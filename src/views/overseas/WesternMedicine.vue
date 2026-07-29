<script setup>
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import GlassCard from '@/components/GlassCard.vue'

const searchQuery = ref('')
const selectedCategory = ref('all')
const currentPage = ref(1)
const pageSize = ref(12)

const categories = [
  { value: 'all', label: '全部药品' },
  { value: 'antibiotic', label: '抗生素' },
  { value: 'cardiovascular', label: '心血管' },
  { value: 'digestive', label: '消化系统' },
  { value: 'respiratory', label: '呼吸系统' },
  { value: 'nervous', label: '神经系统' },
  { value: 'other', label: '其他' }
]

const medicines = [
  { id: 1, name: '阿莫西林胶囊', enName: 'Amoxicillin Capsules', category: 'antibiotic', spec: '0.5g×24粒', manufacturer: '华北制药', price: '¥12.50', stock: 560, expiry: '2027-06' },
  { id: 2, name: '头孢克肟分散片', enName: 'Cefixime Dispersible Tablets', category: 'antibiotic', spec: '0.1g×12片', manufacturer: '广州白云山', price: '¥28.00', stock: 320, expiry: '2027-03' },
  { id: 3, name: '阿托伐他汀钙片', enName: 'Atorvastatin Calcium Tablets', category: 'cardiovascular', spec: '10mg×28片', manufacturer: '辉瑞制药', price: '¥45.60', stock: 180, expiry: '2026-12' },
  { id: 4, name: '硝苯地平缓释片', enName: 'Nifedipine Extended-release Tablets', category: 'cardiovascular', spec: '30mg×30片', manufacturer: '拜耳医药', price: '¥32.80', stock: 240, expiry: '2027-01' },
  { id: 5, name: '奥美拉唑肠溶胶囊', enName: 'Omeprazole Enteric Capsules', category: 'digestive', spec: '20mg×21粒', manufacturer: '丽珠集团', price: '¥18.90', stock: 420, expiry: '2027-04' },
  { id: 6, name: '多潘立酮片', enName: 'Domperidone Tablets', category: 'digestive', spec: '10mg×30片', manufacturer: '西安杨森', price: '¥15.20', stock: 380, expiry: '2026-11' },
  { id: 7, name: '孟鲁司特钠片', enName: 'Montelukast Sodium Tablets', category: 'respiratory', spec: '10mg×28片', manufacturer: '默沙东', price: '¥38.50', stock: 150, expiry: '2027-02' },
  { id: 8, name: '布地奈德吸入剂', enName: 'Budesonide Inhalation', category: 'respiratory', spec: '200μg×200吸', manufacturer: '阿斯利康', price: '¥156.00', stock: 60, expiry: '2026-09' },
  { id: 9, name: '甲钴胺片', enName: 'Mecobalamin Tablets', category: 'nervous', spec: '0.5mg×100片', manufacturer: '卫材药业', price: '¥42.00', stock: 290, expiry: '2027-05' },
  { id: 10, name: '盐酸氟桂利嗪胶囊', enName: 'Flunarizine Hydrochloride Capsules', category: 'nervous', spec: '5mg×20粒', manufacturer: '西安杨森', price: '¥26.30', stock: 210, expiry: '2026-10' },
  { id: 11, name: '氯雷他定片', enName: 'Loratadine Tablets', category: 'other', spec: '10mg×12片', manufacturer: '拜耳医药', price: '¥19.80', stock: 450, expiry: '2027-07' },
  { id: 12, name: '维生素C片', enName: 'Vitamin C Tablets', category: 'other', spec: '0.1g×100片', manufacturer: '东北制药', price: '¥5.60', stock: 890, expiry: '2028-01' },
  { id: 13, name: '左氧氟沙星片', enName: 'Levofloxacin Tablets', category: 'antibiotic', spec: '0.5g×10片', manufacturer: '第一三共', price: '¥22.40', stock: 340, expiry: '2027-03' },
  { id: 14, name: '缬沙坦胶囊', enName: 'Valsartan Capsules', category: 'cardiovascular', spec: '80mg×28粒', manufacturer: '诺华制药', price: '¥36.70', stock: 190, expiry: '2026-08' },
  { id: 15, name: '铝碳酸镁咀嚼片', enName: 'Hydrotalcite Chewable Tablets', category: 'digestive', spec: '0.5g×48片', manufacturer: '拜耳医药', price: '¥25.90', stock: 270, expiry: '2027-02' }
]

