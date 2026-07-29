// 版本履历接口（对接 go-server /api/versions）
const apiBase = '/api';

const request = async (url, options = {}) => {
  const token = localStorage.getItem('auth_token');
  const headers = {
    'Content-Type': 'application/json',
    ...options.headers
  };
  if (token) {
    headers['Authorization'] = 'Bearer ' + token;
  }
  const response = await fetch(apiBase + url, { ...options, headers });
  const resp = await response.json().catch(() => ({}));
  if (resp && resp.code !== undefined && resp.code !== 0) {
    throw new Error(resp.message || '请求失败');
  }
  return resp;
};

// 将后端返回的记录转换为前端使用的字段：
// - id 映射为版本号字符串（前端模板用 v.id 显示版本号）
// - date 由 release_date 映射（模板用 v.date）
// - dbId 保留数据库主键，用于编辑/删除
const normalize = (v) => ({
  ...v,
  id: v.version_no,
  date: v.release_date,
  dbId: v.id
});

export const versionApi = {
  // 列表 / 筛选：region(domestic/overseas)、app_version、used_device、keyword
  async getVersions(params = {}) {
    const q = new URLSearchParams();
    if (params.region) q.set('region', params.region);
    if (params.app_version) q.set('app_version', params.app_version);
    if (params.used_device) q.set('used_device', params.used_device);
    if (params.keyword) q.set('keyword', params.keyword);
    const qs = q.toString();
    const resp = await request('/versions' + (qs ? '?' + qs : ''));
    const list = (resp.data || []).map(normalize);
    return list;
  },

  async getVersion(id) {
    const resp = await request('/versions/' + id);
    return resp.data;
  },

  // 新增：{ region, version_no, title, app_version, controller_version, used_device, device_type_color, release_date, features, source }
  async createVersion(data) {
    return request('/versions', { method: 'POST', body: JSON.stringify(data) });
  },

  async updateVersion(id, data) {
    return request('/versions/' + id, { method: 'PUT', body: JSON.stringify(data) });
  },

  async deleteVersion(id) {
    return request('/versions/' + id, { method: 'DELETE' });
  },

  // 批量导入：{ items: [...] }
  async importVersions(items) {
    return request('/versions/import', { method: 'POST', body: JSON.stringify({ items }) });
  }
};
