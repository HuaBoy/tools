import { ref, reactive } from 'vue';

// 第三方账号登录状态管理（用于顶部菜单栏同步）
// 直接导出reactive对象，使其在所有组件间共享响应式状态
export const loginStatus = reactive({
  mp: false,        // mp.holyview.cn (主平台/云平台)
  smart: false,     // 218.90.146.230 (智能制造平台)
  factory: false    // 工厂数据查询
});

// 状态变化监听器
const listeners = new Set();

export const updateLoginStatus = (platform, status) => {
  if (loginStatus[platform] !== undefined) {
    loginStatus[platform] = status;
    // 通知所有监听器
    listeners.forEach(callback => {
      try {
        callback(platform, status);
      } catch (e) {
        console.error('Login status listener error:', e);
      }
    });
    // 触发 window 自定义事件，跨标签/跨组件同步
    try {
      window.dispatchEvent(new CustomEvent('login-status-changed', {
        detail: { platform, status }
      }));
    } catch (e) {
      // ignore
    }
  }
};

export const onLoginStatusChange = (callback) => {
  listeners.add(callback);
  return () => listeners.delete(callback);
};

export const getLoginStatus = (platform) => {
  return loginStatus[platform] || false;
};

export const getAllLoginStatus = () => {
  return { ...loginStatus };
};

// 凭据存储
const CREDENTIALS_KEY = 'third_party_credentials';

export const saveCredentials = (platform, credentials) => {
  try {
    const all = JSON.parse(localStorage.getItem(CREDENTIALS_KEY) || '{}');
    all[platform] = {
      ...credentials,
      saveTime: Date.now()
    };
    localStorage.setItem(CREDENTIALS_KEY, JSON.stringify(all));
  } catch (e) {
    console.error('Save credentials error:', e);
  }
};

export const getCredentials = (platform) => {
  try {
    const all = JSON.parse(localStorage.getItem(CREDENTIALS_KEY) || '{}');
    return all[platform] || null;
  } catch (e) {
    return null;
  }
};

export const clearCredentials = (platform) => {
  try {
    const all = JSON.parse(localStorage.getItem(CREDENTIALS_KEY) || '{}');
    delete all[platform];
    localStorage.setItem(CREDENTIALS_KEY, JSON.stringify(all));
  } catch (e) {
    console.error('Clear credentials error:', e);
  }
};

// 初始化时从localStorage恢复登录状态
export const initLoginStatus = () => {
  // 检查凭据管理中是否有保存的凭据
  ['mp', 'smart', 'factory'].forEach(platform => {
    const cred = getCredentials(platform);
    if (cred && cred.accessToken) {
      loginStatus[platform] = true;
    }
  });

  // 兼容旧的localStorage key
  try {
    // API测试助手
    const testerCred = localStorage.getItem('tester_credentials');
    if (testerCred) {
      const data = JSON.parse(testerCred);
      if (data.expireTime && Date.now() < data.expireTime) {
        loginStatus.mp = true;
      }
    }
    // 工厂数据查询
    const factoryCred = localStorage.getItem('factory_data_login_record');
    if (factoryCred) {
      const data = JSON.parse(factoryCred);
      if (data.expireTime && Date.now() < data.expireTime) {
        loginStatus.smart = true;
      }
    }
  } catch (e) {
    console.error('Init login status error:', e);
  }
};
