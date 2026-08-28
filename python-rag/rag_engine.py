"""
LlamaIndex RAG 引擎
核心：用 LlamaIndex 的 SentenceSplitter 做分块，Ollama 做向量化，pgvector 做检索
使用同步 psycopg（避免 Windows ProactorEventLoop 问题）
"""
import json
import threading
from typing import List, Dict, Any, Tuple, Optional

import httpx
import psycopg
from llama_index.core.text_splitter import SentenceSplitter

from config import EMBED_MODEL, OLLAMA_BASE_URL


class RAGEngine:
    def __init__(self, db_url: str, ollama_url: str, embed_model: str):
        self.db_url = db_url
        self.ollama_url = ollama_url
        self.embed_model = embed_model
        self._db: Optional[psycopg.Connection] = None
        self._lock = threading.Lock()  # 保护同步 DB 连接
        self._http: Optional[httpx.AsyncClient] = None
        # LlamaIndex SentenceSplitter：512 token/块，80 token 重叠
        self.splitter = SentenceSplitter(
            chunk_size=512,
            chunk_overlap=80,
            paragraph_separator="\n",
            secondary_chunking_regex=r"[。！？；\n]",
        )

    async def init(self):
        self._http = httpx.AsyncClient(timeout=60)
        self._db = psycopg.connect(self.db_url, autocommit=True)
        print(f"[RAGEngine] 已连接 PostgreSQL (autocommit) + Ollama ({self.ollama_url})")

    async def close(self):
        if self._http:
            await self._http.aclose()
        if self._db:
            self._db.close()

    # ── 向量化 ────────────────────────────────────

    async def _embed(self, text: str) -> List[float]:
        """调用 Ollama 获取文本 embedding"""
        resp = await self._http.post(
            f"{self.ollama_url}/api/embeddings",
            json={"model": self.embed_model, "prompt": text},
        )
        resp.raise_for_status()
        return resp.json()["embedding"]

    async def _embed_batch(self, texts: List[str]) -> List[List[float]]:
        """批量向量化（逐条调用，Ollama 不支持原生 batch）"""
        vectors = []
        for t in texts:
            vectors.append(await self._embed(t))
        return vectors

    # ── pgvector 工具 ─────────────────────────────

    @staticmethod
    def _vec_literal(vec: List[float]) -> str:
        parts = ",".join(f"{v}" for v in vec)
        return f"[{parts}]"

    # ── 核心 API ──────────────────────────────────

    async def index_document(self, doc_id: int, content: str) -> int:
        """
        用 LlamaIndex SentenceSplitter 分块 → 向量化 → 存入 pgvector
        返回分块数量
        """
        # 1. LlamaIndex 分块
        nodes = self.splitter.split_text(content)
        chunks = [node for node in nodes if node.strip()]

        if not chunks:
            raise ValueError("文档内容为空，无法分块")

        # 2. 删除旧分块（如果是重新索引）
        with self._lock:
            try:
                cur = self._db.cursor()
                cur.execute("DELETE FROM public.knowledge_chunks WHERE document_id = %s", (doc_id,))
            except Exception as e:
                print(f"[RAGEngine] 删除旧分块失败: {e}")
                self._db.rollback()

        # 3. 批量向量化
        vectors = await self._embed_batch(chunks)

        # 4. 批量插入
        with self._lock:
            cur = self._db.cursor()
            for i, (chunk, vec) in enumerate(zip(chunks, vectors)):
                vec_lit = self._vec_literal(vec)
                cur.execute(
                    """INSERT INTO public.knowledge_chunks
                       (document_id, chunk_index, content, embedding)
                       VALUES (%s, %s, %s, %s::vector)""",
                    (doc_id, i, chunk, vec_lit),
                )

        # 5. 更新文档状态
        with self._lock:
            cur = self._db.cursor()
            cur.execute(
                """UPDATE public.knowledge_documents
                   SET status='ready', chunk_count=%s, updated_at=NOW()
                   WHERE id=%s""",
                (len(chunks), doc_id),
            )

        print(f"[RAGEngine] 文档 {doc_id} 已索引，{len(chunks)} 个分块")
        return len(chunks)

    async def search(self, query: str, top_k: int = 8, min_score: float = 0.3) -> List[Dict[str, Any]]:
        """向量语义搜索（文档分块 + FAQ）"""
        vec = await self._embed(query)
        vec_lit = self._vec_literal(vec)

        results = []

        with self._lock:
            cur = self._db.cursor()

            # 1. 检索文档分块
            cur.execute(
                """
                SELECT c.document_id, d.title AS document_name, d.category,
                       c.content, 1 - (c.embedding <=> %s::vector) AS score
                FROM public.knowledge_chunks c
                JOIN public.knowledge_documents d ON d.id = c.document_id
                WHERE d.status = 'ready'
                ORDER BY c.embedding <=> %s::vector
                LIMIT %s
                """,
                (vec_lit, vec_lit, top_k),
            )
            for row in cur.fetchall():
                r = {
                    "document_id": row[0],
                    "document_name": row[1],
                    "category": row[2],
                    "content": row[3],
                    "score": float(row[4]),
                    "source": "document",
                }
                if r["score"] >= min_score:
                    results.append(r)

            # 2. 检索 FAQ
            cur.execute(
                """
                SELECT id, question, answer, category,
                       1 - (embedding <=> %s::vector) AS score
                FROM public.knowledge_faqs
                ORDER BY embedding <=> %s::vector
                LIMIT %s
                """,
                (vec_lit, vec_lit, top_k),
            )
            for row in cur.fetchall():
                r = {
                    "faq_id": row[0],
                    "question": row[1],
                    "answer": row[2],
                    "category": row[3],
                    "score": float(row[4]),
                    "source": "faq",
                    "content": row[1] + "\n" + row[2],
                }
                if r["score"] >= min_score:
                    results.append(r)

        # 按分数降序
        results.sort(key=lambda x: x["score"], reverse=True)
        return results[:top_k]

    async def build_context(self, query: str, top_k: int = 6, min_score: float = 0.3) -> Tuple[str, List[Dict]]:
        """搜索 + 构建 RAG 上下文提示词"""
        results = await self.search(query, top_k, min_score)
        parts = []
        for i, r in enumerate(results):
            parts.append(f"[参考资料 {i+1}]")
            if r["source"] == "faq":
                parts.append(f"问题: {r['question']}\n答案: {r['answer']}")
            else:
                parts.append(f"来源文档: {r['document_name']}\n内容: {r['content']}")
            parts.append("")

        context = "\n".join(parts)
        return context, results

    async def index_faq(self, faq_id: int, question: str, answer: str):
        """FAQ 向量化并存入（仅更新 embedding）"""
        text = f"{question}\n{answer}"
        vec = await self._embed(text)
        vec_lit = self._vec_literal(vec)

        with self._lock:
            cur = self._db.cursor()
            cur.execute(
                "UPDATE public.knowledge_faqs SET embedding = %s::vector WHERE id = %s",
                (vec_lit, faq_id),
            )

    async def delete_document(self, doc_id: int):
        """删除文档分块"""
        with self._lock:
            cur = self._db.cursor()
            cur.execute("DELETE FROM public.knowledge_chunks WHERE document_id = %s", (doc_id,))

    async def delete_faq(self, faq_id: int):
        """删除 FAQ"""
        with self._lock:
            cur = self._db.cursor()
            cur.execute("DELETE FROM public.knowledge_faqs WHERE id = %s", (faq_id,))
