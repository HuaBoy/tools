<script setup>
import { ref, nextTick, watch, onMounted, onUnmounted } from 'vue';
import { useChatStore } from '@/stores/chat';

const chatStore = useChatStore();
const inputMessage = ref('');
const chatContainer = ref(null);
const isSending = ref(false);

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