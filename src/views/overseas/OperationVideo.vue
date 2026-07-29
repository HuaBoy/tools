<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import GlassCard from '@/components/GlassCard.vue'
import { listOperationVideos } from '@/api/operationVideo'

const videos = ref([])
const total = ref(0)
const loading = ref(false)
const searchQuery = ref('')
const selectedCategory = ref('all')
const currentPage = ref(1)
const pageSize = 12
const playingVideo = ref(null)

const categories = [
  { value: 'all', label: '全部视频' },
  { value: 'install', label: '安装教程' },
  { value: 'operate', label: '操作培训' },
  { value: 'maintain', label: '维护指导' }
]

async function fetchData() {
  loading.value = true
  try {
    const params = { current: currentPage.value, size: pageSize }
    if (selectedCategory.value !== 'all') params.category = selectedCategory.value
    if (searchQuery.value.trim()) params.keyword = searchQuery.value.trim()
    const res = await listOperationVideos(params)
    videos.value = res.data?.data || []
    total.value = res.data?.total || 0
  } catch (e) {
    ElMessage.error('加载视频列表失败')
  } finally {
    loading.value = false
  }
}

const totalPages = computed(() => Math.ceil(total.value / pageSize))

function onCategoryChange(cat) {
  selectedCategory.value = cat
  currentPage.value = 1
  fetchData()
}
function onSearch() {
  currentPage.value = 1
  fetchData()
}
function onPageChange(page) {
  currentPage.value = page
  fetchData()
}

const categoryBadge = (cat) => {
  const map = { install: '安装教程', operate: '操作培训', maintain: '维护指导' }
  return map[cat] || cat
}
const categoryColor = (cat) => {
  const map = { install: '#36CFC9', operate: '#165DFF', maintain: '#00B42A' }
  return map[cat] || '#888'
}
function formatViews(n) {
  if (n >= 10000) return (n / 10000).toFixed(1) + '万'
  if (n >= 1000) return (n / 1000).toFixed(1) + 'k'
  return n || 0
}
function playVideo(video) { playingVideo.value = video }
function closePlayer() { playingVideo.value = null }

onMounted(fetchData)
</script>

<template>
  <div class="operation-video">
    <GlassCard title="海外操作视频">
      <div class="toolbar">
        <div class="search-wrap">
          <input v-model="searchQuery" type="text" class="search-input" placeholder="搜索视频名称或描述..." @keyup.enter="onSearch" />
          <svg class="search-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="11" cy="11" r="8" /><line x1="21" y1="21" x2="16.65" y2="16.65" />
          </svg>
        </div>
        <div class="category-tabs">
          <button v-for="cat in categories" :key="cat.value"
            :class="['cat-tab', { active: selectedCategory === cat.value }]"
            @click="onCategoryChange(cat.value)">{{ cat.label }}</button>
        </div>
      </div>
    </GlassCard>

    <!-- 视频播放器浮层 -->
    <div v-if="playingVideo" class="video-player-overlay" @click.self="closePlayer">
      <div class="video-player-card">
        <div class="player-header">
          <h3>{{ playingVideo.title }}</h3>
          <button class="close-btn" @click="closePlayer">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18" /><line x1="6" y1="6" x2="18" y2="18" />
            </svg>
          </button>
        </div>
        <div class="player-wrapper">
          <iframe :src="playingVideo.url" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>
        </div>
        <p class="player-desc">{{ playingVideo.description }}</p>
      </div>
    </div>

    <div v-if="loading" class="loading-state">加载中...</div>

    <div v-else class="video-grid">
      <div v-for="video in videos" :key="video.id" class="video-card" @click="playVideo(video)">
        <div class="video-thumb">
          <div class="play-overlay">
            <svg width="48" height="48" viewBox="0 0 24 24" fill="white"><polygon points="5 3 19 12 5 21 5 3" /></svg>
          </div>
          <div class="duration-badge">{{ video.duration }}</div>
        </div>
        <div class="video-info">
          <div class="video-meta">
            <span class="category-badge" :style="{ background: categoryColor(video.category) + '20', color: categoryColor(video.category) }">{{ categoryBadge(video.category) }}</span>
            <span class="view-count">{{ formatViews(video.views) }} 次播放</span>
          </div>
          <h4 class="video-title">{{ video.title }}</h4>
          <p class="video-desc">{{ video.description }}</p>
        </div>
      </div>

      <div v-if="!loading && videos.length === 0" class="empty-state">
        <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="var(--text-tertiary)" stroke-width="1.5">
          <polygon points="23 7 16 12 23 17 23 7" /><rect x="1" y="5" width="15" height="14" rx="2" ry="2" />
        </svg>
        <p>暂无匹配的视频</p>
      </div>
    </div>

    <div v-if="totalPages > 1" class="pagination">
      <button :disabled="currentPage <= 1" class="page-btn" @click="onPageChange(currentPage - 1)">上一页</button>
      <button v-for="p in totalPages" :key="p" :class="['page-btn', { active: p === currentPage }]" @click="onPageChange(p)">{{ p }}</button>
      <button :disabled="currentPage >= totalPages" class="page-btn" @click="onPageChange(currentPage + 1)">下一页</button>
    </div>
  </div>
