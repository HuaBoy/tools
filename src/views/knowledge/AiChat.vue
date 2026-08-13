<script setup>
import { ref, reactive, nextTick, onMounted, computed } from 'vue';
import { ElMessage } from 'element-plus';
import GlassCard from '@/components/GlassCard.vue';
import { useLogsStore } from '@/stores/logs';
import { listConversations, getConversationMessages } from '@/api/knowledge';

const logsStore = useLogsStore();

// ===== 会话状态 =====
const conversations = ref([]);
const currentConvId = ref(0);
const messages = ref([]);
const inputText = ref('');
const thinking = ref(false);
const chatLoading = ref(false);
const messageListRef = ref(null);

// 默认欢迎消息
const welcomeMsg = {
  role: 'assistant',
  content: '你好！我是企业内部智能助手。\n\n你可以问我关于：\n· 设备故障排查\n· 操作流程指引\n· 业务规范制度\n· 常见问题解答\n\n我会基于知识库中的文档和问题库来回答你的问题。'
};

const loadConversations = async () => {
  try {
    conversations.value = await listConversations();
  } catch (e) {
    console.warn('加载会话失败:', e);
  }
};

const switchConversation = async (id) => {
  currentConvId.value = id;
  chatLoading.value = true;
  try {
    const history = await getConversationMessages(id);
    messages.value = history.length ? history : [welcomeMsg];
  } catch (e) {
    ElMessage.error('加载历史失败：' + (e.message || ''));
    messages.value = [welcomeMsg];
  } finally {
    chatLoading.value = false;
    scrollToBottom();
  }
};

const newConversation = () => {
  currentConvId.value = 0;
  messages.value = [welcomeMsg];
  inputText.value = '';
};

// ===== 发送消息（SSE 流式） =====
const sendMessage = async () => {
  const text = inputText.value.trim();
  if (!text || thinking.value) return;

  // 追加用户消息
  messages.value.push({ role: 'user', content: text });
  inputText.value = '';
  scrollToBottom();

  thinking.value = true;
  // 追加空的 assistant 消息，逐步填充
  const aiMsg = { role: 'assistant', content: '' };
  messages.value.push(aiMsg);
  scrollToBottom();

  const token = localStorage.getItem('auth_token');

  try {
    const resp = await fetch('/api/v1/knowledge/chat', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer ' + (token || '')
      },
      body: JSON.stringify({
        question: text,
        conversation_id: currentConvId.value,
        stream: true
      })
    });

    if (!resp.ok) {
      const err = await resp.json().catch(() => ({}));
      throw new Error(err.message || '请求失败 (' + resp.status + ')');
    }

    // 当前会话 ID 更新
    let convId = currentConvId.value;
    let fullAnswer = '';

    const reader = resp.body.getReader();
    const decoder = new TextDecoder();
    let buffer = '';

    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      buffer += decoder.decode(value, { stream: true });

      // 解析 SSE 帧：data: {...}\n\n
      const lines = buffer.split('\n');
      buffer = lines.pop();

      for (const line of lines) {
        if (!line.startsWith('data: ')) continue;
        const jsonStr = line.slice(6).trim();
        if (!jsonStr) continue;

        let data;
        try {
          data = JSON.parse(jsonStr);
        } catch { continue; }

        if (data.type === 'meta') {
          convId = data.conversation_id;
          currentConvId.value = convId;
        } else if (data.type === 'chunk') {
          aiMsg.content += data.content;
          fullAnswer += data.content;
          scrollToBottom();
        } else if (data.type === 'error') {
          aiMsg.content = '⚠️ 生成失败：' + data.error;
          scrollToBottom();
        } else if (data.type === 'done') {
          if (!aiMsg.content) {
            aiMsg.content = data.answer || '';
          }
        }
      }
    }

    if (convId && !conversations.value.some(c => c.id === convId)) {
      await loadConversations();
    }
    logsStore.addLog('对话', 'AI助手', `提问: ${text.slice(0, 30)}`);
  } catch (e) {
    aiMsg.content = '❌ 调用失败：' + (e.message || '网络异常');
    ElMessage.error('AI 调用失败：' + (e.message || ''));
  } finally {
    thinking.value = false;
    scrollToBottom();
  }
};

// ===== 快速提问 =====
const quickQuestions = [
  '设备无法连接怎么办？',
  '如何排查通信故障？',
  '授权码验证失败怎么处理？',
  '日志解析异常如何解决？'
];

const useQuickQuestion = (q) => {
  inputText.value = q;
  sendMessage();
};

// ===== 工具 =====
const scrollToBottom = async () => {
  await nextTick();
  if (messageListRef.value) {
    messageListRef.value.scrollTop = messageListRef.value.scrollHeight;
  }
};

const formatTime = () => {
  const d = new Date();
  return `${String(d.getHours()).padStart(2,'0')}:${String(d.getMinutes()).padStart(2,'0')}`;
};

