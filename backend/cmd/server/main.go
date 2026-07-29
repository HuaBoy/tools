package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/lib/pq"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tester-platform/backend/internal/config"
	"github.com/tester-platform/backend/internal/router"
	"github.com/tester-platform/backend/internal/supabase"
	"github.com/tester-platform/backend/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		fmt.Println("警告: .env 文件未找到，使用环境变量")
	}

	// 初始化日志
	setupLogger()

	// 加载配置
	if err := config.Load(); err != nil {
		log.Fatal().Err(err).Msg("配置加载失败")
	}

	log.Info().
		Str("mode", config.App.Server.Mode).
		Int("port", config.App.Server.Port).
		Bool("supabase_configured", config.IsConfigured()).
		Msg("启动 tester-platform backend")

	// 初始化 Supabase 客户端
	client := supabase.NewClient()

	// 初始化 JWT 管理器
	jwtManager := jwt.NewManager(
		config.App.JWT.Secret,
		config.App.JWT.ExpireHours,
	)

	// 初始化数据库连接（直连 Postgres，用于 sys_user 登录）
	var db *sql.DB
	if dsn := config.App.Supabase.DatabaseURL; dsn != "" {
		var err error
		db, err = sql.Open("postgres", dsn)
		if err != nil {
			log.Fatal().Err(err).Msg("数据库连接失败")
		}
		defer db.Close()
		log.Info().Msg("数据库连接成功")

		// 启动时自动初始化（建表 + 默认数据 + 重置密码）
		if err := runMigrations(db); err != nil {
			log.Warn().Err(err).Msg("数据库迁移失败")
		}
		if err := ensureAdminPassword(db); err != nil {
			log.Warn().Err(err).Msg("初始化 admin 密码失败")
		}
	} else {
		log.Fatal().Msg("DB_DSN / DATABASE_URL 未配置")
	}

	// 设置路由
	r := router.Setup(client, jwtManager, db)

	// 启动服务
	addr := fmt.Sprintf("%s:%d", config.App.Server.Host, config.App.Server.Port)
	log.Info().Str("addr", addr).Msg("服务监听中...")
	if err := r.Run(addr); err != nil {
		log.Fatal().Err(err).Msg("服务启动失败")
	}
}

// runMigrations 自动建表（幂等）
func runMigrations(db *sql.DB) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS public.operation_manuals (
			id SERIAL PRIMARY KEY,
			title TEXT NOT NULL DEFAULT '',
			category TEXT NOT NULL DEFAULT '',
			description TEXT NOT NULL DEFAULT '',
			file_url TEXT NOT NULL DEFAULT '',
			file_size TEXT NOT NULL DEFAULT '',
			updated_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS public.operation_videos (
			id SERIAL PRIMARY KEY,
			title TEXT NOT NULL DEFAULT '',
			category TEXT NOT NULL DEFAULT '',
			description TEXT NOT NULL DEFAULT '',
			url TEXT NOT NULL DEFAULT '',
			duration TEXT NOT NULL DEFAULT '',
			views INTEGER DEFAULT 0,
			date DATE DEFAULT CURRENT_DATE
		)`,
	}
	for _, m := range migrations {
		if _, err := db.Exec(m); err != nil {
			return fmt.Errorf("迁移失败: %w", err)
		}
	}
	log.Info().Msg("数据库迁移完成 (operation_manuals / operation_videos)")
	return nil
}

func ensureAdminPassword(db *sql.DB) error {
	// 检查 admin 是否存在于 sys_user
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM sys_user WHERE username = 'admin'").Scan(&count)
	if err != nil {
		return fmt.Errorf("查询 admin 失败: %w", err)
	}
	if count == 0 {
		log.Warn().Msg("sys_user 中未找到 admin 用户，跳过密码重置")
		return nil
	}

	// 生成 800228 的 bcrypt 哈希
	hash, err := bcrypt.GenerateFromPassword([]byte("800228"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("生成密码哈希失败: %w", err)
	}

	// 更新密码
	_, err = db.Exec("UPDATE sys_user SET password = $1 WHERE username = 'admin'", string(hash))
	if err != nil {
		return fmt.Errorf("更新 admin 密码失败: %w", err)
	}
	log.Info().Msg("admin 密码已重置为 800228")
	return nil
}

func setupLogger() {
	// 设置日志级别
	level := zerolog.InfoLevel
	if lvl := os.Getenv("LOG_LEVEL"); lvl != "" {
		if l, err := zerolog.ParseLevel(lvl); err == nil {
			level = l
		}
	}
	zerolog.SetGlobalLevel(level)

	// 美化输出
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "2006-01-02 15:04:05"})

	// 写文件
	if logFile := os.Getenv("LOG_FILE"); logFile != "" {
		// 确保目录存在
		dir := filepath.Dir(logFile)
		if err := os.MkdirAll(dir, 0755); err == nil {
			f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
			if err == nil {
				// 同时输出到控制台和文件
				multi := zerolog.MultiLevelWriter(
					zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "2006-01-02 15:04:05"},
					f,
				)
				log.Logger = log.Output(multi)
			} else {
				log.Warn().Err(err).Msg("无法打开日志文件，仅输出到控制台")
			}
		}
	}
}