</template>

<style scoped>
.operation-video { max-width: 100%; }
.toolbar { display: flex; flex-direction: column; gap: 16px; }
.search-wrap { position: relative; width: 100%; max-width: 400px; }
.search-input { width: 100%; padding: 10px 14px 10px 38px; background: var(--bg-input); border: 1px solid var(--border-color); border-radius: 8px; color: var(--text-primary); font-size: 14px; outline: none; box-sizing: border-box; }
.search-input:focus { border-color: rgba(22, 93, 255, 0.6); }
.search-input::placeholder { color: var(--text-tertiary); }
.search-icon { position: absolute; left: 12px; top: 50%; transform: translateY(-50%); color: var(--text-tertiary); }
.category-tabs { display: flex; gap: 8px; flex-wrap: wrap; }
.cat-tab { padding: 6px 16px; border-radius: 20px; border: 1px solid var(--border-color); background: transparent; color: var(--text-secondary); font-size: 13px; cursor: pointer; transition: all 0.2s; }
.cat-tab:hover { border-color: rgba(22, 93, 255, 0.4); color: var(--text-primary); }
.cat-tab.active { background: rgba(22, 93, 255, 0.1); border-color: rgba(22, 93, 255, 0.4); color: #165DFF; }
.loading-state { text-align: center; padding: 60px; color: var(--text-tertiary); font-size: 14px; }

.video-player-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.7); display: flex; align-items: center; justify-content: center; z-index: 1000; padding: 20px; }
.video-player-card { background: var(--bg-card); border-radius: 12px; width: 100%; max-width: 800px; overflow: hidden; }
.player-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; border-bottom: 1px solid var(--border-color); }
.player-header h3 { margin: 0; font-size: 16px; color: var(--text-primary); }
.close-btn { background: none; border: none; color: var(--text-secondary); cursor: pointer; padding: 4px; border-radius: 4px; }
.close-btn:hover { background: var(--bg-hover); }
.player-wrapper { position: relative; padding-bottom: 56.25%; height: 0; }
.player-wrapper iframe { position: absolute; inset: 0; width: 100%; height: 100%; }
.player-desc { padding: 12px 20px; margin: 0; font-size: 13px; color: var(--text-secondary); line-height: 1.6; }

.video-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); gap: 20px; margin-top: 20px; }
.video-card { background: var(--bg-card); border: 1px solid var(--border-color); border-radius: 12px; overflow: hidden; cursor: pointer; transition: all 0.2s; }
.video-card:hover { border-color: rgba(22, 93, 255, 0.3); box-shadow: 0 4px 20px rgba(0,0,0,0.08); transform: translateY(-2px); }
.video-thumb { position: relative; background: linear-gradient(135deg, #1a1a2e, #16213e); height: 180px; display: flex; align-items: center; justify-content: center; }
.play-overlay { opacity: 0.8; transition: opacity 0.2s; }
.video-card:hover .play-overlay { opacity: 1; }
.duration-badge { position: absolute; bottom: 8px; right: 8px; background: rgba(0,0,0,0.75); color: white; padding: 2px 8px; border-radius: 4px; font-size: 12px; }
.video-info { padding: 14px; }
.video-meta { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; }
.category-badge { padding: 2px 8px; border-radius: 4px; font-size: 11px; font-weight: 500; }
.view-count { font-size: 12px; color: var(--text-tertiary); }
.video-title { font-size: 15px; font-weight: 600; color: var(--text-primary); margin: 0 0 6px 0; line-height: 1.4; }
.video-desc { font-size: 13px; color: var(--text-secondary); line-height: 1.5; margin: 0; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
.empty-state { grid-column: 1 / -1; display: flex; flex-direction: column; align-items: center; gap: 12px; padding: 60px 20px; color: var(--text-tertiary); }
.empty-state p { margin: 0; font-size: 14px; }
.pagination { display: flex; justify-content: center; gap: 6px; margin-top: 24px; flex-wrap: wrap; }
.page-btn { padding: 6px 14px; border-radius: 6px; border: 1px solid var(--border-color); background: var(--bg-card); color: var(--text-secondary); font-size: 13px; cursor: pointer; transition: all 0.2s; }
.page-btn:hover:not(:disabled):not(.active) { border-color: rgba(22, 93, 255, 0.4); color: var(--text-primary); }
.page-btn.active { background: rgba(22, 93, 255, 0.1); border-color: rgba(22, 93, 255, 0.4); color: #165DFF; }
.page-btn:disabled { opacity: 0.4; cursor: not-allowed; }
</style>
