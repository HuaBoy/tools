import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useChatStore = defineStore('chat', () => {
  const messages = ref([
    {
      id: 1,
      type: 'ai',
      content: '你好，我是起爆业务AI助手，可以帮你分析日志、查询故障、翻译术语、追溯数据。'
    }
  ])

  function addMessage(type, content) {
    messages.value.push({
      id: Date.now(),
      type,
      content
    })
  }

  function clearMessages() {
    messages.value = [
      {
        id: 1,
        type: 'ai',
        content: '你好，我是起爆业务AI助手，可以帮你分析日志、查询故障、翻译术语、追溯数据。'
      }
    ]
  }

  return {
    messages,
    addMessage,
    clearMessages
  }
})
