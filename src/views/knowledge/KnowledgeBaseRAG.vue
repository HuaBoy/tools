<script setup>
import { ref, onMounted, computed } from 'vue';
import { ElMessage, ElLoading } from 'element-plus';
import GlassCard from '@/components/GlassCard.vue';
import { useLogsStore } from '@/stores/logs';
import knowledgeManager from '@/modules/rag/knowledgeManager.js';

const logsStore = useLogsStore();
const searchQuery = ref('');
const showAddForm = ref(false);
const showImportForm = ref(false);
const searchResults = ref([]);
const isSearching = ref(false);
const knowledgeStats = ref(null);
const isInitializing = ref(false);

// 分类选项
const categoryOptions = [
  { code: 'hardware', name: '硬件', color: '#F53F3F' },
  { code: 'firmware', name: '固件', color: '#FF7D00' },
  { code: 'network', name: '组网', color: '#165DFF' },
  { code: 'auth', name: '授权', color: '#722ED1' },
  { code: 'log', name: '日志', color: '#00B42A' },
  { code: 'document', name: '文档', color: '#0FC6C2' },
  { code: 'other', name: '其他', color: '#64748B' }
];

// 导入文件相关
const importFile = ref(null);
const importTags = ref('document');
const importCategory = ref('document');

// 新增知识表单
const newKnowledge = ref({
  question: '',
  answer: '',
  tags: 'other',
  category: 'other'
});

// 初始化知识库
const initializeKnowledgeBase = async () => {
  isInitializing.value = true;
  try {
    await knowledgeManager.initialize();
    
    // 如果没有数据，导入示例数据
    const stats = knowledgeManager.getStats();
    if (stats.totalItems === 0) {
      await importSampleData();
    }
    
    knowledgeStats.value = stats;
    isInitializing.value = false;
    ElMessage.success('知识库初始化完成');
  } catch (error) {
    console.error('知识库初始化失败:', error);
    isInitializing.value = false;
    ElMessage.error('知识库初始化失败，请刷新页面重试');
  }
};

// 导入示例数据（包含AI问题数据库的所有数据）
const importSampleData = async () => {
  const sampleData = [
    {
      question: '设备与控制中心通信连接超时',
      answer: '1. 检查网络信号强度，确保信号强度大于-70dBm\n2. 重启通信模块，等待30秒后重试\n3. 检查天线连接是否牢固\n4. 尝试更换通信模块测试',
      tags: ['network'],
      category: 'network'
    },
    {
      question: '起爆参数配置超出允许范围',
      answer: '1. 检查参数配置是否符合规范要求\n2. 重新校准参数值到合理范围\n3. 使用默认参数模板恢复配置\n4. 联系技术支持获取参数配置指导',
      tags: ['firmware'],
      category: 'firmware'
    },
    {
      question: '设备电源电压低于正常工作范围',
      answer: '1. 检查电池电量是否充足\n2. 更换备用电池测试\n3. 检查充电器输出是否正常\n4. 检查电源线路是否有损坏',
      tags: ['hardware'],
      category: 'hardware'
    },
    {
      question: '起爆指令发送后未成功执行',
      answer: '1. 检查雷管连接状态是否正常\n2. 重新发送起爆指令\n3. 检查起爆网络完整性\n4. 更换故障雷管',
      tags: ['network'],
      category: 'network'
    },
    {
      question: '设备数据与云端同步失败',
      answer: '1. 检查网络连接是否正常\n2. 重新发起数据同步\n3. 检查数据格式是否正确\n4. 联系技术支持协助排查',
      tags: ['network'],
      category: 'network'
    },
    {
      question: '参数警告：起爆参数接近临界值',
      answer: '1. 检查参数值是否合理\n2. 根据实际情况调整参数\n3. 确认是否需要继续执行\n4. 参考操作手册参数建议',
      tags: ['firmware'],
      category: 'firmware'
    },
    {
      question: '授权码验证失败',
      answer: '1. 检查授权码输入是否正确\n2. 确认授权码是否过期\n3. 联系管理员获取新授权码\n4. 检查网络连接是否正常',
      tags: ['auth'],
      category: 'auth'
    },
    {
      question: '日志解析异常',
      answer: '1. 检查日志文件格式是否正确\n2. 确保日志文件未损坏\n3. 使用日志解密工具处理加密日志\n4. 联系技术支持协助分析',
      tags: ['log'],
      category: 'log'
    },
    {
      question: '接口连接问题',
      answer: '1. 检查接口线缆连接是否牢固\n2. 更换接口线缆测试\n3. 检查接口是否有损坏\n4. 重启设备后重试',
      tags: ['hardware'],
      category: 'hardware'
    },
    {
      question: '程序升级失败',
      answer: '1. 检查网络连接是否稳定\n2. 确认固件文件完整性\n3. 确保设备电量充足\n4. 重新发起升级流程',
      tags: ['firmware'],
      category: 'firmware'
    }
  ];

  try {
    await knowledgeManager.batchAddKnowledgeItems(sampleData);
    ElMessage.success('AI问题数据库数据已成功导入到RAG知识库');
  } catch (error) {
    console.error('示例数据导入失败:', error);
  }
};

