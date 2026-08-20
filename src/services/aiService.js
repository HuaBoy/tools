// AI 服务 - 统一封装 LLM 调用
// 默认使用本地 Ollama 部署，敏感数据不出内网

const PROVIDERS = {
  ollama: {
    name: '本地 Ollama',
    chat: 'http://localhost:11434/v1/chat/completions',
    models: 'http://localhost:11434/api/tags',
    model: 'deepseek-r1:7b',
    needKey: false,
    badge: 'Ollama 本地'
  },
  deepseek: {
    name: 'DeepSeek 云端',
    chat: 'https://api.deepseek.com/v1/chat/completions',
    models: 'https://api.deepseek.com/v1/models',
    model: 'deepseek-chat',
    needKey: true,
    badge: 'DeepSeek'
  },
  hunyuan: {
    name: '腾讯混元',
    chat: 'https://api.hunyuan.cloud.tencent.com/v1/chat/completions',
    model: 'hunyuan-lite',
    needKey: true,
    badge: '混元'
  }
}

let provider = localStorage.getItem('ai_provider') || 'ollama'
let apiKey = localStorage.getItem('ai_api_key') || ''
let ollamaUrl = localStorage.getItem('ai_ollama_url') || 'http://localhost:11434'

const getKey = () => apiKey

const getProvider = () => {
  const p = PROVIDERS[provider]
  if (provider === 'ollama' && p) {
    return { ...p, chat: ollamaUrl + '/v1/chat/completions', models: ollamaUrl + '/api/tags' }
  }
  return p
}

export const aiService = {
  setApiKey(key) { apiKey = key; localStorage.setItem('ai_api_key', key) },
  getApiKey() { return apiKey },

  setProvider(p) {
    provider = p
    localStorage.setItem('ai_provider', p)
    // 切换到 ollama 时无需 key，切到云端时需检查 key
  },
  currentProvider() { return PROVIDERS[provider] },
  currentProviderKey() { return provider },
  providerBadge() { return PROVIDERS[provider]?.badge || provider },

  setOllamaUrl(url) {
    ollamaUrl = url.replace(/\/+$/, '')
    localStorage.setItem('ai_ollama_url', ollamaUrl)
  },
  getOllamaUrl() { return ollamaUrl },

  async validateKey(key) {
    try {
      const p = getProvider()
      const headers = {}
      if (p.needKey) headers.Authorization = 'Bearer ' + (key || apiKey)
      const res = await fetch(p.models, { headers })
      return res.ok
    } catch { return false }
  },

  // 测试连接
  async testConnection() {
    try {
      const p = getProvider()
      const headers = { 'Content-Type': 'application/json' }
      if (p.needKey) headers.Authorization = 'Bearer ' + apiKey

      const res = await fetch(p.chat, {
        method: 'POST',
        headers,
        body: JSON.stringify({
          model: p.model,
          messages: [{ role: 'user', content: 'ping' }],
          max_tokens: 5
        })
      })
      return { ok: res.ok, status: res.status }
    } catch (e) {
      return { ok: false, error: e.message }
    }
  },

  async chat(messages, opts = {}) {
    const p = getProvider()
    const key = opts.apiKey || apiKey
    if (p.needKey && !key) {
      if (provider === 'ollama') {
        throw new Error('本地 Ollama 服务不可用，请确认 Ollama 已启动')
      }
      throw new Error('AI API Key 未配置')
    }

    const headers = { 'Content-Type': 'application/json' }
    if (p.needKey) headers.Authorization = 'Bearer ' + key

    const res = await fetch(p.chat, {
      method: 'POST',
      headers,
      body: JSON.stringify({
        model: opts.model || p.model,
        messages,
        temperature: opts.temperature ?? 0.7,
        max_tokens: opts.maxTokens ?? 2048
      })
    })
    if (!res.ok) {
      const err = await res.json().catch(() => ({}))
      if (provider === 'ollama') {
        throw new Error('本地 Ollama 服务连接失败: ' + (err.error?.message || res.statusText))
      }
      throw new Error(err.error?.message || 'AI 调用失败')
    }
    const data = await res.json()
    return { content: data.choices?.[0]?.message?.content || '', usage: data.usage }
  },

  async analyze(context, instruction, contextMessages = []) {
    const messages = [
      { role: 'system', content: '你是工业起爆系统AI运维专家。用简洁中文回答，给出可执行建议。' }
    ]
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