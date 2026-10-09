<script setup>
import { ref, reactive, onMounted, computed } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import GlassCard from '@/components/GlassCard.vue';
import { useLogsStore } from '@/stores/logs';
import {
  getKnowledgeStats,
  listDocuments,
  uploadDocument,
  deleteDocument,
  getDocumentContent,
  updateDocumentContent,
  uploadImage,
  getFileUrl,
  searchKnowledge,
  listFAQs,
  createFAQ,
  deleteFAQ
} from '@/api/knowledge';

const logsStore = useLogsStore();

// ===== 顶部状态 =====
const activeTab = ref('docs'); // docs / faqs / search
const stats = ref(null);
const loading = ref(false);

// ===== 文档管理 =====
const docs = ref([]);
const docTotal = ref(0);
const docPage = ref(1);
const docPageSize = ref(12);
const docCategory = ref('all');
const uploading = ref(false);
const uploadProgress = ref(0);
const uploadDialog = ref(false);
const uploadForm = reactive({ title: '', category: 'other', description: '' });
const uploadFile = ref(null);

const categoryOptions = [
  { code: 'hardware', name: '硬件', color: '#F53F3F' },
  { code: 'firmware', name: '固件', color: '#FF7D00' },
  { code: 'network', name: '组网', color: '#165DFF' },
  { code: 'auth', name: '授权', color: '#722ED1' },
  { code: 'log', name: '日志', color: '#00B42A' },
  { code: 'document', name: '文档', color: '#0FC6C2' },
  { code: 'sales', name: '销售', color: '#F7BA1E' },
  { code: 'other', name: '其他', color: '#64748B' }
];

const getCategoryInfo = (code) => {
  return categoryOptions.find(c => c.code === code) || categoryOptions[7];
};

const loadDocs = async () => {
  loading.value = true;
  try {
    const res = await listDocuments({
      page: docPage.value,
      pageSize: docPageSize.value,
      category: docCategory.value
    });
    docs.value = res.data;
    docTotal.value = res.total;
  } catch (e) {
    ElMessage.error('加载文档失败：' + (e.message || ''));
  } finally {
    loading.value = false;
  }
};

const onDocPageChange = (p) => { docPage.value = p; loadDocs(); };
const onCategoryChange = () => { docPage.value = 1; loadDocs(); };

// 上传
const handleFileSelect = (e) => { uploadFile.value = e.target.files[0]; };
const fileInputRef = ref(null);
const openFilePicker = () => fileInputRef.value?.click();

const handleUpload = async () => {
  if (!uploadFile.value) {
    ElMessage.warning('请选择文件');
    return;
  }
  const allowed = ['.txt', '.md', '.markdown', '.json', '.csv', '.log', '.docx', '.pdf', '.xlsx', '.xls'];
  const ext = (uploadFile.value.name.match(/\.[^.]+$/) || [''])[0].toLowerCase();
  if (!allowed.includes(ext)) {
    ElMessage.error('不支持的格式：' + ext + '。支持：' + allowed.join(' '));
    return;
  }

  uploading.value = true;
  uploadProgress.value = 0;
  try {
    const doc = await uploadDocument(
      uploadFile.value,
      {
        title: uploadForm.title || uploadFile.value.name,
        category: uploadForm.category,
        description: uploadForm.description
      },
      (e) => {
        if (e.lengthComputable && e.total) {
          uploadProgress.value = Math.round((e.loaded / e.total) * 100);
        }
      }
    );
    ElMessage.success(`文档「${doc.title}」上传成功，生成 ${doc.chunk_count} 个知识分块`);
    logsStore.addLog('上传', '知识库', `上传文档: ${doc.title}`);
    uploadDialog.value = false;
    uploadFile.value = null;
    uploadForm.title = '';
    uploadForm.description = '';
    docPage.value = 1;
    await loadDocs();
    await loadStats();
  } catch (e) {
    ElMessage.error('上传失败：' + (e.message || ''));
  } finally {
    uploading.value = false;
    uploadProgress.value = 0;
  }
};

const handleDeleteDoc = async (doc) => {
  try {
    await ElMessageBox.confirm(`确认删除文档「${doc.title}」及其全部知识分块？`, '删除确认', {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning'
    });
  } catch { return; }

  try {
    await deleteDocument(doc.id);
    ElMessage.success('已删除');
    await loadDocs();
    await loadStats();
  } catch (e) {
    ElMessage.error('删除失败：' + (e.message || ''));
  }
};

const formatSize = (bytes) => {
  if (!bytes) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB'];
  let i = 0;
  let v = bytes;
  while (v >= 1024 && i < units.length - 1) { v /= 1024; i++; }
  return v.toFixed(1) + ' ' + units[i];
};

// ===== 在线编辑器 =====
const editorDialog = ref(false);
const editorDoc = ref(null);
const editorContent = ref('');
const editorSaving = ref(false);
const editorLoading = ref(false);
const activeSheetIndex = ref(0);

// 图片类型集合
const IMAGE_EXTS = ['.jpg', '.jpeg', '.png', '.gif', '.webp', '.bmp', '.svg'];
const isImageDoc = (doc) => {
  const ext = (doc.file_type || '').toLowerCase();
  return IMAGE_EXTS.includes(ext);
};
const isEditableDoc = (doc) => {
  const ext = (doc.file_type || '').toLowerCase();
  return ['.txt', '.md', '.markdown', '.json', '.csv', '.log'].includes(ext);
};
const isPreviewableDoc = (doc) => {
  const ext = (doc.file_type || '').toLowerCase();
  return ['.docx', '.xlsx', '.xls'].includes(ext);
};
const isExcelDoc = (doc) => {
  const ext = (doc.file_type || '').toLowerCase();
  return ['.xlsx', '.xls'].includes(ext);
};

