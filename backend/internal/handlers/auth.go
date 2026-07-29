package handlers

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tester-platform/backend/internal/models"
	"github.com/tester-platform/backend/internal/supabase"
	"github.com/tester-platform/backend/pkg/jwt"
	"github.com/tester-platform/backend/pkg/response"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	Client     *supabase.Client
	JWTManager *jwt.Manager
	DB         *sql.DB
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(client *supabase.Client, jwtManager *jwt.Manager, db *sql.DB) *AuthHandler {
	return &AuthHandler{
		Client:     client,
		JWTManager: jwtManager,
		DB:         db,
	}
}

// sysUser 从 sys_user 表读取的用户（字段名容错）
type sysUser struct {
	ID          string
	Username    string
	Email       string
	Role        string
	Password    string
	Permissions []string
}

// Login 登录（基于 sys_user 表直接校验，不依赖 Supabase Auth）
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	if req.Password == "" {
		response.BadRequest(c, "密码不能为空")
		return
	}

	user, err := h.findSysUser(req.Username)
	if err != nil {
		response.Unauthorized(c, "登录失败: "+err.Error())
		return
	}
	if user == nil {
		response.Unauthorized(c, "用户不存在")
		return
	}

	fmt.Printf("[DEBUG] user=%s pass_len=%d pass_pref=%s input=%s\n",
		user.Username, len(user.Password), user.Password[:min(20, len(user.Password))], req.Password)

	// 校验密码
	if !verifySysPassword(user.Password, req.Password) {
		response.Unauthorized(c, "密码错误")
		return
	}

	// 生成本系统 JWT
	token, err := h.JWTManager.Generate(user.ID, user.Username, user.Email, user.Role, user.Permissions)
	if err != nil {
		response.ServerError(c, "生成 token 失败: "+err.Error())
		return
	}

	response.Success(c, models.LoginResponse{
		Token:       token,
		ExpiresIn:   h.JWTManager.ExpireHours() * 3600,
		UserID:      user.ID,
		Username:    user.Username,
		Email:       user.Email,
		Role:        user.Role,
		Permissions: user.Permissions,
	})
}

// findSysUser 在 sys_user 中查找用户，兼容多种字段命名
func (h *AuthHandler) findSysUser(username string) (*sysUser, error) {
	rows, qerr := h.querySysUserRows(map[string]string{"limit": "1000"})
	if qerr != nil {
		return nil, qerr
	}

	// 第一轮：精确匹配 username 列
	for _, r := range rows {
		if un := getStr(r, "username"); un != "" && strings.EqualFold(un, username) {
			return rowToSysUser(r)
		}
	}

	// 第二轮：宽匹配（user_name, name, login, account...）
	for _, r := range rows {
		if matchSysUser(r, username) {
			return rowToSysUser(r)
		}
	}
	return nil, nil
}

func matchSysUser(m map[string]interface{}, username string) bool {
	un := getStr(m, "username", "user_name", "name", "login", "account", "user_login")
	return un != "" && strings.EqualFold(un, username)
}

func (h *AuthHandler) querySysUserRows(params map[string]string) ([]map[string]interface{}, error) {
	query := "SELECT * FROM public.sys_user"
	if limit, ok := params["limit"]; ok {
		query += " LIMIT " + limit
	}

	rows, err := h.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("查询 sys_user 失败: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("获取列名失败: %w", err)
	}

	var results []map[string]interface{}
	for rows.Next() {
		vals := make([]interface{}, len(cols))
		valPtrs := make([]interface{}, len(cols))
		for i := range vals {
			valPtrs[i] = &vals[i]
		}
		if err := rows.Scan(valPtrs...); err != nil {
			continue
		}
		row := make(map[string]interface{}, len(cols))
		for i, col := range cols {
			// lib/pq 把 text/varchar 扫描为 []byte，统一转 string
			switch v := vals[i].(type) {
			case []byte:
				row[col] = string(v)
			case nil:
				row[col] = ""
			default:
				row[col] = v
			}
		}
		results = append(results, row)
	}
	return results, nil
}

func rowToSysUser(m map[string]interface{}) (*sysUser, error) {
	u := &sysUser{
		ID:       getStr(m, "id", "user_id", "uuid"),
		Username: getStr(m, "username", "user_name", "name", "login", "account"),
		Email:    getStr(m, "email", "mail", "user_email", "email_address"),
		Role:     getStr(m, "role", "user_role", "user_type"),
		Password: getStr(m, "password", "pwd", "passwd", "password_hash", "hash"),
	}
	if u.ID == "" {
		if v, ok := m["id"]; ok {
			u.ID = fmt.Sprintf("%v", v)
		}
	}
	if u.Role == "" {
		u.Role = "user"
	}
	u.Permissions = getStrSlice(m, "permissions", "perms", "permission")
	return u, nil
}

func getStr(m map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			if s, ok := v.(string); ok && s != "" {
				return s
			}
			if b, ok := v.([]byte); ok && len(b) > 0 {
				return string(b)
			}
		}
	}
	return ""
}

func getStrSlice(m map[string]interface{}, keys ...string) []string {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			switch arr := v.(type) {
			case []string:
				return arr
			case []interface{}:
				out := make([]string, 0, len(arr))
				for _, it := range arr {
					if s, ok := it.(string); ok {
						out = append(out, s)
					}
				}
				return out
			case string:
				var out []string
				if json.Unmarshal([]byte(arr), &out) == nil {
					return out
				}
			}
		}
	}
	return []string{}
}

// verifySysPassword 校验 sys_user 中的密码，兼容多种存储方式
func verifySysPassword(stored, input string) bool {
	if stored == "" || input == "" {
		return false
	}
	// bcrypt / pgcrypto bf ($2a$ / $2b$ / $2y$)
	if strings.HasPrefix(stored, "$2a$") || strings.HasPrefix(stored, "$2b$") || strings.HasPrefix(stored, "$2y$") {
		return bcrypt.CompareHashAndPassword([]byte(stored), []byte(input)) == nil
	}
	lower := strings.ToLower(stored)
	if isHex(lower) {
		switch len(lower) {
		case 32: // md5
			return subtle.ConstantTimeCompare([]byte(md5hex(input)), []byte(lower)) == 1
		case 40: // sha1
			return subtle.ConstantTimeCompare([]byte(sha1hex(input)), []byte(lower)) == 1
		case 64: // sha256
			return subtle.ConstantTimeCompare([]byte(sha256hex(input)), []byte(lower)) == 1
		}
	}
	// 明文
	return subtle.ConstantTimeCompare([]byte(stored), []byte(input)) == 1
}

func isHex(s string) bool {
	if s == "" {
		return false
	}
	for _, ch := range s {
		if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f')) {
			return false
		}
	}
	return true
}

func md5hex(s string) string {
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}

func sha1hex(s string) string {
	sum := sha1.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}

func sha256hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
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

// queryRowsToMaps 通用：执行 SQL 查询并将结果转换为 []map[string]interface{}
func queryRowsToMaps(db *sql.DB, query string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}
	for rows.Next() {
		vals := make([]interface{}, len(cols))
		valPtrs := make([]interface{}, len(cols))
		for i := range vals {
			valPtrs[i] = &vals[i]
		}
		if err := rows.Scan(valPtrs...); err != nil {
			continue
		}
		row := make(map[string]interface{}, len(cols))
		for i, col := range cols {
			switch v := vals[i].(type) {
			case []byte:
				row[col] = string(v)
			case nil:
				row[col] = ""
			default:
				row[col] = v
			}
		}
		results = append(results, row)
	}
	return results, nil
}
