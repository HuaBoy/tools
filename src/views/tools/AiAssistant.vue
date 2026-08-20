<script setup>
import { ref, nextTick, watch, onMounted, onUnmounted } from 'vue';
import { useRouter } from 'vue-router';
import { useChatStore } from '@/stores/chat';
import deviceData from '@/views/knowledge/deviceData.json';

const chatStore = useChatStore();
const router = useRouter();
const inputMessage = ref('');
const chatContainer = ref(null);
const isSending = ref(false);

// ==================== 工具调用：AI起爆数据查询 ====================
const BLAST_QUERY_URL = '/api/blade-detonate/blastTask/page';
const BLAST_COLUMNS = [
  { key: 'tenantName', label: '雷管企业' },
  { key: 'deptName', label: '使用单位' },
  { key: 'controllerCode', label: '控制器编号' },
  { key: 'deviceCode', label: '手持机编号' },
  { key: 'deviceVersion', label: '手持机版本' },
  { key: 'detonatorCount', label: '爆破数量' },
  { key: 'taskName', label: '工程名称' },
  { key: 'blasterUserName', label: '作业人员' },
  { key: 'explosionDate', label: '爆破时间' },
  { key: 'uploadDlTime', label: '上传时间' }
];

// 解析「查询SN设备信息」意图：识别 DZ 开头的 SN 编号 + 查询关键词
const parseDeviceQueryIntent = (msg) => {
  const hasQueryIntent = /查询|查|设备|信息|sn|编号|检索|搜索|记录|数据/.test(msg.toLowerCase());
  const snMatch = msg.match(/\bDZ\w*\d+\b/i);
  if (hasQueryIntent && snMatch) {
    return { deviceCode: snMatch[0].toUpperCase() };
  }
  return null;
};

