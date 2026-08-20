<script setup>
import { ref, computed, onMounted, nextTick, watch } from 'vue'
import { useRouter } from 'vue-router'
import { aiService } from '@/services/aiService'
import { useLocalDiagnosis } from '@/composables/useLocalDiagnosis'
import { getTenantName } from '@/utils/tenant.js'
import { showLoginDialog } from '@/utils/platformLogin.js'

const router = useRouter()

// ===== AI 对话 =====
const inputText = ref('')
const inputRef = ref(null)
const aiThinking = ref(false)
const chatExpanded = ref(false)
const chatMessages = ref([]) // { role: 'user'|'assistant', content: string, time: Date }
const chatContainerRef = ref(null)
const recentTasks = ref([])

// ===== 欢迎屏幕 & 引导交互 =====
const activeCapTab = ref('device')
const inputPlaceholder = ref('描述你的需求，如：查询SN编号为869850022329161的设备信息和爆破作业...')
const lastSuggestions = ref([])

const capabilityCategories = [
  { id: 'device', name: '设备查询', icon: '📟' },
  { id: 'blast', name: '爆破作业', icon: '💥' },
  { id: 'log', name: '日志分析', icon: '🔍' },
  { id: 'translate', name: '智能翻译', icon: '🌐' },
  { id: 'system', name: '系统导航', icon: '🧭' },
  { id: 'qa', name: '技术问答', icon: '💡' }
]

const guideMap = {
  device: [
    { icon: '📟', text: '查询设备详细信息', example: '查询设备SN编号 DZ600000016', tag: '常用' },
    { icon: '📋', text: '查询设备绑定关系', example: '查询手持机869850022329161的信息', tag: '' },
    { icon: '🔄', text: '批量设备状态查询', example: '查询SN编号为 DZ600000016 和 DZ600000017 的设备信息', tag: '批量' },
    { icon: '🏭', text: '按厂家筛选设备', example: '查询淮南舜泰的设备信息', tag: '' }
  ],
  blast: [
    { icon: '💥', text: '查询爆破作业记录', example: '查询手持机869850022329161的爆破作业', tag: '常用' },
    { icon: '📅', text: '按日期查询爆破记录', example: '查询最近7天的爆破作业', tag: '' },
    { icon: '🏗', text: '查询特定工程爆破记录', example: '查询XX工程的爆破作业', tag: '' },
    { icon: '📊', text: '爆破作业数据统计', example: '统计本月爆破作业次数', tag: '分析' }
  ],
  log: [
    { icon: '🔍', text: '分析日志异常', example: '分析以下日志片段：\n[ERROR] Connection timeout at 2026-08-01', tag: '常用' },
    { icon: '🐛', text: '诊断故障原因', example: '设备上报数据失败，日志显示ERR_TIMEOUT', tag: '' },
    { icon: '📈', text: '分析日志趋势', example: '分析以下日志中的错误模式', tag: '' }
  ],
  translate: [
    { icon: '🌐', text: '中英互译', example: '将以下内容翻译为英文：电子雷管起爆系统', tag: '常用' },
    { icon: '🔧', text: '行业术语翻译', example: '翻译以下爆破行业术语', tag: '专业' },
    { icon: '📝', text: '文档翻译', example: '翻译以下技术文档内容', tag: '' }
  ],
  system: [
    { icon: '🔐', text: '三方授权管理', example: '打开三方授权页面', tag: '导航', path: '/auth/third-party' },
    { icon: '🔓', text: '日志解密工具', example: '打开日志解密工具', tag: '导航', path: '/log/decrypt' },
    { icon: '📊', text: '数据查询工具', example: '打开数据查询页面', tag: '导航', path: '/data/query' },
    { icon: '📚', text: '知识库管理', example: '打开知识库', tag: '导航', path: '/knowledge/rag' },
    { icon: '🔗', text: '全链路追溯', example: '打开链路追溯', tag: '导航', path: '/data/trace' }
  ],
  qa: [
    { icon: '❓', text: '设备操作指南', example: '如何对电子雷管进行在线注册？', tag: '' },
    { icon: '📖', text: '技术原理讲解', example: '解释电子雷管延期起爆的原理', tag: '' },
    { icon: '⚠️', text: '故障排查指导', example: '手持机无法连接控制器，如何排查？', tag: '常用' },
    { icon: '🔧', text: '使用技巧', example: '爆破作业数据如何导出？', tag: '' }
  ]
}

const activeGuides = computed(() => guideMap[activeCapTab.value] || [])

// ===== 历史对话列表 =====
const chatHistory = ref([]) // { id, firstMessage, time }
const showHistoryPanel = ref(false)

// ===== 内联日期选择器（爆破作业查询） =====
const showDatePicker = ref(false)
const datePickerDeviceCode = ref('')
const datePickerQueryDevice = ref(false)
const datePickerQueryBlast = ref(false)
const datePickerOriginalText = ref('')
const dateRangeStart = ref('')
const dateRangeEnd = ref('')
const dateRangePreset = ref('week') // 'today' | 'week' | 'month' | 'threeMonth' | 'year' | 'custom'

