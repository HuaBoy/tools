// 统一对接 go-server 后端管理接口（/api）
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

  const response = await fetch(apiBase + url, {
    ...options,
    headers
  });

  const resp = await response.json().catch(() => ({}));
  // 业务码非 0 视为失败
  if (resp && resp.code !== undefined && resp.code !== 0) {
    throw new Error(resp.message || '请求失败');
  }
  return resp;
};

export const adminApi = {
  // ===================== 用户管理 =====================
  // 列表查询：每页 page_size 条，支持分页；keyword 按用户名/昵称模糊，roleId 按角色过滤
  async getUsers(page = 1, pageSize = 10, keyword = '', roleId = '') {
    const params = new URLSearchParams();
    params.set('page', page);
    params.set('page_size', pageSize);
    if (keyword) params.set('username', keyword);
    if (roleId) params.set('role_id', roleId);
    const resp = await request('/users?' + params.toString());
    return {
      data: (resp.data && resp.data.list) || [],
      total: (resp.data && resp.data.total) || 0
    };
  },

  async getUserById(id) {
    const resp = await request('/users/' + id);
    return resp.data;
  },

  // 新增用户：{ username, password, nickname, email, phone, role_id, factory_id, status }
  async createUser(data) {
    return request('/users', { method: 'POST', body: JSON.stringify(data) });
  },

  // 编辑用户（不修改密码，改密码走 resetPassword）：{ nickname, email, phone, role_id, factory_id, status }
  async updateUser(id, data) {
    return request('/users/' + id, { method: 'PUT', body: JSON.stringify(data) });
  },

  async deleteUser(id) {
    return request('/users/' + id, { method: 'DELETE' });
  },

  // 管理员重置密码：返回 { new_password }
  async resetPassword(id, newPassword) {
    const resp = await request('/users/' + id + '/reset-password', {
      method: 'POST',
      body: JSON.stringify({ new_password: newPassword || '' })
    });
    return resp.data;
  },

  // ===================== 角色（权限）管理 =====================
  async getRoles() {
    const resp = await request('/roles');
    return resp.data || [];
  },

  async getRole(id) {
    const resp = await request('/roles/' + id);
    return resp.data;
  },

  async createRole(data) {
    return request('/roles', { method: 'POST', body: JSON.stringify(data) });
  },

  async updateRole(id, data) {
    return request('/roles/' + id, { method: 'PUT', body: JSON.stringify(data) });
  },

  async deleteRole(id) {
    return request('/roles/' + id, { method: 'DELETE' });
  },

  // 为角色分配功能菜单（全量覆盖）
  async setRoleMenus(id, menuIds) {
    return request('/roles/' + id + '/menus', {
      method: 'POST',
      body: JSON.stringify({ menu_ids: menuIds })
    });
  },

  // ===================== 功能（菜单）管理 =====================
  // 返回树形结构
  async getMenus() {
    const resp = await request('/menus');
    return resp.data || [];
  },

  async createMenu(data) {
    return request('/menus', { method: 'POST', body: JSON.stringify(data) });
  },

  async updateMenu(id, data) {
    return request('/menus/' + id, { method: 'PUT', body: JSON.stringify(data) });
  },

  async deleteMenu(id) {
    return request('/menus/' + id, { method: 'DELETE' });
  }
};

export default adminApi;
