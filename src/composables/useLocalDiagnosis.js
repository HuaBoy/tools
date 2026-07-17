/**
 * 本地智能诊断引擎
 * 不消耗 API Token，基于知识库 + 规则引擎实现类 AI 体验
 */
import { ref, reactive } from 'vue'

// ===== 知识库 =====
const knowledgeBase = [
  {
    id: 1,
    category: 'network',
    keywords: ['通信', '连接', '超时', '信号', '组网', '天线', '离线', '断连'],
    question: '设备与控制中心通信连接超时',
    answer: '1. 检查网络信号强度，确保信号强度大于-70dBm\n2. 重启通信模块，等待30秒后重试\n3. 检查天线连接是否牢固\n4. 尝试更换通信模块测试',
    severity: 'high'
  },
  {
    id: 2,
    category: 'firmware',
    keywords: ['参数', '配置', '范围', '校准', '超限', '临界'],
    question: '起爆参数配置超出允许范围',
    answer: '1. 检查参数配置是否符合规范要求\n2. 重新校准参数值到合理范围\n3. 使用默认参数模板恢复配置\n4. 联系技术支持获取参数配置指导',
    severity: 'high'
  },
  {
    id: 3,
    category: 'hardware',
    keywords: ['电源', '电压', '电池', '充电', '电量', '供电'],
    question: '设备电源电压低于正常工作范围',
    answer: '1. 检查电池电量是否充足\n2. 更换备用电池测试\n3. 检查充电器输出是否正常\n4. 检查电源线路是否有损坏',
    severity: 'critical'
  },
  {
    id: 4,
    category: 'network',
    keywords: ['起爆', '指令', '执行', '雷管', '发送'],
    question: '起爆指令发送后未成功执行',
    answer: '1. 检查雷管连接状态是否正常\n2. 重新发送起爆指令\n3. 检查起爆网络完整性\n4. 更换故障雷管',
    severity: 'critical'
  },
  {
    id: 5,
    category: 'network',
    keywords: ['数据', '同步', '云端', '上传', '下载'],
    question: '设备数据与云端同步失败',
    answer: '1. 检查网络连接是否正常\n2. 重新发起数据同步\n3. 检查数据格式是否正确\n4. 联系技术支持协助排查',
    severity: 'medium'
  },
  {
    id: 6,
    category: 'firmware',
    keywords: ['警告', '临界', '参数'],
    question: '参数警告：起爆参数接近临界值',
    answer: '1. 检查参数值是否合理\n2. 根据实际情况调整参数\n3. 确认是否需要继续执行\n4. 参考操作手册参数建议',
    severity: 'medium'
  },
  {
    id: 7,
    category: 'auth',
    keywords: ['授权', '验证', '授权码', '过期'],
    question: '授权码验证失败',
    answer: '1. 检查授权码输入是否正确\n2. 确认授权码是否过期\n3. 联系管理员获取新授权码\n4. 检查网络连接是否正常',
    severity: 'medium'
  },
  {
    id: 8,
    category: 'log',
    keywords: ['日志', '解析', '格式', '加密', '解密', '损坏'],
    question: '日志解析异常',
    answer: '1. 检查日志文件格式是否正确\n2. 确保日志文件未损坏\n3. 使用日志解密工具处理加密日志\n4. 联系技术支持协助分析',
    severity: 'low'
  },
  {
    id: 9,
    category: 'hardware',
    keywords: ['接口', '线缆', '连接', '松动', '接触'],
    question: '接口连接问题',
    answer: '1. 检查接口线缆连接是否牢固\n2. 更换接口线缆测试\n3. 检查接口是否有损坏\n4. 重启设备后重试',
    severity: 'low'
  },
  {
    id: 10,
    category: 'firmware',
    keywords: ['升级', '固件', 'OTA', '更新', '失败'],
    question: '程序升级失败',
    answer: '1. 检查网络连接是否稳定\n2. 确认固件文件完整性\n3. 确保设备电量充足\n4. 重新发起升级流程',
    severity: 'medium'
  },
  {
    id: 11,
    category: 'chip',
    keywords: ['芯片', 'ID', '序列号', '认证', '识别'],
    question: '芯片 ID 识别异常',
    answer: '1. 检查芯片是否安装正确\n2. 清理芯片触点\n3. 重新插拔芯片\n4. 更换芯片测试是否为硬件故障',
    severity: 'high'
  },
  {
    id: 12,
    category: 'chip',
    keywords: ['烧录', '程序', '写入', '失败', '芯片'],
    question: '芯片程序烧录失败',
    answer: '1. 检查烧录器连接状态\n2. 确认固件文件与芯片型号匹配\n3. 检查芯片是否被写保护\n4. 更换空白芯片测试',
    severity: 'high'
  },
  {
    id: 13,
    category: 'module',
    keywords: ['模块', '初始化', '启动', '自检'],
    question: '模块上电初始化失败',
    answer: '1. 检查供电电压是否正常\n2. 观察模块指示灯状态\n3. 重新上电复位\n4. 检查模块与底板连接',
    severity: 'high'
  },
  {
    id: 14,
    category: 'module',
    keywords: ['模块', '通信', '响应', '无应答'],
    question: '模块与上位机通信无响应',
    answer: '1. 检查通信接口连接\n2. 确认通信协议配置一致\n3. 检查模块地址设置\n4. 尝试单独连接该模块排查',
    severity: 'medium'
  },
  {
    id: 15,
    category: 'blasting',
    keywords: ['起爆', '延期', '时间', '不准确', '偏差'],
    question: '起爆延期时间偏差过大',
    answer: '1. 检查延期参数设置是否正确\n2. 确认时钟同步状态\n3. 检查环境温度是否在范围内\n4. 重新校准延期模块',
    severity: 'critical'
  },
  {
    id: 16,
    category: 'blasting',
    keywords: ['雷管', '检测', '电阻', '断路', '短路'],
    question: '雷管回路检测异常',
    answer: '1. 检查雷管接线是否正确\n2. 测量雷管电阻值\n3. 更换雷管测试\n4. 检查起爆母线是否完好',
    severity: 'critical'
  },
  {
    id: 17,
    category: 'log',
    keywords: ['日志', '分析', '异常', '报错', 'ERROR', 'WARN'],
    question: '日志频繁报错分析',
    answer: '1. 导出完整日志文件\n2. 使用日志过滤工具筛选错误级别\n3. 按时间线分析错误触发条件\n4. 对照操作手册确认操作流程',
    severity: 'medium'
  },
  {
    id: 18,
    category: 'trace',
    keywords: ['追溯', '批次', '追踪', '生产', '检测'],
    question: '批次数据追溯链路中断',
    answer: '1. 确认批次号是否录入系统\n2. 检查各环节数据是否上传完整\n3. 排查数据缺失环节\n4. 联系对应产线补录数据',
    severity: 'medium'
  }
]

