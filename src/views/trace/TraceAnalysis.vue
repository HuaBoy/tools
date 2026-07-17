<script setup>
import { ref, reactive, onMounted, computed } from 'vue';
import { ElMessage } from 'element-plus';
import GlassCard from '@/components/GlassCard.vue';
import { useLogsStore } from '@/stores/logs';
import { useRouter, useRoute } from 'vue-router';
import CryptoJS from 'crypto-js';
import { requestWithAutoRelogin } from '@/utils/autoRelogin.js';
import { showLoginDialog, getLatestCredentials, performLogin } from '@/utils/platformLogin.js';
import { saveCredentials, updateLoginStatus } from '@/utils/loginStatus.js';
import { getTenantName, getTenantOptions } from '@/utils/tenant.js';
const logsStore = useLogsStore();
const router = useRouter();
const route = useRoute();

const TRACE_URL = '/api/blade-detonator-factory/detonatorProducttestData/page';

const formatDate = (date) => {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
};

const getDefaultDates = () => {
  const now = new Date();
  const oneYearAgo = new Date(now);
  oneYearAgo.setFullYear(oneYearAgo.getFullYear() - 1);
  return {
    startDate: formatDate(oneYearAgo),
    endDate: formatDate(now)
  };
};

const defaultDates = getDefaultDates();

const searchForm = reactive({
  tenantId: '',
  companyId: '',
  uploadStartDate: defaultDates.startDate,
  uploadEndDate: defaultDates.endDate,
  batchNo: '',
  shellCodeStart: '',
  uid: '',
  idlStart: '',
  explosionMark: '',
  type: '',
  moduleType: '',
  traceableCode: '',
  factoryId: '',
  startDate: '',
  endDate: '',
  explosionStartDate: '',
  explosionEndDate: '',
  current: 1,
  size: 10
});

const filterOptions = ref([
  { key: 'batchNo', label: '批次号' },
  { key: 'uploadStartDate', label: '上传开始日期' },
  { key: 'uploadEndDate', label: '上传结束日期' },
  { key: 'tenantId', label: '租户ID' },
  { key: 'companyId', label: '公司ID' },
  { key: 'shellCodeStart', label: '管壳码' },
  { key: 'uid', label: 'UID' },
  { key: 'idlStart', label: 'IDL开始' },
  { key: 'explosionMark', label: '起爆标志' },
  { key: 'type', label: '类型' },
  { key: 'moduleType', label: '模块类型' },
  { key: 'traceableCode', label: '溯源码' },
  { key: 'factoryId', label: '厂商ID' },
  { key: 'startDate', label: '开始日期' },
  { key: 'endDate', label: '结束日期' },
  { key: 'explosionStartDate', label: '起爆开始日期' },
  { key: 'explosionEndDate', label: '起爆结束日期' },
]);

const filterChecked = reactive({
  batchNo: true,
  uploadStartDate: true,
  uploadEndDate: true,
  tenantId: false,
  companyId: false,
  shellCodeStart: false,
  uid: false,
  idlStart: false,
  explosionMark: false,
  type: false,
  moduleType: false,
  traceableCode: false,
  factoryId: false,
  startDate: false,
  endDate: false,
  explosionStartDate: false,
  explosionEndDate: false,
});

const showFilterConfig = ref(false);
const isSearching = ref(false);

// ========== 多批次支持 ==========
// 解析批次号：支持中英文逗号、换行、空格分隔
const parseBatchNos = (text) => {
  if (!text) return [];
  return text
    .split(/[\n,，;；\s]+/)
    .map(b => b.trim())
    .filter(b => b.length > 0);
};

// 各批次查询结果：{ batchNo, loading, success, data, records, total, error }
const batchResults = ref([]);
const activeBatchIndex = ref(0);

// 解析当前批次号列表（computed）
const batchNoList = computed(() => parseBatchNos(searchForm.batchNo));

// 总记录数
const totalRecordCount = computed(() => {
  return batchResults.value.reduce((sum, r) => sum + (r.total || 0), 0);
});

// 当前激活的批次结果
const activeBatchResult = computed(() => {
  if (batchResults.value.length === 0) return null;
  return batchResults.value[activeBatchIndex.value] || null;
});
// ========== 多批次支持 END ==========

const showLoginModal = ref(false);
const selectedBatchNo = ref('');
const smartLoginForm = reactive({
  tenantId: '000000',
  username: '',
  password: ''
});
const isSmartLogin = ref(false);
const smartLoginSuccess = ref(false);
const smartLoginMessage = ref('');

const searchHistory = ref([]);
const showSearchHistory = ref(false);
const MAX_HISTORY_COUNT = 10;
const HISTORY_STORAGE_KEY = 'trace_analysis_search_history';

const loadSearchHistory = () => {
  try {
    const historyStr = localStorage.getItem(HISTORY_STORAGE_KEY);
    if (historyStr) {
      searchHistory.value = JSON.parse(historyStr);
    }
  } catch (e) {
    console.error('加载搜索历史失败:', e);
    searchHistory.value = [];
  }
};

const saveSearchHistory = () => {
  const historyItem = {
    id: Date.now(),
    timestamp: new Date().toLocaleString('zh-CN'),
    params: {}
  };

  Object.keys(searchForm).forEach(key => {
    if (searchForm[key] !== '' && searchForm[key] !== null && searchForm[key] !== undefined) {
      historyItem.params[key] = searchForm[key];
    }
  });

  if (Object.keys(historyItem.params).length > 0) {
    searchHistory.value = [historyItem, ...searchHistory.value.filter(h => {
      const same = Object.keys(h.params).length === Object.keys(historyItem.params).length &&
        Object.keys(h.params).every(k => h.params[k] === historyItem.params[k]);
      return !same;
    })].slice(0, MAX_HISTORY_COUNT);

    localStorage.setItem(HISTORY_STORAGE_KEY, JSON.stringify(searchHistory.value));
  }
};

