<script setup>
import { ref, reactive, onMounted, onUnmounted, computed, watch } from 'vue';
import { useRoute } from 'vue-router';
import { ElMessage } from 'element-plus';
import GlassCard from '@/components/GlassCard.vue';
import { useLogsStore } from '@/stores/logs';
import CryptoJS from 'crypto-js';
import { getTenantName, getTenantOptions } from '@/utils/tenant.js';
import { performLogin, syncGlobalLoginState } from '@/utils/platformLogin.js';

const logsStore = useLogsStore();
const route = useRoute();

// 接口配置（走 Vite 代理）
const QUERY_URL = '/api/blade-detonate/blastDeviceFactory/page';
const BLAST_QUERY_URL = '/api/blade-detonate/blastTask/page';
const LOGIN_URL = '/api/blade-auth/oauth/token';
const TARGET_ORIGIN = 'https://mp.holyview.cn:9443';

// ==================== Tab 切换 ====================
const activeTab = ref('device'); // 'device' 设备数据 | 'blast' 爆破记录

// ==================== 筛选配置（参考批次数据追溯页面） ====================
// 全部可选筛选字段
const filterOptions = ref([
  { key: 'tenantId', label: '管厂' },
  { key: 'companyName', label: '作业单位' },
  { key: 'controllerCode', label: '控制器编号' },
  { key: 'controllerVersion', label: '控制器版本' },
  { key: 'deviceCode', label: '手持机编号' },
  { key: 'softwareVersion', label: '软件版本' },
  { key: 'deviceVersion', label: '手持机版本' },
  { key: 'deviceHardware', label: '手持机类型' },
  { key: 'deviceScene', label: '版本场景' },
  { key: 'batchNo', label: '批次号' },
  { key: 'provinceCode', label: '省份编码' },
  { key: 'cityCode', label: '城市编码' }
]);

// 默认勾选：手持机编号 / 控制器编号 / 软件版本
const filterChecked = reactive({
  tenantId: false,
  companyName: false,
  controllerCode: true,
  controllerVersion: false,
  deviceCode: true,
  softwareVersion: true,
  deviceVersion: false,
  deviceHardware: false,
  deviceScene: false,
  batchNo: false,
  provinceCode: false,
  cityCode: false
});

const showFilterConfig = ref(false);
const FILTER_CONFIG_KEY = 'dataQuery_filterConfig';

const loadFilterConfig = () => {
  try {
    const saved = localStorage.getItem(FILTER_CONFIG_KEY);
    if (saved) {
      const parsed = JSON.parse(saved);
      if (parsed && typeof parsed === 'object') {
        // 合并：保留已存在的勾选状态
        for (const k in parsed) {
          if (k in filterChecked) {
            filterChecked[k] = !!parsed[k];
          }
        }
      }
    }
  } catch (e) {
    console.warn('加载筛选配置失败:', e);
  }
};

const saveFilterConfig = () => {
  try {
    localStorage.setItem(FILTER_CONFIG_KEY, JSON.stringify(filterChecked));
  } catch (e) {
    console.warn('保存筛选配置失败:', e);
  }
};

const openFilterConfig = () => {
  showFilterConfig.value = true;
};

const confirmFilterConfig = () => {
  showFilterConfig.value = false;
  const visibleCount = Object.values(filterChecked).filter(Boolean).length;
  ElMessage.success(`已配置 ${visibleCount} 个筛选字段`);
};

const selectAllFilters = () => {
  for (const k in filterChecked) filterChecked[k] = true;
};

const deselectAllFilters = () => {
  for (const k in filterChecked) filterChecked[k] = false;
};

// 单个字段切换（参考智能制造系统）
const toggleFilter = (key) => {
  if (filterChecked.hasOwnProperty(key)) {
    filterChecked[key] = !filterChecked[key];
  }
};

// 字段可见性判断（参考智能制造系统）
const isFilterVisible = (key) => {
  return filterChecked[key] === true;
};

// 搜索条件
const searchForm = ref({
  tenantId: '',
  tenantName: '',
  companyName: '',
  controllerCode: '',
  controllerVersion: '',
  deviceCode: '',
  softwareVersion: '',
  deviceVersion: '',
  deviceHardware: '',
  deviceScene: '',
  batchNo: '',
  provinceCode: '',
  cityCode: '',
  current: 1,
  size: 10
});

// ==================== 搜索历史（参考批次数据追溯页面） ====================
const searchHistory = ref([]);
const showSearchHistory = ref(false);
const MAX_HISTORY_COUNT = 10;
const HISTORY_STORAGE_KEY = 'dataQuery_searchHistory';

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
  Object.keys(searchForm.value).forEach((key) => {
    if (searchForm.value[key] !== '' && searchForm.value[key] !== null && searchForm.value[key] !== undefined) {
      historyItem.params[key] = searchForm.value[key];
    }
  });
  if (Object.keys(historyItem.params).length > 0) {
    // 去重：相同 params 的旧记录移除
    searchHistory.value = [historyItem, ...searchHistory.value.filter((h) => {
      const same = Object.keys(h.params).length === Object.keys(historyItem.params).length &&
        Object.keys(h.params).every((k) => h.params[k] === historyItem.params[k]);
      return !same;
    })].slice(0, MAX_HISTORY_COUNT);
    localStorage.setItem(HISTORY_STORAGE_KEY, JSON.stringify(searchHistory.value));
  }
};

const getFilterLabel = (key) => {
  const opt = filterOptions.value.find((o) => o.key === key);
  return opt ? opt.label : key;
};

const applySearchHistory = (item) => {
  // 先清空非 current/size 的字段
  Object.keys(searchForm.value).forEach((key) => {
    if (key !== 'current' && key !== 'size') {
      searchForm.value[key] = '';
    }
  });
  // 再赋值
  Object.keys(item.params).forEach((key) => {
    if (searchForm.value.hasOwnProperty(key)) {
      searchForm.value[key] = item.params[key];
    }
  });
  // 自动展开隐藏字段（如果历史中有未显示的字段）
  Object.keys(item.params).forEach((k) => {
    if (filterChecked.hasOwnProperty(k) && !filterChecked[k]) {
      filterChecked[k] = true;
      saveFilterConfig();
    }
  });
  showSearchHistory.value = false;
  searchForm.value.current = 1;
  handleSearch();
};

const deleteSearchHistory = (id) => {
  searchHistory.value = searchHistory.value.filter((h) => h.id !== id);
  localStorage.setItem(HISTORY_STORAGE_KEY, JSON.stringify(searchHistory.value));
};

const clearSearchHistory = () => {
  searchHistory.value = [];
  localStorage.removeItem(HISTORY_STORAGE_KEY);
  ElMessage.success('已清空搜索历史');
};

// ==================== 爆破记录（Tab 2） ====================
const blastSearchForm = ref({
  explosionDate: '',
  uploadDlTime: '',
  tenantName: '',
  deptName: '',
  controllerCode: '',
  controllerVersion: '',
  deviceCode: '',
  deviceVersion: '',
  detonatorCount: '',
  taskName: '',
  blasterUserName: '',
  blasterUserPhone: '',
  current: 1,
  size: 10
});

const blastFilterOptions = ref([
  { key: 'explosionDate', label: '爆破时间' },
  { key: 'uploadDlTime', label: '上传时间' },
  { key: 'tenantName', label: '雷管企业' },
  { key: 'deptName', label: '使用单位' },
  { key: 'controllerCode', label: '控制器编号' },
  { key: 'controllerVersion', label: '控制器版本' },
  { key: 'deviceCode', label: '手持机编号' },
  { key: 'deviceVersion', label: '手持机版本' },
  { key: 'detonatorCount', label: '爆破数量' },
  { key: 'taskName', label: '工程名称' },
  { key: 'blasterUserName', label: '作业人员' },
  { key: 'blasterUserPhone', label: '联系方式' }
]);

const blastFilterChecked = reactive({
  explosionDate: true,
  uploadDlTime: true,
  tenantName: false,
  deptName: true,
  controllerCode: true,
  controllerVersion: false,
  deviceCode: true,
  deviceVersion: false,
  detonatorCount: false,
  taskName: true,
  blasterUserName: false,
  blasterUserPhone: false
});

