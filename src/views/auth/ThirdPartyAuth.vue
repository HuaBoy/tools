<script setup>
import { ref, reactive, computed, onMounted, onUnmounted, watch } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import GlassCard from '@/components/GlassCard.vue';
import CryptoJS from 'crypto-js';
import { updateLoginStatus, saveCredentials as saveUnifiedCredentials, loginStatus, onLoginStatusChange } from '@/utils/loginStatus.js';
import { performLogin } from '@/utils/platformLogin.js';
import { platformApis } from '@/data/platformApis.js';

// 接口说明区块：当前选中的平台
const activeApiDocPlatform = ref('mp');
const activeApiDoc = computed(() => platformApis.find(p => p.id === activeApiDocPlatform.value) || platformApis[0]);

const API_LOGIN_URL = 'https://mp.holyview.cn:9443/api/blade-auth/oauth/token';
const FACTORY_LOGIN_URL = '/iot-api/api/blade-auth/oauth/token';
const API_AUTH_HEADER = 'Basic ' + btoa('saber:saber_secret');
const FACTORY_AUTH_HEADER = 'Basic ' + btoa('saber_identity_client:saber_identity_secret');

const apiLoginForm = reactive({
  username: '',
  password: ''
});
const factoryLoginForm = reactive({
  username: '',
  password: ''
});

const apiIsLogging = ref(false);
const factoryIsLogging = ref(false);
const apiLoginMessage = ref('');
const factoryLoginMessage = ref('');
const apiIsLoggedIn = ref(false);
const factoryIsLoggedIn = ref(false);
const apiAccountInfo = ref(null);
const factoryAccountInfo = ref(null);

const apiShowAccountDropdown = ref(false);
const factoryShowAccountDropdown = ref(false);

const DEEPSEEK_KEY = 'deepseek_api_key';
const deepseekApiKey = ref(localStorage.getItem(DEEPSEEK_KEY) || '');
const deepseekKeySaved = ref(false);
const deepseekTestResult = ref(null);
const deepseekIsTesting = ref(false);
const deepseekShowKey = ref(false);

// Ollama 本地模型配置
const ollamaUrl = ref(localStorage.getItem('ai_ollama_url') || 'http://localhost:11434');
const ollamaConnected = ref(false);
const ollamaIsTesting = ref(false);
const ollamaTestResult = ref(null);
const ollamaModelList = ref([]);
const ollamaSelectedModel = ref(localStorage.getItem('ai_ollama_model') || 'deepseek-r1:7b');

const handleSaveOllamaUrl = () => {
  let url = ollamaUrl.value.trim().replace(/\/+$/, '');
  if (!url) url = 'http://localhost:11434';
  ollamaUrl.value = url;
  localStorage.setItem('ai_ollama_url', url);
  localStorage.setItem('ai_ollama_model', ollamaSelectedModel.value);
};

const handleTestOllama = async () => {
  ollamaIsTesting.value = true;
  ollamaTestResult.value = null;
  handleSaveOllamaUrl();
  try {
    // 1. 检查 Ollama 服务是否可达
    const tagsRes = await fetch(ollamaUrl.value + '/api/tags', { signal: AbortSignal.timeout(5000) });
    if (!tagsRes.ok) {
      ollamaTestResult.value = { success: false, message: `Ollama 服务响应异常 (${tagsRes.status})` };
      return;
    }
    const tagsData = await tagsRes.json();
    ollamaModelList.value = (tagsData.models || []).map(m => m.name);
    if (!ollamaModelList.value.length) {
      ollamaTestResult.value = { success: false, message: 'Ollama 已连接，但没有已安装的模型。请先运行: ollama pull deepseek-r1:7b' };
      return;
    }
    // 如果当前选中模型不在列表中，自动选第一个
    if (!ollamaModelList.value.includes(ollamaSelectedModel.value)) {
      ollamaSelectedModel.value = ollamaModelList.value[0];
    }
    // 2. 用简单请求测试模型推理能力
    const chatRes = await fetch(ollamaUrl.value + '/v1/chat/completions', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        model: ollamaSelectedModel.value,
        messages: [{ role: 'user', content: 'ping' }],
        max_tokens: 5
      }),
      signal: AbortSignal.timeout(15000)
    });
    if (chatRes.ok) {
      ollamaConnected.value = true;
      ollamaTestResult.value = { success: true, message: `连接成功！模型: ${ollamaSelectedModel.value}，共 ${ollamaModelList.value.length} 个模型可用` };
      ElMessage.success('Ollama 连接成功');
    } else {
      ollamaTestResult.value = { success: false, message: `模型调用失败 (${chatRes.status})，请确认模型已下载` };
    }
  } catch (e) {
    ollamaConnected.value = false;
    ollamaTestResult.value = { success: false, message: `连接失败: ${e.message}。请确认 Ollama 容器已启动` };
  } finally {
    ollamaIsTesting.value = false;
  }
};

// 页面加载时自动检测 Ollama 连接
const autoDetectOllama = async () => {
  try {
    const res = await fetch(ollamaUrl.value + '/api/tags', { signal: AbortSignal.timeout(3000) });
    if (res.ok) {
      const data = await res.json();
      ollamaModelList.value = (data.models || []).map(m => m.name);
      ollamaConnected.value = ollamaModelList.value.length > 0;
    }
  } catch { /* 静默 */ }
};

const handleSaveDeepseekKey = () => {
  if (!deepseekApiKey.value.trim()) {
    ElMessage.warning('请输入 DeepSeek API Key');
    return;
  }
  localStorage.setItem(DEEPSEEK_KEY, deepseekApiKey.value.trim());
  deepseekKeySaved.value = true;
  ElMessage.success('DeepSeek API Key 已保存');
  setTimeout(() => { deepseekKeySaved.value = false; }, 3000);
};

const handleClearDeepseekKey = () => {
  deepseekApiKey.value = '';
  localStorage.removeItem(DEEPSEEK_KEY);
  deepseekTestResult.value = null;
  ElMessage.info('DeepSeek API Key 已清除');
};

const handleTestDeepseekKey = async () => {
  if (!deepseekApiKey.value.trim()) {
    ElMessage.warning('请先输入 API Key');
    return;
  }
  deepseekIsTesting.value = true;
  deepseekTestResult.value = null;
  try {
    const response = await fetch('https://api.deepseek.com/v1/models', {
      headers: { 'Authorization': `Bearer ${deepseekApiKey.value.trim()}` }
    });
    if (response.ok) {
      deepseekTestResult.value = { success: true, message: '连接成功！API Key 有效' };
      ElMessage.success('DeepSeek API Key 验证通过');
    } else {
      const err = await response.json().catch(() => ({}));
      deepseekTestResult.value = { success: false, message: `验证失败 (${response.status}): ${err.error?.message || '请检查 API Key'}` };
    }
  } catch (e) {
    deepseekTestResult.value = { success: false, message: `网络错误: ${e.message}` };
  } finally {
    deepseekIsTesting.value = false;
  }
};

const API_STORAGE_KEY = 'tester_credentials';
const FACTORY_STORAGE_KEY = 'factory_data_login_record';
const API_ACCOUNTS_LIST_KEY = 'api_accounts_list';
const FACTORY_ACCOUNTS_LIST_KEY = 'factory_accounts_list';
const UNIFIED_LATEST_KEY = 'unified_latest_login';

const apiAccountsList = ref([]);
const factoryAccountsList = ref([]);

// 跨平台最新登录信息（用于顶部展示）
const unifiedLatest = ref(null);

// 跨平台最新登录账号（兼容两个平台）
const latestAccount = computed(() => {
  if (unifiedLatest.value) return unifiedLatest.value;
  // 兼容旧数据：从两个列表中找最新
  const all = [
    ...apiAccountsList.value.map(a => ({ ...a, platform: 'mp' })),
    ...factoryAccountsList.value.map(a => ({ ...a, platform: 'smart' }))
  ];
  if (all.length === 0) return null;
  return all.sort((a, b) => b.lastUsedTime - a.lastUsedTime)[0];
});

const apiRecentAccount = computed(() => {
  if (apiAccountsList.value.length === 0) return null;
  return [...apiAccountsList.value].sort((a, b) => b.lastUsedTime - a.lastUsedTime)[0];
});

const factoryRecentAccount = computed(() => {
  if (factoryAccountsList.value.length === 0) return null;
  return [...factoryAccountsList.value].sort((a, b) => b.lastUsedTime - a.lastUsedTime)[0];
});

