import { ref, reactive } from 'vue';
import { ElMessage } from 'element-plus';
import CryptoJS from 'crypto-js';
import { saveCredentials, updateLoginStatus, getCredentials } from './loginStatus.js';

// 平台配置
const PLATFORMS = {
  mp: {
    name: '云系统',
    authHeader: 'Basic ' + btoa('saber:saber_secret'),
    tokenKey: 'mp_token',
    loginUrl: '/api/blade-auth/oauth/token',
    domain: 'mp.holyview.cn'
  },
  smart: {
    name: '智能制造系统',
    authHeader: 'Basic ' + btoa('saber_identity_client:saber_identity_secret'),
    tokenKey: 'smart_factory_token',
    loginUrl: '/smart-api/api/blade-auth/oauth/token',
    domain: '218.90.146.230:20001'
  },
  iot: {
    name: '工厂数据系统',
    authHeader: 'Basic ' + btoa('saber_identity_client:saber_identity_secret'),
    tokenKey: 'iot_token',
    loginUrl: '/iot-api/api/blade-auth/oauth/token',
    domain: 'iot'
  }
};

export const getPlatformConfig = (platform) => PLATFORMS[platform];

// 获取最近一次登录的账号密码（从所有平台历史中找）
export const getLatestCredentials = (platform = null) => {
  try {
    // 1. 优先使用跨平台最新登录
    const unifiedLatest = localStorage.getItem('unified_latest_login');
    if (unifiedLatest) {
      const data = JSON.parse(unifiedLatest);
      if (data && data.username && data.password) {
        // 如果指定了平台，且最新账号不匹配该平台，则继续查找
        if (platform && data.platform && data.platform !== platform) {
          // 继续往下找
        } else {
          return {
            username: data.username,
            password: data.password,
            tenantId: '000000',
            platform: data.platform
          };
        }
      }
    }

    // 2. 从指定平台的账号历史中找最新
    const listKeys = {
      mp: 'api_accounts_list',
      smart: 'factory_accounts_list',
      iot: 'factory_accounts_list'  // iot 复用 smart 的账号列表
    };

    // 收集所有账号
    const allAccounts = [];
    Object.keys(listKeys).forEach(p => {
      const list = localStorage.getItem(listKeys[p]);
      if (list) {
        try {
          const items = JSON.parse(list);
          items.forEach(item => {
            allAccounts.push({ ...item, platform: p });
          });
        } catch (e) {
          // ignore
        }
      }
    });

    if (allAccounts.length === 0) return null;

    // 按时间倒序
    allAccounts.sort((a, b) => b.lastUsedTime - a.lastUsedTime);

    // 优先匹配指定平台
    if (platform) {
      const same = allAccounts.find(a => a.platform === platform);
      if (same) {
        return {
          username: same.username,
          password: same.password,
          tenantId: '000000',
          platform
        };
      }
    }

    // 返回最新一个
    return {
      username: allAccounts[0].username,
      password: allAccounts[0].password,
      tenantId: '000000',
      platform: allAccounts[0].platform
    };
  } catch (e) {
    console.error('Get latest credentials error:', e);
    return null;
  }
};

// 检查是否已登录
export const isPlatformLoggedIn = (platform) => {
  const config = PLATFORMS[platform];
  if (!config) return false;
  const token = localStorage.getItem(config.tokenKey);
  if (!token) return false;

  // 检查凭据中的 token 过期时间
  const cred = getCredentials(platform);
  if (cred && cred.tokenExpire && Date.now() > cred.tokenExpire) {
    return false;
  }
  return true;
};

