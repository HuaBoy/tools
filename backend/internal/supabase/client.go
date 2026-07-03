package supabase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/tester-platform/backend/internal/config"
)

// Client Supabase 客户端
type Client struct {
	BaseURL    string
	AnonKey    string
	ServiceKey string
	HTTPClient *http.Client
}

// NewClient 创建 Supabase 客户端
func NewClient() *Client {
	cfg := config.App.Supabase
	return &Client{
		BaseURL:    strings.TrimRight(cfg.URL, "/"),
		AnonKey:    cfg.AnonKey,
		ServiceKey: cfg.ServiceKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewClientWithToken 创建带自定义 token 的客户端（用于用户请求）
func (c *Client) NewClientWithToken(token string) *Client {
	client := &Client{
		BaseURL:    c.BaseURL,
		AnonKey:    c.AnonKey,
		ServiceKey: c.ServiceKey,
		HTTPClient: c.HTTPClient,
	}
	return client
}

// RequestOptions 请求选项
type RequestOptions struct {
	Method   string
	Path     string
	Body     interface{}
	Headers  map[string]string
	UseAdmin bool // 是否使用 service_role key
}

// QueryResult 通用查询结果
type QueryResult struct {
	Data  interface{} `json:"data"`
	Error *APIError   `json:"error,omitempty"`
	Count int         `json:"count,omitempty"`
}

// APIError API 错误
type APIError struct {
	Message string `json:"message"`
	Code    string `json:"code"`
	Details string `json:"details,omitempty"`
}

// Request 通用请求方法
func (c *Client) Request(opts RequestOptions) (*QueryResult, error) {
	url := c.BaseURL + opts.Path

	var bodyReader io.Reader
	if opts.Body != nil {
		bodyBytes, err := json.Marshal(opts.Body)
		if err != nil {
			return nil, fmt.Errorf("序列化请求体失败: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequest(opts.Method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置默认头部
	apiKey := c.AnonKey
	authHeader := "Bearer " + c.AnonKey
	if opts.UseAdmin && c.ServiceKey != "" {
		apiKey = c.ServiceKey
		authHeader = "Bearer " + c.ServiceKey
	}

	req.Header.Set("apikey", apiKey)
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation,count=exact")

	// 设置自定义头部
	for key, value := range opts.Headers {
		req.Header.Set(key, value)
	}

	// 如果有自定义 Authorization 头，使用它
	if auth, ok := opts.Headers["Authorization"]; ok {
		req.Header.Set("Authorization", auth)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode >= 400 {
		var apiErr APIError
		_ = json.Unmarshal(body, &apiErr)
		if apiErr.Message == "" {
			apiErr.Message = string(body)
		}
		apiErr.Code = fmt.Sprintf("HTTP_%d", resp.StatusCode)
		return &QueryResult{Error: &apiErr}, nil
	}

	// 解析响应，优先解析为 data
	var data interface{}
	if len(body) > 0 {
		if err := json.Unmarshal(body, &data); err != nil {
			return nil, fmt.Errorf("解析响应失败: %w", err)
		}
	}

	// 处理 count header
	count := 0
	if contentRange := resp.Header.Get("Content-Range"); contentRange != "" {
		// 格式: "0-9/100"
		parts := strings.Split(contentRange, "/")
		if len(parts) == 2 {
			_, _ = fmt.Sscanf(parts[1], "%d", &count)
		}
	}

	return &QueryResult{Data: data, Count: count}, nil
}

// Query 查询数据（带过滤、分页、排序）
func (c *Client) Query(table string, params map[string]string, useAdmin bool) (*QueryResult, error) {
	path := fmt.Sprintf("/rest/v1/%s", table)
	if len(params) > 0 {
		path += "?"
		parts := make([]string, 0)
		for k, v := range params {
			parts = append(parts, fmt.Sprintf("%s=%s", k, v))
		}
		path += strings.Join(parts, "&")
	}

	return c.Request(RequestOptions{
		Method:   http.MethodGet,
		Path:     path,
		UseAdmin: useAdmin,
	})
}

// Insert 插入数据
func (c *Client) Insert(table string, data interface{}, useAdmin bool) (*QueryResult, error) {
	return c.Request(RequestOptions{
		Method:   http.MethodPost,
		Path:     fmt.Sprintf("/rest/v1/%s", table),
		Body:     data,
		UseAdmin: useAdmin,
	})
}

// Update 更新数据
func (c *Client) Update(table string, data interface{}, filters map[string]string, useAdmin bool) (*QueryResult, error) {
	path := fmt.Sprintf("/rest/v1/%s?", table)
	parts := make([]string, 0)
	for k, v := range filters {
		parts = append(parts, fmt.Sprintf("%s=eq.%s", k, v))
	}
	path += strings.Join(parts, "&")

	return c.Request(RequestOptions{
		Method:   http.MethodPatch,
		Path:     path,
		Body:     data,
		UseAdmin: useAdmin,
	})
}

// Delete 删除数据
func (c *Client) Delete(table string, filters map[string]string, useAdmin bool) (*QueryResult, error) {
	path := fmt.Sprintf("/rest/v1/%s?", table)
	parts := make([]string, 0)
	for k, v := range filters {
		parts = append(parts, fmt.Sprintf("%s=eq.%s", k, v))
	}
	path += strings.Join(parts, "&")

	return c.Request(RequestOptions{
		Method:   http.MethodDelete,
		Path:     path,
		UseAdmin: useAdmin,
	})
}

// ========================================
// Auth API
// ========================================

// SignInRequest 登录请求
type SignInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// SignInResponse 登录响应
type SignInResponse struct {
	AccessToken  string      `json:"access_token"`
	TokenType    string      `json:"token_type"`
	ExpiresIn    int         `json:"expires_in"`
	ExpiresAt    int64       `json:"expires_at"`
	RefreshToken string      `json:"refresh_token"`
	User         interface{} `json:"user"`
}

// SignInWithPassword 邮箱密码登录
func (c *Client) SignInWithPassword(email, password string) (*SignInResponse, error) {
	reqBody := map[string]interface{}{
		"email":    email,
		"password": password,
	}

	result, err := c.Request(RequestOptions{
		Method: http.MethodPost,
		Path:   "/auth/v1/token?grant_type=password",
		Body:   reqBody,
	})
	if err != nil {
		return nil, err
	}

	if result.Error != nil {
		return nil, fmt.Errorf("登录失败: %s", result.Error.Message)
	}

	dataBytes, _ := json.Marshal(result.Data)
	var resp SignInResponse
	if err := json.Unmarshal(dataBytes, &resp); err != nil {
		return nil, fmt.Errorf("解析登录响应失败: %w", err)
	}

	return &resp, nil
}

// SignUpRequest 注册请求
type SignUpRequest struct {
	Email    string                 `json:"email"`
	Password string                 `json:"password"`
	Data     map[string]interface{} `json:"data,omitempty"`
}

// SignUpResponse 注册响应
type SignUpResponse struct {
	User interface{} `json:"user"`
}

// SignUp 注册新用户
func (c *Client) SignUp(email, password string, userData map[string]interface{}) (*SignUpResponse, error) {
	reqBody := SignUpRequest{
		Email:    email,
		Password: password,
		Data:     userData,
	}

	result, err := c.Request(RequestOptions{
		Method: http.MethodPost,
		Path:   "/auth/v1/signup",
		Body:   reqBody,
	})
	if err != nil {
		return nil, err
	}

	if result.Error != nil {
		return nil, fmt.Errorf("注册失败: %s", result.Error.Message)
	}

	dataBytes, _ := json.Marshal(result.Data)
	var resp SignUpResponse
	if err := json.Unmarshal(dataBytes, &resp); err != nil {
		return nil, fmt.Errorf("解析注册响应失败: %w", err)
	}

	return &resp, nil
}

// GetUser 获取当前用户信息
func (c *Client) GetUser(accessToken string) (interface{}, error) {
	result, err := c.Request(RequestOptions{
		Method: http.MethodGet,
		Path:   "/auth/v1/user",
		Headers: map[string]string{
			"Authorization": "Bearer " + accessToken,
		},
	})
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, fmt.Errorf("获取用户失败: %s", result.Error.Message)
	}
	return result.Data, nil
}
