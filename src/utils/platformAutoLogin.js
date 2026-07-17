/**
 * 全局平台自动登录工具
 * 每 5 小时自动检查并登录云系统和智能制造系统
 * 使用统一的 performLogin，确保全局状态同步
 */
import { performLogin } from './platformLogin.js';

const AUTO_LOGIN_INTERVAL = 5 * 60 * 60 * 1000; // 5 小时
const AUTO_LOGIN_TIME_KEY = 'auto_login_last_time';

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

    // 云系统自动登录（使用统一 performLogin，自动同步全局状态）
    const apiAccount = getLatestFromHistory('api_accounts_list');
    if (apiAccount) {
      const result = await performLogin('mp', apiAccount.username, apiAccount.password);
      if (result.success) {
        console.log('[自动登录] 云系统登录成功，Header 状态已同步');
      } else {
        console.warn('[自动登录] 云系统登录失败:', result.message);
      }
    }

    // 智能制造系统自动登录（使用统一 performLogin，自动同步全局状态）
    const factoryAccount = getLatestFromHistory('factory_accounts_list');
    if (factoryAccount) {
      const result = await performLogin('iot', factoryAccount.username, factoryAccount.password);
      if (result.success) {
        console.log('[自动登录] 智能制造系统登录成功，Header 状态已同步');
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
