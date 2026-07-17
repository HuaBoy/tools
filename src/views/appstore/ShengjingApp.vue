<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Edit, Delete, Download, Document, Upload, Search } from '@element-plus/icons-vue'
import { localAppStore, DEFAULT_DEVICES } from '@/utils/localAppStore'
import { putFile, getFile, deleteFile, downloadBlob, formatSize } from '@/utils/indexedDB'

const apps = ref([])
const searchKey = ref('')
const deviceFilter = ref('')

const dialogVisible = ref(false)
const dialogMode = ref('create') // create | edit
const submitting = ref(false)

const form = reactive({
  id: null,
  name: '',
  intro: '',
  manualText: '',
  devices: []
})
const apkFile = ref(null)     // { file?, name, size, type, id? }
const manualFile = ref(null)  // { file?, name, size, type, id? }

const apkInput = ref(null)
const manualInput = ref(null)

const deviceOptions = computed(() => localAppStore.getDevices())

const filteredApps = computed(() => {
  return apps.value.filter((a) => {
    const matchKey = !searchKey.value || a.name.toLowerCase().includes(searchKey.value.toLowerCase())
    const matchDev = !deviceFilter.value || (a.devices || []).includes(deviceFilter.value)
    return matchKey && matchDev
  })
})

function refresh() {
  apps.value = localAppStore.getAll()
}

function openCreate() {
  dialogMode.value = 'create'
  Object.assign(form, { id: null, name: '', intro: '', manualText: '', devices: [] })
  apkFile.value = null
  manualFile.value = null
  dialogVisible.value = true
}

function openEdit(app) {
  dialogMode.value = 'edit'
  Object.assign(form, {
    id: app.id,
    name: app.name,
    intro: app.intro || '',
    manualText: app.manualText || '',
    devices: [...(app.devices || [])]
  })
  apkFile.value = app.apkFile ? { ...app.apkFile } : null
  manualFile.value = app.manualFile ? { ...app.manualFile } : null
  dialogVisible.value = true
}

// ===== 文件选择 =====
function pickApk() { apkInput.value?.click() }
function pickManual() { manualInput.value?.click() }

function onApkChange(e) {
  const f = e.target.files?.[0]
  if (!f) return
  if (!/\.apk$/i.test(f.name)) {
    ElMessage.warning('请选择 .apk 格式的安装包')
    e.target.value = ''
    return
  }
  apkFile.value = { file: f, name: f.name, size: f.size, type: f.type || 'application/vnd.android.package-archive' }
  e.target.value = ''
}

function onManualChange(e) {
  const f = e.target.files?.[0]
  if (!f) return
  manualFile.value = { file: f, name: f.name, size: f.size, type: f.type || 'application/octet-stream' }
  e.target.value = ''
}

function removeApk() { apkFile.value = null }
function removeManual() { manualFile.value = null }

// 设备新增时持久化
function onDevicesChange(val) {
  val.forEach((d) => {
    if (!DEFAULT_DEVICES.includes(d)) localAppStore.addDevice(d)
  })
}

