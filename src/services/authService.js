/**
 * 认证服务
 * 统一登录/注册/登出/Token 管理，兼容本地账号、Supabase、三方平台
 */
import { api } from './http'
import supabase from '@/utils/supabase'
import { localUserStore } from '@/utils/localUserStore'

// ======== 登录 ========
export async function login(username, password) {
  // 优先 Supabase
  if (supabase?.auth) {
    const { data, error } = await supabase.auth.signInWithPassword({
      email: username.includes('@') ? username : `${username}@ai-blasting.local`,
      password
    })
    if (error) throw error
    return { user: data.user, session: data.session, token: data.session?.access_token }
  }
  // 降级本地账号
  const result = localUserStore.validateLogin(username, password)
  if (!result.success) throw new Error(result.message)
  return { user: result.user, token: result.token || 'local-token', local: true }
}

// ======== 注册 ========
export async function register(username, password, email) {
  if (supabase?.auth) {
    const { data, error } = await supabase.auth.signUp({ email, password })
    if (error) throw error
    return data
  }
  const result = localUserStore.createUser({ username, password, email })
  if (!result.success) throw new Error(result.message)
  return result
}

// ======== 登出 ========
export async function logout() {
  if (supabase?.auth) {
    await supabase.auth.signOut()
  }
  localUserStore.clearSession()
}

// ======== 获取当前用户信息 ========
export async function getProfile() {
  try {
    return await api.get('/api/v1/auth/profile')
  } catch {
    // 降级：从 localStorage 读取
    const saved = localStorage.getItem('auth_user')
    return saved ? JSON.parse(saved) : null
  }
}

// ======== 刷新 Token ========
export async function refreshToken() {
  if (supabase?.auth) {
    const { data } = await supabase.auth.refreshSession()
    return data.session?.access_token
  }
  return null
}

// ======== 修改密码 ========
export async function changePassword(oldPwd, newPwd) {
  return api.post('/api/v1/auth/change-password', { oldPassword: oldPwd, newPassword: newPwd })
}

/**
 * 三方平台登录（云系统 / 智能制造 / 工厂数据）
 * @param {'mp'|'smart'|'iot'} platform 平台标识
 * @param {Object} credentials 登录凭证
 */
export async function thirdPartyLogin(platform, credentials) {
  return api.post('/api/v1/auth/third-party-login', { platform, ...credentials })
}