const loadAccountsList = (key) => {
  try {
    const list = localStorage.getItem(key);
    if (list) {
      return JSON.parse(list);
    }
  } catch (e) {
    console.error('加载账号列表失败:', e);
  }
  return [];
};

const saveAccountsList = (key, list) => {
  try {
    localStorage.setItem(key, JSON.stringify(list));
  } catch (e) {
    console.error('保存账号列表失败:', e);
  }
};

const addToAccountsList = (list, username, password) => {
  const now = Date.now();
  const existingIndex = list.findIndex(item => item.username === username);
  if (existingIndex > -1) {
    list[existingIndex].password = password;
    list[existingIndex].lastUsedTime = now;
  } else {
    list.push({
      username,
      password,
      lastUsedTime: now
    });
  }
  list.sort((a, b) => b.lastUsedTime - a.lastUsedTime);
  if (list.length > 10) {
    list.splice(10);
  }
  return list;
};

const removeFromAccountsList = (list, username) => {
  const index = list.findIndex(item => item.username === username);
  if (index > -1) {
    list.splice(index, 1);
  }
  return list;
};

// 保存跨平台最新登录信息
const saveUnifiedLatest = (username, password, platform) => {
  const data = {
    username,
    password,
    platform,
    lastUsedTime: Date.now()
  };
  unifiedLatest.value = data;
  localStorage.setItem(UNIFIED_LATEST_KEY, JSON.stringify(data));
};

const loadUnifiedLatest = () => {
  try {
    const data = localStorage.getItem(UNIFIED_LATEST_KEY);
    if (data) {
      unifiedLatest.value = JSON.parse(data);
    }
  } catch (e) {
    console.error('加载跨平台最新登录信息失败:', e);
  }
};

const loadApiLoginRecord = () => {
  try {
    // 优先检查新的统一凭据 key
    const unifiedCredStr = localStorage.getItem('third_party_credentials');
    if (unifiedCredStr) {
      let allCred;
      try {
        allCred = JSON.parse(unifiedCredStr);
      } catch (e) {
        // 兼容的JSON格式错误，清除脏数据
        localStorage.removeItem('third_party_credentials');
        allCred = {};
      }
      // 确保是对象
      if (typeof allCred !== 'object' || allCred === null) {
        localStorage.removeItem('third_party_credentials');
        allCred = {};
      }
      const mpCred = allCred.mp;
      if (mpCred && mpCred.accessToken) {
        const now = Date.now();
        if (!mpCred.tokenExpire || now < mpCred.tokenExpire) {
          apiIsLoggedIn.value = true;
          apiAccountInfo.value = { access_token: mpCred.accessToken, ...mpCred };
          apiLoginForm.username = mpCred.username || '';
          apiLoginForm.password = '';  // 密码不保存在统一凭据中
          // 不再调用 updateLoginStatus，避免无限循环
          return true;
        }
      }
    }

    // 兼容旧的 localStorage key
    const recordStr = localStorage.getItem(API_STORAGE_KEY);
    if (recordStr) {
      let record;
      try {
        record = JSON.parse(recordStr);
      } catch (e) {
        localStorage.removeItem(API_STORAGE_KEY);
        record = null;
      }
      if (record) {
        const now = Date.now();
        if (record.expireTime && now < record.expireTime) {
          apiIsLoggedIn.value = true;
          apiAccountInfo.value = record.accountInfo;
          apiLoginForm.username = record.username || '';
          apiLoginForm.password = record.password || '';
          // 不再调用 updateLoginStatus，避免无限循环
          return true;
        } else {
          // Token 过期，但保留账号密码供自动填充
          if (record.username) {
            apiLoginForm.username = record.username;
          }
          if (record.password) {
            apiLoginForm.password = record.password;
          }
          if (!apiIsLoggedIn.value) {
            localStorage.removeItem(API_STORAGE_KEY);
          }
        }
      }
    }
  } catch (e) {
    console.error('加载API登录记录失败:', e);
  }
  return false;
};

const loadFactoryLoginRecord = () => {
  try {
    // 优先检查新的统一凭据 key
    const unifiedCredStr = localStorage.getItem('third_party_credentials');
    if (unifiedCredStr) {
      let allCred;
      try {
        allCred = JSON.parse(unifiedCredStr);
      } catch (e) {
        localStorage.removeItem('third_party_credentials');
        allCred = {};
      }
      if (typeof allCred !== 'object' || allCred === null) {
        localStorage.removeItem('third_party_credentials');
        allCred = {};
      }
      // 智能制造可以是 smart 或 iot 凭据
      const smartCred = allCred.smart || allCred.iot;
      if (smartCred && smartCred.accessToken) {
        const now = Date.now();
        if (!smartCred.tokenExpire || now < smartCred.tokenExpire) {
          factoryIsLoggedIn.value = true;
          factoryAccountInfo.value = { access_token: smartCred.accessToken, ...smartCred };
          factoryLoginForm.username = smartCred.username || '';
          factoryLoginForm.password = '';
          // 不再调用 updateLoginStatus，避免无限循环
          return true;
        }
      }
    }

    // 兼容旧的 localStorage key
    const recordStr = localStorage.getItem(FACTORY_STORAGE_KEY);
    if (recordStr) {
      let record;
      try {
        record = JSON.parse(recordStr);
      } catch (e) {
        localStorage.removeItem(FACTORY_STORAGE_KEY);
        record = null;
      }
      if (record) {
        const now = Date.now();
        if (record.expireTime && now < record.expireTime) {
          factoryIsLoggedIn.value = true;
          factoryAccountInfo.value = record.accountInfo;
          factoryLoginForm.username = record.username || '';
          factoryLoginForm.password = record.password || '';
          // 不再调用 updateLoginStatus，避免无限循环
          return true;
        } else {
          // Token 过期，但保留账号密码供自动填充
          if (record.username) {
            factoryLoginForm.username = record.username;
          }
          if (record.password) {
            factoryLoginForm.password = record.password;
          }
          if (!factoryIsLoggedIn.value) {
            localStorage.removeItem(FACTORY_STORAGE_KEY);
          }
        }
      }
    }
  } catch (e) {
    console.error('加载工厂登录记录失败:', e);
  }
  return false;
};

const saveApiLoginRecord = (account) => {
  const now = Date.now();
  const expireTime = now + 30 * 24 * 60 * 60 * 1000;
  const record = {
    username: apiLoginForm.username,
    password: apiLoginForm.password,
    accessToken: account.access_token,
    accountInfo: account,
    expireTime: expireTime
  };
  localStorage.setItem(API_STORAGE_KEY, JSON.stringify(record));
  apiAccountsList.value = addToAccountsList(apiAccountsList.value, apiLoginForm.username, apiLoginForm.password);
  saveAccountsList(API_ACCOUNTS_LIST_KEY, apiAccountsList.value);
  // 同时保存到跨平台最新
  saveUnifiedLatest(apiLoginForm.username, apiLoginForm.password, 'mp');
};

const saveFactoryLoginRecord = (account) => {
  const now = Date.now();
  const expireTime = now + 30 * 24 * 60 * 60 * 1000;
  const record = {
    username: factoryLoginForm.username,
    password: factoryLoginForm.password,
    accountInfo: account,
    loginTime: now,
    expireTime: expireTime
  };
  localStorage.setItem(FACTORY_STORAGE_KEY, JSON.stringify(record));
  factoryAccountsList.value = addToAccountsList(factoryAccountsList.value, factoryLoginForm.username, factoryLoginForm.password);
  saveAccountsList(FACTORY_ACCOUNTS_LIST_KEY, factoryAccountsList.value);
  // 同时保存到跨平台最新
  saveUnifiedLatest(factoryLoginForm.username, factoryLoginForm.password, 'smart');
};

const handleApiLogin = async () => {
  if (!apiLoginForm.username || !apiLoginForm.password) {
    apiLoginMessage.value = '请输入用户名和密码';
    return;
  }
  apiIsLogging.value = true;
  apiLoginMessage.value = '';
  try {
    const result = await performLogin('mp', apiLoginForm.username, apiLoginForm.password);
    if (result.success) {
      apiIsLoggedIn.value = true;
      apiAccountInfo.value = result.data;
      saveApiLoginRecord(result.data);
      apiLoginMessage.value = '登录成功！';
      // performLogin 已自动调用 updateLoginStatus('mp', true) 和 saveCredentials
      ElMessage.success('云系统登录成功');
    } else {
      apiLoginMessage.value = result.message || '登录失败';
    }
  } catch (error) {
    apiLoginMessage.value = '登录请求失败：' + error.message;
  } finally {
    apiIsLogging.value = false;
  }
};