const getFilterLabel = (key) => {
  const opt = filterOptions.value.find(o => o.key === key);
  return opt ? opt.label : key;
};

const applySearchHistory = (item) => {
  handleReset();
  Object.keys(item.params).forEach(key => {
    if (searchForm.hasOwnProperty(key)) {
      searchForm[key] = item.params[key];
    }
  });
  showSearchHistory.value = false;
  handleSearch();
};

const clearSearchHistory = () => {
  searchHistory.value = [];
  localStorage.removeItem(HISTORY_STORAGE_KEY);
  ElMessage.success('搜索历史已清除');
};

// 单个批次查询
const querySingleBatch = async (batchNo) => {
  const params = new URLSearchParams();
  Object.entries(searchForm).forEach(([key, value]) => {
    if (key === 'batchNo') {
      params.append(key, batchNo);
    } else if (value) {
      params.append(key, value);
    }
  });

  const { response, data } = await requestWithAutoRelogin('mp', {
    url: `${TRACE_URL}?${params.toString()}`,
    method: 'GET',
    headers: {
      'accept': 'application/json, text/plain, */*',
      'origin': 'https://mp.holyview.cn:9443',
      'referer': 'https://mp.holyview.cn:9443/'
    }
  });

  if (data && data.code === 200 && data.success && data.data && data.data.records) {
    return {
      success: true,
      data: data.data,
      records: data.data.records,
      total: data.data.total,
      current: data.data.current,
      size: data.data.size,
      pages: data.data.pages
    };
  } else if (response && response.status === 401) {
    return { success: false, error: '登录已失效，请重新登录' };
  } else {
    return { success: false, error: (data && data.msg) || '未查询到数据' };
  }
};

// 多批次查询主函数
const handleSearch = async () => {
  const batchNos = batchNoList.value;

  if (batchNos.length === 0) {
    ElMessage.warning('请输入至少一个批次号');
    return;
  }

  // 检查登录状态：未登录则弹窗或跳转
  const token = localStorage.getItem('mp_token');
  if (!token) {
    const latest = getLatestCredentials('mp');
    if (latest) {
      // 有最近凭据，弹出登录对话框
      const result = await showLoginDialog('mp');
      if (!result || !result.success) {
        // 用户取消登录
        if (result && result.reason === 'no_credentials') {
          window.location.href = '#/auth/third-party';
        }
        return;
      }
    } else {
      // 没有最近凭据，跳转到三方授权页面
      ElMessage.warning('未找到登录凭据，请先在授权页面登录');
      setTimeout(() => {
        window.location.href = '#/auth/third-party';
      }, 500);
      return;
    }
  }

  isSearching.value = true;
  batchResults.value = batchNos.map(batchNo => ({
    batchNo,
    loading: true,
    success: false,
    data: null,
    records: [],
    total: 0,
    current: 1,
    size: 10,
    pages: 0,
    error: null
  }));
  activeBatchIndex.value = 0;

  // 顺序查询（避免并发过多）
  for (let i = 0; i < batchNos.length; i++) {
    const batchNo = batchNos[i];
    try {
      const result = await querySingleBatch(batchNo);
      batchResults.value[i] = {
        batchNo,
        loading: false,
        ...result
      };
      if (result.success) {
        logsStore.addLog('查询', '数据追溯', `批次号: ${batchNo}, 共${result.total}条`);
      }
    } catch (error) {
      batchResults.value[i] = {
        batchNo,
        loading: false,
        success: false,
        error: error.message || '网络错误',
        data: null,
        records: [],
        total: 0
      };
    }
  }

  isSearching.value = false;
  const successCount = batchResults.value.filter(r => r.success).length;
  const totalCount = batchResults.value.reduce((sum, r) => sum + (r.total || 0), 0);

  if (batchNos.length === 1) {
    const first = batchResults.value[0];
    if (first.success) {
      ElMessage.success(`查询成功，共 ${first.total} 条记录`);
    } else {
      ElMessage.warning(first.error || '未查询到数据');
    }
  } else {
    ElMessage.success(`完成 ${successCount}/${batchNos.length} 个批次查询，共 ${totalCount} 条记录`);
  }
  saveSearchHistory();
};

const handleReset = () => {
  const dates = getDefaultDates();
  Object.keys(searchForm).forEach(key => {
    if (key === 'current') {
      searchForm[key] = 1;
    } else if (key === 'size') {
      searchForm[key] = 10;
    } else if (key === 'uploadStartDate') {
      searchForm[key] = dates.startDate;
    } else if (key === 'uploadEndDate') {
      searchForm[key] = dates.endDate;
    } else {
      searchForm[key] = '';
    }
  });
  batchResults.value = [];
  activeBatchIndex.value = 0;
};

const toggleFilter = (key) => {
  if (filterChecked.hasOwnProperty(key)) {
    filterChecked[key] = !filterChecked[key];
  }
};

const isFilterVisible = (key) => {
  return filterChecked[key] === true;
};

const handleBatchNoClick = (batchNo) => {
  router.push({
    path: '/trace/factory-data',
    query: {
      batchNo: batchNo,
      autoQuery: 'true'  // 自动触发查询
    }
  });
};

// 低压ID点击 → 跳转到智能制造系统页面并自动带入 batchNo + idHex 筛选条件
const handleLowVoltageIdClick = (batchNo, idHex) => {
  if (!idHex || idHex === '-') return;
  router.push({
    path: '/trace/factory-data',
    query: {
      batchNo: batchNo || '',
      idHex: idHex,
      autoQuery: 'true'
    }
  });
};

