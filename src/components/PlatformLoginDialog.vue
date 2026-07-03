<script setup>
import { computed, onMounted, onUnmounted, watch, ref, nextTick } from 'vue';
import { ElMessage } from 'element-plus';
import {
  loginDialogState,
  handleDialogLogin,
  closeLoginDialog,
  goToThirdPartyAuth,
  PLATFORMS
} from '@/utils/platformLogin.js';
import { getLatestCredentials } from '@/utils/platformLogin.js';

const platformName = computed(() => {
  return PLATFORMS[loginDialogState.platform]?.name || '平台';
});

const platformDomain = computed(() => {
  return PLATFORMS[loginDialogState.platform]?.domain || '';
});

const hasCredentials = computed(() => {
  return !!(loginDialogState.username && loginDialogState.password);
});

const handleSubmit = () => {
  handleDialogLogin();
};

const handleClose = () => {
  closeLoginDialog();
  if (loginDialogState._resolve) {
    loginDialogState._resolve({ success: false, reason: 'cancelled' });
    loginDialogState._resolve = null;
  }
};

const handleGoToAuth = () => {
  goToThirdPartyAuth(loginDialogState.platform);
  if (loginDialogState._resolve) {
    loginDialogState._resolve({ success: false, reason: 'no_credentials' });
    loginDialogState._resolve = null;
  }
};

const handleUseLatest = () => {
  const latest = getLatestCredentials(loginDialogState.platform);
  if (latest) {
    loginDialogState.username = latest.username;
    loginDialogState.password = latest.password;
    ElMessage.info('已填充最近登录的账号');
  }
};

// 处理回车键提交
const handleKeydown = (e) => {
  if (loginDialogState.visible && e.key === 'Enter' && !e.shiftKey) {
    if (e.target.tagName !== 'TEXTAREA') {
      e.preventDefault();
      handleSubmit();
    }
  }
};

onMounted(() => {
  document.addEventListener('keydown', handleKeydown);
});

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeydown);
});

// 监听 visible 变化
watch(() => loginDialogState.visible, (newVal) => {
  if (newVal) {
    nextTick(() => {
      // 自动聚焦密码框（如果已填充账号）或用户名框
      const passwordInput = document.querySelector('.platform-login-dialog .password-input');
      if (loginDialogState.password && passwordInput) {
        passwordInput.focus();
      } else {
        const usernameInput = document.querySelector('.platform-login-dialog .username-input');
        if (usernameInput) usernameInput.focus();
      }
    });
  }
});
</script>

<template>
  <Teleport to="body">
    <Transition name="dialog-fade">
      <div v-if="loginDialogState.visible" class="platform-login-overlay" @click.self="handleClose">
        <div class="platform-login-dialog">
          <div class="dialog-header">
            <div class="header-icon">
              <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <rect x="3" y="11" width="18" height="11" rx="2" ry="2"/>
                <path d="M7 11V7a5 5 0 0 1 10 0v4"/>
              </svg>
            </div>
            <div class="header-content">
              <h3>需要登录 {{ platformName }}</h3>
              <p class="header-tip">调用 {{ platformName }} 接口需要先登录</p>
            </div>
            <button class="close-btn" @click="handleClose" title="关闭">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <line x1="18" y1="6" x2="6" y2="18" />
                <line x1="6" y1="6" x2="18" y2="18" />
              </svg>
            </button>
          </div>

          <div class="dialog-body">
            <div v-if="hasCredentials" class="info-banner">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <circle cx="12" cy="12" r="10"/>
                <line x1="12" y1="16" x2="12" y2="12"/>
                <line x1="12" y1="8" x2="12.01" y2="8"/>
              </svg>
              <span>已为您预填最近一次登录的账号</span>
            </div>

            <div class="form-group">
              <label>用户名</label>
              <input
                v-model="loginDialogState.username"
                type="text"
                class="form-input username-input"
                placeholder="请输入用户名"
                :disabled="loginDialogState.loading"
              />
            </div>

            <div class="form-group">
              <label>密码</label>
              <div class="password-wrapper">
                <input
                  v-model="loginDialogState.password"
                  :type="loginDialogState.showPassword ? 'text' : 'password'"
                  class="form-input password-input"
                  placeholder="请输入密码"
                  :disabled="loginDialogState.loading"
                />
                <button
                  type="button"
                  class="toggle-password"
                  @click="loginDialogState.showPassword = !loginDialogState.showPassword"
                  :title="loginDialogState.showPassword ? '隐藏密码' : '显示密码'"
                >
                  <svg v-if="loginDialogState.showPassword" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/>
                    <circle cx="12" cy="12" r="3"/>
                  </svg>
                  <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"/>
                    <line x1="1" y1="1" x2="23" y2="23"/>
                  </svg>
                </button>
              </div>
            </div>

            <div v-if="loginDialogState.errorMessage" class="error-message">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <circle cx="12" cy="12" r="10"/>
                <line x1="12" y1="8" x2="12" y2="12"/>
                <line x1="12" y1="16" x2="12.01" y2="16"/>
              </svg>
              <span>{{ loginDialogState.errorMessage }}</span>
            </div>

            <div class="form-tips">
              <span class="domain-tip">平台地址：{{ platformDomain }}</span>
            </div>
          </div>

          <div class="dialog-footer">
            <button
              type="button"
              class="btn-link"
              @click="handleGoToAuth"
              :disabled="loginDialogState.loading"
            >
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M9 18l6-6-6-6"/>
              </svg>
              <span>去授权页面</span>
            </button>
            <div class="footer-actions">
              <button
                type="button"
                class="btn-secondary"
                @click="handleClose"
                :disabled="loginDialogState.loading"
              >
                取消
              </button>
              <button
                v-if="!hasCredentials"
                type="button"
                class="btn-primary"
                @click="handleUseLatest"
                :disabled="loginDialogState.loading"
              >
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M21 12a9 9 0 11-9-9c2.52 0 4.93 1 6.74 2.74L21 8"/>
                  <path d="M21 3v5h-5"/>
                </svg>
                <span>使用最近账号</span>
              </button>
              <button
                type="button"
                class="btn-primary"
                @click="handleSubmit"
                :disabled="loginDialogState.loading || !loginDialogState.username || !loginDialogState.password"
              >
                <svg v-if="loginDialogState.loading" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="spin">
                  <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
                </svg>
                <span>{{ loginDialogState.loading ? '登录中...' : '立即登录' }}</span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.platform-login-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  padding: 20px;
}