const formatContent = (content) => {
  // 简单的文本格式化：保留换行
  return content;
};

const isOwn = (msg) => msg.role === 'user';

const getConversationTitle = (id) => {
  const conv = conversations.value.find(c => c.id === id);
  return conv ? conv.title : '新对话';
};

onMounted(() => {
  loadConversations();
  messages.value = [welcomeMsg];
});

const handleKeydown = (e) => {
  if (e.key === 'Enter' && !e.shiftKey && !e.isComposing) {
    e.preventDefault();
    sendMessage();
  }
};
</script>

<template>
  <div class="ai-chat-page">
    <!-- 左侧会话列表 -->
    <div class="chat-layout">
      <div class="sidebar">
        <button class="new-chat-btn" @click="newConversation">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <line x1="12" y1="5" x2="12" y2="19" /><line x1="5" y1="12" x2="19" y2="12" />
          </svg>
          新对话
        </button>
        <div class="conv-list">
          <div
            v-for="conv in conversations"
            :key="conv.id"
            class="conv-item"
            :class="{ active: currentConvId === conv.id }"
            @click="switchConversation(conv.id)"
          >
            <span class="conv-icon">💬</span>
            <span class="conv-title">{{ conv.title }}</span>
            <span class="conv-msgcount">{{ conv.msg_count }}</span>
          </div>
          <div v-if="conversations.length === 0" class="conv-empty">
            <p>暂无历史会话</p>
          </div>
        </div>
      </div>

      <!-- 右侧对话区域 -->
      <div class="chat-main">
        <div class="chat-header">
          <div class="chat-title">
            <span class="ai-badge">AI</span>
            <span>知识库智能助手</span>
            <span class="conv-name" v-if="currentConvId">· {{ getConversationTitle(currentConvId) }}</span>
          </div>
          <span class="model-tip">本地模型 · 数据不出内网</span>
        </div>

        <div class="message-list" ref="messageListRef">
          <div v-if="chatLoading" class="loading-hint">
            <div class="spinner"></div>
            <span>加载会话历史...</span>
          </div>

          <template v-for="(msg, idx) in messages" :key="idx">
            <div class="msg-row" :class="isOwn(msg) ? 'own' : 'ai'">
              <div class="msg-avatar" :class="isOwn(msg) ? 'own-avatar' : 'ai-avatar'">
                {{ isOwn(msg) ? '我' : 'AI' }}
              </div>
              <div class="msg-bubble" :class="isOwn(msg) ? 'own-bubble' : 'ai-bubble'">
                <div class="msg-content">{{ msg.content }}<span v-if="thinking && idx === messages.length - 1 && isOwn(msg) === false" class="cursor-blink">▌</span></div>
              </div>
            </div>
          </template>

          <div v-if="messages.length === 1 && !chatLoading" class="quick-questions">
            <div class="quick-title">💡 你可以问我：</div>
            <div class="quick-list">
              <button v-for="q in quickQuestions" :key="q" class="quick-btn" @click="useQuickQuestion(q)" :disabled="thinking">
                {{ q }}
              </button>
            </div>
          </div>
        </div>

        <div class="chat-input-area">
          <div class="input-box">
            <textarea
              v-model="inputText"
              placeholder="输入你的问题，基于知识库智能回答（Enter 发送，Shift+Enter 换行）"
              rows="2"
              @keydown="handleKeydown"
            ></textarea>
            <button class="send-btn" @click="sendMessage" :disabled="thinking || !inputText.trim()">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <line x1="22" y1="2" x2="11" y2="13" />
                <polygon points="22 2 15 22 11 13 2 9 22 2" />
              </svg>
              发送
            </button>
          </div>
          <div class="input-tip">回答由本地模型生成，基于知识库内容，仅供参考</div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.ai-chat-page {
  padding: 20px;
  height: calc(100vh - 110px);
}

.chat-layout {
  display: flex;
  gap: 16px;
  height: 100%;
}

/* 左侧会话列表 */
.sidebar {
  width: 240px;
  flex-shrink: 0;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 14px;
  display: flex;
  flex-direction: column;
  padding: 12px;
  overflow: hidden;
}

.new-chat-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 10px;
  background: linear-gradient(135deg, #165DFF, #4080FF);
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  margin-bottom: 12px;
  transition: all .2s;
}

.new-chat-btn:hover { transform: translateY(-1px); box-shadow: 0 4px 12px rgba(22,93,255,.3); }

.conv-list {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.conv-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  border-radius: 8px;
  cursor: pointer;
  transition: all .2s;
  color: var(--text-secondary);
}

.conv-item:hover { background: var(--bg-input); }

.conv-item.active {
  background: rgba(22,93,255,.1);
  color: #165DFF;
  font-weight: 500;
}

.conv-icon { font-size: 14px; }

