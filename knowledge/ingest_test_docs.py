#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
测试知识库入库脚本（自包含，不依赖 RAG 微服务是否在线）
==================================================
功能：
  1. 读取 test-docs/ 下的 *.md 文档，写入 knowledge_documents 并分块向量化入库。
  2. 写入一组内置 FAQ 到 knowledge_faqs（供客服页「快捷提问」与检索使用）。
  3. 可重复执行：已存在的同名文档会先清旧分块再重索引；已存在的 FAQ 自动跳过。

运行环境：能同时访问 PostgreSQL(pgvector) 与 Ollama 的机器
          （通常是部署 python-rag 的同一台服务器）。

用法：
  pip install psycopg httpx
  # 按需覆盖环境变量
  set DATABASE_URL=postgres://postgres:密码@127.0.0.1:54322/postgres?sslmode=disable
  set OLLAMA_BASE_URL=http://127.0.0.1:11434
  set EMBED_MODEL=bge-m3
  python ingest_test_docs.py
"""
import os
import re
import glob
import sys

import psycopg
import httpx

# ── 配置（可用环境变量覆盖）──────────────────────────
DATABASE_URL = os.getenv("DATABASE_URL", os.getenv("DB_DSN", ""))
OLLAMA_BASE_URL = os.getenv("OLLAMA_BASE_URL", "http://127.0.0.1:11434")
EMBED_MODEL = os.getenv("EMBED_MODEL", "bge-m3")
CATEGORY = "test"  # 测试文档归类，便于在知识库管理里识别/清理

HERE = os.path.dirname(os.path.abspath(__file__))
DOCS_DIR = os.path.join(HERE, "test-docs")

# 内置 FAQ（客服快捷提问来源）
FAQ_LIST = [
    {
        "question": "起爆器开机后停在自检界面，提示通信模块异常怎么办？",
        "answer": "先重启设备；若仍异常，检查通信天线是否连接牢固、物联网卡是否有效、所在区域信号覆盖。如为现场临时无信号，可切换至离线作业模式（需提前在系统内报备）。",
        "category": "起爆器",
    },
    {
        "question": "起爆器电量消耗很快，是否正常？",
        "answer": "低温环境下电池放电加快属正常现象；若常温也异常耗电，建议校准电池或联系售后检测电源模块。作业前请保持电量不低于 60%。",
        "category": "起爆器",
    },
    {
        "question": "手持采集器扫不出雷管 UID 怎么办？",
        "answer": "检查采集器电量与读码距离（建议贴近 5cm 内）；确认雷管未被提前激活；尝试重新对码或重启采集器。",
        "category": "雷管录入",
    },
    {
        "question": "组网在线数量比录入数量少是什么原因？",
        "answer": "常见原因：雷管未全部布置、通信被遮挡、UID 重复录入。请在现场逐一排查后重新组网，确认数量一致再进入起爆流程。",
        "category": "雷管录入",
    },
    {
        "question": "延期时间校验不通过怎么处理？",
        "answer": "检查是否存在段别重复、相邻段差过小、总延时超过设备上限。按爆破设计修正后重新校验。",
        "category": "雷管录入",
    },
    {
        "question": "起爆记录在哪里查看和导出？",
        "answer": "在系统的「起爆记录 / 追溯」模块可查看每次起爆的成功与失败清单，并支持按时间、区域导出报表，用于合规追溯。",
        "category": "数据服务",
    },
    {
        "question": "知识库里的资料怎么更新？",
        "answer": "管理员可在「知识库管理」上传最新操作手册、规范文档，系统会自动向量化并建立检索；常见问题可在「智能知识库(RAG)」的 FAQ 中维护。",
        "category": "数据服务",
    },
    {
        "question": "智能客服答不上来的问题怎么办？",
        "answer": "本智能客服以本地知识库为准，仅回答资料覆盖的内容。如涉及未收录的专业问题，请转人工客服或查阅最新版官方规程。",
        "category": "数据服务",
    },
]


def log(*a):
    print("[ingest]", *a, flush=True)


def embed(text: str) -> list:
    r = httpx.post(
        f"{OLLAMA_BASE_URL}/api/embeddings",
        json={"model": EMBED_MODEL, "prompt": text},
        timeout=120,
    )
    r.raise_for_status()
    return r.json()["embedding"]


def vec_literal(vec: list) -> str:
    return "[" + ",".join(f"{v}" for v in vec) + "]"


def chunk_text(text: str, size: int = 512, overlap: int = 80) -> list:
    """按中文句末标点切分，控制每块长度与重叠。"""
    text = re.sub(r"\n{2,}", "\n", text).strip()
    parts = re.split(r"(?<=[。！？；\n])", text)
    chunks, cur = [], ""
    for p in parts:
        if cur and len(cur) + len(p) > size:
            chunks.append(cur.strip())
            cur = cur[-overlap:] + p
        else:
            cur += p
    if cur.strip():
        chunks.append(cur.strip())
    return [c for c in chunks if c.strip()]


def upsert_document(conn, title: str, content: str) -> int:
    with conn.cursor() as cur:
        cur.execute(
            "SELECT id FROM public.knowledge_documents WHERE title=%s AND category=%s",
            (title, CATEGORY),
        )
        row = cur.fetchone()
        if row:
            doc_id = row[0]
            cur.execute("DELETE FROM public.knowledge_chunks WHERE document_id=%s", (doc_id,))
            log(f"已存在文档「{title}」(id={doc_id})，清除旧分块后重索引")
        else:
            cur.execute(
                """INSERT INTO public.knowledge_documents
                   (title, category, description, file_name, file_type, file_size, status, created_by)
                   VALUES (%s,%s,'测试文档自动入库',%s,'.md',%s,'processing','seed')
                   RETURNING id""",
                (title, CATEGORY, title + ".md", len(content.encode("utf-8"))),
            )
            doc_id = cur.fetchone()[0]
            log(f"新建文档「{title}」(id={doc_id})")
    conn.commit()

    chunks = chunk_text(content)
    vecs = [embed(c) for c in chunks]
    with conn.cursor() as cur:
        for i, (c, v) in enumerate(zip(chunks, vecs)):
            cur.execute(
                """INSERT INTO public.knowledge_chunks (document_id, chunk_index, content, embedding)
                   VALUES (%s,%s,%s,%s::vector)""",
                (doc_id, i, c, vec_literal(v)),
            )
        cur.execute(
            """UPDATE public.knowledge_documents
               SET status='ready', chunk_count=%s, updated_at=NOW() WHERE id=%s""",
            (len(chunks), doc_id),
        )
    conn.commit()
    log(f"  └─ 索引完成：{len(chunks)} 个分块")
    return doc_id


def upsert_faq(conn, q: str, a: str, cat: str):
    with conn.cursor() as cur:
        cur.execute("SELECT id FROM public.knowledge_faqs WHERE question=%s", (q,))
        if cur.fetchone():
            log(f"FAQ 已存在，跳过：{q[:24]}…")
            return
        vec = vec_literal(embed(f"{q}\n{a}"))
        cur.execute(
            """INSERT INTO public.knowledge_faqs
               (question, answer, category, tags, embedding, created_by)
               VALUES (%s,%s,%s,'["test"]'::jsonb,%s::vector,'seed')""",
            (q, a, cat, vec),
        )
    conn.commit()
    log(f"FAQ 已写入：{q[:24]}…")


def main():
    if not DATABASE_URL:
        log("错误：未设置 DATABASE_URL / DB_DSN 环境变量")
        sys.exit(1)
    log(f"连接数据库 …")
    conn = psycopg.connect(DATABASE_URL, autocommit=False)
    try:
        # 1) 文档
        md_files = sorted(glob.glob(os.path.join(DOCS_DIR, "*.md")))
        if not md_files:
            log(f"未在 {DOCS_DIR} 找到 .md 文档")
        for fp in md_files:
            title = os.path.splitext(os.path.basename(fp))[0]
            with open(fp, "r", encoding="utf-8") as f:
                content = f.read()
            log(f"处理文档：{title}")
            upsert_document(conn, title, content)

        # 2) FAQ
        log("写入 FAQ …")
        for f in FAQ_LIST:
            upsert_faq(conn, f["question"], f["answer"], f["category"])

        log("✅ 测试知识库入库完成")
    finally:
        conn.close()


if __name__ == "__main__":
    main()
