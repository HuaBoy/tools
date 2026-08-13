-- =====================================================
-- 知识库 + 问题库 + AI 对话 建表脚本 (pgvector)
-- 适用于: 内网 PostgreSQL (pgvector 插件)
-- 模型维度: bge-m3 = 1024 维
-- =====================================================

-- 1. 启用 pgvector 扩展（幂等）
CREATE EXTENSION IF NOT EXISTS vector;

-- 2. 文档表
CREATE TABLE IF NOT EXISTS public.knowledge_documents (
    id           SERIAL PRIMARY KEY,
    title        TEXT NOT NULL DEFAULT '',
    category     TEXT NOT NULL DEFAULT 'other',
    description  TEXT NOT NULL DEFAULT '',
    file_name    TEXT NOT NULL DEFAULT '',
    file_type    TEXT NOT NULL DEFAULT '',
    file_size    BIGINT NOT NULL DEFAULT 0,
    storage_path TEXT NOT NULL DEFAULT '',
    status       TEXT NOT NULL DEFAULT 'processing',  -- processing / ready / failed
    chunk_count  INT  NOT NULL DEFAULT 0,
    created_by   TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 3. 文档分块表（含向量）
CREATE TABLE IF NOT EXISTS public.knowledge_chunks (
    id           SERIAL PRIMARY KEY,
    document_id  INT NOT NULL REFERENCES public.knowledge_documents(id) ON DELETE CASCADE,
    chunk_index  INT NOT NULL DEFAULT 0,
    content      TEXT NOT NULL,
    embedding    vector(1024),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 向量索引 (HNSW, 余弦距离)
CREATE INDEX IF NOT EXISTS idx_knowledge_chunks_embedding
    ON public.knowledge_chunks USING hnsw (embedding vector_cosine_ops);

-- 4. 问题库表（FAQ）
CREATE TABLE IF NOT EXISTS public.knowledge_faqs (
    id         SERIAL PRIMARY KEY,
    question   TEXT NOT NULL,
    answer     TEXT NOT NULL,
    category   TEXT NOT NULL DEFAULT 'other',
    tags       TEXT NOT NULL DEFAULT '[]',          -- JSON 数组
    embedding  vector(1024),
    created_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_knowledge_faqs_embedding
    ON public.knowledge_faqs USING hnsw (embedding vector_cosine_ops);

-- 5. 对话记录表
CREATE TABLE IF NOT EXISTS public.knowledge_conversations (
    id         SERIAL PRIMARY KEY,
    title      TEXT NOT NULL DEFAULT '新对话',
    user_id    TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 6. 对话消息表
CREATE TABLE IF NOT EXISTS public.knowledge_messages (
    id              SERIAL PRIMARY KEY,
    conversation_id INT NOT NULL REFERENCES public.knowledge_conversations(id) ON DELETE CASCADE,
    role            TEXT NOT NULL,                  -- user / assistant
    content         TEXT NOT NULL,
    sources         TEXT NOT NULL DEFAULT '[]',     -- JSON 引用来源
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_knowledge_messages_conversation
    ON public.knowledge_messages(conversation_id);
