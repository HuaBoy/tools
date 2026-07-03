/**
 * 全局平台自动登录工具
 * 每 5 小时自动检查并登录云系统和智能制造系统
 */
import CryptoJS from 'crypto-js';
import { saveCredentials, updateLoginStatus } from './loginStatus.js';

const AUTO_LOGIN_INTERVAL = 5 * 60 * 60 * 1000; // 5 小时
const AUTO_LOGIN_TIME_KEY = 'auto_login_last_time';
const API_LOGIN_URL = 'https://mp.holyview.cn:9443/api/blade-auth/oauth/token';
const FACTORY_LOGIN_URL = '/iot-api/api/blade-auth/oauth/token';
const API_AUTH_HEADER = 'Basic ' + btoa('saber:saber_secret');
const FACTORY_AUTH_HEADER = 'Basic ' + btoa('saber_identity_client:saber_identity_secret');

/**
 * 云系统自动登录
 * @param {string} username - 用户名
 * @param {string} password - 密码（明文）
 */
const autoLoginApi = async (username, password) => {
  if (!username || !password) return { success: false, reason: 'no_credentials' };
  try {
    const passwordMd5 = CryptoJS.MD5(password).toString();
    const params = new URLSearchParams();
    params.append('tenantId', '000000');
    params.append('username', username);
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
      // 保存 token
      localStorage.setItem('mp_token', result.data.access_token);
      // 保存凭据
      saveCredentials('mp', {
        tenantId: '000000',
        username,
        password: passwordMd5,
        accessToken: result.data.access_token,
        tokenExpire: Date.now() + (result.data.expires_in || 7200) * 1000
      });
      updateLoginStatus('mp', true);
      return { success: true };
    }
    return { success: false, message: result.msg || '登录失败' };
  } catch (e) {
    return { success: false, message: e.message };
  }
};

/**
 * 智能制造系统自动登录
 * @param {string} username - 用户名
 * @param {string} password - 密码（明文）
 */
const autoLoginFactory = async (username, password) => {
  if (!username || !password) return { success: false, reason: 'no_credentials' };
  try {
    const passwordMd5 = CryptoJS.MD5(password).toString();
    const params = new URLSearchParams();
    params.append('tenantId', '000000');
    params.append('username', username);
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
      // 保存 token
      localStorage.setItem('smart_factory_token', result.data.access_token);
      // 保存凭据
      saveCredentials('smart', {
        tenantId: '000000',
        username,
        password: passwordMd5,
        accessToken: result.data.access_token,
        tokenExpire: Date.now() + (result.data.expires_in || 7200) * 1000
      });
      updateLoginStatus('smart', true);
      return { success: true };
    }
    return { success: false, message: result.msg || '登录失败' };
  } catch (e) {
    return { success: false, message: e.message };
  }
};

/**
 * 获取账号历史中的最新账号
 * @param {string} listKey - 账号历史列表 key
 */
const getLatestFromHistory = (listKey) => {
  try {
    const list = JSON.parse(localStorage.getItem(listKey) || '[]');
    if (list.length > 0) {
      return list[0]; // 已按 lastUsedTime 倒序
    }
  } catch (e) {
    // ignore
  }
  return null;
};

/**
 * 执行全局自动登录
 * 检查时间间隔，如果超过 5 小时则执行登录
 */
export const performPlatformLogin = async () => {
  try {
    const lastTime = parseInt(localStorage.getItem(AUTO_LOGIN_TIME_KEY) || '0');
    const now = Date.now();
    // 如果距离上次自动登录不足 5 小时，跳过
    if (lastTime && (now - lastTime < AUTO_LOGIN_INTERVAL)) {
      return;
    }

    // 标记本次检查时间
    localStorage.setItem(AUTO_LOGIN_TIME_KEY, String(now));

    // 云系统自动登录
    const apiAccount = getLatestFromHistory('api_accounts_list');
    if (apiAccount) {
      const result = await autoLoginApi(apiAccount.username, apiAccount.password);
      if (result.success) {
        console.log('[自动登录] 云系统登录成功');
      } else {
        console.warn('[自动登录] 云系统登录失败:', result.message);
      }
    }

    // 智能制造系统自动登录
    const factoryAccount = getLatestFromHistory('factory_accounts_list');
    if (factoryAccount) {
      const result = await autoLoginFactory(factoryAccount.username, factoryAccount.password);
      if (result.success) {
        console.log('[自动登录] 智能制造系统登录成功');
      } else {
        console.warn('[自动登录] 智能制造系统登录失败:', result.message);
      }
    }
  } catch (e) {
    console.error('[自动登录] 执行失败:', e);
  }
};

export default {
  performPlatformLogin
};
