import { reactive } from 'vue'
import { deleteFile } from '@/utils/indexedDB'

const APPS_KEY = 'shengjing_apps'
const CUSTOM_DEVICES_KEY = 'shengjing_devices_custom'

// 预置适配设备
export const DEFAULT_DEVICES = [
  'DT40',
  '小勇士',
  '煤许一代',
  '煤许二代',
  '全面屏',
  '联想瑞'
]

const apps = reactive({})

const loadApps = () => {
  try {
    const stored = localStorage.getItem(APPS_KEY)
    if (stored) {
      const parsed = JSON.parse(stored)
      Object.keys(parsed).forEach((id) => { apps[id] = parsed[id] })
    }
  } catch (e) {
    console.error('加载盛景应用数据失败:', e)
  }
}

const saveApps = () => {
  try {
    localStorage.setItem(APPS_KEY, JSON.stringify(apps))
  } catch (e) {
    console.error('保存盛景应用数据失败:', e)
    throw new Error('保存失败，可能本地存储空间不足')
  }
}

const generateId = () => 'app-' + Date.now().toString(36) + Math.random().toString(36).substr(2, 5)

// ===== 自定义设备 =====
const loadCustomDevices = () => {
  try {
    const stored = localStorage.getItem(CUSTOM_DEVICES_KEY)
    return stored ? JSON.parse(stored) : []
  } catch {
    return []
  }
}

const saveCustomDevices = (list) => {
  localStorage.setItem(CUSTOM_DEVICES_KEY, JSON.stringify(list))
}

export const localAppStore = {
  // ===== 应用 CRUD =====
  getAll() {
    return Object.values(apps).sort((a, b) => new Date(b.created_at) - new Date(a.created_at))
  },

  getById(id) {
    return apps[id] || null
  },

  create(appData) {
    const now = new Date().toISOString()
    const id = generateId()
    const app = {
      id,
      name: appData.name,
      intro: appData.intro || '',
      manualText: appData.manualText || '',
      manualFile: appData.manualFile || null,
      apkFile: appData.apkFile || null,
      devices: appData.devices || [],
      created_at: now,
      updated_at: now
    }
    apps[id] = app
    saveApps()
    return app
  },

  update(id, updates) {
    if (!apps[id]) throw new Error('应用不存在')
    Object.assign(apps[id], updates, { updated_at: new Date().toISOString() })
    saveApps()
    return apps[id]
  },

  async remove(id) {
    if (!apps[id]) throw new Error('应用不存在')
    const app = apps[id]
    // 清理关联的文件
    if (app.apkFile?.id) {
      try { await deleteFile(app.apkFile.id) } catch (e) { console.warn('删除APK失败', e) }
    }
    if (app.manualFile?.id) {
      try { await deleteFile(app.manualFile.id) } catch (e) { console.warn('删除手册失败', e) }
    }
    delete apps[id]
    saveApps()
    return true
  },

  // ===== 设备 =====
  getDevices() {
    const custom = loadCustomDevices()
    return [...DEFAULT_DEVICES, ...custom]
  },

  addDevice(deviceName) {
    const name = (deviceName || '').trim()
    if (!name) return
    const list = loadCustomDevices()
    if (DEFAULT_DEVICES.includes(name) || list.includes(name)) return
    list.push(name)
    saveCustomDevices(list)
  }
}

loadApps()