const filteredList = computed(() => {
  let list = medicines
  if (selectedCategory.value !== 'all') {
    list = list.filter(m => m.category === selectedCategory.value)
  }
  if (searchQuery.value.trim()) {
    const q = searchQuery.value.toLowerCase()
    list = list.filter(m =>
      m.name.toLowerCase().includes(q) ||
      m.enName.toLowerCase().includes(q) ||
      m.manufacturer.toLowerCase().includes(q)
    )
  }
  return list
})

const paginatedList = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  return filteredList.value.slice(start, start + pageSize.value)
})

const total = computed(() => filteredList.value.length)
const totalPages = computed(() => Math.ceil(total.value / pageSize.value))

const categoryLabel = (v) => categories.find(c => c.value === v)?.label || v

function handlePageChange(page) { currentPage.value = page }

function getStockStatus(stock) {
  if (stock <= 100) return { label: '低库存', color: '#F77234' }
  if (stock <= 300) return { label: '中等', color: '#F6AB16' }
  return { label: '充足', color: '#00B42A' }
}
</script>

<template>
  <div class="western-medicine">
    <GlassCard title="西药管理">
      <div class="toolbar">
        <div class="search-wrap">
          <input
            v-model="searchQuery"
            type="text"
            class="search-input"
            placeholder="搜索药品名称/英文名/生产商..."
          />
          <svg class="search-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="11" cy="11" r="8" /><line x1="21" y1="21" x2="16.65" y2="16.65" />
          </svg>
        </div>
        <div class="category-tabs">
          <button
            v-for="cat in categories"
            :key="cat.value"
            :class="['cat-tab', { active: selectedCategory === cat.value }]"
            @click="selectedCategory = cat.value; currentPage = 1"
          >{{ cat.label }}</button>
        </div>
      </div>
    </GlassCard>

    <div class="medicine-list">
      <div
        v-for="item in paginatedList"
        :key="item.id"
        class="medicine-card"
      >
        <div class="card-body">
          <div class="card-top">
            <span class="drug-icon">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                <rect x="4" y="4" width="16" height="16" rx="2" />
                <line x1="8" y1="9" x2="16" y2="9" />
                <line x1="8" y1="13" x2="12" y2="13" />
              </svg>
            </span>
            <span class="category-tag">{{ categoryLabel(item.category) }}</span>
          </div>
          <h3 class="drug-name">{{ item.name }}</h3>
          <p class="drug-en-name">{{ item.enName }}</p>
          <div class="drug-details">
            <div class="detail-row">
              <span class="detail-label">规格</span>
              <span class="detail-value">{{ item.spec }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">生产商</span>
              <span class="detail-value">{{ item.manufacturer }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">效期</span>
              <span class="detail-value">{{ item.expiry }}</span>
            </div>
          </div>
          <div class="card-footer">
            <div class="price-stock">
              <span class="price">{{ item.price }}</span>
              <span
                class="stock-badge"
                :style="{ background: getStockStatus(item.stock).color + '20', color: getStockStatus(item.stock).color }"
              >{{ getStockStatus(item.stock).label }} ({{ item.stock }})</span>
            </div>
          </div>
        </div>
      </div>

      <div v-if="filteredList.length === 0" class="empty-state">
        <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="var(--text-tertiary)" stroke-width="1.5">
          <rect x="4" y="4" width="16" height="16" rx="2" />
          <line x1="8" y1="9" x2="16" y2="9" />
          <line x1="8" y1="13" x2="12" y2="13" />
        </svg>
        <p>暂无匹配的药品</p>
      </div>
    </div>

    <div v-if="totalPages > 1" class="pagination">
      <button
        :disabled="currentPage <= 1"
        class="page-btn"
        @click="handlePageChange(currentPage - 1)"
      >上一页</button>
      <template v-for="p in totalPages" :key="p">
        <button
          :class="['page-btn', { active: p === currentPage }]"
          @click="handlePageChange(p)"
        >{{ p }}</button>
      </template>
      <button
        :disabled="currentPage >= totalPages"
        class="page-btn"
        @click="handlePageChange(currentPage + 1)"
      >下一页</button>
    </div>
  </div>
</template>

<style scoped>
.western-medicine { max-width: 100%; }

.toolbar { display: flex; flex-direction: column; gap: 16px; }

.search-wrap {
  position: relative; width: 100%; max-width: 400px;
}
.search-input {
  width: 100%; padding: 10px 14px 10px 38px;
  background: var(--bg-input); border: 1px solid var(--border-color);
  border-radius: 8px; color: var(--text-primary); font-size: 14px;
  outline: none; box-sizing: border-box;
}
.search-input:focus { border-color: rgba(22, 93, 255, 0.6); }
.search-input::placeholder { color: var(--text-tertiary); }
.search-icon {
  position: absolute; left: 12px; top: 50%;
  transform: translateY(-50%); color: var(--text-tertiary);
}

.category-tabs {
  display: flex; gap: 8px; flex-wrap: wrap;
}
.cat-tab {
  padding: 6px 16px; border-radius: 20px; border: 1px solid var(--border-color);
  background: transparent; color: var(--text-secondary); font-size: 13px;
  cursor: pointer; transition: all 0.2s;
}
.cat-tab:hover { border-color: rgba(22, 93, 255, 0.4); color: var(--text-primary); }
.cat-tab.active {
  background: rgba(22, 93, 255, 0.1); border-color: rgba(22, 93, 255, 0.4);
  color: #165DFF;
}

.medicine-list {
  display: grid; grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 16px; margin-top: 20px;
}
.medicine-card {
  background: var(--bg-card); border: 1px solid var(--border-color);
  border-radius: 12px; transition: all 0.2s; overflow: hidden;
}
.medicine-card:hover {
  border-color: rgba(22, 93, 255, 0.3);
  box-shadow: 0 4px 20px rgba(0,0,0,0.08);
  transform: translateY(-2px);
}
.card-body { padding: 20px; }

.card-top {
  display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px;
}
.drug-icon { color: #36CFC9; }
.category-tag {
  padding: 2px 10px; border-radius: 4px;
  background: rgba(22, 93, 255, 0.1); color: #165DFF;
  font-size: 12px; font-weight: 500;
}

.drug-name {
  font-size: 16px; font-weight: 600; color: var(--text-primary);
  margin: 0 0 4px 0;
}
.drug-en-name {
  font-size: 12px; color: var(--text-tertiary);
  margin: 0 0 14px 0;
}

.drug-details { margin-bottom: 14px; }
.detail-row {
  display: flex; justify-content: space-between; padding: 4px 0;
  font-size: 13px;
}
.detail-label { color: var(--text-tertiary); }
.detail-value { color: var(--text-primary); }

.card-footer {
  padding-top: 12px; border-top: 1px solid var(--border-color);
}
.price-stock {
  display: flex; justify-content: space-between; align-items: center;
}
.price { font-size: 18px; font-weight: 700; color: #F77234; }
.stock-badge {
  padding: 2px 8px; border-radius: 4px; font-size: 12px; font-weight: 500;
}

.pagination {
  display: flex; justify-content: center; gap: 6px; margin-top: 24px;
  flex-wrap: wrap;
}
.page-btn {
  padding: 6px 14px; border-radius: 6px;
  border: 1px solid var(--border-color); background: var(--bg-card);
  color: var(--text-secondary); font-size: 13px; cursor: pointer;
  transition: all 0.2s;
}
.page-btn:hover:not(:disabled):not(.active) {
  border-color: rgba(22, 93, 255, 0.4); color: var(--text-primary);
}
.page-btn.active {
  background: rgba(22, 93, 255, 0.1); border-color: rgba(22, 93, 255, 0.4);
  color: #165DFF;
}
.page-btn:disabled { opacity: 0.4; cursor: not-allowed; }

.empty-state {
  grid-column: 1 / -1; display: flex; flex-direction: column;
  align-items: center; gap: 12px; padding: 60px 20px;
  color: var(--text-tertiary);
}
.empty-state p { margin: 0; font-size: 14px; }
</style>
