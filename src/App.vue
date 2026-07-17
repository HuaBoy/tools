<script setup>
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import Header from '@/components/layout/Header.vue'
import Sidebar from '@/components/layout/Sidebar.vue'
import PlatformLoginDialog from '@/components/PlatformLoginDialog.vue'
import { useAuthStore } from '@/stores/auth'
import { initLoginStatus } from '@/utils/loginStatus.js'
import { performPlatformLogin } from '@/utils/platformAutoLogin.js'

const route = useRoute()
const isLoaded = ref(false)
const sidebarVisible = ref(true)
const isMobile = ref(false)
const authStore = useAuthStore()

const isLoginPage = computed(() => route.path === '/login')

const handleToggleSidebar = () => {
  sidebarVisible.value = !sidebarVisible.value
}

const handleCloseSidebar = () => {
  sidebarVisible.value = false
}

const checkMobile = () => {
  isMobile.value = window.innerWidth <= 768
  if (isMobile.value) {
    sidebarVisible.value = false
  }
}

// 全局自动登录定时器
let autoLoginTimer = null
const AUTO_LOGIN_CHECK_INTERVAL = 5 * 60 * 1000 // 每 5 分钟检查一次是否需要自动登录

const startAutoLoginTimer = () => {
  if (autoLoginTimer) return
  autoLoginTimer = setInterval(() => {
    performPlatformLogin()
  }, AUTO_LOGIN_CHECK_INTERVAL)
  // 启动后立即执行一次检查
  setTimeout(() => {
    performPlatformLogin()
  }, 3000)
}

const stopAutoLoginTimer = () => {
  if (autoLoginTimer) {
    clearInterval(autoLoginTimer)
    autoLoginTimer = null
  }
}

onMounted(async () => {
  checkMobile()
  window.addEventListener('resize', checkMobile)
  // 初始化第三方平台登录状态
  initLoginStatus()
  // 应用启动时加载认证状态（兼容未配置 Supabase 的本地模式）
  try {
    await authStore.loadAuthState()
  } catch (e) {
    console.warn('加载认证状态失败:', e)
  }
  // 启动全局自动登录定时器
  startAutoLoginTimer()
  setTimeout(() => {
    isLoaded.value = true
  }, 100)
})

onUnmounted(() => {
  window.removeEventListener('resize', checkMobile)
  stopAutoLoginTimer()
})
</script>

<template>
  <div class="app-container grid-bg" :class="{ 'app-loaded': isLoaded }">
    <template v-if="isLoginPage">
      <router-view />
    </template>
    
    <template v-else>
      <Header @toggle-sidebar="handleToggleSidebar" />
      <div class="main-content">
        <Sidebar
          :visible="sidebarVisible"
          :is-mobile="isMobile"
          @close="handleCloseSidebar"
        />
        <div class="content-area" :class="{ 'content-full': isMobile && !sidebarVisible }">
          <router-view v-slot="{ Component }">
            <component :is="Component" />
          </router-view>
        </div>
      </div>
    </template>

    <!-- 全局平台登录弹窗 -->
    <PlatformLoginDialog />
  </div>
</template>

<style scoped>
.app-container {
  min-height: 100vh;
}

.main-content {
  display: flex;
  padding-top: 60px;
}

.content-area {
  flex: 1;
  padding: 20px;
  padding-left: 240px;
  min-height: calc(100vh - 60px);
  transition: padding-left 0.3s ease;
  overflow-x: auto;
}

.content-full {
  padding-left: 20px;
}

@media screen and (max-width: 768px) {
  .content-area {
    padding-left: 20px;
    padding-top: 16px;
    padding-bottom: 16px;
    min-height: calc(100vh - 56px);
  }

  .main-content {
    padding-top: 56px;
  }
}

@media screen and (max-width: 480px) {
  .content-area {
    padding: 12px;
  }
}
</style>
