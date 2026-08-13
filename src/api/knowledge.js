import { api } from '@/services/http'

const BASE = '/api/v1/knowledge'

// ===== 统计与模型状态 =====
export async function getKnowledgeStats() {
  const resp = await api.get(`${BASE}/stats`)
  return resp.data || {}
}

// ===== 文档管理 =====
export async function listDocuments(params = {}) {
  const q = { current: params.page || 1, size: params.pageSize || 20 }
  if (params.category) q.category = params.category
  const resp = await api.get(`${BASE}/documents`, q)
  return {
    data: (resp.data && resp.data.list) || [],
    total: (resp.data && resp.data.total) || 0
  }
}

export async function getDocument(id) {
  const resp = await api.get(`${BASE}/documents/${id}`)
  return resp.data
}

export async function uploadDocument(file, meta = {}) {
  const fd = new FormData()
  fd.append('file', file)
  if (meta.title) fd.append('title', meta.title)
  if (meta.category) fd.append('category', meta.category)
  if (meta.description) fd.append('description', meta.description)
  const resp = await api.upload(`${BASE}/documents/upload`, fd)
  return resp.data
}

export async function deleteDocument(id) {
  const resp = await api.del(`${BASE}/documents/${id}`)
  return resp.data
}

// ===== 语义搜索 =====
export async function searchKnowledge(query, params = {}) {
  const q = { q: query }
  if (params.topK) q.top_k = params.topK
  if (params.minScore) q.min_score = params.minScore
  const resp = await api.get(`${BASE}/search`, q)
  return {
    query: resp.data?.query,
    results: resp.data?.results || [],
    count: resp.data?.count || 0
  }
}

// ===== 问题库 FAQ =====
export async function listFAQs(params = {}) {
  const q = { current: params.page || 1, size: params.pageSize || 20 }
  if (params.category) q.category = params.category
  const resp = await api.get(`${BASE}/faqs`, q)
  return {
    data: (resp.data && resp.data.list) || [],
    total: (resp.data && resp.data.total) || 0
  }
}

export async function createFAQ(data) {
  const resp = await api.post(`${BASE}/faqs`, data)
  return resp.data
}

export async function deleteFAQ(id) {
  const resp = await api.del(`${BASE}/faqs/${id}`)
  return resp.data
}

// ===== AI 对话（非流式） =====
export async function sendChat(question, conversationId = 0) {
  const resp = await api.post(`${BASE}/chat`, {
    question,
    conversation_id: conversationId,
    stream: false
  }, { timeout: 600000 })
  return resp.data
}

// ===== 会话管理 =====
export async function listConversations() {
  const resp = await api.get(`${BASE}/conversations`)
  return (resp.data || [])
}

export async function getConversationMessages(id) {
  const resp = await api.get(`${BASE}/conversations/${id}/messages`)
  return (resp.data || [])
}

export default {
  getKnowledgeStats,
  listDocuments,
  getDocument,
  uploadDocument,
  deleteDocument,
  searchKnowledge,
  listFAQs,
  createFAQ,
  deleteFAQ,
  sendChat,
  listConversations,
  getConversationMessages
}