// 解析 Excel 文本为结构化工作表数据
const parsedSheets = computed(() => {
  if (!editorDoc.value || !isExcelDoc(editorDoc.value) || !editorContent.value) return [];
  const text = editorContent.value;
  const sheets = [];
  const lines = text.split('\n');
  let currentSheet = null;

  for (const line of lines) {
    // 检测工作表标题行: "--- 工作表: xxx ---"
    const sheetMatch = line.match(/^---\s*工作表:\s*(.+?)\s*---$/);
    if (sheetMatch) {
      if (currentSheet) sheets.push(currentSheet);
      currentSheet = { name: sheetMatch[1], headers: [], rows: [] };
      continue;
    }
    if (!currentSheet) continue;

    // 检测表头行: "[表头] col1\tcol2\t..."
    const headerMatch = line.match(/^\[表头\]\s*(.+)$/);
    if (headerMatch) {
      currentSheet.headers = headerMatch[1].split('\t').filter(c => c.trim() !== '');
      continue;
    }

    // 空行跳过
    if (line.trim() === '') continue;

    // 数据行: tab 分隔
    const cells = line.split('\t');
    if (cells.some(c => c.trim() !== '')) {
      currentSheet.rows.push(cells);
    }
  }
  if (currentSheet) sheets.push(currentSheet);
  return sheets;
});

const getFileIcon = (doc) => {
  const ext = (doc.file_type || '').toLowerCase();
  if (ext === '.pdf') return '📕';
  if (ext === '.docx') return '📘';
  if (ext === '.xlsx' || ext === '.xls') return '📗';
  if (ext === '.md' || ext === '.markdown') return '📝';
  if (IMAGE_EXTS.includes(ext)) return '🖼️';
  if (ext === '.json') return '📋';
  if (ext === '.csv') return '📊';
  return '📄';
};

const openEditor = async (doc) => {
  editorDoc.value = doc;
  editorDialog.value = true;
  editorContent.value = '';
  editorLoading.value = true;
  activeSheetIndex.value = 0;
  try {
    const res = await getDocumentContent(doc.id);
    editorContent.value = res.content || '';
  } catch (e) {
    ElMessage.error('加载内容失败：' + (e.message || ''));
  } finally {
    editorLoading.value = false;
  }
};

const handleSaveContent = async () => {
  if (!editorDoc.value) return;
  editorSaving.value = true;
  try {
    await updateDocumentContent(editorDoc.value.id, editorContent.value);
    ElMessage.success('文档已保存并重新向量化');
    editorDialog.value = false;
    await loadDocs();
    await loadStats();
  } catch (e) {
    ElMessage.error('保存失败：' + (e.message || ''));
  } finally {
    editorSaving.value = false;
  }
};

// ===== 图片上传 =====
const imageUploadDialog = ref(false);
const imageUploadFile = ref(null);
const imageUploadForm = reactive({ title: '', description: '', category: 'image' });
const imageUploading = ref(false);
const imageFileInputRef = ref(null);
const openImagePicker = () => imageFileInputRef.value?.click();

const handleImageSelect = (e) => { imageUploadFile.value = e.target.files[0]; };

const handleImageUpload = async () => {
  if (!imageUploadFile.value) {
    ElMessage.warning('请选择图片');
    return;
  }
  imageUploading.value = true;
  try {
    const doc = await uploadImage(imageUploadFile.value, {
      title: imageUploadForm.title || imageUploadFile.value.name,
      description: imageUploadForm.description,
      category: imageUploadForm.category
    });
    ElMessage.success(`图片「${doc.title}」上传成功`);
    logsStore.addLog('上传', '知识库', `上传图片: ${doc.title}`);
    imageUploadDialog.value = false;
    imageUploadFile.value = null;
    imageUploadForm.title = '';
    imageUploadForm.description = '';
    docPage.value = 1;
    await loadDocs();
    await loadStats();
  } catch (e) {
    ElMessage.error('上传失败：' + (e.message || ''));
  } finally {
    imageUploading.value = false;
  }
};

const statusMap = {
  processing: { label: '处理中', color: '#FF7D00' },
  ready: { label: '已入库', color: '#00B42A' },
  failed: { label: '失败', color: '#F53F3F' }
};

// ===== 问题库 =====
const faqs = ref([]);
const faqTotal = ref(0);
const faqPage = ref(1);
const faqPageSize = ref(12);
const faqCategory = ref('all');
const faqDialog = ref(false);
const faqForm = reactive({ question: '', answer: '', category: 'other', tags: [] });

const loadFAQs = async () => {
  loading.value = true;
  try {
    const res = await listFAQs({
      page: faqPage.value,
      pageSize: faqPageSize.value,
      category: faqCategory.value
    });
    faqs.value = res.data;
    faqTotal.value = res.total;
  } catch (e) {
    ElMessage.error('加载问题库失败：' + (e.message || ''));
  } finally {
    loading.value = false;
  }
};

const onFaqPageChange = (p) => { faqPage.value = p; loadFAQs(); };
const onFaqCategoryChange = () => { faqPage.value = 1; loadFAQs(); };

const handleAddFAQ = async () => {
  if (!faqForm.question.trim() || !faqForm.answer.trim()) {
    ElMessage.warning('请填写问题与答案');
    return;
  }
  try {
    await createFAQ({
      question: faqForm.question.trim(),
      answer: faqForm.answer.trim(),
      category: faqForm.category,
      tags: faqForm.tags.length ? faqForm.tags : [faqForm.category]
    });
    ElMessage.success('问题已入库');
    logsStore.addLog('新增', '问题库', `新增问题: ${faqForm.question.slice(0, 20)}`);
    faqDialog.value = false;
    faqForm.question = '';
    faqForm.answer = '';
    await loadFAQs();
    await loadStats();
  } catch (e) {
    ElMessage.error('新增失败：' + (e.message || ''));
  }
};

