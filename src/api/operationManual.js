import { api } from '@/services/http'

const BASE = '/api/v1/operation-manuals'

export function listOperationManuals(params = {}) {
  return api.get(BASE, { params })
}

export function getOperationManual(id) {
  return api.get(`${BASE}/${id}`)
}

export function createOperationManual(data) {
  return api.post(BASE, data)
}

export function updateOperationManual(id, data) {
  return api.patch(`${BASE}/${id}`, data)
}

export function deleteOperationManual(id) {
  return api.delete(`${BASE}/${id}`)
}
