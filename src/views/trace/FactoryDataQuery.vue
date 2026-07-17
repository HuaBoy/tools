<script setup>
import { ref, reactive, onMounted, onUnmounted, computed, watch } from 'vue';
import { ElMessage } from 'element-plus';
import GlassCard from '@/components/GlassCard.vue';
import { useLogsStore } from '@/stores/logs';
import { useRoute } from 'vue-router';
import CryptoJS from 'crypto-js';
import { saveCredentials as saveCred, clearCredentials as clearCred, updateLoginStatus, loginStatus, onLoginStatusChange } from '@/utils/loginStatus.js';
import { requestWithAutoRelogin } from '@/utils/autoRelogin.js';
import { showLoginDialog, performLogin } from '@/utils/platformLogin.js';
import { getTenantName, getTenantOptions } from '@/utils/tenant.js';
const logsStore = useLogsStore();
const route = useRoute();

const FACTORY_URL = '/api/blade-iot/factoryDataQuery/page';
const LOGIN_URL = '/iot-api/api/blade-auth/oauth/token';

// 登录请求头（完整配置，使用代理）
const loginHeaders = {
  'Accept': 'application/json, text/plain, */*',
  'Accept-Encoding': 'gzip, deflate',
  'Accept-Language': 'zh-CN,zh;q=0.9',
  'Authorization': 'Basic ' + btoa('saber:saber_secret'),
  'Cache-Control': 'no-cache',
  'Captcha-Code': '',
  'Captcha-Key': '',
  'Connection': 'keep-alive',
  'Host': '218.90.146.230:20001',
  'Origin': 'http://218.90.146.230:20001',
  'Pragma': 'no-cache',
  'Referrer': 'http://218.90.146.230:20001/',
  'Tenant-Id': '000000',
  'User-Agent': 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/149.0.0.0 Safari/537.36'
};

const searchForm = reactive({
  deviceType: 'D',
  deviceTypeText: '一测',
  modeTypeDiy: 'batchNo',
  modeTypeDiyText: '批次号',
  mode: '',
  batchNo: '',
  idHex: '',
  blastTenantId: '',
  current: 1,
  size: 10
});

const filterOptions = ref([
  { key: 'deviceType', label: '设备类型' },
  { key: 'modeTypeDiy', label: '查询模式' },
  { key: 'batchNo', label: '批次号' },
  { key: 'idHex', label: '低压ID' },
  { key: 'blastTenantId', label: '租户ID' },
]);

const filterChecked = reactive({
  deviceType: true,
  modeTypeDiy: true,
  batchNo: true,
  idHex: true,
  blastTenantId: false,
});

const deviceTypeOptions = [
  { value: 'D', label: '一测' },
  { value: 'B', label: '成测' },
  { value: 'R', label: '发火电阻' },
  { value: 'K', label: '快检' },
  { value: 'C', label: '注码' },
  { value: 'E', label: '二测' }
];

const modeTypeOptions = [
  { value: 'batchNo', label: '批次号' },
  { value: 'shellCode', label: '管壳码' },
  { value: 'uid', label: 'UID' }
];

const showFilterConfig = ref(false);
const searchResult = ref(null);
const isSearching = ref(false);
const selectedRecord = ref(null);