// 智能搜索
const handleSearch = async () => {
  if (!searchQuery.value.trim()) {
    searchResults.value = [];
    return;
  }

  isSearching.value = true;
  searchResults.value = [];

  try {
    const results = await knowledgeManager.searchKnowledge(searchQuery.value, {
      topK: 10,
      minScore: 0.3,
      includeAnswer: true
    });

    searchResults.value = results;
    logsStore.addLog('查询', 'AI问题数据库', `智能搜索: ${searchQuery.value}`);
    
    if (results.length === 0) {
      ElMessage.info('未找到相关问题，您可以添加新的知识条目');
    }
  } catch (error) {
    console.error('搜索失败:', error);
    ElMessage.error('搜索失败，请稍后重试');
  } finally {
    isSearching.value = false;
  }
};

// 新增知识
const addKnowledge = async () => {
  if (!newKnowledge.value.question.trim() || !newKnowledge.value.answer.trim()) {
    ElMessage.warning('请填写完整的问题描述和解决方案');
    return;
  }

  try {
    const knowledgeItem = await knowledgeManager.addKnowledgeItem({
      question: newKnowledge.value.question.trim(),
      answer: newKnowledge.value.answer.trim(),
      tags: [newKnowledge.value.tags],
      category: newKnowledge.value.tags
    });

    // 更新统计信息
    knowledgeStats.value = knowledgeManager.getStats();
    
    // 清空表单
    newKnowledge.value = { question: '', answer: '', tags: 'other', category: 'other' };
    showAddForm.value = false;
    
    ElMessage.success('知识条目添加成功');
    logsStore.addLog('新增', 'AI问题数据库', `新增知识: ${knowledgeItem.question}`);
  } catch (error) {
    console.error('添加知识失败:', error);
    ElMessage.error('添加失败，请稍后重试');
  }
};

// 导入文档
const handleImport = async () => {
  if (!importFile.value) {
    ElMessage.warning('请选择要导入的文件');
    return;
  }

  const loading = ElLoading.service({
    lock: true,
    text: '正在处理文档...',
    background: 'rgba(0, 0, 0, 0.7)'
  });

  try {
    const fileName = importFile.value?.name || '未知文件';
    const results = await knowledgeManager.importDocument(importFile.value, {
      tags: [importTags.value],
      category: importCategory.value
    });

    // 更新统计信息
    knowledgeStats.value = knowledgeManager.getStats();
    
    // 清空表单
    importFile.value = null;
    showImportForm.value = false;
    
    ElMessage.success(`成功导入 ${results.length} 个文档片段`);
    logsStore.addLog('导入', '智能知识库(RAG)', `导入文档: ${fileName}`);
  } catch (error) {
    console.error('文档导入失败:', error);
    ElMessage.error(`导入失败: ${error.message}`);
  } finally {
    loading.close();
  }
};

