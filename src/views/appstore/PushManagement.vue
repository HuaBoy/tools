<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Promotion, Delete, Refresh, Document, Box, Cpu } from '@element-plus/icons-vue'
import { localAppStore } from '@/utils/localAppStore'
import { localPushStore } from '@/utils/localPushStore'

const apps = ref([])
const tasks = ref([])
const pushing = ref(false)

const selectedAppId = ref('')
const snInputText = ref('')
const note = ref('')

// 连续编号生成
const rangePrefix = ref('')
const rangeStart = ref('')
const rangeEnd = ref('')
const rangeWidth = ref(4)

// ===== 计算 SN 列表（去重） =====
const snList = computed(() => {
  const parts = snInputText.value
    .split(/[\n,，;；\s]+/)
    .map((s) => s.trim())
    .filter(Boolean)
  return [...new Set(parts)]
})

const selectedApp = computed(() => apps.value.find((a) => a.id === selectedAppId.value) || null)

function refreshApps() {
  apps.value = localAppStore.getAll()
}

function refreshTasks() {
  tasks.value = localPushStore.getAll()
}

const unsub = localPushStore.subscribe(refreshTasks)

onMounted(() => {
  refreshApps()
  refreshTasks()
})
onUnmounted(() => unsub && unsub())

// ===== 连续编号生成，追加到输入区 =====
function addRange() {
  const prefix = rangePrefix.value.trim()
  const start = parseInt(rangeStart.value, 10)
  const end = parseInt(rangeEnd.value, 10)
  const width = parseInt(rangeWidth.value, 10) || 1
  if (isNaN(start) || isNaN(end)) {
    ElMessage.warning('请输入有效的起始 / 结束编号')
    return
  }
  const lo = Math.min(start, end)
  const hi = Math.max(start, end)
  const limit = 10000
  if (hi - lo + 1 > limit) {
    ElMessage.warning(`单次连续编号过多（>${limit}），请缩小范围`)
    return
  }
  const lines = []
  for (let i = lo; i <= hi; i++) {
    lines.push(prefix + String(i).padStart(width, '0'))
  }
  const existing = snInputText.value.trim()
  snInputText.value = existing ? existing + '\n' + lines.join('\n') : lines.join('\n')
  ElMessage.success(`已生成 ${lines.length} 个连续编号`)
}

// ===== 清空 SN 输入 =====
function clearSn() {
  snInputText.value = ''
}

// ===== 执行推送 =====
async function doPush() {
  if (!selectedAppId.value) {
    ElMessage.warning('请先选择要推送的应用（APK）')
    return
  }
  if (!snList.value.length) {
    ElMessage.warning('请输入至少一个 SN 编号')
    return
  }
  if (pushing.value) return

  const app = selectedApp.value
  if (!app.apkFile) {
    ElMessage.warning('该应用尚未上传 APK 安装包，无法推送')
    return
  }

  pushing.value = true
  try {
    await localPushStore.create({
      appId: app.id,
      appName: app.name,
      apkFileName: app.apkFile.name,
      sns: snList.value,
      note: note.value
    })
    ElMessage.success(`推送完成：共 ${snList.value.length} 台`)
    // 重置输入
    snInputText.value = ''
    note.value = ''
    selectedAppId.value = ''
  } catch (e) {
    ElMessage.error('推送失败：' + (e.message || '未知错误'))
  } finally {
    pushing.value = false
  }
}

// ===== 删除记录 =====
function handleDelete(task) {
  ElMessageBox.confirm(`确定删除「${task.appName}」的这条推送记录（共 ${task.total} 台）？`, '删除确认', {
    type: 'warning'
  }).then(() => {
    localPushStore.remove(task.id)
    ElMessage.success('已删除')
  }).catch(() => {})
}

function fmtTime(t) {
  try { return new Date(t).toLocaleString('zh-CN') } catch { return t }
}
</script>

