package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tester-platform/backend/internal/supabase"
	"github.com/tester-platform/backend/pkg/response"
)

// RoleHandler 角色管理处理器
type RoleHandler struct {
	Client *supabase.Client
}

// NewRoleHandler 创建角色处理器
func NewRoleHandler(client *supabase.Client) *RoleHandler {
	return &RoleHandler{Client: client}
}

// List 角色列表
func (h *RoleHandler) List(c *gin.Context) {
	result, err := h.Client.Query("roles", map[string]string{
		"select": "*",
		"order":  "created_at.desc",
	}, true)
	if err != nil || result.Error != nil {
		response.ServerError(c, "查询角色失败")
		return
	}

	response.Success(c, gin.H{
		"list":  result.Data,
		"total": result.Count,
	})
}

// Get 获取角色
func (h *RoleHandler) Get(c *gin.Context) {
	id := c.Param("id")

	result, err := h.Client.Query("roles", map[string]string{
		"id":     "eq." + id,
		"select": "*",
	}, true)
	if err != nil || result.Error != nil {
		response.NotFound(c, "角色不存在")
		return
	}

	roles, _ := result.Data.([]interface{})
	if len(roles) == 0 {
		response.NotFound(c, "角色不存在")
		return
	}

	response.Success(c, roles[0])
}

// Create 创建角色
func (h *RoleHandler) Create(c *gin.Context) {
	var req struct {
		Name        string   `json:"name" binding:"required"`
		Code        string   `json:"code" binding:"required"`
		Description string   `json:"description"`
		Permissions []string `json:"permissions"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	data := map[string]interface{}{
		"name":        req.Name,
		"code":        req.Code,
		"description": req.Description,
		"permissions": req.Permissions,
	}

	result, err := h.Client.Insert("roles", data, true)
	if err != nil || result.Error != nil {
		response.BadRequest(c, "创建角色失败: "+err.Error())
		return
	}

	response.SuccessWithMessage(c, "角色创建成功", result.Data)
}

// Update 更新角色
func (h *RoleHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Name        *string   `json:"name,omitempty"`
		Description *string   `json:"description,omitempty"`
		Permissions *[]string `json:"permissions,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	updateData := make(map[string]interface{})
	if req.Name != nil {
		updateData["name"] = *req.Name
	}
	if req.Description != nil {
		updateData["description"] = *req.Description
	}
	if req.Permissions != nil {
		updateData["permissions"] = *req.Permissions
	}

	if len(updateData) == 0 {
		response.BadRequest(c, "没有要更新的字段")
		return
	}

	result, err := h.Client.Update("roles", updateData, map[string]string{
		"id": "eq." + id,
	}, true)
	if err != nil || result.Error != nil {
		response.ServerError(c, "更新角色失败")
		return
	}

	response.SuccessWithMessage(c, "角色更新成功", result.Data)
}

// Delete 删除角色
func (h *RoleHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	result, err := h.Client.Delete("roles", map[string]string{
		"id": "eq." + id,
	}, true)
	if err != nil || result.Error != nil {
		response.ServerError(c, "删除角色失败")
		return
	}

	response.SuccessWithMessage(c, "角色删除成功", result.Data)
}

// AssignRole 分配角色给用户
func (h *RoleHandler) AssignRole(c *gin.Context) {
	var req struct {
		UserID string `json:"user_id" binding:"required"`
		RoleID string `json:"role_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 查询角色权限
	roleResult, err := h.Client.Query("roles", map[string]string{
		"id":     "eq." + req.RoleID,
		"select": "permissions",
	}, true)
	if err != nil || roleResult.Error != nil {
		response.NotFound(c, "角色不存在")
		return
	}

	roles, _ := roleResult.Data.([]interface{})
	if len(roles) == 0 {
		response.NotFound(c, "角色不存在")
		return
	}

	role := roles[0].(map[string]interface{})
	permissions := []string{}
	if perms, ok := role["permissions"].([]interface{}); ok {
		for _, p := range perms {
			if ps, ok := p.(string); ok {
				permissions = append(permissions, ps)
			}
		}
	}

	// 更新用户的 permissions
	result, err := h.Client.Update("user_profiles", map[string]interface{}{
		"permissions": permissions,
	}, map[string]string{
		"id": "eq." + req.UserID,
	}, true)
	if err != nil || result.Error != nil {
		response.ServerError(c, "分配角色失败")
		return
	}

	response.SuccessWithMessage(c, "角色分配成功", result.Data)
}

// helper for strconv
var _ = strconv.Itoa