const handleSmartLogin = async () => {
  if (!smartLoginForm.username || !smartLoginForm.password) {
    smartLoginMessage.value = '请输入用户名和密码';
    return;
  }

  isSmartLogin.value = true;
  smartLoginMessage.value = '';

  try {
    // 使用统一 performLogin，自动同步全局状态（Header、token、凭据、三方授权等）
    const result = await performLogin('smart', smartLoginForm.username, smartLoginForm.password);

    if (result.success) {
      smartLoginSuccess.value = true;
      smartLoginMessage.value = '登录成功！即将跳转...';

      setTimeout(() => {
        showLoginModal.value = false;
        smartLoginForm.username = '';
        smartLoginForm.password = '';
        ElMessage.success('登录智能制造平台成功');
      }, 1500);
    } else {
      smartLoginMessage.value = result.message || '登录失败，请检查用户名和密码';
    }
  } catch (error) {
    smartLoginMessage.value = '登录请求失败：' + error.message;
  } finally {
    isSmartLogin.value = false;
  }
};

const closeLoginModal = () => {
  showLoginModal.value = false;
  smartLoginForm.username = '';
  smartLoginForm.password = '';
  smartLoginMessage.value = '';
  smartLoginSuccess.value = false;
};

// 翻页：使用激活的批次
const handlePageChange = (page) => {
  const cur = activeBatchResult.value;
  if (!cur || page < 1 || page > cur.pages) return;
  // 简单处理：翻页时重新查询当前批次
  const newBatchNos = [...batchNoList.value];
  newBatchNos[activeBatchIndex.value] = cur.batchNo;
  searchForm.batchNo = newBatchNos.join(',');
  searchForm.current = page;
  // 单独查询当前批次
  isSearching.value = true;
  cur.loading = true;
  querySingleBatch(cur.batchNo).then(result => {
    Object.assign(cur, { loading: false, ...result });
    isSearching.value = false;
  });
};

// 导出：单个批次（下载）
const exportBatch = async (batchResult) => {
  if (!batchResult || !batchResult.success || batchResult.total === 0) {
    ElMessage.warning('该批次无数据可导出');
    return;
  }

  try {
    const params = new URLSearchParams();
    params.append('uploadStartDate', searchForm.uploadStartDate || '');
    params.append('date', '');
    params.append('date1', '');
    params.append('batchNo', batchResult.batchNo || '');
    params.append('uploadEndDate', searchForm.uploadEndDate || '');

    const EXPORT_URL = 'https://mp.holyview.cn:9443/api/blade-detonator-factory/detonatorProducttestData/export';
    const fullUrl = `${EXPORT_URL}?${params.toString()}`;

    const { response } = await requestWithAutoRelogin('mp', {
      url: fullUrl,
      method: 'GET',
      headers: {
        'accept': 'application/json, text/plain, */*',
        'origin': 'https://mp.holyview.cn:9443',
        'referer': 'https://mp.holyview.cn:9443/'
      }
    });

    if (response.ok) {
      const blob = await response.blob();
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `批次数据追溯_${batchResult.batchNo}_${new Date().toISOString().split('T')[0]}.xlsx`;
      a.click();
      URL.revokeObjectURL(url);
      logsStore.addLog('导出', '数据追溯', `批次号: ${batchResult.batchNo}`);
      return true;
    } else {
      ElMessage.error('导出失败');
      return false;
    }
  } catch (error) {
    ElMessage.error('导出请求失败：' + error.message);
    return false;
  }
};

// 批量顺序导出
const isBatchExporting = ref(false);
const batchExportProgress = ref({ current: 0, total: 0, currentBatch: '' });

const handleBatchExport = async () => {
  const successBatches = batchResults.value.filter(r => r.success && r.total > 0);
  if (successBatches.length === 0) {
    ElMessage.warning('没有可导出的批次数据');
    return;
  }

  isBatchExporting.value = true;
  batchExportProgress.value = {
    current: 0,
    total: successBatches.length,
    currentBatch: ''
  };

  let exportedCount = 0;
  for (let i = 0; i < successBatches.length; i++) {
    const r = successBatches[i];
    batchExportProgress.value.current = i + 1;
    batchExportProgress.value.currentBatch = r.batchNo;
    const ok = await exportBatch(r);
    if (ok) exportedCount++;
  }

  isBatchExporting.value = false;
  ElMessage.success(`批量导出完成，成功 ${exportedCount}/${successBatches.length}`);
  logsStore.addLog('批量导出', '数据追溯', `共${successBatches.length}个批次，成功${exportedCount}`);
};

onMounted(() => {
  loadSearchHistory();

  // 处理 URL 参数（从首页跳转过来）
  const queryBatchNo = route.query.batchNo;
  const autoQuery = route.query.autoQuery === 'true' || route.query.autoQuery === true;

  if (queryBatchNo) {
    searchForm.batchNo = String(queryBatchNo);
    logsStore.addLog('参数填充', '数据追溯', `批次号: ${queryBatchNo}`);

    if (autoQuery) {
      // 延迟执行确保页面完全渲染
      setTimeout(() => {
        handleQuery();
      }, 500);
    }
  }
});
</script>

