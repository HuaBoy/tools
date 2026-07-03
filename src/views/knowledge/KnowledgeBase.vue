<script setup>import { ref, computed } from 'vue';
import { ElMessage } from 'element-plus';
import GlassCard from '@/components/GlassCard.vue';
import { useLogsStore } from '@/stores/logs';

const logsStore = useLogsStore();
const searchQuery = ref('');
const showAddForm = ref(false);
const searchResults = ref([]);
const isSearching = ref(false);

const categoryOptions = [
  { code: 'hardware', name: '硬件', color: '#F53F3F' },
  { code: 'firmware', name: '固件', color: '#FF7D00' },
  { code: 'network', name: '组网', color: '#165DFF' },
  { code: 'auth', name: '授权', color: '#722ED1' },
  { code: 'log', name: '日志', color: '#00B42A' },
  { code: 'other', name: '其他', color: '#64748B' }
];

const newKnowledge = ref({
  question: '',
  answer: '',
  tags: 'other'
});

const knowledgeBase = ref([
  {
    id: 1,
    tags: 'network',
    question: '设备与控制中心通信连接超时',
    answer: '1. 检查网络信号强度，确保信号强度大于-70dBm\n2. 重启通信模块，等待30秒后重试\n3. 检查天线连接是否牢固\n4. 尝试更换通信模块测试',
    create_time: '2024-01-15 10:30:00'
  },
  {
    id: 2,
    tags: 'firmware',
    question: '起爆参数配置超出允许范围',
    answer: '1. 检查参数配置是否符合规范要求\n2. 重新校准参数值到合理范围\n3. 使用默认参数模板恢复配置\n4. 联系技术支持获取参数配置指导',
    create_time: '2024-01-16 14:20:00'
  },
  {
    id: 3,
    tags: 'hardware',
    question: '设备电源电压低于正常工作范围',
    answer: '1. 检查电池电量是否充足\n2. 更换备用电池测试\n3. 检查充电器输出是否正常\n4. 检查电源线路是否有损坏',
    create_time: '2024-01-17 09:15:00'
  },
  {
    id: 4,
    tags: 'network',
    question: '起爆指令发送后未成功执行',
    answer: '1. 检查雷管连接状态是否正常\n2. 重新发送起爆指令\n3. 检查起爆网络完整性\n4. 更换故障雷管',
    create_time: '2024-01-18 16:45:00'
  },
  {
    id: 5,
    tags: 'network',
    question: '设备数据与云端同步失败',
    answer: '1. 检查网络连接是否正常\n2. 重新发起数据同步\n3. 检查数据格式是否正确\n4. 联系技术支持协助排查',
    create_time: '2024-01-19 11:00:00'
  },
  {
    id: 6,
    tags: 'firmware',
    question: '参数警告：起爆参数接近临界值',
    answer: '1. 检查参数值是否合理\n2. 根据实际情况调整参数\n3. 确认是否需要继续执行\n4. 参考操作手册参数建议',
    create_time: '2024-01-20 13:30:00'
  },
  {
    id: 7,
    tags: 'auth',
    question: '授权码验证失败',
    answer: '1. 检查授权码输入是否正确\n2. 确认授权码是否过期\n3. 联系管理员获取新授权码\n4. 检查网络连接是否正常',
    create_time: '2024-01-21 15:00:00'
  },
  {
    id: 8,
    tags: 'log',
    question: '日志解析异常',
    answer: '1. 检查日志文件格式是否正确\n2. 确保日志文件未损坏\n3. 使用日志解密工具处理加密日志\n4. 联系技术支持协助分析',
    create_time: '2024-01-22 10:15:00'
  },
  {
    id: 9,
    tags: 'hardware',
    question: '接口连接问题',
    answer: '1. 检查接口线缆连接是否牢固\n2. 更换接口线缆测试\n3. 检查接口是否有损坏\n4. 重启设备后重试',
    create_time: '2024-01-23 09:45:00'
  },
  {
    id: 10,
    tags: 'firmware',
    question: '程序升级失败',
    answer: '1. 检查网络连接是否稳定\n2. 确认固件文件完整性\n3. 确保设备电量充足\n4. 重新发起升级流程',
    create_time: '2024-01-24 14:30:00'
  }
]);

const calculateSimilarity = (text1, text2) => {
  const words1 = text1.toLowerCase().split(/\s+/).filter(w => w.length > 1);
  const words2 = text2.toLowerCase().split(/\s+/).filter(w => w.length > 1);
  
  if (words1.length === 0 || words2.length === 0) return 0;
  
  const intersection = words1.filter(w => words2.includes(w));
  const union = [...new Set([...words1, ...words2])];
  
  return intersection.length / union.length;
};

