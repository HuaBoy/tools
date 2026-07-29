// 应用管理 + 推送管理接口（对接 go-server /api/apps、/api/pushes）
// 使用纯 fetch 实现（与 version.js 一致），避免引入 axios 依赖
const apiBase = '/api'

const request = async (url, options = {}) => {
  const token = localStorage.getItem('auth_token')
  const headers = {
    'Content-Type': 'application/json',
    ...options.headers
  }
  if (token) {
    headers['Authorization'] = 'Bearer ' + token
  }
  const response = await fetch(apiBase + url, { ...options, headers })
  const resp = await response.json().catch(() => ({}))
  // 令牌失效 / 未登录：清理本地凭证并跳转登录页
  if (response.status === 401 || (resp && resp.code === 401)) {
    localStorage.removeItem('auth_token')
    localStorage.removeItem('auth_user')
    localStorage.removeItem('auth_user_info')
    localStorage.removeItem('last_activity')
    if (window.location.pathname !== '/login') {
      window.location.href = '/login'
    }
    throw new Error(resp.message || '登录已失效，请重新登录')
  }
  if (resp && resp.code !== undefined && resp.code !== 0) {
    throw new Error(resp.message || '请求失败')
  }
  return resp
}

const buildQuery = (params) => {
  if (!params) return ''
  const clean = Object.fromEntries(
    Object.entries(params).filter(([, v]) => v !== undefined && v !== '')
  )
  const qs = new URLSearchParams(clean).toString()
  return qs ? '?' + qs : ''
}

// ===== 应用管理 =====
export const appStoreApi = {
  // 查询（支持 name 模糊搜索）
  list: (params) => request('/apps' + buildQuery(params)),
  // 详情
  get: (id) => request(`/apps/${id}`),
  // 新增
  create: (data) => request('/apps', { method: 'POST', body: JSON.stringify(data) }),
  // 编辑
  update: (id, data) => request(`/apps/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  // 删除
  remove: (id) => request(`/apps/${id}`, { method: 'DELETE' }),
  // 上传 APK / 图标文件（multipart），返回 { apk_url, apk_name, apk_size, icon_url }
  upload: async (apkFile, iconFile) => {
    const token = localStorage.getItem('auth_token')
    const form = new FormData()
    if (apkFile) form.append('file', apkFile)
    if (iconFile) form.append('icon', iconFile)
    const headers = {}
    if (token) headers['Authorization'] = 'Bearer ' + token
    const response = await fetch(apiBase + '/apps/upload', { method: 'POST', headers, body: form })
    const resp = await response.json().catch(() => ({}))
    if (response.status === 401 || (resp && resp.code === 401)) {
      localStorage.removeItem('auth_token')
      localStorage.removeItem('auth_user')
      localStorage.removeItem('auth_user_info')
      localStorage.removeItem('last_activity')
      if (window.location.pathname !== '/login') window.location.href = '/login'
      throw new Error(resp.message || '登录已失效，请重新登录')
    }
    if (!resp || resp.code !== 0) throw new Error((resp && resp.message) || '上传失败')
    return resp
  }
}

// ===== 推送管理 =====
export const pushApi = {
  // 查询推送任务列表
  list: (params) => request('/pushes' + buildQuery(params)),
  // 创建推送任务（data.target=specified 时需带 target_sn 逗号分隔）
  create: (data) => request('/pushes', { method: 'POST', body: JSON.stringify(data) }),
  // 推送（下发）
  send: (id) => request(`/pushes/${id}/send`, { method: 'POST' }),
  // 删除推送任务
  remove: (id) => request(`/pushes/${id}`, { method: 'DELETE' }),
  // 设备视角：按 SN 拉取该设备可见的版本列表
  deviceUpdates: (sn) => request(`/devices/${encodeURIComponent(sn)}/updates`)
}
