// 数据服务 - 查询/追溯/导出
import { api } from './http'

export const dataService = {
  query(params) {
    return api.get('/api/v1/data/query', params)
  },
  trace(batchId) {
    return api.get('/api/v1/data/trace/' + batchId)
  },
  analyzeBatch(batchId) {
    return api.get('/api/v1/data/batch-analysis/' + batchId)
  },
  queryFactory(params) {
    return api.post('/api/blade-detonate/blastDeviceFactory/page', params)
  },
  exportData(params) {
    return api.get('/api/v1/data/export', { ...params, responseType: 'blob' })
  },
  getStats() {
    return api.get('/api/v1/data/stats')
  }
}