const handleFactoryLogin = async () => {
  if (!factoryLoginForm.username || !factoryLoginForm.password) {
    factoryLoginMessage.value = '请输入用户名和密码';
    return;
  }
  factoryIsLogging.value = true;
  factoryLoginMessage.value = '';
  try {
    const result = await performLogin('iot', factoryLoginForm.username, factoryLoginForm.password);
    if (result.success) {
      factoryIsLoggedIn.value = true;
      factoryAccountInfo.value = result.data;
      saveFactoryLoginRecord(result.data);
      factoryLoginMessage.value = '登录成功！';
      // performLogin 已自动同步 updateLoginStatus('smart', true) + saveCredentials
      ElMessage.success('智能制造系统登录成功');
    } else {
      factoryLoginMessage.value = result.message || '登录失败';
    }
  } catch (error) {
    factoryLoginMessage.value = '登录请求失败：' + error.message;
  } finally {
    factoryIsLogging.value = false;
  }
};

const handleApiLogout = () => {
  apiIsLoggedIn.value = false;
  apiAccountInfo.value = null;
  apiLoginMessage.value = '';
  try {
    // 清除新的统一凭据
    const unifiedStr = localStorage.getItem('third_party_credentials');
    if (unifiedStr) {
      const all = JSON.parse(unifiedStr);
      delete all.mp;
      localStorage.setItem('third_party_credentials', JSON.stringify(all));
    }
    // 清除 token
    localStorage.removeItem('mp_token');
    // 标记旧的凭据为过期
    const recordStr = localStorage.getItem(API_STORAGE_KEY);
    if (recordStr) {
      const record = JSON.parse(recordStr);
      localStorage.setItem(API_STORAGE_KEY, JSON.stringify({
        ...record,
        expireTime: 0
      }));
    }
  } catch (e) {
    console.error('退出登录失败:', e);
  }
  updateLoginStatus('mp', false);
  ElMessage.success('云系统已退出');
};

const handleFactoryLogout = () => {
  factoryIsLoggedIn.value = false;
  factoryAccountInfo.value = null;
  factoryLoginMessage.value = '';
  try {
    // 清除新的统一凭据
    const unifiedStr = localStorage.getItem('third_party_credentials');
    if (unifiedStr) {
      const all = JSON.parse(unifiedStr);
      delete all.smart;
      delete all.iot;  // 兼容 iot 平台
      localStorage.setItem('third_party_credentials', JSON.stringify(all));
    }
    // 清除 token
    localStorage.removeItem('smart_factory_token');
    localStorage.removeItem('iot_token');
    // 标记旧的凭据为过期
    const recordStr = localStorage.getItem(FACTORY_STORAGE_KEY);
    if (recordStr) {
      const record = JSON.parse(recordStr);
      localStorage.setItem(FACTORY_STORAGE_KEY, JSON.stringify({
        ...record,
        expireTime: 0
      }));
    }
  } catch (e) {
    console.error('退出登录失败:', e);
  }
  updateLoginStatus('smart', false);
  ElMessage.success('智能制造系统已退出');
};

const handleSelectApiAccount = (account) => {
  if (account) {
    apiLoginForm.username = account.username;
    apiLoginForm.password = account.password;
    apiLoginMessage.value = '';
  }
};

const handleSelectFactoryAccount = (account) => {
  if (account) {
    factoryLoginForm.username = account.username;
    factoryLoginForm.password = account.password;
    factoryLoginMessage.value = '';
  }
};

const handleDeleteApiAccount = async (username, event) => {
  event.stopPropagation();
  try {
    await ElMessageBox.confirm(
      `确定要删除账号 "${username}" 吗？`,
      '删除确认',
      {
        confirmButtonText: '确定删除',
        cancelButtonText: '取消',
        type: 'warning'
      }
    );
    apiAccountsList.value = removeFromAccountsList(apiAccountsList.value, username);
    saveAccountsList(API_ACCOUNTS_LIST_KEY, apiAccountsList.value);
    ElMessage.success('账号已删除');
  } catch (e) {
    if (e !== 'cancel') {
      console.error(e);
    }
  }
};

const handleDeleteFactoryAccount = async (username, event) => {
  event.stopPropagation();
  try {
    await ElMessageBox.confirm(
      `确定要删除账号 "${username}" 吗？`,
      '删除确认',
      {
        confirmButtonText: '确定删除',
        cancelButtonText: '取消',
        type: 'warning'
      }
    );
    factoryAccountsList.value = removeFromAccountsList(factoryAccountsList.value, username);
    saveAccountsList(FACTORY_ACCOUNTS_LIST_KEY, factoryAccountsList.value);
    ElMessage.success('账号已删除');
  } catch (e) {
    if (e !== 'cancel') {
      console.error(e);
    }
  }
};

// 格式化最后使用时间
const formatLastUsed = (timestamp) => {
  const now = Date.now();
  const diff = now - timestamp;
  if (diff < 60 * 1000) return '刚刚';
  if (diff < 60 * 60 * 1000) return `${Math.floor(diff / (60 * 1000))} 分钟前`;
  if (diff < 24 * 60 * 60 * 1000) return `${Math.floor(diff / (60 * 60 * 1000))} 小时前`;
  if (diff < 7 * 24 * 60 * 60 * 1000) return `${Math.floor(diff / (24 * 60 * 60 * 1000))} 天前`;
  return new Date(timestamp).toLocaleDateString('zh-CN');
};

// 一键登录（直接调用登录）
const handleQuickLoginApi = async (account) => {
  apiLoginForm.username = account.username;
  apiLoginForm.password = account.password;
  await handleApiLogin();
};

const handleQuickLoginFactory = async (account) => {
  factoryLoginForm.username = account.username;
  factoryLoginForm.password = account.password;
  await handleFactoryLogin();
};

// 同步账号：将一个平台的账号密码同步到另一个平台
const handleSyncToOther = (source) => {
  if (source === 'api') {
    // 同步到智能制造
    factoryLoginForm.username = apiLoginForm.username;
    factoryLoginForm.password = apiLoginForm.password;
    ElMessage.success('已同步账号到智能制造系统');
  } else {
    // 同步到云系统
    apiLoginForm.username = factoryLoginForm.username;
    apiLoginForm.password = factoryLoginForm.password;
    ElMessage.success('已同步账号到云系统');
  }
};

// 应用最新账号到指定平台表单
const applyLatestToForm = (platform) => {
  if (!latestAccount.value) return;
  if (platform === 'api') {
    apiLoginForm.username = latestAccount.value.username;
    apiLoginForm.password = latestAccount.value.password;
  } else {
    factoryLoginForm.username = latestAccount.value.username;
    factoryLoginForm.password = latestAccount.value.password;
  }
};
// 注释: applyLatestToForm 保留备用
void applyLatestToForm;

onMounted(() => {
  // 加载跨平台最新登录信息
  loadUnifiedLatest();

  // 加载账号历史列表
  apiAccountsList.value = loadAccountsList(API_ACCOUNTS_LIST_KEY);
  factoryAccountsList.value = loadAccountsList(FACTORY_ACCOUNTS_LIST_KEY);

  // 加载登录记录
  loadApiLoginRecord();
  loadFactoryLoginRecord();

  // 如果没有登录过，自动填充最近一次使用的账号
  if (!apiIsLoggedIn.value && apiRecentAccount.value) {
    apiLoginForm.username = apiRecentAccount.value.username;
    apiLoginForm.password = apiRecentAccount.value.password;
  }
  if (!factoryIsLoggedIn.value && factoryRecentAccount.value) {
    factoryLoginForm.username = factoryRecentAccount.value.username;
    factoryLoginForm.password = factoryRecentAccount.value.password;
  }

  // 检查是否需要自动登录（每 5 小时）
  checkAndAutoLogin();

  // 订阅全局登录状态变化，实时更新本地状态
  unsubscribeStatus = onLoginStatusChange((platform) => {
    handleGlobalStatusChange(platform);
  });
  void loginStatus;

  // 自动检测本地 Ollama 连接
  autoDetectOllama();
});

// 自动登录相关常量
const AUTO_LOGIN_INTERVAL = 5 * 60 * 60 * 1000; // 5 小时
const AUTO_LOGIN_TIME_KEY = 'auto_login_last_time';

