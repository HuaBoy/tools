<script setup>
import { ref, reactive, nextTick, onMounted, computed } from 'vue';
import { ElMessage } from 'element-plus';
import { useLogsStore } from '@/stores/logs';
import {
  listConversations,
  getConversationMessages,
  deleteConversation,
  listFAQs
} from '@/api/knowledge';
import { ElMessageBox } from 'element-plus';

const logsStore = useLogsStore();

// ===== 会话状态 =====
const conversations = ref([]);
const currentConvId = ref(0);
const messages = ref([]);
const inputText = ref('');
const thinking = ref(false);
const chatLoading = ref(false);
const messageListRef = ref(null);

// 本地 FAQ 快捷提问
const faqQuestions = ref([]);
const quickFallback = [
  '设备无法连接怎么办？',
  '如何排查通信故障？',
  '授权码验证失败怎么处理？',
  '日志解析异常如何解决？'
];

// 默认欢迎消息（客服口吻）
const welcomeMsg = {
  role: 'assistant',
  content: '您好，我是本产品的智能客服助手 👩‍💻\n\n我可以基于本地知识库（操作手册、常见问题、设备资料）为您解答：\n· 设备使用与故障排查\n· 操作流程指引\n· 常见问题解答\n· 售后与运维支持\n\n请问有什么可以帮您？'
};

const quickQuestions = computed(() =>
  faqQuestions.value.length ? faqQuestions.value : quickFallback
);

const loadConversations = async () => {
  try {
    conversations.value = await listConversations();
  } catch (e) {
    console.warn('加载会话失败:', e);
  }
};

// 加载本地 FAQ 作为快捷提问（以本地数据为准）
const loadFaqQuestions = async () => {
  try {
    const { data } = await listFAQs({ pageSize: 8 });
    if (data && data.length) {
      faqQuestions.value = data.map(f => f.question).filter(Boolean).slice(0, 6);
    }
  } catch (e) {
    console.warn('加载 FAQ 失败:', e);
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

// ===== 删除会话 =====
const deleteConv = async (conv, e) => {
  e.stopPropagation();
  try {
    await ElMessageBox.confirm(
      `确定删除会话「${conv.title || '新对话'}」？删除后不可恢复。`,
      '删除确认',
      { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' }
    );
  } catch {
    return;
  }
  try {
    await deleteConversation(conv.id);
    ElMessage.success('已删除');
    if (currentConvId.value === conv.id) {
      newConversation();
    }
    await loadConversations();
  } catch (e2) {
    ElMessage.error('删除失败：' + (e2.message || ''));
  }
};

// ===== 发送消息（SSE 流式，mode=customer_service 走本地知识库） =====
const sendMessage = async () => {
  const text = inputText.value.trim();
  if (!text || thinking.value) return;

  messages.value.push({ role: 'user', content: text });
  inputText.value = '';
  scrollToBottom();

  thinking.value = true;
  const aiMsg = { role: 'assistant', content: '', sources: [] };
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
        stream: true,
        mode: 'customer_service'
      })
    });

    if (!resp.ok) {
      const err = await resp.json().catch(() => ({}));
      throw new Error(err.message || '请求失败 (' + resp.status + ')');
    }

    let convId = currentConvId.value;
    let isNewConv = false;

    const reader = resp.body.getReader();
    const decoder = new TextDecoder();
    let buffer = '';

    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      buffer += decoder.decode(value, { stream: true });

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
          isNewConv = data.is_new || false;
          currentConvId.value = convId;
        } else if (data.type === 'chunk') {
          aiMsg.content += data.content;
          scrollToBottom();
        } else if (data.type === 'error') {
          aiMsg.content = '⚠️ 生成失败：' + data.error;
          scrollToBottom();
        } else if (data.type === 'done') {
          if (!aiMsg.content) aiMsg.content = data.answer || '';
          // 回传的本地检索来源（以本地数据为准的证据）
          if (data.sources && Array.isArray(data.sources)) {
            aiMsg.sources = data.sources;
          }
        }
      }
    }

    if (isNewConv && convId) {
      await loadConversations();
    }
    logsStore.addLog('客服', '智能客服', `提问: ${text.slice(0, 30)}`);
  } catch (e) {
    aiMsg.content = '❌ 调用失败：' + (e.message || '网络异常');
    ElMessage.error('客服调用失败：' + (e.message || ''));
  } finally {
    thinking.value = false;
    scrollToBottom();
  }
};

const useQuickQuestion = (q) => {
  inputText.value = q;
  sendMessage();
};

