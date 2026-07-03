package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tester-platform/backend/internal/models"
	"github.com/tester-platform/backend/internal/supabase"
	"github.com/tester-platform/backend/pkg/response"
)

// UserHandler 用户管理处理器
type UserHandler struct {
	Client *supabase.Client
}

// NewUserHandler 创建用户处理器
func NewUserHandler(client *supabase.Client) *UserHandler {
	return &UserHandler{Client: client}
}

// List 用户列表
func (h *UserHandler) List(c *gin.Context) {
	current, _ := strconv.Atoi(c.DefaultQuery("current", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	keyword := c.Query("keyword")
	role := c.Query("role")

	if current < 1 {
		current = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	params := map[string]string{
		"select": "id,username,email,role,permissions,avatar_url,created_at,updated_at",
		"order":  "created_at.desc",
		"offset": strconv.Itoa((current - 1) * size),
		"limit":  strconv.Itoa(size),
	}

	if keyword != "" {
		params["or"] = "(username.ilike.*" + keyword + "*,email.ilike.*" + keyword + "*)"
	}
	if role != "" {
		params["role"] = "eq." + role
	}

	result, err := h.Client.Query("user_profiles", params, true)
	if err != nil {
		response.ServerError(c, "查询用户失败: "+err.Error())
		return
	}
	if result.Error != nil {
		response.ServerError(c, "查询用户失败: "+result.Error.Message)
		return
	}

	response.Page(c, result.Data, result.Count, current, size)
}

// Get 获取单个用户
func (h *UserHandler) Get(c *gin.Context) {
	id := c.Param("id")

	result, err := h.Client.Query("user_profiles", map[string]string{
		"id":     "eq." + id,
		"select": "id,username,email,role,permissions,avatar_url,created_at,updated_at",
	}, true)
	if err != nil || result.Error != nil {
		response.NotFound(c, "用户不存在")
		return
	}

	users, _ := result.Data.([]interface{})
	if len(users) == 0 {
		response.NotFound(c, "用户不存在")
		return
	}

	response.Success(c, users[0])
}

// Create 创建用户
func (h *UserHandler) Create(c *gin.Context) {
	var req models.UserCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userData := map[string]interface{}{
		"username":    req.Username,
		"role":        req.Role,
		"permissions": req.Permissions,
	}

	signUpResult, err := h.Client.SignUp(req.Email, req.Password, userData)
	if err != nil {
		response.BadRequest(c, "创建用户失败: "+err.Error())
		return
	}

	userMap, ok := signUpResult.User.(map[string]interface{})
	if !ok {
		response.BadRequest(c, "创建用户失败: 无法获取用户信息")
		return
	}

	userId, ok := userMap["id"].(string)
	if !ok {
		response.BadRequest(c, "创建用户失败: 无法获取用户ID")
		return
	}

	profileData := map[string]interface{}{
		"auth_id":     userId,
		"username":    req.Username,
		"email":       req.Email,
		"role":        req.Role,
		"permissions": req.Permissions,
	}

	_, err = h.Client.Insert("user_profiles", profileData, true)
	if err != nil {
		response.BadRequest(c, "创建用户配置失败: "+err.Error())
		return
	}

	response.SuccessWithMessage(c, "用户创建成功", nil)
}

// Update 更新用户
func (h *UserHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req models.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 只更新非空字段
	updateData := make(map[string]interface{})
	if req.Username != nil {
		updateData["username"] = *req.Username
	}
	if req.Email != nil {
		updateData["email"] = *req.Email
	}
	if req.Role != nil {
		updateData["role"] = *req.Role
	}
	if req.Permissions != nil {
		updateData["permissions"] = *req.Permissions
	}
	if req.AvatarURL != nil {
		updateData["avatar_url"] = *req.AvatarURL
	}

	if len(updateData) == 0 {
		response.BadRequest(c, "没有要更新的字段")
		return
	}

	result, err := h.Client.Update("user_profiles", updateData, map[string]string{
		"id": "eq." + id,
	}, true)
	if err != nil || result.Error != nil {
		response.ServerError(c, "更新用户失败")
		return
	}

	response.SuccessWithMessage(c, "用户更新成功", result.Data)
}

// Delete 删除用户
func (h *UserHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	_, err := h.Client.Request(supabase.RequestOptions{
		Method:   "DELETE",
		Path:     "/auth/v1/admin/users/" + id,
		UseAdmin: true,
	})
	if err != nil {
		response.ServerError(c, "删除用户失败")
		return
	}

	response.SuccessWithMessage(c, "用户删除成功", nil)
}

// ResetPassword 重置密码
func (h *UserHandler) ResetPassword(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 1. 查询用户邮箱
	userResult, err := h.Client.Query("user_profiles", map[string]string{
		"id":     "eq." + id,
		"select": "email",
	}, true)
	if err != nil || userResult.Error != nil {
		response.NotFound(c, "用户不存在")
		return
	}

	users, _ := userResult.Data.([]interface{})
	if len(users) == 0 {
		response.NotFound(c, "用户不存在")
		return
	}

	email := users[0].(map[string]interface{})["email"].(string)

	// 2. 调用 Supabase Admin API 重置密码
	// 注意：需要 service_role key
	resetReq := map[string]interface{}{
		"email": email,
		"password": req.NewPassword,
	}

	_, err = h.Client.Request(supabase.RequestOptions{
		Method:   "PUT",
		Path:     "/auth/v1/admin/users/" + id,
		Body:     resetReq,
		UseAdmin: true,
	})
	if err != nil {
		response.ServerError(c, "重置密码失败: "+err.Error())
		return
	}

	response.SuccessWithMessage(c, "密码重置成功", nil)
}