<template>
  <div class="trace-analysis">
    <GlassCard title="AI数据追溯">
      <div class="search-section">
        <div class="filter-config-header">
          <button 
            class="config-btn"
            @click="showFilterConfig = !showFilterConfig"
          >
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <line x1="4" y1="4" x2="20" y2="20" />
              <line x1="20" y1="4" x2="4" y2="20" />
              <line x1="12" y1="4" x2="12" y2="20" />
              <line x1="4" y1="12" x2="20" y2="12" />
            </svg>
            <span>{{ showFilterConfig ? '收起配置' : '筛选配置' }}</span>
          </button>
          
          <button 
            class="config-btn history-btn"
            :class="{ active: showSearchHistory }"
            @click="showSearchHistory = !showSearchHistory"
          >
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8" />
              <path d="M3 3v5h5" />
            </svg>
            <span>搜索历史</span>
            <span v-if="searchHistory.length > 0" class="history-count">{{ searchHistory.length }}</span>
          </button>
        </div>
        
        <div v-if="showSearchHistory" class="search-history-panel">
          <div class="search-history-header">
            <span>搜索历史记录</span>
            <button class="clear-history-btn" @click="clearSearchHistory">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <polyline points="3 6 5 6 21 6" />
                <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
              </svg>
              <span>清空</span>
            </button>
          </div>
          <div v-if="searchHistory.length === 0" class="search-history-empty">
            <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
              <path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8" />
              <path d="M3 3v5h5" />
            </svg>
            <span>暂无搜索历史</span>
          </div>
          <div v-else class="search-history-list">
            <div 
              v-for="item in searchHistory" 
              :key="item.id" 
              class="search-history-item"
              @click="applySearchHistory(item)"
            >
              <div class="search-history-time">{{ item.timestamp }}</div>
              <div class="search-history-params">
                <span 
                  v-for="(value, key) in item.params" 
                  :key="key"
                  class="search-param-tag"
                >
                  {{ getFilterLabel(key) }}: {{ value }}
                </span>
              </div>
            </div>
          </div>
        </div>
        
        <div v-if="showFilterConfig" class="filter-config-panel">
          <div class="filter-options">
            <label 
              v-for="option in filterOptions" 
              :key="option.key"
              class="filter-option"
            >
              <input 
                type="checkbox" 
                :checked="filterChecked[option.key]"
                @change="toggleFilter(option.key)"
              />
              <span>{{ option.label }}</span>
            </label>
          </div>
        </div>
        
        <div class="search-grid">
          <div v-if="isFilterVisible('batchNo')" class="form-group batch-no-group">
            <label>
              批次号
              <span v-if="batchNoList.length > 0" class="batch-count">{{ batchNoList.length }} 个</span>
            </label>
            <textarea
              v-model="searchForm.batchNo"
              class="form-input batch-textarea"
              placeholder="支持多个批次号，换行、逗号、空格分隔。例如：&#10;BATCH001&#10;BATCH002,BATCH003"
              rows="3"
            ></textarea>
          </div>
          
          <div v-if="isFilterVisible('uploadStartDate')" class="form-group">
            <label>上传开始日期</label>
            <input 
              v-model="searchForm.uploadStartDate"
              type="date"
              class="form-input"
            />
          </div>
          
          <div v-if="isFilterVisible('uploadEndDate')" class="form-group">
            <label>上传结束日期</label>
            <input 
              v-model="searchForm.uploadEndDate"
              type="date"
              class="form-input"
            />
          </div>
          
          <div v-if="isFilterVisible('tenantId')" class="form-group">
            <label>管厂/租户</label>
            <select
              v-model="searchForm.tenantId"
              class="form-input"
              @change="searchForm.tenantName = getTenantName(searchForm.tenantId)"
            >
              <option value="">全部</option>
              <option v-for="opt in tenantOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
            </select>
            <div v-if="searchForm.tenantId" class="tenant-name-hint">已选：{{ searchForm.tenantName }}</div>
          </div>
          
          <div v-if="isFilterVisible('companyId')" class="form-group">
            <label>公司ID</label>
            <input 
              v-model="searchForm.companyId"
              type="text"
              class="form-input"
              placeholder="companyId"
            />
          </div>
          
          <div v-if="isFilterVisible('shellCodeStart')" class="form-group">
            <label>管壳码</label>
            <input 
              v-model="searchForm.shellCodeStart"
              type="text"
              class="form-input"
              placeholder="请输入管壳码"
            />
          </div>
          
          <div v-if="isFilterVisible('uid')" class="form-group">
            <label>UID</label>
            <input 
              v-model="searchForm.uid"
              type="text"
              class="form-input"
              placeholder="请输入UID"
            />
          </div>
          
          <div v-if="isFilterVisible('idlStart')" class="form-group">
            <label>IDL开始</label>
            <input 
              v-model="searchForm.idlStart"
              type="text"
              class="form-input"
              placeholder="idlStart"
            />
          </div>
          
          <div v-if="isFilterVisible('explosionMark')" class="form-group">
            <label>起爆标志</label>
            <select v-model="searchForm.explosionMark" class="form-input">
              <option value="">全部</option>
              <option value="0">未爆破</option>
              <option value="1">已爆破</option>
            </select>
          </div>
          
          <div v-if="isFilterVisible('type')" class="form-group">
            <label>类型</label>
            <input 
              v-model="searchForm.type"
              type="text"
              class="form-input"
              placeholder="type"
            />
          </div>
          
          <div v-if="isFilterVisible('moduleType')" class="form-group">
            <label>模块类型</label>
            <input 
              v-model="searchForm.moduleType"
              type="text"
              class="form-input"
              placeholder="moduleType"
            />
          </div>
          
          <div v-if="isFilterVisible('traceableCode')" class="form-group">
            <label>溯源码</label>
            <input 
              v-model="searchForm.traceableCode"
              type="text"
              class="form-input"
              placeholder="traceableCode"
            />
          </div>
          
          <div v-if="isFilterVisible('factoryId')" class="form-group">
            <label>厂商ID</label>
            <input 
              v-model="searchForm.factoryId"
              type="text"
              class="form-input"
              placeholder="factoryId"
            />
          </div>
          
          <div v-if="isFilterVisible('startDate')" class="form-group">
            <label>开始日期</label>
            <input 
              v-model="searchForm.startDate"
              type="date"
              class="form-input"
            />
          </div>
          
          <div v-if="isFilterVisible('endDate')" class="form-group">
            <label>结束日期</label>
            <input 
              v-model="searchForm.endDate"
              type="date"
              class="form-input"
            />
          </div>
          
          <div v-if="isFilterVisible('explosionStartDate')" class="form-group">
            <label>起爆开始日期</label>
            <input 
              v-model="searchForm.explosionStartDate"
              type="date"
              class="form-input"
            />
          </div>
          
          <div v-if="isFilterVisible('explosionEndDate')" class="form-group">
            <label>起爆结束日期</label>
            <input 
              v-model="searchForm.explosionEndDate"
              type="date"
              class="form-input"
            />
          </div>
        </div>
        
        <div class="form-actions">
          <button 
            class="action-btn primary"
            :disabled="isSearching"
            @click="handleSearch"
          >
            <svg v-if="!isSearching" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="11" cy="11" r="8" />
              <line x1="21" y1="21" x2="16.65" y2="16.65" />
            </svg>
            <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="spin">
              <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
            </svg>
            <span>{{ isSearching ? '查询中...' : '查询数据' }}</span>
          </button>
          
          <button 
            class="action-btn secondary"
            @click="handleReset"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8" />
              <path d="M3 3v5h5" />
            </svg>
            <span>重置</span>
          </button>
          
          <button
            class="action-btn export"
            :disabled="!activeBatchResult?.success || activeBatchResult.total === 0"
            @click="exportBatch(activeBatchResult)"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
              <polyline points="7 10 12 15 17 10" />
              <line x1="12" y1="15" x2="12" y2="3" />
            </svg>
            <span>导出当前批次</span>
          </button>

          <button
            v-if="batchResults.length > 1"
            class="action-btn batch-export"
            :disabled="isBatchExporting || batchResults.filter(r => r.success && r.total > 0).length === 0"
            @click="handleBatchExport"
          >
            <svg v-if="!isBatchExporting" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
              <polyline points="17 8 12 3 7 8" />
              <line x1="12" y1="3" x2="12" y2="15" />
            </svg>
            <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="spin">
              <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
            </svg>
            <span>{{ isBatchExporting ? `导出中 ${batchExportProgress.current}/${batchExportProgress.total}` : '批量导出' }}</span>
          </button>
        </div>
      </div>

      <!-- 批量导出加载层 -->
      <Teleport to="body">
        <div v-if="isBatchExporting" class="export-loading-overlay">
          <div class="export-loading-content">
            <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="#165DFF" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="spin">
              <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
            </svg>
            <div class="export-loading-text">
              <div class="export-loading-title">正在批量导出数据</div>
              <div class="export-loading-detail">
                进度：{{ batchExportProgress.current }} / {{ batchExportProgress.total }}
                <span v-if="batchExportProgress.currentBatch"> · 当前批次：{{ batchExportProgress.currentBatch }}</span>
              </div>
              <div class="export-progress-bar">
                <div class="export-progress-fill" :style="{ width: (batchExportProgress.current / batchExportProgress.total * 100) + '%' }"></div>
              </div>
            </div>
          </div>
        </div>
      </Teleport>

      <!-- 多批次 Tab 切换结果区 -->
      <div v-if="batchResults.length > 0" class="batch-tabs-panel">
        <div class="batch-tabs-header">
          <div class="batch-tabs-scroll">
            <div class="batch-tabs">
              <div
                v-for="(r, idx) in batchResults"
                :key="r.batchNo + '-' + idx"
                class="batch-tab"
                :class="{ active: activeBatchIndex === idx }"
                @click="activeBatchIndex = idx"
              >
                <span class="batch-tab-name">{{ r.batchNo }}</span>
                <span v-if="r.loading" class="batch-tab-loading spin">
                  <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
                  </svg>
                </span>
                <span v-else-if="r.success" class="batch-tab-count">{{ r.total }}</span>
                <span v-else class="batch-tab-error" :title="r.error">!</span>
              </div>
            </div>
          </div>
          <div v-if="batchResults.length > 1" class="batch-total-info">
            总计 {{ totalRecordCount }} 条记录
          </div>
        </div>

        <div v-if="activeBatchResult" class="batch-tab-content">
          <div v-if="activeBatchResult.loading" class="tab-loading">
            <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="#165DFF" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="spin">
              <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
            </svg>
            <span>正在查询批次 {{ activeBatchResult.batchNo }} ...</span>
          </div>
          <div v-else-if="activeBatchResult.success && activeBatchResult.records && activeBatchResult.records.length > 0" class="records-panel">
            <div class="records-header">
              <span>数据列表 - {{ activeBatchResult.batchNo }}</span>
              <span class="pagination-info">共 {{ activeBatchResult.total }} 条 | 第 {{ activeBatchResult.current }} 页 | 每页 {{ activeBatchResult.size }} 条</span>
            </div>
            <div class="table-container">
              <table class="trace-table">
                <thead>
                  <tr>
                    <th>序号</th>
                    <th>批次号</th>
                    <th>低压ID</th>
                    <th>管壳码</th>
                    <th>UID</th>
                    <th>爆破状态</th>
                    <th>三码绑定时间</th>
                    <th>生产数据上传时间</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(record, index) in activeBatchResult.records" :key="record.id">
                    <td>{{ (activeBatchResult.current - 1) * activeBatchResult.size + index + 1 }}</td>
                    <td class="batch-no clickable" @click="handleBatchNoClick(record.batchNo)">{{ record.batchNo || '-' }}</td>
                    <td class="id-hex clickable" @click="handleLowVoltageIdClick(record.batchNo, record.idHex)">{{ record.idHex || '-' }}</td>
                    <td>{{ record.shellCode || '-' }}</td>
                    <td class="uid">{{ record.uid || '-' }}</td>
                    <td>
                      <span
                        class="explosion-status"
                        :class="record.explosionMark === '1' ? 'exploded' : 'not-exploded'"
                      >
                        {{ record.explosionMark === '1' ? '已爆破' : '未爆破' }}
                      </span>
                    </td>
                    <td>{{ record.checkTime || '-' }}</td>
                    <td>{{ record.createTime || '-' }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
            <div v-if="activeBatchResult.pages > 1" class="pagination-container">
              <div class="pagination">
                <button
                  class="page-btn"
                  :disabled="activeBatchResult.current <= 1"
                  @click="handlePageChange(activeBatchResult.current - 1)"
                >
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <polyline points="15 18 9 12 15 6" />
                  </svg>
                </button>
                <span class="page-info">{{ activeBatchResult.current }} / {{ activeBatchResult.pages }}</span>
                <button
                  class="page-btn"
                  :disabled="activeBatchResult.current >= activeBatchResult.pages"
                  @click="handlePageChange(activeBatchResult.current + 1)"
                >
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <polyline points="9 18 15 12 9 6" />
                  </svg>
                </button>
              </div>
            </div>
          </div>
          <div v-else-if="activeBatchResult.success && (!activeBatchResult.records || activeBatchResult.records.length === 0)" class="tab-empty">
            <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="#94A3B8" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="12" r="10" />
              <line x1="12" y1="8" x2="12" y2="12" />
              <line x1="12" y1="16" x2="12.01" y2="16" />
            </svg>
            <span>批次 {{ activeBatchResult.batchNo }} 未查询到数据</span>
          </div>
          <div v-else class="tab-error">
            <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="#F53F3F" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="12" r="10" />
              <line x1="15" y1="9" x2="9" y2="15" />
              <line x1="9" y1="9" x2="15" y2="15" />
            </svg>
            <span>批次 {{ activeBatchResult.batchNo }} 查询失败：{{ activeBatchResult.error }}</span>
          </div>
        </div>
      </div>
      
      <Teleport to="body">
        <div v-if="showLoginModal" class="modal-overlay" @click.self="closeLoginModal">
          <div class="modal-content">
            <div class="modal-header">
              <h3>登录智能制造平台</h3>
              <button class="close-btn" @click="closeLoginModal">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <line x1="18" y1="6" x2="6" y2="18" />
                  <line x1="6" y1="6" x2="18" y2="18" />
                </svg>
              </button>
            </div>
            
            <div class="modal-body">
              <p class="batch-info">批次号：<span>{{ selectedBatchNo }}</span></p>
              
              <div class="login-form">
                <div class="form-group">
                  <label>用户名</label>
                  <input 
                    v-model="smartLoginForm.username"
                    type="text"
                    class="form-input"
                    placeholder="请输入用户名"
                  />
                </div>
                
                <div class="form-group">
                  <label>密码</label>
                  <input 
                    v-model="smartLoginForm.password"
                    type="password"
                    class="form-input"
                    placeholder="请输入密码"
                  />
                </div>
                
                <div v-if="smartLoginMessage" class="message" :class="smartLoginSuccess ? 'success' : 'error'">
                  {{ smartLoginMessage }}
                </div>
              </div>
            </div>
            
            <div class="modal-footer">
              <button class="modal-btn secondary" @click="closeLoginModal">取消</button>
              <button 
                class="modal-btn primary"
                :disabled="isSmartLogin"
                @click="handleSmartLogin"
              >
                <svg v-if="isSmartLogin" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="spin">
                  <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
                </svg>
                <span>{{ isSmartLogin ? '登录中...' : '登录' }}</span>
              </button>
            </div>
          </div>
        </div>
      </Teleport>
    </GlassCard>
  </div>
</template>

<style scoped>
.trace-analysis {
  max-width: 100%;
}

.search-section {
  margin-bottom: 20px;
}

.filter-config-header {
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.config-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background: rgba(22, 93, 255, 0.1);
  border: 1px solid rgba(22, 93, 255, 0.3);
  border-radius: 6px;
  color: #165DFF;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s;
  
  &:hover {
    background: rgba(22, 93, 255, 0.2);
  }
}

.filter-config-panel {
  background: rgba(15, 23, 42, 0.05);
  border-radius: 8px;
  padding: 12px 16px;
  margin-bottom: 16px;
}

.filter-options {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
}

.filter-option {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #64748B;
  cursor: pointer;
  
  input[type="checkbox"] {
    width: 14px;
    height: 14px;
    cursor: pointer;
  }
}

.search-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-group label {
  font-size: 12px;
  color: #64748B;
}

.tenant-name-hint {
  font-size: 11px;
  color: #165DFF;
  padding: 2px 6px;
  background: rgba(22, 93, 255, 0.06);
  border-radius: 4px;
}

.form-input {
  padding: 8px 10px;
  border: 1px solid #E2E8F0;
  border-radius: 6px;
  font-size: 13px;
  background: white;
  
  &:focus {
    outline: none;
    border-color: #165DFF;
  }
}

.form-actions {
  display: flex;
  gap: 12px;
  margin-top: 20px;
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 20px;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
  
  &.primary {
    background: linear-gradient(135deg, #165DFF 0%, #0F4CD0 100%);
    color: #FFFFFF;
    
    &:hover:not(:disabled) {
      transform: translateY(-2px);
      box-shadow: 0 4px 12px rgba(22, 93, 255, 0.3);
    }
  }
  
  &.secondary {
    background: rgba(100, 116, 139, 0.2);
    border: 1px solid rgba(100, 116, 139, 0.3);
    color: #94A3B8;
    
    &:hover:not(:disabled) {
      background: rgba(100, 116, 139, 0.3);
    }
  }
  
  &.export {
    background: rgba(0, 180, 42, 0.1);
    border: 1px solid rgba(0, 180, 42, 0.3);
    color: #00B42A;
    
    &:hover:not(:disabled) {
      background: rgba(0, 180, 42, 0.2);
    }
  }
  
  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
}

.result-panel {
  margin-top: 20px;
  padding: 16px;
  border-radius: 8px;
  
  &.success {
    background: rgba(0, 180, 42, 0.1);
    border: 1px solid rgba(0, 180, 42, 0.3);
  }
  
  &.error {
    background: rgba(245, 63, 63, 0.1);
    border: 1px solid rgba(245, 63, 63, 0.3);
  }
}

.result-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
  
  svg {
    flex-shrink: 0;
  }
  
  span {
    font-size: 14px;
    font-weight: 600;
    
    .success & {
      color: #00B42A;
    }
    
    .error & {
      color: #F53F3F;
    }
  }
}

.result-content {
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 12px;
  color: #64748B;
  white-space: pre-wrap;
  word-break: break-all;
  margin: 0;
  max-height: 300px;
  overflow-y: auto;
}

.records-panel {
  margin-top: 20px;
}

.records-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 13px;
  font-weight: 600;
  color: #1E293B;
  margin-bottom: 12px;
}