// ===== 来源引用展示 =====
const formatSourceTitle = (s) => {
  if (s.source === 'faq') return s.question || '常见问题';
  return s.document_name || '本地文档';
};
const formatScore = (s) => {
  const v = Number(s.score || 0);
  return Math.round(v * 100) + '%';
};

// ===== 工具 =====
const scrollToBottom = async () => {
  await nextTick();
  if (messageListRef.value) {
    messageListRef.value.scrollTop = messageListRef.value.scrollHeight;
  }
};

const isOwn = (msg) => msg.role === 'user';
const getConversationTitle = (id) => {
  const conv = conversations.value.find(c => c.id === id);
  if (!conv) return '新对话';
  const title = conv.title || '新对话';
  return title.length > 16 ? title.slice(0, 16) + '...' : title;
};

onMounted(() => {
  loadConversations();
  loadFaqQuestions();
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
    <div class="chat-layout">
      <!-- 左侧会话列表 -->
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
            <span class="conv-title">{{ conv.title && conv.title.length > 16 ? conv.title.slice(0, 16) + '...' : (conv.title || '新对话') }}</span>
            <span class="conv-msgcount">{{ conv.msg_count }}</span>
            <button class="conv-delete" @click="deleteConv(conv, $event)" title="删除会话">
              <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2"/>
              </svg>
            </button>
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
            <span class="cs-badge">客服</span>
            <span>智能客服问答</span>
            <span class="conv-name" v-if="currentConvId">· {{ getConversationTitle(currentConvId) }}</span>
          </div>
          <span class="model-tip">本地知识库 · 数据不出内网</span>
        </div>

        <div class="message-list" ref="messageListRef">
          <div v-if="chatLoading" class="loading-hint">
            <div class="spinner"></div>
            <span>加载会话历史...</span>
          </div>

          <template v-for="(msg, idx) in messages" :key="idx">
            <div class="msg-row" :class="isOwn(msg) ? 'own' : 'ai'">
              <div class="msg-avatar" :class="isOwn(msg) ? 'own-avatar' : 'cs-avatar'">
                {{ isOwn(msg) ? '我' : '客服' }}
              </div>
              <div class="msg-bubble" :class="isOwn(msg) ? 'own-bubble' : 'ai-bubble'">
                <div class="msg-content">{{ msg.content }}<span v-if="thinking && idx === messages.length - 1 && !isOwn(msg)" class="cursor-blink">▌</span></div>

                <!-- 本地知识库来源引用（以本地数据为准的证据） -->
                <div
                  v-if="!isOwn(msg) && msg.sources && msg.sources.length"
                  class="source-box"
                >
                  <div class="source-title">📚 参考本地知识库：</div>
                  <div class="source-chips">
                    <span
                      v-for="(s, si) in msg.sources"
                      :key="si"
                      class="source-chip"
                      :class="s.source === 'faq' ? 'chip-faq' : 'chip-doc'"
                      :title="formatSourceTitle(s)"
                    >
                      <span class="chip-icon">{{ s.source === 'faq' ? '❓' : '📄' }}</span>
                      <span class="chip-text">{{ formatSourceTitle(s) }}</span>
                      <span class="chip-score">{{ formatScore(s) }}</span>
                    </span>
                  </div>
                </div>
              </div>
            </div>
          </template>

          <div v-if="messages.length === 1 && !chatLoading" class="quick-questions">
            <div class="quick-title">💡 您可以这样问我：</div>
            <div class="quick-list">
              <button v-for="q in quickQuestions" :key="q" class="quick-btn" @click="useQuickQuestion(q)" :disabled="thinking">
                {{ q }}
              </button>
            </div>
            <div class="quick-hint">以上问题来自本地常见问题库</div>
          </div>
        </div>

        <div class="chat-input-area">
          <div class="input-box">
            <textarea
              v-model="inputText"
              placeholder="描述您的问题，客服将基于本地知识库为您解答（Enter 发送，Shift+Enter 换行）"
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
          <div class="input-tip">回答由本地模型生成，均基于本地知识库内容，数据不出内网</div>
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
  background: linear-gradient(135deg, #00B42A, #0FC6C2);
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  margin-bottom: 12px;
  transition: all .2s;
}

.new-chat-btn:hover { transform: translateY(-1px); box-shadow: 0 4px 12px rgba(0,180,42,.3); }

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
.conv-item:hover .conv-delete { opacity: 1; }

.conv-delete {
  opacity: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  border: none;
  background: transparent;
  border-radius: 6px;
  color: var(--text-tertiary);
  cursor: pointer;
  flex-shrink: 0;
  transition: all .15s;
}

.conv-delete:hover { background: rgba(255,77,79,.12); color: #FF4D4F; }

.conv-item.active { background: rgba(0,180,42,.1); color: #00B42A; font-weight: 500; }

.conv-icon { font-size: 14px; }
.conv-title { flex: 1; font-size: 13px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.conv-msgcount { font-size: 11px; background: var(--bg-input); color: var(--text-tertiary); padding: 1px 7px; border-radius: 10px; }
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

.cs-badge {
  background: linear-gradient(135deg, #00B42A, #0FC6C2);
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

.loading-hint { display: flex; align-items: center; justify-content: center; gap: 10px; padding: 30px; color: var(--text-tertiary); }
.spinner { width: 18px; height: 18px; border: 2px solid rgba(0,180,42,.2); border-top-color: #00B42A; border-radius: 50%; animation: spin 1s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }

/* 消息气泡 */
.msg-row { display: flex; gap: 10px; align-items: flex-start; }
.msg-row.own { flex-direction: row-reverse; }

.msg-avatar {
  width: 34px; height: 34px; border-radius: 10px;
  display: flex; align-items: center; justify-content: center;
  font-size: 13px; font-weight: 600; flex-shrink: 0;
}

.cs-avatar { background: linear-gradient(135deg, #00B42A, #0FC6C2); color: white; }
.own-avatar { background: linear-gradient(135deg, #165DFF, #3b6fff); color: white; }

.msg-bubble {
  max-width: 72%;
  padding: 12px 16px;
  border-radius: 12px;
  font-size: 14px;
  line-height: 1.7;
  white-space: pre-wrap;
  word-break: break-word;
}

.ai-bubble { background: var(--bg-input); border: 1px solid var(--border-color); color: var(--text-primary); border-top-left-radius: 4px; }
.own-bubble { background: linear-gradient(135deg, #165DFF, #3b6fff); color: white; border-top-right-radius: 4px; }

.cursor-blink { animation: blink 1s step-end infinite; color: #00B42A; }
@keyframes blink { 50% { opacity: 0; } }

/* 来源引用 */
.source-box {
  margin-top: 10px;
  padding: 10px 12px;
  background: rgba(0,180,42,.06);
  border: 1px solid rgba(0,180,42,.18);
  border-radius: 10px;
}
.source-title { font-size: 12px; color: #00B42A; font-weight: 600; margin-bottom: 8px; }
.source-chips { display: flex; flex-wrap: wrap; gap: 6px; }
.source-chip {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  max-width: 240px;
  padding: 4px 9px;
  border-radius: 14px;
  font-size: 12px;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
}
.chip-faq { border-color: rgba(22,93,255,.3); }
.chip-doc { border-color: rgba(0,180,42,.3); }
.chip-icon { flex-shrink: 0; }
.chip-text { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; color: var(--text-secondary); }
.chip-score { flex-shrink: 0; color: var(--text-tertiary); font-size: 11px; }

/* 快速提问 */
.quick-questions { text-align: center; padding: 30px 0 10px; }
.quick-title { font-size: 14px; color: var(--text-tertiary); margin-bottom: 14px; }
.quick-list { display: flex; flex-wrap: wrap; gap: 10px; justify-content: center; }
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
.quick-btn:hover:not(:disabled) { border-color: #00B42A; color: #00B42A; background: rgba(0,180,42,.06); }
.quick-btn:disabled { opacity: .5; cursor: not-allowed; }
.quick-hint { margin-top: 12px; font-size: 11px; color: var(--text-tertiary); }

/* 输入区 */
.chat-input-area { padding: 14px 20px 12px; border-top: 1px solid var(--border-color); }
.input-box { display: flex; align-items: flex-end; gap: 10px; background: var(--bg-input); border: 1px solid var(--border-color); border-radius: 12px; padding: 10px 12px; transition: border-color .2s; }
.input-box:focus-within { border-color: #00B42A; box-shadow: 0 0 0 3px rgba(0,180,42,.08); }
.input-box textarea { flex: 1; background: none; border: none; outline: none; resize: none; color: var(--text-primary); font-size: 14px; line-height: 1.6; font-family: inherit; max-height: 120px; }
.send-btn {
  display: inline-flex; align-items: center; gap: 6px; padding: 8px 18px;
  background: linear-gradient(135deg, #00B42A, #0FC6C2); color: white; border: none;
  border-radius: 8px; font-size: 13px; font-weight: 500; cursor: pointer; transition: all .2s; flex-shrink: 0;
}
.send-btn:hover:not(:disabled) { transform: translateY(-1px); box-shadow: 0 4px 12px rgba(0,180,42,.3); }
.send-btn:disabled { opacity: .5; cursor: not-allowed; transform: none; }
.input-tip { text-align: center; font-size: 12px; color: var(--text-tertiary); padding-top: 8px; }

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