const handleDeleteFAQ = async (faq) => {
  try {
    await ElMessageBox.confirm(`确认删除问题「${faq.question.slice(0, 30)}」？`, '删除确认', {
      confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning'
    });
  } catch { return; }
  try {
    await deleteFAQ(faq.id);
    ElMessage.success('已删除');
    await loadFAQs();
    await loadStats();
  } catch (e) {
    ElMessage.error('删除失败：' + (e.message || ''));
  }
};

// ===== 智能搜索 =====
const searchQuery = ref('');
const searchResults = ref([]);
const searching = ref(false);
const searched = ref(false);

const handleSearch = async () => {
  const q = searchQuery.value.trim();
  if (!q) return;
  searching.value = true;
  searched.value = false;
  try {
    const res = await searchKnowledge(q, { topK: 10, minScore: 0.25 });
    searchResults.value = res.results;
    searched.value = true;
    logsStore.addLog('搜索', '知识库', `语义搜索: ${q}`);
  } catch (e) {
    ElMessage.error('搜索失败：' + (e.message || ''));
  } finally {
    searching.value = false;
  }
};

const formatScore = (s) => ((s || 0) * 100).toFixed(1) + '%';
const scoreColor = (s) => {
  if (s >= 0.7) return '#00B42A';
  if (s >= 0.5) return '#FF7D00';
  return '#F53F3F';
};

// ===== 统计 =====
const loadStats = async () => {
  try {
    stats.value = await getKnowledgeStats();
  } catch { /* 统计失败不阻塞页面 */ }
};

const switchTab = (tab) => {
  activeTab.value = tab;
};

onMounted(() => {
  loadStats();
  loadDocs();
  loadFAQs();
});
</script>