// 表格列定义
const tableColumns = [
  { key: 'deviceType', label: '设备类型', formatter: (val) => {
      const map = { 'D': '一测', 'B': '成测', 'R': '发火电阻', 'K': '快检', 'C': '注码', 'E': '二测' };
      return map[val] || (val || '-');
    }
  },
  { key: 'deviceCode', label: '设备编号', formatter: null },
  { key: 'batchNo', label: '批次号', formatter: null },
  { key: 'mode', label: '模式', formatter: (val) => val === 0 || val === '0' ? '正常' : val === 1 || val === '1' ? '重测' : (val || '-') },
  { key: 'idHex', label: 'ID', formatter: null },
  { key: 'innerCode', label: '内码', formatter: null },
  { key: 'operatorName', label: '操作人', formatter: null },
  { key: 'dataErrorName', label: '检测结果汇总', formatter: null },
  { key: 'checkTime', label: '测试时间', formatter: null },
  { key: 'lowCurrent', label: '低压电流', formatter: (val) => val !== null && val !== undefined ? (val / 100).toFixed(2) : '-' },
  { key: 'lowReverseCurrent', label: '低压反向电流', formatter: (val) => val !== null && val !== undefined ? (val / 100).toFixed(2) : '-' },
  { key: 'highCurrent', label: '高压电流', formatter: (val) => val !== null && val !== undefined ? (val / 100).toFixed(2) : '-' },
  { key: 'highReverseCurrent', label: '高压反向电流', formatter: (val) => val !== null && val !== undefined ? (val / 100).toFixed(2) : '-' },
  { key: 'bombLowPressure', label: '起爆低压', formatter: (val) => val !== null && val !== undefined ? (val / 100).toFixed(2) : '-' },
  { key: 'bombHighPressure', label: '起爆高压', formatter: (val) => val !== null && val !== undefined ? (val / 100).toFixed(2) : '-' },
  { key: 'diodeForwardElectricity', label: '二极管正向电流', formatter: (val) => val !== null && val !== undefined ? (val / 100).toFixed(2) : '-' },
  { key: 'diodeReverseElectricity', label: '二极管反向电流', formatter: (val) => val !== null && val !== undefined ? (val / 100).toFixed(2) : '-' },
  { key: 'reverseLowId', label: '反向低压ID', formatter: null },
  { key: 'lowFrequency', label: '低压频率', formatter: null },
  { key: 'highFrequency', label: '高压频率', formatter: null },
  { key: 'eepromAddressFrequency', label: '注码频率', formatter: null },
  { key: 'lowDelayPrecision', label: '低压延时精度', formatter: (val) => val !== null && val !== undefined ? (val / 100).toFixed(2) : '-' },
  { key: 'highDelayPrecision', label: '高压延时精度', formatter: (val) => val !== null && val !== undefined ? (val / 100).toFixed(2) : '-' },
  { key: 'address0', label: '地址0', formatter: null },
  { key: 'address12', label: '地址12', formatter: null },
  { key: 'chipName', label: '芯片类型', formatter: null },
  { key: 'moduleName', label: '模块类型', formatter: null },
  { key: 'lowDynamicCurrent', label: '低压动态电流', formatter: null },
  { key: 'highDynamicCurrent', label: '高压动态电流', formatter: null },
  { key: 'traceableCode', label: '追溯码', formatter: null }
];

const formatCellValue = (record, column) => {
  const value = record[column.key];
  if (value === null || value === undefined || value === '') return '-';
  if (column.formatter) return column.formatter(value);
  return value;
};

// 登录相关
const isLoggedIn = ref(false);
const isLogging = ref(false);
const accountInfo = ref(null);
const loginForm = reactive({
  username: '',
  password: ''
});
const loginMessage = ref('');
const loginSuccess = ref(false);

// localStorage key
const LOGIN_STORAGE_KEY = 'factory_data_login_record';

// 搜索历史相关
const searchHistory = ref([]);
const showSearchHistory = ref(false);
const MAX_HISTORY_COUNT = 10;
const HISTORY_STORAGE_KEY = 'factory_data_search_history';

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
  if (opt) return opt.label;
  const modeOpt = modeTypeOptions.find(o => o.value === key);
  if (modeOpt) return modeOpt.label;
  if (key === 'tenantId') return '管厂/租户';
  return key;
};

const applySearchHistory = (item) => {
  handleReset();
  Object.keys(item.params).forEach(key => {
    if (searchForm.hasOwnProperty(key)) {
      searchForm[key] = item.params[key];
    }
  });
  if (searchForm.deviceType) {
    handleDeviceTypeChange(searchForm.deviceType);
  }
  if (searchForm.modeTypeDiy) {
    handleModeTypeChange(searchForm.modeTypeDiy);
  }
  showSearchHistory.value = false;
  handleSearch();
};

const clearSearchHistory = () => {
  searchHistory.value = [];
  localStorage.removeItem(HISTORY_STORAGE_KEY);
  ElMessage.success('搜索历史已清除');
};

const toggleFilter = (key) => {
  if (filterChecked.hasOwnProperty(key)) {
    filterChecked[key] = !filterChecked[key];
  }
};