<template>
  <div class="push-page">
    <header class="pp-header">
      <div class="ph-left">
        <div class="ph-icon">
          <el-icon><Promotion /></el-icon>
        </div>
        <div>
          <h1>推送管理</h1>
          <p>按 SN 编号推送盛景应用（APK）至目标设备，支持连续编号与多台批量推送</p>
        </div>
      </div>
    </header>

    <div class="pp-body">
      <!-- 左：新建推送 -->
      <section class="pp-panel">
        <div class="pp-panel-title">
          <el-icon><Plus /></el-icon><span>新建推送</span>
        </div>

        <!-- 选择应用 -->
        <div class="field">
          <label>推送应用（APK）</label>
          <el-select v-model="selectedAppId" placeholder="选择要推送的盛景应用" style="width: 100%">
            <el-option
              v-for="a in apps"
              :key="a.id"
              :label="a.name"
              :value="a.id"
              :disabled="!a.apkFile"
            >
              <span>{{ a.name }}</span>
              <span v-if="!a.apkFile" style="color:#F53F3F;font-size:12px">（无APK）</span>
            </el-option>
          </el-select>
          <p v-if="selectedApp && !selectedApp.apkFile" class="field-tip err">该应用未上传 APK，请先在「应用管理」补充</p>
          <p v-else-if="selectedApp" class="field-tip">将推送：{{ selectedApp.apkFile?.name }}</p>
        </div>

        <!-- 连续编号生成 -->
        <div class="field">
          <label>连续编号生成（可选）</label>
          <div class="range-row">
            <el-input v-model="rangePrefix" placeholder="前缀 如 DT40-" style="width: 110px" />
            <el-input v-model="rangeStart" placeholder="起始 如 1" style="width: 90px" />
            <span class="tilde">~</span>
            <el-input v-model="rangeEnd" placeholder="结束 如 100" style="width: 90px" />
            <el-input v-model="rangeWidth" placeholder="位数" title="编号补零位数" style="width: 70px" />
            <el-button :icon="Refresh" @click="addRange">生成并追加</el-button>
          </div>
          <p class="field-tip">示例：前缀 <b>DT40-</b>、起始 <b>1</b>、结束 <b>100</b>、位数 <b>4</b> → 生成 DT40-0001 ~ DT40-0100</p>
        </div>

        <!-- 手动/批量 SN 输入 -->
        <div class="field">
          <label>SN 编号（多台：每行 / 逗号 / 空格分隔）</label>
          <el-input
            v-model="snInputText"
            type="textarea"
            :rows="8"
            placeholder="DT40-0001&#10;DT40-0002&#10;或一行多个：DT40-0003,DT40-0004"
          />
          <div class="sn-bar">
            <span class="sn-count">已解析 <b>{{ snList.length }}</b> 台（自动去重）</span>
            <el-button text size="small" :icon="Delete" @click="clearSn">清空</el-button>
          </div>
          <div v-if="snList.length" class="sn-preview">
            <el-tag
              v-for="sn in snList.slice(0, 60)"
              :key="sn"
              size="small"
              type="info"
              effect="plain"
              class="sn-chip"
            >{{ sn }}</el-tag>
            <span v-if="snList.length > 60" class="sn-more">… 等共 {{ snList.length }} 台</span>
          </div>
        </div>

        <!-- 备注 -->
        <div class="field">
          <label>备注（可选）</label>
          <el-input v-model="note" placeholder="如：V2.3 版本灰度推送" />
        </div>

        <el-button
          type="primary"
          :icon="Promotion"
          :loading="pushing"
          :disabled="!selectedAppId || !snList.length"
          class="push-btn"
          @click="doPush"
        >
          {{ pushing ? '推送中…' : `推送至 ${snList.length} 台设备` }}
        </el-button>
      </section>

      <!-- 右：推送记录 -->
      <section class="pp-panel">
        <div class="pp-panel-title">
          <el-icon><Document /></el-icon><span>推送记录</span>
          <span class="rec-count">{{ tasks.length }} 条</span>
        </div>

        <el-empty v-if="!tasks.length" description="暂无推送记录" :image-size="80" />

        <div v-else class="rec-list">
          <div v-for="t in tasks" :key="t.id" class="rec-card">
            <div class="rec-top">
              <div class="rec-app">
                <el-icon class="rec-ico"><Box /></el-icon>
                <div>
                  <div class="rec-name">{{ t.appName }}</div>
                  <div class="rec-sub">{{ t.apkFileName }}</div>
                </div>
              </div>
              <el-tag :type="t.status === 'done' ? 'success' : 'warning'" size="small" effect="dark">
                {{ t.status === 'done' ? '已完成' : '推送中' }}
              </el-tag>
            </div>

            <div class="rec-meta">
              <span><el-icon><Cpu /></el-icon> {{ t.total }} 台</span>
              <span class="ok">成功 {{ t.success }}</span>
              <span class="fail" v-if="t.failed">失败 {{ t.failed }}</span>
              <span class="time">{{ fmtTime(t.created_at) }}</span>
            </div>

            <div v-if="t.note" class="rec-note">📝 {{ t.note }}</div>

            <!-- 逐台结果 -->
            <div class="rec-sns">
              <el-tag
                v-for="sn in t.sns"
                :key="sn"
                size="small"
                :type="t.results[sn] === 'failed' ? 'danger' : 'success'"
                effect="plain"
                class="sn-chip"
              >{{ sn }}</el-tag>
            </div>

            <div class="rec-actions">
              <el-button text type="danger" size="small" :icon="Delete" @click="handleDelete(t)">删除记录</el-button>
            </div>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<style scoped>
