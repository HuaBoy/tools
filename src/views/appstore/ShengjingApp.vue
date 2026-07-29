<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { appStoreApi } from '@/api/appstore'

const apps = ref([])
const loading = ref(false)
const searchKey = ref('')

const dialogVisible = ref(false)
const dialogMode = ref('create') // create | edit
const saving = ref(false)
const formRef = ref()
const form = ref(emptyForm())
const formRules = {
  name: [{ required: true, message: '请输入应用名称', trigger: 'blur' }],
  app_key: [{ required: true, message: '请输入应用标识', trigger: 'blur' }]
}

function emptyForm() {
  return {
    id: null,
    name: '',
    app_key: '',
    package_name: '',
    version: '',
    description: '',
    icon_url: '',
    download_url: '',
    status: 1
  }
}

// 文件上传状态
const apkFile = ref(null)
const iconFile = ref(null)
const uploadingApk = ref(false)
const uploadingIcon = ref(false)

function resetUploadState() {
  apkFile.value = null
  iconFile.value = null
  uploadingApk.value = false
  uploadingIcon.value = false
}

async function onApkChange(e) {
  const f = e.target.files && e.target.files[0]
  if (!f) return
  uploadingApk.value = true
  try {
    const res = await appStoreApi.upload(f, null)
    if (res && res.code === 0 && res.data && res.data.apk_url) {
      form.value.download_url = res.data.apk_url
      apkFile.value = f
      ElMessage.success('APK 上传成功')
    } else {
      ElMessage.error('APK 上传失败')
    }
  } catch (err) {
    ElMessage.error('APK 上传失败：' + (err.message || '网络错误'))
  } finally {
    uploadingApk.value = false
  }
}

async function onIconChange(e) {
  const f = e.target.files && e.target.files[0]
  if (!f) return
  uploadingIcon.value = true
  try {
    const res = await appStoreApi.upload(null, f)
    if (res && res.code === 0 && res.data && res.data.icon_url) {
      form.value.icon_url = res.data.icon_url
      iconFile.value = f
      ElMessage.success('图标上传成功')
    } else {
      ElMessage.error('图标上传失败')
    }
  } catch (err) {
    ElMessage.error('图标上传失败：' + (err.message || '网络错误'))
  } finally {
    uploadingIcon.value = false
  }
}

async function loadApps() {
  loading.value = true
  try {
    const res = await appStoreApi.list({ name: searchKey.value || undefined })
    if (res && res.code === 0) {
      apps.value = res.data || []
    } else {
      apps.value = []
    }
  } catch (e) {
    ElMessage.error('加载应用列表失败：' + (e.message || '网络错误'))
  } finally {
    loading.value = false
  }
}

function onSearch() { loadApps() }

function openCreate() {
  dialogMode.value = 'create'
  form.value = emptyForm()
  resetUploadState()
  dialogVisible.value = true
}

function openEdit(row) {
  dialogMode.value = 'edit'
  resetUploadState()
  form.value = {
    id: row.id,
    name: row.name,
    app_key: row.app_key,
    package_name: row.package_name,
    version: row.version,
    description: row.description,
    icon_url: row.icon_url,
    download_url: row.download_url,
    status: row.status
  }
  dialogVisible.value = true
}

async function submitForm() {
  await formRef.value.validate()
  saving.value = true
  try {
    const payload = {
      name: form.value.name,
      app_key: form.value.app_key,
      package_name: form.value.package_name,
      version: form.value.version,
      description: form.value.description,
      icon_url: form.value.icon_url,
      download_url: form.value.download_url,
      status: form.value.status
    }
    let res
    if (dialogMode.value === 'create') {
      res = await appStoreApi.create(payload)
    } else {
      res = await appStoreApi.update(form.value.id, payload)
    }
    if (res && res.code === 0) {
      ElMessage.success(dialogMode.value === 'create' ? '应用创建成功' : '应用更新成功')
      dialogVisible.value = false
      loadApps()
    } else {
      ElMessage.error((res && res.message) || '保存失败')
    }
  } catch (e) {
    ElMessage.error('保存失败：' + (e.message || '网络错误'))
  } finally {
    saving.value = false
  }
}