const isFilterVisible = (key) => {
  return filterChecked[key] === true;
};

// 加载保存的登录记录（30天有效，兼容新旧两种 key）
const loadLoginRecord = () => {
  try {
    // 优先检查新的统一凭据 key
    const unifiedStr = localStorage.getItem('third_party_credentials');
    if (unifiedStr) {
      const all = JSON.parse(unifiedStr);
      // iot 和 smart 凭据都可以识别为智能制造
      const cred = all.iot || all.smart;
      if (cred && cred.accessToken) {
        const now = Date.now();
        if (!cred.tokenExpire || now < cred.tokenExpire) {
          isLoggedIn.value = true;
          accountInfo.value = { access_token: cred.accessToken, ...cred };
          loginForm.username = cred.username || '';
          updateLoginStatus('smart', true);
          updateLoginStatus('iot', true);
          return true;
        }
      }
    }

    // 兼容旧的 localStorage key
    const recordStr = localStorage.getItem(LOGIN_STORAGE_KEY);
    if (recordStr) {
      const record = JSON.parse(recordStr);
      const now = Date.now();
      if (record.expireTime && now < record.expireTime) {
        // 30天内有效
        isLoggedIn.value = true;
        accountInfo.value = record.accountInfo;
        loginForm.username = record.username || '';

        // 同步登录状态到顶部菜单栏（智能制造平台）
        updateLoginStatus('smart', true);

        return true;
      } else {
        // 过期清除
        localStorage.removeItem(LOGIN_STORAGE_KEY);
      }
    }
  } catch (e) {
    console.error('加载登录记录失败:', e);
  }
  return false;
};

// 保存登录记录（30天有效）
const saveLoginRecord = (account) => {
  try {
    const now = Date.now();
    const expireTime = now + 30 * 24 * 60 * 60 * 1000; // 30天
    const record = {
      username: loginForm.username,
      accountInfo: account,
      loginTime: now,
      expireTime: expireTime
    };
    localStorage.setItem(LOGIN_STORAGE_KEY, JSON.stringify(record));
  } catch (e) {
    console.error('保存登录记录失败:', e);
  }
};

// 登录智能制造平台
const handleLogin = async () => {
  if (!loginForm.username || !loginForm.password) {
    loginMessage.value = '请输入用户名和密码';
    loginSuccess.value = false;
    return;
  }
  
  isLogging.value = true;
  loginMessage.value = '';
  
  try {
    const result = await performLogin('iot', loginForm.username, loginForm.password);
    
    if (result.success) {
      loginSuccess.value = true;
      loginMessage.value = '登录成功！';
      isLoggedIn.value = true;
      accountInfo.value = result.data;
      saveLoginRecord(result.data);

      // performLogin 已自动同步全局状态（updateLoginStatus、saveCredentials）
      // 额外保存 FactoryDataQuery 特有的登录记录
      saveCred('smart', {
        tenantId: '000000',
        username: loginForm.username,
        password: CryptoJS.MD5(loginForm.password).toString(),
        accessToken: result.data.access_token,
        tokenExpire: Date.now() + 30 * 24 * 60 * 60 * 1000
      });

      ElMessage.success('登录成功');

      if (searchForm.batchNo) {
        setTimeout(() => {
          handleSearch();
        }, 500);
      }
    } else {
      loginSuccess.value = false;
      loginMessage.value = result.message || '登录失败，请检查用户名和密码';
    }
  } catch (error) {
    loginSuccess.value = false;
    loginMessage.value = '登录请求失败：' + error.message;
  } finally {
    isLogging.value = false;
  }
};

// 退出登录
const handleLogout = () => {
  isLoggedIn.value = false;
  accountInfo.value = null;
  loginForm.username = '';
  loginForm.password = '';
  loginMessage.value = '';
  loginSuccess.value = false;
  searchResult.value = null;
  localStorage.removeItem(LOGIN_STORAGE_KEY);

  // 同步清除登录状态
  updateLoginStatus('smart', false);
  clearCred('smart');

  ElMessage.success('已退出登录');
};

