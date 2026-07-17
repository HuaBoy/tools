// 工作流执行状态
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useWorkflowStore = defineStore('workflow', () => {
  const tasks = ref([])
  const running = ref(false)

  const activeTasks = computed(() => tasks.value.filter(t => t.status === 'running'))
  const completedTasks = computed(() => tasks.value.filter(t => t.status === 'done'))
  const failedTasks = computed(() => tasks.value.filter(t => t.status === 'failed'))

  function addTask(task) { tasks.value.unshift({ ...task, id: Date.now(), status: 'pending', createdAt: new Date() }) }
  function updateTask(id, updates) {
    const idx = tasks.value.findIndex(t => t.id === id)
    if (idx > -1) tasks.value[idx] = { ...tasks.value[idx], ...updates }
  }
  function removeTask(id) { tasks.value = tasks.value.filter(t => t.id !== id) }

  return { tasks, running, activeTasks, completedTasks, failedTasks, addTask, updateTask, removeTask }
})