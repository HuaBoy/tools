import { defineStore } from 'pinia';
import { supabase } from '@/utils/supabase';
import { localUserStore } from '@/utils/localUserStore';

export const useAuthStore = defineStore('auth', {
  state: () => ({
    isLoggedIn: false,
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
      const user = localUserStore.validateLogin(username, password);
      
      if (user) {
        this.isLoggedIn = true;
        this.user = { id: user.id, email: user.email, username: user.username };
        this.role = user.role;
        this.permissions = user.permissions || [];
        this.lastActivityTime = Date.now();
        localStorage.setItem('auth_user', user.username);
        localStorage.setItem('last_activity', String(Date.now()));
        return true;
      }

      return false;
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
    },

    refreshActivity() {
      if (this.isLoggedIn) {
        this.lastActivityTime = Date.now();
        localStorage.setItem('last_activity', String(Date.now()));
      }
    },

    async loadAuthState() {
      const savedUser = localStorage.getItem('auth_user');
      const savedTime = localStorage.getItem('last_activity');
      
      if (savedUser && savedTime) {
        const lastTime = parseInt(savedTime, 10);
        const twoDaysInMs = 2 * 24 * 60 * 60 * 1000;
        
        if (Date.now() - lastTime <= twoDaysInMs) {
          const user = localUserStore.getUserByUsername(savedUser);
          if (user) {
            this.isLoggedIn = true;
            this.user = { id: user.id, email: user.email, username: user.username };
            this.role = user.role;
            this.permissions = user.permissions || [];
            this.lastActivityTime = lastTime;
            return true;
          }
        } else {
          this.logout();
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