const blastTableColumns = [
  { key: 'tenantName', label: '雷管企业' },
  { key: 'deptName', label: '使用单位' },
  { key: 'controllerCode', label: '控制器编号' },
  { key: 'controllerVersion', label: '控制器版本' },
  { key: 'deviceCode', label: '手持机编号' },
  { key: 'deviceVersion', label: '手持机版本' },
  { key: 'detonatorCount', label: '爆破数量' },
  { key: 'taskName', label: '工程名称' },
  { key: 'blasterUserName', label: '作业人员' },
  { key: 'blasterUserPhone', label: '联系方式' },
  { key: 'explosionDate', label: '爆破时间' },
  { key: 'uploadDlTime', label: '上传时间' }
];

const blastSearchResult = ref(null);
const isBlastSearching = ref(false);
const blastSearchHistory = ref([]);
const showBlastSearchHistory = ref(false);
const showBlastFilterConfig = ref(false);
const BLAST_HISTORY_KEY = 'dataQuery_blast_searchHistory';

const loadBlastSearchHistory = () => {
  try {
    const str = localStorage.getItem(BLAST_HISTORY_KEY);
    if (str) blastSearchHistory.value = JSON.parse(str);
  } catch (e) {
    console.error('加载爆破记录搜索历史失败:', e);
    blastSearchHistory.value = [];
  }
};

const saveBlastSearchHistory = () => {
  const item = {
    id: Date.now(),
    timestamp: new Date().toLocaleString('zh-CN'),
    params: {}
  };
  Object.keys(blastSearchForm.value).forEach((key) => {
    const v = blastSearchForm.value[key];
    if (v !== '' && v !== null && v !== undefined && key !== 'current' && key !== 'size') {
      item.params[key] = v;
    }
  });
  if (Object.keys(item.params).length > 0) {
    blastSearchHistory.value = [item, ...blastSearchHistory.value.filter((h) => {
      const same = Object.keys(h.params).length === Object.keys(item.params).length &&
        Object.keys(h.params).every((k) => h.params[k] === item.params[k]);
      return !same;
    })].slice(0, MAX_HISTORY_COUNT);
    localStorage.setItem(BLAST_HISTORY_KEY, JSON.stringify(blastSearchHistory.value));
  }
};

const getBlastFilterLabel = (key) => {
  const opt = blastFilterOptions.value.find((o) => o.key === key);
  return opt ? opt.label : key;
};

const applyBlastSearchHistory = (item) => {
  Object.keys(blastSearchForm.value).forEach((key) => {
    if (key !== 'current' && key !== 'size') {
      blastSearchForm.value[key] = '';
    }
  });
  Object.keys(item.params).forEach((key) => {
    if (blastSearchForm.value.hasOwnProperty(key)) {
      blastSearchForm.value[key] = item.params[key];
    }
  });
  Object.keys(item.params).forEach((k) => {
    if (blastFilterChecked.hasOwnProperty(k) && !blastFilterChecked[k]) {
      blastFilterChecked[k] = true;
    }
  });
  showBlastSearchHistory.value = false;
  blastSearchForm.value.current = 1;
  handleBlastSearch();
};

const clearBlastSearchHistory = () => {
  blastSearchHistory.value = [];
  localStorage.removeItem(BLAST_HISTORY_KEY);
  ElMessage.success('已清空爆破记录搜索历史');
};

const toggleBlastFilter = (key) => {
  if (blastFilterChecked.hasOwnProperty(key)) {
    blastFilterChecked[key] = !blastFilterChecked[key];
  }
};

// 爆破记录 - 上传开始日期 / 上传结束日期（默认最近一周，参考批次数据追溯）
const formatDate = (date) => {
  const y = date.getFullYear();
  const m = String(date.getMonth() + 1).padStart(2, '0');
  const d = String(date.getDate()).padStart(2, '0');
  return `${y}-${m}-${d}`;
};

const getLastWeekRange = () => {
  const today = new Date();
  const start = new Date(today);
  start.setDate(today.getDate() - 6);
  return { startDate: formatDate(start), endDate: formatDate(today) };
};

const lastWeekRange = getLastWeekRange();
const blastUploadStartDate = ref(lastWeekRange.startDate);
const blastUploadEndDate = ref(lastWeekRange.endDate);

const isBlastFilterVisible = (key) => {
  return blastFilterChecked[key] === true;
};

const handleBlastReset = () => {
  blastSearchForm.value = {
    explosionDate: '',
    uploadDlTime: '',
    tenantName: '',
    deptName: '',
    controllerCode: '',
    controllerVersion: '',
    deviceCode: '',
    deviceVersion: '',
    detonatorCount: '',
    taskName: '',
    blasterUserName: '',
    blasterUserPhone: '',
    current: 1,
    size: 10
  };
  const lastWeek = getLastWeekRange();
  blastUploadStartDate.value = lastWeek.startDate;
  blastUploadEndDate.value = lastWeek.endDate;
  blastSearchResult.value = null;
};

const handleBlastSearch = async () => {
  if (!accessToken.value) {
    ElMessage.warning('请先登录');
    return;
  }
  isBlastSearching.value = true;
  saveBlastSearchHistory();
  try {
    const params = new URLSearchParams();
    // 上传日期区间（爆破记录专属，默认最近一周）
    if (blastUploadStartDate.value) params.append('startDate', blastUploadStartDate.value);
    if (blastUploadEndDate.value) params.append('endDate', blastUploadEndDate.value);
    // 顶部筛选项 - 通用字段（设备数据和爆破记录都根据这些值进行筛选）
    if (searchForm.value.controllerCode) params.append('controllerCode', searchForm.value.controllerCode);
    if (searchForm.value.controllerVersion) params.append('controllerVersion', searchForm.value.controllerVersion);
    if (searchForm.value.deviceCode) params.append('deviceCode', searchForm.value.deviceCode);
    if (searchForm.value.deviceVersion) params.append('deviceVersion', searchForm.value.deviceVersion);
    // 爆破记录专属字段
    if (blastSearchForm.value.explosionDate) params.append('explosionDate', blastSearchForm.value.explosionDate);
    if (blastSearchForm.value.uploadDlTime) params.append('uploadDlTime', blastSearchForm.value.uploadDlTime);
    if (blastSearchForm.value.tenantName) params.append('tenantName', blastSearchForm.value.tenantName);
    if (blastSearchForm.value.deptName) params.append('deptName', blastSearchForm.value.deptName);
    if (blastSearchForm.value.detonatorCount) params.append('detonatorCount', blastSearchForm.value.detonatorCount);
    if (blastSearchForm.value.taskName) params.append('taskName', blastSearchForm.value.taskName);
    if (blastSearchForm.value.blasterUserName) params.append('blasterUserName', blastSearchForm.value.blasterUserName);
    if (blastSearchForm.value.blasterUserPhone) params.append('blasterUserPhone', blastSearchForm.value.blasterUserPhone);
    params.append('current', blastSearchForm.value.current);
    params.append('size', blastSearchForm.value.size);

    const response = await fetch(`${BLAST_QUERY_URL}?${params.toString()}`, {
      method: 'GET',
      headers: {
        'accept': 'application/json, text/plain, */*',
        'authorization': 'Basic ' + btoa('saber:saber_secret'),
        'blade-auth': `bearer ${accessToken.value}`,
        'tenant-id': '000000'
      }
    });
    const result = await response.json();
    if (result.code === 200 && result.data) {
      const records = result.data.records || [];
      const total = result.data.total || 0;
      const current = result.data.current || blastSearchForm.value.current;
      const size = result.data.size || blastSearchForm.value.size;
      blastSearchResult.value = { records, total, current, size, success: true };
      ElMessage.success(`查询完成，共 ${total} 条爆破记录`);
      logsStore.addLog('查询', 'AI起爆记录', `共${total}条`);
    } else if (result.code === 401) {
      ElMessage.error('Token已过期，请重新登录');
      isLoggedIn.value = false;
      accessToken.value = '';
      localStorage.removeItem('mp_token');
      blastSearchResult.value = null;
    } else {
      ElMessage.error(result.msg || result.message || '查询失败');
      blastSearchResult.value = { records: [], total: 0, current: 1, size: blastSearchForm.value.size, success: false, error: result.msg || result.message };
    }
  } catch (e) {
    ElMessage.error('网络连接失败: ' + e.message);
    blastSearchResult.value = { records: [], total: 0, current: 1, size: blastSearchForm.value.size, success: false, error: e.message };
  } finally {
    isBlastSearching.value = false;
  }
};

