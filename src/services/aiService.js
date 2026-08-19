// AI 服务 - 统一封装 LLM 调用

const PROVIDERS = {
  deepseek: {
    chat: 'https://api.deepseek.com/v1/chat/completions',
    models: 'https://api.deepseek.com/v1/models',
    model: 'deepseek-chat'
  },
  hunyuan: {
    chat: 'https://api.hunyuan.cloud.tencent.com/v1/chat/completions',
    model: 'hunyuan-lite'
  }
}

let provider = localStorage.getItem('ai_provider') || 'deepseek'
let apiKey = localStorage.getItem('ai_api_key') || ''

const getKey = () => apiKey
const getProvider = () => PROVIDERS[provider]

export const aiService = {
  setApiKey(key) { apiKey = key; localStorage.setItem('ai_api_key', key) },
  getApiKey() { return apiKey },

  setProvider(p) { provider = p; localStorage.setItem('ai_provider', p) },
  currentProvider() { return provider },

  async validateKey(key) {
    try {
      const p = getProvider()
      const res = await fetch(p.models || p.chat, {
        headers: { Authorization: 'Bearer ' + (key || apiKey) }
      })
      return res.ok
    } catch { return false }
  },

  async chat(messages, opts = {}) {
    const p = getProvider()
    const key = opts.apiKey || apiKey
    if (!key) throw new Error('AI API Key 未配置')

    const res = await fetch(p.chat, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Authorization: 'Bearer ' + key },
      body: JSON.stringify({
        model: opts.model || p.model,
        messages,
        temperature: opts.temperature ?? 0.7,
        max_tokens: opts.maxTokens ?? 2048
      })
    })
    if (!res.ok) {
      const err = await res.json().catch(() => ({}))
      throw new Error(err.error?.message || 'AI 调用失败')
    }
    const data = await res.json()
    return { content: data.choices?.[0]?.message?.content || '', usage: data.usage }
  },

  async analyze(context, instruction, contextMessages = []) {
    const messages = [
      { role: 'system', content: '你是工业起爆系统AI运维专家。用简洁中文回答，给出可执行建议。' }
    ]
    // 注入最近对话上下文（最多8轮）
    if (contextMessages.length) {
      messages.push(...contextMessages.slice(-16))
    }
    messages.push({ role: 'user', content: '背景数据：' + String(context) + '\n\n任务：' + instruction })
    return this.chat(messages, { temperature: 0.3 })
  },

  async generateReport(template, dataSet) {
    return this.chat([
      { role: 'system', content: '你是专业运维报告生成助手。输出Markdown格式报告。' },
      { role: 'user', content: '模板：' + template + '\n\n数据：' + JSON.stringify(dataSet) + '\n\n请生成完整报告。' }
    ], { temperature: 0.5, maxTokens: 4096 })
  },

  async diagnose(batchId, logData, traceData) {
    return this.analyze(
      '批次号: ' + batchId + '\n日志: ' + JSON.stringify(logData) + '\n链路: ' + JSON.stringify(traceData),
      '请诊断该批次是否存在异常，列出问题点和建议措施。'
    )
  }
}