const handleSearch = async () => {
  if (!searchQuery.value.trim()) {
    searchResults.value = [];
    return;
  }
  
  isSearching.value = true;
  
  await new Promise(resolve => setTimeout(resolve, 500));
  
  const results = knowledgeBase.value.map(item => ({
    ...item,
    similarity: calculateSimilarity(searchQuery.value, item.question + ' ' + item.answer)
  }))
  .filter(item => item.similarity >= 0.3)
  .sort((a, b) => b.similarity - a.similarity)
  .slice(0, 5);
  
  searchResults.value = results;
  isSearching.value = false;
  logsStore.addLog('查询', 'AI问题数据库', `搜索: ${searchQuery.value}`);
};

const addKnowledge = () => {
  if (!newKnowledge.value.question.trim() || !newKnowledge.value.answer.trim()) {
    ElMessage.warning('请填写完整的问题描述和解决方案');
    return;
  }
  
  const newItem = {
    id: Date.now(),
    tags: newKnowledge.value.tags,
    question: newKnowledge.value.question.trim(),
    answer: newKnowledge.value.answer.trim(),
    create_time: new Date().toLocaleString('zh-CN')
  };
  
  knowledgeBase.value.unshift(newItem);
  newKnowledge.value = { question: '', answer: '', tags: 'other' };
  showAddForm.value = false;
  ElMessage.success('知识条目添加成功');
  logsStore.addLog('新增', 'AI问题数据库', `新增知识: ${newItem.question}`);
};

const getCategoryInfo = (code) => {
  return categoryOptions.find(c => c.code === code) || categoryOptions[5];
};

const formatSimilarity = (score) => {
  return (score * 100).toFixed(1) + '%';
};

const filteredKnowledge = computed(() => {
  if (!searchQuery.value.trim()) {
    return knowledgeBase.value;
  }
  return searchResults.value;
});

const hasSearchResults = computed(() => {
  return searchQuery.value.trim() && searchResults.value.length > 0;
});

const noSearchResults = computed(() => {
  return searchQuery.value.trim() && !isSearching.value && searchResults.value.length === 0;
});
</script>

