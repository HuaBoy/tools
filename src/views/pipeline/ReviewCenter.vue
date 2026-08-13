<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import GlassCard from '@/components/GlassCard.vue'
import { useLogsStore } from '@/stores/logs'
import { getRequirements, getRequirement, createReview, getDeliverables } from '@/api/pipeline'

const logsStore = useLogsStore()

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

const stageDefs = [
  { key: 'A', name: '需求分析', emp: '员工A' },
  { key: 'B', name: '产品设计', emp: '员工B' },
  { key: 'C', name: '前端开发', emp: '员工C' },
  { key: 'D', name: '后端开发', emp: '员工D' },
  { key: 'E', name: '嵌入式', emp: '员工E' },
  { key: 'F', name: '测试', emp: '员工F' },
  { key: 'G', name: '验证', emp: '员工G' }
]

const filter = ref('review') // review / all
const records = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const isLoading = ref(false)

const showDialog = ref(false)
const current = ref(null) // 当前审核的需求详情
const detailLoading = ref(false)
const submitting = ref(false)
const reviewForm = reactive({
  action: 'approve',
  target_stage: '',
  reason: ''
})

const load = async () => {
  isLoading.value = true
  try {
    const res = await getRequirements({ status: filter.value, page: page.value, pageSize: pageSize.value })
    records.value = res.data
    total.value = res.total
  } catch (e) {
    ElMessage.error('加载失败: ' + e.message)
  } finally {
    isLoading.value = false
  }
}

const pendingCount = computed(() => (filter.value === 'review' ? total.value : 0))

const openReview = async (row) => {
  showDialog.value = true
  current.value = null
  detailLoading.value = true
  reviewForm.action = 'approve'
  reviewForm.target_stage = ''
  reviewForm.reason = ''
  try {
    const detail = await getRequirement(row.id)
    const dls = await getDeliverables(row.id)
    current.value = { ...detail.requirement, stages: detail.stages || [], reviews: detail.reviews || [], deliverables: dls }
  } catch (e) {
    ElMessage.error('加载详情失败: ' + e.message)
  } finally {
    detailLoading.value = false
  }
}

const submitReview = async () => {
  if (!current.value) return
  if (reviewForm.action === 'reject' && !reviewForm.target_stage) {
    ElMessage.warning('打回时必须选择目标环节')
    return
  }
  submitting.value = true
  try {
    await createReview(current.value.id, {
      action: reviewForm.action,
      target_stage: reviewForm.target_stage,
      reason: reviewForm.reason.trim()
    })
    const verb = reviewForm.action === 'approve' ? '通过' : `打回至环节 ${reviewForm.target_stage}`
    ElMessage.success(`审核${verb}完成`)
    logsStore.addLog('审核', 'AI研发流水线', `${current.value.reqNo} ${verb}`)
    showDialog.value = false
    load()
  } catch (e) {
    ElMessage.error('提交失败: ' + e.message)
  } finally {
    submitting.value = false
  }
}

const handlePageChange = (p) => { page.value = p; load() }
const switchFilter = (v) => { filter.value = v; page.value = 1; load() }

onMounted(load)
</script>

