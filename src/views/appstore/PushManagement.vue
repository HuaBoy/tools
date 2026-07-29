<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { appStoreApi, pushApi } from '@/api/appstore'

const apps = ref([])
const tasks = ref([])
const loading = ref(false)

const form = ref({
  app_id: '',
  title: '',
  content: '',
  target: 'all',
  sn: ''
})

const targetOptions = [
  { value: 'all', label: '全部设备' },
  { value: 'specified', label: '指定设备' }
]

async function loadApps() {
  try {
    const res = await appStoreApi.list()
    if (res && res.code === 0) apps.value = res.data || []
  } catch { /* ignore */ }
}

async function loadTasks() {
  loading.value = true
  try {
    const res = await pushApi.list()
    if (res && res.code === 0) tasks.value = res.data || []
  } catch (e) {
    ElMessage.error('加载推送记录失败：' + (e.message || '网络错误'))
  } finally {
    loading.value = false
  }
}

async function doPush() {
  if (!form.value.app_id) { ElMessage.warning('请选择应用'); return }
  if (!form.value.title) { ElMessage.warning('请输入推送标题'); return }
  if (form.value.target === 'specified' && !form.value.sn.trim()) {
    ElMessage.warning('指定设备时请填写 SN'); return
  }
  loading.value = true
  try {
    const createRes = await pushApi.create({
      app_id: Number(form.value.app_id),
      title: form.value.title,
      content: form.value.content,
      target: form.value.target,
      target_sn: form.value.target === 'specified' ? form.value.sn.trim() : ''
    })
    if (!(createRes && createRes.code === 0)) {
      ElMessage.error((createRes && createRes.message) || '创建推送失败')
      loading.value = false
      return
    }
    const id = createRes.data && createRes.data.id
    const sendRes = await pushApi.send(id)
    if (sendRes && sendRes.code === 0) {
      ElMessage.success('推送已下发')
      form.value = { app_id: '', title: '', content: '', target: 'all', sn: '' }
      loadTasks()
    } else {
      ElMessage.error((sendRes && sendRes.message) || '推送下发失败')
    }
  } catch (e) {
    ElMessage.error('推送失败：' + (e.message || '网络错误'))
  } finally {
    loading.value = false
  }
}

// 设备视角：按 SN 预览该设备可见的版本（验证 SN 过滤）
const previewSn = ref('')
const previewList = ref([])
const previewLoading = ref(false)
async function previewDevice() {
  if (!previewSn.value.trim()) { ElMessage.warning('请输入设备 SN'); return }
  previewLoading.value = true
  try {
    const res = await pushApi.deviceUpdates(previewSn.value.trim())
    if (res && res.code === 0) {
      previewList.value = res.data || []
    } else {
      previewList.value = []
    }
  } catch (e) {
    ElMessage.error('查询失败：' + (e.message || '网络错误'))
  } finally {
    previewLoading.value = false
  }
}

async function removeTask(row) {
  try {
    await ElMessageBox.confirm(`确认删除该推送记录？`, '提示', { type: 'warning' })
  } catch { return }
  try {
    const res = await pushApi.remove(row.id)
    if (res && res.code === 0) {
      ElMessage.success('删除成功')
      loadTasks()
    } else {
      ElMessage.error((res && res.message) || '删除失败')
    }
  } catch (e) {
    ElMessage.error('删除失败：' + (e.message || '网络错误'))
  }
}

const appNameMap = (id) => {
  const a = apps.value.find(x => x.id === id)
  return a ? a.name : '—'
}

onMounted(() => { loadApps(); loadTasks() })
</script>

