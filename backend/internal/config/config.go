package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config 应用配置
type Config struct {
	Server   ServerConfig
	Supabase SupabaseConfig
	JWT      JWTConfig
	CORS     CORSConfig
	Log      LogConfig
	Ollama   OllamaConfig
}

// OllamaConfig Ollama 配置
type OllamaConfig struct {
	URL           string // Ollama 服务地址
	EmbeddingModel string // 嵌入模型
	ChatModel     string // 对话模型
	StoreDir      string // 文档存储目录
}

// ServerConfig 服务配置
type ServerConfig struct {
	Port int
	Host string
	Mode string
}

// SupabaseConfig Supabase 配置
type SupabaseConfig struct {
	URL         string
	AnonKey     string
	ServiceKey  string
	DatabaseURL string
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret      string
	ExpireHours int
}

// CORSConfig CORS 配置
type CORSConfig struct {
	Origins []string
}

// LogConfig 日志配置
type LogConfig struct {
	Level string
	File  string
}

// App 全局应用配置
var App *Config

// Load 加载配置
func Load() error {
	// 加载 .env 文件（如果存在）
	if err := godotenv.Load(); err != nil {
		// .env 文件不存在不算错误，可以使用环境变量
		fmt.Println("警告: .env 文件未找到，使用环境变量")
	}

	App = &Config{
		Server: ServerConfig{
			Port: getEnvInt("SERVER_PORT", 8080),
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Mode: getEnv("SERVER_MODE", "debug"),
		},
		Supabase: SupabaseConfig{
			URL:         getEnv("SUPABASE_URL", ""),
			AnonKey:     getEnv("SUPABASE_ANON_KEY", ""),
			ServiceKey:  getEnv("SUPABASE_SERVICE_KEY", ""),
			DatabaseURL: getEnv("DB_DSN", getEnv("DATABASE_URL", "")),
		},
		JWT: JWTConfig{
			Secret:      getEnv("JWT_SECRET", "default-secret-please-change"),
			ExpireHours: getEnvInt("JWT_EXPIRE_HOURS", 720),
		},
		CORS: CORSConfig{
			Origins: strings.Split(getEnv("CORS_ORIGINS", "*"), ","),
		},
		Log: LogConfig{
			Level: getEnv("LOG_LEVEL", "debug"),
			File:  getEnv("LOG_FILE", ""),
		},
		Ollama: OllamaConfig{
			URL:            getEnv("OLLAMA_URL", "http://localhost:11434"),
			EmbeddingModel: getEnv("OLLAMA_EMBEDDING_MODEL", "bge-m3"),
			ChatModel:      getEnv("OLLAMA_CHAT_MODEL", "deepseek-r1:7b"),
			StoreDir:       getEnv("KNOWLEDGE_STORE_DIR", "./data/knowledge"),
		},
	}

	// 校验必填配置
	if App.Supabase.URL == "" || App.Supabase.URL == "https://your-project.supabase.co" {
		return fmt.Errorf("SUPABASE_URL 未配置，请在 .env 文件中设置")
	}
	if App.Supabase.AnonKey == "" || App.Supabase.AnonKey == "your-anon-key" {
		return fmt.Errorf("SUPABASE_ANON_KEY 未配置，请在 .env 文件中设置")
	}
	if App.Supabase.ServiceKey == "" {
		fmt.Println("警告: SUPABASE_SERVICE_KEY 未配置，部分功能可能不可用")
	}

	return nil
}

// IsConfigured 检查 Supabase 是否已配置
func IsConfigured() bool {
	return App != nil &&
		App.Supabase.URL != "" &&
		App.Supabase.URL != "https://your-project.supabase.co" &&
		App.Supabase.AnonKey != "" &&
		App.Supabase.AnonKey != "your-anon-key"
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
