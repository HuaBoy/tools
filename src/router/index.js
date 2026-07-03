import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    redirect: '/home'
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue')
  },
  {
    path: '/home',
    name: 'Home',
    component: () => import('@/views/Home.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/auth/converter',
    name: 'AuthConverter',
    component: () => import('@/views/auth/AuthConverter.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/auth/third-party',
    name: 'ThirdPartyAuth',
    component: () => import('@/views/auth/ThirdPartyAuth.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/log/decrypt',
    name: 'LogDecrypt',
    component: () => import('@/views/log/LogDecrypt.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/log/analysis',
    name: 'LogAnalysis',
    component: () => import('@/views/log/LogAnalysis.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/data/query',
    name: 'DataQuery',
    component: () => import('@/views/data/DataQuery.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/data/trace',
    name: 'DataTrace',
    component: () => import('@/views/data/DataTrace.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/knowledge/base',
    name: 'KnowledgeBase',
    component: () => import('@/views/knowledge/KnowledgeBase.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/knowledge/manual',
    name: 'VersionManual',
    component: () => import('@/views/knowledge/VersionManual.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/knowledge/history',
    name: 'VersionHistory',
    component: () => import('@/views/knowledge/VersionHistory.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/knowledge/audio',
    name: 'AudioRecorder',
    component: () => import('@/views/knowledge/AudioRecorder.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/tools/translate',
    name: 'AiTranslate',
    component: () => import('@/views/tools/AiTranslate.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/tools/qrcode',
    name: 'QrCode',
    component: () => import('@/views/tools/QrCode.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/tools/assistant',
    name: 'AiAssistant',
    component: () => import('@/views/tools/AiAssistant.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/tools/tester',
    name: 'AiTester',
    component: () => import('@/views/tools/AiTester.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/tools/converter',
    name: 'FormatConverter',
    component: () => import('@/views/tools/FormatConverter.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/trace/analysis',
    name: 'TraceAnalysis',
    component: () => import('@/views/trace/TraceAnalysis.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/trace/factory-data',
    name: 'FactoryDataQuery',
    component: () => import('@/views/trace/FactoryDataQuery.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/admin',
    name: 'AdminPanel',
    component: () => import('@/views/admin/AdminPanel.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/admin/users',
    name: 'UserManagement',
    component: () => import('@/views/admin/UserManagement.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/admin/permissions',
    name: 'PermissionManagement',
    component: () => import('@/views/admin/PermissionManagement.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/admin/features',
    name: 'FeatureManagement',
    component: () => import('@/views/admin/FeatureManagement.vue'),
    meta: { requiresAuth: true }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  // 登录页不需要认证
  if (to.path === '/login') {
    // 如果已经登录且凭证有效，直接跳转到首页
    const savedUser = localStorage.getItem('auth_user')
    const savedTime = localStorage.getItem('last_activity')

    if (savedUser && savedTime) {
      const lastTime = parseInt(savedTime, 10)
      const twoDaysInMs = 2 * 24 * 60 * 60 * 1000

      if (Date.now() - lastTime <= twoDaysInMs) {
        next('/home')
        return
      }
    }

    next()
    return
  }
  
  // 需要认证的页面
  if (to.meta.requiresAuth) {
    const savedUser = localStorage.getItem('auth_user')
    const savedTime = localStorage.getItem('last_activity')
    
    if (savedUser && savedTime) {
      const lastTime = parseInt(savedTime, 10)
      const twoDaysInMs = 2 * 24 * 60 * 60 * 1000
      
      if (Date.now() - lastTime <= twoDaysInMs) {
        localStorage.setItem('last_activity', String(Date.now()))
        next()
        return
      }
    }
    
    // 未登录或凭证过期，跳转到登录页
    next('/login')
    return
  }
  
  next()
})

export default router