import { defineStore } from 'pinia';
import { supabase } from '@/utils/supabase';
import { localUserStore } from '@/utils/localUserStore';

export const useAuthStore = defineStore('auth', {
  state: () => ({
    isLoggedIn: false,
    token: localStorage.getItem('auth_token') || null,
    user: null,
    userProfile: null,
    role: null,
    permissions: [],
    lastActivityTime: Date.now()
  }),

  getters: {
    isAdmin: (state) => state.role === 'admin',
    
    hasPermission: (state) => (permission) => {
      if (state.role === 'admin') return true;
      return state.permissions.includes(permission);
    },
    
    isSessionExpired: (state) => {
      const twoDaysInMs = 2 * 24 * 60 * 60 * 1000;
      return Date.now() - state.lastActivityTime > twoDaysInMs;
    },
    
    authStatus: (state) => {
      return state.isLoggedIn ? 'valid' : 'invalid';
    },
    
    statusText: (state) => {
      return state.isLoggedIn ? '授权有效' : '未授权';
    },
    
    remainingDays: (state) => {
      if (!state.isLoggedIn) return 0;
      const twoDaysInMs = 2 * 24 * 60 * 60 * 1000;
      const remainingMs = twoDaysInMs - (Date.now() - state.lastActivityTime);
      const remainingDays = Math.ceil(remainingMs / (24 * 60 * 60 * 1000));
      return Math.max(0, remainingDays);
    }
  },

  actions: {
    async login(username, password) {
      const response = await fetch('/api/v1/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password })
      });

      let result = {};
      try {
        result = await response.json();
      } catch (e) {
        throw new Error('服务器返回异常');
      }

      if (response.ok && result.code === 200 && result.data && result.data.token) {
        // 后端返回扁平结构：{ token, user_id, username, email, role, permissions }
        const { token, user_id, username: uname, email: uemail, role: urole, permissions: uperms } = result.data;
        const role = urole || 'user';
        this.isLoggedIn = true;
        this.token = token;
        this.user = {
          id: user_id,
          username: uname || username,
          email: uemail || ''
        };
        this.userProfile = this.user;
        this.role = role;
        this.permissions = role === 'admin' ? ['*'] : (uperms || []);
        this.lastActivityTime = Date.now();
        localStorage.setItem('auth_token', token);
        localStorage.setItem('auth_user', this.user.username);
        localStorage.setItem('auth_user_info', JSON.stringify({ ...this.user, role, permissions: this.permissions }));
        localStorage.setItem('last_activity', String(Date.now()));
        return true;
      }

      // 登录失败：抛出后端返回的错误信息
      throw new Error(result.message || '用户名或密码错误');
    },

    async register(email, password, username) {
      const { data, error } = await supabase.auth.signUp({
        email,
        password,
        options: {
          data: {
            username,
            role: 'user'
          }
        }
      });

      if (error) {
        throw new Error(error.message);
      }

      if (data.user) {
        await supabase.from('user_profiles').insert([
          {
            id: data.user.id,
            username,
            email,
            role: 'user',
            permissions: []
          }
        ]);
        return true;
      }
      return false;
    },

    async logout() {
      try {
        const { error } = await supabase.auth.signOut();
        if (error) {
          console.warn('Supabase 登出失败:', error.message);
        }
      } catch (e) {
        console.warn('Supabase 登出异常:', e.message);
      }
      this.reset();
    },

    async loadUserProfile() {
      if (!this.user) return;

      const { data, error } = await supabase
        .from('user_profiles')
        .select('*')
        .eq('id', this.user.id)
        .single();

      if (error) {
        console.warn('加载用户信息失败:', error.message);
        return;
      }

      if (data) {
        this.userProfile = data;
        this.role = data.role;
        this.permissions = data.permissions || [];
      }
    },

    reset() {
      this.isLoggedIn = false;
      this.user = null;
      this.userProfile = null;
      this.role = null;
      this.permissions = [];
      this.lastActivityTime = 0;
      localStorage.removeItem('last_activity');
      localStorage.removeItem('auth_token');
      localStorage.removeItem('auth_user');
      localStorage.removeItem('auth_user_info');
    },

    refreshActivity() {
      if (this.isLoggedIn) {
        this.lastActivityTime = Date.now();
        localStorage.setItem('last_activity', String(Date.now()));
      }
    },

    async loadAuthState() {
      const token = localStorage.getItem('auth_token');
      const savedInfo = localStorage.getItem('auth_user_info');
      const savedTime = localStorage.getItem('last_activity');

      if (token && savedInfo && savedTime) {
        const lastTime = parseInt(savedTime, 10);
        const twoDaysInMs = 2 * 24 * 60 * 60 * 1000;

        if (Date.now() - lastTime <= twoDaysInMs) {
          try {
            const info = JSON.parse(savedInfo);
            this.isLoggedIn = true;
            this.user = { id: info.id, username: info.username, email: info.email || '' };
            this.role = info.role;
            this.permissions = info.role === 'admin' ? ['*'] : [];
            this.lastActivityTime = lastTime;
            return true;
          } catch (e) {
            this.reset();
            return false;
          }
        } else {
          this.reset();
          return false;
        }
      }

      return false;
    },

    checkSession() {
      if (this.isLoggedIn && this.isSessionExpired) {
        this.logout();
        return true;
      }
      return false;
    }
  }
});