.platform-login-dialog {
  background: #FFFFFF;
  border-radius: 16px;
  width: 100%;
  max-width: 460px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.2);
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.dialog-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 20px 24px;
  border-bottom: 1px solid #F1F5F9;
  background: linear-gradient(135deg, rgba(22, 93, 255, 0.03) 0%, rgba(15, 76, 208, 0.06) 100%);
}

.header-icon {
  width: 44px;
  height: 44px;
  border-radius: 10px;
  background: linear-gradient(135deg, #165DFF 0%, #0F4CD0 100%);
  color: #FFFFFF;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.header-content {
  flex: 1;
  min-width: 0;
}

.header-content h3 {
  font-size: 16px;
  font-weight: 600;
  color: #1E293B;
  margin: 0 0 2px;
}

.header-tip {
  font-size: 12px;
  color: #64748B;
  margin: 0;
}

.close-btn {
  width: 32px;
  height: 32px;
  border: none;
  background: transparent;
  border-radius: 6px;
  cursor: pointer;
  color: #94A3B8;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.close-btn:hover {
  background: #F1F5F9;
  color: #1E293B;
}

.dialog-body {
  padding: 20px 24px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.info-banner {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  background: rgba(22, 93, 255, 0.05);
  border: 1px solid rgba(22, 93, 255, 0.15);
  border-radius: 8px;
  font-size: 12px;
  color: #165DFF;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-group label {
  font-size: 13px;
  font-weight: 500;
  color: #475569;
}

.form-input {
  width: 100%;
  padding: 10px 14px;
  border: 1px solid #E2E8F0;
  border-radius: 8px;
  font-size: 14px;
  color: #1E293B;
  background: #FFFFFF;
  transition: all 0.2s;
  outline: none;
}

.form-input:focus {
  border-color: #165DFF;
  box-shadow: 0 0 0 3px rgba(22, 93, 255, 0.1);
}

.form-input:disabled {
  background: #F8FAFC;
  cursor: not-allowed;
}

.form-input::placeholder {
  color: #94A3B8;
}

.password-wrapper {
  position: relative;
}

.password-wrapper .form-input {
  padding-right: 40px;
}

.toggle-password {
  position: absolute;
  right: 8px;
  top: 50%;
  transform: translateY(-50%);
  width: 28px;
  height: 28px;
  border: none;
  background: transparent;
  border-radius: 4px;
  cursor: pointer;
  color: #94A3B8;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.toggle-password:hover {
  background: #F1F5F9;
  color: #475569;
}

.error-message {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 12px;
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid rgba(239, 68, 68, 0.2);
  border-radius: 6px;
  font-size: 12px;
  color: #EF4444;
}

.form-tips {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: #94A3B8;
}

.domain-tip {
  font-family: 'SF Mono', 'Monaco', 'Cascadia Code', 'Courier New', monospace;
}

.dialog-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 16px 24px;
  background: #FAFAFA;
  border-top: 1px solid #F1F5F9;
}

.btn-link {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 6px 8px;
  background: transparent;
  border: none;
  border-radius: 4px;
  font-size: 12px;
  color: #165DFF;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-link:hover:not(:disabled) {
  background: rgba(22, 93, 255, 0.05);
}

.btn-link:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.footer-actions {
  display: flex;
  gap: 8px;
}

.btn-secondary {
  padding: 8px 16px;
  background: #FFFFFF;
  border: 1px solid #E2E8F0;
  border-radius: 6px;
  font-size: 13px;
  color: #475569;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-secondary:hover:not(:disabled) {
  background: #F8FAFC;
  border-color: #CBD5E1;
}

.btn-secondary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-primary {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 18px;
  background: linear-gradient(135deg, #165DFF 0%, #0F4CD0 100%);
  border: none;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
  color: #FFFFFF;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(22, 93, 255, 0.3);
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

/* 过渡动画 */
.dialog-fade-enter-active,
.dialog-fade-leave-active {
  transition: opacity 0.2s ease;
}

.dialog-fade-enter-active .platform-login-dialog,
.dialog-fade-leave-active .platform-login-dialog {
  transition: transform 0.2s ease;
}

.dialog-fade-enter-from,
.dialog-fade-leave-to {
  opacity: 0;
}

.dialog-fade-enter-from .platform-login-dialog,
.dialog-fade-leave-to .platform-login-dialog {
  transform: scale(0.95) translateY(-10px);
}

/* 响应式 */
@media (max-width: 480px) {
  .platform-login-overlay {
    padding: 12px;
  }

  .dialog-header {
    padding: 16px 18px;
  }

  .dialog-body {
    padding: 16px 18px;
  }

  .dialog-footer {
    padding: 12px 18px;
    flex-wrap: wrap;
  }

  .footer-actions {
    width: 100%;
    justify-content: stretch;
  }

  .footer-actions button {
    flex: 1;
    justify-content: center;
  }
}
</style>
