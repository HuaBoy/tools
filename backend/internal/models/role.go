package models

import "time"

// Role 角色模型
type Role struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
	Permissions []string  `json:"permissions"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Feature 功能模型
type Feature struct {
	ID          int                    `json:"id"`
	Code        string                 `json:"code"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	Enabled     bool                   `json:"enabled"`
	Order       int                    `json:"order"`
	Config      map[string]interface{} `json:"config"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// FeatureCreateRequest 创建功能请求
type FeatureCreateRequest struct {
	Code        string                 `json:"code" binding:"required"`
	Name        string                 `json:"name" binding:"required"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	Enabled     *bool                  `json:"enabled"`
	Order       *int                   `json:"order"`
	Config      map[string]interface{} `json:"config"`
}

// FeatureUpdateRequest 更新功能请求
type FeatureUpdateRequest struct {
	Code        *string                `json:"code,omitempty"`
	Name        *string                `json:"name,omitempty"`
	Description *string                `json:"description,omitempty"`
	Category    *string                `json:"category,omitempty"`
	Enabled     *bool                  `json:"enabled,omitempty"`
	Order       *int                   `json:"order,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
}