// ==================== 统一检索（同时查询设备数据 + 爆破记录） ====================
const handleUnifiedSearch = async () => {
  if (!accessToken.value) {
    ElMessage.warning('请先登录');
    return;
  }
  // 并行触发两个查询
  await Promise.all([
    handleSearch(),
    handleBlastSearch()
  ]);
};

const blastTotalPages = computed(() => {
  if (!blastSearchResult.value || !blastSearchResult.value.total) return 1;
  return Math.ceil(blastSearchResult.value.total / blastSearchForm.value.size) || 1;
});

const handleBlastPageChange = (page) => {
  if (page < 1 || page > blastTotalPages.value) return;
  blastSearchForm.value.current = page;
  handleBlastSearch();
};

const handleBlastSizeChange = (val) => {
  blastSearchForm.value.size = val;
  blastSearchForm.value.current = 1;
  handleBlastSearch();
};

const blastDisplayedPages = computed(() => {
  if (!blastSearchResult.value) return [];
  const current = blastSearchResult.value.current || 1;
  const pages = blastTotalPages.value;
  const result = [];
  if (pages <= 5) {
    for (let i = 1; i <= pages; i++) result.push(i);
  } else {
    if (current <= 3) {
      for (let i = 1; i <= 4; i++) result.push(i);
    } else if (current >= pages - 2) {
      for (let i = pages - 3; i <= pages; i++) result.push(i);
    } else {
      for (let i = current - 1; i <= current + 1; i++) result.push(i);
    }
  }
  return result;
});

const blastShowEllipsisBefore = computed(() => {
  if (!blastSearchResult.value) return false;
  return blastTotalPages.value > 5 && (blastSearchResult.value.current || 1) > 4;
});

const blastShowEllipsisAfter = computed(() => {
  if (!blastSearchResult.value) return false;
  return blastTotalPages.value > 5 && (blastSearchResult.value.current || 1) < blastTotalPages.value - 3;
});