.pagination-info {
  font-size: 12px;
  color: #94A3B8;
  font-weight: normal;
}

.pagination-container {
  margin-top: 16px;
  display: flex;
  justify-content: center;
}

.pagination {
  display: flex;
  align-items: center;
  gap: 8px;
}

.page-btn {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: white;
  border: 1px solid #E2E8F0;
  border-radius: 6px;
  color: #64748B;
  cursor: pointer;
  transition: all 0.2s;
  
  &:hover:not(:disabled) {
    background: rgba(22, 93, 255, 0.1);
    border-color: #165DFF;
    color: #165DFF;
  }
  
  &:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }
}

.page-numbers {
  display: flex;
  align-items: center;
  gap: 4px;
}

.page-num-btn {
  min-width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: white;
  border: 1px solid #E2E8F0;
  border-radius: 6px;
  color: #64748B;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
  
  &:hover:not(.active) {
    border-color: #165DFF;
    color: #165DFF;
  }
  
  &.active {
    background: #165DFF;
    border-color: #165DFF;
    color: white;
  }
}

.ellipsis {
  padding: 0 4px;
  color: #94A3B8;
  font-size: 13px;
}

.page-size-selector {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-left: 16px;
  font-size: 12px;
  color: #64748B;
}

.page-size-input {
  padding: 4px 8px;
  border: 1px solid #E2E8F0;
  border-radius: 4px;
  font-size: 12px;
  background: white;
  color: #1E293B;
  cursor: pointer;
  
  &:focus {
    outline: none;
    border-color: #165DFF;
  }
}

