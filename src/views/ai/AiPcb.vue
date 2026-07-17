<script setup>
import { ref, onMounted } from 'vue'

const checklist = ref({
  eda: false,
  gateway: false,
  skill: false
})

// 本机 Bridge 状态：idle | checking | down | up(未连EDA) | ready(已连EDA)
const bridgeState = ref('idle')
const bridgePort = ref(null)

const links = {
  eda: 'https://lceda.cn/',
  gateway: 'https://ext.lceda.cn/item/oshwhub/run-api-gateway',
  skill: 'https://github.com/easyeda/easyeda-api-skill',
  apiDoc: 'https://docs.lceda.cn/cn/API/EasyEDA-API/index.html'
}

function openEda() {
  // 浏览器无法直接拉起本机桌面程序，故打开官网（网页版 EDA）作入口
  window.open('https://lceda.cn/editor', '_blank', 'noopener')
}

// Bridge 监听 49620-49629，启动时自动选可用端口，故需扫描整段
const BRIDGE_PORTS = Array.from({ length: 10 }, (_, i) => 49620 + i)

async function probePort(port) {
  const ctrl = new AbortController()
  const timer = setTimeout(() => ctrl.abort(), 1500)
  try {
    const r = await fetch(`http://127.0.0.1:${port}/health`, { signal: ctrl.signal })
    if (!r.ok) return null
    const j = await r.json()
    return j.service === 'easyeda-bridge' ? j : null
  } catch {
    return null
  } finally {
    clearTimeout(timer)
  }
}

async function checkBridge() {
  bridgeState.value = 'checking'
  for (const port of BRIDGE_PORTS) {
    const j = await probePort(port)
    if (j) {
      bridgePort.value = port
      bridgeState.value = j.edaConnected ? 'ready' : 'up'
      // 探针能确认的部分自动勾选：EDA 已打开 + 网关已连
      checklist.value.eda = true
      checklist.value.gateway = !!j.edaConnected
      return
    }
  }
  bridgeState.value = 'down'
  bridgePort.value = null
  checklist.value.eda = false
  checklist.value.gateway = false
}

onMounted(checkBridge)
</script>

<template>
  <div class="ai-pcb-page">
    <header class="page-head">
      <h1>AI-PCB 设计助手</h1>
      <p class="subtitle">基于嘉立创官方 AI 能力，对话式绘制原理图与 PCB</p>
    </header>

    <div class="notice">
      本功能采用<strong>嘉立创官方方案</strong>（easyeda-api-skill + Run API Gateway），
      由你本机的嘉立创 EDA 专业版与你的 AI 编程工具（Claude Code / Cursor / QwenCode 等）协作完成，
      本平台仅作入口与引导，<strong>不涉及后端服务与云端存储</strong>。
    </div>

    <section class="card">
      <h2>① 打开嘉立创 EDA 专业版</h2>
      <p>请确保本机已手动打开「嘉立创 EDA 专业版」桌面客户端（浏览器无法自动拉起本机程序）。</p>
      <div class="actions">
        <button class="btn primary" @click="openEda">打开嘉立创 EDA 官网</button>
        <a class="btn ghost" :href="links.eda" target="_blank" rel="noopener">前往官网下载</a>
      </div>

    </section>

    <section class="card">
      <h2>② 环境自检</h2>
      <p class="hint">页面会自动探测你本机运行的 Bridge 服务（端口 49620–49629）。</p>

      <div class="status" :class="bridgeState">
        <template v-if="bridgeState === 'checking'">🔄 正在检测本机 Bridge 服务…</template>
        <template v-else-if="bridgeState === 'ready'">✅ 环境就绪：Bridge 运行中（端口 {{ bridgePort }}）且 EDA 已连接，回到你的 AI 工具开始对话即可。</template>
        <template v-else-if="bridgeState === 'up'">⚠️ Bridge 已运行（端口 {{ bridgePort }}），但 EDA 未连接。请打开嘉立创 EDA 专业版并启用「Run API Gateway」扩展。</template>
        <template v-else>❌ 未检测到本机 Bridge 服务。请确认已在本地运行 <code>npm run server</code>（已配置开机自启）；若本页通过 HTTPS 访问，浏览器会拦截对本机 http 端口的检测（混合内容），请改用 HTTP 访问本页。</template>
      </div>

      <label class="check"><input type="checkbox" v-model="checklist.eda" /> 已安装并打开嘉立创 EDA 专业版</label>
      <label class="check"><input type="checkbox" v-model="checklist.gateway" /> 已在 EDA 扩展广场安装并启用「Run API Gateway」扩展</label>
      <label class="check"><input type="checkbox" v-model="checklist.skill" /> 已在 AI 工具中安装 easyeda-api-skill</label>

      <div class="actions" style="margin-top: 10px;">
        <button class="btn ghost" :disabled="bridgeState === 'checking'" @click="checkBridge">重新检测</button>
      </div>
    </section>

    <section class="card">
      <h2>③ 三步开始对话画 PCB</h2>
      <ol class="steps">
        <li>在 EDA 中打开（或新建）目标工程，并启用 <a :href="links.gateway" target="_blank" rel="noopener">Run API Gateway</a> 扩展，保持客户端运行。</li>
        <li>在你的 AI 编程工具（Claude Code / Cursor / QwenCode 等）中加载 <a :href="links.skill" target="_blank" rel="noopener">easyeda-api-skill</a>（按仓库 README 配置）。</li>
        <li>直接对话，例如：“画一个基于 ESP32 的最小系统原理图，并布局到 50×40mm 的 PCB”，AI 会通过网关操作你的 EDA 完成绘制。</li>
      </ol>
    </section>

    <section class="card links">
      <h2>官方资源</h2>
      <ul>
        <li><a :href="links.eda" target="_blank" rel="noopener">嘉立创 EDA 官网</a></li>
        <li><a :href="links.gateway" target="_blank" rel="noopener">Run API Gateway 扩展（扩展广场）</a></li>
        <li><a :href="links.skill" target="_blank" rel="noopener">easyeda-api-skill（GitHub）</a></li>
        <li><a :href="links.apiDoc" target="_blank" rel="noopener">嘉立创 EDA API 文档</a></li>
      </ul>
    </section>

    <div class="tip">提示：AI 绘制的原理图 / PCB 直接生成在你本机 EDA 的当前工程中，记得及时保存与 DRC 检查。</div>
  </div>