<template>
  <div class="push-mgmt">
    <div class="pm-header">
      <h2 class="pm-title">推送管理</h2>
      <p class="pm-sub">创建推送任务并下发至设备（服务端持久化）</p>
    </div>

    <div class="pm-layout">
      <div class="pm-form card">
        <h3 class="card-title">新建推送</h3>
        <div class="field">
          <label>选择应用</label>
          <select v-model="form.app_id" class="input">
            <option value="">请选择应用</option>
            <option v-for="a in apps" :key="a.id" :value="a.id">{{ a.name }}</option>
          </select>
        </div>
        <div class="field">
          <label>推送标题</label>
          <input v-model="form.title" class="input" placeholder="如：V1.2.0 版本已发布" />
        </div>
        <div class="field">
          <label>推送内容</label>
          <textarea v-model="form.content" class="input" rows="4" placeholder="推送正文内容"></textarea>
        </div>
        <div class="field">
          <label>推送目标</label>
          <select v-model="form.target" class="input">
            <option v-for="t in targetOptions" :key="t.value" :value="t.value">{{ t.label }}</option>
          </select>
        </div>
        <div v-if="form.target === 'specified'" class="field">
          <label>目标 SN（逗号分隔）</label>
          <input v-model="form.sn" class="input" placeholder="如 10001,10002" />
        </div>
        <button class="btn-primary" :disabled="loading" @click="doPush">推送</button>
      </div>

      <div class="pm-list card">
        <h3 class="card-title">推送记录</h3>
        <div v-loading="loading" class="task-table">
          <div class="task-head">
            <span class="col col-app">应用</span>
            <span class="col col-title">标题</span>
            <span class="col col-target">目标</span>
            <span class="col col-sn">SN</span>
            <span class="col col-status">状态</span>
            <span class="col col-op">操作</span>
          </div>
          <div v-for="t in tasks" :key="t.id" class="task-row">
            <span class="col col-app">{{ appNameMap(t.app_id) }}</span>
            <span class="col col-title">{{ t.title }}</span>
            <span class="col col-target">{{ t.target === 'all' ? '全部' : '指定' }}</span>
            <span class="col col-sn">{{ t.target === 'all' ? '—' : (t.target_sn || '—') }}</span>
            <span class="col col-status">
              <span class="status-tag" :class="t.status === 1 ? 'on' : 'off'">
                {{ t.status === 1 ? '已下发' : '待下发' }}
              </span>
            </span>
            <span class="col col-op">
              <button class="op-btn danger" @click="removeTask(t)">删除</button>
            </span>
          </div>
          <div v-if="!loading && tasks.length === 0" class="empty-tip">暂无推送记录</div>
        </div>
      </div>

      <div class="pm-preview card">
        <h3 class="card-title">设备视角预览（验证 SN 过滤）</h3>
        <div class="field" style="display:flex;gap:8px;align-items:flex-end;">
          <div style="flex:1">
            <label>输入设备 SN</label>
            <input v-model="previewSn" class="input" placeholder="如 10001" />
          </div>
          <button class="btn-primary" style="width:auto;padding:9px 18px" :disabled="previewLoading" @click="previewDevice">查询可见版本</button>
        </div>
        <div v-if="previewList.length" class="preview-list">
          <div v-for="p in previewList" :key="p.push_id" class="preview-item">
            <img v-if="p.icon_url" :src="p.icon_url" class="preview-icon" />
            <div class="preview-info">
              <div class="preview-name">{{ p.app_name }} <span class="preview-ver">v{{ p.version }}</span></div>
              <div class="preview-title">{{ p.title }}</div>
              <a :href="p.download_url" class="preview-dl" target="_blank">下载安装包</a>
            </div>
          </div>
        </div>
        <div v-else-if="!previewLoading" class="empty-tip">该 SN 暂无可更新的版本</div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.push-mgmt { padding: 20px; color: var(--text-primary); }
.pm-header { margin-bottom: 18px; }
.pm-title { font-size: 20px; font-weight: 700; margin: 0; }
.pm-sub { font-size: 12px; color: var(--text-tertiary); margin: 4px 0 0; }

.pm-layout { display: grid; grid-template-columns: 360px 1fr; gap: 16px; align-items: start; }
.card { background: rgba(255,255,255,0.04); border: 1px solid var(--border-color); border-radius: 16px; padding: 18px; }
.card-title { font-size: 14px; font-weight: 600; margin: 0 0 14px; }

.field { margin-bottom: 14px; }
.field label { display: block; font-size: 12px; color: var(--text-tertiary); margin-bottom: 6px; }
.input {
  width: 100%; padding: 9px 12px; border-radius: 10px; font-size: 13px;
  background: rgba(255,255,255,0.04); border: 1px solid var(--border-color);
  color: var(--text-primary); outline: none; box-sizing: border-box; resize: vertical;
}
.input:focus { border-color: rgba(22,93,255,0.5); }
.btn-primary {
  width: 100%; background: linear-gradient(135deg, #165DFF, #0F4CD0); color: #fff; border: none;
  padding: 10px; border-radius: 10px; cursor: pointer; font-size: 14px; font-weight: 600;
}
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }

.task-table { font-size: 13px; }
.task-head, .task-row { display: grid; grid-template-columns: 1.2fr 1.8fr 0.7fr 1fr 0.9fr 0.7fr; gap: 8px; padding: 10px 6px; align-items: center; }
.task-head { color: var(--text-tertiary); font-size: 12px; border-bottom: 1px solid var(--border-color); }
.col-sn { font-family: monospace; font-size: 12px; color: var(--text-secondary); word-break: break-all; }

.pm-preview { margin-top: 16px; }
.preview-list { display: flex; flex-direction: column; gap: 12px; }
.preview-item { display: flex; gap: 12px; align-items: center; background: rgba(255,255,255,0.03); border: 1px solid var(--border-color); border-radius: 12px; padding: 12px; }
.preview-icon { width: 44px; height: 44px; border-radius: 10px; object-fit: cover; background: rgba(255,255,255,0.06); }
.preview-info { flex: 1; min-width: 0; }
.preview-name { font-size: 14px; font-weight: 600; }
.preview-ver { font-size: 12px; color: var(--text-tertiary); font-weight: 400; margin-left: 4px; }
.preview-title { font-size: 12px; color: var(--text-secondary); margin: 4px 0 6px; }
.preview-dl { font-size: 12px; color: #165DFF; text-decoration: none; }
.preview-dl:hover { text-decoration: underline; }
.task-row { border-bottom: 1px solid rgba(255,255,255,0.04); }
.col-title { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.status-tag { font-size: 11px; padding: 2px 10px; border-radius: 8px; }
.status-tag.on { background: rgba(0,180,42,0.15); color: #00B42A; }
.status-tag.off { background: rgba(245,63,63,0.15); color: #F53F3F; }
.op-btn { font-size: 12px; padding: 4px 12px; border-radius: 8px; cursor: pointer; background: rgba(255,255,255,0.06); border: 1px solid var(--border-color); color: var(--text-secondary); }
.op-btn.danger:hover { border-color: rgba(245,63,63,0.5); color: #F53F3F; }
.empty-tip { text-align: center; color: var(--text-tertiary); padding: 30px; }

@media screen and (max-width: 900px) {
  .pm-layout { grid-template-columns: 1fr; }
}
</style>
