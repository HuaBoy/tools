<script setup>
import { ref, onMounted } from 'vue';
import { useAuthStore } from '@/stores/auth';
import { useRouter } from 'vue-router';
import { ElMessage } from 'element-plus';

const authStore = useAuthStore();
const router = useRouter();

const username = ref('admin');
const password = ref('800228');
const isLoading = ref(false);

onMounted(async () => {
  await authStore.loadAuthState();
  if (authStore.isLoggedIn) {
    router.push('/home');
  }
});

const handleLogin = async () => {
  if (!username.value.trim()) {
    ElMessage.warning('请输入用户名');
    return;
  }

  if (!password.value.trim()) {
    ElMessage.warning('请输入密码');
    return;
  }

  isLoading.value = true;

  // 请求后端登录接口 /api/login
  try {
    console.log('[Login] 请求登录接口，用户名:', username.value.trim());
    const success = await authStore.login(username.value.trim(), password.value.trim());

    isLoading.value = false;

    if (success) {
      ElMessage.success('登录成功');
      await router.push('/home');
    } else {
      ElMessage.error('账号或密码错误');
    }
  } catch (error) {
    isLoading.value = false;
    console.error('[Login] 登录异常:', error);
    const msg = (error?.message || '').includes('Failed to fetch')
      ? '无法连接服务器，请检查网络或后端服务是否正常'
      : (error?.message || '登录失败，请重试');
    ElMessage.error(msg);
  }
};
</script>

<template>
  <div class="login-container">
    <div class="login-card">
      <div class="login-header">
        <div class="logo">
          <img src="/logo.png" alt="Holyview 盛景" class="logo-img" />
        </div>
        <h1 class="title">AI起爆一体化工具集</h1>
        <p class="subtitle">请登录系统</p>
      </div>
      
      <form class="login-form" @submit.prevent="handleLogin">
        <div class="form-group">
          <label class="form-label">用户名</label>
          <input 
            v-model="username"
            type="text" 
            class="form-input"
            placeholder="请输入用户名"
            autofocus
          />
        </div>
        
        <div class="form-group">
          <label class="form-label">密码</label>
          <input 
            v-model="password"
            type="password" 
            class="form-input"
            placeholder="请输入密码"
          />
        </div>
        
        <button type="submit" class="login-btn" :disabled="isLoading">
          <span v-if="isLoading">登录中...</span>
          <span v-else>登录</span>
        </button>
      </form>
      
      <div class="login-hint">
        <p>默认管理员：<strong>admin / 800228</strong></p>
        <p>账号登录后，2天无操作将自动退出</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #F8FAFC 0%, #E2E8F0 100%);
  padding: 20px;
}

.login-card {
  width: 100%;
  max-width: 420px;
  background: #FFFFFF;
  border-radius: 16px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.1);
  padding: 40px;
}

.login-header {
  text-align: center;
  margin-bottom: 32px;
}

.logo {
  margin-bottom: 16px;
}

.logo-img {
  height: 80px;
  width: auto;
  object-fit: contain;
}

.title {
  font-size: 24px;
  font-weight: 700;
  color: #1E293B;
  margin: 0 0 8px 0;
}

.subtitle {
  font-size: 14px;
  color: #64748B;
  margin: 0;
}

.login-form {
  margin-bottom: 24px;
}

.form-group {
  margin-bottom: 20px;
}

.form-label {
  display: block;
  font-size: 13px;
  color: #64748B;
  margin-bottom: 8px;
}

.form-input {
  width: 100%;
  padding: 14px 16px;
  border: 1px solid rgba(22, 93, 255, 0.2);
  border-radius: 10px;
  font-size: 14px;
  color: #1E293B;
  outline: none;
  transition: border-color 0.2s;
  box-sizing: border-box;
  
  &:focus {
    border-color: rgba(22, 93, 255, 0.6);
  }
  
  &::placeholder {
    color: #94A3B8;
  }
}

.login-btn {
  width: 100%;
  padding: 14px;
  background: linear-gradient(135deg, #165DFF 0%, #0F4CD0 100%);
  border: none;
  border-radius: 10px;
  color: #FFFFFF;
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  
  &:hover:not(:disabled) {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(22, 93, 255, 0.3);
  }
  
  &:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }
}

.login-hint {
  text-align: center;
  padding-top: 16px;
  border-top: 1px solid rgba(22, 93, 255, 0.1);
}

.login-hint p {
  font-size: 12px;
  color: #94A3B8;
  margin: 4px 0;
}


</style>