// 检查并执行自动登录
const checkAndAutoLogin = () => {
  const lastTime = parseInt(localStorage.getItem(AUTO_LOGIN_TIME_KEY) || '0');
  const now = Date.now();
  // 如果距离上次自动登录超过 5 小时，且有账号密码
  if (now - lastTime >= AUTO_LOGIN_INTERVAL) {
    // 标记为已检查
    localStorage.setItem(AUTO_LOGIN_TIME_KEY, String(now));
    // 延迟执行自动登录（确保页面完全加载）
    setTimeout(() => {
      performAutoLogin();
    }, 2000);
  }
};

// 执行自动登录
const performAutoLogin = async () => {
  // 云系统自动登录
  if (apiLoginForm.username && apiLoginForm.password) {
    try {
      await performApiLoginAuto();
    } catch (e) {
      console.error('云系统自动登录失败:', e);
    }
  }
  // 智能制造系统自动登录
  if (factoryLoginForm.username && factoryLoginForm.password) {
    try {
      await performFactoryLoginAuto();
    } catch (e) {
      console.error('智能制造系统自动登录失败:', e);
    }
  }
};

// 云系统自动登录实现
const performApiLoginAuto = async () => {
  const passwordMd5 = CryptoJS.MD5(apiLoginForm.password).toString();
  const params = new URLSearchParams();
  params.append('tenantId', '000000');
  params.append('username', apiLoginForm.username);
  params.append('password', passwordMd5);
  params.append('grant_type', 'password');
  params.append('scope', 'all');
  params.append('type', 'account');

  const response = await fetch(API_LOGIN_URL + '?' + params.toString(), {
    method: 'POST',
    headers: {
      'Authorization': API_AUTH_HEADER,
      'Accept': 'application/json, text/plain, */*',
      'Content-Type': 'application/x-www-form-urlencoded',
      'Tenant-Id': '000000'
    }
  });
  const result = await response.json();
  if (result.code === 200 && result.data && result.data.access_token) {
    // 保存登录信息
    saveApiLoginRecord({
      username: apiLoginForm.username,
      password: apiLoginForm.password,
      accountInfo: { access_token: result.data.access_token }
    });
    ElMessage.success('云系统已自动登录');
  }
};

// 智能制造系统自动登录实现
const performFactoryLoginAuto = async () => {
  const passwordMd5 = CryptoJS.MD5(factoryLoginForm.password).toString();
  const params = new URLSearchParams();
  params.append('tenantId', '000000');
  params.append('username', factoryLoginForm.username);
  params.append('password', passwordMd5);
  params.append('grant_type', 'password');
  params.append('scope', 'all');
  params.append('type', 'account');

  const response = await fetch(FACTORY_LOGIN_URL + '?' + params.toString(), {
    method: 'POST',
    headers: {
      'Authorization': FACTORY_AUTH_HEADER,
      'Accept': 'application/json, text/plain, */*',
      'Content-Type': 'application/x-www-form-urlencoded',
      'Tenant-Id': '000000'
    }
  });
  const result = await response.json();
  if (result.code === 200 && result.data && result.data.access_token) {
    // 保存登录信息
    saveFactoryLoginRecord({
      username: factoryLoginForm.username,
      password: factoryLoginForm.password,
      accountInfo: { access_token: result.data.access_token }
    });
    ElMessage.success('智能制造系统已自动登录');
  }
};

// 取消订阅
let unsubscribeStatus = null;
onUnmounted(() => {
  if (unsubscribeStatus) {
    unsubscribeStatus();
  }
});

// 监听全局登录状态变化，实时同步到本地
const handleGlobalStatusChange = (platform) => {
  if (platform === 'mp' || platform === 'smart') {
    // 重新从 localStorage 加载登录状态
    if (platform === 'mp') {
      loadApiLoginRecord();
    } else {
      loadFactoryLoginRecord();
    }
  }
};

// 监听 loginStatus 变化，触发响应式更新
watch(() => loginStatus.mp, (newVal) => {
  // 当全局状态变化时，强制重新加载
  if (newVal !== apiIsLoggedIn.value) {
    loadApiLoginRecord();
  }
});

watch(() => loginStatus.smart, (newVal) => {
  if (newVal !== factoryIsLoggedIn.value) {
    loadFactoryLoginRecord();
  }
});

// 监听 window 自定义事件（确保即使组件外其他地方触发状态变更也能同步）
const handleWindowLoginChange = (event) => {
  const { platform } = event.detail || {};
  if (platform === 'mp' || platform === 'smart' || platform === 'iot') {
    handleGlobalStatusChange(platform === 'iot' ? 'smart' : platform);
  }
};
window.addEventListener('login-status-changed', handleWindowLoginChange);

onUnmounted(() => {
  window.removeEventListener('login-status-changed', handleWindowLoginChange);
  if (unsubscribeStatus) {
    unsubscribeStatus();
  }
});
</script>

