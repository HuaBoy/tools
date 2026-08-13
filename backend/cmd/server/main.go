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
		// ===== 用户登录表（新部署时初始化，兼容 sys_user 登录方案） =====
		`CREATE TABLE IF NOT EXISTS public.sys_user (
			id SERIAL PRIMARY KEY,
			username TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL DEFAULT '',
			email TEXT NOT NULL DEFAULT '',
			role TEXT NOT NULL DEFAULT 'user',
			permissions TEXT NOT NULL DEFAULT '[]',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`INSERT INTO public.sys_user (username, password, email, role, permissions)
		 SELECT 'admin', '', 'admin@local', 'admin', '["*"]'
		 WHERE NOT EXISTS (SELECT 1 FROM public.sys_user WHERE username = 'admin')`,
		// ===== 知识库 (pgvector) =====
		`CREATE EXTENSION IF NOT EXISTS vector`,
		`CREATE TABLE IF NOT EXISTS public.knowledge_documents (
			id SERIAL PRIMARY KEY,
			title TEXT NOT NULL DEFAULT '',
			category TEXT NOT NULL DEFAULT 'other',
			description TEXT NOT NULL DEFAULT '',
			file_name TEXT NOT NULL DEFAULT '',
			file_type TEXT NOT NULL DEFAULT '',
			file_size BIGINT NOT NULL DEFAULT 0,
			storage_path TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'processing',
			chunk_count INT NOT NULL DEFAULT 0,
			created_by TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS public.knowledge_chunks (
			id SERIAL PRIMARY KEY,
			document_id INT NOT NULL REFERENCES public.knowledge_documents(id) ON DELETE CASCADE,
			chunk_index INT NOT NULL DEFAULT 0,
			content TEXT NOT NULL,
			embedding vector(1024),
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_knowledge_chunks_embedding
			ON public.knowledge_chunks USING hnsw (embedding vector_cosine_ops)`,
		`CREATE TABLE IF NOT EXISTS public.knowledge_faqs (
			id SERIAL PRIMARY KEY,
			question TEXT NOT NULL,
			answer TEXT NOT NULL,
			category TEXT NOT NULL DEFAULT 'other',
			tags TEXT NOT NULL DEFAULT '[]',
			embedding vector(1024),
			created_by TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_knowledge_faqs_embedding
			ON public.knowledge_faqs USING hnsw (embedding vector_cosine_ops)`,
		`CREATE TABLE IF NOT EXISTS public.knowledge_conversations (
			id SERIAL PRIMARY KEY,
			title TEXT NOT NULL DEFAULT '新对话',
			user_id TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS public.knowledge_messages (
			id SERIAL PRIMARY KEY,
			conversation_id INT NOT NULL REFERENCES public.knowledge_conversations(id) ON DELETE CASCADE,
			role TEXT NOT NULL,
			content TEXT NOT NULL,
			sources TEXT NOT NULL DEFAULT '[]',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_knowledge_messages_conversation
			ON public.knowledge_messages(conversation_id)`,
		// ===== AI 研发流水线（REQ-2026-001）=====
		`CREATE TABLE IF NOT EXISTS public.ai_pipeline_requirements (
			id SERIAL PRIMARY KEY,
			req_no TEXT NOT NULL UNIQUE,
			title TEXT NOT NULL DEFAULT '',
			description TEXT NOT NULL DEFAULT '',
			kano_category TEXT NOT NULL DEFAULT 'O',
			aarrr_impacts TEXT NOT NULL DEFAULT '[]',
			priority TEXT NOT NULL DEFAULT 'P2',
			status TEXT NOT NULL DEFAULT 'draft',
			source TEXT NOT NULL DEFAULT '',
			created_by INT NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS public.ai_pipeline_stages (
			id SERIAL PRIMARY KEY,
			req_id INT NOT NULL REFERENCES public.ai_pipeline_requirements(id) ON DELETE CASCADE,
			stage TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'todo',
			started_at TIMESTAMPTZ,
			finished_at TIMESTAMPTZ,
			note TEXT NOT NULL DEFAULT ''
		)`,
		`CREATE INDEX IF NOT EXISTS idx_ai_pipeline_stages_req
			ON public.ai_pipeline_stages(req_id)`,
		`CREATE TABLE IF NOT EXISTS public.ai_pipeline_deliverables (
			id SERIAL PRIMARY KEY,
			req_id INT NOT NULL REFERENCES public.ai_pipeline_requirements(id) ON DELETE CASCADE,
			stage TEXT NOT NULL DEFAULT '',
			title TEXT NOT NULL DEFAULT '',
			file_type TEXT NOT NULL DEFAULT '',
			file_path TEXT NOT NULL DEFAULT '',
			url TEXT NOT NULL DEFAULT '',
			created_by TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS public.ai_pipeline_reviews (
			id SERIAL PRIMARY KEY,
			req_id INT NOT NULL REFERENCES public.ai_pipeline_requirements(id) ON DELETE CASCADE,
			action TEXT NOT NULL DEFAULT '',
			target_stage TEXT NOT NULL DEFAULT '',
			reason TEXT NOT NULL DEFAULT '',
			reviewed_by INT NOT NULL DEFAULT 0,
			reviewed_name TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		// 流水线权限码（供功能管理界面按需分配给角色）
		`INSERT INTO public.features (code, name, description, category, enabled, "order")
		 SELECT v.code, v.name, v.description, v.category, true, v.ord FROM (VALUES
			('pipeline:view', '流水线查看', '查看 AI 研发流水线（需求池/看板/交付物）', 'pipeline', 90),
			('pipeline:edit', '流水线编辑', '编辑需求、维护交付物', 'pipeline', 91),
			('pipeline:review', '流水线审核', '需求审核通过/打回', 'pipeline', 92)
		 ) AS v(code, name, description, category, ord)
		 WHERE NOT EXISTS (SELECT 1 FROM public.features WHERE code = v.code)`,
	}
	for _, m := range migrations {
		if _, err := db.Exec(m); err != nil {
			return fmt.Errorf("迁移失败: %w", err)
		}
	}
	log.Info().Msg("数据库迁移完成 (operation_manuals / operation_videos / knowledge / ai_pipeline)")
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
