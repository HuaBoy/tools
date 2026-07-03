package handlers

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tester-platform/backend/internal/models"
	"github.com/tester-platform/backend/internal/supabase"
	"github.com/tester-platform/backend/pkg/jwt"
	"github.com/tester-platform/backend/pkg/response"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	Client     *supabase.Client
	JWTManager *jwt.Manager
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(client *supabase.Client, jwtManager *jwt.Manager) *AuthHandler {
	return &AuthHandler{
		Client:     client,
		JWTManager: jwtManager,
	}
}

// Login 登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 调用 Supabase Auth API
	authResp, err := h.Client.SignInWithPassword(req.Email, req.Password)
	if err != nil {
		response.Unauthorized(c, "登录失败: "+err.Error())
		return
	}

	// 获取用户详细信息（包括 role、permissions）
	userInfo, err := h.Client.Query("user_profiles", map[string]string{
		"email": "eq." + req.Email,
		"select": "id,username,email,role,permissions",
	}, true)
	if err != nil || userInfo.Error != nil {
		response.Unauthorized(c, "获取用户信息失败")
		return
	}

	// 解析用户数据
	users, _ := userInfo.Data.([]interface{})
	if len(users) == 0 {
		response.Unauthorized(c, "用户不存在")
		return
	}

	user := users[0].(map[string]interface{})
	role, _ := user["role"].(string)
	if role == "" {
		role = "user"
	}
	permissions := []string{}
	if perms, ok := user["permissions"].([]interface{}); ok {
		for _, p := range perms {
			if ps, ok := p.(string); ok {
				permissions = append(permissions, ps)
			}
		}
	}

	userID, _ := user["id"].(string)
	username, _ := user["username"].(string)

	// 生成本系统 JWT token
	token, err := h.JWTManager.Generate(userID, username, req.Email, role, permissions)
	if err != nil {
		response.ServerError(c, "生成 token 失败: "+err.Error())
		return
	}

	response.Success(c, models.LoginResponse{
		Token:       token,
		ExpiresIn:   h.JWTManager.ExpireHours() * 3600,
		UserID:      userID,
		Username:    username,
		Email:       req.Email,
		Role:        role,
		Permissions: permissions,
	})

	// 存储 Supabase token 到响应头（供前端调用 Supabase 使用）
	c.Header("X-Supabase-Token", authResp.AccessToken)
}

// GetProfile 获取当前用户信息
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := c.GetString("userID")
	username := c.GetString("username")
	email := c.GetString("email")
	role := c.GetString("role")
	permissions, _ := c.Get("permissions")

	c.JSON(200, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"user_id":     userID,
			"username":    username,
			"email":       email,
			"role":        role,
			"permissions": permissions,
		},
	})
}

// Logout 登出
func (h *AuthHandler) Logout(c *gin.Context) {
	// 简单实现：前端清除 token 即可
	// 如需 Supabase 端登出，可调用 supabase.auth.signOut()
	response.Success(c, gin.H{"message": "登出成功"})
}

// Register 注册
func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Email    string                 `json:"email" binding:"required,email"`
		Password string                 `json:"password" binding:"required,min=6"`
		Username string                 `json:"username" binding:"required"`
		Role     string                 `json:"role"`
		Data     map[string]interface{} `json:"data"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 准备用户元数据
	userData := map[string]interface{}{
		"username": req.Username,
	}
	if req.Role != "" {
		userData["role"] = req.Role
	}
	for k, v := range req.Data {
		userData[k] = v
	}

	// 调用 Supabase 注册
	_, err := h.Client.SignUp(req.Email, req.Password, userData)
	if err != nil {
		// 如果错误是"user already exists"，提供友好提示
		if strings.Contains(err.Error(), "already") {
			response.BusinessError(c, 1001, "用户已存在")
			return
		}
		response.BadRequest(c, "注册失败: "+err.Error())
		return
	}

	response.SuccessWithMessage(c, "注册成功，请登录", nil)
}