<template>
  <div class="third-party-auth">
    <GlassCard title="三方账号授权">
      <!-- 最近登录账号快捷登录区域 -->
      <div v-if="latestAccount" class="latest-account-card">
        <div class="latest-header">
          <div class="latest-icon">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="12" r="10"/>
              <polyline points="12 6 12 12 16 14"/>
            </svg>
          </div>
          <div class="latest-title">
            <span class="latest-label">最近登录</span>
            <span class="latest-time">{{ formatLastUsed(latestAccount.lastUsedTime) }}</span>
          </div>
        </div>

        <div class="latest-body">
          <div class="latest-info">
            <div class="latest-username">{{ latestAccount.username }}</div>
            <div class="latest-password">{{ latestAccount.password }}</div>
          </div>

          <div class="latest-actions">
            <button
              v-if="!apiIsLoggedIn || apiLoginForm.username !== latestAccount.username"
              class="quick-btn api"
              :disabled="apiIsLogging"
              @click="handleQuickLoginApi(latestAccount)"
            >
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z" />
              </svg>
              <span>{{ apiIsLogging ? '登录中...' : '登录云系统' }}</span>
            </button>

            <button
              v-if="!factoryIsLoggedIn || factoryLoginForm.username !== latestAccount.username"
              class="quick-btn factory"
              :disabled="factoryIsLogging"
              @click="handleQuickLoginFactory(latestAccount)"
            >
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <rect x="3" y="3" width="7" height="7" />
                <rect x="14" y="3" width="7" height="7" />
                <rect x="14" y="14" width="7" height="7" />
                <rect x="3" y="14" width="7" height="7" />
              </svg>
              <span>{{ factoryIsLogging ? '登录中...' : '登录智能制造' }}</span>
            </button>
          </div>
        </div>
      </div>

      <div class="auth-container">
        <!-- 本地 Ollama 模型配置（推荐，数据不出内网） -->
        <div class="auth-card ollama-card recommended">
          <div class="auth-header">
            <div class="auth-icon ollama">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <rect x="2" y="3" width="20" height="14" rx="2"/>
                <line x1="8" y1="21" x2="16" y2="21"/>
                <line x1="12" y1="17" x2="12" y2="21"/>
              </svg>
            </div>
            <h3>本地 Ollama 模型</h3>
            <span class="auth-domain">localhost:11434</span>
            <span class="status-badge recommended-badge">推荐</span>
            <span v-if="ollamaConnected" class="status-badge logged">已连接</span>
            <span v-else class="status-badge unlogged">未连接</span>
          </div>

          <div class="security-tip">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#52c41a" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
            </svg>
            <span>本地部署 · 数据不出内网 · 零泄密风险</span>
          </div>

          <div class="auth-form">
            <div class="form-group">
              <label>Ollama 服务地址</label>
              <div class="input-with-dropdown">
                <input
                  v-model="ollamaUrl"
                  type="text"
                  class="form-input key-input-monospace"
                  placeholder="http://localhost:11434"
                  @blur="handleSaveOllamaUrl"
                />
              </div>
              <p class="key-hint">
                Docker 中 Ollama 容器地址，默认 http://localhost:11434
              </p>
            </div>

            <div v-if="ollamaModelList.length" class="form-group">
              <label>可用模型</label>
              <div class="model-list">
                <span v-for="m in ollamaModelList" :key="m" class="model-tag" :class="{ active: ollamaSelectedModel === m }" @click="ollamaSelectedModel = m; handleSaveOllamaUrl()">
                  {{ m }}
                </span>
              </div>
            </div>

            <div v-if="ollamaTestResult" class="form-message" :class="ollamaTestResult.success ? 'success' : 'error'">
              {{ ollamaTestResult.message }}
            </div>

            <div class="form-actions">
              <button
                class="sync-btn"
                :disabled="ollamaIsTesting"
                @click="handleTestOllama"
              >
                <svg v-if="!ollamaIsTesting" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <polyline points="20 6 9 17 4 12"/>
                </svg>
                <svg v-else width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="spin">
                  <path d="M21 12a9 9 0 1 1-6.219-8.56"/>
                </svg>
                <span>{{ ollamaIsTesting ? '检测中...' : '检测连接' }}</span>
              </button>
            </div>
          </div>
        </div>

        <!-- DeepSeek API Key 配置（云端备选） -->
        <div class="auth-card deepseek-card">
          <div class="auth-header">
            <div class="auth-icon deepseek">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <circle cx="12" cy="12" r="10"/>
                <path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3"/>
                <line x1="12" y1="17" x2="12.01" y2="17"/>
              </svg>
            </div>
            <h3>DeepSeek AI（云端）</h3>
            <span class="auth-domain">api.deepseek.com</span>
            <span v-if="deepseekApiKey" class="status-badge logged">已配置</span>
            <span v-else class="status-badge unlogged">未配置</span>
          </div>

          <div class="auth-form">
            <div class="form-group">
              <label>API Key</label>
              <div class="input-with-dropdown">
                <input
                  v-model="deepseekApiKey"
                  :type="deepseekShowKey ? 'text' : 'password'"
                  class="form-input key-input-monospace"
                  placeholder="请输入 DeepSeek API Key（sk-...）"
                />
                <button
                  type="button"
                  class="dropdown-toggle"
                  @click="deepseekShowKey = !deepseekShowKey"
                  :title="deepseekShowKey ? '隐藏' : '显示'"
                >
                  <svg v-if="!deepseekShowKey" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/>
                    <circle cx="12" cy="12" r="3"/>
                  </svg>
                  <svg v-else width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94"/>
                    <path d="M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19"/>
                    <line x1="1" y1="1" x2="23" y2="23"/>
                  </svg>
                </button>
              </div>
              <p class="key-hint">
                获取 API Key：
                <a href="https://platform.deepseek.com/api_keys" target="_blank" rel="noopener">platform.deepseek.com</a>
              </p>
            </div>

            <div v-if="deepseekTestResult" class="form-message" :class="deepseekTestResult.success ? 'success' : 'error'">
              {{ deepseekTestResult.message }}
            </div>

            <div v-if="deepseekKeySaved" class="form-message success">
              API Key 已保存成功
            </div>

            <div class="form-actions">
              <button
                v-if="!deepseekApiKey"
                class="auth-btn primary"
                :disabled="!deepseekApiKey.trim()"
                @click="handleSaveDeepseekKey"
              >
                保存 Key
              </button>
              <template v-else>
                <button
                  class="sync-btn"
                  :disabled="deepseekIsTesting"
                  @click="handleTestDeepseekKey"
                >
                  <svg v-if="!deepseekIsTesting" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <polyline points="20 6 9 17 4 12"/>
                  </svg>
                  <svg v-else width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="spin">
                    <path d="M21 12a9 9 0 1 1-6.219-8.56"/>
                  </svg>
                  <span>{{ deepseekIsTesting ? '测试中...' : '测试连接' }}</span>
                </button>
                <button class="auth-btn primary" @click="handleSaveDeepseekKey">
                  更新 Key
                </button>
                <button class="auth-btn danger-outline" @click="handleClearDeepseekKey">
                  清除
                </button>
              </template>
            </div>
          </div>
        </div>

        <div class="auth-card">
          <div class="auth-header">
            <div class="auth-icon api">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z" />
              </svg>
            </div>
            <h3>云系统</h3>
            <span class="auth-domain">mp.holyview.cn</span>
            <span v-if="apiIsLoggedIn" class="status-badge logged">已登录</span>
            <span v-else class="status-badge unlogged">未登录</span>
          </div>

          <div v-if="!apiIsLoggedIn" class="auth-form">
            <div class="form-group">
              <label>用户名</label>
              <div class="input-with-dropdown">
                <input
                  v-model="apiLoginForm.username"
                  type="text"
                  class="form-input"
                  placeholder="请输入用户名"
                  @focus="apiShowAccountDropdown = true"
                  @blur="setTimeout(() => apiShowAccountDropdown = false, 200)"
                />
                <button
                  v-if="apiAccountsList.length > 0"
                  type="button"
                  class="dropdown-toggle"
                  @mousedown.prevent="apiShowAccountDropdown = !apiShowAccountDropdown"
                >
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" :class="{ rotated: apiShowAccountDropdown }">
                    <polyline points="6 9 12 15 18 9" />
                  </svg>
                </button>
                <div v-if="apiShowAccountDropdown && apiAccountsList.length > 0" class="account-dropdown">
                  <div class="dropdown-header">
                    <span>历史账号</span>
                    <span class="dropdown-count">共 {{ apiAccountsList.length }} 个</span>
                  </div>
                  <div class="dropdown-list">
                    <div
                      v-for="account in apiAccountsList"
                      :key="account.username"
                      class="dropdown-item"
                      :class="{ active: account.username === apiLoginForm.username }"
                      @mousedown.prevent="handleSelectApiAccount(account)"
                    >
                      <div class="account-info">
                        <div class="account-username">{{ account.username }}</div>
                        <div class="account-time">{{ formatLastUsed(account.lastUsedTime) }}</div>
                      </div>
                      <div class="dropdown-actions">
                        <button
                          type="button"
                          class="quick-fill-btn"
                          @mousedown.stop.prevent="handleQuickLoginApi(account)"
                          title="直接登录"
                        >
                          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                            <path d="M5 12h14M12 5l7 7-7 7" />
                          </svg>
                        </button>
                        <button
                          type="button"
                          class="delete-btn"
                          @mousedown.stop.prevent="handleDeleteApiAccount(account.username, $event)"
                          title="删除该账号"
                        >
                          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                            <line x1="18" y1="6" x2="6" y2="18" />
                            <line x1="6" y1="6" x2="18" y2="18" />
                          </svg>
                        </button>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
            <div class="form-group">
              <label>密码</label>
              <input
                v-model="apiLoginForm.password"
                type="password"
                class="form-input"
                placeholder="请输入密码"
              />
            </div>
            <div v-if="apiLoginMessage" class="form-message" :class="apiLoginMessage.includes('成功') ? 'success' : 'error'">
              {{ apiLoginMessage }}
            </div>
            <div class="form-actions">
              <button
                v-if="factoryLoginForm.username && factoryLoginForm.password && !apiIsLogging"
                type="button"
                class="sync-btn"
                @click="handleSyncToOther('factory')"
                title="使用智能制造的账号登录"
              >
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <polyline points="17 1 21 5 17 9" />
                  <path d="M3 11V9a4 4 0 0 1 4-4h14" />
                  <polyline points="7 23 3 19 7 15" />
                  <path d="M21 13v2a4 4 0 0 1-4 4H3" />
                </svg>
                <span>使用智能制造账号</span>
              </button>
              <button
                class="auth-btn primary"
                :disabled="apiIsLogging"
                @click="handleApiLogin"
              >
                <svg v-if="apiIsLogging" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="spin">
                  <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
                </svg>
                <span>{{ apiIsLogging ? '登录中...' : '登录' }}</span>
              </button>
            </div>
          </div>

          <div v-else class="auth-success">
            <div class="success-icon">
              <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="#10B981" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <polyline points="20 6 9 17 4 12" />
              </svg>
            </div>
            <p>登录成功</p>
            <p class="username">{{ apiLoginForm.username }}</p>
            <button class="auth-btn danger" @click="handleApiLogout">
              退出登录
            </button>
          </div>
        </div>

        <div class="auth-card">
          <div class="auth-header">
            <div class="auth-icon factory">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <rect x="3" y="3" width="7" height="7" />
                <rect x="14" y="3" width="7" height="7" />
                <rect x="14" y="14" width="7" height="7" />
                <rect x="3" y="14" width="7" height="7" />
              </svg>
            </div>
            <h3>智能制造系统</h3>
            <span class="auth-domain">218.90.146.230:20001</span>
            <span v-if="factoryIsLoggedIn" class="status-badge logged">已登录</span>
            <span v-else class="status-badge unlogged">未登录</span>
          </div>

          <div v-if="!factoryIsLoggedIn" class="auth-form">
            <div class="form-group">
              <label>用户名</label>
              <div class="input-with-dropdown">
                <input
                  v-model="factoryLoginForm.username"
                  type="text"
                  class="form-input"
                  placeholder="请输入用户名"
                  @focus="factoryShowAccountDropdown = true"
                  @blur="setTimeout(() => factoryShowAccountDropdown = false, 200)"
                />
                <button
                  v-if="factoryAccountsList.length > 0"
                  type="button"
                  class="dropdown-toggle"
                  @mousedown.prevent="factoryShowAccountDropdown = !factoryShowAccountDropdown"
                >
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" :class="{ rotated: factoryShowAccountDropdown }">
                    <polyline points="6 9 12 15 18 9" />
                  </svg>
                </button>
                <div v-if="factoryShowAccountDropdown && factoryAccountsList.length > 0" class="account-dropdown">
                  <div class="dropdown-header">
                    <span>历史账号</span>
                    <span class="dropdown-count">共 {{ factoryAccountsList.length }} 个</span>
                  </div>
                  <div class="dropdown-list">
                    <div
                      v-for="account in factoryAccountsList"
                      :key="account.username"
                      class="dropdown-item"
                      :class="{ active: account.username === factoryLoginForm.username }"
                      @mousedown.prevent="handleSelectFactoryAccount(account)"
                    >
                      <div class="account-info">
                        <div class="account-username">{{ account.username }}</div>
                        <div class="account-time">{{ formatLastUsed(account.lastUsedTime) }}</div>
                      </div>
                      <div class="dropdown-actions">
                        <button
                          type="button"
                          class="quick-fill-btn"
                          @mousedown.stop.prevent="handleQuickLoginFactory(account)"
                          title="直接登录"
                        >
                          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                            <path d="M5 12h14M12 5l7 7-7 7" />
                          </svg>
                        </button>
                        <button
                          type="button"
                          class="delete-btn"
                          @mousedown.stop.prevent="handleDeleteFactoryAccount(account.username, $event)"
                          title="删除该账号"
                        >
                          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                            <line x1="18" y1="6" x2="6" y2="18" />
                            <line x1="6" y1="6" x2="18" y2="18" />
                          </svg>
                        </button>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
            <div class="form-group">
              <label>密码</label>
              <input
                v-model="factoryLoginForm.password"
                type="password"
                class="form-input"
                placeholder="请输入密码"
              />
            </div>
            <div v-if="factoryLoginMessage" class="form-message" :class="factoryLoginMessage.includes('成功') ? 'success' : 'error'">
              {{ factoryLoginMessage }}
            </div>
            <div class="form-actions">
              <button
                v-if="apiLoginForm.username && apiLoginForm.password && !factoryIsLogging"
                type="button"
                class="sync-btn"
                @click="handleSyncToOther('api')"
                title="使用云系统的账号登录"
              >
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <polyline points="17 1 21 5 17 9" />
                  <path d="M3 11V9a4 4 0 0 1 4-4h14" />
                  <polyline points="7 23 3 19 7 15" />
                  <path d="M21 13v2a4 4 0 0 1-4 4H3" />
                </svg>
                <span>使用云系统账号</span>
              </button>
              <button
                class="auth-btn primary"
                :disabled="factoryIsLogging"
                @click="handleFactoryLogin"
              >
                <svg v-if="factoryIsLogging" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="spin">
                  <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
                </svg>
                <span>{{ factoryIsLogging ? '登录中...' : '登录' }}</span>
              </button>
            </div>
          </div>

          <div v-else class="auth-success">
            <div class="success-icon">
              <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="#10B981" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <polyline points="20 6 9 17 4 12" />
              </svg>
            </div>
            <p>登录成功</p>
            <p class="username">{{ factoryLoginForm.username }}</p>
            <button class="auth-btn danger" @click="handleFactoryLogout">
              退出登录
            </button>
          </div>
        </div>
      </div>

      <div class="tips-panel">
        <div class="tip-item">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#165DFF" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="10" />
            <line x1="12" y1="16" x2="12" y2="12" />
            <line x1="12" y1="8" x2="12.01" y2="8" />
          </svg>
          <span>登录状态本地保存，期间无需重复登录</span>
        </div>
        <div class="tip-item">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#165DFF" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="10" />
            <line x1="12" y1="16" x2="12" y2="12" />
            <line x1="12" y1="8" x2="12.01" y2="8" />
          </svg>
          <span>顶部"最近登录"账号可一键登录任意平台</span>
        </div>
        <div class="tip-item">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#165DFF" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="10" />
            <line x1="12" y1="16" x2="12" y2="12" />
            <line x1="12" y1="8" x2="12.01" y2="8" />
          </svg>
          <span>支持"使用XX账号"按钮将一个平台的账号同步到另一个平台</span>
        </div>
        <div class="tip-item">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#165DFF" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="10" />
            <line x1="12" y1="16" x2="12" y2="12" />
            <line x1="12" y1="8" x2="12.01" y2="8" />
          </svg>
          <span>点击退出登录可清除本地缓存</span>
        </div>
      </div>

      <!-- 接口及接口说明 -->
      <div class="api-doc-section">
        <div class="api-doc-header">
          <h3 class="api-doc-title">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <polyline points="16 18 22 12 16 6"/>
              <polyline points="8 6 2 12 8 18"/>
            </svg>
            接口及接口说明
          </h3>
          <span class="api-doc-subtitle">选择平台查看对应接口列表与说明</span>
        </div>

        <!-- 平台切换 Tab -->
        <div class="api-doc-tabs">
          <button
            v-for="p in platformApis"
            :key="p.id"
            class="api-doc-tab"
            :class="{ active: activeApiDocPlatform === p.id }"
            @click="activeApiDocPlatform = p.id"
          >
            <span class="api-doc-tab-name">{{ p.name }}</span>
            <span class="api-doc-tab-count">{{ p.apis.length }} 个接口</span>
          </button>
        </div>

        <!-- 平台基础信息 -->
        <div class="api-doc-platform">
          <div class="api-doc-platform-info">
            <span class="api-doc-label">服务地址：</span>
            <code class="api-doc-domain">{{ activeApiDoc.domain }}</code>
          </div>
          <div class="api-doc-platform-info">
            <span class="api-doc-label">Token 存储：</span>
            <code class="api-doc-token-key">{{ activeApiDoc.loginTokenKey }}</code>
          </div>
          <p class="api-doc-desc">{{ activeApiDoc.description }}</p>

          <div class="api-doc-auth">
            <div class="api-doc-auth-title">认证方式（登录前请求头）</div>
            <div class="api-doc-header-row" v-for="h in activeApiDoc.baseAuth" :key="h.key">
              <code class="api-doc-header-key">{{ h.key }}</code>
              <code class="api-doc-header-value">{{ h.value }}</code>
              <span class="api-doc-header-desc">{{ h.desc }}</span>
            </div>
          </div>
        </div>

        <!-- 接口列表 -->
        <div class="api-doc-list">
          <div class="api-doc-item" v-for="api in activeApiDoc.apis" :key="api.method + api.path">
            <div class="api-doc-item-head">
              <span class="api-doc-method" :class="api.method.toLowerCase()">{{ api.method }}</span>
              <code class="api-doc-path">{{ api.path }}</code>
              <span class="api-doc-name">{{ api.name }}</span>
            </div>
            <p class="api-doc-item-desc">{{ api.desc }}</p>

            <!-- 请求头 -->
            <template v-if="api.headers && api.headers.length">
              <div class="api-doc-block">
                <div class="api-doc-block-title">请求头</div>
                <div class="api-doc-header-row" v-for="h in api.headers" :key="h.key">
                  <code class="api-doc-header-key">{{ h.key }}</code>
                  <code class="api-doc-header-value">{{ h.value }}</code>
                  <span class="api-doc-header-desc">{{ h.desc }}</span>
                </div>
              </div>
            </template>

            <!-- 参数 -->
            <template v-if="api.params && api.params.length">
              <div class="api-doc-block">
                <div class="api-doc-block-title">参数说明</div>
                <div class="api-doc-params-table">
                  <div class="api-doc-params-head">
                    <span>参数名</span>
                    <span>必填</span>
                    <span>说明</span>
                  </div>
                  <div class="api-doc-params-row" v-for="p in api.params" :key="p.key">
                    <code class="api-doc-param-key">{{ p.key }}</code>
                    <span class="api-doc-param-required" :class="{ yes: p.required }">{{ p.required ? '是' : '否' }}</span>
                    <span class="api-doc-param-desc">{{ p.desc }}</span>
                  </div>
                </div>
              </div>
            </template>

            <!-- 返回说明 -->
            <div class="api-doc-block" v-if="api.response">
              <div class="api-doc-block-title">返回说明</div>
              <p class="api-doc-response">{{ api.response }}</p>
            </div>
          </div>
        </div>
      </div>
    </GlassCard>
  </div>