<template>
  <GlassCard class="review-center">
    <div class="rc-header">
      <div class="rc-title">
        <h2>审核中心</h2>
        <p>老板专属：对完成验证的需求统一审核，通过即完结，打回自动返回对应环节返工</p>
      </div>
      <div class="rc-tabs">
        <button class="tab-btn" :class="{ active: filter === 'review' }" @click="switchFilter('review')">待审核 ({{ pendingCount }})</button>
        <button class="tab-btn" :class="{ active: filter === 'all' }" @click="switchFilter('all')">全部需求</button>
      </div>
    </div>

    <div class="content-panel">
      <el-table :data="records" v-loading="isLoading" border stripe class="rc-table">
        <template #empty>
          <div class="empty-state">
            <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M9 11l3 3L22 4"/><path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11"/></svg>
            <span>{{ filter === 'review' ? '暂无待审核的需求' : '暂无需求' }}</span>
            <p>{{ filter === 'review' ? '需求走完验证环节后会自动进入这里，等待你的裁决' : '请先在「需求池管理」录入需求' }}</p>
          </div>
        </template>
        <el-table-column type="index" label="#" width="50" />
        <el-table-column prop="reqNo" label="编号" width="140">
          <template #default="{ row }"><span class="req-no">{{ row.reqNo }}</span></template>
        </el-table-column>
        <el-table-column prop="title" label="标题" min-width="200" show-overflow-tooltip />
        <el-table-column label="卡诺" width="100">
          <template #default="{ row }"><span class="tag" :class="'kano-' + (row.kanoCategory || 'O')">{{ kanoDict[row.kanoCategory] || row.kanoCategory }}</span></template>
        </el-table-column>
        <el-table-column label="优先级" width="90">
          <template #default="{ row }"><span class="tag prio" :class="'prio-' + (row.priority || 'P2')">{{ row.priority }}</span></template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <span class="status-dot" :style="{ background: (statusDict[row.status] || {}).color }"></span>
            <span :style="{ color: (statusDict[row.status] || {}).color }">{{ (statusDict[row.status] || {}).label || row.status }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="创建时间" width="160" />
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <button class="action-link" @click="openReview(row)">审核</button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pager" v-if="total > pageSize">
        <el-pagination
          :current-page="page"
          :page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          @current-change="handlePageChange"
        />
      </div>
    </div>

    <el-dialog v-model="showDialog" title="需求审核" width="680px">
      <div v-loading="detailLoading" class="review-body" v-if="current">
        <div class="rv-head">
          <span class="req-no">{{ current.reqNo }}</span>
          <span class="rv-title">{{ current.title }}</span>
          <span class="tag" :class="'kano-' + (current.kanoCategory || 'O')">{{ kanoDict[current.kanoCategory] || current.kanoCategory }}</span>
        </div>
        <p class="rv-desc">{{ current.description || '（无描述）' }}</p>

        <div class="rv-section">
          <div class="rv-section-title">交付物</div>
          <div v-if="current.deliverables && current.deliverables.length" class="dl-list">
            <div v-for="dl in current.deliverables" :key="dl.id" class="dl-item">
              <span class="dl-stage">{{ dl.stage }}</span>
              <span class="dl-title">{{ dl.title }}</span>
              <a v-if="dl.url" class="dl-link" :href="dl.url" target="_blank" rel="noopener">打开链接</a>
              <span v-if="dl.filePath" class="dl-path">{{ dl.filePath }}</span>
            </div>
          </div>
          <div v-else class="muted">该需求暂无登记交付物</div>
        </div>

        <div class="rv-section">
          <div class="rv-section-title">审核历史</div>
          <div v-if="current.reviews && current.reviews.length" class="rv-history">
            <div v-for="rv in current.reviews" :key="rv.id" class="rv-item">
              <span class="rv-action" :class="rv.action === 'approve' ? 'ok' : 'no'">{{ rv.action === 'approve' ? '通过' : '打回' }}</span>
              <span v-if="rv.targetStage">→ 环节{{ rv.targetStage }}</span>
              <span class="muted"> {{ rv.reason || '' }}</span>
              <span class="muted rv-meta">{{ rv.reviewedName }} · {{ rv.createdAt }}</span>
            </div>
          </div>
          <div v-else class="muted">暂无审核记录</div>
        </div>

        <div class="rv-section">
          <div class="rv-section-title">我的裁决</div>
          <div class="action-row">
            <label class="radio-item"><input type="radio" value="approve" v-model="reviewForm.action" /> 通过（完结需求）</label>
            <label class="radio-item"><input type="radio" value="reject" v-model="reviewForm.action" /> 打回返工</label>
          </div>
          <div v-if="reviewForm.action === 'reject'" class="reject-box">
            <div class="form-item">
              <label>打回目标环节 *</label>
              <select v-model="reviewForm.target_stage" class="form-select">
                <option value="">请选择目标环节</option>
                <option v-for="st in stageDefs" :key="st.key" :value="st.key">环节{{ st.key }} {{ st.name }}（{{ st.emp }}）</option>
              </select>
            </div>
            <div class="form-item">
              <label>打回原因</label>
              <textarea v-model="reviewForm.reason" rows="2" placeholder="说明打回原因，对应员工将据此返工"></textarea>
            </div>
          </div>
          <div v-else class="form-item">
            <label>审核意见（可选）</label>
            <textarea v-model="reviewForm.reason" rows="2" placeholder="通过时的补充意见"></textarea>
          </div>
        </div>
      </div>
      <template #footer>
        <button class="btn-ghost" @click="showDialog = false">取消</button>
        <button class="btn-primary" :class="{ danger: reviewForm.action === 'reject' }" :disabled="submitting || detailLoading" @click="submitReview">
          {{ submitting ? '提交中...' : (reviewForm.action === 'approve' ? '✅ 审核通过' : '⛔ 打回返工') }}
        </button>
      </template>
    </el-dialog>
  </GlassCard>
</template>

<style scoped>
.review-center { width: 100%; }

.rc-header { display: flex; align-items: flex-start; justify-content: space-between; gap: 16px; margin-bottom: 16px; flex-wrap: wrap; }
.rc-title h2 { font-size: 20px; font-weight: 700; color: var(--text-primary); margin: 0; }
.rc-title p { font-size: 13px; color: var(--text-tertiary); margin: 4px 0 0; }

.rc-tabs { display: flex; gap: 8px; }
.tab-btn { height: 34px; padding: 0 16px; border-radius: 8px; font-size: 13px; cursor: pointer; border: 1px solid var(--border-color); background: var(--bg-input); color: var(--text-secondary); }
.tab-btn.active { background: #165DFF; border-color: #165DFF; color: #fff; font-weight: 600; }

.content-panel { background: #ffffff; border: 1px solid #E2E8F0; border-radius: 12px; overflow: hidden; padding: 16px; }
.rc-table { width: 100%; }
.pager { display: flex; justify-content: flex-end; margin-top: 16px; }

.empty-state { display: flex; flex-direction: column; align-items: center; gap: 8px; padding: 50px 0; color: var(--text-tertiary); }
.empty-state p { font-size: 12px; margin: 0; }

.req-no { font-family: 'Cascadia Code', Consolas, monospace; font-weight: 600; font-size: 13px; color: var(--text-secondary); }
.muted { color: var(--text-tertiary); font-size: 12px; }
.action-link { background: none; border: none; color: #165DFF; cursor: pointer; font-size: 13px; padding: 0 4px; }
.action-link:hover { text-decoration: underline; }

.tag { display: inline-block; padding: 2px 8px; border-radius: 4px; font-size: 12px; line-height: 18px; }
.kano-M { background: #E8F3FF; color: #165DFF; }
.kano-O { background: #E8FFEA; color: #00B42A; }
.kano-A { background: #FFF3E8; color: #F77234; }
.kano-I { background: #F2F3F5; color: #86909C; }
.kano-R { background: #FFECE8; color: #F53F3F; }
.prio-P0 { background: #FFECE8; color: #F53F3F; font-weight: 600; }
.prio-P1 { background: #FFF3E8; color: #F77234; }
.prio-P2 { background: #F2F3F5; color: #4E5969; }

.status-dot { display: inline-block; width: 8px; height: 8px; border-radius: 50%; margin-right: 6px; }

.review-body { min-height: 200px; }
.rv-head { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.rv-title { font-size: 15px; font-weight: 700; color: var(--text-primary); }
.rv-desc { font-size: 13px; color: var(--text-secondary); margin: 8px 0 0; white-space: pre-line; }

.rv-section { margin-top: 16px; }
.rv-section-title { font-size: 13px; font-weight: 600; color: var(--text-primary); margin-bottom: 8px; padding-left: 8px; border-left: 3px solid #165DFF; }
.dl-list { display: flex; flex-direction: column; gap: 6px; }
.dl-item { display: flex; align-items: center; gap: 8px; background: var(--bg-input); border-radius: 8px; padding: 8px 12px; font-size: 13px; }
.dl-stage { width: 22px; height: 22px; border-radius: 50%; background: #165DFF; color: #fff; font-size: 12px; font-weight: 700; display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
.dl-title { color: var(--text-primary); }
.dl-link { color: #165DFF; text-decoration: none; }
.dl-path { color: var(--text-tertiary); font-size: 12px; }

.rv-history { display: flex; flex-direction: column; gap: 6px; }
.rv-item { display: flex; align-items: center; gap: 8px; background: var(--bg-input); border-radius: 8px; padding: 8px 12px; font-size: 13px; flex-wrap: wrap; }
.rv-action { padding: 1px 8px; border-radius: 4px; font-size: 12px; font-weight: 600; }
.rv-action.ok { background: #E8FFEA; color: #00B42A; }
.rv-action.no { background: #FFECE8; color: #F53F3F; }
.rv-meta { margin-left: auto; }

.action-row { display: flex; gap: 20px; margin-bottom: 10px; }
.radio-item { display: flex; align-items: center; gap: 6px; font-size: 13px; color: var(--text-primary); cursor: pointer; }
.reject-box { display: flex; flex-direction: column; gap: 12px; }

.form-item { display: flex; flex-direction: column; gap: 6px; margin-top: 10px; }
.form-item label { font-size: 12px; color: var(--text-secondary); }
.form-item select, .form-item textarea {
  border: 1px solid var(--border-color); border-radius: 8px; padding: 8px 10px;
  background: var(--bg-input); color: var(--text-primary); font-size: 13px; outline: none; resize: vertical;
}
.form-item select:focus, .form-item textarea:focus { border-color: #165DFF; }
.form-select { width: 100%; }

.btn-ghost { height: 36px; padding: 0 16px; border-radius: 8px; font-size: 13px; cursor: pointer; border: 1px solid var(--border-color); background: var(--bg-input); color: var(--text-secondary); }
.btn-ghost:hover { border-color: #165DFF; color: #165DFF; }
.btn-primary { height: 36px; padding: 0 18px; border-radius: 8px; font-size: 13px; font-weight: 600; cursor: pointer; border: none; background: #165DFF; color: #fff; }
.btn-primary:hover { background: #0F4CD0; }
.btn-primary.danger { background: #F53F3F; }
.btn-primary.danger:hover { background: #D03030; }
.btn-primary:disabled { opacity: .6; cursor: not-allowed; }
</style>
