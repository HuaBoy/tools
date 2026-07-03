import { reactive } from 'vue';

const STORAGE_KEY = 'local_users';

const defaultUsers = [
  {
    id: 'admin-001',
    username: 'admin',
    email: 'admin@example.com',
    password: '123456',
    role: 'admin',
    permissions: ['user:view', 'user:create', 'user:edit', 'user:delete', 'role:assign', 'permission:manage', 'feature:toggle', 'data:query', 'data:export', 'log:view'],
    created_at: new Date('2026-01-01').toISOString(),
    updated_at: new Date('2026-01-01').toISOString()
  },
  {
    id: 'user-001',
    username: 'testuser',
    email: 'test@example.com',
    password: '123456',
    role: 'user',
    permissions: ['data:query', 'data:export'],
    created_at: new Date('2026-01-02').toISOString(),
    updated_at: new Date('2026-01-02').toISOString()
  }
];

const users = reactive({});

const loadUsers = () => {
  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored) {
      const parsed = JSON.parse(stored);
      Object.keys(parsed).forEach(id => {
        users[id] = parsed[id];
      });
    } else {
      defaultUsers.forEach(user => {
        users[user.id] = { ...user };
      });
      saveUsers();
    }
  } catch (e) {
    console.error('加载用户数据失败:', e);
    defaultUsers.forEach(user => {
      users[user.id] = { ...user };
    });
  }
};

const saveUsers = () => {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(users));
  } catch (e) {
    console.error('保存用户数据失败:', e);
  }
};

const generateId = () => {
  return 'user-' + Date.now().toString(36) + Math.random().toString(36).substr(2, 5);
};

export const localUserStore = {
  getAllUsers() {
    return Object.values(users);
  },

  getUserById(id) {
    return users[id] || null;
  },

  getUserByUsername(username) {
    return Object.values(users).find(u => u.username === username) || null;
  },

  getUserByEmail(email) {
    return Object.values(users).find(u => u.email === email) || null;
  },

  createUser(userData) {
    const now = new Date().toISOString();
    const newUser = {
      id: generateId(),
      username: userData.username,
      email: userData.email || '',
      password: userData.password || '123456',
      role: userData.role || 'user',
      permissions: userData.permissions || [],
      created_at: now,
      updated_at: now
    };

    if (this.getUserByUsername(newUser.username)) {
      throw new Error('用户名已存在');
    }

    if (newUser.email && this.getUserByEmail(newUser.email)) {
      throw new Error('邮箱已存在');
    }

    users[newUser.id] = newUser;
    saveUsers();
    return newUser;
  },

  updateUser(id, updates) {
    if (!users[id]) {
      throw new Error('用户不存在');
    }

    const user = users[id];

    if (updates.username && updates.username !== user.username) {
      const existing = this.getUserByUsername(updates.username);
      if (existing && existing.id !== id) {
        throw new Error('用户名已存在');
      }
    }

    if (updates.email && updates.email !== user.email) {
      const existing = this.getUserByEmail(updates.email);
      if (existing && existing.id !== id) {
        throw new Error('邮箱已存在');
      }
    }

    Object.keys(updates).forEach(key => {
      if (key !== 'id') {
        user[key] = updates[key];
      }
    });
    user.updated_at = new Date().toISOString();

    saveUsers();
    return user;
  },

  deleteUser(id) {
    if (!users[id]) {
      throw new Error('用户不存在');
    }

    const deleted = users[id];
    delete users[id];
    saveUsers();
    return deleted;
  },

  validateLogin(username, password) {
    const user = this.getUserByUsername(username);
    if (!user) {
      return null;
    }
    if (user.password === password) {
      return user;
    }
    return null;
  },

  resetPassword(id, newPassword) {
    if (!users[id]) {
      throw new Error('用户不存在');
    }
    users[id].password = newPassword;
    users[id].updated_at = new Date().toISOString();
    saveUsers();
    return users[id];
  },

  searchUsers(keyword, role = '') {
    let results = Object.values(users);

    if (keyword) {
      const lowerKeyword = keyword.toLowerCase();
      results = results.filter(u =>
        u.username.toLowerCase().includes(lowerKeyword) ||
        u.email.toLowerCase().includes(lowerKeyword)
      );
    }

    if (role) {
      results = results.filter(u => u.role === role);
    }

    return results.sort((a, b) => new Date(b.created_at) - new Date(a.created_at));
  },

  paginateUsers(page = 1, limit = 10, keyword = '', role = '') {
    const filtered = this.searchUsers(keyword, role);
    const total = filtered.length;
    const start = (page - 1) * limit;
    const end = start + limit;
    const data = filtered.slice(start, end);

    return {
      data,
      total,
      current: page,
      size: limit,
      pages: Math.ceil(total / limit)
    };
  },

  getRoles() {
    return [
      { id: 'admin', name: '超级管理员', description: '拥有所有权限' },
      { id: 'manager', name: '管理员', description: '管理用户和权限' },
      { id: 'editor', name: '编辑', description: '可编辑内容' },
      { id: 'viewer', name: '查看者', description: '仅查看权限' },
      { id: 'user', name: '普通用户', description: '基础操作权限' }
    ];
  },

  getPermissions() {
    return [
      { id: 'user:view', name: '查看用户', description: '查看用户列表' },
      { id: 'user:create', name: '创建用户', description: '创建新用户' },
      { id: 'user:edit', name: '编辑用户', description: '编辑用户信息' },
      { id: 'user:delete', name: '删除用户', description: '删除用户' },
      { id: 'role:view', name: '查看角色', description: '查看角色列表' },
      { id: 'role:assign', name: '分配角色', description: '为用户分配角色' },
      { id: 'permission:manage', name: '权限管理', description: '管理权限' },
      { id: 'feature:view', name: '查看功能', description: '查看功能列表' },
      { id: 'feature:toggle', name: '开关功能', description: '开启/关闭功能' },
      { id: 'data:query', name: '数据查询', description: '查询数据' },
      { id: 'data:export', name: '数据导出', description: '导出数据' },
      { id: 'log:view', name: '查看日志', description: '查看操作日志' }
    ];
  },

  getRolePermissions(roleId) {
    const rolePermissions = {
      admin: ['user:view', 'user:create', 'user:edit', 'user:delete', 'role:view', 'role:assign', 'permission:manage', 'feature:view', 'feature:toggle', 'data:query', 'data:export', 'log:view'],
      manager: ['user:view', 'user:create', 'user:edit', 'user:delete', 'role:view', 'role:assign', 'feature:view', 'feature:toggle', 'data:query', 'data:export'],
      editor: ['data:query', 'data:export', 'feature:view'],
      viewer: ['data:query', 'feature:view'],
      user: ['data:query']
    };
    return rolePermissions[roleId] || [];
  }
};

loadUsers();