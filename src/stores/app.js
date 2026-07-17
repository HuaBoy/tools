// 全局应用状态
import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useAppStore = defineStore('app', () => {
  const platform = ref('web')
  const isOnline = ref(navigator.onLine)
  const sidebarCollapsed = ref(false)

  function setPlatform(p) { platform.value = p }
  function toggleSidebar() { sidebarCollapsed.value = !sidebarCollapsed.value }

  window.addEventListener('online', () => { isOnline.value = true })
  window.addEventListener('offline', () => { isOnline.value = false })

  return { platform, isOnline, sidebarCollapsed, setPlatform, toggleSidebar }
})