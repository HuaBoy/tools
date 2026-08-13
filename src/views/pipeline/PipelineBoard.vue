<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import GlassCard from '@/components/GlassCard.vue'
import { getPipelineBoard } from '@/api/pipeline'

const isLoading = ref(false)
const board = ref({ requirements: [], stages: [], stats: {}, kanoStats: {} })

// 环节定义（员工 A~G）
const stageDefs = [
  { key: 'A', name: '需求分析', emp: '员工A' },
  { key: 'B', name: '产品设计', emp: '员工B' },
  { key: 'C', name: '前端开发', emp: '员工C' },
  { key: 'D', name: '后端开发', emp: '员工D' },
  { key: 'E', name: '嵌入式', emp: '员工E' },
  { key: 'F', name: '测试', emp: '员工F' },
  { key: 'G', name: '验证', emp: '员工G' }
]

const statusDict = {
  draft: { label: '草稿', color: '#86909C' },
  backlog: { label: '已评估', color: '#0FC6C2' },
  prd: { label: 'PRD中', color: '#165DFF' },
  dev: { label: '开发中', color: '#722ED1' },
  test: { label: '测试中', color: '#F77234' },
  verify: { label: '验证中', color: '#F7BA1E' },
  review: { label: '待审核', color: '#F53F3F' },
  rework: { label: '返工中', color: '#D91AD9' },
  done: { label: '已完结', color: '#00B42A' },
  rejected: { label: '已淘汰', color: '#C9CDD4' }
}
const kanoDict = { M: '基本型', O: '期望型', A: '兴奋型', I: '无差异', R: '反向型' }
const priorityDict = { P0: 'P0', P1: 'P1', P2: 'P2' }

const stageStatusDict = {
  todo: { label: '待办', color: '#C9CDD4' },
  doing: { label: '进行中', color: '#165DFF' },
  done: { label: '已完成', color: '#00B42A' },
  rework: { label: '返工', color: '#F53F3F' },
  na: { label: '不适用', color: '#E5E6EB' }
}

// 需求 → 环节状态映射
const stageMap = computed(() => {
  const map = {}
  board.value.stages.forEach((s) => {
    if (!map[s.reqId]) map[s.reqId] = {}
    map[s.reqId][s.stage] = s.status
  })
  return map
})

const statCards = computed(() => {
  const order = ['review', 'rework', 'dev', 'verify', 'test', 'prd', 'backlog', 'draft', 'done', 'rejected']
  return order.map((k) => ({
    key: k,
    label: (statusDict[k] || {}).label || k,
    color: (statusDict[k] || {}).color,
    count: board.value.stats[k] || 0
  }))
})

const load = async () => {
  isLoading.value = true
  try {
    const resp = await getPipelineBoard()
    board.value = resp
  } catch (e) {
    ElMessage.error('加载失败: ' + e.message)
  } finally {
    isLoading.value = false
  }
}

onMounted(load)
</script>

<template>
  <GlassCard class="pipeline-board">
    <div class="pb-header">
      <div class="pb-title">
        <h2>流水线看板</h2>
        <p>可视化每个需求在 A~G 七个 AI 员工环节中的实时流转状态</p>
      </div>
      <button class="btn-ghost" @click="load">刷新</button>
    </div>

    <div class="stat-grid">
      <div v-for="s in statCards" :key="s.key" class="stat-card">
        <div class="stat-count" :style="{ color: s.color }">{{ s.count }}</div>
        <div class="stat-label">{{ s.label }}</div>
      </div>
    </div>

    <div class="board-list" v-loading="isLoading">
      <div v-for="req in board.requirements" :key="req.id" class="board-card">
        <div class="bc-head">
          <div class="bc-left">
            <span class="req-no">{{ req.reqNo }}</span>
            <span class="req-title">{{ req.title }}</span>
            <span class="tag kano" :class="'kano-' + (req.kanoCategory || 'O')">{{ kanoDict[req.kanoCategory] || req.kanoCategory }}</span>
            <span class="tag prio" :class="'prio-' + (req.priority || 'P2')">{{ priorityDict[req.priority] || req.priority }}</span>
            <span v-for="t in (req.aarrrImpacts || [])" :key="t" class="tag aarrr">{{ t }}</span>
          </div>
          <span class="status-pill" :style="{ background: (statusDict[req.status] || {}).color }">{{ (statusDict[req.status] || {}).label || req.status }}</span>
        </div>

        <div class="stage-track">
          <div v-for="(st, idx) in stageDefs" :key="st.key" class="stage-node">
            <div class="stage-line" :class="{ 'has-flow': idx < 6 }">
              <div class="stage-dot" :class="(stageMap[req.id] && stageMap[req.id][st.key]) || 'todo'"
                   :style="'todo' === ((stageMap[req.id] && stageMap[req.id][st.key]) || 'todo') ? {} : { background: stageStatusDict[(stageMap[req.id] && stageMap[req.id][st.key]) || 'todo'].color }">
                {{ st.key }}
              </div>
            </div>
            <div class="stage-name">{{ st.name }}</div>
            <div class="stage-emp">{{ st.emp }}</div>
            <div class="stage-status" :style="{ color: stageStatusDict[(stageMap[req.id] && stageMap[req.id][st.key]) || 'todo'].color }">
              {{ stageStatusDict[(stageMap[req.id] && stageMap[req.id][st.key]) || 'todo'].label }}
            </div>
          </div>
        </div>
      </div>

      <div v-if="!isLoading && !board.requirements.length" class="empty-state">
        <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/><polyline points="3.27 6.96 12 12.01 20.73 6.96"/><line x1="12" y1="22.08" x2="12" y2="12"/></svg>
        <span>暂无需求</span>
        <p>请先在「需求池管理」中录入需求</p>
      </div>
    </div>
  </GlassCard>