const handleBlastExport = () => {
  if (!blastSearchResult.value || !blastSearchResult.value.records || blastSearchResult.value.records.length === 0) {
    ElMessage.warning('没有可导出的数据');
    return;
  }
  const records = blastSearchResult.value.records;
  const keys = blastTableColumns.map((c) => c.key);
  const headers = blastTableColumns.map((c) => c.label);
  const csvContent = [
    headers.join(','),
    ...records.map((row) =>
      keys.map((k) => {
        const v = row[k];
        if (v === null || v === undefined) return '';
        const str = String(v);
        return str.includes(',') || str.includes('"') || str.includes('\n') ? `"${str.replace(/"/g, '""')}"` : str;
      }).join(',')
    )
  ].join('\n');
  const blob = new Blob(['\uFEFF' + csvContent], { type: 'text/csv;charset=utf-8' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `AI起爆记录_${new Date().toISOString().split('T')[0]}.csv`;
  a.click();
  URL.revokeObjectURL(url);
  ElMessage.success(`已导出 ${records.length} 条记录`);
  logsStore.addLog('导出', 'AI起爆记录', `${records.length}条`);
};

// 切换 tab 时重置另一 tab 的搜索历史下拉
watch(activeTab, (newVal) => {
  showSearchHistory.value = false;
  showBlastSearchHistory.value = false;
});

// ==================== 认证状态 ====================
const accessToken = ref(localStorage.getItem('mp_token') || '');
const isLoggedIn = ref(!!accessToken.value);
const loginForm = ref({
  username: localStorage.getItem('mp_username') || '',
  password: ''
});
const isLoggingIn = ref(false);

// 工具函数
const md5 = (str) => CryptoJS.MD5(str).toString();
const getAuthHeader = () => 'Basic ' + btoa('saber:saber_secret');

const persistToken = (token, expiresIn) => {
  // 持久化 token + 过期时间（避免每次刷新都要重新登录）
  const expireMs = Date.now() + (expiresIn ? expiresIn * 1000 : 24 * 3600 * 1000);
  localStorage.setItem('mp_token', token);
  localStorage.setItem('mp_token_expire', String(expireMs));
};

const loadSavedCredentials = () => {
  try {
    // 1. 优先从 mp_token 恢复
    const token = localStorage.getItem('mp_token');
    const expireStr = localStorage.getItem('mp_token_expire');
    if (token) {
      const expire = expireStr ? Number(expireStr) : 0;
      if (!expire || expire > Date.now()) {
        accessToken.value = token;
        isLoggedIn.value = true;
      } else {
        // 已过期，清除
        localStorage.removeItem('mp_token');
        localStorage.removeItem('mp_token_expire');
      }
    }
    // 2. 从 tester_credentials 恢复（来自三方账号授权页面）
    const saved = localStorage.getItem('tester_credentials');
    if (saved) {
      const data = JSON.parse(saved);
      const now = Date.now();
      if (data.expireTime > now) {
        loginForm.value.username = data.username || loginForm.value.username;
        if (data.accessToken) {
          accessToken.value = data.accessToken;
          isLoggedIn.value = true;
          localStorage.setItem('mp_token', data.accessToken);
        }
      }
    }
  } catch (e) {
    console.error('加载凭据失败:', e);
  }
};

// ==================== 检索结果（参考智能制造系统） ====================
const searchResult = ref(null);
const isSearching = ref(false);

// 表格列定义（按用户指定：8 列）
const tableColumns = [
  {
    key: 'tenantId',
    label: '管厂',
    formatter: (val) => {
      const name = getTenantName(val);
      return name && name !== val ? `${val} ${name}` : val;
    }
  },
  { key: 'companyName', label: '作业单位' },
  { key: 'controllerCode', label: '控制器编号' },
  { key: 'controllerVersion', label: '控制器版本' },
  { key: 'deviceCode', label: '手持机编号' },
  { key: 'deviceVersion', label: '手持机版本' },
  { key: 'deviceHardware', label: '手持机类型' },
  { key: 'deviceScene', label: '版本场景' }
];

const formatCellValue = (record, column) => {
  const value = record[column.key];
  if (value === null || value === undefined || value === '') return '-';
  if (column.formatter) return column.formatter(value);
  return value;
};

// 登录
const handleLogin = async () => {
  if (!loginForm.value.username || !loginForm.value.password) {
    ElMessage.warning('请输入用户名和密码');
    return;
  }

  isLoggingIn.value = true;
  try {
    const result = await performLogin('mp', loginForm.value.username, loginForm.value.password);
    if (result.success) {
      accessToken.value = result.data.access_token;
      isLoggedIn.value = true;
      // 持久化 token + 过期时间
      persistToken(result.data.access_token, result.data.expires_in);
      localStorage.setItem('mp_username', loginForm.value.username);
      ElMessage.success('登录成功');
      logsStore.addLog('登录', 'AI起爆数据查询', `用户: ${loginForm.value.username}`);
      // 登录成功后自动同时查询设备数据 + 爆破记录
      handleUnifiedSearch();
    } else {
      ElMessage.error(`登录失败: ${result.message}`);
    }
  } catch (e) {
    ElMessage.error('网络连接失败: ' + e.message);
  } finally {
    isLoggingIn.value = false;
  }
};

// 退出登录
const handleLogout = () => {
  accessToken.value = '';
  isLoggedIn.value = false;
  localStorage.removeItem('mp_token');
  localStorage.removeItem('mp_token_expire');
  searchResult.value = null;
  blastSearchResult.value = null;
  ElMessage.success('已退出登录');
};

// 检索数据
const handleSearch = async () => {
  if (!accessToken.value) {
    ElMessage.warning('请先登录');
    return;
  }

  isSearching.value = true;
  // 记录搜索历史（在请求前记录）
  saveSearchHistory();
  try {
    const params = new URLSearchParams();
    if (searchForm.value.tenantId) params.append('tenantId', searchForm.value.tenantId);
    if (searchForm.value.companyName) params.append('companyName', searchForm.value.companyName);
    if (searchForm.value.controllerCode) params.append('controllerCode', searchForm.value.controllerCode);
    if (searchForm.value.controllerVersion) params.append('controllerVersion', searchForm.value.controllerVersion);
    if (searchForm.value.deviceCode) params.append('deviceCode', searchForm.value.deviceCode);
    if (searchForm.value.softwareVersion) params.append('softwareVersion', searchForm.value.softwareVersion);
    if (searchForm.value.deviceVersion) params.append('deviceVersion', searchForm.value.deviceVersion);
    if (searchForm.value.deviceHardware) params.append('deviceHardware', searchForm.value.deviceHardware);
    if (searchForm.value.deviceScene) params.append('deviceScene', searchForm.value.deviceScene);
    if (searchForm.value.batchNo) params.append('batchNo', searchForm.value.batchNo);
    if (searchForm.value.provinceCode) params.append('provinceCode', searchForm.value.provinceCode);
    if (searchForm.value.cityCode) params.append('cityCode', searchForm.value.cityCode);
    params.append('current', searchForm.value.current);
    params.append('size', searchForm.value.size);

    const response = await fetch(`${QUERY_URL}?${params.toString()}`, {
      method: 'GET',
      headers: {
        'accept': 'application/json, text/plain, */*',
        'authorization': 'Basic ' + btoa('saber:saber_secret'),
        'blade-auth': `bearer ${accessToken.value}`,
        'tenant-id': '000000'
      }
    });

    const result = await response.json();

    if (result.code === 200 && result.data) {
      const records = result.data.records || [];
      const total = result.data.total || 0;
      const current = result.data.current || searchForm.value.current;
      const size = result.data.size || searchForm.value.size;
      searchResult.value = { records, total, current, size, success: true };
      ElMessage.success(`检索完成，共 ${total} 条记录`);
      logsStore.addLog('查询', 'AI起爆数据查询', `共${total}条`);
    } else if (result.code === 401) {
      ElMessage.error('Token已过期，请重新登录');
      isLoggedIn.value = false;
      accessToken.value = '';
      localStorage.removeItem('mp_token');
      searchResult.value = null;
    } else {
      ElMessage.error(result.msg || result.message || '检索失败');
      searchResult.value = { records: [], total: 0, current: 1, size: searchForm.value.size, success: false, error: result.msg || result.message };
    }
  } catch (e) {
    ElMessage.error('网络连接失败: ' + e.message);
    searchResult.value = { records: [], total: 0, current: 1, size: searchForm.value.size, success: false, error: e.message };
  } finally {
    isSearching.value = false;
  }
};

// 重置
const handleReset = () => {
  searchForm.value = {
    tenantId: '',
    tenantName: '',
    companyName: '',
    controllerCode: '',
    controllerVersion: '',
    deviceCode: '',
    softwareVersion: '',
    deviceVersion: '',
    deviceHardware: '',
    deviceScene: '',
    batchNo: '',
    provinceCode: '',
    cityCode: '',
    current: 1,
    size: 10
  };
  searchResult.value = null;
};

// 导出 Excel
const handleExport = () => {
  if (!searchResult.value || !searchResult.value.records || searchResult.value.records.length === 0) {
    ElMessage.warning('没有可导出的数据');
    return;
  }
  const records = searchResult.value.records;
  const keys = tableColumns.map((c) => c.key);
  const headers = tableColumns.map((c) => c.label);
  const csvContent = [
    headers.join(','),
    ...records.map((row) =>
      keys.map((k) => {
        const v = row[k];
        if (v === null || v === undefined) return '';
        const str = String(v);
        return str.includes(',') || str.includes('"') || str.includes('\n') ? `"${str.replace(/"/g, '""')}"` : str;
      }).join(',')
    )
  ].join('\n');
  const blob = new Blob(['\uFEFF' + csvContent], { type: 'text/csv;charset=utf-8' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `AI起爆数据_${new Date().toISOString().split('T')[0]}.csv`;
  a.click();
  URL.revokeObjectURL(url);
  ElMessage.success(`已导出 ${records.length} 条记录`);
  logsStore.addLog('导出', 'AI起爆数据查询', `${records.length}条`);
};

// 计算总页数（参考智能制造系统）
const totalPages = computed(() => {
  if (!searchResult.value || !searchResult.value.total) return 1;
  return Math.ceil(searchResult.value.total / searchForm.value.size) || 1;
});

// 分页（参考智能制造系统）
const handlePageChange = (page) => {
  if (page < 1 || page > totalPages.value) return;
  searchForm.value.current = page;
  handleSearch();
};

const handleSizeChange = (val) => {
  searchForm.value.size = val;
  searchForm.value.current = 1;
  handleSearch();
};

const displayedPages = computed(() => {
  if (!searchResult.value) return [];
  const current = searchResult.value.current || 1;
  const pages = totalPages.value;
  const visiblePages = 5;
  const result = [];
  if (pages <= visiblePages) {
    for (let i = 1; i <= pages; i++) result.push(i);
  } else {
    if (current <= 3) {
      for (let i = 1; i <= 4; i++) result.push(i);
    } else if (current >= pages - 2) {
      for (let i = pages - 3; i <= pages; i++) result.push(i);
    } else {
      for (let i = current - 1; i <= current + 1; i++) result.push(i);
    }
  }
  return result;
});

const showEllipsisBefore = computed(() => {
  if (!searchResult.value) return false;
  return totalPages.value > 5 && (searchResult.value.current || 1) > 4;
});

const showEllipsisAfter = computed(() => {
  if (!searchResult.value) return false;
  return totalPages.value > 5 && (searchResult.value.current || 1) < totalPages.value - 3;
});

onMounted(() => {
  loadSavedCredentials();
  loadFilterConfig();
  loadSearchHistory();
  loadBlastSearchHistory();
  // 支持 URL 参数预填：?deviceCode=DZ600000016（首页 AI 助手本地拆解后跳转）
  const urlDeviceCode = route.query.deviceCode;
  if (urlDeviceCode) {
    searchForm.value.deviceCode = String(urlDeviceCode);
  }
  if (accessToken.value) {
    // 已登录则自动同时查询设备数据 + 爆破记录
    handleUnifiedSearch();
  }
});

onUnmounted(() => {
});
</script>

<template>
  <div class="data-query">
    <!-- 登录卡片（未登录时显示） -->
    <GlassCard v-if="!isLoggedIn" title="AI起爆数据查询 - 登录">
      <div class="login-form">
        <div class="form-row">
          <div class="form-item">
            <label class="form-label">用户名</label>
            <input
              v-model="loginForm.username"
              type="text"
              class="form-input"
              placeholder="请输入用户名"
            />
          </div>
          <div class="form-item">
            <label class="form-label">密码</label>
            <input
              v-model="loginForm.password"
              type="password"
              class="form-input"
              placeholder="请输入密码"
              @keyup.enter="handleLogin"
            />
          </div>
        </div>
        <div class="form-actions">
          <button
            class="search-btn"
            :disabled="isLoggingIn"
            @click="handleLogin"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M15 3h4a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-4" />
              <polyline points="10 17 15 12 10 7" />
              <line x1="15" y1="12" x2="3" y2="12" />
            </svg>
            <span>{{ isLoggingIn ? '登录中...' : '登录' }}</span>
          </button>
        </div>
        <div class="login-tip">
          💡 提示：可前往「三方账号授权」页面登录云系统，凭据会自动同步
        </div>
      </div>
    </GlassCard>

    <!-- 检索卡片（已登录时显示） -->
    <GlassCard v-else title="AI起爆数据查询" :extra="`已登录: ${loginForm.username}`">
      <template #extra>
        <div class="card-extra">
          <span class="user-info">已登录: {{ loginForm.username }}</span>
          <button class="logout-btn" @click="handleLogout">退出</button>
        </div>
      </template>

      <div class="search-section">
        <!-- 顶部工具栏：筛选配置 + 搜索历史（固定不变） -->
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
            <button class="clear-history-btn" @click.stop="clearSearchHistory">清空</button>
          </div>
          <div v-if="searchHistory.length === 0" class="search-history-empty">暂无搜索历史</div>
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
                >{{ getFilterLabel(key) }}: {{ value }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- 筛选配置面板 -->
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

        <!-- 搜索字段（固定：手持机&控制器编号、软件版本等） -->
        <div class="search-grid">
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
          <div v-if="isFilterVisible('companyName')" class="form-group">
            <label>作业单位</label>
            <input v-model="searchForm.companyName" type="text" class="form-input" placeholder="输入作业单位名称" />
          </div>
          <div v-if="isFilterVisible('controllerCode')" class="form-group">
            <label>控制器编号</label>
            <input v-model="searchForm.controllerCode" type="text" class="form-input" placeholder="输入控制器编号" />
          </div>
          <div v-if="isFilterVisible('controllerVersion')" class="form-group">
            <label>控制器版本</label>
            <input v-model="searchForm.controllerVersion" type="text" class="form-input" placeholder="如 8.5-B" />
          </div>
          <div v-if="isFilterVisible('deviceCode')" class="form-group">
            <label>手持机编号</label>
            <input v-model="searchForm.deviceCode" type="text" class="form-input" placeholder="输入手持机编号" />
          </div>
          <div v-if="isFilterVisible('softwareVersion')" class="form-group">
            <label>软件版本</label>
            <input v-model="searchForm.softwareVersion" type="text" class="form-input" placeholder="如 L.1.2.11T2" />
          </div>
          <div v-if="isFilterVisible('deviceVersion')" class="form-group">
            <label>手持机版本</label>
            <input v-model="searchForm.deviceVersion" type="text" class="form-input" placeholder="如 1.0.0" />
          </div>
          <div v-if="isFilterVisible('deviceHardware')" class="form-group">
            <label>手持机类型</label>
            <input v-model="searchForm.deviceHardware" type="text" class="form-input" placeholder="如 DT40 / 小勇士" />
          </div>
          <div v-if="isFilterVisible('deviceScene')" class="form-group">
            <label>版本场景</label>
            <input v-model="searchForm.deviceScene" type="text" class="form-input" placeholder="如 煤矿 / 隧道" />
          </div>
          <div v-if="isFilterVisible('batchNo')" class="form-group">
            <label>批次号</label>
            <input v-model="searchForm.batchNo" type="text" class="form-input" placeholder="输入批次号" />
          </div>
          <div v-if="isFilterVisible('provinceCode')" class="form-group">
            <label>省份编码</label>
            <input v-model="searchForm.provinceCode" type="text" class="form-input" placeholder="省份编码" />
          </div>
          <div v-if="isFilterVisible('cityCode')" class="form-group">
            <label>城市编码</label>
            <input v-model="searchForm.cityCode" type="text" class="form-input" placeholder="城市编码" />
          </div>

          <!-- 上传日期区间（仅爆破记录 Tab 显示，默认最近一周，参考批次数据追溯） -->
          <div v-if="activeTab === 'blast'" class="form-group">
            <label>上传开始日期</label>
            <input
              v-model="blastUploadStartDate"
              type="date"
              class="form-input"
            />
          </div>

          <div v-if="activeTab === 'blast'" class="form-group">
            <label>上传结束日期</label>
            <input
              v-model="blastUploadEndDate"
              type="date"
              class="form-input"
            />
          </div>
        </div>

        <!-- 操作按钮（点击同时查询设备数据 + 爆破记录） -->
        <div class="form-actions">
          <button
            class="action-btn primary"
            :disabled="isSearching || isBlastSearching"
            @click="handleUnifiedSearch"
          >
            <svg v-if="!isSearching && !isBlastSearching" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="11" cy="11" r="8" />
              <line x1="21" y1="21" x2="16.65" y2="16.65" />
            </svg>
            <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="spin">
              <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
            </svg>
            <span>{{ (isSearching || isBlastSearching) ? '检索中...' : '检索数据' }}</span>
          </button>

          <button
            class="action-btn secondary"
            :disabled="isSearching || isBlastSearching"
            @click="activeTab === 'device' ? handleReset() : handleBlastReset()"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8" />
              <path d="M3 3v5h5" />
            </svg>
            <span>重置</span>
          </button>

          <button
            class="action-btn success"
            :disabled="activeTab === 'device'
              ? (!searchResult || !searchResult.records || searchResult.records.length === 0)
              : (!blastSearchResult || !blastSearchResult.records || blastSearchResult.records.length === 0)"
            @click="activeTab === 'device' ? handleExport() : handleBlastExport()"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
              <polyline points="7 10 12 15 17 10" />
              <line x1="12" y1="15" x2="12" y2="3" />
            </svg>
            <span>导出Excel</span>
          </button>
        </div>
      </div>
    </GlassCard>

    <!-- Tab 切换 + 数据列表（参考智能制造系统） -->
    <div class="records-panel">
      <!-- Tab 头部 -->
      <div class="records-tab-header">
        <div class="tab-bar">
          <button
            class="tab-btn"
            :class="{ active: activeTab === 'device' }"
            @click="activeTab = 'device'"
          >
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <rect x="4" y="4" width="16" height="16" rx="2" />
              <rect x="9" y="9" width="6" height="6" />
            </svg>
            <span>设备数据</span>
          </button>
          <button
            class="tab-btn"
            :class="{ active: activeTab === 'blast' }"
            @click="activeTab = 'blast'"
          >
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M12 2v6" />
              <path d="M12 22a8 8 0 0 0 8-8c0-3-2-5-4-6-1 0-2 1-4 1s-3-1-4-1c-2 1-4 3-4 6a8 8 0 0 0 8 8z" />
            </svg>
            <span>爆破记录</span>
          </button>
        </div>
        <span class="pagination-info">
          <template v-if="activeTab === 'device' && searchResult">
            共 {{ searchResult.total }} 条 | 第 {{ searchResult.current }} 页 | 每页 {{ searchResult.size }} 条
          </template>
          <template v-else-if="activeTab === 'blast' && blastSearchResult">
            共 {{ blastSearchResult.total }} 条 | 第 {{ blastSearchResult.current }} 页 | 每页 {{ blastSearchResult.size }} 条
          </template>
        </span>
      </div>

      <!-- ============== Tab: 设备数据 ============== -->
      <div v-show="activeTab === 'device'">
        <!-- 操作按钮已上移到顶部共享区 -->

        <!-- 设备数据结果状态 -->
        <div v-if="searchResult" class="result-panel" :class="searchResult.success ? 'success' : 'error'">
          <div class="result-header">
            <svg v-if="searchResult.success" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <polyline points="20 6 9 17 4 12" />
            </svg>
            <svg v-else width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="12" r="10" />
              <line x1="15" y1="9" x2="9" y2="15" />
              <line x1="9" y1="9" x2="15" y2="15" />
            </svg>
            <span>{{ searchResult.success ? '检索成功' : '检索失败' }}</span>
          </div>
          <pre v-if="!searchResult.success" class="result-content">{{ JSON.stringify(searchResult.error, null, 2) }}</pre>
        </div>

        <!-- 设备数据表格 -->
        <div v-if="searchResult?.success && searchResult.records" class="table-container">
          <table class="trace-table">
            <thead>
              <tr>
                <th>序号</th>
                <th v-for="col in tableColumns" :key="col.key">{{ col.label }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(record, index) in searchResult.records" :key="index">
                <td>{{ (searchResult.current - 1) * searchResult.size + index + 1 }}</td>
                <td v-for="col in tableColumns" :key="col.key" class="cell-td">
                  <div class="cell-content" :title="String(formatCellValue(record, col))">
                    {{ formatCellValue(record, col) }}
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- 设备数据空态 -->
        <div v-else-if="!searchResult" class="empty-panel">
          <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="11" cy="11" r="8" />
            <line x1="21" y1="21" x2="16.65" y2="16.65" />
          </svg>
          <p>{{ isSearching ? '检索中...' : '暂无数据，请输入条件后点击"检索数据"' }}</p>
        </div>

        <!-- 设备数据分页 -->
        <div v-if="searchResult?.success && searchResult.records" class="pagination-container">
          <div class="pagination">
            <button
              class="page-btn"
              :disabled="(searchResult.current || 1) <= 1 || isSearching"
              @click="handlePageChange((searchResult.current || 1) - 1)"
            >
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <polyline points="15 18 9 12 15 6" />
              </svg>
            </button>
            <div class="page-numbers">
              <button
                v-for="page in displayedPages"
                :key="page"
                class="page-num-btn"
                :class="{ active: page === (searchResult.current || 1) }"
                @click="handlePageChange(page)"
              >{{ page }}</button>
              <span v-if="showEllipsisBefore" class="ellipsis">...</span>
              <span v-if="showEllipsisAfter" class="ellipsis">...</span>
            </div>
            <button
              class="page-btn"
              :disabled="(searchResult.current || 1) >= totalPages || isSearching"
              @click="handlePageChange((searchResult.current || 1) + 1)"
            >
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <polyline points="9 18 15 12 9 6" />
              </svg>
            </button>
            <div class="page-size-selector">
              <span>每页</span>
              <select
                v-model.number="searchForm.size"
                class="page-size-input"
                @change="handleSizeChange(searchForm.size)"
              >
                <option :value="10">10</option>
                <option :value="20">20</option>
                <option :value="50">50</option>
                <option :value="100">100</option>
              </select>
              <span>条</span>
            </div>
          </div>
        </div>
      </div>

      <!-- ============== Tab: 爆破记录 ============== -->
      <div v-show="activeTab === 'blast'">
        <!-- 操作按钮已上移到顶部共享区，日期区间也已在顶部筛选框内 -->

        <!-- 爆破记录结果状态 -->
        <div v-if="blastSearchResult" class="result-panel" :class="blastSearchResult.success ? 'success' : 'error'">
          <div class="result-header">
            <svg v-if="blastSearchResult.success" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <polyline points="20 6 9 17 4 12" />
            </svg>
            <svg v-else width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="12" r="10" />
              <line x1="15" y1="9" x2="9" y2="15" />
              <line x1="9" y1="9" x2="15" y2="15" />
            </svg>
            <span>{{ blastSearchResult.success ? '查询成功' : '查询失败' }}</span>
          </div>
          <pre v-if="!blastSearchResult.success" class="result-content">{{ JSON.stringify(blastSearchResult.error, null, 2) }}</pre>
        </div>

        <!-- 爆破记录表格 -->
        <div v-if="blastSearchResult?.success && blastSearchResult.records" class="table-container">
          <table class="trace-table">
            <thead>
              <tr>
                <th>序号</th>
                <th v-for="col in blastTableColumns" :key="col.key">{{ col.label }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(record, index) in blastSearchResult.records" :key="index">
                <td>{{ (blastSearchResult.current - 1) * blastSearchResult.size + index + 1 }}</td>
                <td v-for="col in blastTableColumns" :key="col.key" class="cell-td">
                  <div class="cell-content" :title="String(formatCellValue(record, col))">
                    {{ formatCellValue(record, col) }}
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- 爆破记录空态 -->
        <div v-else-if="!blastSearchResult" class="empty-panel">
          <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <path d="M12 2v6" />
            <path d="M12 22a8 8 0 0 0 8-8c0-3-2-5-4-6-1 0-2 1-4 1s-3-1-4-1c-2 1-4 3-4 6a8 8 0 0 0 8 8z" />
          </svg>
          <p>{{ isBlastSearching ? '查询中...' : '暂无爆破记录，请输入条件后点击"查询爆破记录"' }}</p>
        </div>

        <!-- 爆破记录分页 -->
        <div v-if="blastSearchResult?.success && blastSearchResult.records" class="pagination-container">
          <div class="pagination">
            <button
              class="page-btn"
              :disabled="(blastSearchResult.current || 1) <= 1 || isBlastSearching"
              @click="handleBlastPageChange((blastSearchResult.current || 1) - 1)"
            >
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <polyline points="15 18 9 12 15 6" />
              </svg>
            </button>
            <div class="page-numbers">
              <button
                v-for="page in blastDisplayedPages"
                :key="page"
                class="page-num-btn"
                :class="{ active: page === (blastSearchResult.current || 1) }"
                @click="handleBlastPageChange(page)"
              >{{ page }}</button>
              <span v-if="blastShowEllipsisBefore" class="ellipsis">...</span>
              <span v-if="blastShowEllipsisAfter" class="ellipsis">...</span>
            </div>
            <button
              class="page-btn"
              :disabled="(blastSearchResult.current || 1) >= blastTotalPages || isBlastSearching"
              @click="handleBlastPageChange((blastSearchResult.current || 1) + 1)"
            >
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <polyline points="9 18 15 12 9 6" />
              </svg>
            </button>
            <div class="page-size-selector">
              <span>每页</span>
              <select
                v-model.number="blastSearchForm.size"
                class="page-size-input"
                @change="handleBlastSizeChange(blastSearchForm.size)"
              >
                <option :value="10">10</option>
                <option :value="20">20</option>
                <option :value="50">50</option>
                <option :value="100">100</option>
              </select>
              <span>条</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default { name: 'DataQuery' };
</script>

<style scoped>
.data-query {
  max-width: 100%;
}

/* 卡片右上角 */
:deep(.el-card__header) {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.card-extra {
  display: flex;
  align-items: center;
  gap: 12px;
}
.user-info {
  font-size: 13px;
  color: var(--text-tertiary);
}
.logout-btn {
  padding: 4px 12px;
  background: rgba(245, 63, 63, 0.1);
  border: 1px solid rgba(245, 63, 63, 0.3);
  border-radius: 4px;
  color: #F53F3F;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s;
}
.logout-btn:hover {
  background: rgba(245, 63, 63, 0.2);
}

/* 登录表单 */
.login-form {
  padding: 24px 16px;
  max-width: 600px;
}
.login-tip {
  margin-top: 16px;
  padding: 10px 14px;
  background: rgba(22, 93, 255, 0.08);
  border-left: 3px solid #165DFF;
  border-radius: 4px;
  font-size: 12px;
  color: var(--text-secondary);
}

.search-form {
  padding: 16px;
  background: var(--bg-input);
  border-radius: 8px;
}

.form-row {
  display: flex;
  gap: 16px;
  flex-wrap: wrap;
}

.form-item {
  flex: 1;
  min-width: 180px;
}
.form-item-small {
  flex: 0 0 100px;
  min-width: 100px;
}

.form-label {
  display: block;
  font-size: 12px;
  color: var(--text-tertiary);
  margin-bottom: 8px;
}

.form-input {
  width: 100%;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 10px 12px;
  color: var(--text-primary);
  font-size: 13px;
  outline: none;
  transition: border-color 0.2s;
}
.form-input:focus {
  border-color: rgba(22, 93, 255, 0.6);
}
.form-input::placeholder {
  color: var(--text-tertiary);
}

.form-actions {
  display: flex;
  gap: 12px;
  margin-top: 16px;
  flex-wrap: wrap;
}

/* 高级筛选区 */
.advanced-filters {
  margin-top: 14px;
  padding: 14px;
  background: rgba(64, 158, 255, 0.04);
  border: 1px dashed rgba(64, 158, 255, 0.25);
  border-radius: 8px;
  animation: slideDown 0.25s ease;
}
@keyframes slideDown {
  from { opacity: 0; transform: translateY(-6px); }
  to { opacity: 1; transform: translateY(0); }
}

/* 工具栏：更多筛选 / 筛选配置 / 搜索历史 */
.form-toolbar {
  display: flex;
  gap: 8px;
  margin-top: 14px;
  flex-wrap: wrap;
  position: relative;
}
.toolbar-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 6px;
  color: var(--text-secondary);
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s;
}
.toolbar-btn:hover {
  background: rgba(64, 158, 255, 0.1);
  border-color: rgba(64, 158, 255, 0.4);
  color: var(--text-primary);
}
.toolbar-btn.is-active {
  background: rgba(64, 158, 255, 0.15);
  border-color: rgba(64, 158, 255, 0.6);
  color: #409EFF;
}
.history-badge {
  display: inline-block;
  min-width: 18px;
  padding: 0 5px;
  background: #f53f3f;
  color: #fff;
  border-radius: 9px;
  font-size: 10px;
  line-height: 16px;
  text-align: center;
}

/* 历史下拉面板 */
.history-wrapper {
  position: relative;
}
.history-dropdown {
  position: absolute;
  top: 100%;
  left: 0;
  margin-top: 6px;
  min-width: 360px;
  max-width: 480px;
  max-height: 400px;
  overflow-y: auto;
  background: var(--bg-card, #1e1e2e);
  border: 1px solid rgba(255, 255, 255, 0.15);
  border-radius: 8px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.3);
  z-index: 100;
}
.history-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 14px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  font-size: 12px;
  color: var(--text-tertiary);
}
.history-clear-all {
  background: none;
  border: none;
  color: #f53f3f;
  font-size: 12px;
  cursor: pointer;
  padding: 2px 6px;
}
.history-clear-all:hover {
  background: rgba(245, 63, 63, 0.1);
  border-radius: 4px;
}
.history-empty {
  padding: 30px 14px;
  text-align: center;
  color: var(--text-tertiary);
  font-size: 13px;
}
.history-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.04);
  transition: background 0.15s;
}
.history-item:hover {
  background: rgba(64, 158, 255, 0.08);
}
.history-item-main {
  flex: 1;
  cursor: pointer;
  min-width: 0;
}
.history-item-conditions {
  font-size: 13px;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.history-item-time {
  font-size: 11px;
  color: var(--text-tertiary);
  margin-top: 2px;
}
.history-item-remove {
  width: 22px;
  height: 22px;
  border: none;
  background: none;
  color: var(--text-tertiary);
  font-size: 18px;
  line-height: 1;
  border-radius: 4px;
  cursor: pointer;
  flex-shrink: 0;
}
.history-item-remove:hover {
  background: rgba(245, 63, 63, 0.15);
  color: #f53f3f;
}

/* 弹窗 */
.modal-mask {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  animation: fadeIn 0.2s ease;
}
@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}
.modal-container {
  width: 520px;
  max-width: 90vw;
  max-height: 80vh;
  background: var(--bg-card, #1e1e2e);
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 12px;
  box-shadow: 0 16px 48px rgba(0, 0, 0, 0.4);
  display: flex;
  flex-direction: column;
  animation: scaleIn 0.2s ease;
}
@keyframes scaleIn {
  from { opacity: 0; transform: scale(0.95); }
  to { opacity: 1; transform: scale(1); }
}
.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}
.modal-header h3 {
  margin: 0;
  font-size: 16px;
  color: var(--text-primary);
}
.modal-close {
  background: none;
  border: none;
  color: var(--text-tertiary);
  font-size: 24px;
  line-height: 1;
  cursor: pointer;
  padding: 0 4px;
}
.modal-close:hover {
  color: var(--text-primary);
}
.modal-body {
  padding: 16px 20px;
  overflow-y: auto;
  flex: 1;
}
.modal-hint {
  margin: 0 0 12px;
  font-size: 12px;
  color: var(--text-tertiary);
}
.config-actions {
  display: flex;
  gap: 8px;
  align-items: center;
  margin-bottom: 12px;
  padding-bottom: 10px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}
.config-action-btn {
  padding: 4px 10px;
  background: rgba(64, 158, 255, 0.1);
  border: 1px solid rgba(64, 158, 255, 0.3);
  border-radius: 4px;
  color: #409EFF;
  font-size: 12px;
  cursor: pointer;
}
.config-action-btn:hover {
  background: rgba(64, 158, 255, 0.2);
}
.config-count {
  margin-left: auto;
  font-size: 12px;
  color: var(--text-tertiary);
}
.config-list {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
}
.config-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 10px;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.15s;
}
.config-item:hover {
  background: rgba(64, 158, 255, 0.08);
  border-color: rgba(64, 158, 255, 0.3);
}
.config-item input[type="checkbox"] {
  margin: 0;
  cursor: pointer;
}
.config-item-label {
  flex: 1;
  font-size: 13px;
  color: var(--text-primary);
}
.config-item-key {
  font-size: 11px;
  color: var(--text-tertiary);
  font-family: monospace;
}
.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  padding: 12px 20px;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
}
.modal-btn {
  padding: 8px 18px;
  border-radius: 6px;
  font-size: 13px;
  cursor: pointer;
  border: 1px solid transparent;
  transition: all 0.2s;
}
.modal-btn.cancel {
  background: rgba(255, 255, 255, 0.05);
  border-color: rgba(255, 255, 255, 0.15);
  color: var(--text-secondary);
}
.modal-btn.cancel:hover {
  background: rgba(255, 255, 255, 0.1);
}
.modal-btn.confirm {
  background: #409EFF;
  color: #fff;
  border-color: #409EFF;
}
.modal-btn.confirm:hover:not(:disabled) {
  background: #66b1ff;
  border-color: #66b1ff;
}
.modal-btn.confirm:disabled {
  background: rgba(64, 158, 255, 0.4);
  border-color: rgba(64, 158, 255, 0.4);
  cursor: not-allowed;
}

.search-btn,
.reset-btn,
.export-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 10px 20px;
  border: none;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}
.search-btn {
  background: linear-gradient(135deg, #165DFF 0%, #0F4CD0 100%);
  color: #FFFFFF;
}
.search-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(22, 93, 255, 0.3);
}
.reset-btn {
  background: var(--bg-input);
  color: var(--text-secondary);
  border: 1px solid var(--border-color);
}
.reset-btn:hover:not(:disabled) {
  background: rgba(22, 93, 255, 0.05);
}
.export-btn {
  background: rgba(0, 180, 42, 0.1);
  color: #00B42A;
  border: 1px solid rgba(0, 180, 42, 0.3);
}
.export-btn:hover:not(:disabled) {
  background: rgba(0, 180, 42, 0.2);
}
.search-btn:disabled,
.reset-btn:disabled,
.export-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* 数据统计 */
.data-stats {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 10px 14px;
  background: var(--bg-input);
  border-radius: 6px;
  font-size: 13px;
  color: var(--text-secondary);
  margin-bottom: 12px;
}
.data-stats strong {
  color: #165DFF;
  font-weight: 600;
}

/* 表格 */
.tasks-table {
  overflow-x: auto;
}
.tasks-table table {
  width: 100%;
  border-collapse: collapse;
}
.tasks-table th,
.tasks-table td {
  padding: 10px 12px;
  text-align: left;
  font-size: 12px;
  border-bottom: 1px solid var(--border-color);
  white-space: nowrap;
}
.tasks-table th {
  color: var(--text-tertiary);
  font-weight: 500;
  background: var(--bg-input);
  position: sticky;
  top: 0;
  z-index: 1;
}
.tasks-table td {
  color: var(--text-secondary);
}
.tasks-table td.text-null {
  color: var(--text-tertiary);
  font-style: italic;
}
.col-index {
  width: 60px;
  text-align: center !important;
  color: var(--text-tertiary) !important;
}
.tasks-table tbody tr:hover {
  background: rgba(22, 93, 255, 0.03);
}

/* 空状态 */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  color: var(--text-tertiary);
}
.empty-state svg {
  opacity: 0.4;
  margin-bottom: 12px;
}
.empty-state p {
  font-size: 13px;
}