.table-container {
  overflow-x: auto;
  background: #FAFAFA;
  border-radius: 8px;
  border: 1px solid #E2E8F0;
}

.trace-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.trace-table thead {
  background: rgba(22, 93, 255, 0.1);
}

.trace-table th {
  padding: 12px 16px;
  text-align: left;
  font-weight: 600;
  color: #1E293B;
  border-bottom: 1px solid #E2E8F0;
  white-space: nowrap;
}

.trace-table td {
  padding: 10px 16px;
  border-bottom: 1px solid rgba(224, 232, 240, 0.5);
  color: #64748B;
}

.trace-table tbody tr:hover {
  background: rgba(22, 93, 255, 0.05);
}

.batch-no {
  font-family: 'Consolas', 'Monaco', monospace;
  font-weight: 500;
  color: #165DFF;
}

.id-hex,
.uid {
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 12px;
  color: #1E293B;
}

.explosion-status {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
  
  &.exploded {
    background: rgba(0, 180, 42, 0.2);
    color: #00B42A;
  }
  
  &.not-exploded {
    background: rgba(245, 63, 63, 0.2);
    color: #F53F3F;
  }
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

@media screen and (max-width: 1200px) {
  .search-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}

@media screen and (max-width: 768px) {
  .search-grid {
    grid-template-columns: repeat(2, 1fr);
  }
  
  .form-actions {
    flex-direction: column;
  }
  
  .action-btn {
    width: 100%;
    justify-content: center;
  }
  
  .trace-table {
    font-size: 12px;
  }
  
  .trace-table th,
  .trace-table td {
    padding: 8px 10px;
  }
  
  .filter-options {
    gap: 12px;
  }
}

.batch-no.clickable,
.id-hex.clickable {
  color: #165DFF;
  cursor: pointer;
  text-decoration: underline;
  
  &:hover {
    color: #0D47A1;
    text-decoration: none;
  }
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: white;
  border-radius: 12px;
  width: 90%;
  max-width: 420px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid #E2E8F0;
  
  h3 {
    margin: 0;
    font-size: 16px;
    font-weight: 600;
    color: #1E293B;
  }
}

.close-btn {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: none;
  border-radius: 6px;
  color: #94A3B8;
  cursor: pointer;
  transition: all 0.2s;
  
  &:hover {
    background: rgba(0, 0, 0, 0.05);
    color: #64748B;
  }
}

.modal-body {
  padding: 20px;
}

.batch-info {
  margin: 0 0 16px 0;
  font-size: 13px;
  color: #64748B;
  
  span {
    font-weight: 600;
    color: #165DFF;
  }
}

.login-form {
  .form-group {
    margin-bottom: 12px;
  }
  
  .form-input {
    width: 100%;
  }
}

.message {
  margin-top: 12px;
  padding: 10px 12px;
  border-radius: 6px;
  font-size: 12px;
  
  &.success {
    background: rgba(34, 197, 94, 0.1);
    color: #16A34A;
  }
  
  &.error {
    background: rgba(239, 68, 68, 0.1);
    color: #DC2626;
  }
}

.modal-footer {
  display: flex;
  gap: 12px;
  padding: 16px 20px;
  border-top: 1px solid #E2E8F0;
  justify-content: flex-end;
}

.modal-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  border: 1px solid transparent;
  
  &.primary {
    background: #165DFF;
    color: white;
    
    &:hover:not(:disabled) {
      background: #0D47A1;
    }
    
    &:disabled {
      opacity: 0.6;
      cursor: not-allowed;
    }
  }
  
  &.secondary {
    background: #F1F5F9;
    color: #64748B;
    
    &:hover {
      background: #E2E8F0;
    }
  }
}