function getPresetDateRange(preset) {
  const now = new Date()
  const end = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-${String(now.getDate()).padStart(2, '0')}`
  let start = ''
  switch (preset) {
    case 'today':
      start = end; break
    case 'week': {
      const d = new Date(); d.setDate(d.getDate() - 6)
      start = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`; break
    }
    case 'month': {
      const d = new Date(); d.setDate(d.getDate() - 29)
      start = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`; break
    }
    case 'threeMonth': {
      const d = new Date(); d.setMonth(d.getMonth() - 3)
      start = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`; break
    }
    case 'year': {
      const d = new Date(); d.setFullYear(d.getFullYear() - 1)
      start = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`; break
    }
    default:
      start = ''
  }
  dateRangeStart.value = start
  dateRangeEnd.value = end
  dateRangePreset.value = preset
}

function openDatePicker(deviceCode, queryDevice, queryBlast, originalText) {
  datePickerDeviceCode.value = deviceCode
  datePickerQueryDevice.value = queryDevice
  datePickerQueryBlast.value = queryBlast
  datePickerOriginalText.value = originalText
  getPresetDateRange('week')
  showDatePicker.value = true
}

function closeDatePicker() {
  showDatePicker.value = false
  datePickerDeviceCode.value = ''
}

function handleDateRangePreset(preset) {
  dateRangePreset.value = preset
  if (preset !== 'custom') getPresetDateRange(preset)
}

function executeBlastWithDateRange() {
  if (!dateRangeStart.value || !dateRangeEnd.value) return
  showDatePicker.value = false

  const intent = {
    deviceCode: datePickerDeviceCode.value,
    queryDevice: datePickerQueryDevice.value,
    queryBlast: datePickerQueryBlast.value
  }
  const dateInfo = `${dateRangeStart.value} 至 ${dateRangeEnd.value}`
  const deviceLabel = intent.deviceCode ? `设备 ${intent.deviceCode} ` : ''

  pushMessage('user', `📅 日期范围：${dateInfo}`)
  pushMessage('assistant', `🔎 正在查询${deviceLabel}在 ${dateInfo} 的爆破作业...`)

  aiThinking.value = true
  doBlastQuery(intent, dateRangeStart.value, dateRangeEnd.value).finally(() => {
    aiThinking.value = false
    showDatePicker.value = false
  })
}

async function doBlastQuery(intent, startDate, endDate) {
  try {
    let reply = ''
    if (intent.queryDevice && intent.deviceCode) {
      const deviceReply = await queryDeviceInfoInChat(intent.deviceCode)
      reply += deviceReply
    }
    if (intent.queryBlast) {
      if (reply) reply += '\n\n'
      const blastReply = await queryBlastTaskInChat(intent.deviceCode, startDate, endDate)
      reply += blastReply
    }
    chatMessages.value.pop()
    pushMessage('assistant', reply)
    recentTasks.value.unshift({ id: Date.now(), query: datePickerOriginalText.value, reply: reply.slice(0, 120), time: new Date() })
    if (recentTasks.value.length > 10) recentTasks.value.pop()
  } catch (e) {
    chatMessages.value.pop()
    pushMessage('assistant', '❌ 查询失败：' + (e.message || '请稍后重试'))
  }
}

// ===== 聊天记录持久化 =====
const CHAT_STORAGE_KEY = 'home_chatMessages'

function saveChatMessages() {
  try {
    const data = chatMessages.value.map(m => ({ role: m.role, content: m.content, time: m.time?.toISOString?.() || m.time }))
    localStorage.setItem(CHAT_STORAGE_KEY, JSON.stringify(data))
  } catch { /* ignore */ }
}

function loadChatMessages() {
  try {
    const saved = localStorage.getItem(CHAT_STORAGE_KEY)
    if (saved) {
      const data = JSON.parse(saved)
      chatMessages.value = data.map(m => ({ ...m, time: m.time ? new Date(m.time) : new Date() }))
    }
  } catch { /* ignore */ }
}

async function scrollToBottom() {
  await nextTick()
  const el = chatContainerRef.value
  if (el) el.scrollTop = el.scrollHeight
}

watch(chatMessages, () => {
  scrollToBottom()
  saveChatMessages()
}, { deep: true })

function pushMessage(role, content) {
  chatMessages.value.push({ role, content, time: new Date() })
}

// 格式化时间
function formatTime(date) {
  if (!date) return ''
  const d = date instanceof Date ? date : new Date(date)
  const pad = n => String(n).padStart(2, '0')
  return `${d.getMonth() + 1}/${d.getDate()} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

// 记录历史对话：每次对话开始时记录第一句话和时间
function recordChatHistory(firstMessage) {
  chatHistory.value.push({
    id: Date.now(),
    firstMessage: firstMessage.length > 40 ? firstMessage.slice(0, 40) + '...' : firstMessage,
    time: new Date()
  })
}

function clearChat() {
  chatMessages.value = []
  localStorage.removeItem(CHAT_STORAGE_KEY)
}

function loadHistoryItem(item) {
  // 点击历史对话条目，恢复对话（此处简单提示）
  pushMessage('assistant', `📋 历史对话（${formatTime(item.time)}）：${item.firstMessage}`)
}

// 快捷提示关键词
const quickHints = [
  { label: '设备查询', example: '查询设备SN编号 DZ600000016', icon: '📟' },
  { label: '爆破作业', example: '查询手持机869850022329161的爆破作业', icon: '💥' },
  { label: '设备+作业', example: '查询SN编号为869850022329161的设备信息和爆破作业', icon: '📋' },
  { label: '日志诊断', example: '分析以下日志片段', icon: '🔍' },
  { label: '翻译', example: '将以下内容翻译为英文', icon: '🌐' }
]

function applyHint(hint) {
  inputText.value = hint.example
  inputRef.value?.focus()
}

function applyGuide(guide) {
  if (guide.path) {
    router.push(guide.path)
    return
  }
  inputText.value = guide.example || guide.text
  inputRef.value?.focus()
}

function replayTask(task) {
  inputText.value = task.query
  inputRef.value?.focus()
}

// 根据最后一条AI回复生成引导建议
function generateSuggestions(reply) {
  const suggestions = []
  if (/设备|SN|手持机|控制器/.test(reply)) {
    if (/爆破/.test(reply)) suggestions.push('查询该设备的详细注册信息')
    suggestions.push('查询该设备的爆破作业记录')
    suggestions.push('查看该设备的固件版本信息')
  }
  if (/爆破|作业|工程/.test(reply)) {
    suggestions.push('查询更多时间段的爆破记录')
    suggestions.push('分析爆破作业数据趋势')
  }
  if (/错误|失败|异常|ERROR/.test(reply)) {
    suggestions.push('分析导致该错误的可能原因')
    suggestions.push('提供该问题的解决方案')
    suggestions.push('查看更多相关故障排查指南')
  }
  if (/翻译|translate/.test(reply)) {
    suggestions.push('将翻译结果复制到剪贴板')
    suggestions.push('翻译更多内容')
  }
  // 通用建议
  if (suggestions.length === 0) {
    suggestions.push('查询设备SN编号 DZ600000016')
    suggestions.push('查询爆破作业记录')
  }
  lastSuggestions.value = suggestions.slice(0, 4)
}

// ===== 本地诊断 =====
const { isAnalyzing: localThinking, results: localResults, logPatterns, issueType, suggestedTools, diagnose } = useLocalDiagnosis()
const localQuery = ref('')
const localResult = ref(null)

// 场景卡片
const roleScenarios = [
  { id: 'diagnose', icon: '🔍', title: '智能诊断', desc: '输入问题描述或日志，本地引擎自动匹配知识库' },
  { id: 'trace', icon: '🔗', title: '全链路追溯', desc: '芯片→模块→设备全链路追踪', path: '/data/trace' },
  { id: 'report', icon: '📋', title: '一键生成报告', desc: '选择批次，自动生成质量/市场分析报告' },
  { id: 'log', icon: '📄', title: '日志智能分析', desc: '上传日志，自动解密 + 模式匹配 + 问题归类', path: '/log/analysis' },
  { id: 'factory', icon: '🏭', title: '工厂数据查询', desc: '雷管厂数据统计、用量分析、产能追踪', path: '/trace/factory-data' },
  { id: 'translate', icon: '🌐', title: '行业术语翻译', desc: '8语种互译，含爆破行业专业术语库', path: '/tools/translate' }
]

// 统计数据
const stats = computed(() => [
  { label: '芯片在线', value: '--', icon: '💾' },
  { label: '模块在线', value: '--', icon: '📡' },
  { label: '设备在线', value: '--', icon: '🟢' },
  { label: '知识库', value: '18条', icon: '💡' }
])

// ===== AI 对话方法 =====
async function handleSend() {
  const text = inputText.value.trim()
  if (!text || aiThinking.value) return

  // 清除之前的引导建议
  lastSuggestions.value = []

  // 展开动画
  chatExpanded.value = true
  setTimeout(() => { chatExpanded.value = false }, 600)

  // 添加用户消息到聊天记录
  const isFirstMessage = chatMessages.value.length === 0
  pushMessage('user', text)
  inputText.value = ''

  // 首条消息时记录历史对话
  if (isFirstMessage) {
    recordChatHistory(text)
  }

  const cmd = parseQuickCmd(text)
  if (cmd) { pushMessage('assistant', `🔗 已为你跳转到对应功能页面。\n\n如需在对话中直接查询，请输入：查询SN编号为xxx的设备信息`); return }

  // 本地意图拆解：设备/手持机 SN 信息查询 + 爆破作业查询
  const deviceIntent = parseDeviceQueryIntent(text)
  if (deviceIntent) {
    // 如果只查设备信息（不含爆破作业），直接查
    if (deviceIntent.queryDevice && !deviceIntent.queryBlast) {
      aiThinking.value = true
      pushMessage('assistant', `🔎 正在检索设备 ${deviceIntent.deviceCode} 的设备信息...`)
      try {
        const reply = await queryDeviceInfoInChat(deviceIntent.deviceCode)
        chatMessages.value.pop()
        pushMessage('assistant', reply)
        generateSuggestions(reply)
        recentTasks.value.unshift({ id: Date.now(), query: text, reply: reply.slice(0, 120), time: new Date() })
        if (recentTasks.value.length > 10) recentTasks.value.pop()
      } finally {
        aiThinking.value = false
      }
      return
    }

    // 含爆破作业查询 → 先查设备信息（如果有 SN 且需要查设备），然后弹出日期选择器
    if (deviceIntent.queryDevice && deviceIntent.deviceCode) {
      aiThinking.value = true
      pushMessage('assistant', `🔎 正在检索设备 ${deviceIntent.deviceCode} 的设备信息...`)
      try {
        const deviceReply = await queryDeviceInfoInChat(deviceIntent.deviceCode)
        chatMessages.value.pop()
        pushMessage('assistant', deviceReply)
      } finally {
        aiThinking.value = false
      }
    }

    // 弹出内联日期选择器
    openDatePicker(deviceIntent.deviceCode, deviceIntent.queryDevice, deviceIntent.queryBlast, text)
    return
  }

  // 无法识别意图 → 显示关键词提示
  const hintResult = matchHintIntent(text)
  if (hintResult) {
    pushMessage('assistant', hintResult)
    generateSuggestions(hintResult)
    return
  }

  if (!aiService.getApiKey()) {
    const msg = '⚠️ AI 服务尚未配置。\n\n请前往「系统管理 > 三方账号授权」页面配置 DeepSeek API Key。\n\n💡 或使用上方快捷入口查询设备信息、爆破作业。\n\n如需直接查询设备信息，请输入：\n• 查询设备SN编号 DZ600000016\n• 查询手持机869850022329161的爆破作业\n• 查询SN编号为xxx的设备信息和爆破作业'
    pushMessage('assistant', msg)
    generateSuggestions(msg)
    return
  }

  aiThinking.value = true

  try {
    // 构建上下文：将最近对话历史作为上下文传给 AI
    const contextMessages = chatMessages.value.slice(-10).map(m => ({
      role: m.role,
      content: m.content
    }))
    const result = await aiService.analyze(text, '理解用户意图，如果是查询类请求，请告诉用户如何使用平台功能。如果是指令类，请给出执行步骤。请用简洁中文回答。', contextMessages)
    pushMessage('assistant', result.content)
    generateSuggestions(result.content)
    recentTasks.value.unshift({ id: Date.now(), query: text, reply: result.content.slice(0, 120), time: new Date() })
    if (recentTasks.value.length > 10) recentTasks.value.pop()
  } catch (e) {
    pushMessage('assistant', '❌ AI 调用失败：' + (e.message || '网络异常，请稍后重试'))
  } finally {
    aiThinking.value = false
  }
}

function parseQuickCmd(text) {
  const lower = text.toLowerCase()
  if (/诊断/.test(text)) return '/log/analysis'
  if (/日志/.test(text) && /解密/.test(text)) return '/log/decrypt'
  if (/数据/.test(text) && /查询/.test(text)) return '/data/query'
  if (/链路|追溯/.test(text)) return '/data/trace'
  if (/批次/.test(text) && /追溯|分析/.test(text)) return '/trace/analysis'
  if (/工厂|制造|一测/.test(text)) return '/trace/factory-data'
  if (/翻译/.test(text)) return '/tools/translate'
  if (/授权/.test(text)) return '/auth/third-party'
  if (/管理|后台/.test(text)) return '/admin'
  if (/知识|问题/.test(text)) return '/knowledge/rag'
  if (/手册/.test(text)) return '/knowledge/manual'
  if (/远程|手机/.test(text)) return '/tools/remote-phone'
  if (/二维码/.test(text)) return '/tools/qrcode'
  if (/转换|格式/.test(text)) return '/tools/converter'
  if (/助手/.test(text)) return '/tools/assistant'
  if (/应用|apk|app/i.test(text)) return '/appstore/shengjing'
  if (/推送|push|下发/i.test(text)) return '/appstore/push'
  if (/云系统/.test(text)) return '/tools/tester'
  return null
}

// 本地意图拆解：设备/手持机 SN 信息查询 + 爆破作业查询（不依赖 AI 服务 / API Key）
function parseDeviceQueryIntent(text) {
  const hasQueryIntent = /查询|检索|搜索|查找|查一下|看看|查/.test(text)
  const hasDeviceMark = /设备|手持机|产品|SN|编号|sn|机器/.test(text)
  const hasBlastMark = /爆破|作业|任务|工程/.test(text)
  const snMatch = text.match(/\b(DZ[a-zA-Z0-9-]+|\d{8,})\b/i) // 支持 DZ 开头或纯数字 SN（如 869850022329161）

  // 有 SN 的情况（原有逻辑）
  if (hasQueryIntent && (hasDeviceMark || hasBlastMark) && snMatch) {
    const deviceCode = snMatch[0].toUpperCase()
    const queryDevice = hasDeviceMark || /设备/.test(text)
    const queryBlast = hasBlastMark || /爆破|作业|任务/.test(text)
    return { deviceCode, queryDevice: queryDevice || !queryBlast, queryBlast: queryBlast || !queryDevice }
  }

  // 无 SN 但明确查询爆破作业 → 走爆破作业接口（deviceCode 为空）
  if (hasQueryIntent && hasBlastMark && !snMatch) {
    return { deviceCode: '', queryDevice: false, queryBlast: true }
  }

  return null
}

// 意图不匹配时，给出关键词提示（引导用户正确输入）
function matchHintIntent(text) {
  const snMatch = text.match(/\b(DZ[a-zA-Z0-9-]+|\d{8,})\b/i)
  // 包含 SN 但缺少明确查询意图 → 提示
  if (snMatch && !/查询|检索|搜索|查找|查/.test(text)) {
    const sn = snMatch[0]
    return `💡 检测到设备编号「${sn}」，但未明确查询意图。你可以这样输入：\n\n• 查询设备SN编号 ${sn}\n• 查询手持机 ${sn} 的爆破作业\n• 查询SN编号为 ${sn} 的设备信息和爆破作业\n\n点击下方关键词可快速填入 ⬇️`
  }
  return null
}

// ===== 爆破作业查询 =====
const BLAST_TASK_URL = '/api/blade-detonate/blastTask/page'

async function queryBlastTaskInChat(deviceCode, startDate, endDate) {
  let token = getDeviceQueryToken()
  if (!token) {
    const res = await showLoginDialog('mp')
    if (!res || !res.success) {
      return `⚠️ 查询爆破作业需要登录云系统。\n\n请在弹出的登录窗口中完成登录后重试。`
    }
    token = getDeviceQueryToken()
    if (!token) return `⚠️ 登录完成后仍未获取到凭证，请前往「三方账号授权」页面配置后重试。`
  }
  try {
    const params = new URLSearchParams()
    if (deviceCode) params.append('deviceCode', deviceCode)
    params.append('current', 1)
    params.append('size', 50)
    if (startDate) params.append('startDate', startDate)
    if (endDate) params.append('endDate', endDate)
    const response = await fetch(`${BLAST_TASK_URL}?${params.toString()}`, {
      method: 'GET',
      headers: {
        'accept': 'application/json, text/plain, */*',
        'authorization': 'Basic ' + btoa('saber:saber_secret'),
        'blade-auth': `bearer ${token}`,
        'tenant-id': '000000'
      }
    })
    const result = await response.json()
    if (result.code === 200 && result.data) {
      const records = result.data.records || []
      const total = result.data.total || 0
      const dateInfo = startDate && endDate ? `（${startDate} 至 ${endDate}）` : ''
      return formatBlastReply(deviceCode, records, total, dateInfo)
    }
    if (result.code === 401) {
      localStorage.removeItem('mp_token')
      localStorage.removeItem('mp_token_expire')
      return `⚠️ 云系统登录已过期，请重新登录后重试。`
    }
    return `❌ 爆破作业查询失败：${result.msg || result.message || '未知错误'}`
  } catch (e) {
    return '❌ 网络请求失败：' + (e.message || '请稍后重试')
  }
}

function formatBlastReply(deviceCode, records, total, dateInfo = '') {
  const fmt = (v) => (v === null || v === undefined || v === '') ? '-' : v
  const deviceLabel = deviceCode ? `设备 ${deviceCode} ` : '全部设备 '
  const header = `💥 ${deviceLabel}爆破作业记录${dateInfo}，共 ${total} 条`
  if (!records.length) {
    return `📭 未查询到${deviceLabel}的爆破作业记录。\n\n可能原因：\n• 该设备尚未执行过爆破作业\n• 爆破数据尚未上传`
  }
  const lines = records.map((r, i) => {
    return `\n──── 作业 ${i + 1} ────\n` +
      `🏗 工程名称：${fmt(r.taskName)}\n` +
      `🏢 使用单位：${fmt(r.deptName)}\n` +
      `🏭 雷管企业：${fmt(r.tenantName)}\n` +
      `👤 作业人员：${fmt(r.blasterUserName)}\n` +
      `📞 联系方式：${fmt(r.blasterUserPhone)}\n` +
      `🎛 控制器编号：${fmt(r.controllerCode)}\n` +
      `🔧 控制器版本：${fmt(r.controllerVersion)}\n` +
      `📟 手持机编号：${fmt(r.deviceCode)}\n` +
      `📱 手持机版本：${fmt(r.deviceVersion)}\n` +
      `💣 爆破数量：${fmt(r.detonatorCount)}\n` +
      `📅 爆破时间：${fmt(r.explosionDate)}\n` +
      `📤 上传时间：${fmt(r.uploadDlTime)}`
  })
  return `${header}\n${lines.join('\n')}`
}

// ===== 对话式设备信息查询（本地直查接口，无需跳转 / AI 服务） =====
const DEVICE_QUERY_URL = '/api/blade-detonate/blastDeviceFactory/page'

function getDeviceQueryToken() {
  let token = localStorage.getItem('mp_token')
  if (token) {
    const expire = Number(localStorage.getItem('mp_token_expire') || 0)
    if (!expire || expire > Date.now()) return token
  }
  try {
    const saved = localStorage.getItem('tester_credentials')
    if (saved) {
      const data = JSON.parse(saved)
      if (data.accessToken && data.expireTime > Date.now()) return data.accessToken
    }
  } catch (e) { /* ignore */ }
  return ''
}

async function queryDeviceInfoInChat(deviceCode) {
  let token = getDeviceQueryToken()
  if (!token) {
    const res = await showLoginDialog('mp')
    if (!res || !res.success) {
      return `⚠️ 查询设备信息需要登录云系统。\n\n请在弹出的登录窗口中完成登录后重试，或前往「三方账号授权」页面配置账号。`
    }
    token = getDeviceQueryToken()
    if (!token) return `⚠️ 登录完成后仍未获取到凭证，请前往「三方账号授权」页面配置后重试。`
  }
  try {
    const params = new URLSearchParams()
    params.append('deviceCode', deviceCode)
    params.append('current', 1)
    params.append('size', 10)
    const response = await fetch(`${DEVICE_QUERY_URL}?${params.toString()}`, {
      method: 'GET',
      headers: {
        'accept': 'application/json, text/plain, */*',
        'authorization': 'Basic ' + btoa('saber:saber_secret'),
        'blade-auth': `bearer ${token}`,
        'tenant-id': '000000'
      }
    })
    const result = await response.json()
    if (result.code === 200 && result.data) {
      const records = result.data.records || []
      const total = result.data.total || 0
      return formatDeviceReply(deviceCode, records, total)
    }
    if (result.code === 401) {
      localStorage.removeItem('mp_token')
      localStorage.removeItem('mp_token_expire')
      return `⚠️ 云系统登录已过期，请重新登录后重试。`
    }
    return `❌ 查询失败：${result.msg || result.message || '未知错误'}`
  } catch (e) {
    return '❌ 网络请求失败：' + (e.message || '请稍后重试')
  }
}

function formatDeviceReply(deviceCode, records, total) {
  const fmt = (v) => (v === null || v === undefined || v === '') ? '-' : v
  const header = `✅ 已查到设备 ${deviceCode} 的信息，共 ${total} 条记录`
  if (!records.length) {
    return `📭 未查询到设备 ${deviceCode} 的注册信息。\n\n可能原因：\n• 设备编号输入有误\n• 该设备尚未录入系统`
  }
  const lines = records.map((r, i) => {
    const tenantName = getTenantName(r.tenantId)
    return `\n──── 记录 ${i + 1} ────\n` +
      `🏭 管厂：${fmt(tenantName && tenantName !== r.tenantId ? `${r.tenantId} ${tenantName}` : r.tenantId)}\n` +
      `🏢 作业单位：${fmt(r.companyName)}\n` +
      `🎛 控制器编号：${fmt(r.controllerCode)}\n` +
      `🔧 控制器版本：${fmt(r.controllerVersion)}\n` +
      `📟 手持机编号：${fmt(r.deviceCode)}\n` +
      `📱 手持机版本：${fmt(r.deviceVersion)}\n` +
      `🔩 手持机类型：${fmt(r.deviceHardware)}\n` +
      `🧭 版本场景：${fmt(r.deviceScene)}`
  })
  return `${header}\n${lines.join('\n')}`
}

function fillScenario(scenario) {
  if (scenario.id === 'diagnose') {
    setTimeout(() => document.querySelector('.local-input')?.focus(), 100)
    return
  }
  if (scenario.path) { router.push(scenario.path); return }
  inputText.value = scenario.prompt || ''
  inputRef.value?.focus()
}

function handleKeydown(e) {
  if (e.key === 'Enter' && !e.shiftKey && !e.isComposing) {
    e.preventDefault()
    handleSend()
  }
}

// ===== 本地诊断方法 =====
async function runLocalDiagnosis() {
  const q = localQuery.value.trim()
  if (!q || localThinking.value) return
  localResult.value = await diagnose(q)
}

function handleLocalKeydown(e) {
  if (e.key === 'Enter' && !e.isComposing) {
    e.preventDefault()
    runLocalDiagnosis()
  }
}

// ===== 初始化 =====
onMounted(() => {
  setTimeout(() => inputRef.value?.focus(), 500)
  loadChatMessages()
  scrollToBottom()
  try {
    const saved = localStorage.getItem('ai_tasks')
    if (saved) recentTasks.value = JSON.parse(saved)
  } catch { }
})
</script>

<template>
  <div class="workbench">
    <!-- ===== 顶部横幅 ===== -->
    <header class="wb-header">
      <div class="hb-left">
        <div class="hb-icon">
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
            <polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/>
          </svg>
        </div>
        <div>
          <h1 class="hb-title">AI智能助手工作台</h1>
          <p class="hb-subtitle">智能对话 · 设备查询 · 爆破作业分析</p>
        </div>
      </div>
      <div class="hb-right">
        <!-- 移除统计芯片 -->
      </div>
    </header>

    <!-- ===== 主体：AI 对话全宽 ===== -->
    <div class="wb-body">
      <div class="wb-main ai-full">
        <!-- AI 对话 -->
        <div class="card ai-chat" :class="{ 'ai-chat-expanded': chatExpanded, 'ai-chat-loading': aiThinking }">
          <div class="card-head">
            <span class="ch-dot"></span>
            <span>AI 智能助手</span>
            <span class="ch-badge">DeepSeek</span>
            <span v-if="chatMessages.length" class="ch-msg-count">{{ chatMessages.length }} 条消息</span>
            <span v-if="chatHistory.length" class="ch-history-toggle" @click.stop="showHistoryPanel = !showHistoryPanel" :title="showHistoryPanel ? '收起历史' : '展开历史'">📋</span>
            <span v-if="chatMessages.length" class="ch-clear" @click.stop="clearChat" title="清空对话">🗑</span>
          </div>

          <!-- 历史对话列表 -->
          <div v-if="showHistoryPanel && chatHistory.length && !chatMessages.length" class="chat-history-panel">
            <div class="chp-title">📋 历史对话</div>
            <div v-for="item in [...chatHistory].reverse()" :key="item.id" class="chp-item" @click="loadHistoryItem(item)">
              <span class="chp-msg">{{ item.firstMessage }}</span>
              <span class="chp-time">{{ formatTime(item.time) }}</span>
            </div>
          </div>

          <!-- 欢迎屏幕（空对话且非历史面板时显示） -->
          <div v-if="!chatMessages.length && !aiThinking && !showHistoryPanel" class="ai-welcome">
            <div class="welcome-hero">
              <div class="welcome-icon-wrap">
                <div class="welcome-icon">
                  <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                    <polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/>
                  </svg>
                </div>
                <div class="welcome-pulse"></div>
              </div>
              <h2 class="welcome-title">你好，我是 AI 智能助手</h2>
              <p class="welcome-desc">我可以帮你查询设备信息、爆破作业记录、分析日志、解答技术问题。请选择一个场景开始，或直接输入你的需求。</p>
            </div>

            <!-- 能力分类标签 -->
            <div class="capability-tabs">
              <div v-for="cat in capabilityCategories" :key="cat.id"
                   class="cap-tab" :class="{ active: activeCapTab === cat.id }"
                   @click="activeCapTab = cat.id">
                <span class="cap-tab-icon">{{ cat.icon }}</span>
                <span class="cap-tab-name">{{ cat.name }}</span>
              </div>
            </div>

            <!-- 当前分类的引导卡片 -->
            <div class="guide-cards">
              <div v-for="g in activeGuides" :key="g.text" class="guide-card" @click="applyGuide(g)">
                <div class="gc-header">
                  <span class="gc-icon">{{ g.icon }}</span>
                  <span class="gc-tag" v-if="g.tag">{{ g.tag }}</span>
                </div>
                <div class="gc-text">{{ g.text }}</div>
                <div class="gc-example" v-if="g.example">示例：{{ g.example }}</div>
              </div>
            </div>

            <!-- 最近任务快捷入口 -->
            <div v-if="recentTasks.length" class="recent-quick">
              <div class="rq-title">📝 最近查询</div>
              <div class="rq-list">
                <div v-for="t in recentTasks.slice(0, 3)" :key="t.id" class="rq-item" @click="replayTask(t)">
                  <span class="rq-icon">🔄</span>
                  <span class="rq-text">{{ t.query }}</span>
                  <span class="rq-time">{{ formatTime(t.time) }}</span>
                </div>
              </div>
            </div>
          </div>

          <!-- 聊天记录区域 -->
          <div v-if="chatMessages.length" ref="chatContainerRef" class="chat-container">
            <div v-for="(msg, idx) in chatMessages" :key="idx" class="chat-message" :class="msg.role">
              <div class="chat-avatar">
                <template v-if="msg.role === 'user'">👤</template>
                <template v-else>
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/></svg>
                </template>
              </div>
              <div class="chat-bubble">
                <div class="chat-role">{{ msg.role === 'user' ? '我' : 'AI 助手' }}</div>
                <div class="chat-text">{{ msg.content }}</div>
              </div>
            </div>
            <!-- 思考中动画 -->
            <div v-if="aiThinking" class="chat-message assistant">
              <div class="chat-avatar">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/></svg>
              </div>
              <div class="chat-bubble thinking">
                <div class="chat-role">AI 助手</div>
                <div class="chat-text">
                  <span class="thinking-dots"><span></span><span></span><span></span></span>
                </div>
              </div>
            </div>

            <!-- 回复后的引导建议 -->
            <div v-if="!aiThinking && chatMessages.length && lastSuggestions.length" class="post-reply-suggestions">
              <div class="prs-label">💡 你可能还想了解：</div>
              <div class="prs-chips">
                <div v-for="(s, i) in lastSuggestions" :key="i" class="prs-chip" @click="applyHint({ example: s })">
                  {{ s }}
                </div>
              </div>
            </div>
          </div>

          <!-- 内联日期选择器（爆破作业查询） -->
          <div v-if="showDatePicker" class="date-picker-panel">
            <div class="dp-header">
              <span class="dp-title">📅 选择查询日期范围</span>
              <span class="dp-device" v-if="datePickerDeviceCode">设备：{{ datePickerDeviceCode }}</span>
              <span class="dp-close" @click="closeDatePicker" title="关闭">✕</span>
            </div>
            <div class="dp-presets">
              <button class="dp-preset" :class="{ active: dateRangePreset === 'today' }" @click="handleDateRangePreset('today')">今天</button>
              <button class="dp-preset" :class="{ active: dateRangePreset === 'week' }" @click="handleDateRangePreset('week')">近7天</button>
              <button class="dp-preset" :class="{ active: dateRangePreset === 'month' }" @click="handleDateRangePreset('month')">近30天</button>
              <button class="dp-preset" :class="{ active: dateRangePreset === 'threeMonth' }" @click="handleDateRangePreset('threeMonth')">近3个月</button>
              <button class="dp-preset" :class="{ active: dateRangePreset === 'year' }" @click="handleDateRangePreset('year')">近1年</button>
              <button class="dp-preset" :class="{ active: dateRangePreset === 'custom' }" @click="handleDateRangePreset('custom')">自定义</button>
            </div>
            <div class="dp-dates">
              <div class="dp-date-group">
                <label class="dp-label">开始日期</label>
                <input type="date" v-model="dateRangeStart" class="dp-input" :disabled="dateRangePreset !== 'custom'" />
              </div>
              <span class="dp-sep">至</span>
              <div class="dp-date-group">
                <label class="dp-label">结束日期</label>
                <input type="date" v-model="dateRangeEnd" class="dp-input" :disabled="dateRangePreset !== 'custom'" />
              </div>
            </div>
            <div class="dp-actions">
              <button class="dp-cancel" @click="closeDatePicker">取消</button>
              <button class="dp-confirm" :disabled="!dateRangeStart || !dateRangeEnd" @click="executeBlastWithDateRange">
                🔎 查询爆破作业
              </button>
            </div>
          </div>

          <div class="ai-input-area">
            <div class="ai-input-wrap">
              <textarea
                ref="inputRef"
                v-model="inputText"
                class="ai-textarea"
                :placeholder="inputPlaceholder"
                rows="2"
                @keydown="handleKeydown"
              ></textarea>
              <button class="ai-send" :class="{ loading: aiThinking }" :disabled="!inputText.trim() || aiThinking" @click="handleSend">
                <template v-if="aiThinking"><span class="spinner"></span></template>
                <template v-else>
                  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="22" y1="2" x2="11" y2="13"/><polygon points="22 2 15 22 11 13 2 9 22 2"/></svg>
                </template>
              </button>
            </div>
            <div class="ai-input-footer">
              <p class="ai-hint">Enter 发送 · Shift+Enter 换行 · 对话支持上下文记忆</p>
              <div class="ai-quick-actions">
                <span class="aqa-item" @click="applyHint({ example: '查询设备SN编号 DZ600000016' })">📟 设备查询</span>
                <span class="aqa-divider">|</span>
                <span class="aqa-item" @click="applyHint({ example: '查询手持机869850022329161的爆破作业' })">💥 爆破作业</span>
                <span class="aqa-divider">|</span>
                <span class="aqa-item" @click="applyHint({ example: '分析以下日志片段' })">🔍 日志分析</span>
                <span class="aqa-divider">|</span>
                <span class="aqa-item" @click="applyHint({ example: '将以下内容翻译为英文' })">🌐 翻译</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* ===== 整体 ===== */
.workbench {
  width: 100%;
  min-height: calc(100vh - 60px);
  padding: 20px 24px 32px;
  box-sizing: border-box;
  background: #F7F8FA;
}

/* ===== 顶部 ===== */
.wb-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 18px 24px;
  margin-bottom: 16px;
  background: #fff;
  border-radius: 14px;
  border: 1px solid #E8ECF1;
}
.hb-left { display: flex; align-items: center; gap: 14px; }
.hb-icon {
  width: 42px; height: 42px; border-radius: 12px;
  background: linear-gradient(135deg, #165DFF, #00D2AC);
  display: flex; align-items: center; justify-content: center;
  color: #fff;
}
.hb-title { font-size: 20px; font-weight: 700; color: #1E293B; margin: 0; line-height: 1.3; }
.hb-subtitle { font-size: 12px; color: #94A3B8; margin: 2px 0 0 0; }
.hb-right { display: flex; gap: 16px; }
.stat-chip {
  display: flex; align-items: center; gap: 6px;
  padding: 8px 14px; background: #F8FAFC; border-radius: 10px;
  border: 1px solid #E8ECF1;
}
.sc-icon { font-size: 14px; }
.sc-val { font-size: 15px; font-weight: 700; color: #1E293B; font-family: monospace; }
.sc-lbl { font-size: 11px; color: #94A3B8; }

/* ===== 主体布局 ===== */
.wb-body { display: flex; gap: 18px; align-items: flex-start; }
.wb-main { flex: 1; min-width: 0; }
.wb-main.ai-full { max-width: 900px; margin: 0 auto; }

/* ===== 卡片 ===== */
.card {
  background: #fff; border-radius: 14px; border: 1px solid #E8ECF1;
  padding: 18px 20px; margin-bottom: 16px;
}
.card-head {
  display: flex; align-items: center; gap: 10px;
  font-size: 14px; font-weight: 600; color: #1E293B; margin-bottom: 14px;
}
.ch-dot { width: 8px; height: 8px; border-radius: 50%; background: #165DFF; }
.ch-badge {
  font-size: 10px; padding: 2px 8px; border-radius: 20px;
  background: rgba(22,93,255,0.1); color: #165DFF; font-weight: 500;
}

/* ===== AI 对话 ===== */
.ai-input-area {
  display: flex; gap: 10px; align-items: flex-start;
}
.ai-textarea {
  flex: 1; border: 1px solid #E8ECF1; border-radius: 10px;
  padding: 12px 14px; font-size: 14px; color: #1E293B;
  resize: none; outline: none; font-family: inherit; line-height: 1.6;
  background: #F8FAFC; transition: border-color 0.2s;
}
.ai-textarea:focus { border-color: #165DFF; background: #fff; }
.ai-textarea::placeholder { color: #94A3B8; }
.ai-send {
  width: 42px; height: 42px; border-radius: 10px; border: none;
  background: linear-gradient(135deg, #165DFF, #0F4CD0); color: #fff;
  font-size: 20px; cursor: pointer; flex-shrink: 0; transition: all 0.2s;
  display: flex; align-items: center; justify-content: center; font-weight: 700;
}
.ai-send:hover:not(:disabled) { transform: scale(1.05); box-shadow: 0 4px 12px rgba(22,93,255,0.35); }
.ai-send:disabled { opacity: 0.35; cursor: not-allowed; }
.spinner {
  width: 16px; height: 16px; border: 2px solid rgba(255,255,255,0.3);
  border-top-color: #fff; border-radius: 50%; animation: spin 0.7s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg) } }
.ai-hint { font-size: 11px; color: #94A3B8; margin: 8px 0 0 0; }

/* AI 回复（兼容旧样式，已被 chat 替代） */
.ai-reply {
  margin-top: 16px; border: 1px solid rgba(22,93,255,0.12);
  border-radius: 12px; overflow: hidden;
}
.ar-head {
  display: flex; align-items: center; gap: 8px;
  padding: 10px 16px; background: rgba(22,93,255,0.04);
  font-size: 12px; color: #165DFF; font-weight: 600;
  border-bottom: 1px solid rgba(22,93,255,0.08);
}
.ar-body {
  padding: 16px; font-size: 13px; line-height: 1.8; color: #475569; white-space: pre-wrap;
}

/* ===== AI 对话卡片动画 ===== */
.ai-chat { transition: transform 0.3s cubic-bezier(0.34, 1.56, 0.64, 1), box-shadow 0.3s ease; }
.ai-chat-expanded {
  animation: aiPulse 0.6s cubic-bezier(0.34, 1.56, 0.64, 1);
}
.ai-chat-loading {
  box-shadow: 0 0 0 2px rgba(22,93,255,0.15), 0 4px 24px rgba(22,93,255,0.08);
}
@keyframes aiPulse {
  0% { transform: scale(1); }
  30% { transform: scale(1.02); box-shadow: 0 8px 32px rgba(22,93,255,0.15); }
  100% { transform: scale(1); }
}

/* 消息计数 & 清空按钮 */
.ch-msg-count {
  font-size: 10px; padding: 2px 8px; border-radius: 20px;
  background: rgba(0,180,42,0.08); color: #00B42A; font-weight: 500; margin-left: auto;
}
.ch-clear {
  font-size: 14px; cursor: pointer; opacity: 0.5; transition: opacity 0.2s; padding: 2px 4px;
}
.ch-clear:hover { opacity: 1; }

.ch-history-toggle {
  font-size: 14px; cursor: pointer; opacity: 0.6; transition: opacity 0.2s; padding: 2px 4px; margin-left: auto; margin-right: 4px;
}
.ch-history-toggle:hover { opacity: 1; }

/* ===== 历史对话列表 ===== */
.chat-history-panel { padding: 6px 0 4px; border-bottom: 1px solid rgba(255,255,255,0.06); max-height: 200px; overflow-y: auto; }
.chp-title { font-size: 12px; color: #94A3B8; margin-bottom: 6px; font-weight: 600; }
.chp-item {
  display: flex; justify-content: space-between; align-items: center;
  padding: 8px 10px; margin-bottom: 4px;
  border-radius: 8px; background: rgba(255,255,255,0.03);
  cursor: pointer; transition: background 0.15s;
}
.chp-item:hover { background: rgba(255,255,255,0.08); }
.chp-msg {
  flex: 1; font-size: 13px; color: #E2E8F0;
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap; margin-right: 8px;
}
.chp-time { font-size: 11px; color: #64748B; white-space: nowrap; }

/* ===== 快捷提示 ===== */
.ai-hints { padding: 8px 0 4px; }
.ai-hints-title { font-size: 12px; color: #94A3B8; margin-bottom: 10px; }
.ai-hints-grid { display: flex; flex-wrap: wrap; gap: 8px; }
.ai-hint-chip {
  display: inline-flex; align-items: center; gap: 4px;
  padding: 6px 12px; background: rgba(22,93,255,0.04); border: 1px solid rgba(22,93,255,0.12);
  border-radius: 20px; font-size: 12px; color: #165DFF; cursor: pointer;
  transition: all 0.15s;
}
.ai-hint-chip:hover { background: rgba(22,93,255,0.08); border-color: rgba(22,93,255,0.25); transform: translateY(-1px); }
.ahc-icon { font-size: 14px; }
.ahc-label { font-weight: 500; }

/* ===== 内联日期选择器 ===== */
.date-picker-panel {
  border: 1px solid rgba(22,93,255,0.15);
  border-radius: 12px; margin: 8px 0 12px;
  background: linear-gradient(135deg, rgba(22,93,255,0.03), rgba(0,210,172,0.03));
  animation: msgIn 0.3s ease-out;
  overflow: hidden;
}
.dp-header {
  display: flex; align-items: center; gap: 10px;
  padding: 12px 16px; background: rgba(22,93,255,0.06);
  border-bottom: 1px solid rgba(22,93,255,0.08);
}
.dp-title { font-size: 13px; font-weight: 600; color: #1E293B; }
.dp-device {
  font-size: 11px; padding: 2px 8px; border-radius: 4px;
  background: rgba(22,93,255,0.08); color: #165DFF; font-family: monospace;
}
.dp-close {
  margin-left: auto; font-size: 14px; color: #94A3B8;
  cursor: pointer; padding: 2px 6px; border-radius: 4px; transition: all 0.15s;
}
.dp-close:hover { color: #F53F3F; background: rgba(245,63,63,0.08); }

.dp-presets {
  display: flex; flex-wrap: wrap; gap: 6px;
  padding: 10px 16px;
}
.dp-preset {
  padding: 5px 14px; border-radius: 20px; border: 1px solid #E8ECF1;
  background: #fff; color: #475569; font-size: 12px; cursor: pointer;
  transition: all 0.15s; font-weight: 500;
}
.dp-preset:hover { border-color: #165DFF; color: #165DFF; background: rgba(22,93,255,0.04); }
.dp-preset.active {
  border-color: #165DFF; color: #fff; background: linear-gradient(135deg, #165DFF, #0F4CD0);
}

.dp-dates {
  display: flex; align-items: flex-end; gap: 10px;
  padding: 0 16px 10px;
}
.dp-date-group { flex: 1; }
.dp-label { display: block; font-size: 11px; color: #64748B; margin-bottom: 4px; font-weight: 500; }
.dp-input {
  width: 100%; box-sizing: border-box;
  border: 1px solid #E8ECF1; border-radius: 8px;
  padding: 8px 10px; font-size: 13px; color: #1E293B;
  background: #fff; outline: none; transition: border-color 0.2s;
}
.dp-input:focus { border-color: #165DFF; }
.dp-input:disabled { background: #F1F5F9; color: #94A3B8; cursor: not-allowed; }
.dp-sep { font-size: 13px; color: #94A3B8; padding-bottom: 8px; }

.dp-actions {
  display: flex; justify-content: flex-end; gap: 8px;
  padding: 10px 16px; border-top: 1px solid rgba(22,93,255,0.08);
}
.dp-cancel {
  padding: 7px 18px; border-radius: 8px; border: 1px solid #E8ECF1;
  background: #fff; color: #475569; font-size: 13px; cursor: pointer;
  transition: all 0.15s;
}
.dp-cancel:hover { border-color: #94A3B8; background: #F8FAFC; }
.dp-confirm {
  padding: 7px 20px; border-radius: 8px; border: none;
  background: linear-gradient(135deg, #165DFF, #0F4CD0); color: #fff;
  font-size: 13px; font-weight: 600; cursor: pointer;
  transition: all 0.15s;
}
.dp-confirm:hover:not(:disabled) { transform: translateY(-1px); box-shadow: 0 4px 12px rgba(22,93,255,0.3); }
.dp-confirm:disabled { opacity: 0.4; cursor: not-allowed; }

/* ===== 聊天记录 ===== */
.chat-container {
  max-height: 480px; overflow-y: auto; padding: 4px 0 8px;
  scroll-behavior: smooth;
}
.chat-container::-webkit-scrollbar { width: 4px; }
.chat-container::-webkit-scrollbar-track { background: transparent; }
.chat-container::-webkit-scrollbar-thumb { background: #D0D5DD; border-radius: 4px; }

.chat-message {
  display: flex; gap: 10px; padding: 8px 0; animation: msgIn 0.3s ease-out;
}
.chat-message.user { flex-direction: row-reverse; }

@keyframes msgIn {
  from { opacity: 0; transform: translateY(8px); }
  to { opacity: 1; transform: translateY(0); }
}

.chat-avatar {
  width: 32px; height: 32px; border-radius: 50%; flex-shrink: 0;
  display: flex; align-items: center; justify-content: center; font-size: 16px;
}
.chat-message.user .chat-avatar {
  background: linear-gradient(135deg, #165DFF, #0F4CD0); color: #fff;
}
.chat-message.assistant .chat-avatar {
  background: linear-gradient(135deg, #00D2AC, #008A6E); color: #fff;
}

.chat-bubble {
  max-width: 85%; padding: 10px 14px; border-radius: 12px; font-size: 13px; line-height: 1.7;
}
.chat-message.user .chat-bubble {
  background: linear-gradient(135deg, #165DFF, #0F4CD0); color: #fff;
  border-bottom-right-radius: 4px;
}
.chat-message.assistant .chat-bubble {
  background: #F1F5F9; color: #1E293B; border-bottom-left-radius: 4px;
}
.chat-bubble.thinking { min-width: 60px; }

.chat-role { font-size: 11px; font-weight: 600; margin-bottom: 4px; opacity: 0.7; }
.chat-text { white-space: pre-wrap; word-break: break-word; }

/* 思考中动画 */
.thinking-dots { display: inline-flex; gap: 4px; padding: 4px 0; }
.thinking-dots span {
  width: 6px; height: 6px; border-radius: 50%; background: #165DFF;
  animation: dotBounce 1.2s ease-in-out infinite;
}
.thinking-dots span:nth-child(2) { animation-delay: 0.2s; }
.thinking-dots span:nth-child(3) { animation-delay: 0.4s; }
@keyframes dotBounce {
  0%, 80%, 100% { transform: translateY(0); opacity: 0.4; }
  40% { transform: translateY(-6px); opacity: 1; }
}

/* ===== 场景卡片 ===== */
.sc-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 10px; }
.sc-card {
  display: flex; align-items: center; gap: 10px;
  padding: 14px; background: #F8FAFC; border: 1px solid #E8ECF1;
  border-radius: 10px; cursor: pointer; transition: all 0.15s;
}
.sc-card:hover { border-color: #165DFF; background: rgba(22,93,255,0.02); transform: translateY(-1px); }
.sc-icon { font-size: 24px; flex-shrink: 0; }
.sc-info { flex: 1; min-width: 0; }
.sc-name { font-size: 13px; font-weight: 600; color: #1E293B; margin-bottom: 2px; }
.sc-desc { font-size: 11px; color: #94A3B8; line-height: 1.3; }
.sc-card svg { color: #94A3B8; flex-shrink: 0; }

/* ===== 最近执行 ===== */
.recent-list { display: flex; flex-direction: column; gap: 6px; }
.recent-row {
  display: flex; justify-content: space-between; gap: 12px;
  padding: 10px 14px; background: #F8FAFC; border-radius: 8px;
  border: 1px solid #E8ECF1;
}
.rq { font-size: 13px; color: #1E293B; font-weight: 500; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.rp { font-size: 11px; color: #94A3B8; flex-shrink: 0; max-width: 40%; }

/* ===== 本地诊断 ===== */
.local-dx { border-color: rgba(0,210,172,0.3); }
.ld-input-wrap {
  display: flex; gap: 8px; margin-bottom: 8px;
}
.local-input {
  flex: 1; border: 1px solid #E8ECF1; border-radius: 10px;
  padding: 10px 12px; font-size: 13px; color: #1E293B;
  outline: none; background: #F8FAFC; transition: border-color 0.2s;
}
.local-input:focus { border-color: #00D2AC; background: #fff; }
.local-input::placeholder { color: #94A3B8; }
.ld-btn {
  width: 40px; height: 40px; border-radius: 10px; border: none;
  background: linear-gradient(135deg, #00D2AC, #008A6E); color: #fff;
  font-size: 16px; cursor: pointer; flex-shrink: 0; transition: all 0.2s;
}
.ld-btn:hover:not(:disabled) { transform: scale(1.05); }
.ld-btn:disabled { opacity: 0.4; cursor: not-allowed; }
.ld-hint { font-size: 10px; color: #94A3B8; margin: 0; }

/* 诊断结果 */
.ld-results { margin-top: 14px; }
.ld-section { margin-bottom: 12px; }
.ld-sec-title { font-size: 11px; font-weight: 600; color: #64748B; margin-bottom: 8px; text-transform: uppercase; letter-spacing: 0.3px; }
.ld-pattern {
  display: flex; align-items: flex-start; gap: 8px;
  padding: 8px 10px; border-radius: 6px; margin-bottom: 4px;
  font-size: 11px; font-family: monospace;
}
.ld-pattern.high { background: #FFF2F0; }
.ld-pattern.medium { background: #FFF7E6; }
.ld-pattern.low { background: #F6FFED; }
.ldp-dot { width: 6px; height: 6px; border-radius: 50%; margin-top: 4px; flex-shrink: 0; }
.ld-pattern.high .ldp-dot { background: #F53F3F; }
.ld-pattern.medium .ldp-dot { background: #FF7D00; }
.ld-pattern.low .ldp-dot { background: #00B42A; }
.ldp-text { color: #475569; word-break: break-all; }

.ld-match {
  padding: 10px; border-radius: 8px; margin-bottom: 6px;
  background: #F8FAFC; border: 1px solid #E8ECF1;
}
.ld-match.best { border-color: #00D2AC; background: rgba(0,210,172,0.03); }
.ldm-head { display: flex; align-items: center; gap: 6px; margin-bottom: 4px; }
.ldm-sev { font-size: 10px; padding: 1px 6px; border-radius: 4px; color: #fff; }
.ldm-cat { font-size: 10px; color: #94A3B8; }
.ldm-score { font-size: 10px; color: #00D2AC; font-weight: 600; margin-left: auto; }
.ldm-q { font-size: 12px; color: #1E293B; line-height: 1.4; }

.ld-tools { display: flex; gap: 6px; flex-wrap: wrap; }
.ld-tool-btn {
  padding: 4px 10px; border-radius: 6px; border: 1px solid rgba(22,93,255,0.2);
  background: rgba(22,93,255,0.04); color: #165DFF; font-size: 11px;
  cursor: pointer; transition: all 0.15s;
}
.ld-tool-btn:hover { background: rgba(22,93,255,0.1); }
.ld-empty { font-size: 12px; color: #94A3B8; text-align: center; padding: 12px 0; }

/* ===== 快捷工具 ===== */
.qt-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 6px; }
.qt-item {
  padding: 9px 12px; border-radius: 8px; cursor: pointer;
  font-size: 12px; color: #475569; background: #F8FAFC;
  border: 1px solid transparent; transition: all 0.15s;
}
.qt-item:hover { border-color: rgba(22,93,255,0.2); background: rgba(22,93,255,0.03); color: #165DFF; }

/* ===== 硬件卡片 ===== */
.hw-card { background: linear-gradient(135deg, rgba(22,93,255,0.03), rgba(0,210,172,0.02)); border-color: rgba(22,93,255,0.12); }
.hw-row { display: flex; flex-direction: column; gap: 8px; }
.hw-item {
  display: flex; align-items: center; justify-content: space-between;
  padding: 10px 14px; background: #fff; border-radius: 8px;
  border: 1px solid #E8ECF1; font-size: 13px; color: #475569;
}
.hw-tag {
  font-size: 10px; padding: 2px 8px; border-radius: 4px;
  background: rgba(22,93,255,0.08); color: #165DFF; font-weight: 500;
}

/* ===== 响应式 ===== */
/* ===== 欢迎屏幕 ===== */
.ai-welcome { padding: 10px 0; }
.welcome-hero { text-align: center; padding: 40px 20px 30px; }
.welcome-icon-wrap { position: relative; display: inline-block; margin-bottom: 20px; }
.welcome-icon {
  width: 80px; height: 80px; border-radius: 20px;
  background: linear-gradient(135deg, #7B68EE, #FF6B35);
  display: flex; align-items: center; justify-content: center;
  color: #fff; position: relative; z-index: 1;
}
.welcome-pulse {
  position: absolute; top: -6px; left: -6px; right: -6px; bottom: -6px;
  border-radius: 24px; background: linear-gradient(135deg, rgba(123,104,238,0.2), rgba(255,107,53,0.2));
  animation: pulse-glow 3s ease-in-out infinite;
}
@keyframes pulse-glow {
  0%, 100% { transform: scale(1); opacity: 0.5; }
  50% { transform: scale(1.05); opacity: 0.8; }
}
.welcome-title { font-size: 22px; font-weight: 700; color: #1D2129; margin: 0 0 8px; }
.welcome-desc { font-size: 14px; color: #86909C; margin: 0; max-width: 500px; margin: 0 auto; line-height: 1.6; }

/* ===== 能力分类标签 ===== */
.capability-tabs {
  display: flex; gap: 8px; padding: 0 4px; margin-bottom: 20px;
  overflow-x: auto; scrollbar-width: none;
}
.capability-tabs::-webkit-scrollbar { display: none; }
.cap-tab {
  display: flex; align-items: center; gap: 6px;
  padding: 8px 16px; border-radius: 20px; cursor: pointer;
  background: #F2F3F5; border: 2px solid transparent;
  font-size: 13px; font-weight: 500; color: #4E5969;
  white-space: nowrap; transition: all 0.2s;
}
.cap-tab:hover { background: #E5E6EB; }
.cap-tab.active {
  background: linear-gradient(135deg, rgba(123,104,238,0.1), rgba(255,107,53,0.1));
  border-color: #7B68EE; color: #7B68EE;
}
.cap-tab-icon { font-size: 16px; }
.cap-tab-name { font-size: 13px; }

/* ===== 引导卡片 ===== */
.guide-cards {
  display: grid; grid-template-columns: repeat(2, 1fr); gap: 12px;
  margin-bottom: 24px;
}
.guide-card {
  padding: 16px; border-radius: 12px; cursor: pointer;
  background: #fff; border: 1px solid #E5E6EB;
  transition: all 0.2s;
}
.guide-card:hover {
  border-color: #7B68EE; transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(123,104,238,0.1);
}
.gc-header { display: flex; align-items: center; gap: 8px; margin-bottom: 8px; }
.gc-icon { font-size: 20px; }
.gc-tag {
  font-size: 11px; padding: 2px 8px; border-radius: 10px;
  background: rgba(123,104,238,0.1); color: #7B68EE; font-weight: 600;
}
.gc-text { font-size: 14px; font-weight: 600; color: #1D2129; margin-bottom: 6px; }
.gc-example {
  font-size: 12px; color: #86909C; line-height: 1.5;
  background: #F7F8FA; padding: 6px 8px; border-radius: 6px;
}

/* ===== 最近查询快捷入口 ===== */
.recent-quick { margin-top: 8px; }
.rq-title { font-size: 13px; font-weight: 600; color: #4E5969; margin-bottom: 8px; }
.rq-list { display: flex; flex-direction: column; gap: 6px; }
.rq-item {
  display: flex; align-items: center; gap: 8px;
  padding: 8px 12px; border-radius: 8px; cursor: pointer;
  background: #F7F8FA; transition: background 0.2s;
}
.rq-item:hover { background: #E5E6EB; }
.rq-icon { font-size: 14px; }
.rq-text { font-size: 13px; color: #1D2129; flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.rq-time { font-size: 11px; color: #C9CDD4; white-space: nowrap; }

/* ===== 回复后引导建议 ===== */
.post-reply-suggestions { padding: 8px 0 4px; }
.prs-label { font-size: 12px; color: #86909C; margin-bottom: 8px; }
.prs-chips { display: flex; flex-wrap: wrap; gap: 8px; }
.prs-chip {
  padding: 6px 14px; border-radius: 16px; cursor: pointer;
  background: #F2F3F5; font-size: 12px; color: #4E5969;
  border: 1px solid #E5E6EB; transition: all 0.2s;
}
.prs-chip:hover {
  background: rgba(123,104,238,0.08); border-color: #7B68EE; color: #7B68EE;
}

/* ===== 输入区域增强 ===== */
.ai-input-area { border-top: 1px solid #F2F3F5; padding: 16px; }
.ai-input-wrap {
  display: flex; align-items: flex-end; gap: 12px;
  background: #F7F8FA; border-radius: 12px; padding: 12px;
  border: 2px solid #F2F3F5; transition: border-color 0.2s;
}
.ai-input-wrap:focus-within { border-color: #7B68EE; }
.ai-textarea {
  flex: 1; border: none; background: transparent; resize: none;
  font-size: 14px; color: #1D2129; outline: none; line-height: 1.5;
}
.ai-textarea::placeholder { color: #C9CDD4; }
.ai-send {
  width: 40px; height: 40px; border-radius: 10px;
  background: linear-gradient(135deg, #7B68EE, #5B4BC7);
  color: #fff; border: none; cursor: pointer;
  display: flex; align-items: center; justify-content: center;
  flex-shrink: 0; transition: all 0.2s;
}
.ai-send:hover:not(:disabled) { transform: scale(1.05); }
.ai-send:disabled { opacity: 0.5; cursor: not-allowed; }
.ai-send.loading { animation: rotate 1s linear infinite; }
@keyframes rotate { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }
.ai-input-footer { display: flex; align-items: center; justify-content: space-between; margin-top: 8px; padding: 0 4px; }
.ai-hint { font-size: 11px; color: #C9CDD4; margin: 0; }
.ai-quick-actions { display: flex; align-items: center; gap: 4px; }
.aqa-item {
  font-size: 11px; color: #86909C; cursor: pointer;
  padding: 2px 6px; border-radius: 6px; transition: all 0.2s;
}
.aqa-item:hover { background: rgba(123,104,238,0.08); color: #7B68EE; }
.aqa-divider { color: #E5E6EB; font-size: 11px; }

/* ===== 响应式 ===== */
@media screen and (max-width: 768px) {
  .guide-cards { grid-template-columns: 1fr; }
  .capability-tabs { gap: 4px; }
  .cap-tab { padding: 6px 12px; font-size: 12px; }
  .welcome-hero { padding: 24px 16px 20px; }
  .welcome-title { font-size: 18px; }
  .hb-right { display: none; }
  .ai-input-footer { flex-direction: column; align-items: flex-start; gap: 6px; }
  .ai-quick-actions { flex-wrap: wrap; }
}
</style>
