package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/tester-platform/backend/internal/config"
	"github.com/tester-platform/backend/pkg/response"
)

// HealthHandler 健康检查
type HealthHandler struct{}

// NewHealthHandler 创建健康检查处理器
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Health 健康检查
func (h *HealthHandler) Health(c *gin.Context) {
	response.Success(c, gin.H{
		"status":      "ok",
		"version":     "1.0.0",
		"supabase":    config.IsConfigured(),
	})
}

// Config 获取前端配置（如 Supabase 公开配置）
func (h *HealthHandler) Config(c *gin.Context) {
	cfg := config.App.Supabase
	response.Success(c, gin.H{
		"supabase_url":    cfg.URL,
		"supabase_anon":   cfg.AnonKey,
	})
}