// 分类配置
const categories = {
  hardware: { name: '硬件', color: '#F53F3F' },
  firmware: { name: '固件', color: '#FF7D00' },
  network: { name: '组网', color: '#165DFF' },
  auth: { name: '授权', color: '#722ED1' },
  log: { name: '日志', color: '#00B42A' },
  chip: { name: '芯片', color: '#14C9C9' },
  module: { name: '模块', color: '#F7BA1E' },
  blasting: { name: '起爆', color: '#FF7D00' },
  trace: { name: '追溯', color: '#0FC6C2' },
  other: { name: '其他', color: '#64748B' }
}

// 严重程度
const severityMap = {
  critical: { label: '紧急', color: '#F53F3F', weight: 5 },
  high: { label: '重要', color: '#FF7D00', weight: 4 },
  medium: { label: '一般', color: '#165DFF', weight: 3 },
  low: { label: '提示', color: '#00B42A', weight: 2 }
}

/**
 * 分词 + 计算匹配得分
 */
function calculateScore(query, item) {
  const q = query.toLowerCase()
  let score = 0

  // 1. 关键词精确匹配
  for (const kw of item.keywords) {
    if (q.includes(kw.toLowerCase())) score += 3
  }

  // 2. 标题匹配
  const titleWords = item.question.toLowerCase().split(/\s+/)
  const queryWords = q.split(/\s+/)
  for (const tw of titleWords) {
    for (const qw of queryWords) {
      if (tw.includes(qw) || qw.includes(tw)) score += 2
    }
  }

  // 3. 答案内容匹配
  const ansWords = item.answer.toLowerCase().split(/\s+/)
  for (const aw of ansWords) {
    for (const qw of queryWords) {
      if (aw.includes(qw) || qw.includes(aw)) score += 1
    }
  }

  // 4. 严重程度加权
  const sev = severityMap[item.severity]
  if (sev) score *= (1 + sev.weight * 0.1)

  return Math.round(score * 100) / 100
}

/**
 * 日志模式匹配 - 从日志文本中提取关键异常
 */
function parseLogPatterns(logText) {
  const patterns = []
  const lines = logText.split('\n')

  for (const line of lines) {
    const l = line.toLowerCase()
    if (l.includes('error') || l.includes('fail') || l.includes('exception') || l.includes('错误')) {
      patterns.push({ type: 'error', line: line.trim(), severity: 'high' })
    } else if (l.includes('warn') || l.includes('警告')) {
      patterns.push({ type: 'warning', line: line.trim(), severity: 'medium' })
    } else if (l.includes('timeout') || l.includes('超时')) {
      patterns.push({ type: 'timeout', line: line.trim(), severity: 'medium' })
    } else if (l.includes('retry') || l.includes('重试')) {
      patterns.push({ type: 'retry', line: line.trim(), severity: 'low' })
    }
  }

  return patterns
}