// 最近7天日期范围（yyyy-MM-dd）
const getRecentWeekRange = () => {
  const end = new Date();
  const start = new Date();
  start.setDate(end.getDate() - 6);
  const fmt = (d) => `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
  return { startDate: fmt(start), endDate: fmt(end) };
};

// 调用 AI起爆数据查询接口
const queryBlastData = async (deviceCode) => {
  const token = localStorage.getItem('mp_token');
  if (!token) {
    return { error: 'NO_TOKEN' };
  }
  const { startDate, endDate } = getRecentWeekRange();
  const params = new URLSearchParams({
    startDate,
    endDate,
    deviceCode,
    current: '1',
    size: '10'
  });
  const response = await fetch(`${BLAST_QUERY_URL}?${params.toString()}`, {
    method: 'GET',
    headers: {
      'accept': 'application/json, text/plain, */*',
      'authorization': 'Basic ' + btoa('saber:saber_secret'),
      'blade-auth': `bearer ${token}`,
      'tenant-id': '000000'
    }
  });
  const result = await response.json();
  if (result.code === 200 && result.data) {
    return { records: result.data.records || [], total: result.data.total || 0 };
  }
  if (result.code === 401) {
    return { error: 'TOKEN_EXPIRED' };
  }
  return { error: result.msg || result.message || '查询失败' };
};

// 处理设备信息查询（工具调用主逻辑）
const handleDeviceQuery = async (intent) => {
  chatStore.addMessage('ai', `正在为您查询设备 ${intent.deviceCode} 的AI起爆数据（最近7天），请稍候...`);
  const result = await queryBlastData(intent.deviceCode);
  if (result.error === 'NO_TOKEN') {
    chatStore.addMessage('ai', '未登录云平台，无法查询AI起爆数据。请先登录后重试。', { type: 'error', hint: 'goto-login' });
    return;
  }
  if (result.error === 'TOKEN_EXPIRED') {
    chatStore.addMessage('ai', '云平台登录已过期，请重新登录后再查询。', { type: 'error', hint: 'goto-login' });
    return;
  }
  if (result.error) {
    chatStore.addMessage('ai', `查询失败：${result.error}`, { type: 'error' });
    return;
  }
  if (result.total === 0) {
    chatStore.addMessage('ai', `设备 ${intent.deviceCode} 最近7天没有AI起爆记录。`, { type: 'empty' });
    return;
  }
  chatStore.addMessage(
    `已为您查询到设备 ${intent.deviceCode} 的AI起爆数据，共 ${result.total} 条（展示前 ${result.records.length} 条）：`,
    { type: 'table', columns: BLAST_COLUMNS, records: result.records, total: result.total, deviceCode: intent.deviceCode }
  );
};

// 格式化单元格值
const formatCell = (value) => {
  if (value === null || value === undefined || value === '') return '-';
  return value;
};

// 跳转到 AI起爆数据查询页面（带设备号预填）
const goToDataQuery = (deviceCode) => {
  router.push({ path: '/data/query', query: deviceCode ? { deviceCode } : {} });
};

// ==================== 工具调用：设备固件版本查询 ====================

// 解析「固件版本」意图：识别固件/版本/软件 等关键词，可选设备类别名称
const parseFirmwareQueryIntent = (msg) => {
  const lower = msg.toLowerCase();
  const hasFirmwareIntent = /固件|firmware|软件版本|版本信息|软硬件|上位机|下位机|系统版本/.test(lower);
  const hasQueryIntent = /查询|查|查看|看看|知道|是什么|多少|列出|显示|检索/.test(lower);
  if (!hasFirmwareIntent && !hasQueryIntent) return null;

  // 提取设备类别关键词（支持从 deviceData.json 的 name 和 id 匹配）
  const categoryKeywords = [];
  deviceData.forEach(cat => {
    categoryKeywords.push({ keyword: cat.name, id: cat.id });
    categoryKeywords.push({ keyword: cat.id, id: cat.id });
  });

  let matchedCategory = null;
  for (const ck of categoryKeywords) {
    if (lower.includes(ck.keyword.toLowerCase())) {
      matchedCategory = ck.id;
      break;
    }
  }

  // 关键词简写匹配
  if (!matchedCategory) {
    const aliases = {
      '起爆器': 'qbq', '起爆': 'qbq', '雷管': 'lg',
      '导爆管': 'dbgsjk', '导爆': 'dbgsjk', '数码': 'dbgsjk',
      '注码': 'jzzm', '检测': 'jzzm',
      '数模': 'smmsjzx', '模拟': 'smmsjzx',
      '测试仪': 'qbkcsy', '检测仪': 'qbkcsy',
      '中继': 'zjx', '中继器': 'zjx'
    };
    for (const [alias, catId] of Object.entries(aliases)) {
      if (lower.includes(alias)) {
        matchedCategory = catId;
        break;
      }
    }
  }

  return { category: matchedCategory };
};

// 查询设备固件版本（从本地 deviceData.json）
const queryFirmwareInfo = (categoryFilter) => {
  let categories = deviceData;
  if (categoryFilter) {
    categories = deviceData.filter(cat => cat.id === categoryFilter);
  }

  const results = [];
  categories.forEach(cat => {
    // 收集该类别下所有唯一版本
    const versionMap = new Map();
    cat.versions.forEach(v => {
      const key = `${v.upperSoftVer}|${v.lowerSoftVer}|${v.upperSoftName || '-'}`;
      if (!versionMap.has(key)) {
        versionMap.set(key, {
          upperSoftVer: v.upperSoftVer,
          lowerSoftVer: v.lowerSoftVer,
          upperSoftName: v.upperSoftName || '-',
          count: 0
        });
      }
      versionMap.get(key).count++;
    });

    results.push({
      categoryId: cat.id,
      categoryName: cat.name,
      color: cat.color,
      bgColor: cat.bgColor,
      totalDevices: cat.versions.length,
      versions: Array.from(versionMap.values())
    });
  });

  return results;
};

// 处理固件版本查询
const handleFirmwareQuery = async (intent) => {
  const queryLabel = intent.category ? `「${deviceData.find(c => c.id === intent.category)?.name || intent.category}」` : '全部设备类别';
  chatStore.addMessage('ai', `正在为您查询 ${queryLabel} 的固件版本信息，请稍候...`);

  // 模拟短暂延迟
  await new Promise(resolve => setTimeout(resolve, 400));

  const results = queryFirmwareInfo(intent.category);

  if (results.length === 0) {
    chatStore.addMessage('ai', `未找到匹配的设备类别。如需查看所有设备固件版本，请直接说"查看固件版本信息"。`);
    return;
  }

  // 构建结构化回复
  let totalCategories = results.length;
  let totalVersions = 0;
  results.forEach(r => totalVersions += r.versions.length);

  let summary = `已查询到 ${totalCategories} 个设备类别，共 ${totalVersions} 个版本组合：`;

  chatStore.addMessage('ai', summary, {
    type: 'firmware',
    categories: results,
    totalCategories,
    totalVersions
  });
};

// 跳转到设备版本页面
const goToDeviceVersion = (categoryId) => {
  router.push({ path: '/knowledge/production-history', query: categoryId ? { category: categoryId } : {} });
};

const aiResponses = [
  '根据您的问题，我查询到相关信息如下：',
  '这是一个常见问题，建议按照以下步骤处理：',
  '分析完成，主要问题在于参数配置异常，请检查相关设置。',
  '日志分析结果显示通信故障，建议检查网络连接。',
  '根据起爆数据，任务执行正常，未发现异常。',
  '翻译结果已生成，专业术语已准确转换。',
  '数据追溯链路已完整建立，可以查看详细信息。',
  '故障代码对应的处理方案已找到，请参考知识库。'
];

const getRandomResponse = () => {
  const responses = [
    {
      keyword: ['故障', 'error', '报错'],
      text: '这是一个常见故障问题。建议按照以下步骤处理：\n1. 检查设备连接状态\n2. 查看日志获取详细错误信息\n3. 参考知识库中的对应故障代码解决方案\n4. 如问题仍未解决，请联系技术支持'
    },
    {
      keyword: ['日志', 'log'],
      text: '日志分析完成。主要发现：\n- 通信故障：3次\n- 参数异常：2次\n- 时序错误：1次\n建议重点关注通信模块，可能存在网络不稳定问题。'
    },
    {
      keyword: ['翻译', 'translate'],
      text: '专业术语翻译完成。爆破行业常用术语对照：\n- 起爆器 -> Initiator\n- 导爆管 -> Detonating Cord\n- 延时时间 -> Delay Time\n- 起爆网络 -> Blasting Network'
    },
    {
      keyword: ['追溯', 'trace'],
      text: '数据追溯链路已建立：\n1. 授权记录 -> LIC2026-QB-001\n2. 加密日志 -> LOG_20240115.dat\n3. 起爆任务 -> TASK_00123\n4. 故障记录 -> 无异常\n全链路状态：正常'
    },
    {
      keyword: ['授权', 'license'],
      text: '当前授权状态：有效\n授权码：LIC2026-QB-001\n剩余天数：365天\n如需转换设备码，请使用「授权码转换工具」页面。'
    },
    {
      keyword: ['查询', 'search', '数据'],
      text: '数据查询功能支持：\n- 设备号模糊搜索\n- 批次号查询\n- 时间范围筛选\n- AI智能检索\n建议在「AI起爆数据查询」页面进行详细查询。'
    }
  ];
  const msg = inputMessage.value.toLowerCase();
  for (const item of responses) {
    if (item.keyword.some(k => msg.includes(k))) {
      return item.text;
    }
  }
  return aiResponses[Math.floor(Math.random() * aiResponses.length)];
};

const sendMessage = async () => {
  if (!inputMessage.value.trim() || isSending.value) return;
  
  const message = inputMessage.value.trim();
  chatStore.addMessage('user', message);
  inputMessage.value = '';
  isSending.value = true;
  
  await nextTick();
  scrollToBottom();

  // 1. 工具调用：AI起爆数据查询（查询SN设备信息）
  const deviceIntent = parseDeviceQueryIntent(message);
  if (deviceIntent) {
    try {
      await handleDeviceQuery(deviceIntent);
    } catch (e) {
      chatStore.addMessage('ai', `网络异常：${e.message}`, { type: 'error' });
    }
    isSending.value = false;
    await nextTick();
    scrollToBottom();
    return;
  }

  // 2. 工具调用：设备固件版本查询
  const firmwareIntent = parseFirmwareQueryIntent(message);
  if (firmwareIntent) {
    try {
      await handleFirmwareQuery(firmwareIntent);
    } catch (e) {
      chatStore.addMessage('ai', `固件版本查询异常：${e.message}`, { type: 'error' });
    }
    isSending.value = false;
    await nextTick();
    scrollToBottom();
    return;
  }

  // 3. 默认：规则回复
  await new Promise(resolve => setTimeout(resolve, 800));
  const response = getRandomResponse();
  chatStore.addMessage('ai', response);
  isSending.value = false;
  
  await nextTick();
  scrollToBottom();
};

const scrollToBottom = () => {
  if (chatContainer.value) {
    chatContainer.value.scrollTop = chatContainer.value.scrollHeight;
  }
};

const handleKeyPress = (e) => {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault();
    sendMessage();
  }
};

watch(() => chatStore.messages.length, () => {
  nextTick(scrollToBottom);
});

onMounted(() => {
  scrollToBottom();
});
</script>

<template>
  <div class="ai-assistant-page">
    <div class="page-header">
      <div class="header-title">
        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />
        </svg>
        <h1>AI运维智能助手</h1>
      </div>
      <p class="header-desc">智能问答系统，为您提供专业的运维支持服务</p>
    </div>
    
    <div class="chat-wrapper">
      <div ref="chatContainer" class="chat-container">
        <div 
          v-for="msg in chatStore.messages" 
          :key="msg.id" 
          class="message-item"
          :class="msg.type"
        >
          <div class="message-avatar">
            <svg v-if="msg.type === 'ai'" width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />
            </svg>
            <svg v-else width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" />
              <circle cx="12" cy="7" r="4" />
            </svg>
          </div>
          <div class="message-content">
            <p>{{ msg.content }}</p>
            <!-- 工具调用结果：表格 -->
            <div v-if="msg.payload && msg.payload.type === 'table'" class="tool-result">
              <div class="tool-result-header">
                <span>查询结果：{{ msg.payload.total }} 条记录</span>
                <button class="goto-btn" @click="goToDataQuery(msg.payload.deviceCode)">
                  在「AI起爆数据查询」中查看完整数据 →
                </button>
              </div>
              <div class="tool-table-wrapper">
                <table class="tool-table">
                  <thead>
                    <tr>
                      <th>#</th>
                      <th v-for="col in msg.payload.columns" :key="col.key">{{ col.label }}</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="(record, idx) in msg.payload.records" :key="idx">
                      <td>{{ idx + 1 }}</td>
                      <td v-for="col in msg.payload.columns" :key="col.key">{{ formatCell(record[col.key]) }}</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
            <!-- 工具调用结果：错误/空态 + 登录引导 -->
            <div v-if="msg.payload && msg.payload.hint === 'goto-login'" class="tool-result error">
              <button class="goto-btn" @click="goToDataQuery()">前往「AI起爆数据查询」登录 →</button>
            </div>
            <!-- 工具调用结果：固件版本信息 -->
            <div v-if="msg.payload && msg.payload.type === 'firmware'" class="tool-result firmware-result">
              <div class="tool-result-header">
                <span>查询结果：{{ msg.payload.totalCategories }} 个类别，{{ msg.payload.totalVersions }} 个版本组合</span>
                <button class="goto-btn" @click="goToDeviceVersion()">
                  在「生产履历」中查看完整信息 →
                </button>
              </div>
              <div class="firmware-cards">
                <div 
                  v-for="cat in msg.payload.categories" 
                  :key="cat.categoryId" 
                  class="firmware-card"
                  :style="{ borderLeftColor: cat.color, background: cat.bgColor + '20' }"
                >
                  <div class="firmware-card-header" :style="{ background: cat.bgColor }">
                    <span class="firmware-card-title">{{ cat.categoryName }}</span>
                    <span class="firmware-card-count">{{ cat.totalDevices }} 台设备</span>
                  </div>
                  <div class="firmware-version-list">
                    <div v-for="(ver, idx) in cat.versions" :key="idx" class="firmware-version-item">
                      <div class="version-row">
                        <span class="version-label">上位机</span>
                        <span class="version-tag upper" :style="{ background: cat.bgColor, color: cat.color }">{{ ver.upperSoftVer }}</span>
                      </div>
                      <div class="version-row">
                        <span class="version-label">下位机</span>
                        <span class="version-tag lower">{{ ver.lowerSoftVer }}</span>
                      </div>
                      <div class="version-row">
                        <span class="version-label">数量</span>
                        <span class="version-count">{{ ver.count }} 台</span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
        
        <div v-if="isSending" class="message-item ai">
          <div class="message-avatar">
            <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />
            </svg>
          </div>
          <div class="message-content typing">
            <span class="typing-dot"></span>
            <span class="typing-dot"></span>
            <span class="typing-dot"></span>
          </div>
        </div>
        
        <div v-if="chatStore.messages.length === 0" class="welcome-message">
          <div class="welcome-icon">
            <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
              <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />
            </svg>
          </div>
          <h3>欢迎使用AI运维智能助手</h3>
          <p>我可以帮助您解决运维相关问题，例如：</p>
          <ul class="example-questions">
            <li>查询SN编号为DZ600000016的设备信息</li>
            <li>查看起爆器的固件版本信息</li>
            <li>故障代码E001怎么处理？</li>
            <li>如何分析起爆器日志？</li>
            <li>请翻译专业术语</li>
            <li>如何追溯数据链路？</li>
          </ul>
        </div>
      </div>
      
      <div class="chat-input-area">
        <textarea 
          v-model="inputMessage"
          placeholder="输入您的问题，我来为您解答..."
          class="chat-input"
          @keydown="handleKeyPress"
        ></textarea>
        <button 
          class="send-btn" 
          :disabled="!inputMessage.trim() || isSending"
          @click="sendMessage"
        >
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <line x1="22" y1="2" x2="11" y2="13" />
            <polygon points="22 2 15 22 11 13 2 9 22 2" />
          </svg>
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.ai-assistant-page {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.page-header {
  padding: 20px 24px;
  border-bottom: 1px solid var(--border-color);
  margin-bottom: 20px;
}

.header-title {
  display: flex;
  align-items: center;
  gap: 12px;
  color: var(--text-primary);
  font-size: 20px;
  font-weight: 600;
  margin-bottom: 8px;
}

.header-title svg {
  color: #165DFF;
}

.header-desc {
  color: var(--text-secondary);
  font-size: 14px;
  margin: 0;
}

.chat-wrapper {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: var(--bg-input);
  border-radius: 12px;
  border: 1px solid var(--border-color);
  overflow: hidden;
}

.chat-container {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
}

.message-item {
  display: flex;
  gap: 12px;
  margin-bottom: 20px;
  max-width: 80%;
  
  &.ai {
    .message-content {
      background: rgba(22, 93, 255, 0.15);
      border: 1px solid rgba(22, 93, 255, 0.3);
      color: #E2E8F0;
    }
    
    .message-avatar {
      color: #165DFF;
    }
  }
  
  &.user {
    margin-left: auto;
    flex-direction: row-reverse;
    
    .message-content {
      background: rgba(0, 180, 42, 0.15);
      border: 1px solid rgba(0, 180, 42, 0.3);
      color: #FFFFFF;
    }
    
    .message-avatar {
      color: #00B42A;
    }
  }
}

.message-avatar {
  flex-shrink: 0;
}

.message-content {
  padding: 14px 18px;
  border-radius: 16px;
  font-size: 14px;
  line-height: 1.6;
  
  &.typing {
    display: flex;
    gap: 8px;
    padding: 14px 16px;
  }
}

.typing-dot {
  width: 8px;
  height: 8px;
  background: var(--text-secondary);
  border-radius: 50%;
  animation: typing 1.4s infinite ease-in-out;
  
  &:nth-child(1) { animation-delay: 0s; }
  &:nth-child(2) { animation-delay: 0.2s; }
  &:nth-child(3) { animation-delay: 0.4s; }
}

@keyframes typing {
  0%, 80%, 100% { opacity: 0.3; }
  40% { opacity: 1; }
}

/* ==================== 工具调用结果 ==================== */
.tool-result {
  margin-top: 12px;
  border: 1px solid rgba(22, 93, 255, 0.25);
  border-radius: 10px;
  overflow: hidden;
  background: rgba(22, 93, 255, 0.06);
  
  &.error {
    border-color: rgba(245, 63, 63, 0.3);
    background: rgba(245, 63, 63, 0.06);
    padding: 10px 14px;
  }
}

.tool-result-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 8px 14px;
  border-bottom: 1px solid rgba(22, 93, 255, 0.15);
  font-size: 13px;
  color: var(--text-secondary);
}

.goto-btn {
  background: linear-gradient(135deg, #165DFF 0%, #0F4CD0 100%);
  border: none;
  border-radius: 6px;
  padding: 4px 10px;
  color: #FFFFFF;
  font-size: 12px;
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.2s;
  
  &:hover {
    opacity: 0.85;
  }
}

.tool-table-wrapper {
  overflow-x: auto;
  max-height: 260px;
  overflow-y: auto;
}

.tool-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
  color: var(--text-primary);
  
  th, td {
    padding: 6px 10px;
    text-align: left;
    border-bottom: 1px solid rgba(22, 93, 255, 0.12);
    white-space: nowrap;
    max-width: 160px;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  
  th {
    background: rgba(22, 93, 255, 0.1);
    color: var(--text-secondary);
    font-weight: 600;
    position: sticky;
    top: 0;
  }
  
  tbody tr:hover {
    background: rgba(22, 93, 255, 0.06);
  }
}

/* ==================== 固件版本信息卡片 ==================== */
.firmware-result {
  border-color: rgba(0, 180, 42, 0.3);
  background: rgba(0, 180, 42, 0.04);
  
  .tool-result-header {
    border-bottom-color: rgba(0, 180, 42, 0.15);
  }
}

.firmware-cards {
  padding: 10px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.firmware-card {
  border-radius: 8px;
  border: 1px solid rgba(0, 0, 0, 0.08);
  border-left: 4px solid;
  overflow: hidden;
}

.firmware-card-header {
  padding: 8px 12px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.firmware-card-title {
  font-weight: 600;
  font-size: 13px;
  color: var(--text-primary);
}

.firmware-card-count {
  font-size: 12px;
  color: var(--text-secondary);
}

.firmware-version-list {
  padding: 8px 12px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.firmware-version-item {
  display: flex;
  align-items: center;
  gap: 8px;
  background: rgba(255, 255, 255, 0.6);
  padding: 4px 10px;
  border-radius: 6px;
  font-size: 12px;
}

.version-row {
  display: flex;
  align-items: center;
  gap: 4px;
}

.version-label {
  color: var(--text-tertiary);
  font-size: 11px;
}

.version-tag {
  padding: 2px 8px;
  border-radius: 4px;
  font-weight: 600;
  font-size: 12px;
  
  &.upper {
    border: 1px solid;
  }
  
  &.lower {
    background: rgba(255, 152, 0, 0.15);
    color: #E65100;
    border: 1px solid rgba(255, 152, 0, 0.3);
  }
}

.version-count {
  font-weight: 500;
  color: var(--text-primary);
}

.welcome-message {
  text-align: center;
  padding: 60px 40px;
  color: var(--text-secondary);
}

.welcome-icon {
  width: 64px;
  height: 64px;
  margin: 0 auto 20px;
  color: #165DFF;
  opacity: 0.8;
}

.welcome-message h3 {
  color: var(--text-primary);
  font-size: 18px;
  font-weight: 600;
  margin-bottom: 12px;
}

.welcome-message p {
  margin-bottom: 20px;
  font-size: 14px;
}

.example-questions {
  list-style: none;
  padding: 0;
  margin: 0;
}

.example-questions li {
  padding: 8px 16px;
  background: rgba(22, 93, 255, 0.1);
  border-radius: 8px;
  margin-bottom: 8px;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
  
  &:hover {
    background: rgba(22, 93, 255, 0.2);
    color: var(--text-primary);
  }
  
  &:last-child {
    margin-bottom: 0;
  }
}

.chat-input-area {
  padding: 16px 20px;
  border-top: 1px solid var(--border-color);
  display: flex;
  gap: 12px;
  background: var(--bg-input);
}

.chat-input {
  flex: 1;
  background: var(--bg-input);
  border: 1px solid var(--border-color);
  border-radius: 10px;
  padding: 12px 16px;
  color: var(--text-primary);
  font-size: 14px;
  resize: none;
  outline: none;
  min-height: 48px;
  max-height: 150px;
  
  &::placeholder {
    color: var(--text-tertiary);
  }
  
  &:focus {
    border-color: rgba(22, 93, 255, 0.6);
  }
}

.send-btn {
  background: linear-gradient(135deg, #165DFF 0%, #0F4CD0 100%);
  border: none;
  border-radius: 10px;
  padding: 12px 18px;
  color: #FFFFFF;
  cursor: pointer;
  transition: all 0.2s;
  flex-shrink: 0;
  
  &:hover:not(:disabled) {
    transform: translateY(-2px);
    box-shadow: 0 4px 16px rgba(22, 93, 255, 0.4);
  }
  
  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
}

@media screen and (max-width: 768px) {
  .chat-container {
    padding: 16px;
  }
  
  .message-item {
    max-width: 95%;
  }
  
  .message-content {
    padding: 10px 14px;
    font-size: 13px;
  }
  
  .page-header {
    padding: 16px;
  }
  
  .header-title {
    font-size: 16px;
  }
  
  .welcome-message {
    padding: 40px 20px;
  }
}
</style>