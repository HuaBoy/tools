<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import GlassCard from '@/components/GlassCard.vue'
import { listOperationManuals } from '@/api/operationManual'

const manuals = ref([])
const total = ref(0)
const loading = ref(false)
const searchQuery = ref('')
const selectedCategory = ref('all')
const currentPage = ref(1)
const pageSize = 20

const categories = [
  { value: 'all', label: '全部手册' },
  { value: 'install', label: '安装说明' },
  { value: 'operation', label: '操作指南' },
  { value: 'maintain', label: '维护保养' },
  { value: 'trouble', label: '故障排除' }
]

async function fetchData() {
  loading.value = true
  try {
    const params = { current: currentPage.value, size: pageSize }
    if (selectedCategory.value !== 'all') params.category = selectedCategory.value
    if (searchQuery.value.trim()) params.keyword = searchQuery.value.trim()
    const res = await listOperationManuals(params)
    manuals.value = res.data?.data || []
    total.value = res.data?.total || 0
  } catch (e) {
    ElMessage.error('加载手册列表失败')
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
  const map = { install: '安装', operation: '操作', maintain: '维护', trouble: '故障' }
  return map[cat] || cat
}
const categoryColor = (cat) => {
  const map = { install: '#36CFC9', operation: '#165DFF', maintain: '#00B42A', trouble: '#F77234' }
  return map[cat] || '#888'
}

onMounted(fetchData)
</script>

<template>
  <div class="operation-manual">
    <GlassCard title="海外操作手册">
      <div class="toolbar">
        <div class="search-wrap">
          <input v-model="searchQuery" type="text" class="search-input" placeholder="搜索手册名称或描述..." @keyup.enter="onSearch" />
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

    <div v-if="loading" class="loading-state">加载中...</div>

    <div v-else class="manual-list">
      <div v-for="item in manuals" :key="item.id" class="manual-card">
        <div class="card-body">
          <div class="card-header">
            <span class="category-badge"
              :style="{ background: categoryColor(item.category) + '20', color: categoryColor(item.category) }">{{ categoryBadge(item.category) }}</span>
            <span class="update-date">{{ item.updated_at?.substring(0, 10) }}</span>
          </div>
          <h3 class="manual-title">{{ item.title }}</h3>
          <p class="manual-desc">{{ item.description }}</p>
          <div class="card-footer">
            <span class="file-size">{{ item.file_size }}</span>
            <a :href="item.file_url" target="_blank" class="download-btn">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
                <polyline points="7 10 12 15 17 10" /><line x1="12" y1="15" x2="12" y2="3" />
              </svg>
              下载手册
            </a>
          </div>
        </div>
      </div>
      <div v-if="!loading && manuals.length === 0" class="empty-state">
        <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="var(--text-tertiary)" stroke-width="1.5">
          <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" /><polyline points="14 2 14 8 20 8" />
        </svg>
        <p>暂无匹配的手册</p>
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
.operation-manual { max-width: 100%; }
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

.manual-list { display: grid; grid-template-columns: repeat(auto-fill, minmax(340px, 1fr)); gap: 16px; margin-top: 20px; }
.manual-card { background: var(--bg-card); border: 1px solid var(--border-color); border-radius: 12px; transition: all 0.2s; overflow: hidden; }
.manual-card:hover { border-color: rgba(22, 93, 255, 0.3); box-shadow: 0 4px 20px rgba(0,0,0,0.08); transform: translateY(-2px); }
.card-body { padding: 20px; }
.card-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
.category-badge { padding: 2px 10px; border-radius: 4px; font-size: 12px; font-weight: 500; }
.update-date { font-size: 12px; color: var(--text-tertiary); }
.manual-title { font-size: 16px; font-weight: 600; color: var(--text-primary); margin: 0 0 8px 0; line-height: 1.4; }
.manual-desc { font-size: 13px; color: var(--text-secondary); line-height: 1.6; margin: 0 0 16px 0; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
.card-footer { display: flex; justify-content: space-between; align-items: center; padding-top: 12px; border-top: 1px solid var(--border-color); }
.file-size { font-size: 12px; color: var(--text-tertiary); }
.download-btn { display: flex; align-items: center; gap: 6px; padding: 6px 14px; border-radius: 6px; background: rgba(22, 93, 255, 0.08); border: 1px solid rgba(22, 93, 255, 0.2); color: #165DFF; font-size: 12px; cursor: pointer; text-decoration: none; transition: all 0.2s; }
.download-btn:hover { background: rgba(22, 93, 255, 0.16); }
.empty-state { grid-column: 1 / -1; display: flex; flex-direction: column; align-items: center; gap: 12px; padding: 60px 20px; color: var(--text-tertiary); }
.empty-state p { margin: 0; font-size: 14px; }
.pagination { display: flex; justify-content: center; gap: 6px; margin-top: 24px; flex-wrap: wrap; }
.page-btn { padding: 6px 14px; border-radius: 6px; border: 1px solid var(--border-color); background: var(--bg-card); color: var(--text-secondary); font-size: 13px; cursor: pointer; transition: all 0.2s; }
.page-btn:hover:not(:disabled):not(.active) { border-color: rgba(22, 93, 255, 0.4); color: var(--text-primary); }
.page-btn.active { background: rgba(22, 93, 255, 0.1); border-color: rgba(22, 93, 255, 0.4); color: #165DFF; }
.page-btn:disabled { opacity: 0.4; cursor: not-allowed; }
</style>