@media screen and (max-width: 480px) {
  .search-grid {
    grid-template-columns: 1fr;
  }
  
  .filter-options {
    flex-direction: column;
    gap: 8px;
  }
}

.history-btn {
  &.active {
    background: rgba(22, 93, 255, 0.2);
    border-color: #165DFF;
  }
  
  .history-count {
    background: #165DFF;
    color: white;
    font-size: 10px;
    padding: 1px 6px;
    border-radius: 10px;
    min-width: 18px;
    text-align: center;
  }
}

.search-history-panel {
  background: rgba(15, 23, 42, 0.03);
  border: 1px solid #E2E8F0;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 16px;
}

.search-history-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  font-size: 13px;
  font-weight: 600;
  color: #1E293B;
}

.clear-history-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  background: rgba(245, 63, 63, 0.1);
  border: 1px solid rgba(245, 63, 63, 0.3);
  border-radius: 4px;
  color: #F53F3F;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s;
  
  &:hover {
    background: rgba(245, 63, 63, 0.2);
  }
}

.search-history-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 30px 0;
  color: #94A3B8;
  font-size: 13px;
  gap: 8px;
}

.search-history-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

/* ===== 多批次支持样式 ===== */
.batch-no-group {
  grid-column: 1 / -1;
}

.batch-no-group label {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.batch-count {
  background: rgba(22, 93, 255, 0.1);
  color: #165DFF;
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 10px;
  font-weight: 500;
}

.batch-textarea {
  resize: vertical;
  min-height: 70px;
  font-family: 'Consolas', 'Monaco', monospace;
  line-height: 1.5;
}

.batch-tabs-panel {
  margin-top: 20px;
}

.batch-tabs-header {
  display: flex;
  align-items: flex-end;
  border-bottom: 1px solid #E2E8F0;
  margin-bottom: 16px;
  gap: 12px;
  width: 100%;
  min-width: 0;
}

.batch-tabs-scroll {
  flex: 1 1 auto;
  min-width: 0;
  overflow-x: auto;
  overflow-y: hidden;
  scrollbar-width: thin;

  &::-webkit-scrollbar {
    height: 4px;
  }
  &::-webkit-scrollbar-thumb {
    background: #CBD5E1;
    border-radius: 2px;
  }
}

.batch-tabs {
  display: flex;
  gap: 4px;
  min-width: max-content;
  border-bottom: none;
  margin-bottom: 0;
  padding-bottom: 0;
}

.batch-tab {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 16px;
  background: transparent;
  border: none;
  border-bottom: 2px solid transparent;
  color: #64748B;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
  margin-bottom: -1px; /* 让 active 状态的蓝色下边框压住 header 的灰色下边框 */
}

.batch-tab:hover {
  color: #165DFF;
}

.batch-tab.active {
  color: #165DFF;
  border-bottom-color: #165DFF;
  font-weight: 600;
}

.batch-total-info {
  flex: 0 0 auto;
  padding: 10px 16px;
  font-size: 13px;
  color: #165DFF;
  font-weight: 600;
  background: rgba(22, 93, 255, 0.06);
  border-radius: 8px 8px 0 0;
  white-space: nowrap;
  border-bottom: 2px solid #165DFF;
  margin-bottom: -1px;
}

.batch-tab-name {
  font-family: 'Consolas', 'Monaco', monospace;
}

.batch-tab-count {
  background: rgba(0, 180, 42, 0.1);
  color: #00B42A;
  font-size: 11px;
  padding: 1px 6px;
  border-radius: 8px;
  min-width: 18px;
  text-align: center;
}

.batch-tab-loading {
  color: #165DFF;
  display: inline-flex;
}

.batch-tab-error {
  background: rgba(245, 63, 63, 0.1);
  color: #F53F3F;
  font-size: 11px;
  padding: 1px 6px;
  border-radius: 50%;
  width: 16px;
  height: 16px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
}

.batch-tab-content {
  min-height: 200px;
}

.tab-loading,
.tab-empty,
.tab-error {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 0;
  gap: 12px;
  font-size: 14px;
}

.tab-loading {
  color: #165DFF;
}

.tab-empty {
  color: #94A3B8;
}

.tab-error {
  color: #F53F3F;
}

.page-info {
  font-size: 13px;
  color: #64748B;
  padding: 0 12px;
}

/* 批量导出按钮 */
.action-btn.batch-export {
  background: rgba(99, 102, 241, 0.1);
  border: 1px solid rgba(99, 102, 241, 0.3);
  color: #6366F1;

  &:hover:not(:disabled) {
    background: rgba(99, 102, 241, 0.2);
  }
}

/* 批量导出加载层 */
.export-loading-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(15, 23, 42, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  backdrop-filter: blur(2px);
}

.export-loading-content {
  background: white;
  border-radius: 16px;
  padding: 32px 48px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  min-width: 320px;
}

.export-loading-text {
  text-align: center;
  width: 100%;
}

.export-loading-title {
  font-size: 16px;
  font-weight: 600;
  color: #1E293B;
  margin-bottom: 8px;
}

.export-loading-detail {
  font-size: 13px;
  color: #64748B;
  margin-bottom: 12px;
}

.export-progress-bar {
  width: 100%;
  height: 6px;
  background: rgba(22, 93, 255, 0.1);
  border-radius: 3px;
  overflow: hidden;
}

.export-progress-fill {
  height: 100%;
  background: linear-gradient(90deg, #165DFF 0%, #6366F1 100%);
  transition: width 0.3s ease;
}

.search-history-item {
  background: white;
  border: 1px solid #E2E8F0;
  border-radius: 6px;
  padding: 12px;
  cursor: pointer;
  transition: all 0.2s;
  
  &:hover {
    background: rgba(22, 93, 255, 0.05);
    border-color: rgba(22, 93, 255, 0.3);
  }
}

.search-history-time {
  font-size: 11px;
  color: #94A3B8;
  margin-bottom: 8px;
}

.search-history-params {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.search-param-tag {
  background: rgba(22, 93, 255, 0.1);
  color: #165DFF;
  font-size: 12px;
  padding: 3px 8px;
  border-radius: 4px;
}
</style>