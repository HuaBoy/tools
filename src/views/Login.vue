<script setup>
import { ref, onMounted } from 'vue';
import { useAuthStore } from '@/stores/auth';
import { useRouter } from 'vue-router';
import { ElMessage } from 'element-plus';
import { isSupabaseReady } from '@/utils/supabase';

const authStore = useAuthStore();
const router = useRouter();

const username = ref('');
const password = ref('');
const isLoading = ref(false);

// 注册相关
const showRegister = ref(false);
const registerUsername = ref('');
const registerEmail = ref('');
const registerPassword = ref('');
const isRegistering = ref(false);

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

  await new Promise(resolve => setTimeout(resolve, 300));

  // 本地模式快捷登录：密码为 123456 时直接通过本地账号登录
  if (password.value.trim() === '123456') {
    try {
      const success = await authStore.login(username.value.trim(), password.value.trim());
      isLoading.value = false;

      if (success) {
        ElMessage.success('登录成功');
        await router.push('/home');
      } else {
        ElMessage.error('登录失败，请重试');
      }
    } catch (error) {
      isLoading.value = false;
      ElMessage.error('登录失败：' + (error?.message || '未知错误'));
    }
    return;
  }

  // 非默认密码：未配置 Supabase 时拒绝
  if (!isSupabaseReady) {
    isLoading.value = false;
    ElMessage.error('当前为本地模式，默认密码为：123456');
    return;
  }

  // 已配置 Supabase：尝试远程登录
  try {
    console.log('[Login] 开始远程登录流程，用户名:', username.value.trim());
    const success = await authStore.login(username.value.trim(), password.value.trim());

    isLoading.value = false;

    if (success) {
      ElMessage.success('登录成功');
      await router.push('/home');
    } else {
      console.warn('[Login] 远程登录返回失败');
      ElMessage.error('账号或密码错误');
    }
  } catch (error) {
    isLoading.value = false;
    console.error('[Login] 远程登录异常:', error);
    const msg = (error?.message || '').includes('Failed to fetch')
      ? '网络请求失败，请检查网络或 Supabase 配置'
      : '登录失败：' + (error?.message || '未知错误');
    ElMessage.error(msg);
  }
};

const handleRegister = async () => {
  if (!registerUsername.value.trim()) {
    ElMessage.warning('请输入用户名');
    return;
  }
  if (!registerEmail.value.trim()) {
    ElMessage.warning('请输入邮箱');
    return;
  }
  if (!registerPassword.value.trim()) {
    ElMessage.warning('请输入密码');
    return;
  }

  isRegistering.value = true;

  try {
    console.log('[Register] 开始注册流程:', { username: registerUsername.value, email: registerEmail.value });
    const success = await authStore.register(registerEmail.value.trim(), registerPassword.value.trim(), registerUsername.value.trim());
    
    if (success) {
      ElMessage.success('注册成功，请登录');
      showRegister.value = false;
      username.value = registerEmail.value;
      registerUsername.value = '';
      registerEmail.value = '';
      registerPassword.value = '';
    } else {
      ElMessage.error('注册失败，请重试');
    }
  } catch (error) {
    console.error('[Register] 注册异常:', error);
    const msg = (error?.message || '').includes('Failed to fetch')
      ? '网络请求失败，请检查网络或 Supabase 配置'
      : '注册失败：' + (error?.message || '未知错误');
    ElMessage.error(msg);
  } finally {
    isRegistering.value = false;
  }
};
</script>

<template>
  <div class="login-container">
    <div class="login-card">
      <div class="login-header">
        <div class="logo">
          <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <path d="M12 2L2 7l10 5 10-5-10-5z" />
            <path d="M2 17l10 5 10-5" />
            <path d="M2 12l10 5 10-5" />
          </svg>
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
            placeholder="请输入密码（默认：123456）"
          />
        </div>
        
        <button type="submit" class="login-btn" :disabled="isLoading">
          <span v-if="isLoading">登录中...</span>
          <span v-else>登录</span>
        </button>
      </form>
      
      <div class="login-hint">
        <p v-if="!isSupabaseReady">当前为本地模式，默认密码：<strong>123456</strong></p>
        <p v-else>默认密码：123456</p>
        <p>账号登录后，2天无操作将自动退出</p>
      </div>
      
      <div v-if="isSupabaseReady" class="register-section">
        <p>还没有账号？<button class="register-btn" @click="showRegister = !showRegister">{{ showRegister ? '收起' : '立即注册' }}</button></p>
        
        <div v-if="showRegister" class="register-form">
          <div class="form-group">
            <label class="form-label">用户名</label>
            <input 
              v-model="registerUsername"
              type="text" 
              class="form-input"
              placeholder="请输入用户名"
            />
          </div>
          <div class="form-group">
            <label class="form-label">邮箱</label>
            <input 
              v-model="registerEmail"
              type="email" 
              class="form-input"
              placeholder="请输入邮箱"
            />
          </div>
          <div class="form-group">
            <label class="form-label">密码</label>
            <input 
              v-model="registerPassword"
              type="password" 
              class="form-input"
              placeholder="请输入密码"
            />
          </div>
          <button type="button" class="register-submit-btn" @click="handleRegister" :disabled="isRegistering">
            <span v-if="isRegistering">注册中...</span>
            <span v-else>注册</span>
          </button>
        </div>
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
  color: #165DFF;
  margin-bottom: 16px;
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

.register-section {
  text-align: center;
  padding-top: 16px;
  border-top: 1px solid rgba(22, 93, 255, 0.1);
  margin-top: 16px;
}

.register-section p {
  font-size: 13px;
  color: #64748B;
  margin: 0;
}

.register-btn {
  background: none;
  border: none;
  color: #165DFF;
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  text-decoration: underline;
  transition: color 0.2s;
  
  &:hover {
    color: #0F4CD0;
  }
}

.register-form {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px dashed rgba(22, 93, 255, 0.15);
}

.register-submit-btn {
  width: 100%;
  padding: 14px;
  background: linear-gradient(135deg, #00B42A 0%, #00A627 100%);
  border: none;
  border-radius: 10px;
  color: #FFFFFF;
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  
  &:hover:not(:disabled) {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(0, 180, 42, 0.3);
  }
  
  &:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }
}
</style>