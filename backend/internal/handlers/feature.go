package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tester-platform/backend/internal/models"
	"github.com/tester-platform/backend/internal/supabase"
	"github.com/tester-platform/backend/pkg/response"
)

// FeatureHandler 功能管理处理器
type FeatureHandler struct {
	Client *supabase.Client
}

// NewFeatureHandler 创建功能处理器
func NewFeatureHandler(client *supabase.Client) *FeatureHandler {
	return &FeatureHandler{Client: client}
}

// List 功能列表
func (h *FeatureHandler) List(c *gin.Context) {
	category := c.Query("category")
	enabled := c.Query("enabled")

	params := map[string]string{
		"select": "*",
		"order":  "order.asc,id.asc",
	}

	if category != "" {
		params["category"] = "eq." + category
	}
	if enabled != "" {
		params["enabled"] = "eq." + enabled
	}

	result, err := h.Client.Query("features", params, true)
	if err != nil || result.Error != nil {
		response.ServerError(c, "查询功能失败")
		return
	}

	response.Success(c, gin.H{
		"list":  result.Data,
		"total": result.Count,
	})
}

// Create 创建功能
func (h *FeatureHandler) Create(c *gin.Context) {
	var req models.FeatureCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	if req.Enabled == nil {
		enabled := true
		req.Enabled = &enabled
	}
	if req.Order == nil {
		order := 0
		req.Order = &order
	}

	data := map[string]interface{}{
		"code":        req.Code,
		"name":        req.Name,
		"description": req.Description,
		"category":    req.Category,
		"enabled":     *req.Enabled,
		"order":       *req.Order,
		"config":      req.Config,
	}

	result, err := h.Client.Insert("features", data, true)
	if err != nil || result.Error != nil {
		response.BadRequest(c, "创建功能失败: "+err.Error())
		return
	}

	response.SuccessWithMessage(c, "功能创建成功", result.Data)
}

// Update 更新功能
func (h *FeatureHandler) Update(c *gin.Context) {
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		response.BadRequest(c, "无效的 ID")
		return
	}

	var req models.FeatureUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	updateData := make(map[string]interface{})
	if req.Code != nil {
		updateData["code"] = *req.Code
	}
	if req.Name != nil {
		updateData["name"] = *req.Name
	}
	if req.Description != nil {
		updateData["description"] = *req.Description
	}
	if req.Category != nil {
		updateData["category"] = *req.Category
	}
	if req.Enabled != nil {
		updateData["enabled"] = *req.Enabled
	}
	if req.Order != nil {
		updateData["order"] = *req.Order
	}
	if req.Config != nil {
		updateData["config"] = req.Config
	}

	if len(updateData) == 0 {
		response.BadRequest(c, "没有要更新的字段")
		return
	}

	result, err := h.Client.Update("features", updateData, map[string]string{
		"id": "eq." + strconv.Itoa(idInt),
	}, true)
	if err != nil || result.Error != nil {
		response.ServerError(c, "更新功能失败")
		return
	}

	response.SuccessWithMessage(c, "功能更新成功", result.Data)
}

// Delete 删除功能
func (h *FeatureHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		response.BadRequest(c, "无效的 ID")
		return
	}

	result, err := h.Client.Delete("features", map[string]string{
		"id": "eq." + strconv.Itoa(idInt),
	}, true)
	if err != nil || result.Error != nil {
		response.ServerError(c, "删除功能失败")
		return
	}

	response.SuccessWithMessage(c, "功能删除成功", result.Data)
}

// Toggle 切换功能启用状态
func (h *FeatureHandler) Toggle(c *gin.Context) {
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		response.BadRequest(c, "无效的 ID")
		return
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	result, err := h.Client.Update("features", map[string]interface{}{
		"enabled": req.Enabled,
	}, map[string]string{
		"id": "eq." + strconv.Itoa(idInt),
	}, true)
	if err != nil || result.Error != nil {
		response.ServerError(c, "切换功能状态失败")
		return
	}

	response.SuccessWithMessage(c, "状态已更新", result.Data)
}
