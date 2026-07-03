import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useLogsStore = defineStore('logs', () => {
  const operationLogs = ref([])

  function addLog(action, module, detail = '') {
    operationLogs.value.unshift({
      id: Date.now(),
      time: new Date().toLocaleString('zh-CN'),
      action,
      module,
      detail
    })
    if (operationLogs.value.length > 100) {
      operationLogs.value.pop()
    }
  }

  function clearLogs() {
    operationLogs.value = []
  }

  return {
    operationLogs,
    addLog,
    clearLogs
  }
})