const handleSearch = async () => {
  // 检查登录状态：未登录则弹出登录对话框
  if (!isLoggedIn.value) {
    // 总是弹出登录对话框（即使没有最近凭据）
    const result = await showLoginDialog('iot');
    if (!result || !result.success) {
      // 用户取消登录，不做任何操作
      return;
    }
    // 登录成功，更新本地状态
    isLoggedIn.value = true;
    accountInfo.value = { access_token: result.token || '' };
  }

  isSearching.value = true;
  searchResult.value = null;
  try {
    const params = new URLSearchParams();
    if (searchForm.deviceType) {
      params.append('deviceType', searchForm.deviceType);
      params.append('$deviceType', searchForm.deviceTypeText);
    }
    if (searchForm.modeTypeDiy) {
      params.append('modeTypeDiy', searchForm.modeTypeDiy);
      params.append('$modeTypeDiy', searchForm.modeTypeDiyText);
    }
    if (searchForm.mode) {
      params.append('$mode', searchForm.mode);
    }
    if (searchForm.batchNo) {
      params.append('batchNo', searchForm.batchNo);
    }
    if (searchForm.idHex) {
      params.append('idHex', searchForm.idHex);
    }
    if (searchForm.blastTenantId) {
      params.append('blastTenantId', searchForm.blastTenantId);
    }
    params.append('current', searchForm.current);
    params.append('size', searchForm.size);

    // 使用401自动重新登录的请求函数
    const { response, data } = await requestWithAutoRelogin('iot', {
      url: `${FACTORY_URL}?${params.toString()}`,
      method: 'GET',
      headers: {
        'Accept': 'application/json, text/plain, */*',
        'Accept-Encoding': 'gzip, deflate',
        'Accept-Language': 'zh-CN,zh;q=0.9',
        'Authorization': 'Basic ' + btoa('saber_identity_client:saber_identity_secret'),
        'Cache-Control': 'no-cache',
        'Connection': 'keep-alive',
        'Host': '218.90.146.230:20001',
        'Origin': 'http://218.90.146.230:20001',
        'Pragma': 'no-cache',
        'Referrer': 'http://218.90.146.230:20001/',
        'Tenant-Id': '000000',
        'User-Agent': 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/149.0.0.0 Safari/537.36',
        'Blade-Auth': accountInfo.value ? `bearer ${accountInfo.value.access_token}` : ''
      }
    });
    if (data && data.code === 200 && data.data && data.data.records) {
      searchResult.value = {
        success: true,
        data: data.data,
        records: data.data.records,
        total: data.data.total,
        current: data.data.current,
        size: data.data.size,
        pages: data.data.pages
      };
      ElMessage.success(`查询成功，共 ${data.data.total} 条记录`);
      logsStore.addLog('查询', '工厂数据', `设备类型: ${searchForm.deviceType}, 批次号: ${searchForm.batchNo}, 共${data.data.total}条`);
      saveSearchHistory();
    }
    else if (response && response.status === 401) {
      // 自动重登失败
      searchResult.value = { success: false, error: '登录已失效，请重新登录' };
      ElMessage.error('登录已失效，请重新登录');
    }
    else {
      searchResult.value = { success: false, error: (data && data.msg) || '未查询到数据' };
      ElMessage.warning((data && data.msg) || '未查询到数据');
    }
  }
  catch (error) {
    searchResult.value = { success: false, error: error.message };
    ElMessage.error('网络连接失败');
    logsStore.addLog('查询', '工厂数据', `网络错误: ${error.message}`);
  }
  finally {
    isSearching.value = false;
  }
};

const showRecordDetail = (record) => {
  selectedRecord.value = record;
};

const handleReset = () => {
  Object.keys(searchForm).forEach(key => {
    if (key === 'current') {
      searchForm[key] = 1;
    } else if (key === 'size') {
      searchForm[key] = 10;
    } else {
      searchForm[key] = '';
    }
  });
  searchResult.value = null;
};

const handleDeviceTypeChange = (val) => {
  const opt = deviceTypeOptions.find(o => o.value === val);
  searchForm.deviceTypeText = opt ? opt.label : '';
};

const handleModeTypeChange = (val) => {
  const opt = modeTypeOptions.find(o => o.value === val);
  searchForm.modeTypeDiyText = opt ? opt.label : '';
};

const handlePageChange = (page) => {
  if (page < 1 || page > searchResult.value.pages) return;
  searchForm.current = page;
  handleSearch();
};