<template>
  <div class="knowledge-base-page">
    <!-- 顶部统计卡片 -->
    <GlassCard title="内部知识库" class="stats-card">
      <div class="stats-grid">
        <div class="stat-item">
          <div class="stat-icon" style="background: rgba(22,93,255,.12); color:#165DFF;">📄</div>
          <div class="stat-info">
            <div class="stat-value">{{ stats?.documents ?? '--' }}</div>
            <div class="stat-label">文档</div>
          </div>
        </div>
        <div class="stat-item">
          <div class="stat-icon" style="background: rgba(0,180,42,.12); color:#00B42A;">🧩</div>
          <div class="stat-info">
            <div class="stat-value">{{ stats?.chunks ?? '--' }}</div>
            <div class="stat-label">知识分块</div>
          </div>
        </div>
        <div class="stat-item">
          <div class="stat-icon" style="background: rgba(114,46,209,.12); color:#722ED1;">❓</div>
          <div class="stat-info">
            <div class="stat-value">{{ stats?.faqs ?? '--' }}</div>
            <div class="stat-label">问题条目</div>
          </div>
        </div>
        <div class="stat-item">
          <div class="stat-icon"
               :style="stats?.ollama_online ? { background:'rgba(0,180,42,.12)', color:'#00B42A' } : { background:'rgba(245,63,63,.12)', color:'#F53F3F' }">
            🧠
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ stats?.ollama_online ? '在线' : '离线' }}</div>
            <div class="stat-label" :title="(stats?.models || []).join(', ')">本地模型</div>
          </div>
        </div>
      </div>
    </GlassCard>

    <!-- 操作栏 -->
    <GlassCard class="action-card">
      <div class="action-bar">
        <div class="tab-segment">
          <button class="tab-btn" :class="{ active: activeTab === 'docs' }" @click="switchTab('docs')">📄 文档知识库</button>
          <button class="tab-btn" :class="{ active: activeTab === 'faqs' }" @click="switchTab('faqs')">❓ 问题库</button>
          <button class="tab-btn" :class="{ active: activeTab === 'search' }" @click="switchTab('search')">🔍 智能搜索</button>
        </div>
        <div class="action-right" v-if="activeTab === 'docs'">
          <div class="upload-row">
            <button class="primary-btn" @click="uploadDialog = true">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
                <polyline points="17 8 12 3 7 8" />
                <line x1="12" y1="3" x2="12" y2="15" />
              </svg>
              上传文档
            </button>
            <button class="image-btn" @click="imageUploadDialog = true">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <rect x="3" y="3" width="18" height="18" rx="2" ry="2" />
                <circle cx="8.5" cy="8.5" r="1.5" />
                <polyline points="21 15 16 10 5 21" />
              </svg>
              上传图片
            </button>
          </div>
        </div>
        <div class="action-right" v-if="activeTab === 'faqs'">
          <button class="primary-btn" @click="faqDialog = true">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <line x1="12" y1="5" x2="12" y2="19" /><line x1="5" y1="12" x2="19" y2="12" />
            </svg>
            新增问题
          </button>
        </div>
      </div>
    </GlassCard>

    <!-- 文档知识库 -->
    <GlassCard v-if="activeTab === 'docs'" title="文档列表">
      <div class="filter-bar">
        <div class="filter-group">
          <label>分类</label>
          <select v-model="docCategory" @change="onCategoryChange">
            <option value="all">全部分类</option>
            <option v-for="c in categoryOptions" :key="c.code" :value="c.code">{{ c.name }}</option>
          </select>
        </div>
        <span class="filter-count">共 {{ docTotal }} 篇文档</span>
      </div>

      <div v-if="docs.length === 0 && !loading" class="empty-state">
        <div class="empty-icon">📄</div>
        <p>还没有文档，点击右上角「上传文档」开始构建知识库</p>
      </div>

      <div class="doc-grid" v-if="docs.length">
        <div v-for="doc in docs" :key="doc.id" class="doc-card" @dblclick="(isEditableDoc(doc) || isPreviewableDoc(doc)) ? openEditor(doc) : null">
          <div class="doc-header">
            <span class="doc-icon" :style="{ background: getCategoryInfo(doc.category).color + '18' }">{{ getFileIcon(doc) }}</span>
            <span class="doc-category" :style="{ color: getCategoryInfo(doc.category).color, background: getCategoryInfo(doc.category).color + '12' }">
              {{ getCategoryInfo(doc.category).name }}
            </span>
          </div>
          <!-- 图片文档显示缩略图 -->
          <div v-if="isImageDoc(doc)" class="doc-image-preview">
            <img :src="getFileUrl(doc.id)" :alt="doc.title" loading="lazy" @click.stop="openEditor(doc)" />
          </div>
          <div class="doc-title" :title="doc.title">{{ doc.title }}</div>
          <div class="doc-meta">
            <span>{{ doc.file_name }}</span>
            <span>{{ formatSize(doc.file_size) }}</span>
          </div>
          <div class="doc-footer">
            <span class="doc-status" :style="{ color: statusMap[doc.status]?.color, background: (statusMap[doc.status]?.color || '#999') + '14' }">
              {{ statusMap[doc.status]?.label || doc.status }}
            </span>
            <span class="doc-chunks">{{ doc.chunk_count }} 分块</span>
            <!-- 编辑按钮 -->
            <button v-if="isEditableDoc(doc) || isPreviewableDoc(doc) || isImageDoc(doc)"
                    class="doc-edit" @click.stop="openEditor(doc)" title="查看/编辑">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" />
                <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" />
              </svg>
            </button>
            <button class="doc-delete" @click="handleDeleteDoc(doc)" title="删除">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <polyline points="3 6 5 6 21 6" />
                <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
              </svg>
            </button>
          </div>
        </div>
      </div>

      <div class="pagination" v-if="docTotal > docPageSize">
        <button :disabled="docPage <= 1" @click="onDocPageChange(docPage - 1)">上一页</button>
        <span>{{ docPage }} / {{ Math.ceil(docTotal / docPageSize) }}</span>
        <button :disabled="docPage >= Math.ceil(docTotal / docPageSize)" @click="onDocPageChange(docPage + 1)">下一页</button>
      </div>
    </GlassCard>

    <!-- 问题库 -->
    <GlassCard v-if="activeTab === 'faqs'" title="问题库">
      <div class="filter-bar">
        <div class="filter-group">
          <label>分类</label>
          <select v-model="faqCategory" @change="onFaqCategoryChange">
            <option value="all">全部分类</option>
            <option v-for="c in categoryOptions" :key="c.code" :value="c.code">{{ c.name }}</option>
          </select>
        </div>
        <span class="filter-count">共 {{ faqTotal }} 条问题</span>
      </div>

      <div v-if="faqs.length === 0 && !loading" class="empty-state">
        <div class="empty-icon">❓</div>
        <p>问题库还是空的，点击右上角「新增问题」添加第一条</p>
      </div>

      <div class="faq-list" v-if="faqs.length">
        <div v-for="faq in faqs" :key="faq.id" class="faq-item">
          <div class="faq-left">
            <div class="faq-question">
              <span class="faq-q-icon" :style="{ background: getCategoryInfo(faq.category).color + '14', color: getCategoryInfo(faq.category).color }">Q</span>
              {{ faq.question }}
            </div>
            <div class="faq-answer">
              <span class="faq-a-icon" :style="{ background: 'rgba(0,180,42,.12)', color: '#00B42A' }">A</span>
              {{ faq.answer }}
            </div>
          </div>
          <div class="faq-right">
            <span class="faq-cat" :style="{ color: getCategoryInfo(faq.category).color, background: getCategoryInfo(faq.category).color + '12' }">
              {{ getCategoryInfo(faq.category).name }}
            </span>
            <button class="faq-delete" @click="handleDeleteFAQ(faq)">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <polyline points="3 6 5 6 21 6" />
                <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
              </svg>
            </button>
          </div>
        </div>
      </div>

      <div class="pagination" v-if="faqTotal > faqPageSize">
        <button :disabled="faqPage <= 1" @click="onFaqPageChange(faqPage - 1)">上一页</button>
        <span>{{ faqPage }} / {{ Math.ceil(faqTotal / faqPageSize) }}</span>
        <button :disabled="faqPage >= Math.ceil(faqTotal / faqPageSize)" @click="onFaqPageChange(faqPage + 1)">下一页</button>
      </div>
    </GlassCard>

    <!-- 智能搜索 -->
    <GlassCard v-if="activeTab === 'search'" title="智能语义搜索">
      <div class="search-box">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="11" cy="11" r="8" />
          <line x1="21" y1="21" x2="16.65" y2="16.65" />
        </svg>
        <input
          v-model="searchQuery"
          type="text"
          placeholder="输入问题或关键词，语义检索文档和问题库..."
          @keyup.enter="handleSearch"
        />
        <button class="search-btn" @click="handleSearch" :disabled="searching">
          {{ searching ? '搜索中...' : '搜索' }}
        </button>
      </div>

      <div v-if="searching" class="search-loading">
        <div class="spinner"></div>
        <span>正在向量化并检索知识库...</span>
      </div>

      <div v-else-if="searched && searchResults.length === 0" class="empty-state">
        <div class="empty-icon">🔍</div>
        <p>未检索到相关内容，建议调整关键词，或先上传相关文档</p>
      </div>

      <div v-if="searchResults.length" class="search-results">
        <div v-for="(r, i) in searchResults" :key="i" class="search-item">
          <div class="search-item-header">
            <span class="score-badge" :style="{ color: scoreColor(r.score), background: scoreColor(r.score) + '14' }">
              相关度 {{ formatScore(r.score) }}
            </span>
            <span class="source-badge" :class="r.source">
              {{ r.source === 'faq' ? '问题库' : '文档' }}
            </span>
            <span class="doc-name" v-if="r.source === 'document'">{{ r.document_name }}</span>
          </div>
          <div class="search-item-content">{{ r.content }}</div>
        </div>
      </div>
    </GlassCard>

    <!-- 上传文档弹窗 -->
    <div v-if="uploadDialog" class="modal-mask" @click.self="uploadDialog = false">
      <div class="modal">
        <div class="modal-header">
          <h3>上传文档</h3>
          <button class="modal-close" @click="uploadDialog = false">×</button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>选择文件</label>
            <div class="file-upload" @click="openFilePicker">
              <input ref="fileInputRef" type="file" style="display:none" @change="handleFileSelect"
                     accept=".txt,.md,.markdown,.json,.csv,.log,.docx,.pdf,.xlsx,.xls,.jpg,.jpeg,.png,.gif,.webp" />
              <template v-if="uploadFile">
                <div class="file-selected">📄 {{ uploadFile.name }}（{{ formatSize(uploadFile.size) }}）</div>
              </template>
              <template v-else>
                <div class="file-placeholder">
                  <span>点击选择文件</span>
                  <small>支持 TXT / MD / JSON / CSV / LOG / DOCX / PDF / XLSX / JPG / PNG / GIF</small>
                </div>
              </template>
            </div>
          </div>
          <div class="form-group">
            <label>文档标题（可选，默认用文件名）</label>
            <input v-model="uploadForm.title" type="text" class="form-input" placeholder="文档标题" />
          </div>
          <div class="form-group">
            <label>分类</label>
            <select v-model="uploadForm.category" class="form-input">
              <option v-for="c in categoryOptions" :key="c.code" :value="c.code">{{ c.name }}</option>
            </select>
          </div>
          <div class="form-group">
            <label>描述（可选）</label>
            <input v-model="uploadForm.description" type="text" class="form-input" placeholder="简短描述文档内容" />
          </div>
        </div>
        <div class="modal-footer">
          <button class="cancel-btn" @click="uploadDialog = false" :disabled="uploading">取消</button>
          <button class="submit-btn" @click="handleUpload" :disabled="uploading">
            {{ uploading ? ('上传并向量化中 ' + uploadProgress + '%') : '上传入库' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 新增问题弹窗 -->
    <div v-if="faqDialog" class="modal-mask" @click.self="faqDialog = false">
      <div class="modal">
        <div class="modal-header">
          <h3>新增问题</h3>
          <button class="modal-close" @click="faqDialog = false">×</button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>问题描述 *</label>
            <textarea v-model="faqForm.question" class="form-textarea" rows="3" placeholder="用户可能提出的问题..."></textarea>
          </div>
          <div class="form-group">
            <label>解决方案 *</label>
            <textarea v-model="faqForm.answer" class="form-textarea" rows="4" placeholder="详细的解决方案或操作步骤..."></textarea>
          </div>
          <div class="form-group">
            <label>分类</label>
            <select v-model="faqForm.category" class="form-input">
              <option v-for="c in categoryOptions" :key="c.code" :value="c.code">{{ c.name }}</option>
            </select>
          </div>
        </div>
        <div class="modal-footer">
          <button class="cancel-btn" @click="faqDialog = false">取消</button>
          <button class="submit-btn" @click="handleAddFAQ">保存入库</button>
        </div>
      </div>
    </div>

    <!-- 在线编辑器弹窗 -->
    <div v-if="editorDialog" class="modal-mask" @click.self="editorDialog = false">
      <div class="modal editor-modal">
        <div class="modal-header">
          <h3>
            <span v-if="editorDoc && isImageDoc(editorDoc)">🖼️ 图片预览</span>
            <span v-else-if="editorDoc && isExcelDoc(editorDoc)">📊 表格预览</span>
            <span v-else-if="editorDoc && isPreviewableDoc(editorDoc)">📘 文档预览</span>
            <span v-else>📝 在线编辑</span>
            — {{ editorDoc?.title }}
          </h3>
          <button class="modal-close" @click="editorDialog = false">×</button>
        </div>
        <div class="modal-body">
          <!-- 加载中 -->
          <div v-if="editorLoading" class="editor-loading">
            <div class="spinner"></div>
            <span>加载内容中...</span>
          </div>
          <!-- 图片预览 -->
          <div v-else-if="editorDoc && isImageDoc(editorDoc)" class="image-preview-container">
            <img :src="getFileUrl(editorDoc.id)" :alt="editorDoc.title" class="image-preview-img" />
            <div class="image-info">
              <p><strong>文件名：</strong>{{ editorDoc.file_name }}</p>
              <p><strong>大小：</strong>{{ formatSize(editorDoc.file_size) }}</p>
            </div>
          </div>
          <!-- Excel 表格预览 -->
          <div v-else-if="editorDoc && isExcelDoc(editorDoc) && parsedSheets.length > 0" class="excel-preview">
            <div v-if="parsedSheets.length > 1" class="sheet-tabs">
              <button v-for="(sheet, idx) in parsedSheets" :key="idx"
                      :class="['sheet-tab-btn', { active: activeSheetIndex === idx }]"
                      @click="activeSheetIndex = idx">
                {{ sheet.name }}
              </button>
            </div>
            <div class="sheet-content">
              <div v-for="(sheet, sIdx) in parsedSheets" :key="sIdx" v-show="sIdx === (parsedSheets.length > 1 ? activeSheetIndex : 0)">
                <div class="sheet-name" v-if="parsedSheets.length > 1">{{ sheet.name }}</div>
                <div class="table-wrapper">
                  <table class="excel-table">
                    <thead v-if="sheet.headers.length > 0">
                      <tr>
                        <th class="row-num-header">#</th>
                        <th v-for="(h, i) in sheet.headers" :key="i">{{ h || '—' }}</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr v-for="(row, rIdx) in sheet.rows" :key="rIdx">
                        <td class="row-num">{{ rIdx + 1 }}</td>
                        <td v-for="(cell, cIdx) in sheet.headers.length" :key="cIdx">
                          {{ row[cIdx] || '' }}
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>
            </div>
          </div>
          <!-- 文本编辑器 -->
          <div v-else class="editor-container">
            <textarea
              v-model="editorContent"
              class="editor-textarea"
              :disabled="editorDoc && isPreviewableDoc(editorDoc)"
              placeholder="文档内容为空"
              spellcheck="false"
            ></textarea>
          </div>
        </div>
        <div class="modal-footer">
          <button class="cancel-btn" @click="editorDialog = false">关闭</button>
          <button v-if="editorDoc && isEditableDoc(editorDoc)"
                  class="submit-btn" @click="handleSaveContent" :disabled="editorSaving">
            {{ editorSaving ? '保存并重新向量化...' : '保存并重新向量化' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 图片上传弹窗 -->
    <div v-if="imageUploadDialog" class="modal-mask" @click.self="imageUploadDialog = false">
      <div class="modal">
        <div class="modal-header">
          <h3>上传图片</h3>
          <button class="modal-close" @click="imageUploadDialog = false">×</button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>选择图片</label>
            <div class="file-upload" @click="openImagePicker">
              <input ref="imageFileInputRef" type="file" style="display:none" @change="handleImageSelect"
                     accept=".jpg,.jpeg,.png,.gif,.webp,.bmp,.svg" />
              <template v-if="imageUploadFile">
                <div class="file-selected">🖼️ {{ imageUploadFile.name }}（{{ formatSize(imageUploadFile.size) }}）</div>
                <div class="image-upload-preview">
                  <img :src="URL.createObjectURL(imageUploadFile)" alt="预览" />
                </div>
              </template>
              <template v-else>
                <div class="file-placeholder">
                  <span>点击选择图片</span>
                  <small>支持 JPG / PNG / GIF / WebP / BMP / SVG</small>
                </div>
              </template>
            </div>
          </div>
          <div class="form-group">
            <label>标题（可选，默认用文件名）</label>
            <input v-model="imageUploadForm.title" type="text" class="form-input" placeholder="图片标题" />
          </div>
          <div class="form-group">
            <label>描述（可选）</label>
            <input v-model="imageUploadForm.description" type="text" class="form-input" placeholder="简短描述图片内容" />
          </div>
          <div class="form-group">
            <label>分类</label>
            <select v-model="imageUploadForm.category" class="form-input">
              <option v-for="c in categoryOptions" :key="c.code" :value="c.code">{{ c.name }}</option>
            </select>
          </div>
        </div>
        <div class="modal-footer">
          <button class="cancel-btn" @click="imageUploadDialog = false">取消</button>
          <button class="submit-btn" @click="handleImageUpload" :disabled="imageUploading">
            {{ imageUploading ? '上传中...' : '上传图片' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.knowledge-base-page {
  padding: 20px;
}

.stats-card {
  margin-bottom: 16px;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
  padding: 16px;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 16px;
  background: var(--bg-input);
  border-radius: 10px;
  border: 1px solid var(--border-color);
}

.stat-icon {
  width: 44px;
  height: 44px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 22px;
  flex-shrink: 0;
}

.stat-value {
  font-size: 22px;
  font-weight: 700;
  color: var(--text-primary);
}

.stat-label {
  font-size: 12px;
  color: var(--text-tertiary);
}

.action-card {
  margin-bottom: 16px;
}

.action-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  flex-wrap: wrap;
  gap: 12px;
}

.action-bar .action-right {
  display: flex;
  flex-direction: row;
  gap: 10px;
}

.upload-row {
  display: flex;
  align-items: center;
  gap: 10px;
}

.tab-segment {
  display: flex;
  gap: 4px;
  padding: 4px;
  background: var(--bg-input);
  border-radius: 10px;
}

.tab-btn {
  padding: 8px 18px;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: var(--text-secondary);
  font-size: 14px;
  cursor: pointer;
  transition: all .2s;
}

.tab-btn:hover { color: #165DFF; }
.tab-btn.active {
  background: white;
  color: #165DFF;
  font-weight: 600;
  box-shadow: 0 2px 8px rgba(0,0,0,.08);
}

.primary-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 18px;
  background: linear-gradient(135deg, #165DFF, #4080FF);
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all .2s;
  box-shadow: 0 2px 8px rgba(22,93,255,.25);
}

.primary-btn:hover { transform: translateY(-1px); box-shadow: 0 4px 14px rgba(22,93,255,.35); }
.primary-btn:disabled { opacity: .5; cursor: not-allowed; transform: none; }

.filter-bar {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 12px 0;
}

.filter-group {
  display: flex;
  align-items: center;
  gap: 8px;
}

.filter-group label { font-size: 13px; color: var(--text-secondary); }

.filter-group select {
  padding: 6px 12px;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  background: var(--bg-card);
  color: var(--text-primary);
  font-size: 13px;
}

.filter-count { font-size: 13px; color: var(--text-tertiary); margin-left: auto; }

/* 文档卡片 */
.doc-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 16px;
}

.doc-card {
  padding: 16px;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  transition: all .25s;
}

.doc-card:hover {
  transform: translateY(-2px);
  border-color: #c5d8ff;
  box-shadow: 0 8px 20px rgba(22,93,255,.12);
}

.doc-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.doc-icon {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
}

.doc-category {
  font-size: 11px;
  padding: 3px 8px;
  border-radius: 4px;
  font-weight: 500;
}

.doc-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 6px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.doc-meta {
  font-size: 12px;
  color: var(--text-tertiary);
  display: flex;
  justify-content: space-between;
  margin-bottom: 12px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.doc-footer {
  display: flex;
  align-items: center;
  gap: 8px;
}

.doc-status {
  font-size: 11px;
  padding: 3px 8px;
  border-radius: 4px;
}

.doc-chunks { font-size: 12px; color: var(--text-tertiary); margin-left: auto; }

.doc-delete {
  border: none;
  background: rgba(245,63,63,.08);
  color: #F53F3F;
  width: 26px;
  height: 26px;
  border-radius: 6px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all .2s;
}

.doc-delete:hover { background: #F53F3F; color: white; }

/* FAQ */
.faq-list { display: flex; flex-direction: column; gap: 12px; }

.faq-item {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  padding: 14px 16px;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 10px;
}

.faq-left { flex: 1; min-width: 0; }

.faq-question {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 8px;
}

.faq-q-icon, .faq-a-icon {
  width: 20px;
  height: 20px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  margin-top: 1px;
}

.faq-answer {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  font-size: 13px;
  color: var(--text-secondary);
  line-height: 1.6;
  white-space: pre-wrap;
  padding-left: 28px;
}

.faq-right { display: flex; flex-direction: column; align-items: flex-end; gap: 10px; flex-shrink: 0; }

.faq-cat { font-size: 11px; padding: 3px 8px; border-radius: 4px; }

.faq-delete {
  border: none;
  background: rgba(245,63,63,.08);
  color: #F53F3F;
  width: 26px;
  height: 26px;
  border-radius: 6px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all .2s;
}

.faq-delete:hover { background: #F53F3F; color: white; }

/* 搜索 */
.search-box {
  display: flex;
  align-items: center;
  gap: 10px;
  background: var(--bg-input);
  border: 1px solid var(--border-color);
  border-radius: 10px;
  padding: 12px 16px;
  margin-bottom: 16px;
}

.search-box svg { color: var(--text-tertiary); flex-shrink: 0; }

.search-box input {
  flex: 1;
  background: none;
  border: none;
  outline: none;
  color: var(--text-primary);
  font-size: 14px;
}

.search-btn {
  padding: 6px 18px;
  background: #165DFF;
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 13px;
  cursor: pointer;
}

.search-btn:disabled { opacity: .5; cursor: not-allowed; }

.search-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 40px;
  color: var(--text-tertiary);
}

.spinner {
  width: 20px;
  height: 20px;
  border: 2px solid rgba(22,93,255,.2);
  border-top-color: #165DFF;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin { to { transform: rotate(360deg); } }

.search-results { display: flex; flex-direction: column; gap: 12px; }

.search-item {
  padding: 14px 16px;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 10px;
}

.search-item-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
}

.score-badge, .source-badge {
  font-size: 11px;
  padding: 3px 8px;
  border-radius: 4px;
  font-weight: 500;
}

.source-badge.document { background: rgba(22,93,255,.1); color: #165DFF; }
.source-badge.faq { background: rgba(114,46,209,.1); color: #722ED1; }

.doc-name { font-size: 12px; color: var(--text-tertiary); margin-left: auto; }

.search-item-content {
  font-size: 13px;
  color: var(--text-secondary);
  line-height: 1.7;
  white-space: pre-wrap;
  max-height: 120px;
  overflow-y: auto;
}

/* 弹窗 */
.modal-mask {
  position: fixed;
  top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(0,0,0,.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  backdrop-filter: blur(2px);
}

.modal {
  background: white;
  border-radius: 14px;
  width: 520px;
  max-width: 92vw;
  max-height: 88vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 20px 60px rgba(0,0,0,.2);
  animation: modalIn .2s ease-out;
}

@keyframes modalIn {
  from { opacity: 0; transform: scale(.96) translateY(-10px); }
  to { opacity: 1; transform: scale(1) translateY(0); }
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid #f0f0f0;
}

.modal-header h3 { margin: 0; font-size: 16px; }

.modal-close {
  border: none;
  background: none;
  font-size: 22px;
  color: #999;
  cursor: pointer;
  width: 28px;
  height: 28px;
  border-radius: 4px;
}

.modal-close:hover { background: #f5f5f5; color: #333; }

.modal-body { padding: 20px; overflow-y: auto; flex: 1; }

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 12px 20px;
  border-top: 1px solid #f0f0f0;
  background: #fafbfc;
  border-radius: 0 0 14px 14px;
}

.form-group { margin-bottom: 16px; }
.form-group label { display: block; font-size: 13px; color: #4e5969; margin-bottom: 6px; font-weight: 500; }

.form-input, .form-textarea {
  width: 100%;
  box-sizing: border-box;
  padding: 8px 12px;
  border: 1px solid #dcdfe6;
  border-radius: 6px;
  font-size: 14px;
  color: #1d2129;
  outline: none;
  transition: border-color .2s;
  font-family: inherit;
}

.form-textarea { resize: vertical; }

.form-input:focus, .form-textarea:focus {
  border-color: #165DFF;
  box-shadow: 0 0 0 3px rgba(22,93,255,.1);
}

.file-upload {
  border: 2px dashed var(--border-color, #dcdfe6);
  border-radius: 8px;
  padding: 24px;
  text-align: center;
  cursor: pointer;
  transition: border-color .2s;
}

.file-upload:hover { border-color: #165DFF; }

.file-selected { color: #165DFF; font-size: 14px; }

.file-placeholder span { display: block; font-size: 14px; color: #666; margin-bottom: 6px; }
.file-placeholder small { font-size: 12px; color: #999; }

.cancel-btn, .submit-btn {
  padding: 8px 20px;
  border-radius: 6px;
  font-size: 14px;
  cursor: pointer;
  border: 1px solid transparent;
  transition: all .2s;
}

.cancel-btn { background: white; color: #4e5969; border-color: #dcdfe6; }
.cancel-btn:hover { border-color: #165DFF; color: #165DFF; }
.submit-btn { background: #165DFF; color: white; }
.submit-btn:hover { background: #0e42d2; }
.submit-btn:disabled { opacity: .5; cursor: not-allowed; }

/* 空状态 */
.empty-state {
  text-align: center;
  padding: 60px 20px;
  color: var(--text-tertiary);
}

.empty-icon { font-size: 42px; margin-bottom: 12px; }
.empty-state p { margin: 0; font-size: 14px; }

/* 分页 */
.pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
  padding: 20px 0 4px;
}

.pagination button {
  padding: 6px 16px;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  background: var(--bg-card);
  color: var(--text-secondary);
  font-size: 13px;
  cursor: pointer;
}

.pagination button:disabled { opacity: .4; cursor: not-allowed; }
.pagination span { font-size: 13px; color: var(--text-tertiary); }

/* 文档卡片 - 图片缩略图 */
.doc-image-preview {
  margin: 8px 0;
  border-radius: 8px;
  overflow: hidden;
  max-height: 160px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f7fa;
}

.doc-image-preview img {
  width: 100%;
  max-height: 160px;
  object-fit: cover;
  cursor: pointer;
  transition: transform .2s;
}

.doc-image-preview img:hover { transform: scale(1.03); }

/* 编辑按钮 */
.doc-edit {
  border: none;
  background: rgba(22,93,255,.08);
  color: #165DFF;
  width: 26px;
  height: 26px;
  border-radius: 6px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all .2s;
}

.doc-edit:hover { background: #165DFF; color: white; }

/* 图片上传按钮 */
.image-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background: rgba(114,46,209,.08);
  color: #722ED1;
  border: 1px solid rgba(114,46,209,.2);
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all .2s;
}

.image-btn:hover { background: #722ED1; color: white; }

/* 编辑器弹窗 */
.editor-modal {
  width: 800px;
  max-width: 95vw;
}

.editor-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 60px;
  color: var(--text-tertiary);
}

.editor-container { width: 100%; }

.editor-textarea {
  width: 100%;
  min-height: 400px;
  max-height: 60vh;
  padding: 16px;
  border: 1px solid var(--border-color, #dcdfe6);
  border-radius: 8px;
  font-family: 'SF Mono', 'Fira Code', 'Consolas', monospace;
  font-size: 13px;
  line-height: 1.6;
  color: var(--text-primary);
  background: var(--bg-input, #f7f8fa);
  resize: vertical;
  outline: none;
  box-sizing: border-box;
  tab-size: 2;
}

.editor-textarea:focus {
  border-color: #165DFF;
  box-shadow: 0 0 0 3px rgba(22,93,255,.1);
}

/* 图片预览 */
.image-preview-container {
  text-align: center;
}

.image-preview-img {
  max-width: 100%;
  max-height: 60vh;
  border-radius: 8px;
  object-fit: contain;
  box-shadow: 0 2px 12px rgba(0,0,0,.1);
}

.image-info {
  text-align: left;
  margin-top: 16px;
  padding: 12px 16px;
  background: var(--bg-input, #f7f8fa);
  border-radius: 8px;
  font-size: 13px;
  color: var(--text-secondary);
}

.image-info p { margin: 4px 0; }

/* 图片上传预览 */
.image-upload-preview {
  margin-top: 12px;
}

.image-upload-preview img {
  max-width: 200px;
  max-height: 150px;
  border-radius: 8px;
  object-fit: cover;
  border: 1px solid var(--border-color, #dcdfe6);
}

/* ===== Excel 表格预览 ===== */
.excel-preview {
  overflow: auto;
  max-height: 65vh;
}

.sheet-tabs {
  display: flex;
  gap: 4px;
  padding: 8px 0;
  border-bottom: 1px solid var(--border-color, #e5e7eb);
  margin-bottom: 8px;
  flex-wrap: wrap;
}

.sheet-tab-btn {
  padding: 6px 16px;
  border: 1px solid var(--border-color, #dcdfe6);
  border-radius: 6px;
  background: var(--bg-input, #f7f8fa);
  color: var(--text-secondary);
  font-size: 13px;
  cursor: pointer;
  transition: all .2s;
}

.sheet-tab-btn:hover { color: #165DFF; border-color: #c5d8ff; }
.sheet-tab-btn.active {
  background: #165DFF;
  color: white;
  border-color: #165DFF;
  font-weight: 600;
}

.sheet-name {
  font-size: 13px;
  color: var(--text-secondary);
  margin-bottom: 8px;
  font-weight: 500;
}

.table-wrapper {
  overflow: auto;
  border: 1px solid var(--border-color, #e5e7eb);
  border-radius: 8px;
}

.excel-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
  white-space: nowrap;
}

.excel-table th,
.excel-table td {
  padding: 8px 12px;
  border: 1px solid var(--border-color, #e5e7eb);
  text-align: left;
  max-width: 300px;
  overflow: hidden;
  text-overflow: ellipsis;
}

.excel-table thead th {
  background: #f0f5ff;
  color: #165DFF;
  font-weight: 600;
  position: sticky;
  top: 0;
  z-index: 1;
}

.excel-table tbody tr:hover {
  background: #f7f8fa;
}

.excel-table tbody tr:nth-child(even) {
  background: #fafbfc;
}

.excel-table tbody tr:nth-child(even):hover {
  background: #f0f5ff;
}

.row-num-header,
.row-num {
  width: 40px;
  text-align: center;
  color: var(--text-tertiary, #999);
  background: #fafbfc;
  font-size: 12px;
}
</style>