async function removeApp(row) {
  try {
    await ElMessageBox.confirm(`确认删除应用「${row.name}」？`, '提示', {
      type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消'
    })
  } catch { return }
  try {
    const res = await appStoreApi.remove(row.id)
    if (res && res.code === 0) {
      ElMessage.success('删除成功')
      loadApps()
    } else {
      ElMessage.error((res && res.message) || '删除失败')
    }
  } catch (e) {
    ElMessage.error('删除失败：' + (e.message || '网络错误'))
  }
}

const total = computed(() => apps.value.length)
const onlineCount = computed(() => apps.value.filter(a => a.status === 1).length)

onMounted(loadApps)
</script>

<template>
  <div class="app-store">
    <div class="store-header">
      <div>
        <h2 class="store-title">盛景应用管理</h2>
        <p class="store-sub">管理 APP 应用版本，数据持久化于服务端</p>
      </div>
      <button class="btn-primary" @click="openCreate">+ 新增应用</button>
    </div>

    <div class="stat-row">
      <div class="stat-card">
        <div class="stat-label">应用总数</div>
        <div class="stat-value">{{ total }}</div>
      </div>
      <div class="stat-card">
        <div class="stat-label">上架中</div>
        <div class="stat-value online">{{ onlineCount }}</div>
      </div>
      <div class="stat-card">
        <div class="stat-label">已下架</div>
        <div class="stat-value offline">{{ total - onlineCount }}</div>
      </div>
    </div>

    <div class="store-toolbar">
      <input
        v-model="searchKey"
        class="search-input"
        placeholder="搜索应用名称 / 标识 / 包名"
        @keyup.enter="onSearch"
        @input="onSearch"
      />
    </div>

    <div v-loading="loading" class="app-grid">
      <div v-for="app in apps" :key="app.id" class="app-card">
        <div class="app-card-head">
          <img v-if="app.icon_url" :src="app.icon_url" class="app-icon" alt="icon" />
          <div v-else class="app-icon placeholder">{{ (app.name || '?').charAt(0) }}</div>
          <div class="app-meta">
            <div class="app-name">{{ app.name }}</div>
            <div class="app-pkg">{{ app.package_name || app.app_key }}</div>
          </div>
          <span class="status-tag" :class="app.status === 1 ? 'on' : 'off'">
            {{ app.status === 1 ? '上架' : '下架' }}
          </span>
        </div>
        <div class="app-version">版本 {{ app.version || '—' }}</div>
        <div class="app-desc">{{ app.description || '暂无描述' }}</div>
        <div class="app-card-foot">
          <a v-if="app.download_url" :href="app.download_url" target="_blank" class="link-btn">下载</a>
          <span v-else class="link-btn disabled">无下载</span>
          <div class="app-ops">
            <button class="op-btn" @click="openEdit(app)">编辑</button>
            <button class="op-btn danger" @click="removeApp(app)">删除</button>
          </div>
        </div>
      </div>
      <div v-if="!loading && apps.length === 0" class="empty-tip">暂无应用，点击右上角新增</div>
    </div>

    <el-dialog
      v-model="dialogVisible"
      :title="dialogMode === 'create' ? '新增应用' : '编辑应用'"
      width="520px"
    >
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="90px">
        <el-form-item label="应用名称" prop="name">
          <el-input v-model="form.name" placeholder="如：盛景智能" />
        </el-form-item>
        <el-form-item label="应用标识" prop="app_key">
          <el-input v-model="form.app_key" placeholder="如：shengjing" />
        </el-form-item>
        <el-form-item label="包名">
          <el-input v-model="form.package_name" placeholder="如：com.shengjing.app" />
        </el-form-item>
        <el-form-item label="版本号">
          <el-input v-model="form.version" placeholder="如：1.2.0" />
        </el-form-item>
        <el-form-item label="应用图标">
          <div class="upload-row">
            <input type="file" accept="image/*" @change="onIconChange" />
            <span v-if="uploadingIcon" class="up-hint">上传中…</span>
            <span v-else-if="form.icon_url" class="up-hint ok">已设置：{{ form.icon_url.split('/').pop() }}</span>
          </div>
          <el-input v-model="form.icon_url" placeholder="（可选）外部图标 URL" style="margin-top:8px" />
        </el-form-item>
        <el-form-item label="APK 文件">
          <div class="upload-row">
            <input type="file" accept=".apk" @change="onApkChange" />
            <span v-if="uploadingApk" class="up-hint">上传中…</span>
            <span v-else-if="form.download_url" class="up-hint ok">已设置安装包</span>
          </div>
          <el-input v-model="form.download_url" placeholder="（可选）外部 APK 下载 URL" style="margin-top:8px" />
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" active-text="上架" inactive-text="下架" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="应用说明 / 操作手册" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="submitForm">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.app-store { padding: 20px; color: var(--text-primary); }
.store-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 18px; }
.store-title { font-size: 20px; font-weight: 700; margin: 0; }
.store-sub { font-size: 12px; color: var(--text-tertiary); margin: 4px 0 0; }
.btn-primary {
  background: linear-gradient(135deg, #165DFF, #0F4CD0); color: #fff; border: none;
  padding: 9px 18px; border-radius: 10px; cursor: pointer; font-size: 13px; font-weight: 600;
}
.btn-primary:hover { filter: brightness(1.08); }

.stat-row { display: grid; grid-template-columns: repeat(3, 1fr); gap: 16px; margin-bottom: 16px; }
.stat-card { background: rgba(255,255,255,0.04); border: 1px solid var(--border-color); border-radius: 14px; padding: 16px 18px; }
.stat-label { font-size: 12px; color: var(--text-tertiary); }
.stat-value { font-size: 26px; font-weight: 700; margin-top: 6px; }
.stat-value.online { color: #00B42A; }
.stat-value.offline { color: #F53F3F; }

.store-toolbar { margin-bottom: 16px; }
.search-input {
  width: 320px; max-width: 100%; padding: 9px 14px; border-radius: 10px;
  background: rgba(255,255,255,0.04); border: 1px solid var(--border-color);
  color: var(--text-primary); font-size: 13px; outline: none;
}
.search-input:focus { border-color: rgba(22,93,255,0.5); }

.app-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 16px; }
.app-card { background: rgba(255,255,255,0.04); border: 1px solid var(--border-color); border-radius: 16px; padding: 16px; }
.app-card-head { display: flex; align-items: center; gap: 12px; }
.app-icon { width: 46px; height: 46px; border-radius: 12px; object-fit: cover; }
.app-icon.placeholder { display: flex; align-items: center; justify-content: center; background: rgba(22,93,255,0.2); color: #165DFF; font-size: 20px; font-weight: 700; }
.app-meta { flex: 1; min-width: 0; }
.app-name { font-size: 15px; font-weight: 600; }
.app-pkg { font-size: 12px; color: var(--text-tertiary); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.status-tag { font-size: 11px; padding: 2px 10px; border-radius: 8px; flex-shrink: 0; }
.status-tag.on { background: rgba(0,180,42,0.15); color: #00B42A; }
.status-tag.off { background: rgba(245,63,63,0.15); color: #F53F3F; }
.app-version { font-size: 12px; color: var(--text-secondary); margin: 12px 0 6px; }
.app-desc { font-size: 13px; color: var(--text-secondary); line-height: 1.6; min-height: 38px; }
.app-card-foot { display: flex; align-items: center; justify-content: space-between; margin-top: 12px; }
.link-btn { font-size: 13px; color: #165DFF; text-decoration: none; font-weight: 600; }
.link-btn.disabled { color: var(--text-tertiary); }
.app-ops { display: flex; gap: 8px; }
.upload-row { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; }
.upload-row input[type="file"] {
  font-size: 12px; color: var(--text-secondary);
  background: rgba(255,255,255,0.04); border: 1px dashed var(--border-color);
  border-radius: 8px; padding: 6px 10px; max-width: 260px;
}
.up-hint { font-size: 12px; color: var(--text-tertiary); }
.up-hint.ok { color: #00B42A; }
.op-btn { font-size: 12px; padding: 4px 12px; border-radius: 8px; cursor: pointer; background: rgba(255,255,255,0.06); border: 1px solid var(--border-color); color: var(--text-secondary); }
.op-btn:hover { border-color: rgba(22,93,255,0.4); }
.op-btn.danger:hover { border-color: rgba(245,63,63,0.5); color: #F53F3F; }
.empty-tip { grid-column: 1 / -1; text-align: center; color: var(--text-tertiary); padding: 40px; }
</style>