/* 分页 */
.pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
  padding: 16px 0;
}
.page-btn {
  padding: 6px 16px;
  background: var(--bg-input);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  color: var(--text-primary);
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s;
}
.page-btn:hover:not(:disabled) {
  background: rgba(22, 93, 255, 0.08);
  border-color: rgba(22, 93, 255, 0.4);
  color: #165DFF;
}
.page-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}
.page-info {
  font-size: 12px;
  color: var(--text-tertiary);
  min-width: 60px;
  text-align: center;
}

/* 响应式 */
@media screen and (max-width: 768px) {
  .form-row {
    flex-direction: column;
    gap: 12px;
  }
  .form-item,
  .form-item-small {
    min-width: 100%;
    flex: 1;
  }
  .form-actions {
    flex-direction: column;
  }
  .search-btn,
  .reset-btn,
  .export-btn {
    width: 100%;
    justify-content: center;
  }
  .tasks-table th,
  .tasks-table td {
    padding: 8px 10px;
    font-size: 11px;
  }
  .data-stats {
    flex-direction: column;
    align-items: flex-start;
    gap: 6px;
  }
}

/* ==================== 智能制造系统风格样式（参考 FactoryDataQuery） ==================== */
.search-section {
  margin-bottom: 20px;
}

.filter-config-header {
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.history-wrapper {
  position: relative;
  display: inline-block;
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
}
.config-btn:hover {
  background: rgba(22, 93, 255, 0.18);
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
  color: var(--text-primary, #1F2937);
}
.clear-history-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 3px 8px;
  background: transparent;
  border: 1px solid #E2E8F0;
  border-radius: 4px;
  color: #F53F3F;
  font-size: 11px;
  cursor: pointer;
}
.clear-history-btn:hover {
  background: rgba(245, 63, 63, 0.05);
}
.search-history-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 24px;
  color: var(--text-tertiary, #94A3B8);
  font-size: 13px;
  gap: 8px;
}
.search-history-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.search-history-item {
  background: white;
  border: 1px solid #E2E8F0;
  border-radius: 6px;
  padding: 10px 12px;
  cursor: pointer;
  transition: all 0.2s;
}
.search-history-item:hover {
  background: rgba(22, 93, 255, 0.05);
  border-color: rgba(22, 93, 255, 0.3);
}
.search-history-time {
  font-size: 11px;
  color: var(--text-tertiary, #94A3B8);
  margin-bottom: 4px;
}
.search-history-params {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}
.search-param-tag {
  display: inline-block;
  padding: 2px 6px;
  background: rgba(22, 93, 255, 0.08);
  border: 1px solid rgba(22, 93, 255, 0.2);
  border-radius: 3px;
  color: #165DFF;
  font-size: 11px;
}

.filter-config-panel {
  background: rgba(22, 93, 255, 0.04);
  border: 1px dashed rgba(22, 93, 255, 0.3);
  border-radius: 6px;
  padding: 12px 16px;
  margin-bottom: 16px;
}
.filter-options {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}
.filter-option {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  cursor: pointer;
  user-select: none;
}
.filter-option input[type="checkbox"] {
  cursor: pointer;
  accent-color: #165DFF;
}

.search-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 20px;
}
.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.form-group label {
  font-size: 12px;
  color: var(--text-secondary, #475569);
  font-weight: 500;
}

.tenant-name-hint {
  font-size: 11px;
  color: #165DFF;
  padding: 2px 6px;
  background: rgba(22, 93, 255, 0.06);
  border-radius: 4px;
  margin-top: 2px;
}

.form-actions {
  display: flex;
  gap: 12px;
  padding-top: 8px;
  border-top: 1px solid #E2E8F0;
}
.action-btn {
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
}
.action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
.action-btn.primary {
  background: #165DFF;
  color: white;
}
.action-btn.primary:hover:not(:disabled) {
  background: #0E4FD8;
}
.action-btn.secondary {
  background: white;
  color: #475569;
  border-color: #E2E8F0;
}
.action-btn.secondary:hover:not(:disabled) {
  background: #F1F5F9;
  border-color: #CBD5E1;
}
.action-btn.success {
  background: #00B42A;
  color: white;
}
.action-btn.success:hover:not(:disabled) {
  background: #009A24;
}
.spin {
  animation: spin 1s linear infinite;
}
@keyframes spin {
  to { transform: rotate(360deg); }
}

/* 爆破记录 - 上传时间区间（顶部筛选框内） */
.form-group-wide {
  grid-column: span 2;
}
.blast-date-picker-top {
  width: 100%;
}
.blast-date-picker-top :deep(.el-range-editor.el-input__wrapper) {
  width: 100%;
}
@media (max-width: 900px) {
  .form-group-wide { grid-column: span 1; }
}

.result-panel {
  margin-top: 20px;
  padding: 12px 16px;
  border-radius: 6px;
}
.result-panel.success {
  background: rgba(0, 180, 42, 0.08);
  border: 1px solid rgba(0, 180, 42, 0.3);
  color: #00B42A;
}
.result-panel.error {
  background: rgba(245, 63, 63, 0.08);
  border: 1px solid rgba(245, 63, 63, 0.3);
  color: #F53F3F;
}
.result-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-weight: 500;
}
.result-content {
  margin-top: 8px;
  padding: 8px;
  background: rgba(245, 63, 63, 0.05);
  border-radius: 4px;
  font-size: 11px;
  color: #475569;
  white-space: pre-wrap;
  word-break: break-all;
}

.records-panel {
  margin-top: 20px;
  background: white;
  border: 1px solid #E2E8F0;
  border-radius: 8px;
  padding: 16px;
}

.records-tab-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid #E2E8F0;
  flex-wrap: wrap;
  gap: 12px;
}

