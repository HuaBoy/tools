import { ElMessage } from 'element-plus';
import {
  saveCredentials,
  getCredentials,
  updateLoginStatus
} from './loginStatus.js';
import { showLoginDialog, getLatestCredentials, isPlatformLoggedIn } from './platformLogin.js';

// 平台配置
const PLATFORMS = {
  mp: {
    name: '云系统',
    loginUrl: '/api/blade-auth/oauth/token',
    proxy: '',  // 走/api代理
    tokenKey: 'mp_token',
    authHeader: 'Basic ' + btoa('saber:saber_secret')
  },
  smart: {
    name: '智能制造平台',
    loginUrl: '/smart-api/api/blade-auth/oauth/token',
    proxy: '/smart-api',
    tokenKey: 'smart_factory_token',
    authHeader: 'Basic ' + btoa('saber_identity_client:saber_identity_secret')
  },
  iot: {
    name: '工厂数据平台',
    loginUrl: '/iot-api/api/blade-auth/oauth/token',
    proxy: '/iot-api',
    tokenKey: 'iot_token',
    authHeader: 'Basic ' + btoa('saber_identity_client:saber_identity_secret')
  }
};

// 检测API响应是否表示登录失效
const isUnauthorized = (response, data) => {
  if (response.status === 401) return true;
  if (data && (data.code === 401 || data.code === '401')) return true;
  if (data && data.error === 'unauthorized') return true;
  if (data && data.msg && /登录失效|token失效|未登录|认证失败|access_denied/i.test(data.msg)) {
    return true;
  }
  return false;
};

// 跳转到三方授权页面
const redirectToAuth = (platform) => {
  const path = '/auth/third-party';
  if (!window.location.hash.endsWith(path)) {
    ElMessage.warning(`未找到${PLATFORMS[platform]?.name || '该平台'}的登录凭据，即将跳转到授权页面`);
    setTimeout(() => {
      window.location.href = '#' + path;
    }, 500);
  }
};

// 401自动重新登录（增强版：支持弹窗和跳转到授权页）
const autoRelogin = async (platform) => {
  const config = PLATFORMS[platform];
  if (!config) return false;

  // 智能兼容：iot 平台的凭据可复用 smart 平台的
  let credentials = getCredentials(platform);
  if ((!credentials || !credentials.username || !credentials.password) && platform === 'iot') {
    credentials = getCredentials('smart');
  }

  // 如果没有凭据，尝试使用最近一次登录的账号密码
  if (!credentials || !credentials.username || !credentials.password) {
    const latest = getLatestCredentials(platform);
    if (latest && latest.username && latest.password) {
      credentials = {
        username: latest.username,
        password: latest.password,
        tenantId: '000000'
      };
    } else {
      // 没有最近账号，跳转到三方授权页面
      redirectToAuth(platform);
      return false;
    }
  }

  try {
    // 调用登录接口
    const loginResponse = await fetch(config.loginUrl, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
        'Authorization': config.authHeader,
        'Tenant-Id': credentials.tenantId || '000000'
      },
      body: new URLSearchParams({
        tenantId: credentials.tenantId || '000000',
        username: credentials.username,
        password: credentials.password,
        grant_type: 'password',
        scope: 'all',
        type: 'account'
      })
    });

    const result = await loginResponse.json();
    if (result.code === 200 && result.data && result.data.access_token) {
      // 更新token
      localStorage.setItem(config.tokenKey, result.data.access_token);
      // 更新凭据中的token
      saveCredentials(platform, {
        ...credentials,
        accessToken: result.data.access_token,
        tokenExpire: Date.now() + (result.data.expires_in || 7200) * 1000
      });
      // 同步登录状态
      updateLoginStatus(platform, true);

      // iot 平台登录同时更新 smart 状态和凭据
      if (platform === 'iot') {
        updateLoginStatus('smart', true);
        // 同步 smart 的 tokenKey
        localStorage.setItem(PLATFORMS.smart.tokenKey, result.data.access_token);
        // 同步 smart 的凭据
        saveCredentials('smart', {
          ...credentials,
          accessToken: result.data.access_token,
          tokenExpire: Date.now() + (result.data.expires_in || 7200) * 1000
        });
      }

      return true;
    }
    return false;
  } catch (e) {
    console.error(`Auto relogin ${platform} failed:`, e);
    return false;
  }
};

// 带401自动重新登录的请求函数
export const requestWithAutoRelogin = async (platform, options) => {
  const config = PLATFORMS[platform];
  if (!config) {
    throw new Error(`Unknown platform: ${platform}`);
  }

  const execute = async (isRetry = false) => {
    // 获取当前token
    const token = localStorage.getItem(config.tokenKey);
    const headers = {
      'Accept': 'application/json, text/plain, */*',
      ...options.headers
    };

    if (token) {
      headers['Blade-Auth'] = `bearer ${token}`;
    }

    // 处理URL
    let url = options.url;
    if (!url.startsWith('http')) {
      const proxy = config.proxy || '';
      if (proxy && !url.startsWith(proxy + '/')) {
        url = proxy + url;
      }
    }

    const response = await fetch(url, {
      ...options,
      headers
    });

    let data = null;
    try {
      const contentType = response.headers.get('content-type');
      if (contentType && contentType.includes('application/json')) {
        data = await response.json();
      }
    } catch (e) {
      // ignore
    }

    // 检测401
    if (isUnauthorized(response, data)) {
      updateLoginStatus(platform, false);

      if (!isRetry) {
        // 检查是否有最近凭据
        const latest = getLatestCredentials(platform);
        if (latest) {
          // 有最近凭据，自动登录
          const reloginSuccess = await autoRelogin(platform);
          if (reloginSuccess) {
            return execute(true);
          }
        }

        // 没有最近凭据或自动登录失败，弹出登录对话框
        const result = await showLoginDialog(platform);
        if (result && result.success) {
          // 用户在弹窗中成功登录
          return execute(true);
        }
      }
    } else if (data && data.code === 200) {
      updateLoginStatus(platform, true);
    }

    return { response, data };
  };

  return execute(false);
};

// 通用请求函数（带401自动重登）
export const safeRequest = async (platform, url, options = {}) => {
  const result = await requestWithAutoRelogin(platform, {
    url,
    ...options
  });
  return result.data;
};

// 主动调用：发起请求前检查登录状态
export const requestWithLoginCheck = async (platform, url, options = {}) => {
  // 检查登录状态
  if (!isPlatformLoggedIn(platform)) {
    // 未登录：检查是否有最近凭据
    const latest = getLatestCredentials(platform);
    if (latest) {
      // 有最近凭据，弹出登录对话框
      const result = await showLoginDialog(platform);
      if (!result || !result.success) {
        // 用户取消或登录失败
        throw new Error('未登录');
      }
    } else {
      // 没有最近凭据，跳转到授权页
      redirectToAuth(platform);
      throw new Error('未找到登录凭据');
    }
  }

  return safeRequest(platform, url, options);
};
