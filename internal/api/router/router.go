package router

import (
	"github.com/gin-contrib/cors"
	"github.com/wb-go/wbf/ginext"

	"github.com/aliskhannn/warehouse-control/internal/api/handler/audit"
	"github.com/aliskhannn/warehouse-control/internal/api/handler/auth"
	"github.com/aliskhannn/warehouse-control/internal/api/handler/item"
	"github.com/aliskhannn/warehouse-control/internal/config"
	"github.com/aliskhannn/warehouse-control/internal/middleware"
)

// New creates a new Gin engine and sets up routes for the API.
func New(
	authHandler *auth.Handler,
	itemHandler *item.Handler,
	auditHandler *audit.Handler,
	cfg *config.Config,
) *ginext.Engine {
	e := ginext.New()

	e.Use(ginext.Logger())
	e.Use(ginext.Recovery())
	e.Use(cors.Default())

	api := e.Group("/api")
	{
		// --- Auth routes ---
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)
		}

		// --- Item routes ---
		itemGroup := api.Group("/items")
		{
			// Public: anyone can view items
			itemGroup.GET("", itemHandler.GetAll)
			itemGroup.GET("/:id", itemHandler.GetByID)

			// Protected routes
			itemGroup.Use(middleware.Auth(cfg.JWT.Secret, cfg.JWT.TTL))
			{
				// Admin: full access (create, update, delete)
				adminGroup := itemGroup.Group("")
				adminGroup.Use(middleware.RequireRole("admin"))
				{
					adminGroup.POST("", itemHandler.Create)
					adminGroup.PUT("/:id", itemHandler.Update)
					adminGroup.DELETE("/:id", itemHandler.Delete)
				}

				// Manager: can create and update but not delete
				managerGroup := itemGroup.Group("")
				managerGroup.Use(middleware.RequireRole("manager"))
				{
					managerGroup.POST("", itemHandler.Create)
					managerGroup.PUT("/:id", itemHandler.Update)
				}

				// Viewer: read-only access is already handled by public GET endpoints
			}
		}

		// --- Audit routes ---
		auditGroup := api.Group("/audit")
		auditGroup.Use(middleware.Auth(cfg.JWT.Secret, cfg.JWT.TTL))
		{
			// Only admin can access audit endpoints
			auditGroup.Use(middleware.RequireRole("admin"))
			auditGroup.GET("/items/:id/history", auditHandler.GetHistory)
			auditGroup.POST("/items/compare", auditHandler.CompareVersions)
		}
	}

	return e
}
