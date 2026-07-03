package models

import "time"

// User 用户模型
type User struct {
	ID          string    `json:"id" db:"id"`
	Username    string    `json:"username" db:"username"`
	Email       string    `json:"email" db:"email"`
	Role        string    `json:"role" db:"role"`
	Permissions []string  `json:"permissions" db:"permissions"`
	AvatarURL   string    `json:"avatar_url" db:"avatar_url"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// UserCreateRequest 创建用户请求
type UserCreateRequest struct {
	Username    string   `json:"username" binding:"required"`
	Email       string   `json:"email" binding:"required,email"`
	Password    string   `json:"password" binding:"required,min=6"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
}

// UserUpdateRequest 更新用户请求
type UserUpdateRequest struct {
	Username    *string   `json:"username,omitempty"`
	Email       *string   `json:"email,omitempty"`
	Role        *string   `json:"role,omitempty"`
	Permissions *[]string `json:"permissions,omitempty"`
	AvatarURL   *string   `json:"avatar_url,omitempty"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token       string   `json:"token"`
	ExpiresIn   int      `json:"expires_in"`
	UserID      string   `json:"user_id"`
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
}
