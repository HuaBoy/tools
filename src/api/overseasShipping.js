import { api } from '@/services/http'

const BASE = '/api/v1/overseas-shipping'

export default {
  // 列表（分页 + 关键字搜索 + 产品类型/国家筛选）
  async list(params = {}) {
    const q = { current: params.page || 1, size: params.pageSize || 10 }
    if (params.keyword) q.keyword = params.keyword
    if (params.productType) q.product_type = params.productType
    if (params.country) q.country = params.country
    const resp = await api.get(BASE, q)
    return {
      data: (resp.data && resp.data.list) || [],
      total: (resp.data && resp.data.total) || 0
    }
  },

  // 单条详情
  async getById(id) {
    const resp = await api.get(`${BASE}/${id}`)
    return resp.data
  },

  // 新增
  async create(data) {
    const resp = await api.post(BASE, data)
    return resp.data
  },

  // 更新（后端使用 PATCH）
  async update(id, data) {
    const resp = await api.patch(`${BASE}/${id}`, data)
    return resp.data
  },

  // 删除
  async remove(id) {
    const resp = await api.delete(`${BASE}/${id}`)
    return resp.data
  },

  // 导出全部（不分页拉取，前端生成 CSV）
  async exportAll() {
    const resp = await api.get(BASE, { current: 1, size: 100000 })
    return (resp.data && resp.data.list) || []
  }
}
