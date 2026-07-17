<template>
  <div class="ai-panel" :class="{ expanded }">
    <div class="ai-panel-header" @click="expanded = !expanded">
      <span class="ai-dot"></span>
      <span class="ai-title">AI 分析结果</span>
      <svg class="ai-arrow" :class="{ rotated: expanded }" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="6 9 12 15 18 9"/></svg>
    </div>
    <div class="ai-panel-body" v-show="expanded">
      <div v-if="loading" class="ai-loading">
        <span class="ai-spinner"></span> AI 正在分析中...
      </div>
      <div v-else-if="error" class="ai-error">{{ error }}</div>
      <div v-else-if="content" class="ai-content" v-html="rendered"></div>
      <div v-else class="ai-placeholder">点击上方「AI 分析」按钮获取智能解读</div>
    </div>
    <div class="ai-panel-actions" v-show="expanded && content">
      <button class="ai-btn" @click="refresh">刷新分析</button>
      <button class="ai-btn primary" @click="copy">复制结论</button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'

const props = defineProps({
  content: String,
  loading: Boolean,
  error: String
})

const emit = defineEmits(['analyze', 'copy'])
const expanded = ref(false)

const rendered = computed(() => {
  if (!props.content) return ''
  return props.content
    .replace(/\n/g, '<br>')
    .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
    .replace(/^- (.+)$/gm, '<li>$1</li>')
    .replace(/(<li>.*?<\/li>)/gs, '<ul>$1</ul>')
})

function refresh() { emit('analyze') }
function copy() { emit('copy', props.content) }
</script>

<style scoped>
.ai-panel {
  border: 1px solid rgba(22, 93, 255, 0.2);
  border-radius: 12px;
  background: linear-gradient(135deg, rgba(22, 93, 255, 0.04) 0%, rgba(15, 76, 208, 0.02) 100%);
  overflow: hidden;
  transition: all 0.3s;
}
.ai-panel.expanded { border-color: rgba(22, 93, 255, 0.4); }
.ai-panel-header {
  display: flex; align-items: center; gap: 8px; padding: 14px 18px;
  cursor: pointer; user-select: none; font-weight: 500; color: #165DFF; font-size: 14px;
}
.ai-dot { width: 8px; height: 8px; border-radius: 50%; background: #165DFF; animation: pulse 2s infinite; }
@keyframes pulse { 0%, 100% { opacity: 1 } 50% { opacity: 0.4 } }
.ai-title { flex: 1; }
.ai-arrow { transition: transform 0.2s; color: #165DFF; }
.ai-arrow.rotated { transform: rotate(180deg); }
.ai-panel-body { padding: 0 18px 16px 18px; font-size: 13px; line-height: 1.8; color: var(--text-secondary); }
.ai-loading { display: flex; align-items: center; gap: 8px; color: #165DFF; }
.ai-spinner { width: 14px; height: 14px; border: 2px solid rgba(22,93,255,0.2); border-top-color: #165DFF; border-radius: 50%; animation: spin 0.8s linear infinite; }
@keyframes spin { to { transform: rotate(360deg) } }
.ai-error { color: #F53F3F; }
.ai-placeholder { color: var(--text-tertiary); font-style: italic; }
.ai-panel-actions { display: flex; gap: 8px; padding: 0 18px 14px; }
.ai-btn { padding: 6px 14px; border-radius: 6px; border: 1px solid rgba(22,93,255,0.2); background: transparent; color: #165DFF; font-size: 12px; cursor: pointer; }
.ai-btn.primary { background: #165DFF; color: #fff; border-color: #165DFF; }
</style>