.tab-bar {
  display: flex;
  align-items: center;
  gap: 4px;
  background: #F1F5F9;
  border-radius: 8px;
  padding: 4px;
}

.tab-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  background: transparent;
  border: none;
  border-radius: 6px;
  color: #64748B;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}
.tab-btn:hover:not(.active) {
  color: #165DFF;
  background: rgba(22, 93, 255, 0.06);
}
.tab-btn.active {
  background: #165DFF;
  color: #ffffff;
  box-shadow: 0 2px 6px rgba(22, 93, 255, 0.25);
}
.tab-btn.active svg {
  stroke: #ffffff;
}
.records-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary, #1F2937);
}
.pagination-info {
  font-size: 12px;
  color: var(--text-tertiary, #94A3B8);
  font-weight: normal;
}

.table-container {
  overflow-x: auto;
  margin-bottom: 16px;
}
.trace-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
}
.trace-table thead {
  background: #F1F5F9;
}
.trace-table th {
  padding: 10px 12px;
  text-align: left;
  font-weight: 600;
  color: var(--text-primary, #1F2937);
  border-bottom: 2px solid #E2E8F0;
  white-space: nowrap;
}
.trace-table td {
  padding: 10px 12px;
  border-bottom: 1px solid #F1F5F9;
  color: var(--text-secondary, #475569);
}
.trace-table tbody tr:hover {
  background: rgba(22, 93, 255, 0.02);
}
.cell-td {
  max-width: 200px;
}
.cell-content {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.pagination-container {
  display: flex;
  justify-content: center;
  padding-top: 12px;
  border-top: 1px solid #F1F5F9;
}
.pagination {
  display: flex;
  align-items: center;
  gap: 8px;
}
.page-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  background: white;
  border: 1px solid #E2E8F0;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s;
  color: #475569;
}
.page-btn:hover:not(:disabled) {
  background: rgba(22, 93, 255, 0.05);
  border-color: rgba(22, 93, 255, 0.3);
  color: #165DFF;
}
.page-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}
.page-numbers {
  display: flex;
  align-items: center;
  gap: 4px;
}
.page-num-btn {
  min-width: 32px;
  height: 32px;
  padding: 0 8px;
  background: white;
  border: 1px solid #E2E8F0;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  color: #475569;
  transition: all 0.2s;
}
.page-num-btn:hover:not(.active) {
  background: rgba(22, 93, 255, 0.05);
  border-color: rgba(22, 93, 255, 0.3);
  color: #165DFF;
}
.page-num-btn.active {
  background: #165DFF;
  border-color: #165DFF;
  color: white;
  font-weight: 600;
}
.ellipsis {
  color: var(--text-tertiary, #94A3B8);
  padding: 0 4px;
}
.page-size-selector {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-left: 16px;
  font-size: 12px;
  color: var(--text-secondary, #475569);
}
.page-size-input {
  height: 30px;
  padding: 0 6px;
  background: white;
  border: 1px solid #E2E8F0;
  border-radius: 4px;
  font-size: 12px;
  color: #475569;
  cursor: pointer;
}

.empty-panel {
  margin-top: 40px;
  padding: 60px 20px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  color: var(--text-tertiary, #94A3B8);
  font-size: 13px;
}

@media (max-width: 1200px) {
  .search-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}
@media (max-width: 900px) {
  .search-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
@media (max-width: 600px) {
  .search-grid {
    grid-template-columns: 1fr;
  }
}

@media screen and (max-width: 480px) {
  .search-form {
    padding: 12px;
  }
  .col-index {
    width: 40px;
  }
}
</style>
