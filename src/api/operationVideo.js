import { api } from '@/services/http'

const BASE = '/api/v1/operation-videos'

export function listOperationVideos(params = {}) {
  return api.get(BASE, { params })
}

export function getOperationVideo(id) {
  return api.get(`${BASE}/${id}`)
}

export function createOperationVideo(data) {
  return api.post(BASE, data)
}

export function updateOperationVideo(id, data) {
  return api.patch(`${BASE}/${id}`, data)
}

export function deleteOperationVideo(id) {
  return api.delete(`${BASE}/${id}`)
}