// ===== 提交 =====
async function submit() {
  if (!form.name.trim()) {
    ElMessage.warning('请填写应用名称')
    return
  }
  submitting.value = true
  try {
    const id = form.id || ('app-' + Date.now().toString(36) + Math.random().toString(36).substr(2, 5))
    const isEdit = dialogMode.value === 'edit'

    // APK
    let apkMeta = null
    const oldApkId = isEdit ? localAppStore.getById(form.id)?.apkFile?.id : null
    if (apkFile.value) {
      if (apkFile.value.file) {
        const key = id + ':apk'
        await putFile(key, {
          name: apkFile.value.name,
          type: apkFile.value.type,
          size: apkFile.value.size,
          blob: apkFile.value.file
        })
        apkMeta = { id: key, name: apkFile.value.name, size: apkFile.value.size, type: apkFile.value.type }
      } else {
        apkMeta = { id: apkFile.value.id, name: apkFile.value.name, size: apkFile.value.size, type: apkFile.value.type }
      }
    } else if (oldApkId) {
      await deleteFile(oldApkId).catch(() => {})
    }

    // 操作手册文件
    let manualMeta = null
    const oldManualId = isEdit ? localAppStore.getById(form.id)?.manualFile?.id : null
    if (manualFile.value) {
      if (manualFile.value.file) {
        const key = id + ':manual'
        await putFile(key, {
          name: manualFile.value.name,
          type: manualFile.value.type,
          size: manualFile.value.size,
          blob: manualFile.value.file
        })
        manualMeta = { id: key, name: manualFile.value.name, size: manualFile.value.size, type: manualFile.value.type }
      } else {
        manualMeta = { id: manualFile.value.id, name: manualFile.value.name, size: manualFile.value.size, type: manualFile.value.type }
      }
    } else if (oldManualId) {
      await deleteFile(oldManualId).catch(() => {})
    }

    const payload = {
      name: form.name.trim(),
      intro: form.intro.trim(),
      manualText: form.manualText.trim(),
      devices: [...form.devices],
      apkFile: apkMeta,
      manualFile: manualMeta
    }

    if (isEdit) {
      localAppStore.update(form.id, payload)
      ElMessage.success('应用已更新')
    } else {
      localAppStore.create(payload)
      ElMessage.success('应用已创建')
    }
    dialogVisible.value = false
    refresh()
  } catch (e) {
    ElMessage.error('保存失败：' + (e.message || e))
  } finally {
    submitting.value = false
  }
}

// ===== 删除 =====
async function handleDelete(app) {
  try {
    await ElMessageBox.confirm(`确认删除应用「${app.name}」？关联的 APK 与手册文件将一并删除。`, '删除确认', {
      type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消'
    })
  } catch {
    return
  }
  try {
    await localAppStore.remove(app.id)
    ElMessage.success('已删除')
    refresh()
  } catch (e) {
    ElMessage.error('删除失败：' + (e.message || e))
  }
}

// ===== 下载 =====
async function downloadApk(app) {
  if (!app.apkFile?.id) {
    ElMessage.warning('该应用未上传 APK 包')
    return
  }
  try {
    const rec = await getFile(app.apkFile.id)
    if (!rec) { ElMessage.error('APK 文件不存在或已损坏'); return }
    downloadBlob(rec.blob, app.apkFile.name)
  } catch (e) {
    ElMessage.error('下载失败：' + (e.message || e))
  }
}

const manualTextVisible = ref(false)
const manualTextContent = ref('')
const manualTextTitle = ref('')

function viewManual(app) {
  if (app.manualText) {
    manualTextContent.value = app.manualText
    manualTextTitle.value = app.name + ' - 操作手册'
    manualTextVisible.value = true
  } else if (app.manualFile?.id) {
    downloadManual(app)
  } else {
    ElMessage.info('该应用暂无操作手册')
  }
}

async function downloadManual(app) {
  if (!app.manualFile?.id) return
  try {
    const rec = await getFile(app.manualFile.id)
    if (!rec) { ElMessage.error('手册文件不存在或已损坏'); return }
    downloadBlob(rec.blob, app.manualFile.name)
  } catch (e) {
    ElMessage.error('下载失败：' + (e.message || e))
  }
}

function formatTime(t) {
  try { return new Date(t).toLocaleString('zh-CN') } catch { return t }
}

onMounted(refresh)
</script>