// 全局登录后同步（确保所有入口登录成功后状态一致）
const syncGlobalLoginState = (platform, username, passwordMd5, accessToken, expiresIn) => {
  const config = PLATFORMS[platform];
  if (!config) return;

  const tokenExpire = Date.now() + (expiresIn || 7200) * 1000;

  // 1. 保存 token 到对应 key
  localStorage.setItem(config.tokenKey, accessToken);

  // 2. 保存凭据到统一管理
  saveCredentials(platform, {
    tenantId: '000000',
    username,
    password: passwordMd5,
    accessToken,
    tokenExpire
  });

  // 3. 更新全局登录状态（触发 Header 刷新）
  updateLoginStatus(platform, true);

  // 4. mp 平台额外同步：兼容 DataQuery.vue 等页面读取 mp_token
  if (platform === 'mp') {
    localStorage.setItem('mp_token', accessToken);
    localStorage.setItem('mp_token_expire', String(tokenExpire));
    localStorage.setItem('mp_username', username);
  }

  // 5. iot/smart 平台联动：iot 登录 → smart 也标记为已登录
  if (platform === 'iot' || platform === 'smart') {
    updateLoginStatus('smart', true);
    localStorage.setItem(PLATFORMS.smart.tokenKey, accessToken);
    saveCredentials('smart', {
      tenantId: '000000',
      username,
      password: passwordMd5,
      accessToken,
      tokenExpire
    });
    // 兼容 FactoryDataQuery.vue 读取的 key
    if (platform === 'iot') {
      localStorage.setItem('iot_token', accessToken);
    }
  }

  // 6. 保存到账号历史
  const listKey = platform === 'mp' ? 'api_accounts_list' : 'factory_accounts_list';
  try {
    const list = JSON.parse(localStorage.getItem(listKey) || '[]');
    const now = Date.now();
    const existingIndex = list.findIndex(item => item.username === username);
    if (existingIndex > -1) {
      list[existingIndex].password = passwordMd5;
      list[existingIndex].lastUsedTime = now;
    } else {
      list.push({ username, password: passwordMd5, lastUsedTime: now });
    }
    list.sort((a, b) => b.lastUsedTime - a.lastUsedTime);
    if (list.length > 10) list.splice(10);
    localStorage.setItem(listKey, JSON.stringify(list));
  } catch (e) {
    // ignore
  }

  // 7. 保存到跨平台最新登录记录
  localStorage.setItem('unified_latest_login', JSON.stringify({
    username,
    password: passwordMd5,
    platform,
    lastUsedTime: Date.now()
  }));

  console.log(`[全局登录同步] ${config.name} 登录成功，Header 状态已同步`);
};

// 执行登录
export const performLogin = async (platform, username, password) => {
  const config = PLATFORMS[platform];
  if (!config) {
    return { success: false, message: '未知平台' };
  }

  if (!username || !password) {
    return { success: false, message: '请输入用户名和密码' };
  }

  try {
    const passwordMd5 = CryptoJS.MD5(password).toString();
    const params = new URLSearchParams();
    params.append('tenantId', '000000');
    params.append('username', username);
    params.append('password', passwordMd5);
    params.append('grant_type', 'password');
    params.append('scope', 'all');
    params.append('type', 'account');

    // 处理URL
    let url = config.loginUrl;
    if (!url.startsWith('http')) {
      // 智能补全代理前缀
      const proxyMap = {
        mp: '',
        smart: '/smart-api',
        iot: '/iot-api'
      };
      const proxy = proxyMap[platform] || '';
      if (proxy && !url.startsWith(proxy + '/')) {
        url = proxy + url;
      }
    }

    const response = await fetch(url + '?' + params.toString(), {
      method: 'POST',
      headers: {
        'Authorization': config.authHeader,
        'Accept': 'application/json, text/plain, */*',
        'Content-Type': 'application/x-www-form-urlencoded',
        'Tenant-Id': '000000'
      }
    });

    const result = await response.json();
    if (result.code === 200 && result.data && result.data.access_token) {
      // 统一同步全局登录状态（Header、token、凭据、历史记录）
      syncGlobalLoginState(platform, username, passwordMd5, result.data.access_token, result.data.expires_in);

      return { success: true, data: result.data };
    }
    return { success: false, message: result.msg || '登录失败' };
  } catch (error) {
    return { success: false, message: '登录请求失败：' + error.message };
  }
};

