<template>
  <div class="sales-workbench">
    <!-- 头部 -->
    <div class="wb-header">
      <div>
        <h2 class="wb-title">销售岗 · 民爆行业洞察</h2>
        <p class="wb-sub">
          数据来源：工信部安全生产司 / 中国爆破器材行业协会 / 各大集团 / 券商研报 / 第三方机构 / 省民爆协会
        </p>
      </div>
      <div class="header-tools">
        <div class="live-badge"><span class="dot"></span>实时 {{ clock }}</div>
        <button class="refresh-btn" :disabled="loading" @click="loadData">刷新</button>
        <div class="seg">
          <button :class="{ active: viewMode === 'public' }" @click="viewMode = 'public'">公开版</button>
          <button :class="{ active: viewMode === 'internal' }" @click="viewMode = 'internal'">内部版</button>
        </div>
      </div>
    </div>

    <div class="note-banner" v-if="meta.note">⚠ {{ meta.note }}</div>

    <div class="update-bar">
      <span>数据版本：{{ meta.updatedAt || '—' }}</span>
      <span>最近刷新：{{ lastRefresh || '—' }}</span>
      <span class="mode-hint" v-if="viewMode === 'public'">公开版已隐藏企业级/明细数据</span>
      <span class="mode-hint internal" v-else>内部版（含企业级对标与省级明细，注意不外发）</span>
    </div>

    <!-- Tab 切换 -->
    <div class="tabs">
      <button
        v-for="t in tabs"
        :key="t.key"
        class="tab-btn"
        :class="{ active: activeTab === t.key }"
        @click="activeTab = t.key"
      >
        {{ t.label }}
      </button>
    </div>

    <!-- 全国总量 -->
    <section v-show="activeTab === 'national'" class="tab-panel">
      <div class="kpi-grid">
        <div v-for="k in nationalView" :key="k.label" class="kpi-card">
          <div class="kpi-label">{{ k.label }}</div>
          <div class="kpi-value">
            {{ k.value }}<span class="kpi-unit">{{ k.unit }}</span>
          </div>
          <div v-if="k.trend" class="kpi-trend" :class="k.trendUp ? 'up' : 'down'">{{ k.trend }}</div>
          <div class="src-tag">来源：{{ k.source }}</div>
        </div>
      </div>
    </section>

    <!-- 区域产销分布 -->
    <section v-show="activeTab === 'regional'" class="tab-panel">
      <div class="panel-card">
        <h3 class="panel-title">各大区工业炸药产量分布（公开版）</h3>
        <table class="data-table">
          <thead>
            <tr><th>区域</th><th>工业炸药产量（万吨）</th><th>占比</th></tr>
          </thead>
          <tbody>
            <tr v-for="r in regional.macro" :key="r.region">
              <td>{{ r.region }}</td>
              <td>{{ r.output }}</td>
              <td>{{ r.share }}</td>
            </tr>
          </tbody>
        </table>
      </div>
      <div class="panel-card internal-box" v-if="viewMode === 'internal'">
        <div class="internal-badge">内部版 · 省级明细（对外不披露）</div>
        <h3 class="panel-title">重点省份产量明细</h3>
        <table class="data-table">
          <thead>
            <tr><th>省份</th><th>工业炸药产量（万吨）</th></tr>
          </thead>
          <tbody>
            <tr v-for="r in regional.province" :key="r.region">
              <td>{{ r.region }}</td>
              <td>{{ r.output }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <!-- CR20 集中度 -->
    <section v-show="activeTab === 'cr20'" class="tab-panel">
      <div class="kpi-grid">
        <div v-for="m in cr20Metrics" :key="m.label" class="kpi-card">
          <div class="kpi-label">{{ m.label }}</div>
          <div class="kpi-value">
            {{ m.value }}<span class="kpi-unit">{{ m.unit }}</span>
          </div>
          <div v-if="m.sub" class="kpi-sub">{{ m.sub }}</div>
          <div class="src-tag">来源：{{ cr20.source }}</div>
        </div>
      </div>
      <div class="panel-card note-box">
        <p>{{ cr20.publicNote }}</p>
      </div>
      <div class="panel-card internal-box" v-if="viewMode === 'internal' && cr20.internal">
        <div class="internal-badge">内部版 · 本集团对标</div>
        <h3 class="panel-title">{{ cr20.internal.label }}</h3>
        <p class="internal-text">{{ cr20.internal.text }}</p>
      </div>
    </section>

    <!-- 行业动态资讯 -->
    <section v-show="activeTab === 'news'" class="tab-panel">
      <div class="news-list">
        <div v-for="(n, i) in newsView" :key="i" class="news-card">
          <div class="news-head">
            <span class="news-tag" :class="{ internal: n.visibility === 'internal' }">{{ n.tag }}</span>
            <span class="news-date">{{ n.date }}</span>
          </div>
          <h4 class="news-title">{{ n.title }}</h4>
          <p class="news-desc">{{ n.desc }}</p>
          <div class="src-tag">来源：{{ n.source }}</div>
        </div>
      </div>
    </section>

    <!-- 页脚数据来源 -->
    <div class="footer-src">
      <span class="footer-label">数据来源：</span>
      <span v-for="s in sources" :key="s" class="footer-src-item">{{ s }}</span>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { salesApi } from '@/api/sales'

const route = useRoute()

const tabs = [
  { key: 'national', label: '全国总量' },
  { key: 'regional', label: '区域产销分布' },
  { key: 'cr20', label: 'CR20集中度' },
  { key: 'news', label: '行业动态资讯' }
]

const activeTab = ref(
  route.query.tab && tabs.find((t) => t.key === route.query.tab) ? route.query.tab : 'national'
)
const viewMode = ref('public') // public | internal
const loading = ref(false)

const meta = ref({ updatedAt: '', note: '' })
const sources = ref([])
const national = ref([])
const regional = ref({ macro: [], province: [] })
const cr20 = ref({})
const news = ref([])

const clock = ref('')
const lastRefresh = ref('')

// 可见性过滤：public 始终显示；internal 仅内部版显示
function filterByView(list) {
  return (list || []).filter((it) => (it.visibility || 'public') === 'public' || viewMode.value === 'internal')
}

const nationalView = computed(() => filterByView(national.value))
const newsView = computed(() => filterByView(news.value))
const cr20Metrics = computed(() => filterByView(cr20.value.metrics))

function pad(n) {
  return String(n).padStart(2, '0')
}
function formatNow() {
  const d = new Date()
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(
    d.getMinutes()
  )}:${pad(d.getSeconds())}`
}

async function loadData() {
  loading.value = true
  try {
    const d = await salesApi.getSalesData()
    meta.value = d.meta || {}
    sources.value = d.sources || []
    national.value = d.national || []
    regional.value = d.regional || { macro: [], province: [] }
    cr20.value = d.cr20 || {}
    news.value = d.news || []
    lastRefresh.value = formatNow()
  } finally {
    loading.value = false
  }
}

let clockTimer = null
let refreshTimer = null

onMounted(() => {
  clock.value = formatNow()
  loadData()
  clockTimer = setInterval(() => {
    clock.value = formatNow()
  }, 1000)
  // 实时刷新：每 60 秒拉取一次最新数据（后续对接外部源即自动更新）
  refreshTimer = setInterval(() => {
    loadData()
  }, 60000)
})

onUnmounted(() => {
  if (clockTimer) clearInterval(clockTimer)
  if (refreshTimer) clearInterval(refreshTimer)
})
</script>

<style scoped lang="scss">
.sales-workbench {
  padding: 16px 20px 28px;
  color: var(--text-primary, #e6edf3);
}

.wb-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  flex-wrap: wrap;
  gap: 12px;
}
.wb-title {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
}
.wb-sub {
  margin: 4px 0 0;
  font-size: 12px;
  color: var(--text-muted, #8b97a6);
}
.header-tools {
  display: flex;
  align-items: center;
  gap: 10px;
}
.live-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--text-muted, #8b97a6);
  font-variant-numeric: tabular-nums;
}
.live-badge .dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #21ba72;
  box-shadow: 0 0 0 0 rgba(33, 186, 114, 0.6);
  animation: pulse 1.6s infinite;
}
@keyframes pulse {
  0% { box-shadow: 0 0 0 0 rgba(33, 186, 114, 0.6); }
  70% { box-shadow: 0 0 0 8px rgba(33, 186, 114, 0); }
  100% { box-shadow: 0 0 0 0 rgba(33, 186, 114, 0); }
}
.refresh-btn {
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(255, 255, 255, 0.12);
  color: var(--text-primary, #e6edf3);
  border-radius: 6px;
  padding: 5px 12px;
  font-size: 12px;
  cursor: pointer;
}
.refresh-btn:hover { background: rgba(255, 255, 255, 0.12); }
.refresh-btn:disabled { opacity: 0.5; cursor: default; }

.seg {
  display: inline-flex;
  border: 1px solid rgba(255, 255, 255, 0.14);
  border-radius: 6px;
  overflow: hidden;
}
.seg button {
  background: transparent;
  border: none;
  color: var(--text-muted, #8b97a6);
  padding: 5px 12px;
  font-size: 12px;
  cursor: pointer;
}
.seg button.active {
  background: var(--accent-color, #165dff);
  color: #fff;
}

.note-banner {
  margin-top: 12px;
  padding: 8px 12px;
  background: rgba(255, 179, 64, 0.12);
  border: 1px solid rgba(255, 179, 64, 0.35);
  border-radius: 6px;
  font-size: 12px;
  color: #ffcf80;
}
.update-bar {
  margin-top: 10px;
  display: flex;
  gap: 18px;
  flex-wrap: wrap;
  font-size: 12px;
  color: var(--text-muted, #8b97a6);
}
.mode-hint { color: #5b8def; }
.mode-hint.internal { color: #ff9d6e; }

.tabs {
  margin-top: 16px;
  display: flex;
  gap: 6px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}
.tab-btn {
  background: transparent;
  border: none;
  border-bottom: 2px solid transparent;
  color: var(--text-muted, #8b97a6);
  padding: 8px 14px;
  font-size: 14px;
  cursor: pointer;
}
.tab-btn.active {
  color: var(--text-primary, #e6edf3);
  border-bottom-color: var(--accent-color, #165dff);
}
.tab-panel { margin-top: 16px; }

.kpi-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: 14px;
}
.kpi-card {
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 10px;
  padding: 16px;
}
.kpi-label { font-size: 13px; color: var(--text-muted, #8b97a6); }
.kpi-value {
  margin-top: 8px;
  font-size: 26px;
  font-weight: 700;
  font-variant-numeric: tabular-nums;
}
.kpi-unit { font-size: 13px; font-weight: 500; color: var(--text-muted, #8b97a6); margin-left: 4px; }
.kpi-trend { margin-top: 6px; font-size: 12px; }
.kpi-trend.up { color: #21ba72; }
.kpi-trend.down { color: #f0483e; }
.kpi-sub { margin-top: 6px; font-size: 12px; color: var(--text-muted, #8b97a6); }

.panel-card {
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 10px;
  padding: 16px;
  margin-bottom: 14px;
}
.panel-title { margin: 0 0 12px; font-size: 15px; font-weight: 600; }

.data-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}
.data-table th,
.data-table td {
  text-align: left;
  padding: 9px 12px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}
.data-table th { color: var(--text-muted, #8b97a6); font-weight: 500; }
.data-table tbody tr:hover { background: rgba(255, 255, 255, 0.03); }

.note-box p {
  margin: 0;
  font-size: 13px;
  color: var(--text-muted, #8b97a6);
  line-height: 1.7;
}

.internal-box {
  border-color: rgba(255, 157, 110, 0.3);
  background: rgba(255, 157, 110, 0.06);
}
.internal-badge {
  display: inline-block;
  font-size: 11px;
  color: #ff9d6e;
  border: 1px solid rgba(255, 157, 110, 0.4);
  border-radius: 4px;
  padding: 2px 8px;
  margin-bottom: 10px;
}
.internal-text { margin: 6px 0 0; font-size: 13px; line-height: 1.7; }

.news-list {
  display: grid;
  gap: 12px;
}
.news-card {
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 10px;
  padding: 14px 16px;
}
.news-head {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 6px;
}
.news-tag {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 4px;
  background: rgba(91, 141, 239, 0.18);
  color: #6fa0ff;
}
.news-tag.internal {
  background: rgba(255, 157, 110, 0.18);
  color: #ff9d6e;
}
.news-date { font-size: 12px; color: var(--text-muted, #8b97a6); }
.news-title { margin: 0; font-size: 15px; font-weight: 600; }
.news-desc { margin: 6px 0 8px; font-size: 13px; color: #c4cdd8; line-height: 1.6; }

.src-tag { font-size: 11px; color: var(--text-muted, #8b97a6); }

.footer-src {
  margin-top: 22px;
  padding-top: 14px;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
  font-size: 11px;
  color: var(--text-muted, #8b97a6);
  line-height: 1.9;
}
.footer-label { color: #aab4c0; }
.footer-src-item {
  display: inline-block;
  margin-right: 10px;
  padding: 1px 6px;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 4px;
}
</style>
