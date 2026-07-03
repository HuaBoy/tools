package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tester-platform/backend/pkg/jwt"
	"github.com/tester-platform/backend/pkg/response"
)

// JWTAuth JWT 认证中间件
func JWTAuth(jwtManager *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Header 获取 token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "请先登录")
			c.Abort()
			return
		}

		// 解析 Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "无效的 Authorization 头")
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := jwtManager.Parse(tokenString)
		if err != nil {
			response.Unauthorized(c, "Token 已过期或无效")
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Set("permissions", claims.Permissions)
		c.Set("token", tokenString)

		c.Next()
	}
}

// RequirePermission 权限检查中间件
func RequirePermission(required ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		permissions, exists := c.Get("permissions")
		if !exists {
			response.Forbidden(c, "无权访问")
			c.Abort()
			return
		}

		userPerms, ok := permissions.([]string)
		if !ok {
			response.Forbidden(c, "权限信息错误")
			c.Abort()
			return
		}

		// 检查必需权限（任一满足即可）
		for _, req := range required {
			for _, perm := range userPerms {
				if perm == req {
					c.Next()
					return
				}
			}
		}

		response.Forbidden(c, "缺少必要权限: "+strings.Join(required, ", "))
		c.Abort()
	}
}

// RequireRole 角色检查中间件
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			response.Forbidden(c, "无权访问")
			c.Abort()
			return
		}

		userRole, ok := role.(string)
		if !ok {
			response.Forbidden(c, "角色信息错误")
			c.Abort()
			return
		}

		for _, r := range roles {
			if r == userRole {
				c.Next()
				return
			}
		}

		response.Forbidden(c, "角色不匹配: "+strings.Join(roles, ", "))
		c.Abort()
	}
}
