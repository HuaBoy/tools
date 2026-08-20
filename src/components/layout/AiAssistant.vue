<script setup>
import { ref, nextTick, watch, onMounted, onUnmounted } from 'vue';
import { useChatStore } from '@/stores/chat';
import { ElMessage } from 'element-plus';
import deviceData from '@/views/knowledge/deviceData.json';

const props = defineProps({
  visible: {
    type: Boolean,
    default: true
  }
});

const emit = defineEmits(['close']);

const chatStore = useChatStore();
const inputMessage = ref('');
const chatContainer = ref(null);
const isSending = ref(false);
const isMobile = ref(false);

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
    },
    {
      keyword: ['固件', 'firmware', '软件版本', '版本信息', '上位机', '下位机'],
      text: buildFirmwareText(null)
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

function buildFirmwareText(categoryFilter) {
  let categories = deviceData;
  if (categoryFilter) categories = deviceData.filter(c => c.id === categoryFilter);
  if (!categories.length) return '未找到匹配的设备类别。';

  let lines = ['📋 设备固件版本信息：\n'];
  categories.forEach(cat => {
    const versionMap = new Map();
    cat.versions.forEach(v => {
      const key = `${v.upperSoftVer}|${v.lowerSoftVer}`;
      if (!versionMap.has(key)) versionMap.set(key, { upper: v.upperSoftVer, lower: v.lowerSoftVer, count: 0 });
      versionMap.get(key).count++;
    });
    lines.push(`${cat.name}（${cat.versions.length} 台设备）`);
    versionMap.forEach(ver => {
      lines.push(`• 上位机 ${ver.upper} / 下位机 ${ver.lower} — ${ver.count} 台`);
    });
  });
  lines.push('\n👉 完整版本管理请前往「生产履历」页面');
  return lines.join('\n');
}

const sendMessage = async () => {
  if (!inputMessage.value.trim() || isSending.value) return;
  
  const message = inputMessage.value.trim();
  chatStore.addMessage('user', message);
  inputMessage.value = '';
  isSending.value = true;
  
  await nextTick();
  scrollToBottom();
  
  // 优先：固件版本查询（本地匹配）
  const lower = message.toLowerCase();
  const hasFirmwareKeyword = /固件|firmware|软件版本|版本信息|上位机|下位机|系统版本/.test(lower);
  if (hasFirmwareKeyword) {
    let matchedCat = null;
    const aliases = {
      '起爆器': 'qbq', '起爆': 'qbq', '雷管': 'lg',
      '导爆管': 'dbgsjk', '导爆': 'dbgsjk', '数码': 'dbgsjk',
      '注码': 'jzzm', '检测': 'jzzm',
      '数模': 'smmsjzx', '模拟': 'smmsjzx',
      '测试仪': 'qbkcsy', '检测仪': 'qbkcsy',
      '中继': 'zjx', '中继器': 'zjx'
    };
    for (const [alias, catId] of Object.entries(aliases)) {
      if (lower.includes(alias)) { matchedCat = catId; break; }
    }
    if (!matchedCat) {
      for (const cat of deviceData) {
        if (lower.includes(cat.name.toLowerCase()) || lower.includes(cat.id.toLowerCase())) {
          matchedCat = cat.id; break;
        }
      }
    }
    const response = buildFirmwareText(matchedCat);
    chatStore.addMessage('ai', response);
    isSending.value = false;
    await nextTick();
    scrollToBottom();
    return;
  }

  await new Promise(resolve => setTimeout(resolve, 1500));
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

const checkMobile = () => {
  isMobile.value = window.innerWidth <= 768;
};

watch(() => chatStore.messages.length, () => {
  nextTick(scrollToBottom);
});

onMounted(() => {
  checkMobile();
  window.addEventListener('resize', checkMobile);
});

onUnmounted(() => {
  window.removeEventListener('resize', checkMobile);
});
</script>

<template>
  <div>
    <button 
      v-if="isMobile && !visible" 
      class="assistant-toggle-btn" 
      @click="emit('close', false)"
    >
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />
      </svg>
    </button>
    
    <aside class="ai-assistant" :class="{ 'assistant-mobile': isMobile, 'assistant-hidden': isMobile && !visible }">
      <div class="assistant-overlay" v-if="isMobile && visible" @click="emit('close', false)"></div>
      
      <div class="assistant-header">
        <div class="assistant-title">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />
          </svg>
          <span>AI运维智能助手</span>
        </div>
        <button v-if="isMobile" class="assistant-close" @click="emit('close', false)">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <line x1="18" y1="6" x2="6" y2="18" />
            <line x1="6" y1="6" x2="18" y2="18" />
          </svg>
        </button>
      </div>
      
      <div ref="chatContainer" class="chat-container">
        <div 
          v-for="msg in chatStore.messages" 
          :key="msg.id" 
          class="message-item"
          :class="msg.type"
        >
          <div class="message-avatar">
            <svg v-if="msg.type === 'ai'" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />
            </svg>
            <svg v-else width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" />
              <circle cx="12" cy="7" r="4" />
            </svg>
          </div>
          <div class="message-content">
            <p>{{ msg.content }}</p>
          </div>
        </div>
        
        <div v-if="isSending" class="message-item ai">
          <div class="message-avatar">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />
            </svg>
          </div>
          <div class="message-content typing">
            <span class="typing-dot"></span>
            <span class="typing-dot"></span>
            <span class="typing-dot"></span>
          </div>
        </div>
      </div>
      
      <div class="chat-input-area">
        <textarea 
          v-model="inputMessage"
          placeholder="输入问题，例如：E001故障怎么处理？"
          class="chat-input"
          @keydown="handleKeyPress"
        ></textarea>
        <button 
          class="send-btn" 
          :disabled="!inputMessage.trim() || isSending"
          @click="sendMessage"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <line x1="22" y1="2" x2="11" y2="13" />
            <polygon points="22 2 15 22 11 13 2 9 22 2" />
          </svg>
        </button>
      </div>
    </aside>
  </div>
</template>

<style scoped>
.ai-assistant {
  position: fixed;
  top: 60px;
  right: 0;
  width: 320px;
  height: calc(100vh - 60px);
  background: rgba(15, 23, 42, 0.95);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border-left: 1px solid rgba(22, 93, 255, 0.2);
  display: flex;
  flex-direction: column;
  z-index: 99;
  transition: transform 0.3s ease;
}

.assistant-mobile {
  top: 56px;
  height: calc(100vh - 56px);
  width: 280px;
  box-shadow: -4px 0 20px rgba(0, 0, 0, 0.3);
}

.assistant-hidden {
  transform: translateX(100%);
}

.assistant-overlay {
  position: fixed;
  top: 56px;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 98;
}

.assistant-toggle-btn {
  position: fixed;
  bottom: 20px;
  right: 20px;
  background: linear-gradient(135deg, #165DFF 0%, #0F4CD0 100%);
  border: none;
  border-radius: 50%;
  padding: 14px;
  color: #FFFFFF;
  cursor: pointer;
  z-index: 100;
  box-shadow: 0 4px 12px rgba(22, 93, 255, 0.4);
  transition: all 0.2s;

  &:hover {
    transform: scale(1.1);
    box-shadow: 0 6px 16px rgba(22, 93, 255, 0.5);
  }
}

.assistant-header {
  padding: 16px 20px;
  border-bottom: 1px solid rgba(22, 93, 255, 0.2);
  background: rgba(22, 93, 255, 0.05);
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.assistant-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-weight: 600;
  color: #FFFFFF;
}

.assistant-close {
  background: rgba(245, 63, 63, 0.1);
  border: 1px solid rgba(245, 63, 63, 0.3);
  border-radius: 6px;
  color: #F53F3F;
  cursor: pointer;
  padding: 4px;
  transition: all 0.2s;

  &:hover {
    background: rgba(245, 63, 63, 0.2);
  }
}

.chat-container {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
}

.message-item {
  display: flex;
  gap: 10px;
  margin-bottom: 16px;
  
  &.ai {
    .message-content {
      background: rgba(22, 93, 255, 0.15);
      border: 1px solid rgba(22, 93, 255, 0.3);
      color: #94A3B8;
    }
    
    .message-avatar {
      color: #165DFF;
    }
  }
  
  &.user {
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
  max-width: 75%;
  padding: 10px 14px;
  border-radius: 12px;
  font-size: 13px;
  line-height: 1.5;
  
  &.typing {
    display: flex;
    gap: 6px;
    padding: 12px 14px;
  }
}

.typing-dot {
  width: 6px;
  height: 6px;
  background: #94A3B8;
  border-radius: 50%;
  animation: typing 1.4s infinite ease-in-out;
  
  &:nth-child(1) { animation-delay: 0s; }
  &:nth-child(2) { animation-delay: 0.2s; }
  &:nth-child(3) { animation-delay: 0.4s; }
}

@keyframes typing {
  0%, 80%, 100% { opacity: 0.2; }
  40% { opacity: 1; }
}

.chat-input-area {
  padding: 12px 16px;
  border-top: 1px solid rgba(22, 93, 255, 0.2);
  display: flex;
  gap: 10px;
}

.chat-input {
  flex: 1;
  background: rgba(15, 23, 42, 0.8);
  border: 1px solid rgba(22, 93, 255, 0.3);
  border-radius: 8px;
  padding: 10px 14px;
  color: #FFFFFF;
  font-size: 13px;
  resize: none;
  outline: none;
  min-height: 44px;
  max-height: 120px;
  
  &::placeholder {
    color: #64748B;
  }
  
  &:focus {
    border-color: rgba(22, 93, 255, 0.6);
  }
}

.send-btn {
  background: linear-gradient(135deg, #165DFF 0%, #0F4CD0 100%);
  border: none;
  border-radius: 8px;
  padding: 10px 14px;
  color: #FFFFFF;
  cursor: pointer;
  transition: all 0.2s;
  flex-shrink: 0;
  
  &:hover:not(:disabled) {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(22, 93, 255, 0.3);
  }
  
  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
}

@media screen and (max-width: 768px) {
  .ai-assistant {
    display: none;
  }
  
  .ai-assistant.assistant-mobile {
    display: flex;
  }
}

@media screen and (max-width: 480px) {
  .assistant-mobile {
    width: 100%;
  }
  
  .message-content {
    max-width: 85%;
    font-size: 12px;
    padding: 8px 12px;
  }
  
  .chat-input {
    font-size: 12px;
    padding: 8px 12px;
  }
  
  .assistant-toggle-btn {
    padding: 12px;
    bottom: 16px;
    right: 16px;
  }
}
</style>