</template>

<style scoped>
.ai-pcb-page { max-width: 860px; margin: 0 auto; padding: 28px 20px 60px; color: #1d2129; }
.page-head h1 { font-size: 24px; margin: 0 0 6px; }
.subtitle { color: #86909c; margin: 0 0 18px; font-size: 14px; }
.notice { background: #eef4ff; border: 1px solid #cfe0ff; color: #1d4ed8; padding: 12px 16px; border-radius: 10px; font-size: 13px; line-height: 1.7; margin-bottom: 20px; }
.card { background: #fff; border: 1px solid #ebedf0; border-radius: 12px; padding: 18px 20px; margin-bottom: 16px; }
.card h2 { font-size: 16px; margin: 0 0 12px; }
.card p { font-size: 13px; color: #4e5969; line-height: 1.7; margin: 0 0 10px; }
.hint { color: #86909c !important; font-size: 12px !important; }
.actions { display: flex; gap: 12px; flex-wrap: wrap; margin-top: 6px; }
.btn { display: inline-flex; align-items: center; justify-content: center; height: 36px; padding: 0 16px; border-radius: 8px; font-size: 13px; cursor: pointer; text-decoration: none; border: 1px solid transparent; }
.btn.primary { background: #165DFF; color: #fff; }
.btn.primary:hover { background: #0F4CD0; }
.btn.ghost { background: #fff; border-color: #e5e6eb; color: #4e5969; }
.btn.ghost:hover { border-color: #165DFF; color: #165DFF; }
.check { display: flex; align-items: center; gap: 8px; font-size: 13px; color: #4e5969; margin: 8px 0; cursor: pointer; }
.check input { width: 16px; height: 16px; }
.ok { color: #00b42a; font-weight: 600; margin-top: 8px; }
.status { padding: 12px 14px; border-radius: 10px; font-size: 13px; line-height: 1.7; margin-bottom: 14px; border: 1px solid transparent; }
.status.checking { background: #f2f3f5; border-color: #e5e6eb; color: #4e5969; }
.status.ready { background: #e8ffea; border-color: #b7ebc4; color: #00871c; }
.status.up { background: #fff7e8; border-color: #ffe0a3; color: #a05b00; }
.status.down { background: #fff1f0; border-color: #ffccc7; color: #cf1322; }
.status code { background: rgba(0,0,0,.05); padding: 1px 5px; border-radius: 4px; font-size: 12px; }
.steps { margin: 0; padding-left: 20px; }
.steps li { font-size: 13px; color: #4e5969; line-height: 1.8; margin-bottom: 8px; }
.links ul { list-style: none; padding: 0; margin: 0; }
.links li { margin: 8px 0; }
a { color: #165DFF; text-decoration: none; }
a:hover { text-decoration: underline; }
.tip { font-size: 12px; color: #86909c; margin-top: 4px; }
</style>
