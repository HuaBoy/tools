import { api } from '@/services/http'

const BASE = '/api/v1/pipeline'

// 需求池
export const getRequirements = (params = {}) => {
  const q = { page: params.page || 1, size: params.pageSize || 10 }
  if (params.keyword) q.keyword = params.keyword
  if (params.status) q.status = params.status
  if (params.kanoCategory) q.kano_category = params.kanoCategory
  if (params.priority) q.priority = params.priority
  if (params.aarrr) q.aarrr = params.aarrr
  if (params.source) q.source = params.source
  return api.get(`${BASE}/requirements`, q).then((resp) => ({
    data: (resp.data && resp.data.list) || [],
    total: (resp.data && resp.data.total) || 0
  }))
}

export const getRequirement = (id) => api.get(`${BASE}/requirements/${id}`).then((resp) => resp.data)

export const createRequirement = (data) => api.post(`${BASE}/requirements`, data).then((resp) => resp.data)

export const updateRequirement = (id, data) => api.patch(`${BASE}/requirements/${id}`, data).then((resp) => resp.data)

export const deleteRequirement = (id) => api.delete(`${BASE}/requirements/${id}`).then((resp) => resp.data)

export const changeRequirementStatus = (id, status) =>
  api.patch(`${BASE}/requirements/${id}/status`, { status }).then((resp) => resp.data)

// 流水线看板
export const getPipelineBoard = () => api.get(`${BASE}/board`).then((resp) => resp.data)

// 环节
export const getStages = (id) => api.get(`${BASE}/requirements/${id}/stages`).then((resp) => resp.data || [])

// 交付物
export const getDeliverables = (id) => api.get(`${BASE}/requirements/${id}/deliverables`).then((resp) => resp.data || [])

export const createDeliverable = (id, data) => api.post(`${BASE}/requirements/${id}/deliverables`, data).then((resp) => resp.data)

export const deleteDeliverable = (id) => api.delete(`${BASE}/deliverables/${id}`).then((resp) => resp.data)

// 审核
export const getReviews = (id) => api.get(`${BASE}/requirements/${id}/reviews`).then((resp) => resp.data || [])

export const createReview = (id, data) => api.post(`${BASE}/requirements/${id}/reviews`, data).then((resp) => resp.data)