<template>
  <div class="sj-app">
    <!-- 顶部 -->
    <div class="sj-header">
      <div class="sj-title">
        <span class="sj-logo">📦</span>
        <div>
          <h1>盛景应用</h1>
          <p>管理应用安装包、介绍、操作手册与适配设备</p>
        </div>
      </div>
      <el-button type="primary" :icon="Plus" @click="openCreate">新增应用</el-button>
    </div>

    <!-- 筛选 -->
    <div class="sj-filter">
      <el-input v-model="searchKey" placeholder="搜索应用名称" clearable :prefix-icon="Search" style="width: 260px" />
      <el-select v-model="deviceFilter" placeholder="按适配设备筛选" clearable style="width: 200px">
        <el-option v-for="d in deviceOptions" :key="d" :label="d" :value="d" />
      </el-select>
      <span class="sj-count">共 {{ filteredApps.length }} 个应用</span>
    </div>

    <!-- 列表 -->
    <div v-if="filteredApps.length" class="sj-grid">
      <div v-for="app in filteredApps" :key="app.id" class="sj-card">
        <div class="sj-card-head">
          <h3 :title="app.name">{{ app.name }}</h3>
          <div class="sj-card-ops">
            <el-button text :icon="Edit" @click="openEdit(app)">编辑</el-button>
            <el-button text type="danger" :icon="Delete" @click="handleDelete(app)">删除</el-button>
          </div>
        </div>

        <p class="sj-intro">{{ app.intro || '暂无应用介绍' }}</p>

        <div class="sj-devices">
          <el-tag v-for="d in app.devices" :key="d" size="small" type="info" effect="plain" class="sj-tag">{{ d }}</el-tag>
          <span v-if="!app.devices || !app.devices.length" class="sj-no-device">未设置适配设备</span>
        </div>

        <div class="sj-meta">
          <span>APK：{{ app.apkFile ? app.apkFile.name + ' (' + formatSize(app.apkFile.size) + ')' : '未上传' }}</span>
          <span>手册：{{ app.manualFile ? app.manualFile.name : (app.manualText ? '文本手册' : '无') }}</span>
        </div>

        <div class="sj-actions">
          <el-button type="primary" plain :icon="Download" :disabled="!app.apkFile" @click="downloadApk(app)">下载 APK</el-button>
          <el-button plain :icon="Document" @click="viewManual(app)">操作手册</el-button>
        </div>

        <div class="sj-time">更新于 {{ formatTime(app.updated_at) }}</div>
      </div>
    </div>

    <el-empty v-else description="暂无应用，点击右上角「新增应用」开始添加" />

    <!-- 新增/编辑弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogMode === 'create' ? '新增盛景应用' : '编辑应用'"
      width="640px"
      :close-on-click-modal="false"
    >
      <el-form label-width="100px" label-position="right">
        <el-form-item label="应用名称" required>
          <el-input v-model="form.name" placeholder="请输入应用名称" maxlength="50" show-word-limit />
        </el-form-item>

        <el-form-item label="应用介绍">
          <el-input v-model="form.intro" type="textarea" :rows="3" placeholder="简要描述应用功能、用途等" maxlength="500" show-word-limit />
        </el-form-item>

        <el-form-item label="操作手册">
          <div class="sj-manual-block">
            <el-input v-model="form.manualText" type="textarea" :rows="3" placeholder="可填写操作手册说明文本（可选）" />
            <div class="sj-upload-row">
              <template v-if="manualFile">
                <span class="sj-file">
                  📄 {{ manualFile.name }}
                  <em>{{ formatSize(manualFile.size) }}</em>
                  <el-button text type="danger" size="small" @click="removeManual">移除</el-button>
                </span>
              </template>
              <el-button v-else plain :icon="Upload" @click="pickManual">上传手册文件（PDF/Word 等）</el-button>
              <input ref="manualInput" type="file" accept=".pdf,.doc,.docx,.ppt,.pptx,.xls,.xlsx,.txt,.md" hidden @change="onManualChange" />
            </div>
          </div>
        </el-form-item>

        <el-form-item label="APK 包">
          <div class="sj-upload-row">
            <template v-if="apkFile">
              <span class="sj-file">
                📦 {{ apkFile.name }}
                <em>{{ formatSize(apkFile.size) }}</em>
                <el-button text type="danger" size="small" @click="removeApk">移除</el-button>
              </span>
            </template>
            <el-button v-else plain :icon="Upload" @click="pickApk">上传 APK 安装包</el-button>
            <input ref="apkInput" type="file" accept=".apk" hidden @change="onApkChange" />
          </div>
        </el-form-item>

        <el-form-item label="适配设备">
          <el-select
            v-model="form.devices"
            multiple
            filterable
            allow-create
            default-first-option
            placeholder="选择或输入新建设备"
            style="width: 100%"
            @change="onDevicesChange"
          >
            <el-option v-for="d in deviceOptions" :key="d" :label="d" :value="d" />
          </el-select>
          <div class="sj-dev-hint">预置：{{ DEFAULT_DEVICES.join('、') }}；可直接输入新建设备</div>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submit">保存</el-button>
      </template>
    </el-dialog>

    <!-- 手册文本查看 -->
    <el-dialog v-model="manualTextVisible" :title="manualTextTitle" width="600px">
      <pre class="sj-manual-text">{{ manualTextContent }}</pre>
    </el-dialog>
  </div>
