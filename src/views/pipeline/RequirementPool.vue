<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import GlassCard from '@/components/GlassCard.vue'
import { useLogsStore } from '@/stores/logs'
import { getRequirements, createRequirement, updateRequirement, deleteRequirement, changeRequirementStatus } from '@/api/pipeline'

const logsStore = useLogsStore()

// 字典
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
  rejected: { label: '已淘汰', color: '#86909C' }
}
const kanoDict = { M: '基本型', O: '期望型', A: '兴奋型', I: '无差异', R: '反向型' }
const aarrrOptions = ['获取', '激活', '留存', '收入', '传播']
const priorityDict = { P0: 'P0 紧急', P1: 'P1 重要', P2: 'P2 常规' }

// 状态机（与后端一致）
const transitions = {
  draft: ['backlog'],
  backlog: ['prd', 'rejected'],
  prd: ['dev', 'rejected'],
  dev: ['test', 'rejected'],
  test: ['verify', 'rejected'],
  verify: ['review', 'rejected'],
  review: ['done'],
  rework: ['backlog', 'prd', 'dev', 'test', 'verify', 'review'],
  rejected: [],
  done: []
}

const records = ref([])
const total = ref(0)
const isLoading = ref(false)
const keyword = ref('')
const statusFilter = ref('')
const kanoFilter = ref('')
const priorityFilter = ref('')
const page = ref(1)
const pageSize = ref(10)

const showDialog = ref(false)
const editId = ref(null)
const submitting = ref(false)
const form = reactive({
  title: '',
  description: '',
  kano_category: 'O',
  aarrr_impacts: [],
  priority: 'P2',
  source: ''
})

const load = async () => {
  isLoading.value = true
  try {
    const res = await getRequirements({
      keyword: keyword.value,
      status: statusFilter.value,
      kanoCategory: kanoFilter.value,
      priority: priorityFilter.value,
      page: page.value,
      pageSize: pageSize.value
    })
    records.value = res.data
    total.value = res.total
  } catch (e) {
    ElMessage.error('加载失败: ' + e.message)
  } finally {
    isLoading.value = false
  }
}

const handleSearch = () => { page.value = 1; load() }
const resetFilters = () => {
  keyword.value = ''; statusFilter.value = ''; kanoFilter.value = ''; priorityFilter.value = ''
  page.value = 1; load()
}

const resetForm = () => {
  form.title = ''; form.description = ''; form.kano_category = 'O'
  form.aarrr_impacts = []; form.priority = 'P2'; form.source = ''
}
const handleAdd = () => { editId.value = null; resetForm(); showDialog.value = true }
const handleEdit = (row) => {
  editId.value = row.id
  form.title = row.title || ''
  form.description = row.description || ''
  form.kano_category = row.kanoCategory || 'O'
  form.aarrr_impacts = row.aarrrImpacts || []
  form.priority = row.priority || 'P2'
  form.source = row.source || ''
  showDialog.value = true
}

const handleSave = async () => {
  if (!form.title.trim()) { ElMessage.warning('请输入需求标题'); return }
  submitting.value = true
  const payload = {
    title: form.title.trim(),
    description: form.description.trim(),
    kano_category: form.kano_category,
    aarrr_impacts: form.aarrr_impacts,
    priority: form.priority,
    source: form.source.trim()
  }
  try {
    if (editId.value) {
      await updateRequirement(editId.value, payload)
      ElMessage.success('更新成功')
      logsStore.addLog('编辑', 'AI研发流水线', `编辑需求: ${payload.title}`)
    } else {
      const resp = await createRequirement(payload)
      ElMessage.success(`创建成功：${resp.reqNo || ''}`)
      logsStore.addLog('新增', 'AI研发流水线', `新建需求: ${payload.title}`)
    }
    showDialog.value = false
    load()
  } catch (e) {
    ElMessage.error('保存失败: ' + e.message)
  } finally {
    submitting.value = false
  }
}