const handleSizeChange = (val) => {
  searchForm.size = val;
  searchForm.current = 1;
  handleSearch();
};

const displayedPages = computed(() => {
  if (!searchResult.value) return [];
  const current = searchResult.value.current;
  const pages = searchResult.value.pages;
  const visiblePages = 5;
  const result = [];
  
  if (pages <= visiblePages) {
    for (let i = 1; i <= pages; i++) {
      result.push(i);
    }
  } else {
    if (current <= 3) {
      for (let i = 1; i <= 4; i++) {
        result.push(i);
      }
    } else if (current >= pages - 2) {
      for (let i = pages - 3; i <= pages; i++) {
        result.push(i);
      }
    } else {
      for (let i = current - 2; i <= current + 2; i++) {
        result.push(i);
      }
    }
  }
  return result;
});

const showEllipsisBefore = computed(() => {
  if (!searchResult.value) return false;
  const current = searchResult.value.current;
  const pages = searchResult.value.pages;
  return pages > 5 && current > 4;
});

const showEllipsisAfter = computed(() => {
  if (!searchResult.value) return false;
  const current = searchResult.value.current;
  const pages = searchResult.value.pages;
  return pages > 5 && current < pages - 3;
});

const handleExportExcel = () => {
  if (!searchResult.value || searchResult.value.records.length === 0) {
    ElMessage.warning('没有数据可导出');
    return;
  }
  
  const records = searchResult.value.records;
  const columns = tableColumns;
  
  let html = '<html xmlns:o="urn:schemas-microsoft-com:office:office" xmlns:x="urn:schemas-microsoft-com:office:excel" xmlns="http://www.w3.org/TR/REC-html40">';
  html += '<head><meta charset="utf-8"><style>';
  html += 'td { mso-number-format:"\\@"; }';
  html += '</style></head><body>';
  html += '<table border="1">';
  
  html += '<tr>';
  html += '<td>序号</td>';
  columns.forEach(col => {
    html += `<td>${col.label}</td>`;
  });
  html += '</tr>';
  
  records.forEach((row, index) => {
    html += '<tr>';
    html += `<td>${(searchForm.current - 1) * searchForm.size + index + 1}</td>`;
    columns.forEach(col => {
      const cellValue = formatCellValue(row, col);
      html += `<td style="mso-number-format:'\\@'">${cellValue}</td>`;
    });
    html += '</tr>';
  });
  
  html += '</table></body></html>';
  
  const blob = new Blob([html], { type: 'application/vnd.ms-excel;charset=utf-8' });
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  const now = new Date();
  const dateStr = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-${String(now.getDate()).padStart(2, '0')}_${String(now.getHours()).padStart(2, '0')}${String(now.getMinutes()).padStart(2, '0')}`;
  link.download = `工厂数据_${dateStr}.xls`;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  URL.revokeObjectURL(url);
  
  ElMessage.success('导出Excel成功');
};

// 实时同步全局登录状态
let unsubscribeStatus = null;
const handleGlobalStatusChange = (platform) => {
  if (platform === 'smart' || platform === 'iot') {
    loadLoginRecord();
  }
};

onMounted(() => {
  loadLoginRecord();
  loadSearchHistory();

  // 订阅全局登录状态变化，实时同步到本地
  unsubscribeStatus = onLoginStatusChange(handleGlobalStatusChange);
  // watch 双向同步（兜底）
  watch(() => loginStatus.smart, (newVal) => {
    if (newVal && !isLoggedIn.value) {
      loadLoginRecord();
    } else if (!newVal && isLoggedIn.value) {
      isLoggedIn.value = false;
      accountInfo.value = null;
    }
  });
  watch(() => loginStatus.iot, (newVal) => {
    if (newVal && !isLoggedIn.value) {
      loadLoginRecord();
    }
  });

  // 处理 URL 参数（从首页跳转过来）
  if (route.query.batchNo) {
    searchForm.batchNo = route.query.batchNo;
  }
  if (route.query.idHex) {
    searchForm.idHex = route.query.idHex;
    // 确保低压ID筛选项可见
    filterChecked.idHex = true;
  }
  if (route.query.deviceType) {
    searchForm.deviceType = route.query.deviceType;
    handleDeviceTypeChange(route.query.deviceType);
  }
  if (route.query.deviceTypeName) {
    searchForm.deviceTypeText = route.query.deviceTypeName;
  }

  const autoQuery = route.query.autoQuery === 'true' || route.query.autoQuery === true;

  if ((route.query.batchNo || route.query.deviceType) && autoQuery) {
    if (isLoggedIn.value) {
      ElMessage.success('已自动填充查询参数');
      setTimeout(() => {
        handleSearch();
      }, 500);
    } else {
      // 未登录时弹出登录对话框，登录成功后会自动执行查询
      ElMessage.warning('请先登录，登录成功后将自动查询数据');
      setTimeout(() => {
        handleSearch();
      }, 500);
    }
  }
});

onUnmounted(() => {
  if (unsubscribeStatus) {
    unsubscribeStatus();
  }
});

watch(() => route.query.batchNo, (newBatchNo) => {
  if (newBatchNo) {
    searchForm.batchNo = newBatchNo;
    searchForm.current = 1;
    if (route.query.deviceType) {
      searchForm.deviceType = route.query.deviceType;
      handleDeviceTypeChange(route.query.deviceType);
    }
    if (isLoggedIn.value) {
      handleSearch();
    }
  }
});

watch(() => route.query.idHex, (newIdHex) => {
  if (newIdHex) {
    searchForm.idHex = newIdHex;
    searchForm.current = 1;
    filterChecked.idHex = true;
    if (isLoggedIn.value) {
      handleSearch();
    }
  }
});
</script>

<template>
  <div class="factory-query">
    <GlassCard title="智能制造系统">
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
            <button class="clear-history-btn" @click.stop="clearSearchHistory">
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
          <div v-if="isFilterVisible('deviceType')" class="form-group">
            <label>设备类型</label>
            <select 
              v-model="searchForm.deviceType" 
              class="form-input"
              @change="handleDeviceTypeChange($event.target.value)"
            >
              <option value="">全部</option>
              <option v-for="opt in deviceTypeOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
            </select>
          </div>
          
          <div v-if="isFilterVisible('modeTypeDiy')" class="form-group">
            <label>查询模式</label>
            <select 
              v-model="searchForm.modeTypeDiy" 
              class="form-input"
              @change="handleModeTypeChange($event.target.value)"
            >
              <option value="">全部</option>
              <option v-for="opt in modeTypeOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
            </select>
          </div>
          
          <div v-if="isFilterVisible('batchNo')" class="form-group">
            <label>批次号</label>
            <input 
              v-model="searchForm.batchNo"
              type="text"
              class="form-input"
              placeholder="请输入批次号"
            />
          </div>
          
          <div v-if="isFilterVisible('idHex')" class="form-group">
            <label>低压ID</label>
            <input 
              v-model="searchForm.idHex"
              type="text"
              class="form-input"
              placeholder="请输入低压ID"
            />
          </div>
          
          <div v-if="isFilterVisible('blastTenantId')" class="form-group">
            <label>租户ID</label>
            <input
              v-model="searchForm.blastTenantId"
              type="text"
              class="form-input"
              placeholder="请输入租户ID"
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
        </div>
        
        <div class="form-actions">
          <button 
            class="action-btn primary"
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
            class="action-btn success"
            :disabled="!isLoggedIn || !searchResult || !searchResult.records || searchResult.records.length === 0"
            @click="handleExportExcel"
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
          <span>{{ searchResult.success ? '查询成功' : '查询失败' }}</span>
        </div>
        <pre v-if="!searchResult.success" class="result-content">{{ JSON.stringify(searchResult.error, null, 2) }}</pre>
      </div>
      
      <div v-if="searchResult?.success && searchResult.records" class="records-panel">
        <div class="records-header">
          <span>数据列表</span>
          <span class="pagination-info">共 {{ searchResult.total }} 条 | 第 {{ searchResult.current }} 页 | 每页 {{ searchResult.size }} 条</span>
        </div>
        
        <div class="table-container">
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
        
        <div class="pagination-container">
          <div class="pagination">
            <button 
              class="page-btn"
              :disabled="searchResult.current <= 1"
              @click="handlePageChange(searchResult.current - 1)"
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
                :class="{ active: page === searchResult.current }"
                @click="handlePageChange(page)"
              >
                {{ page }}
              </button>
              <span v-if="showEllipsisBefore" class="ellipsis">...</span>
              <span v-if="showEllipsisAfter" class="ellipsis">...</span>
            </div>
            
            <button 
              class="page-btn"
              :disabled="searchResult.current >= searchResult.pages"
              @click="handlePageChange(searchResult.current + 1)"
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
    </GlassCard>
  </div>
</template>

<style scoped>
.factory-query {
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

.tenant-name-hint {
  font-size: 11px;
  color: #165DFF;
  padding: 2px 4px;
  background: rgba(22, 93, 255, 0.06);
  border-radius: 4px;
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
  
  &.success {
    background: linear-gradient(135deg, #00B42A 0%, #009A29 100%);
    color: #FFFFFF;
    
    &:hover:not(:disabled) {
      transform: translateY(-2px);
      box-shadow: 0 4px 12px rgba(0, 180, 42, 0.3);
    }
  }
  
  &.small {
    padding: 4px 12px;
    font-size: 12px;
  }
  
  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
}

.login-section {
  background: rgba(22, 93, 255, 0.05);
  border: 1px solid rgba(22, 93, 255, 0.2);
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 20px;
}

.login-form {
  display: flex;
  align-items: flex-end;
  gap: 12px;
  flex-wrap: wrap;
  
  .form-group {
    flex: 1;
    min-width: 180px;
  }
  
  .form-input {
    width: 100%;
  }
}

.logged-section {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: rgba(0, 180, 42, 0.05);
  border: 1px solid rgba(0, 180, 42, 0.2);
  border-radius: 8px;
  padding: 10px 16px;
  margin-bottom: 20px;
}

.logged-info {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #00B42A;
  font-size: 13px;
  font-weight: 500;
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
  background: rgba(0, 0, 0, 0.05);
  padding: 12px;
  border-radius: 6px;
  overflow-x: auto;
  white-space: pre-wrap;
  word-break: break-all;
  color: #64748B;
}

.records-panel {
  margin-top: 20px;
}

.records-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  font-size: 14px;
  font-weight: 600;
  color: #1E293B;
  flex-wrap: wrap;
  gap: 12px;
}

.pagination-info {
  font-size: 12px;
  color: #64748B;
  font-weight: normal;
}

.table-container {
  overflow-x: auto;
  overflow-y: hidden;
  background: #FAFAFA;
  border-radius: 8px;
  border: 1px solid #E2E8F0;
  max-width: 100%;
  width: 100%;
  margin-bottom: 16px;
  display: block;
}

.trace-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
  table-layout: auto;
  min-width: 100%;
}

.trace-table thead {
  background: rgba(22, 93, 255, 0.1);
}

.trace-table th {
  padding: 12px 12px;
  text-align: left;
  font-weight: 600;
  color: #1E293B;
  border-bottom: 1px solid #E2E8F0;
  white-space: nowrap;
  min-width: 80px;
  position: sticky;
  top: 0;
}

.trace-table td {
  padding: 10px 12px;
  border-bottom: 1px solid rgba(224, 232, 240, 0.5);
  color: #64748B;
  vertical-align: top;
  white-space: nowrap;
  min-width: 80px;
}

.trace-table tbody tr:hover {
  background: rgba(22, 93, 255, 0.05);
}

.cell-td {
  position: relative;
}

.cell-content {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 250px;
  cursor: pointer;
  transition: all 0.2s;
}

.cell-content:hover {
  overflow: visible;
  text-overflow: unset;
  background: #FFFFFF;
  padding: 6px 8px;
  border-radius: 4px;
  z-index: 100;
  position: absolute;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  min-width: 200px;
  max-width: 500px;
  border: 1px solid #165DFF;
  white-space: normal;
  word-break: break-all;
}

.pagination-container {
  display: flex;
  justify-content: center;
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid #E2E8F0;
  flex-wrap: wrap;
  gap: 8px;
}

.pagination {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  justify-content: center;
}

.page-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
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

.message {
  width: 100%;
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
  
  .login-form {
    flex-direction: column;
    align-items: stretch;
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
