package router

import (
	"database/sql"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/tester-platform/backend/internal/config"
	"github.com/tester-platform/backend/internal/handlers"
	"github.com/tester-platform/backend/internal/knowledge"
	"github.com/tester-platform/backend/internal/middleware"
	"github.com/tester-platform/backend/internal/ollama"
	"github.com/tester-platform/backend/internal/supabase"
	"github.com/tester-platform/backend/pkg/jwt"
)

// Setup 设置路由
func Setup(client *supabase.Client, jwtManager *jwt.Manager, db *sql.DB) *gin.Engine {
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

	// 初始化 Ollama 客户端 + 知识库服务
	ollamaClient := ollama.NewClient(config.App.Ollama.URL)
	knowledgeService := knowledge.NewService(
		db,
		ollamaClient,
		config.App.Ollama.EmbeddingModel,
		config.App.Ollama.ChatModel,
		config.App.Ollama.StoreDir,
	)

	// 初始化 handlers
	healthHandler := handlers.NewHealthHandler()
	authHandler := handlers.NewAuthHandler(client, jwtManager, db)
	userHandler := handlers.NewUserHandler(client)
	roleHandler := handlers.NewRoleHandler(client)
	featureHandler := handlers.NewFeatureHandler(client)
	overseasShippingHandler := handlers.NewOverseasShippingHandler(db)
	pipelineHandler := handlers.NewPipelineHandler(db)
	manualHandler := handlers.NewOperationManualHandler(db)
	videoHandler := handlers.NewOperationVideoHandler(db)
	knowledgeHandler := handlers.NewKnowledgeHandler(knowledgeService)
	chatHandler := handlers.NewChatHandler(knowledgeService)

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

		// 海外发货管理（登录即可访问）
		overseasShipping := auth.Group("/overseas-shipping")
		{
			overseasShipping.GET("", overseasShippingHandler.List)
			overseasShipping.GET("/:id", overseasShippingHandler.Get)
			overseasShipping.POST("", overseasShippingHandler.Create)
			overseasShipping.PATCH("/:id", overseasShippingHandler.Update)
			overseasShipping.DELETE("/:id", overseasShippingHandler.Delete)
		}

		// ===== AI 研发流水线（REQ-2026-001）=====
		pipeline := auth.Group("/pipeline")
		{
			pipeline.GET("/requirements", pipelineHandler.ListRequirements)
			pipeline.GET("/requirements/:id", pipelineHandler.GetRequirement)
			pipeline.POST("/requirements", pipelineHandler.CreateRequirement)
			pipeline.PATCH("/requirements/:id", pipelineHandler.UpdateRequirement)
			pipeline.DELETE("/requirements/:id", pipelineHandler.DeleteRequirement)
			pipeline.PATCH("/requirements/:id/status", pipelineHandler.ChangeStatus)
			pipeline.GET("/requirements/:id/stages", pipelineHandler.GetStages)
			pipeline.GET("/requirements/:id/deliverables", pipelineHandler.GetDeliverables)
			pipeline.POST("/requirements/:id/deliverables", pipelineHandler.CreateDeliverable)
			pipeline.DELETE("/deliverables/:id", pipelineHandler.DeleteDeliverable)
			pipeline.GET("/requirements/:id/reviews", pipelineHandler.GetReviews)
			pipeline.POST("/requirements/:id/reviews", pipelineHandler.CreateReview)
			pipeline.GET("/board", pipelineHandler.GetBoard)
		}

		// 操作手册
		manual := auth.Group("/operation-manuals")
		{
			manual.GET("", manualHandler.List)
			manual.GET("/:id", manualHandler.Get)
			manual.POST("", manualHandler.Create)
			manual.PATCH("/:id", manualHandler.Update)
			manual.DELETE("/:id", manualHandler.Delete)
		}

		// 操作视频
		video := auth.Group("/operation-videos")
		{
			video.GET("", videoHandler.List)
			video.GET("/:id", videoHandler.Get)
			video.POST("", videoHandler.Create)
			video.PATCH("/:id", videoHandler.Update)
			video.DELETE("/:id", videoHandler.Delete)
		}

		// ===== 知识库 + 问题库 + AI 对话 =====
		knowledgeGroup := auth.Group("/knowledge")
		{
			// 统计与状态
			knowledgeGroup.GET("/stats", knowledgeHandler.Stats)

			// 文档管理
			knowledgeGroup.GET("/documents", knowledgeHandler.List)
			knowledgeGroup.GET("/documents/:id", knowledgeHandler.Get)
			knowledgeGroup.POST("/documents/upload", knowledgeHandler.Upload)
			knowledgeGroup.DELETE("/documents/:id", knowledgeHandler.Delete)

			// 搜索
			knowledgeGroup.GET("/search", knowledgeHandler.Search)
			knowledgeGroup.POST("/search", knowledgeHandler.Search)

			// 问题库 FAQ
			knowledgeGroup.GET("/faqs", knowledgeHandler.ListFAQs)
			knowledgeGroup.POST("/faqs", knowledgeHandler.CreateFAQ)
			knowledgeGroup.DELETE("/faqs/:id", knowledgeHandler.DeleteFAQ)

			// AI 对话
			knowledgeGroup.POST("/chat", chatHandler.Send)

			// 会话管理
			knowledgeGroup.GET("/conversations", chatHandler.Conversations)
			knowledgeGroup.GET("/conversations/:id/messages", chatHandler.History)
		}
	}
	}

	return r
}