</template>

<style scoped>
.sj-app {
  width: 100%;
  padding: 20px 24px 40px;
  box-sizing: border-box;
}
.sj-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 18px;
}
.sj-title { display: flex; align-items: center; gap: 14px; }
.sj-logo {
  width: 46px; height: 46px; border-radius: 12px;
  background: linear-gradient(135deg, #165DFF, #00D2AC);
  display: flex; align-items: center; justify-content: center;
  font-size: 24px;
}
.sj-title h1 { font-size: 20px; font-weight: 700; color: #1E293B; margin: 0; }
.sj-title p { font-size: 12px; color: #94A3B8; margin: 2px 0 0; }

.sj-filter {
  display: flex; align-items: center; gap: 12px;
  margin-bottom: 16px; flex-wrap: wrap;
}
.sj-count { font-size: 12px; color: #94A3B8; }

.sj-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}
@media screen and (max-width: 1280px) {
  .sj-grid { grid-template-columns: repeat(3, 1fr); }
}
@media screen and (max-width: 960px) {
  .sj-grid { grid-template-columns: repeat(2, 1fr); }
}
@media screen and (max-width: 600px) {
  .sj-grid { grid-template-columns: 1fr; }
}
.sj-card {
  background: #fff; border: 1px solid #E8ECF1; border-radius: 14px;
  padding: 18px; display: flex; flex-direction: column;
  transition: box-shadow 0.2s, transform 0.2s;
}
.sj-card:hover { box-shadow: 0 8px 24px rgba(22,93,255,0.1); transform: translateY(-2px); }
.sj-card-head { display: flex; align-items: center; justify-content: space-between; gap: 8px; }
.sj-card-head h3 { font-size: 16px; font-weight: 700; color: #1E293B; margin: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.sj-card-ops { display: flex; gap: 2px; flex-shrink: 0; }
.sj-intro {
  font-size: 13px; color: #64748B; line-height: 1.6; margin: 10px 0;
  display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden;
  min-height: 40px;
}
.sj-devices { display: flex; flex-wrap: wrap; gap: 6px; margin-bottom: 12px; min-height: 24px; }
.sj-tag { font-weight: 500; }
.sj-no-device { font-size: 12px; color: #C0C4CC; }
.sj-meta { display: flex; flex-direction: column; gap: 3px; font-size: 12px; color: #94A3B8; margin-bottom: 12px; }
.sj-meta span { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.sj-actions { display: flex; gap: 10px; margin-top: auto; }
.sj-time { font-size: 11px; color: #C0C4CC; margin-top: 10px; text-align: right; }

.sj-upload-row { display: flex; align-items: center; gap: 10px; flex-wrap: wrap; }
.sj-file { font-size: 13px; color: #475569; display: flex; align-items: center; gap: 8px; }
.sj-file em { font-style: normal; color: #94A3B8; font-size: 12px; }
.sj-manual-block { width: 100%; }
.sj-manual-block .sj-upload-row { margin-top: 10px; }
.sj-dev-hint { font-size: 11px; color: #94A3B8; margin-top: 6px; }
.sj-manual-text {
  white-space: pre-wrap; word-break: break-word;
  font-size: 13px; line-height: 1.8; color: #475569; max-height: 50vh; overflow: auto;
}
</style>