</template>

<style scoped>
.pipeline-board { width: 100%; }

.pb-header { display: flex; align-items: flex-start; justify-content: space-between; gap: 16px; margin-bottom: 16px; flex-wrap: wrap; }
.pb-title h2 { font-size: 20px; font-weight: 700; color: var(--text-primary); margin: 0; }
.pb-title p { font-size: 13px; color: var(--text-tertiary); margin: 4px 0 0; }

.btn-ghost { height: 36px; padding: 0 16px; border-radius: 8px; font-size: 13px; cursor: pointer; border: 1px solid var(--border-color); background: var(--bg-input); color: var(--text-secondary); }
.btn-ghost:hover { border-color: #165DFF; color: #165DFF; }

.stat-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(100px, 1fr)); gap: 12px; margin-bottom: 16px; }
.stat-card { background: #ffffff; border: 1px solid #E2E8F0; border-radius: 12px; padding: 14px 10px; text-align: center; }
.stat-count { font-size: 26px; font-weight: 700; line-height: 1.2; }
.stat-label { font-size: 12px; color: var(--text-tertiary); margin-top: 4px; }

.board-list { display: flex; flex-direction: column; gap: 14px; min-height: 120px; }
.board-card { background: #ffffff; border: 1px solid #E2E8F0; border-radius: 12px; padding: 16px; }
.bc-head { display: flex; align-items: center; justify-content: space-between; gap: 12px; flex-wrap: wrap; margin-bottom: 14px; }
.bc-left { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.req-no { font-family: 'Cascadia Code', Consolas, monospace; font-weight: 600; font-size: 13px; color: var(--text-secondary); }
.req-title { font-size: 14px; font-weight: 600; color: var(--text-primary); }
.status-pill { padding: 3px 12px; border-radius: 999px; color: #fff; font-size: 12px; font-weight: 600; }

.tag { display: inline-block; padding: 2px 8px; border-radius: 4px; font-size: 12px; line-height: 18px; }
.kano-M { background: #E8F3FF; color: #165DFF; }
.kano-O { background: #E8FFEA; color: #00B42A; }
.kano-A { background: #FFF3E8; color: #F77234; }
.kano-I { background: #F2F3F5; color: #86909C; }
.kano-R { background: #FFECE8; color: #F53F3F; }
.prio-P0 { background: #FFECE8; color: #F53F3F; font-weight: 600; }
.prio-P1 { background: #FFF3E8; color: #F77234; }
.prio-P2 { background: #F2F3F5; color: #4E5969; }
.aarrr { background: #F5E8FF; color: #722ED1; }

.stage-track { display: flex; gap: 4px; }
.stage-node { flex: 1; text-align: center; position: relative; }
.stage-line { position: relative; height: 34px; display: flex; align-items: center; justify-content: center; }
.stage-line.has-flow::before {
  content: ''; position: absolute; top: 50%; left: calc(50% + 14px); width: calc(100% - 28px);
  height: 2px; background: #E5E6EB; transform: translateY(-50%);
}
.stage-dot {
  width: 28px; height: 28px; border-radius: 50%; display: flex; align-items: center; justify-content: center;
  font-size: 13px; font-weight: 700; color: #fff; z-index: 1; background: #C9CDD4; transition: all .3s;
}
.stage-dot.doing { animation: pulse 1.6s ease-in-out infinite; }
@keyframes pulse {
  0%, 100% { box-shadow: 0 0 0 0 rgba(22, 93, 255, .4); }
  50% { box-shadow: 0 0 0 8px rgba(22, 93, 255, 0); }
}
.stage-name { font-size: 12px; font-weight: 600; color: var(--text-primary); margin-top: 6px; }
.stage-emp { font-size: 11px; color: var(--text-tertiary); }
.stage-status { font-size: 11px; margin-top: 2px; }

.empty-state { display: flex; flex-direction: column; align-items: center; gap: 8px; padding: 60px 0; color: var(--text-tertiary); }
.empty-state p { font-size: 12px; margin: 0; }

@media screen and (max-width: 760px) {
  .stage-track { overflow-x: auto; padding-bottom: 8px; }
  .stage-node { min-width: 72px; }
}
</style>