<template>
  <div class="knowledge-base">
    <GlassCard title="AI问题数据库">
      <div class="search-section">
        <div class="search-box">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="11" cy="11" r="8" />
            <line x1="21" y1="21" x2="16.65" y2="16.65" />
          </svg>
          <input 
            v-model="searchQuery"
            type="text" 
            class="search-input"
            placeholder="输入问题描述，智能匹配知识库..."
            @keyup.enter="handleSearch"
          />
          <button class="search-btn" @click="handleSearch" :disabled="isSearching">
            <svg v-if="!isSearching" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <polygon points="5 3 19 12 5 21 5 3" />
            </svg>
            <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="spin">
              <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
            </svg>
          </button>
        </div>
        
        <div v-if="hasSearchResults" class="search-stats">
          找到 {{ searchResults.length }} 条相关结果，按相似度排序
        </div>
      </div>

      <div v-if="noSearchResults" class="no-results">
        <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="12" cy="12" r="10" />
          <path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3" />
          <line x1="12" y1="17" x2="12.01" y2="17" />
        </svg>
        <p>未检索到相似问题</p>
        <button class="add-knowledge-btn" @click="showAddForm = true">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <line x1="12" y1="5" x2="12" y2="19" />
            <line x1="5" y1="12" x2="19" y2="12" />
          </svg>
          <span>将问题入库</span>
        </button>
      </div>

      <div v-if="showAddForm" class="add-form">
        <div class="form-header">
          <span class="form-title">新增知识条目</span>
          <button class="close-btn" @click="showAddForm = false">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <line x1="18" y1="6" x2="6" y2="18" />
              <line x1="6" y1="6" x2="18" y2="18" />
            </svg>
          </button>
        </div>
        <div class="form-group">
          <label>问题描述</label>
          <textarea 
            v-model="newKnowledge.question"
            class="form-textarea"
            placeholder="请描述问题..."
            rows="3"
          ></textarea>
        </div>
        <div class="form-group">
          <label>分类标签</label>
          <div class="category-selector">
            <label 
              v-for="cat in categoryOptions" 
              :key="cat.code"
              class="category-option"
            >
              <input type="radio" v-model="newKnowledge.tags" :value="cat.code" />
              <span class="category-dot" :style="{ background: cat.color }"></span>
              <span class="category-name">{{ cat.name }}</span>
            </label>
          </div>
        </div>
        <div class="form-group">
          <label>解决方案</label>
          <textarea 
            v-model="newKnowledge.answer"
            class="form-textarea"
            placeholder="请提供详细的解决方案..."
            rows="4"
          ></textarea>
        </div>
        <div class="form-actions">
          <button class="submit-btn" @click="addKnowledge">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z" />
            </svg>
            <span>保存入库</span>
          </button>
          <button class="cancel-btn" @click="showAddForm = false">取消</button>
        </div>
      </div>

      <div class="category-filter">
        <span class="filter-label">分类筛选:</span>
        <div class="category-tags">
          <button 
            v-for="cat in categoryOptions" 
            :key="cat.code"
            class="category-tag"
            :style="{ borderColor: cat.color, color: cat.color }"
            @click="handleSearch"
          >
            <span class="tag-dot" :style="{ background: cat.color }"></span>
            {{ cat.name }}
          </button>
        </div>
      </div>
    </GlassCard>
    
    <div class="knowledge-list">
      <GlassCard 
        v-for="(item, index) in filteredKnowledge" 
        :key="item.id" 
        :class="{ 'highlight-result': hasSearchResults && index === 0 }"
      >
        <div class="knowledge-item">
          <div class="item-header">
            <div class="item-tags">
              <span 
                class="category-badge"
                :style="{ background: getCategoryInfo(item.tags).color + '20', color: getCategoryInfo(item.tags).color }"
              >
                {{ getCategoryInfo(item.tags).name }}
              </span>
              <span v-if="item.similarity" class="similarity-badge">
                相似度: {{ formatSimilarity(item.similarity) }}
              </span>
            </div>
            <span class="create-time">{{ item.create_time }}</span>
          </div>
          
          <div class="item-question">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="12" r="10" />
              <path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3" />
              <line x1="12" y1="17" x2="12.01" y2="17" />
            </svg>
            <span>{{ item.question }}</span>
          </div>
          
          <div class="item-answer">
            <div class="answer-header">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />
              </svg>
              <span>解决方案</span>
            </div>
            <pre class="answer-content">{{ item.answer }}</pre>
          </div>
        </div>
      </GlassCard>
      
      <div v-if="!searchQuery.value && knowledgeBase.length === 0" class="empty-state">
        <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="12" cy="12" r="10" />
          <path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3" />
          <line x1="12" y1="17" x2="12.01" y2="17" />
        </svg>
        <p>知识库暂无数据</p>
        <button class="add-knowledge-btn" @click="showAddForm = true">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <line x1="12" y1="5" x2="12" y2="19" />
            <line x1="5" y1="12" x2="19" y2="12" />
          </svg>
          <span>添加第一条知识</span>
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.knowledge-base {
  max-width: 100%;
}

.search-section {
  padding: 16px;
  background: var(--bg-input);
  border-radius: 8px;
  margin-bottom: 16px;
}

.search-box {
  display: flex;
  align-items: center;
  gap: 10px;
  background: var(--bg-input);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 12px 16px;
  
  svg:first-child {
    color: var(--text-tertiary);
    flex-shrink: 0;
  }
}

.search-input {
  flex: 1;
  background: none;
  border: none;
  outline: none;
  color: var(--text-primary);
  font-size: 14px;
  
  &::placeholder {
    color: var(--text-tertiary);
  }
}

.search-btn {
  background: rgba(22, 93, 255, 0.2);
  border: 1px solid rgba(22, 93, 255, 0.3);
  border-radius: 6px;
  color: #165DFF;
  cursor: pointer;
  padding: 6px;
  transition: all 0.2s;
  
  &:hover:not(:disabled) {
    background: rgba(22, 93, 255, 0.3);
  }
  
  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
}

.search-stats {
  font-size: 12px;
  color: var(--text-tertiary);
  margin-top: 12px;
  text-align: right;
}

.no-results {
  text-align: center;
  padding: 32px;
  background: var(--bg-input);
  border-radius: 8px;
  margin-bottom: 16px;
  
  svg {
    color: var(--text-tertiary);
    margin-bottom: 12px;
  }
  
  p {
    color: var(--text-secondary);
    margin: 0 0 16px 0;
  }
}

.add-knowledge-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background: rgba(22, 93, 255, 0.1);
  border: 1px solid rgba(22, 93, 255, 0.3);
  border-radius: 6px;
  color: #165DFF;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
  
  &:hover {
    background: rgba(22, 93, 255, 0.2);
  }
}

.add-form {
  background: var(--bg-input);
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 16px;
}

.form-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.form-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.close-btn {
  background: none;
  border: none;
  color: var(--text-tertiary);
  cursor: pointer;
  padding: 4px;
  
  &:hover {
    color: var(--text-secondary);
  }
}

