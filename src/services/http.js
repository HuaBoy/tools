// 统一 HTTP 客户端
import axios from 'axios'
import { useAuthStore } from '@/stores/auth'

const http = axios.create({
  timeout: 30000,
  headers: { 'Content-Type': 'application/json' }
})

http.interceptors.request.use((config) => {
  const authStore = useAuthStore()
  if (authStore.token) {
    config.headers.Authorization = 'Bearer ' + authStore.token
  }
  config.headers['X-Platform'] = import.meta.env.VITE_PLATFORM || 'web'
  return config
}, (e) => Promise.reject(e))

let refreshing = false
let queue = []

http.interceptors.response.use(
  (res) => res.data,
  async (err) => {
    const { config, response } = err
    if (response && response.status === 401 && !config._retry) {
      if (refreshing) {
        return new Promise((resolve) => { queue.push({ resolve, config }) })
      }
      refreshing = true
      config._retry = true
      try {
        const authStore = useAuthStore()
        await authStore.refreshToken?.()
        queue.forEach((item) => {
          item.config.headers.Authorization = 'Bearer ' + authStore.token
          item.resolve(http(item.config))
        })
        queue = []
        return http(config)
      } catch {
        useAuthStore().logout()
        window.location.href = '/login'
      } finally {
        refreshing = false
      }
    }
    return Promise.reject(err)
  }
)

export const api = {
  get: (u, p, c) => http.get(u, { params: p, ...c }),
  post: (u, d, c) => http.post(u, d, c),
  put: (u, d, c) => http.put(u, d, c),
  patch: (u, d, c) => http.patch(u, d, c),
  del: (u, c) => http.delete(u, c),
  upload: (u, fd, onP) => http.post(u, fd, {
    headers: { 'Content-Type': 'multipart/form-data' },
    onUploadProgress: onP
  })
}

export default http