// 文件选择处理
const fileInput = ref(null);
const openFilePicker = () => {
  fileInput.value?.click();
};
const handleFileSelect = (event) => {
  importFile.value = event.target.files[0];
};

// 拖拽文件处理
const isDragging = ref(false);
const handleFileDrop = (event) => {
  isDragging.value = false;
  const file = event.dataTransfer?.files?.[0];
  if (file) {
    importFile.value = file;
  }
};

// 导出知识库
const exportKnowledge = () => {
  try {
    const exportData = knowledgeManager.exportKnowledge('json');
    const blob = new Blob([exportData], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `knowledge_base_${new Date().toISOString().split('T')[0]}.json`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
    
    ElMessage.success('知识库导出成功');
    logsStore.addLog('导出', 'AI问题数据库', '导出知识库');
  } catch (error) {
    console.error('导出失败:', error);
    ElMessage.error('导出失败，请稍后重试');
  }
};

// 清空知识库
const clearKnowledgeBase = async () => {
  try {
    await knowledgeManager.clear();
    knowledgeStats.value = knowledgeManager.getStats();
    searchResults.value = [];
    searchQuery.value = '';
    
    ElMessage.success('知识库已清空');
    logsStore.addLog('清空', 'AI问题数据库', '清空知识库');
  } catch (error) {
    console.error('清空失败:', error);
    ElMessage.error('清空失败，请稍后重试');
  }
};

// 获取分类信息
const getCategoryInfo = (code) => {
  return categoryOptions.find(c => c.code === code) || categoryOptions[6];
};

// 格式化相似度
const formatSimilarity = (score) => {
  return (score * 100).toFixed(1) + '%';
};

// 计算相似度颜色
const getSimilarityColor = (score) => {
  if (score >= 0.8) return '#00B42A';
  if (score >= 0.6) return '#FF7D00';
  if (score >= 0.4) return '#F53F3F';
  return '#64748B';
};

// 计算显示的知识条目
const displayedKnowledge = computed(() => {
  if (searchQuery.value.trim() && searchResults.value.length > 0) {
    return searchResults.value;
  }
  return [];
});

// 页面加载时初始化
onMounted(() => {
  initializeKnowledgeBase();
});
</script>

<template>
  <div class="knowledge-base">
    <!-- 统计信息卡片 -->
    <GlassCard title="智能知识库(RAG)" class="stats-card" v-if="knowledgeStats">
      <div class="stats-grid">
        <div class="stat-item">
          <div class="stat-label">知识条目</div>
          <div class="stat-value">{{ knowledgeStats.totalItems }}</div>
        </div>
        <div class="stat-item">
          <div class="stat-label">分类</div>
          <div class="stat-value">{{ knowledgeStats.categories?.length || 0 }}</div>
        </div>
        <div class="stat-item">
          <div class="stat-label">标签</div>
          <div class="stat-value">{{ knowledgeStats.tags?.length || 0 }}</div>
        </div>
        <div class="stat-item">
          <div class="stat-label">向量数量</div>
          <div class="stat-value">{{ knowledgeStats.vectorStats?.totalVectors || 0 }}</div>
        </div>
      </div>
    </GlassCard>

    <!-- 搜索区域 -->
    <GlassCard title="智能搜索" class="search-card">
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
            :disabled="isInitializing"
          />
          <button class="search-btn" @click="handleSearch" :disabled="isSearching || isInitializing">
            <svg v-if="!isSearching" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <polygon points="5 3 19 12 5 21 5 3" />
            </svg>
            <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="spin">
              <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
            </svg>
          </button>
        </div>
        
        <div v-if="searchResults.length > 0" class="search-stats">
          找到 {{ searchResults.length }} 条相关结果，按相似度排序
        </div>
      </div>

      <!-- 操作按钮组 -->
      <div class="action-buttons">
        <button class="action-btn" @click="showAddForm = !showAddForm" :disabled="isInitializing">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <line x1="12" y1="5" x2="12" y2="19" />
            <line x1="5" y1="12" x2="19" y2="12" />
          </svg>
          新增知识
        </button>
        <button class="action-btn" @click="showImportForm = !showImportForm" :disabled="isInitializing">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
            <polyline points="7 10 12 15 17 10" />
            <line x1="12" y1="15" x2="12" y2="3" />
          </svg>
          导入文档
        </button>
        <button class="action-btn" @click="exportKnowledge" :disabled="isInitializing || knowledgeStats?.totalItems === 0">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
            <polyline points="17 8 12 3 7 8" />
            <line x1="12" y1="3" x2="12" y2="15" />
          </svg>
          导出知识库
        </button>
        <button class="action-btn danger" @click="clearKnowledgeBase" :disabled="isInitializing || knowledgeStats?.totalItems === 0">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="3 6 5 6 21 6" />
            <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
          </svg>
          清空知识库
        </button>
      </div>
    </GlassCard>

    <!-- 新增知识表单 -->
    <GlassCard v-if="showAddForm" title="新增知识条目" class="form-card">
      <div class="form-content">
        <div class="form-group">
          <label>问题描述 *</label>
          <textarea 
            v-model="newKnowledge.question"
            class="form-textarea"
            placeholder="请描述问题..."
            rows="3"
          ></textarea>
        </div>
        <div class="form-group">
          <label>分类标签 *</label>
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
          <label>解决方案 *</label>
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
              <path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z" />
              <polyline points="17 21 17 13 7 13 7 21" />
              <polyline points="7 3 7 8 15 8" />
            </svg>
            保存入库
          </button>
          <button class="cancel-btn" @click="showAddForm = false">取消</button>
        </div>
      </div>
    </GlassCard>

    <!-- 导入文档表单 -->
    <GlassCard v-if="showImportForm" title="导入文档" class="form-card">
      <div class="form-content">
        <div class="form-group">
          <label>选择文件</label>
          <div class="file-upload" :class="{ 'dragging': isDragging }"
               @click="openFilePicker"
               @dragover.prevent="isDragging = true"
               @dragenter.prevent="isDragging = true"
               @dragleave.prevent="isDragging = false"
               @drop.prevent="handleFileDrop">
            <input ref="fileInput" type="file" @change="handleFileSelect" accept=".txt,.md,.json,.pdf,.docx,.xlsx,.xls,.pptx,.ppt" style="display:none" />
            <div class="file-info" v-if="importFile">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z" />
                <polyline points="13 2 13 9 20 9" />
              </svg>
              <span>{{ importFile.name }} ({{ (importFile.size / 1024).toFixed(2) }} KB)</span>
            </div>
            <div class="file-placeholder" v-else>
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
                <polyline points="17 8 12 3 7 8" />
                <line x1="12" y1="3" x2="12" y2="15" />
              </svg>
              <span>点击选择文件或拖拽到此处</span>
              <p class="file-hint">支持格式: TXT, MD, JSON, PDF, DOCX, XLSX, PPTX</p>
            </div>
          </div>
        </div>
        <div class="form-group">
          <label>分类标签</label>
          <select v-model="importTags" class="form-select">
            <option v-for="cat in categoryOptions" :key="cat.code" :value="cat.code">
              {{ cat.name }}
            </option>
          </select>
        </div>
        <div class="form-group">
          <label>分类</label>
          <select v-model="importCategory" class="form-select">
            <option v-for="cat in categoryOptions" :key="cat.code" :value="cat.code">
              {{ cat.name }}
            </option>
          </select>
        </div>
        <div class="form-actions">
          <button class="submit-btn" @click="handleImport" :disabled="!importFile">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
              <polyline points="7 10 12 15 17 10" />
              <line x1="12" y1="15" x2="12" y2="3" />
            </svg>
            开始导入
          </button>
          <button class="cancel-btn" @click="showImportForm = false">取消</button>
        </div>
      </div>
    </GlassCard>

    <!-- 搜索结果 -->
    <div v-if="displayedKnowledge.length > 0" class="search-results">
      <GlassCard 
        v-for="(item, index) in displayedKnowledge" 
        :key="item.id"
        :class="{ 'highlight-result': index === 0 }"
      >
        <div class="knowledge-item">
          <div class="item-header">
            <div class="item-tags">
              <span 
                class="category-badge"
                :style="{ background: getCategoryInfo(item.category).color + '20', color: getCategoryInfo(item.category).color }"
              >
                {{ getCategoryInfo(item.category).name }}
              </span>
              <span 
                v-if="item.similarity" 
                class="similarity-badge"
                :style="{ background: getSimilarityColor(item.similarity) + '20', color: getSimilarityColor(item.similarity) }"
              >
                相似度: {{ formatSimilarity(item.similarity) }}
              </span>
              <span class="source-badge" v-if="item.source">
                {{ item.source === 'manual' ? '手动添加' : '文档导入' }}
              </span>
            </div>
            <span class="create-time">{{ item.metadata?.createdAt?.split('T')[0] }}</span>
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
    </div>

    <!-- 空状态 -->
    <div v-if="!searchQuery.trim() && knowledgeStats?.totalItems === 0" class="empty-state">
      <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
        <circle cx="12" cy="12" r="10" />
        <path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3" />
        <line x1="12" y1="17" x2="12.01" y2="17" />
      </svg>
      <p>知识库暂无数据</p>
      <div class="empty-actions">
        <button class="add-knowledge-btn" @click="showAddForm = true">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <line x1="12" y1="5" x2="12" y2="19" />
            <line x1="5" y1="12" x2="19" y2="12" />
          </svg>
          添加第一条知识
        </button>
        <button class="import-knowledge-btn" @click="showImportForm = true">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
            <polyline points="7 10 12 15 17 10" />
            <line x1="12" y1="15" x2="12" y2="3" />
          </svg>
          导入文档
        </button>
      </div>
    </div>

    <!-- 搜索结果为空 -->
    <div v-if="searchQuery.trim() && !isSearching && displayedKnowledge.length === 0" class="no-results">
      <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
        <circle cx="12" cy="12" r="10" />
        <path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3" />
        <line x1="12" y1="17" x2="12.01" y2="17" />
      </svg>
      <p>未检索到相似问题</p>
      <div class="no-results-actions">
        <button class="add-knowledge-btn" @click="showAddForm = true">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <line x1="12" y1="5" x2="12" y2="19" />
            <line x1="5" y1="12" x2="19" y2="12" />
          </svg>
          将问题入库
        </button>
      </div>
    </div>

    <!-- 初始化状态 -->
    <div v-if="isInitializing" class="loading-state">
      <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" class="spin">
        <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
      </svg>
      <p>正在初始化知识库...</p>
    </div>
  </div>
</template>

<style scoped>
.knowledge-base {
  max-width: 100%;
}

/* 统计卡片 */
.stats-card {
  margin-bottom: 16px;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 16px;
  padding: 16px;
}

.stat-item {
  text-align: center;
  padding: 12px;
  background: var(--bg-input);
  border-radius: 8px;
}

.stat-label {
  font-size: 12px;
  color: var(--text-tertiary);
  margin-bottom: 4px;
}

.stat-value {
  font-size: 24px;
  font-weight: 600;
  color: var(--text-primary);
}

/* 搜索卡片 */
.search-card {
  margin-bottom: 16px;
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
  
  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
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

/* 操作按钮组 */
.action-buttons {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
  padding: 0 16px 16px;
}

.action-btn {
  display: flex;
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
  
  &:hover:not(:disabled) {
    background: rgba(22, 93, 255, 0.2);
  }
  
  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
  
  &.danger {
    background: rgba(245, 63, 63, 0.1);
    border-color: rgba(245, 63, 63, 0.3);
    color: #F53F3F;
    
    &:hover:not(:disabled) {
      background: rgba(245, 63, 63, 0.2);
    }
  }
}

/* 表单卡片 */
.form-card {
  margin-bottom: 16px;
}

.form-content {
  padding: 20px;
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

.form-select {
  width: 100%;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  padding: 10px 12px;
  color: var(--text-primary);
  font-size: 13px;
  outline: none;
  
  &:focus {
    border-color: #165DFF;
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

.file-upload {
  border: 2px dashed var(--border-color);
  border-radius: 8px;
  padding: 20px;
  text-align: center;
  cursor: pointer;
  transition: border-color 0.2s;
  position: relative;
  
  &:hover {
    border-color: #165DFF;
  }

  &.dragging {
    border-color: #165DFF;
    background: rgba(22, 93, 255, 0.08);
  }
  
  input[type="file"] {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    opacity: 0;
    cursor: pointer;
  }
}

.file-info {
  display: flex;
  align-items: center;
  gap: 8px;
  justify-content: center;
  
  svg {
    color: #165DFF;
  }
  
  span {
    color: var(--text-primary);
    font-size: 13px;
  }
}

.file-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  
  svg {
    color: var(--text-tertiary);
  }
  
  span {
    color: var(--text-secondary);
    font-size: 14px;
  }
}

.file-hint {
  font-size: 12px;
  color: var(--text-tertiary);
  margin-top: 4px;
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
  
  &:hover:not(:disabled) {
    transform: translateY(-1px);
  }
  
  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
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

/* 搜索结果 */
.search-results {
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
  flex-wrap: wrap;
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
  font-weight: 500;
}

.source-badge {
  font-size: 11px;
  padding: 3px 8px;
  border-radius: 4px;
  background: rgba(100, 116, 139, 0.1);
  color: #64748B;
  font-weight: 500;
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

/* 空状态 */
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

.empty-actions {
  display: flex;
  gap: 12px;
  justify-content: center;
}

.add-knowledge-btn,
.import-knowledge-btn {
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

.import-knowledge-btn {
  background: rgba(0, 180, 42, 0.1);
  border-color: rgba(0, 180, 42, 0.3);
  color: #00B42A;
  
  &:hover {
    background: rgba(0, 180, 42, 0.2);
  }
}

/* 无结果状态 */
.no-results {
  text-align: center;
  padding: 32px;
  background: var(--bg-input);
  border-radius: 8px;
  margin-top: 16px;
  
  svg {
    color: var(--text-tertiary);
    margin-bottom: 12px;
  }
  
  p {
    color: var(--text-secondary);
    margin: 0 0 16px 0;
  }
}

.no-results-actions {
  display: flex;
  justify-content: center;
}

/* 加载状态 */
.loading-state {
  text-align: center;
  padding: 60px;
  
  svg {
    color: var(--text-tertiary);
    margin-bottom: 16px;
  }
  
  p {
    color: var(--text-tertiary);
    margin: 0;
  }
}

/* 高亮结果 */
.highlight-result {
  border: 2px solid rgba(22, 93, 255, 0.3);
}

/* 旋转动画 */
.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

/* 响应式设计 */
@media screen and (max-width: 768px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: 12px;
    padding: 12px;
  }
  
  .stat-item {
    padding: 8px;
  }
  
  .stat-value {
    font-size: 20px;
  }
  
  .search-section {
    padding: 12px;
  }
  
  .search-box {
    padding: 10px 12px;
  }
  
  .search-input {
    font-size: 13px;
  }
  
  .action-buttons {
    flex-direction: column;
    padding: 0 12px 12px;
  }
  
  .action-btn {
    width: 100%;
    justify-content: center;
  }
  
  .form-content {
    padding: 12px;
  }
  
  .category-selector {
    gap: 8px;
  }
  
  .category-option {
    font-size: 12px;
  }
  
  .item-header {
    flex-direction: column;
    gap: 8px;
    align-items: flex-start;
  }
  
  .empty-state {
    padding: 40px 20px;
  }
  
  .empty-actions {
    flex-direction: column;
  }
}

@media screen and (max-width: 480px) {
  .stats-grid {
    grid-template-columns: 1fr;
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