.push-page { width: 100%; padding: 20px 24px 32px; box-sizing: border-box; background: #F7F8FA; min-height: calc(100vh - 60px); }
.pp-header { display: flex; align-items: center; justify-content: space-between; padding: 18px 24px; margin-bottom: 16px; background: #fff; border-radius: 14px; border: 1px solid #E8ECF1; }
.ph-left { display: flex; align-items: center; gap: 14px; }
.ph-icon { width: 42px; height: 42px; border-radius: 12px; background: linear-gradient(135deg, #165DFF, #00D2AC); display: flex; align-items: center; justify-content: center; color: #fff; font-size: 20px; }
.ph-left h1 { font-size: 20px; font-weight: 700; color: #1E293B; margin: 0; line-height: 1.3; }
.ph-left p { font-size: 12px; color: #94A3B8; margin: 2px 0 0 0; }

.pp-body { display: flex; gap: 18px; align-items: flex-start; }
.pp-panel { flex: 1; min-width: 0; background: #fff; border-radius: 14px; border: 1px solid #E8ECF1; padding: 20px; }
.pp-panel-title { display: flex; align-items: center; gap: 8px; font-size: 15px; font-weight: 600; color: #1E293B; margin-bottom: 18px; }
.pp-panel-title .rec-count { margin-left: auto; font-size: 12px; color: #94A3B8; font-weight: 400; }

.field { margin-bottom: 18px; }
.field > label { display: block; font-size: 13px; color: #475569; font-weight: 500; margin-bottom: 8px; }
.field-tip { font-size: 11px; color: #94A3B8; margin: 6px 0 0; }
.field-tip.err { color: #F53F3F; }
.field-tip b { color: #165DFF; }

.range-row { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.tilde { color: #94A3B8; font-weight: 700; }

.sn-bar { display: flex; align-items: center; justify-content: space-between; margin-top: 8px; }
.sn-count { font-size: 12px; color: #64748B; }
.sn-count b { color: #165DFF; }
.sn-preview { margin-top: 10px; max-height: 120px; overflow-y: auto; display: flex; flex-wrap: wrap; gap: 6px; padding: 8px; background: #F8FAFC; border-radius: 8px; border: 1px solid #E8ECF1; }
.sn-chip { margin: 0 !important; font-family: monospace; }
.sn-more { font-size: 11px; color: #94A3B8; align-self: center; }

.push-btn { width: 100%; height: 44px; font-size: 14px; font-weight: 600; }

.rec-list { display: flex; flex-direction: column; gap: 12px; max-height: calc(100vh - 220px); overflow-y: auto; }
.rec-card { border: 1px solid #E8ECF1; border-radius: 12px; padding: 14px; background: #FCFDFE; }
.rec-top { display: flex; align-items: center; justify-content: space-between; }
.rec-app { display: flex; align-items: center; gap: 10px; min-width: 0; }
.rec-ico { color: #165DFF; font-size: 18px; }
.rec-name { font-size: 14px; font-weight: 600; color: #1E293B; }
.rec-sub { font-size: 11px; color: #94A3B8; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 280px; }
.rec-meta { display: flex; align-items: center; gap: 14px; margin: 10px 0 6px; font-size: 12px; color: #64748B; }
.rec-meta .ok { color: #00B42A; font-weight: 600; }
.rec-meta .fail { color: #F53F3F; font-weight: 600; }
.rec-meta .time { margin-left: auto; color: #B4BCCC; }
.rec-note { font-size: 12px; color: #64748B; background: #F8FAFC; border-radius: 6px; padding: 6px 10px; margin-bottom: 8px; }
.rec-sns { display: flex; flex-wrap: wrap; gap: 6px; margin-bottom: 8px; max-height: 130px; overflow-y: auto; }
.rec-actions { display: flex; justify-content: flex-end; border-top: 1px solid #F0F2F5; padding-top: 8px; }

@media screen and (max-width: 1000px) {
  .pp-body { flex-direction: column; }
}
</style>
