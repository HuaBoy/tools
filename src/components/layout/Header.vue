<script setup>
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { useAuthStore } from '@/stores/auth';
import { ElMessage, ElMessageBox } from 'element-plus';

const emit = defineEmits(['toggle-sidebar']);
const router = useRouter();

const authStore = useAuthStore();
const searchQuery = ref('');

const handleSearch = () => {
  if (searchQuery.value.trim()) {
    ElMessage.info(`搜索: ${searchQuery.value}`);
  }
};

const toggleSidebar = () => {
  emit('toggle-sidebar');
};

const handleLogout = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要退出系统吗？退出后将清除所有缓存数据。',
      '退出确认',
      {
        confirmButtonText: '确定退出',
        cancelButtonText: '取消',
        type: 'warning'
      }
    );

    // 调用后端退出接口（无状态 JWT，失败也允许本地退出）
    const token = localStorage.getItem('auth_token');
    if (token) {
      try {
        await fetch('/api/logout', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Bearer ' + token
          }
        });
      } catch (e) {
        console.warn('调用退出接口失败，仍执行本地退出:', e);
      }
    }

    // 清除所有缓存
    localStorage.removeItem('auth_user');
    localStorage.removeItem('last_activity');
    localStorage.removeItem('tester_credentials');
    localStorage.removeItem('third_party_credentials');
    localStorage.removeItem('mp_token');
    localStorage.removeItem('smart_factory_token');
    localStorage.removeItem('factory_token');
    localStorage.removeItem('smart_factory_login_time');
    localStorage.removeItem('factory_data_login_record');
    localStorage.removeItem('auth_token');
    localStorage.removeItem('auth_user_info');
    localStorage.clear();

    // 重置认证状态
    authStore.reset();

    ElMessage.success('已退出系统，请重新登录');

    // 跳转到登录页，并替换历史记录防止浏览器回退到主页
    router.replace('/login');
  } catch (error) {
    // 用户取消退出
  }
};
</script>

<template>
  <header class="app-header">
    <div class="header-left">
      <button class="sidebar-toggle" @click="toggleSidebar">
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <line x1="3" y1="6" x2="21" y2="6" />
          <line x1="3" y1="12" x2="21" y2="12" />
          <line x1="3" y1="18" x2="21" y2="18" />
        </svg>
      </button>
      <div class="logo">
        <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="M12 2L2 7l10 5 10-5-10-5z" />
          <path d="M2 17l10 5 10-5" />
          <path d="M2 12l10 5 10-5" />
        </svg>
      </div>
      <span class="app-title">AI起爆一体化工具集</span>
    </div>

    <div class="header-right">
      <el-tag :type="authStore.authStatus === 'valid' ? 'success' : 'danger'" class="auth-tag">
        {{ authStore.statusText }} · {{ authStore.remainingDays }}天
      </el-tag>

      <div class="search-box">
        <input
          v-model="searchQuery"
          type="text"
          class="search-input"
          placeholder="全局搜索"
          @keyup.enter="handleSearch"
        />
        <button class="search-btn" @click="handleSearch">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="11" cy="11" r="8" />
            <line x1="21" y1="21" x2="16.65" y2="16.65" />
          </svg>
        </button>
      </div>

      <button class="logout-btn" @click="handleLogout">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1-2-2h4" />
          <polyline points="16 17 21 12 16 7" />
          <line x1="21" y1="12" x2="9" y2="12" />
        </svg>
        <span>退出</span>
      </button>
    </div>
  </header>
</template>

<style scoped>
.app-header {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: 60px;
  background: #FFFFFF;
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border-bottom: 1px solid rgba(22, 93, 255, 0.2);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  z-index: 100;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.sidebar-toggle {
  display: none;
  background: none;
  border: 1px solid rgba(22, 93, 255, 0.2);
  border-radius: 8px;
  color: #64748B;
  cursor: pointer;
  padding: 6px;
  transition: all 0.2s;
}

.sidebar-toggle:hover {
  background: rgba(22, 93, 255, 0.1);
  color: #1E293B;
  border-color: rgba(22, 93, 255, 0.4);
}

.logo {
  color: #165DFF;
}

.app-title {
  font-size: 16px;
  font-weight: 600;
  color: #1E293B;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.auth-tag {
  font-size: 12px;
  padding: 4px 12px;
  border-radius: 20px;
}

.auth-status-group {
  display: flex;
  align-items: center;
  gap: 8px;
}

.auth-status-tag {
  font-size: 11px;
  border-radius: 12px;
  transition: all 0.3s;
}

.search-box {
  display: flex;
  align-items: center;
  background: rgba(248, 250, 252, 0.8);
  border: 1px solid rgba(22, 93, 255, 0.2);
  border-radius: 8px;
  padding: 4px 12px;
}

.search-input {
  background: none;
  border: none;
  outline: none;
  color: #64748B;
  font-size: 13px;
  width: 200px;
}

.search-input::placeholder {
  color: #94A3B8;
}

.search-btn {
  background: none;
  border: none;
  color: #94A3B8;
  cursor: pointer;
  padding: 4px;
  transition: color 0.2s;
}

.search-btn:hover {
  color: #64748B;
}

.logout-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background: #fff2f0;
  border: 1px solid #ffccc7;
  border-radius: 8px;
  color: #ff4d4f;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s;
}

.logout-btn:hover {
  background: #ff4d4f;
  color: white;
  border-color: #ff4d4f;
}

@media screen and (max-width: 768px) {
  .app-header {
    height: 56px;
    padding: 0 12px;
  }

  .sidebar-toggle {
    display: block;
  }

  .app-title {
    font-size: 14px;
  }

  .auth-tag {
    font-size: 10px;
    padding: 2px 8px;
  }

  .search-box {
    display: none;
  }

  .logout-btn {
    padding: 6px 12px;
    font-size: 13px;
  }

  .logout-btn span {
    display: none;
  }

  .header-right {
    gap: 10px;
  }
}

@media screen and (max-width: 480px) {
  .app-header {
    padding: 0 8px;
  }

  .app-title {
    font-size: 13px;
    display: none;
  }

  .auth-tag {
    display: none;
  }
}
</style>
