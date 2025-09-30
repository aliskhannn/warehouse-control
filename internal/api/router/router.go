package router

import (
	"github.com/gin-contrib/cors"
	"github.com/wb-go/wbf/ginext"

	"github.com/aliskhannn/warehouse-control/internal/api/handler/audit"
	"github.com/aliskhannn/warehouse-control/internal/api/handler/auth"
	"github.com/aliskhannn/warehouse-control/internal/api/handler/item"
	"github.com/aliskhannn/warehouse-control/internal/api/handler/user"
	"github.com/aliskhannn/warehouse-control/internal/config"
	"github.com/aliskhannn/warehouse-control/internal/middleware"
)

// New creates a new Gin engine and sets up routes for the API.
func New(
	authHandler *auth.Handler,
	userHandler *user.Handler,
	itemHandler *item.Handler,
	auditHandler *audit.Handler,
	cfg *config.Config,
) *ginext.Engine {
	e := ginext.New()

	e.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	e.Use(ginext.Logger())
	e.Use(ginext.Recovery())

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
			// Public GET routes (all roles)
			itemGroup.GET("", itemHandler.GetAll)
			itemGroup.GET("/:id", itemHandler.GetByID)

			// Protected routes (requires JWT)
			itemGroup.Use(middleware.Auth(cfg.JWT.Secret, cfg.JWT.TTL))
			{
				// POST /items: admin and manager
				itemGroup.POST("", middleware.RequireRole("admin", "manager"), itemHandler.Create)

				// PUT /items/:id: admin and manager
				itemGroup.PUT("/:id", middleware.RequireRole("admin", "manager"), itemHandler.Update)

				// DELETE /items/:id: admin only
				itemGroup.DELETE("/:id", middleware.RequireRole("admin"), itemHandler.Delete)
			}
		}

		// --- User routes ---
		userGroup := api.Group("/users")
		userGroup.Use(middleware.Auth(cfg.JWT.Secret, cfg.JWT.TTL)) // если нужен JWT
		{
			userGroup.GET("/:id", userHandler.GetByID) // GET /api/users/:id
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