</template>

<style scoped>
.third-party-auth {
  max-width: 100%;
}

/* ============== 最近登录账号卡片 ============== */
.latest-account-card {
  background: linear-gradient(135deg, rgba(22, 93, 255, 0.05) 0%, rgba(15, 76, 208, 0.08) 100%);
  border: 1px solid rgba(22, 93, 255, 0.2);
  border-radius: 12px;
  padding: 16px 20px;
  margin-bottom: 24px;
  transition: all 0.2s;
}

.latest-account-card:hover {
  box-shadow: 0 4px 16px rgba(22, 93, 255, 0.1);
  border-color: rgba(22, 93, 255, 0.3);
}

.latest-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
}

.latest-icon {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(22, 93, 255, 0.1);
  color: #165DFF;
  border-radius: 8px;
}

.latest-title {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
}

.latest-label {
  font-size: 14px;
  font-weight: 600;
  color: #1E293B;
}

.latest-time {
  font-size: 12px;
  color: #94A3B8;
}

.latest-body {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
}

.latest-info {
  flex: 1;
  min-width: 200px;
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 10px 14px;
  background: #FFFFFF;
  border-radius: 8px;
  border: 1px solid rgba(22, 93, 255, 0.15);
}

.latest-username {
  font-size: 14px;
  font-weight: 600;
  color: #1E293B;
}