// 暴露同步函数供外部直接调用（用于已有独立登录逻辑的页面改造）
export { syncGlobalLoginState };

// 登录弹窗状态（全局共享）
export const loginDialogState = reactive({
  visible: false,
  platform: 'mp',
  username: '',
  password: '',
  loading: false,
  errorMessage: '',
  onSuccess: null,
  redirectOnCancel: '/auth/third-party',
  showPassword: false
});

// 关闭登录弹窗
export const closeLoginDialog = (redirectTo = null) => {
  loginDialogState.visible = false;
  loginDialogState.errorMessage = '';
  loginDialogState.loading = false;
  if (redirectTo) {
    // 跳转到指定页面
    setTimeout(() => {
      window.location.href = '#' + redirectTo;
    }, 100);
  }
};

// 显示登录弹窗
export const showLoginDialog = (platform, options = {}) => {
  const config = PLATFORMS[platform];
  if (!config) {
    console.error('Unknown platform:', platform);
    return Promise.reject(new Error('Unknown platform'));
  }

  // 获取最近一次登录的账号密码
  const latest = getLatestCredentials(platform);

  loginDialogState.platform = platform;
  loginDialogState.username = latest ? latest.username : '';
  loginDialogState.password = latest ? latest.password : '';
  loginDialogState.errorMessage = '';
  loginDialogState.loading = false;
  loginDialogState.visible = true;
  loginDialogState.showPassword = false;
  loginDialogState.redirectOnCancel = options.redirectOnCancel || '/auth/third-party';
  loginDialogState.onSuccess = options.onSuccess || null;

  return new Promise((resolve) => {
    loginDialogState._resolve = resolve;
  });
};

// 处理登录弹窗的登录按钮
export const handleDialogLogin = async () => {
  if (loginDialogState.loading) return;
  loginDialogState.loading = true;
  loginDialogState.errorMessage = '';

  const result = await performLogin(
    loginDialogState.platform,
    loginDialogState.username,
    loginDialogState.password
  );

  loginDialogState.loading = false;

  if (result.success) {
    ElMessage.success(`${PLATFORMS[loginDialogState.platform].name}登录成功`);
    loginDialogState.visible = false;
    if (loginDialogState.onSuccess) {
      try {
        loginDialogState.onSuccess();
      } catch (e) {
        console.error('onSuccess callback error:', e);
      }
    }
    if (loginDialogState._resolve) {
      loginDialogState._resolve({ success: true });
      loginDialogState._resolve = null;
    }
  } else {
    loginDialogState.errorMessage = result.message;
    if (loginDialogState._resolve) {
      loginDialogState._resolve({ success: false, message: result.message });
    }
  }
};

// 跳转到三方授权页面
export const goToThirdPartyAuth = (platform) => {
  loginDialogState.visible = false;
  const path = loginDialogState.redirectOnCancel;
  if (window.location.hash !== '#' + path) {
    window.location.href = '#' + path;
  }
};

// 确保已登录：未登录则弹窗，失败则跳转
export const ensurePlatformLoggedIn = async (platform, options = {}) => {
  if (isPlatformLoggedIn(platform)) {
    return { success: true, alreadyLoggedIn: true };
  }

  // 检查是否有最近账号
  const latest = getLatestCredentials(platform);
  if (!latest) {
    // 没有最近账号，跳转到三方授权页面
    ElMessage.warning(`未找到${PLATFORMS[platform].name}的登录凭据，请先登录`);
    setTimeout(() => {
      window.location.href = '#' + (options.redirectOnCancel || '/auth/third-party');
    }, 500);
    return { success: false, reason: 'no_credentials' };
  }

  // 显示登录弹窗
  const result = await showLoginDialog(platform, options);
  return result;
};

export { PLATFORMS };
