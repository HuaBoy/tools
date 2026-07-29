<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import GlassCard from '@/components/GlassCard.vue'
import { useLogsStore } from '@/stores/logs'
import overseasShippingApi from '@/api/overseasShipping'

const logsStore = useLogsStore()

const records = ref([])
const total = ref(0)
const isLoading = ref(false)
const keyword = ref('')
const productTypeFilter = ref('')
const countryFilter = ref('')
const page = ref(1)
const pageSize = ref(10)

const showDialog = ref(false)
const editId = ref(null)
const submitting = ref(false)

const form = reactive({
  product_type: '',
  sales_engineer: '',
  country: '',
  customer_name: '',
  controller_sn: '',
  handheld_sn: '',
  controller_hw_version: '',
  controller_upgrade_history: '',
  handheld_upgrade_history: '',
  last_upgrade_date: '',
  assistant: '',
  remark: ''
})

const unique = (arr) => [...new Set(arr.filter(Boolean))]
const productTypeOptions = computed(() => unique(records.value.map(r => r.product_type)))
const countryOptions = computed(() => unique(records.value.map(r => r.country)))

const load = async () => {
  isLoading.value = true
  try {
    const res = await overseasShippingApi.list({
      keyword: keyword.value,
      productType: productTypeFilter.value,
      country: countryFilter.value,
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
const resetFilters = () => { keyword.value = ''; productTypeFilter.value = ''; countryFilter.value = ''; page.value = 1; load() }

const resetForm = () => { Object.keys(form).forEach(k => { form[k] = '' }) }
const handleAdd = () => { editId.value = null; resetForm(); showDialog.value = true }
const handleEdit = (row) => {
  editId.value = row.id
  Object.keys(form).forEach(k => { form[k] = row[k] || '' })
  showDialog.value = true
}

const handleSave = async () => {
  if (!form.customer_name.trim() && !form.controller_sn.trim()) {
    ElMessage.warning('请至少填写「客户名称」或「控制器编号」')
    return
  }
  submitting.value = true
  const payload = {}
  Object.keys(form).forEach(k => { payload[k] = (form[k] || '').trim() })
  const label = payload.customer_name || payload.controller_sn
  try {
    if (editId.value) {
      await overseasShippingApi.update(editId.value, payload)
      ElMessage.success('更新成功')
      logsStore.addLog('编辑', '海外发货管理', `更新: ${label}`)
    } else {
      await overseasShippingApi.create(payload)
      ElMessage.success('新增成功')
      logsStore.addLog('新增', '海外发货管理', `新增: ${label}`)
    }
    showDialog.value = false
    load()
  } catch (e) {
    ElMessage.error('保存失败: ' + e.message)
  } finally {
    submitting.value = false
  }
}

const handleDelete = async (row) => {
  const label = row.customer_name || row.controller_sn
  try {
    await ElMessageBox.confirm(`确定删除「${label}」这条发货记录吗？`, '删除确认', {
      confirmButtonText: '确定删除',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await overseasShippingApi.remove(row.id)
    ElMessage.success('删除成功')
    logsStore.addLog('删除', '海外发货管理', `删除: ${label}`)
    load()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('删除失败: ' + e.message)
  }
}

const handlePageChange = (p) => { page.value = p; load() }
const handleSizeChange = (s) => { pageSize.value = s; page.value = 1; load() }

const handleExport = async () => {
  const all = await overseasShippingApi.exportAll()
  if (!all.length) { ElMessage.warning('暂无数据可导出'); return }
  const headers = ['产品类型', '销售工程师', '国家', '客户名称', '控制器编号', '手持机编号', '控制器硬件版本', '控制器升级履历', '手持机升级履历', '最后升级日期', '协助人员', '备注']
  const keys = ['product_type', 'sales_engineer', 'country', 'customer_name', 'controller_sn', 'handheld_sn', 'controller_hw_version', 'controller_upgrade_history', 'handheld_upgrade_history', 'last_upgrade_date', 'assistant', 'remark']
  const escape = (v) => '"' + String(v || '').replace(/"/g, '""').replace(/\n/g, ' / ').replace(/\r/g, '') + '"'
  const rows = all.map(r => keys.map(k => escape(r[k])).join(','))
  const csv = '﻿' + headers.join(',') + '\n' + rows.join('\n')
  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = '海外发货管理_' + new Date().toISOString().slice(0, 10) + '.csv'
  a.click()
  URL.revokeObjectURL(url)
  ElMessage.success('已导出 ' + all.length + ' 条记录')
}

onMounted(load)
</script>

<template>
  <GlassCard class="overseas-shipping">
    <div class="os-header">
      <div class="os-title">
        <h2>海外发货管理</h2>
        <p>记录海外发货的控制器 / 手持机信息、升级履历与协助人员，支持查询与导出</p>
      </div>
      <button class="btn-primary" @click="handleAdd">+ 新增发货记录</button>
    </div>

    <div class="content-panel">
      <div class="toolbar">
        <div class="search-box">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
          <input v-model="keyword" @keyup.enter="handleSearch" placeholder="搜索客户 / 控制器编号 / 国家 / 销售工程师 / 手持机" />
        </div>
        <select v-model="productTypeFilter" class="filter-select" @change="handleSearch">
          <option value="">全部产品类型</option>
          <option v-for="t in productTypeOptions" :key="t" :value="t">{{ t }}</option>
        </select>
        <select v-model="countryFilter" class="filter-select" @change="handleSearch">
          <option value="">全部国家</option>
          <option v-for="c in countryOptions" :key="c" :value="c">{{ c }}</option>
        </select>
        <button class="btn-ghost" @click="handleSearch">查询</button>
        <button class="btn-ghost" @click="resetFilters">重置</button>
        <button class="btn-ghost export" @click="handleExport">导出 CSV</button>
      </div>

      <el-table :data="records" v-loading="isLoading" border stripe class="os-table" :default-sort="{ prop: 'last_upgrade_date', order: 'descending' }">
        <template #empty>
          <div class="empty-state">
            <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/><polyline points="3.27 6.96 12 12.01 20.73 6.96"/><line x1="12" y1="22.08" x2="12" y2="12"/></svg>
            <span>暂无发货记录</span>
            <p>点击「新增发货记录」添加第一条海外发货信息</p>
          </div>
        </template>
        <el-table-column type="index" label="#" width="50" />
        <el-table-column prop="product_type" label="产品类型" min-width="130" show-overflow-tooltip />
        <el-table-column prop="sales_engineer" label="销售工程师" width="100" />
        <el-table-column prop="country" label="国家" width="100" />
        <el-table-column prop="customer_name" label="客户名称" min-width="150" show-overflow-tooltip />
        <el-table-column prop="controller_sn" label="控制器编号" min-width="140" show-overflow-tooltip />
        <el-table-column prop="handheld_sn" label="手持机编号" min-width="130" show-overflow-tooltip />
        <el-table-column prop="controller_hw_version" label="控制器硬件版本" width="120" />
        <el-table-column prop="controller_upgrade_history" label="控制器升级履历" min-width="180" show-overflow-tooltip />
        <el-table-column prop="handheld_upgrade_history" label="手持机升级履历" min-width="180" show-overflow-tooltip />
        <el-table-column prop="last_upgrade_date" label="最后升级日期" width="120" sortable />
        <el-table-column prop="assistant" label="协助人员" width="100" />
        <el-table-column prop="remark" label="备注" min-width="140" show-overflow-tooltip />
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <button class="action-link" @click="handleEdit(row)">编辑</button>
            <button class="action-link danger" @click="handleDelete(row)">删除</button>
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

    <el-dialog v-model="showDialog" :title="editId ? '编辑发货记录' : '新增发货记录'" width="640px">
      <div class="form-grid">
        <div class="form-item"><label>产品类型</label><input v-model="form.product_type" placeholder="如：数码电子雷管控制器" /></div>
        <div class="form-item"><label>销售工程师</label><input v-model="form.sales_engineer" placeholder="如：张三" /></div>
        <div class="form-item"><label>国家</label><input v-model="form.country" placeholder="如：俄罗斯" /></div>
        <div class="form-item"><label>客户名称</label><input v-model="form.customer_name" placeholder="如：Ural Explosives" /></div>
        <div class="form-item"><label>控制器编号</label><input v-model="form.controller_sn" placeholder="如：CTRL-2024-001928" /></div>
        <div class="form-item"><label>手持机编号</label><input v-model="form.handheld_sn" placeholder="如：HH-2024-007341" /></div>
        <div class="form-item"><label>控制器硬件版本</label><input v-model="form.controller_hw_version" placeholder="如：HW-V2.3" /></div>
        <div class="form-item"><label>最后升级日期</label><input v-model="form.last_upgrade_date" type="date" /></div>
        <div class="form-item full"><label>控制器升级履历</label><textarea v-model="form.controller_upgrade_history" rows="2" placeholder="每行一条，如：2024-01-12 V1.0 → V2.0"></textarea></div>
        <div class="form-item full"><label>手持机升级履历</label><textarea v-model="form.handheld_upgrade_history" rows="2" placeholder="每行一条，如：2024-03-01 V1.5 → V2.0"></textarea></div>
        <div class="form-item"><label>协助人员</label><input v-model="form.assistant" placeholder="如：李娜" /></div>
        <div class="form-item full"><label>备注</label><textarea v-model="form.remark" rows="2" placeholder="可选备注"></textarea></div>
      </div>
      <template #footer>
        <button class="btn-ghost" @click="showDialog = false">取消</button>
        <button class="btn-primary" :disabled="submitting" @click="handleSave">{{ submitting ? '保存中...' : '保存' }}</button>
      </template>
    </el-dialog>
  </GlassCard>
</template>

<style scoped>
.overseas-shipping { width: 100%; }

.os-header {
  display: flex; align-items: flex-start; justify-content: space-between;
  gap: 16px; margin-bottom: 16px; flex-wrap: wrap;
}
.os-title h2 { font-size: 20px; font-weight: 700; color: var(--text-primary); margin: 0; }
.os-title p { font-size: 13px; color: var(--text-tertiary); margin: 4px 0 0; }

.content-panel {
  background: #ffffff; border: 1px solid #E2E8F0; border-radius: 12px;
  overflow: hidden; padding: 16px;
}

.toolbar { display: flex; align-items: center; gap: 10px; margin-bottom: 16px; flex-wrap: wrap; }
.search-box {
  display: flex; align-items: center; gap: 8px; flex: 1; min-width: 240px;
  background: var(--bg-input); border: 1px solid var(--border-color); border-radius: 8px; padding: 8px 12px;
}
.search-box input { border: none; background: transparent; outline: none; width: 100%; color: var(--text-primary); font-size: 13px; }
.filter-select, .btn-ghost, .btn-primary {
  height: 36px; border-radius: 8px; font-size: 13px; cursor: pointer; border: 1px solid var(--border-color);
}
.filter-select { padding: 0 10px; background: var(--bg-input); color: var(--text-primary); }
.btn-ghost { padding: 0 16px; background: var(--bg-input); color: var(--text-secondary); }
.btn-ghost:hover { border-color: #165DFF; color: #165DFF; }
.btn-primary {
  padding: 0 18px; background: #165DFF; color: #fff; border: none; font-weight: 600;
}
.btn-primary:hover { background: #0F4CD0; }
.btn-primary:disabled { opacity: .6; cursor: not-allowed; }
.btn-ghost.export { color: #165DFF; border-color: rgba(22,93,255,.4); }

.os-table { width: 100%; }

.pager { display: flex; justify-content: flex-end; margin-top: 16px; }

.empty-state { display: flex; flex-direction: column; align-items: center; gap: 8px; padding: 50px 0; color: var(--text-tertiary); }
.empty-state p { font-size: 12px; margin: 0; }

.action-link { background: none; border: none; color: #165DFF; cursor: pointer; font-size: 13px; padding: 0 4px; }
.action-link:hover { text-decoration: underline; }
.action-link.danger { color: #F53F3F; }

/* 对话框表单 */
.form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; }
.form-item { display: flex; flex-direction: column; gap: 6px; }
.form-item.full { grid-column: 1 / -1; }
.form-item label { font-size: 12px; color: var(--text-secondary); }
.form-item input, .form-item textarea {
  border: 1px solid var(--border-color); border-radius: 8px; padding: 8px 10px;
  background: var(--bg-input); color: var(--text-primary); font-size: 13px; outline: none; resize: vertical;
}
.form-item input:focus, .form-item textarea:focus { border-color: #165DFF; }

@media screen and (max-width: 560px) {
  .form-grid { grid-template-columns: 1fr; }
}
</style>
