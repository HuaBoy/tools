import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

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
          path: '/knowledge/rag',
          name: 'KnowledgeBaseRAG',
          component: () => import('@/views/knowledge/KnowledgeBaseRAG.vue'),
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
      path: '/knowledge/production-history',
      name: 'ProductionHistory',
      component: () => import('@/views/knowledge/ProductionHistory.vue'),
      meta: { requiresAuth: true }
    },
    // ===== 销售岗（民爆行业洞察） =====
    {
      path: '/sales',
      name: 'SalesInsight',
      component: () => import('@/views/sales/SalesWorkbench.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/sales/:section',
      name: 'SalesInsightSection',
      component: () => import('@/views/sales/SalesWorkbench.vue'),
      meta: { requiresAuth: true }
    },
  {
    path: '/knowledge/audio',
    name: 'AudioRecorder',
    component: () => import('@/views/knowledge/AudioRecorder.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/knowledge/base',
    name: 'KnowledgeBase',
    component: () => import('@/views/knowledge/KnowledgeBase.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/knowledge/ai-chat',
    name: 'KnowledgeAiChat',
    component: () => import('@/views/knowledge/AiChat.vue'),
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
    path: '/tools/remote-phone',
    name: 'RemotePhoneControl',
    component: () => import('@/views/tools/RemotePhoneControl.vue'),
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
  },
  {
    path: '/admin/database',
    name: 'DatabaseManagement',
    component: () => import('@/views/admin/DatabaseManagement.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/stats/big-screen',
    name: 'BigScreen',
    component: () => import('@/views/stats/BigScreen.vue'),
    meta: { requiresAuth: true }
  },
  // ===== 盛景应用 =====
  {
    path: '/appstore/shengjing',
    name: 'ShengjingApp',
    component: () => import('@/views/appstore/ShengjingApp.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/appstore/push',
    name: 'PushManagement',
    component: () => import('@/views/appstore/PushManagement.vue'),
    meta: { requiresAuth: true }
  },
  // ===== 硬件管理（预留） =====
  {
    path: '/hardware/devices',
    name: 'HardwareDevices',
    component: () => import('@/modules/hardware/pages/DeviceManager.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/hardware/chips',
    name: 'HardwareChips',
    component: () => import('@/modules/hardware/pages/ChipManager.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/hardware/firmware',
    name: 'HardwareFirmware',
    component: () => import('@/modules/hardware/pages/FirmwareOTA.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/ai/pcb',
    name: 'AiPcb',
    component: () => import('@/views/ai/AiPcb.vue'),
    meta: { requiresAuth: true }
  },
  // ===== 海外发货管理 =====
  {
    path: '/overseas/shipping',
    name: 'OverseasShipping',
    component: () => import('@/views/overseas/OverseasShipping.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/overseas/manual',
    name: 'OperationManual',
    component: () => import('@/views/overseas/OperationManual.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/overseas/video',
    name: 'OperationVideo',
    component: () => import('@/views/overseas/OperationVideo.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/overseas/blasting-design',
    name: 'BlastingDesign',
    component: () => import('@/views/overseas/BlastingDesign.vue'),
    meta: { requiresAuth: true }
  },
  // ===== AI 研发流水线（REQ-2026-001） =====
  {
    path: '/pipeline/requirements',
    name: 'PipelineRequirementPool',
    component: () => import('@/views/pipeline/RequirementPool.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/pipeline/board',
    name: 'PipelineBoard',
    component: () => import('@/views/pipeline/PipelineBoard.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/pipeline/review',
    name: 'PipelineReview',
    component: () => import('@/views/pipeline/ReviewCenter.vue'),
    meta: { requiresAuth: true }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 预加载路由组件（hover 菜单时提前加载，消除点击延迟）
const preloadCache = {}
export function preloadRoute(path) {
  if (preloadCache[path]) return
  const matched = router.resolve(path)
  const route = matched.matched[0]
  if (route?.components?.default) {
    preloadCache[path] = true
    if (typeof route.components.default === 'function') {
      route.components.default()
    }
  }
}

// 是否已登录且凭证在有效期内（2 天）
function isLoggedIn() {
  const token = localStorage.getItem('auth_token')
  const user = localStorage.getItem('auth_user')
  const time = localStorage.getItem('last_activity')
  if (token && user && time) {
    const lastTime = parseInt(time, 10)
    const twoDaysInMs = 2 * 24 * 60 * 60 * 1000
    return Date.now() - lastTime <= twoDaysInMs
  }
  return false
}

router.beforeEach((to, from, next) => {
  // 登录页不需要认证
  if (to.path === '/login') {
    // 已登录则直接进入主页，避免重复登录
    if (isLoggedIn()) {
      next('/home')
    } else {
      next()
    }
    return
  }

  // 需要认证的页面
  if (to.meta.requiresAuth) {
    if (!isLoggedIn()) {
      // 凭证缺失或已过期：清理残留状态，强制跳转到登录页
      try {
        useAuthStore().reset()
      } catch (e) {
        // pinia 未激活时忽略，localStorage 已可拦截
      }
      next('/login')
      return
    }
    // 刷新最近活跃时间
    localStorage.setItem('last_activity', String(Date.now()))
    next()
    return
  }

  next()
})

export default router