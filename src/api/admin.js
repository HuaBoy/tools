import { localUserStore } from '@/utils/localUserStore';

const apiBase = '/api/v1';

const request = async (url, options = {}) => {
  const token = localStorage.getItem('auth_token');
  const headers = {
    'Content-Type': 'application/json',
    ...options.headers
  };
  if (token) {
    headers['Authorization'] = 'Bearer ' + token;
  }

  const response = await fetch(apiBase + url, {
    ...options,
    headers
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({ message: '请求失败' }));
    throw new Error(errorData.message || '请求失败');
  }

  return response.json();
};

const withFallback = async (apiFn, fallbackFn) => {
  try {
    return await apiFn();
  } catch (error) {
    console.warn('API请求失败，使用本地存储:', error.message);
    return await fallbackFn();
  }
};

export const adminApi = {
  async getUsers(page = 1, limit = 10, search = '', role = '') {
    return withFallback(
      async () => {
        const params = new URLSearchParams({
          current: page,
          size: limit,
          keyword: search,
          role: role
        });
        return request('/users?' + params);
      },
      async () => localUserStore.paginateUsers(page, limit, search, role)
    );
  },

  async getUserById(id) {
    return withFallback(
      async () => request('/users/' + id),
      async () => {
        const user = localUserStore.getUserById(id);
        if (!user) throw new Error('用户不存在');
        return user;
      }
    );
  },

  async createUser(userData) {
    return withFallback(
      async () => request('/users', { method: 'POST', body: JSON.stringify(userData) }),
      async () => localUserStore.createUser(userData)
    );
  },

  async updateUser(id, updates) {
    return withFallback(
      async () => request('/users/' + id, { method: 'PATCH', body: JSON.stringify(updates) }),
      async () => localUserStore.updateUser(id, updates)
    );
  },

  async deleteUser(id) {
    return withFallback(
      async () => request('/users/' + id, { method: 'DELETE' }),
      async () => {
        localUserStore.deleteUser(id);
        return true;
      }
    );
  },

  async resetPassword(id, newPassword) {
    return withFallback(
      async () => request('/users/' + id + '/reset-password', { method: 'POST', body: JSON.stringify({ new_password: newPassword }) }),
      async () => {
        localUserStore.resetPassword(id, newPassword);
        return { newPassword };
      }
    );
  },

  async updateRole(id, role) {
    return withFallback(
      async () => request('/users/' + id, { method: 'PATCH', body: JSON.stringify({ role }) }),
      async () => localUserStore.updateUser(id, { role })
    );
  },

  async getRoles() {
    return localUserStore.getRoles();
  },

  async getPermissions() {
    return localUserStore.getPermissions();
  },

  async updatePermissions(userId, permissions) {
    return withFallback(
      async () => request('/users/' + userId, { method: 'PATCH', body: JSON.stringify({ permissions }) }),
      async () => localUserStore.updateUser(userId, { permissions })
    );
  },

  async getFeatures() {
    const stored = localStorage.getItem('features');
    if (stored) {
      return JSON.parse(stored);
    }
    const defaultFeatures = [
      { id: 'auth_converter', name: '授权码转换', key: 'auth_converter', description: '支持多个授权码录入转换', enabled: true, order: 1 },
      { id: 'log_decrypt', name: '日志解密', key: 'log_decrypt', description: 'AES解密日志文件', enabled: true, order: 2 },
      { id: 'log_analysis', name: '起爆器日志AI分析', key: 'log_analysis', description: '解析日志文件，提取电流检测/雷管检测/起爆指令数据', enabled: true, order: 3 },
      { id: 'data_query', name: '批次数据追溯', key: 'data_query', description: '支持数据导出', enabled: true, order: 4 },
      { id: 'factory_data', name: '工厂数据查询', key: 'factory_data', description: '支持登录、数据保存、设备类型和批次号查询', enabled: true, order: 5 },
      { id: 'third_party_auth', name: '三方账号授权', key: 'third_party_auth', description: '整合API测试助手和工厂数据查询的登录功能', enabled: true, order: 6 },
      { id: 'knowledge_base', name: '智能知识库(RAG)', key: 'knowledge_base', description: 'AI问题查询', enabled: true, order: 7 },
      { id: 'ai_translate', name: 'AI翻译工具', key: 'ai_translate', description: '多语言翻译', enabled: true, order: 8 },
      { id: 'qr_code', name: '二维码生成', key: 'qr_code', description: '生成常规二维码', enabled: true, order: 9 },
      { id: 'ai_assistant', name: 'AI运维智能助手', key: 'ai_assistant', description: 'AI运维辅助', enabled: true, order: 10 }
    ];
    localStorage.setItem('features', JSON.stringify(defaultFeatures));
    return defaultFeatures;
  },

  async getFeatureById(id) {
    const features = await this.getFeatures();
    const feature = features.find(f => f.id === id);
    if (!feature) {
      throw new Error('功能不存在');
    }
    return feature;
  },

  async createFeature(feature) {
    const features = await this.getFeatures();
    const newFeature = {
      id: feature.id || 'feature-' + Date.now(),
      name: feature.name,
      key: feature.key || feature.name.toLowerCase().replace(/\s+/g, '_'),
      description: feature.description || '',
      enabled: feature.enabled !== undefined ? feature.enabled : true,
      order: feature.order || features.length + 1
    };
    features.push(newFeature);
    features.sort((a, b) => a.order - b.order);
    localStorage.setItem('features', JSON.stringify(features));
    return newFeature;
  },

  async updateFeature(id, updates) {
    const features = await this.getFeatures();
    const index = features.findIndex(f => f.id === id);
    if (index === -1) {
      throw new Error('功能不存在');
    }
    features[index] = { ...features[index], ...updates };
    features.sort((a, b) => a.order - b.order);
    localStorage.setItem('features', JSON.stringify(features));
    return features[index];
  },

  async deleteFeature(id) {
    let features = await this.getFeatures();
    features = features.filter(f => f.id !== id);
    features.forEach((f, i) => { f.order = i + 1; });
    localStorage.setItem('features', JSON.stringify(features));
    return true;
  },

  async toggleFeature(id, enabled) {
    const features = await this.getFeatures();
    const feature = features.find(f => f.id === id);
    if (!feature) {
      throw new Error('功能不存在');
    }
    feature.enabled = enabled;
    localStorage.setItem('features', JSON.stringify(features));
    return feature;
  }
};

export default adminApi;