/**
 * 从用户描述中提取关键问题类型
 */
function extractIssueType(query) {
  const q = query.toLowerCase()
  if (/芯片|chip/.test(q)) return 'chip'
  if (/模块|module/.test(q)) return 'module'
  if (/起爆|雷管|爆破|blasting/.test(q)) return 'blasting'
  if (/连接|通信|网络|组网|信号|network|ble|蓝牙/.test(q)) return 'network'
  if (/日志|log|解密/.test(q)) return 'log'
  if (/固件|升级|ota|烧录|firmware/.test(q)) return 'firmware'
  if (/授权|auth|验证|activation/.test(q)) return 'auth'
  if (/电源|电压|电池|供电|hardware|硬件/.test(q)) return 'hardware'
  if (/追溯|批次|trace/.test(q)) return 'trace'
  return 'other'
}

// ===== Composable =====
export function useLocalDiagnosis() {
  const isAnalyzing = ref(false)
  const results = ref([])
  const logPatterns = ref([])
  const issueType = ref('')
  const suggestedTools = ref([])

  // 工具推荐映射
  const toolMap = {
    chip: [{ name: '芯片管理', path: '/hardware/chips' }],
    module: [{ name: '设备管理', path: '/hardware/devices' }],
    blasting: [{ name: '批次数据追溯', path: '/trace/analysis' }, { name: '全链路追溯', path: '/data/trace' }],
    network: [{ name: '日志解密工具', path: '/log/decrypt' }],
    log: [{ name: '日志解密工具', path: '/log/decrypt' }, { name: '起爆器日志AI分析', path: '/log/analysis' }],
    firmware: [{ name: '固件升级', path: '/hardware/firmware' }, { name: '版本手册', path: '/knowledge/manual' }],
    auth: [{ name: '三方账号授权', path: '/auth/third-party' }, { name: '授权码转换工具', path: '/auth/converter' }],
    hardware: [{ name: '设备管理', path: '/hardware/devices' }],
    trace: [{ name: '全链路追溯', path: '/data/trace' }, { name: 'AI起爆数据查询', path: '/data/query' }],
    other: [{ name: '智能知识库(RAG)', path: '/knowledge/rag' }, { name: '起爆器版本手册', path: '/knowledge/manual' }]
  }

  /**
   * 核心诊断方法 - 不消耗 API Token
   */
  async function diagnose(query, isLog = false) {
    isAnalyzing.value = true
    results.value = []
    logPatterns.value = []

    // 模拟分析延迟（本地计算很快，加一点以保持交互感）
    await new Promise(r => setTimeout(r, 200))

    if (!query || !query.trim()) {
      isAnalyzing.value = false
      return { results: [], patterns: [], type: '', tools: [] }
    }

    // 1. 提取问题类型
    issueType.value = extractIssueType(query)
    suggestedTools.value = toolMap[issueType.value] || toolMap.other

    // 2. 如果输入看起来像日志文本，做模式匹配
    if (isLog || query.split('\n').length > 3 || /error|fail|exception|warn/i.test(query)) {
      logPatterns.value = parseLogPatterns(query)
    }

    // 3. 知识库匹配
    const scored = knowledgeBase
      .map(item => ({ ...item, score: calculateScore(query, item) }))
      .filter(item => item.score > 0)
      .sort((a, b) => b.score - a.score)
      .slice(0, 6)

    results.value = scored
    isAnalyzing.value = false

    return {
      results: scored,
      patterns: logPatterns.value,
      type: issueType.value,
      tools: suggestedTools.value
    }
  }

  /**
   * 获取知识库全量数据
   */
  function getKnowledgeBase() {
    return knowledgeBase.map(item => ({
      ...item,
      categoryInfo: categories[item.category] || categories.other,
      severityInfo: severityMap[item.severity]
    }))
  }

  /**
   * 按分类筛选知识库
   */
  function getByCategory(cat) {
    return knowledgeBase.filter(item => item.category === cat)
  }

  function getCategories() {
    return Object.entries(categories).map(([code, info]) => ({
      code,
      ...info,
      count: knowledgeBase.filter(item => item.category === code).length
    }))
  }

  function getSeverityInfo(severity) {
    return severityMap[severity] || severityMap.low
  }

  function getCategoryInfo(code) {
    return categories[code] || categories.other
  }

  return {
    isAnalyzing,
    results,
    logPatterns,
    issueType,
    suggestedTools,
    diagnose,
    getKnowledgeBase,
    getByCategory,
    getCategories,
    getSeverityInfo,
    getCategoryInfo
  }
}