const handleStatusChange = async (row, newStatus) => {
  const label = statusDict[newStatus]?.label || newStatus
  try {
    await ElMessageBox.confirm(`确定将「${row.title}」流转到「${label}」吗？`, '状态流转', {
      confirmButtonText: '确认流转',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await changeRequirementStatus(row.id, newStatus)
    ElMessage.success('状态已更新')
    logsStore.addLog('流转', 'AI研发流水线', `${row.reqNo} → ${label}`)
    load()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('流转失败: ' + e.message)
  }
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(`确定删除需求「${row.reqNo} ${row.title}」吗？删除后不可恢复。`, '删除确认', {
      confirmButtonText: '确定删除',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await deleteRequirement(row.id)
    ElMessage.success('删除成功')
    logsStore.addLog('删除', 'AI研发流水线', `删除: ${row.reqNo}`)
    load()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('删除失败: ' + e.message)
  }
}

const handlePageChange = (p) => { page.value = p; load() }
const handleSizeChange = (s) => { pageSize.value = s; page.value = 1; load() }

onMounted(load)
</script>

<template>
  <GlassCard class="requirement-pool">
    <div class="rp-header">
      <div class="rp-title">
        <h2>需求池管理</h2>
        <p>按卡诺模型 + AARRR 模型对需求进行过滤、评估与管理，驱动 AI 员工流水线</p>
      </div>
      <button class="btn-primary" @click="handleAdd">+ 新建需求</button>
    </div>

    <div class="content-panel">
      <div class="toolbar">
        <div class="search-box">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
          <input v-model="keyword" @keyup.enter="handleSearch" placeholder="搜索编号 / 标题 / 描述" />
        </div>
        <select v-model="statusFilter" class="filter-select" @change="handleSearch">
          <option value="">全部状态</option>
          <option v-for="(v, k) in statusDict" :key="k" :value="k">{{ v.label }}</option>
        </select>
        <select v-model="kanoFilter" class="filter-select" @change="handleSearch">
          <option value="">全部卡诺分类</option>
          <option v-for="(v, k) in kanoDict" :key="k" :value="k">{{ v }} ({{ k }})</option>
        </select>
        <select v-model="priorityFilter" class="filter-select" @change="handleSearch">
          <option value="">全部优先级</option>
          <option v-for="(v, k) in priorityDict" :key="k" :value="k">{{ v }}</option>
        </select>
        <button class="btn-ghost" @click="handleSearch">查询</button>
        <button class="btn-ghost" @click="resetFilters">重置</button>
      </div>

      <el-table :data="records" v-loading="isLoading" border stripe class="rp-table">
        <template #empty>
          <div class="empty-state">
            <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/><polyline points="3.27 6.96 12 12.01 20.73 6.96"/><line x1="12" y1="22.08" x2="12" y2="12"/></svg>
            <span>需求池暂无数据</span>
            <p>点击「新建需求」录入第一条需求，员工A 将按卡诺 + AARRR 评估</p>
          </div>
        </template>
        <el-table-column type="index" label="#" width="50" />
        <el-table-column prop="reqNo" label="编号" width="140">
          <template #default="{ row }"><span class="req-no">{{ row.reqNo }}</span></template>
        </el-table-column>
        <el-table-column prop="title" label="标题" min-width="180" show-overflow-tooltip />
        <el-table-column label="卡诺分类" width="110">
          <template #default="{ row }">
            <span class="tag" :class="'kano-' + (row.kanoCategory || 'O')">{{ kanoDict[row.kanoCategory] || row.kanoCategory }} ({{ row.kanoCategory }})</span>
          </template>
        </el-table-column>
        <el-table-column label="AARRR 影响" min-width="150">
          <template #default="{ row }">
            <span v-for="t in (row.aarrrImpacts || [])" :key="t" class="tag aarrr">{{ t }}</span>
            <span v-if="!(row.aarrrImpacts || []).length" class="muted">—</span>
          </template>
        </el-table-column>
        <el-table-column label="优先级" width="110">
          <template #default="{ row }">
            <span class="tag" :class="'prio-' + (row.priority || 'P2')">{{ priorityDict[row.priority] || row.priority }}</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <span class="status-dot" :style="{ background: (statusDict[row.status] || {}).color }"></span>
            <span :style="{ color: (statusDict[row.status] || {}).color }">{{ (statusDict[row.status] || {}).label || row.status }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="source" label="来源" width="110" show-overflow-tooltip />
        <el-table-column prop="createdAt" label="创建时间" width="160" />
        <el-table-column label="操作" width="230" fixed="right">
          <template #default="{ row }">
            <button class="action-link" @click="handleEdit(row)">编辑</button>
            <button class="action-link" @click="handleStatusChange(row, transitions[row.status] && transitions[row.status][0])" v-if="transitions[row.status] && transitions[row.status].length">流转→{{ statusDict[transitions[row.status][0]]?.label }}</button>
            <button class="action-link danger" @click="handleDelete(row)" v-if="row.status === 'draft' || row.status === 'backlog'">删除</button>
            <span class="muted" v-if="!transitions[row.status] || !transitions[row.status].length">已终结</span>
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
          @size-change="handleSizeChange"
        />
      </div>
    </div>

    <el-dialog v-model="showDialog" :title="editId ? '编辑需求' : '新建需求'" width="640px">
      <div class="form-grid">
        <div class="form-item full"><label>需求标题 *</label><input v-model="form.title" placeholder="一句话描述需求，如：AI 研发流水线可视化后台" /></div>
        <div class="form-item full"><label>需求描述</label><textarea v-model="form.description" rows="3" placeholder="背景、目标、用户故事等补充信息"></textarea></div>
        <div class="form-item"><label>卡诺分类</label>
          <select v-model="form.kano_category" class="form-select">
            <option v-for="(v, k) in kanoDict" :key="k" :value="k">{{ v }} ({{ k }})</option>
          </select>
        </div>
        <div class="form-item"><label>优先级</label>
          <select v-model="form.priority" class="form-select">
            <option v-for="(v, k) in priorityDict" :key="k" :value="k">{{ v }}</option>
          </select>
        </div>
        <div class="form-item"><label>AARRR 影响环节</label>
          <div class="check-group">
            <label v-for="t in aarrrOptions" :key="t" class="check-item">
              <input type="checkbox" :value="t" v-model="form.aarrr_impacts" /> {{ t }}
            </label>
          </div>
        </div>
        <div class="form-item"><label>需求来源</label><input v-model="form.source" placeholder="如：老板 / 客户 / 内部提效 / 市场" /></div>
      </div>
      <template #footer>
        <button class="btn-ghost" @click="showDialog = false">取消</button>
        <button class="btn-primary" :disabled="submitting" @click="handleSave">{{ submitting ? '保存中...' : '保存' }}</button>
      </template>
    </el-dialog>
  </GlassCard>
</template>

<style scoped>
.requirement-pool { width: 100%; }

.rp-header { display: flex; align-items: flex-start; justify-content: space-between; gap: 16px; margin-bottom: 16px; flex-wrap: wrap; }
.rp-title h2 { font-size: 20px; font-weight: 700; color: var(--text-primary); margin: 0; }
.rp-title p { font-size: 13px; color: var(--text-tertiary); margin: 4px 0 0; }

.content-panel { background: #ffffff; border: 1px solid #E2E8F0; border-radius: 12px; overflow: hidden; padding: 16px; }

.toolbar { display: flex; align-items: center; gap: 10px; margin-bottom: 16px; flex-wrap: wrap; }
.search-box { display: flex; align-items: center; gap: 8px; flex: 1; min-width: 240px; background: var(--bg-input); border: 1px solid var(--border-color); border-radius: 8px; padding: 8px 12px; }
.search-box input { border: none; background: transparent; outline: none; width: 100%; color: var(--text-primary); font-size: 13px; }
.filter-select, .btn-ghost, .btn-primary { height: 36px; border-radius: 8px; font-size: 13px; cursor: pointer; border: 1px solid var(--border-color); }
.filter-select { padding: 0 10px; background: var(--bg-input); color: var(--text-primary); }
.btn-ghost { padding: 0 16px; background: var(--bg-input); color: var(--text-secondary); }
.btn-ghost:hover { border-color: #165DFF; color: #165DFF; }
.btn-primary { padding: 0 18px; background: #165DFF; color: #fff; border: none; font-weight: 600; }
.btn-primary:hover { background: #0F4CD0; }
.btn-primary:disabled { opacity: .6; cursor: not-allowed; }

.rp-table { width: 100%; }
.pager { display: flex; justify-content: flex-end; margin-top: 16px; }

.empty-state { display: flex; flex-direction: column; align-items: center; gap: 8px; padding: 50px 0; color: var(--text-tertiary); }
.empty-state p { font-size: 12px; margin: 0; }

.action-link { background: none; border: none; color: #165DFF; cursor: pointer; font-size: 13px; padding: 0 4px; }
.action-link:hover { text-decoration: underline; }
.action-link.danger { color: #F53F3F; }

.req-no { font-family: 'Cascadia Code', Consolas, monospace; font-weight: 600; color: var(--text-secondary); }
.muted { color: var(--text-tertiary); font-size: 12px; }

.tag { display: inline-block; padding: 2px 8px; margin: 1px 4px 1px 0; border-radius: 4px; font-size: 12px; line-height: 18px; }
.kano-M { background: #E8F3FF; color: #165DFF; }
.kano-O { background: #E8FFEA; color: #00B42A; }
.kano-A { background: #FFF3E8; color: #F77234; }
.kano-I { background: #F2F3F5; color: #86909C; }
.kano-R { background: #FFECE8; color: #F53F3F; }
.aarrr { background: #F5E8FF; color: #722ED1; }
.prio-P0 { background: #FFECE8; color: #F53F3F; font-weight: 600; }
.prio-P1 { background: #FFF3E8; color: #F77234; }
.prio-P2 { background: #F2F3F5; color: #4E5969; }

.status-dot { display: inline-block; width: 8px; height: 8px; border-radius: 50%; margin-right: 6px; }

.form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; }
.form-item { display: flex; flex-direction: column; gap: 6px; }
.form-item.full { grid-column: 1 / -1; }
.form-item label { font-size: 12px; color: var(--text-secondary); }
.form-item input, .form-item textarea, .form-select {
  border: 1px solid var(--border-color); border-radius: 8px; padding: 8px 10px;
  background: var(--bg-input); color: var(--text-primary); font-size: 13px; outline: none; resize: vertical;
}
.form-item input:focus, .form-item textarea:focus, .form-select:focus { border-color: #165DFF; }
.form-select { width: 100%; }
.check-group { display: flex; flex-wrap: wrap; gap: 10px; padding-top: 6px; }
.check-item { display: flex; align-items: center; gap: 4px; font-size: 13px; color: var(--text-secondary); cursor: pointer; }

@media screen and (max-width: 560px) {
  .form-grid { grid-template-columns: 1fr; }
}
</style>
