<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { aiService } from '@/services/aiService'
import { useLocalDiagnosis } from '@/composables/useLocalDiagnosis'

const router = useRouter()

// ===== AI 对话 =====
const inputText = ref('')
const inputRef = ref(null)
const aiThinking = ref(false)
const aiReply = ref('')
const recentTasks = ref([])

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

  const cmd = parseQuickCmd(text)
  if (cmd) { router.push(cmd); inputText.value = ''; return }

  if (!aiService.getApiKey()) {
    aiReply.value = '⚠️ AI 服务尚未配置。\n\n请前往「系统管理 > 三方账号授权」页面配置 DeepSeek API Key。\n\n💡 或使用右侧「本地智能诊断」，无需 API Key，基于知识库匹配。'
    return
  }

  aiThinking.value = true
  aiReply.value = ''

  try {
    const result = await aiService.analyze(text, '理解用户意图，如果是查询类请求，请告诉用户如何使用平台功能。如果是指令类，请给出执行步骤。请用简洁中文回答。')
    aiReply.value = result.content
    recentTasks.value.unshift({ id: Date.now(), query: text, reply: result.content, time: new Date() })
    if (recentTasks.value.length > 10) recentTasks.value.pop()
  } catch (e) {
    aiReply.value = '❌ AI 调用失败：' + (e.message || '网络异常，请稍后重试')
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
          <h1 class="hb-title">全栈赋能工作台</h1>
          <p class="hb-subtitle">芯片 · 模块 · 设备 · 全链路打通</p>
        </div>
      </div>
      <div class="hb-right">
        <div v-for="s in stats" :key="s.label" class="stat-chip">
          <span class="sc-icon">{{ s.icon }}</span>
          <span class="sc-val">{{ s.value }}</span>
          <span class="sc-lbl">{{ s.label }}</span>
        </div>
      </div>
    </header>

    <!-- ===== 主体双栏 ===== -->
    <div class="wb-body">
      <!-- 左栏：AI 对话 + 场景入口 -->
      <div class="wb-main">
        <!-- AI 对话 -->
        <div class="card ai-chat">
          <div class="card-head">
            <span class="ch-dot"></span>
            <span>AI 智能助手</span>
            <span class="ch-badge">DeepSeek</span>
          </div>
          <div class="ai-input-area">
            <textarea
              ref="inputRef"
              v-model="inputText"
              class="ai-textarea"
              placeholder="描述你的需求，AI 帮你执行..."
              rows="2"
              @keydown="handleKeydown"
            ></textarea>
            <button class="ai-send" :class="{ loading: aiThinking }" :disabled="!inputText.trim() || aiThinking" @click="handleSend">
              <template v-if="aiThinking"><span class="spinner"></span></template>
              <template v-else>↵</template>
            </button>
          </div>
          <p class="ai-hint">Enter 发送 · Shift+Enter 换行 · 输入功能名可快速跳转</p>

          <!-- AI 回复 -->
          <div v-if="aiReply" class="ai-reply">
            <div class="ar-head">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/></svg>
              <span>AI 分析结果</span>
            </div>
            <div class="ar-body">{{ aiReply }}</div>
          </div>
        </div>

        <!-- 场景卡片 -->
        <div class="card scenario-area">
          <div class="card-head">
            <span>🔥 赋能场景</span>
            <span class="ch-badge" style="background:rgba(0,180,42,0.12);color:#00B42A">全功能</span>
          </div>
          <div class="sc-grid">
            <div v-for="s in roleScenarios" :key="s.id" class="sc-card" @click="fillScenario(s)">
              <span class="sc-icon">{{ s.icon }}</span>
              <div class="sc-info">
                <div class="sc-name">{{ s.title }}</div>
                <div class="sc-desc">{{ s.desc }}</div>
              </div>
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="9 18 15 12 9 6"/></svg>
            </div>
          </div>
        </div>

        <!-- 最近执行 -->
        <div v-if="recentTasks.length" class="card">
          <div class="card-head"><span>📝 最近执行</span></div>
          <div class="recent-list">
            <div v-for="t in recentTasks.slice(0, 5)" :key="t.id" class="recent-row">
              <span class="rq">{{ t.query }}</span>
              <span class="rp">{{ t.reply?.slice(0, 60) }}...</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 右栏：本地诊断 + 快捷入口 -->
      <div class="wb-side">
        <!-- 本地智能诊断 -->
        <div class="card local-dx">
          <div class="card-head">
            <span>⚡ 本地智能诊断</span>
            <span class="ch-badge" style="background:rgba(0,210,172,0.12);color:#008A6E">零Token</span>
          </div>
          <div class="ld-input-wrap">
            <input
              v-model="localQuery"
              class="local-input"
              placeholder="输入问题、日志片段或故障描述..."
              @keydown="handleLocalKeydown"
            />
            <button class="ld-btn" :class="{ loading: localThinking }" :disabled="!localQuery.trim() || localThinking" @click="runLocalDiagnosis">
              {{ localThinking ? '⏳' : '🔎' }}
            </button>
          </div>
          <p class="ld-hint">基于本地知识库 · 不消耗API Token · 毫秒级响应</p>

          <!-- 诊断结果 -->
          <div v-if="localResult" class="ld-results">
            <!-- 日志模式匹配 -->
            <div v-if="localResult.patterns.length" class="ld-section">
              <div class="ld-sec-title">📌 日志异常匹配</div>
              <div v-for="(p, i) in localResult.patterns.slice(0, 5)" :key="i" class="ld-pattern" :class="p.severity">
                <span class="ldp-dot"></span>
                <span class="ldp-text">{{ p.line?.slice(0, 80) }}</span>
              </div>
            </div>

            <!-- 知识库匹配 -->
            <div v-if="localResult.results.length" class="ld-section">
              <div class="ld-sec-title">💡 知识库匹配 ({{ localResult.results.length }}条)</div>
              <div v-for="(r, i) in localResult.results.slice(0, 4)" :key="r.id" class="ld-match" :class="{ best: i === 0 && r.score > 10 }">
                <div class="ldm-head">
                  <span class="ldm-sev" :style="{ background: r.severityInfo?.color }">{{ r.severityInfo?.label }}</span>
                  <span class="ldm-cat">{{ r.categoryInfo?.name }}</span>
                  <span class="ldm-score">{{ Math.round(r.score) }}分</span>
                </div>
                <div class="ldm-q">{{ r.question }}</div>
              </div>
            </div>

            <!-- 推荐工具 -->
            <div v-if="localResult.tools?.length" class="ld-section">
              <div class="ld-sec-title">🔧 推荐工具</div>
              <div class="ld-tools">
                <button v-for="t in localResult.tools" :key="t.path" class="ld-tool-btn" @click="router.push(t.path)">
                  {{ t.name }}
                </button>
              </div>
            </div>

            <!-- 无匹配 -->
            <div v-if="!localResult.results.length && !localResult.patterns.length" class="ld-empty">
              未匹配到相关知识，建议使用 AI 智能助手深入分析
            </div>
          </div>
        </div>

        <!-- 快捷工具 -->
        <div class="card">
          <div class="card-head"><span>🔧 快捷工具</span></div>
          <div class="qt-grid">
            <div class="qt-item" @click="router.push('/auth/third-party')">🔐 三方授权</div>
            <div class="qt-item" @click="router.push('/log/decrypt')">🔓 日志解密</div>
            <div class="qt-item" @click="router.push('/data/query')">📊 数据查询</div>
            <div class="qt-item" @click="router.push('/tools/translate')">🌐 AI翻译</div>
            <div class="qt-item" @click="router.push('/tools/converter')">📄 格式转换</div>
            <div class="qt-item" @click="router.push('/knowledge/rag')">💡 知识库</div>
            <div class="qt-item" @click="router.push('/appstore/shengjing')">📦 盛景应用</div>
            <div class="qt-item" @click="router.push('/appstore/push')">📤 推送管理</div>
            <div class="qt-item" @click="router.push('/tools/assistant')">🤖 运维助手</div>
            <div class="qt-item" @click="router.push('/trace/analysis')">🔗 批次追溯</div>
          </div>
        </div>

        <!-- 硬件能力 -->
        <div class="card hw-card">
          <div class="card-head"><span>🔌 硬件直连能力</span></div>
          <div class="hw-row">
            <div class="hw-item">💾 芯片管理<span class="hw-tag">BLE</span></div>
            <div class="hw-item">📡 模块通信<span class="hw-tag">MQTT</span></div>
            <div class="hw-item">📱 设备OTA<span class="hw-tag">固件</span></div>
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

/* ===== 双栏 ===== */
.wb-body { display: flex; gap: 18px; align-items: flex-start; }
.wb-main { flex: 1; min-width: 0; }
.wb-side { width: 340px; flex-shrink: 0; }

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

/* AI 回复 */
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
@media screen and (max-width: 1100px) {
  .wb-side { width: 280px; }
  .sc-grid { grid-template-columns: repeat(2, 1fr); }
}
@media screen and (max-width: 900px) {
  .wb-body { flex-direction: column; }
  .wb-side { width: 100%; }
  .sc-grid { grid-template-columns: 1fr; }
  .role-tabs { overflow-x: auto; }
  .hb-right { display: none; }
}
</style>
