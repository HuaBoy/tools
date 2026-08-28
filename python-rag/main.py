"""
LlamaIndex RAG 微服务
提供文档解析、分块、向量化、检索能力
Go 后端通过 HTTP 调用本服务
"""
# Windows 修复：psycopg 异步不兼容 ProactorEventLoop，必须在最前面设置
import asyncio
import sys
if sys.platform == "win32":
    asyncio.set_event_loop_policy(asyncio.WindowsSelectorEventLoopPolicy())

import json
import time
import os
from typing import Optional
from contextlib import asynccontextmanager

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel

from config import DATABASE_URL, EMBED_MODEL, OLLAMA_BASE_URL, RAG_PORT
from rag_engine import RAGEngine


rag: Optional[RAGEngine] = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    global rag
    rag = RAGEngine(
        db_url=DATABASE_URL,
        ollama_url=OLLAMA_BASE_URL,
        embed_model=EMBED_MODEL,
    )
    await rag.init()
    yield
    await rag.close()


app = FastAPI(title="LlamaIndex RAG Service", lifespan=lifespan)


# ── 请求/响应模型 ──────────────────────────────────

class IndexRequest(BaseModel):
    document_id: int
    content: str

class SearchRequest(BaseModel):
    query: str
    top_k: int = 8
    min_score: float = 0.3

class FAQRequest(BaseModel):
    id: int
    question: str
    answer: str

class ReindexRequest(BaseModel):
    document_id: int
    content: str


# ── API 端点 ──────────────────────────────────────

@app.get("/health")
async def health():
    return {"status": "ok", "engine": "llamaindex"}


@app.post("/api/rag/index")
async def index_document(req: IndexRequest):
    """分块 → 向量化 → 存入 pgvector（替换 Go 原有的 ChunkText + 逐块 Embedding）"""
    try:
        chunk_count = await rag.index_document(req.document_id, req.content)
        return {"ok": True, "chunk_count": chunk_count}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/api/rag/search")
async def search(req: SearchRequest):
    """向量语义搜索（文档分块 + FAQ）"""
    try:
        results = await rag.search(req.query, req.top_k, req.min_score)
        return {"query": req.query, "results": results, "count": len(results)}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/api/rag/context")
async def build_context(req: SearchRequest):
    """搜索 + 构建 RAG 上下文"""
    try:
        context, results = await rag.build_context(req.query, req.top_k, req.min_score)
        return {"context": context, "results": results}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/api/rag/faq")
async def index_faq(req: FAQRequest):
    """FAQ 向量化并存入"""
    try:
        await rag.index_faq(req.id, req.question, req.answer)
        return {"ok": True}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@app.delete("/api/rag/document/{doc_id}")
async def delete_document(doc_id: int):
    """删除文档分块"""
    try:
        await rag.delete_document(doc_id)
        return {"ok": True}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@app.delete("/api/rag/faq/{faq_id}")
async def delete_faq(faq_id: int):
    """删除 FAQ"""
    try:
        await rag.delete_faq(faq_id)
        return {"ok": True}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="127.0.0.1", port=RAG_PORT)
