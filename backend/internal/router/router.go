package router

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/tester-platform/backend/internal/config"
	"github.com/tester-platform/backend/internal/handlers"
	"github.com/tester-platform/backend/internal/middleware"
	"github.com/tester-platform/backend/internal/supabase"
	"github.com/tester-platform/backend/pkg/jwt"
)

// Setup 设置路由
func Setup(client *supabase.Client, jwtManager *jwt.Manager) *gin.Engine {
	// 设置 Gin 模式
	gin.SetMode(config.App.Server.Mode)

	r := gin.New()

	// 全局中间件
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     config.App.CORS.Origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Requested-With", "X-Supabase-Token"},
		ExposeHeaders:    []string{"Content-Length", "X-Supabase-Token"},
		AllowCredentials: true,
		MaxAge:           12 * 3600,
	}))

	// 初始化 handlers
	healthHandler := handlers.NewHealthHandler()
	authHandler := handlers.NewAuthHandler(client, jwtManager)
	userHandler := handlers.NewUserHandler(client)
	roleHandler := handlers.NewRoleHandler(client)
	featureHandler := handlers.NewFeatureHandler(client)

	// 健康检查（无需认证）
	r.GET("/health", healthHandler.Health)
	r.GET("/api/v1/config", healthHandler.Config)

	// API v1
	api := r.Group("/api/v1")
	{
		// 公开接口
		public := api.Group("/auth")
		{
			public.POST("/login", authHandler.Login)
			public.POST("/register", authHandler.Register)
		}

		// 需要认证的接口
		auth := api.Group("/")
		auth.Use(middleware.JWTAuth(jwtManager))
		{
			// 当前用户
			auth.GET("/auth/profile", authHandler.GetProfile)
			auth.POST("/auth/logout", authHandler.Logout)

			// 用户管理
			users := auth.Group("/users")
			{
				users.GET("", userHandler.List)
				users.GET("/:id", userHandler.Get)
				users.POST("", middleware.RequirePermission("user:create"), userHandler.Create)
				users.PATCH("/:id", middleware.RequirePermission("user:edit"), userHandler.Update)
				users.DELETE("/:id", middleware.RequirePermission("user:delete"), userHandler.Delete)
				users.POST("/:id/reset-password", middleware.RequirePermission("user:edit"), userHandler.ResetPassword)
			}

			// 角色管理
			roles := auth.Group("/roles")
			{
				roles.GET("", roleHandler.List)
				roles.GET("/:id", roleHandler.Get)
				roles.POST("", middleware.RequirePermission("role:assign"), roleHandler.Create)
				roles.PATCH("/:id", middleware.RequirePermission("role:assign"), roleHandler.Update)
				roles.DELETE("/:id", middleware.RequirePermission("role:assign"), roleHandler.Delete)
				roles.POST("/assign", middleware.RequirePermission("role:assign"), roleHandler.AssignRole)
			}

			// 功能管理
			features := auth.Group("/features")
			{
				features.GET("", featureHandler.List)
				features.POST("", middleware.RequirePermission("feature:toggle"), featureHandler.Create)
				features.PATCH("/:id", middleware.RequirePermission("feature:toggle"), featureHandler.Update)
				features.DELETE("/:id", middleware.RequirePermission("feature:toggle"), featureHandler.Delete)
				features.POST("/:id/toggle", middleware.RequirePermission("feature:toggle"), featureHandler.Toggle)
			}
		}
	}

	return r
}
