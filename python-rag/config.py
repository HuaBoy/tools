"""RAG 服务配置"""
import os
from dotenv import load_dotenv

load_dotenv()

# Ollama 配置
OLLAMA_BASE_URL = os.getenv("OLLAMA_BASE_URL", "http://localhost:11434")
EMBED_MODEL = os.getenv("EMBED_MODEL", "bge-m3")
CHAT_MODEL = os.getenv("CHAT_MODEL", "deepseek-r1:7b")

# PostgreSQL 配置（使用 Supabase 的 PostgreSQL）
DATABASE_URL = os.getenv("DATABASE_URL", os.getenv("DB_DSN", ""))

# 服务配置
RAG_PORT = int(os.getenv("RAG_PORT", "8001"))
