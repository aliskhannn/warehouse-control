package main

import (
	"context"
	"errors"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/zlog"

	"github.com/aliskhannn/warehouse-control/internal/api/handler/audit"
	"github.com/aliskhannn/warehouse-control/internal/api/handler/auth"
	"github.com/aliskhannn/warehouse-control/internal/api/handler/item"
	"github.com/aliskhannn/warehouse-control/internal/api/router"
	"github.com/aliskhannn/warehouse-control/internal/api/server"
	"github.com/aliskhannn/warehouse-control/internal/config"
	repoitem "github.com/aliskhannn/warehouse-control/internal/repository/item"
	repouser "github.com/aliskhannn/warehouse-control/internal/repository/user"
	serviceitem "github.com/aliskhannn/warehouse-control/internal/service/item"
	serviceuser "github.com/aliskhannn/warehouse-control/internal/service/user"
)

func main() {
	// Initialize logger, configuration and validator.
	zlog.Init()
	cfg := config.MustLoad()
	val := validator.New()

	// Connect to PostgreSQL master and slave databases.
	opts := &dbpg.Options{
		MaxOpenConns:    cfg.Database.MaxOpenConnections,
		MaxIdleConns:    cfg.Database.MaxIdleConnections,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
	}

	slaveDNSs := make([]string, 0, len(cfg.Database.Slaves))

	for _, s := range cfg.Database.Slaves {
		slaveDNSs = append(slaveDNSs, s.DSN())
	}

	db, err := dbpg.New(cfg.Database.Master.DSN(), slaveDNSs, opts)
	if err != nil {
		zlog.Logger.Fatal().Err(err).Msg("failed to connect to database")
	}

	// Initialize user repository, service, and handler for auth endpoints.
	userRepo := repouser.NewRepository(db)
	userService := serviceuser.NewService(userRepo, cfg)
	authHandler := auth.NewHandler(userService, val)

	// Initialize item repository, service.
	itemRepo := repoitem.NewRepository(db)
	itemService := serviceitem.NewService(itemRepo)

	// Initialize handlers for item and audit endpoints.
	itemHandler := item.NewHandler(itemService, val)
	auditHandler := audit.NewHandler(itemService)

	// Initialize API router and HTTP server.
	r := router.New(authHandler, itemHandler, auditHandler, cfg)
	s := server.New(cfg.Server.HTTPPort, r)

	// Start HTTP server in a separate goroutine.
	go func() {
		if err := s.ListenAndServe(); err != nil {
			zlog.Logger.Fatal().Err(err).Msg("failed to start server")
		}
	}()

	// Setup context to handle SIGINT and SIGTERM for graceful shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Wait for shutdown signal.
	<-ctx.Done()
	zlog.Logger.Print("shutdown signal received")

	// Gracefully shutdown server with timeout.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	zlog.Logger.Print("gracefully shutting down server...\n")
	if err := s.Shutdown(shutdownCtx); err != nil {
		zlog.Logger.Error().Err(err).Msg("failed to shutdown server")
	}
	if errors.Is(shutdownCtx.Err(), context.DeadlineExceeded) {
		zlog.Logger.Info().Msg("timeout exceeded, forcing shutdown")
	}

	zlog.Logger.Print("closing master and slave databases...\n")

	// Close master database connection.
	if err := db.Master.Close(); err != nil {
		zlog.Logger.Printf("failed to close master DB: %v", err)
	}

	// Close slave database connections.
	for i, s := range db.Slaves {
		if err := s.Close(); err != nil {
			zlog.Logger.Printf("failed to close slave DB %d: %v", i, err)
		}
	}
}