.latest-password {
  font-size: 13px;
  color: #64748B;
  font-family: 'SF Mono', 'Monaco', 'Cascadia Code', 'Courier New', monospace;
  background: #F1F5F9;
  padding: 2px 8px;
  border-radius: 4px;
}

.latest-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.quick-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  border: none;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  color: #FFFFFF;
}

.quick-btn.api {
  background: linear-gradient(135deg, #3B82F6 0%, #1D4ED8 100%);
}

.quick-btn.factory {
  background: linear-gradient(135deg, #EC4899 0%, #BE185D 100%);
}

.quick-btn:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.quick-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* ============== 两个平台卡片 ============== */
.auth-container {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 24px;
  margin-bottom: 24px;
}

.auth-card {
  background: #FAFAFA;
  border-radius: 12px;
  padding: 24px;
  border: 1px solid #E2E8F0;
}

.auth-header {
  text-align: center;
  margin-bottom: 24px;
  position: relative;
}

.auth-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 12px;

  &.api {
    background: rgba(59, 130, 246, 0.1);
    color: #3B82F6;
  }

  &.factory {
    background: rgba(236, 72, 153, 0.1);
    color: #EC4899;
  }
}

.auth-header h3 {
  font-size: 18px;
  font-weight: 600;
  color: #1E293B;
  margin: 0 0 4px;
}

.auth-domain {
  font-size: 12px;
  color: #94A3B8;
}

.status-badge {
  position: absolute;
  top: 0;
  right: 0;
  font-size: 11px;
  padding: 3px 10px;
  border-radius: 10px;
  font-weight: 500;
}

.status-badge.logged {
  background: rgba(16, 185, 129, 0.1);
  color: #10B981;
}

.status-badge.unlogged {
  background: rgba(148, 163, 184, 0.1);
  color: #94A3B8;
}

.auth-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-group label {
  font-size: 13px;
  font-weight: 500;
  color: #475569;
}

.form-input {
  width: 100%;
  padding: 10px 14px;
  border: 1px solid #E2E8F0;
  border-radius: 8px;
  font-size: 14px;
  color: #1E293B;
  background: #FFFFFF;

  &:focus {
    outline: none;
    border-color: #165DFF;
    box-shadow: 0 0 0 3px rgba(22, 93, 255, 0.1);
  }

  &::placeholder {
    color: #94A3B8;
  }
}

.input-with-dropdown {
  position: relative;
  width: 100%;
}

.input-with-dropdown .form-input {
  padding-right: 36px;
}

.dropdown-toggle {
  position: absolute;
  right: 8px;
  top: 50%;
  transform: translateY(-50%);
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: none;
  border-radius: 4px;
  color: #64748B;
  cursor: pointer;
  transition: all 0.2s;

  &:hover {
    background: rgba(22, 93, 255, 0.1);
    color: #165DFF;
  }

  svg {
    transition: transform 0.2s;
  }

  svg.rotated {
    transform: rotate(180deg);
  }
}

.account-dropdown {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  right: 0;
  background: #FFFFFF;
  border: 1px solid #E2E8F0;
  border-radius: 8px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
  z-index: 1000;
  max-height: 280px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.dropdown-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 14px;
  border-bottom: 1px solid #F1F5F9;
  font-size: 12px;
  color: #64748B;
  font-weight: 500;
  background: #F8FAFC;
  flex-shrink: 0;
}

.dropdown-count {
  font-size: 11px;
  color: #94A3B8;
}

.dropdown-list {
  overflow-y: auto;
  flex: 1;
}

.dropdown-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 14px;
  cursor: pointer;
  transition: background 0.15s;
  border-bottom: 1px solid #F1F5F9;

  &:last-child {
    border-bottom: none;
  }

  &:hover {
    background: #F8FAFC;
  }

  &.active {
    background: rgba(22, 93, 255, 0.08);
  }
}

.account-info {
  flex: 1;
  min-width: 0;
}

.account-username {
  font-size: 13px;
  font-weight: 500;
  color: #1E293B;
  margin-bottom: 2px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.account-time {
  font-size: 11px;
  color: #94A3B8;
}

.dropdown-actions {
  display: flex;
  gap: 4px;
  flex-shrink: 0;
  margin-left: 8px;
}

.quick-fill-btn {
  width: 22px;
  height: 22px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: none;
  border-radius: 4px;
  color: #94A3B8;
  cursor: pointer;
  transition: all 0.2s;

  &:hover {
    background: rgba(22, 93, 255, 0.1);
    color: #165DFF;
  }
}

.delete-btn {
  width: 22px;
  height: 22px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: none;
  border-radius: 4px;
  color: #94A3B8;
  cursor: pointer;
  flex-shrink: 0;
  transition: all 0.2s;

  &:hover {
    background: rgba(245, 63, 63, 0.1);
    color: #F53F3F;
  }
}

.form-message {
  font-size: 12px;
  padding: 8px 12px;
  border-radius: 6px;
  text-align: center;

  &.success {
    background: rgba(16, 185, 129, 0.1);
    color: #10B981;
  }

  &.error {
    background: rgba(239, 68, 68, 0.1);
    color: #EF4444;
  }
}

.form-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

.sync-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 8px 12px;
  background: rgba(22, 93, 255, 0.05);
  border: 1px solid rgba(22, 93, 255, 0.2);
  border-radius: 6px;
  font-size: 12px;
  color: #165DFF;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;

  &:hover {
    background: rgba(22, 93, 255, 0.1);
    border-color: #165DFF;
  }
}

.auth-btn {
  flex: 1;
  padding: 12px 24px;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;

  &.primary {
    background: linear-gradient(135deg, #165DFF 0%, #0F4CD0 100%);
    color: #FFFFFF;

    &:hover:not(:disabled) {
      transform: translateY(-1px);
      box-shadow: 0 4px 12px rgba(22, 93, 255, 0.3);
    }
  }

  &.danger {
    background: rgba(239, 68, 68, 0.1);
    color: #EF4444;

    &:hover {
      background: rgba(239, 68, 68, 0.2);
    }
  }

  &:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }
}

.auth-success {
  text-align: center;
  padding: 20px 0;
}

.success-icon {
  margin-bottom: 16px;
}

.auth-success p {
  margin: 0 0 8px;
  font-size: 14px;
  color: #1E293B;
}

.auth-success .username {
  font-weight: 600;
  color: #165DFF;
  font-size: 16px;
}

.expire-tip {
  font-size: 12px;
  color: #94A3B8 !important;
}

.tips-panel {
  background: rgba(22, 93, 255, 0.05);
  border-radius: 8px;
  padding: 16px 20px;
}

.tip-item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: #64748B;

  & + & {
    margin-top: 8px;
  }
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