.form-group {
  margin-bottom: 16px;
  
  label {
    display: block;
    font-size: 12px;
    color: var(--text-tertiary);
    margin-bottom: 8px;
  }
}

.form-textarea {
  width: 100%;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  padding: 10px 12px;
  color: var(--text-primary);
  font-size: 13px;
  resize: vertical;
  outline: none;
  font-family: inherit;
  
  &:focus {
    border-color: #165DFF;
  }
  
  &::placeholder {
    color: var(--text-tertiary);
  }
}

.category-selector {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.category-option {
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
  font-size: 13px;
  color: var(--text-secondary);
  
  input[type="radio"] {
    display: none;
  }
  
  &:has(input[type="radio"]:checked) {
    .category-dot {
      transform: scale(1.2);
    }
    
    .category-name {
      font-weight: 500;
    }
  }
}

.category-dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  transition: transform 0.2s;
}

.category-name {
  transition: font-weight 0.2s;
}

.form-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
}

.submit-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background: linear-gradient(135deg, #165DFF 0%, #0F4CD0 100%);
  border: none;
  border-radius: 6px;
  color: #FFFFFF;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
  
  &:hover {
    transform: translateY(-1px);
  }
}

.cancel-btn {
  padding: 8px 16px;
  background: rgba(100, 116, 139, 0.2);
  border: 1px solid rgba(100, 116, 139, 0.3);
  border-radius: 6px;
  color: #94A3B8;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
  
  &:hover {
    background: rgba(100, 116, 139, 0.3);
  }
}

.category-filter {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.filter-label {
  font-size: 12px;
  color: var(--text-tertiary);
}

.category-tags {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.category-tag {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  background: transparent;
  border: 1px solid;
  border-radius: 4px;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s;
  
  &:hover {
    opacity: 0.8;
  }
}

.tag-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.knowledge-list {
  margin-top: 20px;
}

.knowledge-item {
  padding: 16px 0;
}

.item-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.item-tags {
  display: flex;
  gap: 8px;
}

.category-badge {
  font-size: 11px;
  padding: 3px 8px;
  border-radius: 4px;
  font-weight: 500;
}

.similarity-badge {
  font-size: 11px;
  padding: 3px 8px;
  border-radius: 4px;
  background: rgba(0, 180, 42, 0.1);
  color: #00B42A;
}

.create-time {
  font-size: 11px;
  color: var(--text-tertiary);
}

.item-question {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  margin-bottom: 16px;
  padding: 12px;
  background: var(--bg-input);
  border-radius: 6px;
  
  svg {
    color: #165DFF;
    flex-shrink: 0;
    margin-top: 2px;
  }
  
  span {
    font-size: 14px;
    color: var(--text-primary);
    line-height: 1.6;
  }
}

.item-answer {
  background: var(--bg-input);
  border-radius: 6px;
  overflow: hidden;
}

.answer-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  background: rgba(0, 180, 42, 0.1);
  border-bottom: 1px solid var(--border-color);
  
  svg {
    color: #00B42A;
    flex-shrink: 0;
  }
  
  span {
    font-size: 12px;
    color: #00B42A;
    font-weight: 500;
  }
}

.answer-content {
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 13px;
  color: var(--text-secondary);
  white-space: pre-wrap;
  word-break: break-all;
  margin: 0;
  line-height: 1.8;
  padding: 12px;
}

.empty-state {
  text-align: center;
  padding: 60px;
  color: var(--text-tertiary);
  
  svg {
    margin-bottom: 16px;
  }
  
  p {
    margin: 0 0 16px 0;
  }
}

.highlight-result {
  border: 2px solid rgba(22, 93, 255, 0.3);
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

@media screen and (max-width: 768px) {
  .search-section {
    padding: 12px;
  }
  
  .search-box {
    padding: 10px 12px;
  }
  
  .search-input {
    font-size: 13px;
  }
  
  .item-header {
    flex-direction: column;
    gap: 8px;
    align-items: flex-start;
  }
  
  .category-selector {
    gap: 8px;
  }
  
  .category-option {
    font-size: 12px;
  }
  
  .category-filter {
    flex-direction: column;
    align-items: flex-start;
  }
}

@media screen and (max-width: 480px) {
  .empty-state {
    padding: 40px 20px;
  }
  
  .add-form {
    padding: 12px;
  }
  
  .form-actions {
    flex-direction: column;
  }
  
  .submit-btn,
  .cancel-btn {
    width: 100%;
    justify-content: center;
  }
}
</style>