.conv-title {
  flex: 1;
  font-size: 13px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.conv-msgcount {
  font-size: 11px;
  background: var(--bg-input);
  color: var(--text-tertiary);
  padding: 1px 7px;
  border-radius: 10px;
}

.conv-empty { text-align: center; padding: 30px 0; color: var(--text-tertiary); font-size: 13px; }

/* 主对话区 */
.chat-main {
  flex: 1;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 14px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-width: 0;
}

.chat-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 20px;
  border-bottom: 1px solid var(--border-color);
}

.chat-title {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
}

.ai-badge {
  background: linear-gradient(135deg, #165DFF, #722ED1);
  color: white;
  font-size: 11px;
  padding: 3px 8px;
  border-radius: 6px;
  font-weight: 700;
}

.conv-name { font-size: 13px; color: var(--text-tertiary); font-weight: 400; }

.model-tip {
  font-size: 12px;
  color: var(--text-tertiary);
  background: rgba(0,180,42,.1);
  padding: 4px 10px;
  border-radius: 20px;
}

.message-list {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.loading-hint {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 30px;
  color: var(--text-tertiary);
}

.spinner {
  width: 18px;
  height: 18px;
  border: 2px solid rgba(22,93,255,.2);
  border-top-color: #165DFF;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin { to { transform: rotate(360deg); } }

/* 消息气泡 */
.msg-row {
  display: flex;
  gap: 10px;
  align-items: flex-start;
}

.msg-row.own { flex-direction: row-reverse; }

.msg-avatar {
  width: 34px;
  height: 34px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  font-weight: 600;
  flex-shrink: 0;
}

.ai-avatar {
  background: linear-gradient(135deg, #165DFF, #722ED1);
  color: white;
}

.own-avatar {
  background: linear-gradient(135deg, #00B42A, #0FC6C2);
  color: white;
}

.msg-bubble {
  max-width: 72%;
  padding: 12px 16px;
  border-radius: 12px;
  font-size: 14px;
  line-height: 1.7;
  white-space: pre-wrap;
  word-break: break-word;
}

.ai-bubble {
  background: var(--bg-input);
  border: 1px solid var(--border-color);
  color: var(--text-primary);
  border-top-left-radius: 4px;
}

.own-bubble {
  background: linear-gradient(135deg, #165DFF, #3b6fff);
  color: white;
  border-top-right-radius: 4px;
}

.cursor-blink { animation: blink 1s step-end infinite; color: #165DFF; }

@keyframes blink { 50% { opacity: 0; } }

/* 快速提问 */
.quick-questions { text-align: center; padding: 30px 0 10px; }

.quick-title { font-size: 14px; color: var(--text-tertiary); margin-bottom: 14px; }

.quick-list {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  justify-content: center;
}

.quick-btn {
  padding: 10px 18px;
  background: var(--bg-input);
  border: 1px solid var(--border-color);
  border-radius: 20px;
  color: var(--text-secondary);
  font-size: 13px;
  cursor: pointer;
  transition: all .2s;
}

.quick-btn:hover:not(:disabled) {
  border-color: #165DFF;
  color: #165DFF;
  background: rgba(22,93,255,.06);
}

.quick-btn:disabled { opacity: .5; cursor: not-allowed; }

/* 输入区 */
.chat-input-area {
  padding: 14px 20px 12px;
  border-top: 1px solid var(--border-color);
}

.input-box {
  display: flex;
  align-items: flex-end;
  gap: 10px;
  background: var(--bg-input);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  padding: 10px 12px;
  transition: border-color .2s;
}

.input-box:focus-within { border-color: #165DFF; box-shadow: 0 0 0 3px rgba(22,93,255,.08); }

.input-box textarea {
  flex: 1;
  background: none;
  border: none;
  outline: none;
  resize: none;
  color: var(--text-primary);
  font-size: 14px;
  line-height: 1.6;
  font-family: inherit;
  max-height: 120px;
}

.send-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 18px;
  background: linear-gradient(135deg, #165DFF, #4080FF);
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all .2s;
  flex-shrink: 0;
}

.send-btn:hover:not(:disabled) { transform: translateY(-1px); box-shadow: 0 4px 12px rgba(22,93,255,.3); }
.send-btn:disabled { opacity: .5; cursor: not-allowed; transform: none; }

.input-tip {
  text-align: center;
  font-size: 12px;
  color: var(--text-tertiary);
  padding-top: 8px;
}

/* 响应式 */
@media (max-width: 768px) {
  .chat-layout { flex-direction: column; }
  .sidebar { width: 100%; height: 120px; flex-direction: row; gap: 8px; }
  .conv-list { flex-direction: row; overflow-x: auto; }
  .conv-item { white-space: nowrap; }
  .chat-main { flex: 1; }
  .msg-bubble { max-width: 85%; }
}
</style>