/* ============== DeepSeek 配置卡片 ============== */
.auth-icon.deepseek {
  background: rgba(99, 102, 241, 0.1);
  color: #6366F1;
}

.key-input-monospace {
  font-family: 'SF Mono', 'Monaco', 'Cascadia Code', 'Courier New', monospace;
  font-size: 13px !important;
  letter-spacing: 0.5px;
}

.key-hint {
  font-size: 11px;
  color: #94A3B8;
  margin-top: 4px;
}

.key-hint a {
  color: #165DFF;
  text-decoration: none;
}

.key-hint a:hover {
  text-decoration: underline;
}

.auth-btn.danger-outline {
  background: #fff;
  color: #EF4444;
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.auth-btn.danger-outline:hover {
  background: rgba(239, 68, 68, 0.05);
  border-color: #EF4444;
}

/* ============== Ollama 本地模型卡片 ============== */
.auth-icon.ollama {
  background: rgba(34, 197, 94, 0.1);
  color: #22C55E;
}

.ollama-card.recommended {
  border: 2px solid rgba(34, 197, 94, 0.25);
  position: relative;
}

.ollama-card.recommended::before {
  content: '';
  position: absolute;
  top: 0; left: 0; right: 0;
  height: 3px;
  background: linear-gradient(90deg, #22C55E, #16A34A);
  border-radius: 12px 12px 0 0;
}

.security-tip {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  margin: 0 16px 12px;
  background: rgba(34, 197, 94, 0.06);
  border-radius: 8px;
  font-size: 12px;
  color: #16A34A;
  border: 1px solid rgba(34, 197, 94, 0.15);
}

.recommended-badge {
  background: rgba(34, 197, 94, 0.1) !important;
  color: #16A34A !important;
}

.model-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.model-tag {
  padding: 4px 12px;
  border-radius: 16px;
  font-size: 12px;
  cursor: pointer;
  border: 1px solid #E2E8F0;
  background: #F8FAFC;
  color: #64748B;
  transition: all 0.2s;
}

.model-tag:hover {
  border-color: #22C55E;
  color: #22C55E;
}

.model-tag.active {
  background: rgba(34, 197, 94, 0.1);
  border-color: #22C55E;
  color: #16A34A;
  font-weight: 500;
}

/* 响应式 */
@media (max-width: 768px) {
  .latest-body {
    flex-direction: column;
    align-items: stretch;
  }

  .latest-info {
    flex-direction: column;
    align-items: flex-start;
    gap: 6px;
  }

  .latest-actions {
    justify-content: stretch;
  }

  .quick-btn {
    flex: 1;
    justify-content: center;
  }

  .form-actions {
    flex-direction: column;
  }

  .sync-btn {
    width: 100%;
    justify-content: center;
  }

  .api-doc-tabs {
    flex-wrap: wrap;
  }

  .api-doc-header-row {
    flex-wrap: wrap;
  }
}

/* ============== 接口及接口说明 ============== */
.api-doc-section {
  margin-top: 24px;
  border: 1px solid rgba(22, 93, 255, 0.12);
  border-radius: 12px;
  background: linear-gradient(135deg, rgba(22, 93, 255, 0.03), rgba(255, 255, 255, 0.6));
  padding: 20px;
}

.api-doc-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 16px;
}

.api-doc-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
  color: #1F2937;
  margin: 0;
}

.api-doc-subtitle {
  font-size: 12px;
  color: #94A3B8;
}

.api-doc-tabs {
  display: flex;
  gap: 10px;
  margin-bottom: 16px;
}

.api-doc-tab {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  border: 1px solid rgba(22, 93, 255, 0.2);
  border-radius: 8px;
  background: #fff;
  cursor: pointer;
  transition: all 0.2s;
}

.api-doc-tab:hover {
  border-color: #165DFF;
}

.api-doc-tab.active {
  background: #165DFF;
  border-color: #165DFF;
  color: #fff;
}

.api-doc-tab-name {
  font-size: 13px;
  font-weight: 600;
}

.api-doc-tab-count {
  font-size: 11px;
  opacity: 0.8;
  background: rgba(148, 163, 184, 0.15);
  padding: 2px 8px;
  border-radius: 10px;
}

.api-doc-tab.active .api-doc-tab-count {
  background: rgba(255, 255, 255, 0.2);
}

.api-doc-platform {
  background: #fff;
  border: 1px solid rgba(148, 163, 184, 0.2);
  border-radius: 8px;
  padding: 14px 16px;
  margin-bottom: 16px;
}

.api-doc-platform-info {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
  font-size: 13px;
}

.api-doc-label {
  color: #6B7280;
  white-space: nowrap;
}

.api-doc-domain,
.api-doc-token-key {
  font-family: 'SF Mono', 'Monaco', 'Cascadia Code', monospace;
  font-size: 12px;
  color: #165DFF;
  background: rgba(22, 93, 255, 0.06);
  padding: 2px 8px;
  border-radius: 4px;
}

.api-doc-desc {
  font-size: 13px;
  color: #6B7280;
  margin: 8px 0 12px;
  line-height: 1.6;
}

.api-doc-auth {
  border-top: 1px dashed rgba(148, 163, 184, 0.3);
  padding-top: 12px;
}

.api-doc-auth-title,
.api-doc-block-title {
  font-size: 12px;
  font-weight: 600;
  color: #374151;
  margin-bottom: 8px;
}

.api-doc-header-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 6px;
  font-size: 12px;
}

.api-doc-header-key {
  min-width: 130px;
  font-family: 'SF Mono', 'Monaco', 'Cascadia Code', monospace;
  color: #7C3AED;
  background: rgba(124, 58, 237, 0.06);
  padding: 2px 8px;
  border-radius: 4px;
}

.api-doc-header-value {
  font-family: 'SF Mono', 'Monaco', 'Cascadia Code', monospace;
  color: #059669;
  background: rgba(5, 150, 105, 0.06);
  padding: 2px 8px;
  border-radius: 4px;
  word-break: break-all;
}

.api-doc-header-desc {
  color: #94A3B8;
}

.api-doc-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.api-doc-item {
  background: #fff;
  border: 1px solid rgba(148, 163, 184, 0.2);
  border-radius: 8px;
  padding: 16px;
}

.api-doc-item-head {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.api-doc-method {
  font-size: 11px;
  font-weight: 700;
  color: #fff;
  padding: 3px 10px;
  border-radius: 4px;
  letter-spacing: 0.5px;
}

.api-doc-method.get {
  background: #059669;
}

.api-doc-method.post {
  background: #165DFF;
}

.api-doc-path {
  font-family: 'SF Mono', 'Monaco', 'Cascadia Code', monospace;
  font-size: 13px;
  color: #1F2937;
  word-break: break-all;
}

.api-doc-name {
  font-size: 13px;
  font-weight: 600;
  color: #165DFF;
}

.api-doc-item-desc {
  font-size: 13px;
  color: #6B7280;
  margin: 8px 0 12px;
  line-height: 1.6;
}

.api-doc-block {
  border-top: 1px dashed rgba(148, 163, 184, 0.3);
  padding-top: 10px;
  margin-top: 10px;
}

.api-doc-params-table {
  border: 1px solid rgba(148, 163, 184, 0.2);
  border-radius: 6px;
  overflow: hidden;
}

.api-doc-params-head,
.api-doc-params-row {
  display: grid;
  grid-template-columns: 140px 50px 1fr;
  gap: 10px;
  padding: 6px 12px;
  align-items: center;
}

.api-doc-params-head {
  background: rgba(148, 163, 184, 0.08);
  font-size: 11px;
  font-weight: 600;
  color: #6B7280;
}

.api-doc-params-row {
  font-size: 12px;
  border-top: 1px solid rgba(148, 163, 184, 0.12);
}

.api-doc-param-key {
  font-family: 'SF Mono', 'Monaco', 'Cascadia Code', monospace;
  color: #1F2937;
  word-break: break-all;
}

.api-doc-param-required {
  text-align: center;
  font-size: 11px;
  padding: 2px 6px;
  border-radius: 4px;
  width: fit-content;
  background: rgba(148, 163, 184, 0.15);
  color: #94A3B8;
}

.api-doc-param-required.yes {
  background: rgba(239, 68, 68, 0.1);
  color: #EF4444;
}

.api-doc-param-desc {
  color: #6B7280;
}

.api-doc-response {
  font-size: 12px;
  color: #6B7280;
  margin: 0;
  line-height: 1.7;
  word-break: break-all;
}
</style>
