// 推送管理本地存储（沿用盛景应用本地模式，无需后端）
const PUSH_KEY = 'shengjing_push_tasks'

const tasks = []
const listeners = []

const load = () => {
  try {
    const stored = localStorage.getItem(PUSH_KEY)
    if (stored) {
      const parsed = JSON.parse(stored)
      tasks.splice(0, tasks.length, ...parsed)
    }
  } catch (e) {
    console.error('加载推送记录失败:', e)
  }
}

const persist = () => {
  try {
    localStorage.setItem(PUSH_KEY, JSON.stringify(tasks))
  } catch (e) {
    console.error('保存推送记录失败:', e)
    throw new Error('保存失败，可能本地存储空间不足')
  }
}

const generateId = () => 'push-' + Date.now().toString(36) + Math.random().toString(36).substr(2, 5)

export const localPushStore = {
  // 订阅变更（用于列表实时刷新）
  subscribe(fn) {
    listeners.push(fn)
    return () => {
      const i = listeners.indexOf(fn)
      if (i > -1) listeners.splice(i, 1)
    }
  },
  _emit() { listeners.forEach((fn) => fn()) },

  getAll() {
    return [...tasks].sort((a, b) => new Date(b.created_at) - new Date(a.created_at))
  },

  getById(id) {
    return tasks.find((t) => t.id === id) || null
  },

  // 创建并模拟执行一次推送：逐台 SN 推送，记录每台状态
  // 真实环境下这里应替换为对每台设备调用后端推送 API
  async create(taskData) {
    const now = new Date().toISOString()
    const id = generateId()
    const sns = [...new Set(taskData.sns.map((s) => s.trim()).filter(Boolean))]
    const task = {
      id,
      appId: taskData.appId || null,
      appName: taskData.appName || '',
      apkFileName: taskData.apkFileName || '',
      sns,
      total: sns.length,
      success: 0,
      failed: 0,
      status: 'sending', // sending | done
      results: {}, // sn -> 'success' | 'failed'
      note: taskData.note || '',
      created_at: now
    }
    tasks.unshift(task)
    persist()

    // 模拟逐台推送（每 30ms 一台，带进度反馈）；真实对接时换成 await api.push(sn)
    for (const sn of sns) {
      await new Promise((r) => setTimeout(r, 30))
      const ok = true // 本地模拟全部成功；对接后端后按接口返回判断
      task.results[sn] = ok ? 'success' : 'failed'
      if (ok) task.success++
      else task.failed++
      task.status = 'sending'
      persist()
      this._emit()
    }
    task.status = 'done'
    persist()
    this._emit()
    return task
  },

  remove(id) {
    const i = tasks.findIndex((t) => t.id === id)
    if (i > -1) {
      tasks.splice(i, 1)
      persist()
      this._emit()
    }
  },

  clear() {
    tasks.splice(0, tasks.length)
    persist()
    this._emit()
  }
}

load()
