<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { preloadRoute } from '@/router'

const props = defineProps({ visible: { type: Boolean, default: true } })
const emit = defineEmits(['close'])
const router = useRouter()
const route = useRoute()
const isMobile = ref(false)

// 场景化菜单分组
const menuGroups = [
  {
    name: '智能诊断',
    icon: 'search',
    children: [
      { name: 'AI 诊断工作台', path: '/home', icon: 'bot', home: true },
      { name: '起爆器日志AI分析', path: '/log/analysis', icon: 'brain' },
      { name: '日志解密工具', path: '/log/decrypt', icon: 'lock' },
      { name: '全链路追溯', path: '/data/trace', icon: 'git-branch' }
    ]
  },
  {
    name: '自动报告',
    icon: 'file-text',
    children: [
      { name: '批次数据追溯', path: '/trace/analysis', icon: 'bar-chart-2' },
      { name: 'AI起爆数据查询', path: '/data/query', icon: 'search' },
      { name: '数据大屏', path: '/stats/big-screen', icon: 'monitor' }
    ]
  },
  {
    name: '销售岗',
    icon: 'trending-up',
    children: [
      { name: '行业洞察', path: '/sales', icon: 'trending-up' },
      { name: '全国总量', path: '/sales', query: 'national', icon: 'bar-chart-2' },
      { name: '区域分布', path: '/sales', query: 'regional', icon: 'map-pin' },
      { name: 'CR20集中度', path: '/sales', query: 'cr20', icon: 'pie-chart' },
      { name: '行业动态', path: '/sales', query: 'news', icon: 'newspaper' }
    ]
  },
  {
    name: '自动化任务',
    icon: 'zap',
    children: [
      { name: '授权码转换工具', path: '/auth/converter', icon: 'key' },
      { name: '三方账号授权', path: '/auth/third-party', icon: 'shield' },
      { name: '云系统', path: '/tools/tester', icon: 'terminal' }
    ]
  },
  {
    name: '硬件管理',
    icon: 'cpu',
    children: [
      { name: '设备管理', path: '/hardware/devices', icon: 'smartphone', coming: true },
      { name: '芯片管理', path: '/hardware/chips', icon: 'cpu-chip', coming: true },
      { name: '固件升级', path: '/hardware/firmware', icon: 'upload-cloud', coming: true }
    ]
  },
  {
    name: 'AI知识库',
    icon: 'book',
    children: [
      { name: '智能知识库(RAG)', path: '/knowledge/rag', icon: 'brain', highlight: true },
      { name: '起爆器版本手册', path: '/knowledge/manual', icon: 'book-open' },
      { name: '起爆器版本履历', path: '/knowledge/history', icon: 'history' },
      { name: '产线履历', path: '/knowledge/production-history', icon: 'factory' },
      { name: '录音管理', path: '/knowledge/audio', icon: 'mic' },
      { name: '智能制造系统', path: '/trace/factory-data', icon: 'database' }
    ]
  },
  {
    name: '通用工具',
    icon: 'wrench',
    children: [
      { name: 'AI翻译工具', path: '/tools/translate', icon: 'globe' },
      { name: '文档格式转换', path: '/tools/converter', icon: 'file-text' },
      { name: '二维码生成工具', path: '/tools/qrcode', icon: 'qr-code' },
      { name: '远程手机控制', path: '/tools/remote-phone', icon: 'smartphone' }
    ]
  },
  {
    name: '盛景应用',
    icon: 'package',
    children: [
      { name: '应用管理', path: '/appstore/shengjing', icon: 'box' },
      { name: '推送管理', path: '/appstore/push', icon: 'send' }
    ]
  },
  {
    name: '系统管理',
    icon: 'settings',
    children: [
      { name: '后台管理', path: '/admin', icon: 'server' },
      { name: '用户管理', path: '/admin/users', icon: 'users' },
      { name: '权限管理', path: '/admin/permissions', icon: 'key' },
      { name: '功能管理', path: '/admin/features', icon: 'toggle' },
      { name: '服务器数据库', path: '/admin/database', icon: 'database' }
    ]
  },
  {
    name: 'AI-PCB 设计',
    icon: 'cpu',
    children: [
      { name: 'AI-PCB 助手', path: '/ai/pcb', icon: 'wrench', highlight: true }
    ]
  },
  {
    name: '海外业务',
    icon: 'package',
    children: [
      { name: '海外发货管理', path: '/overseas/shipping', icon: 'send' },
      { name: '操作手册', path: '/overseas/manual', icon: 'book-open' },
      { name: '操作视频', path: '/overseas/video', icon: 'video' },
      { name: '爆破设计软件', path: '/overseas/blasting-design', icon: 'zap' }
    ]
  }
]

const expandedGroups = ref(menuGroups.map((_, i) => i))

function toggleGroup(index) {
  const idx = expandedGroups.value.indexOf(index)
  if (idx > -1) expandedGroups.value.splice(idx, 1)
  else expandedGroups.value.push(index)
}
function isActive(target) {
  const path = typeof target === 'string' ? target : target.path
  if (route.path !== path) return false
  if (typeof target === 'object' && target.query) return route.query.tab === target.query
  return !route.query.tab
}
function navigateTo(target, coming) {
  if (typeof target === 'string') {
    if (coming) return
    router.push(target)
    if (isMobile.value) emit('close')
    return
  }
  if (target.coming) return
  router.push(target.query ? { path: target.path, query: { tab: target.query } } : { path: target.path })
  if (isMobile.value) emit('close')
}

function checkMobile() { isMobile.value = window.innerWidth <= 768 }
onMounted(() => { checkMobile(); window.addEventListener('resize', checkMobile) })
onUnmounted(() => { window.removeEventListener('resize', checkMobile) })
</script>

<template>
  <aside class="sidebar" :class="{ 'sidebar-mobile': isMobile, 'sidebar-hidden': isMobile && !visible }">
    <div class="sidebar-overlay" v-if="isMobile && visible" @click="emit('close')"></div>
    <div class="sidebar-content">
      <div class="sidebar-header" v-if="isMobile">
        <span class="sidebar-title">功能菜单</span>
        <button class="sidebar-close" @click="emit('close')">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
        </button>
      </div>

      <nav class="sidebar-nav">
        <!-- AI 工作台首页入口 -->
        <div class="home-entrance" :class="{ active: isActive('/home') }" @click="navigateTo('/home')">
          <div class="home-pulse"></div>
          <div class="home-info">
            <span class="home-label">AI 执行工作台</span>
            <span class="home-desc">对话式运维入口</span>
          </div>
          <span class="home-arrow">→</span>
        </div>

        <div class="sidebar-divider"></div>

        <div v-for="(group, groupIndex) in menuGroups" :key="groupIndex" class="menu-group">
          <div class="group-header" @click="toggleGroup(groupIndex)">
            <div class="group-icon">
              <svg v-if="group.icon === 'search'" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
              <svg v-else-if="group.icon === 'file-text'" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
              <svg v-else-if="group.icon === 'zap'" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/></svg>
              <svg v-else-if="group.icon === 'cpu'" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="4" y="4" width="16" height="16" rx="2" ry="2"/><rect x="9" y="9" width="6" height="6"/><line x1="9" y1="1" x2="9" y2="4"/><line x1="15" y1="1" x2="15" y2="4"/><line x1="9" y1="20" x2="9" y2="23"/><line x1="15" y1="20" x2="15" y2="23"/><line x1="20" y1="9" x2="23" y2="9"/><line x1="20" y1="14" x2="23" y2="14"/><line x1="1" y1="9" x2="4" y2="9"/><line x1="1" y1="14" x2="4" y2="14"/></svg>
              <svg v-else-if="group.icon === 'book'" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z"/><path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z"/></svg>
              <svg v-else-if="group.icon === 'package'" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/><polyline points="3.27 6.96 12 12.01 20.73 6.96"/><line x1="12" y1="22.08" x2="12" y2="12"/></svg>
              <svg v-else-if="group.icon === 'settings'" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09a1.65 1.65 0 0 0-1-1.51 1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09a1.65 1.65 0 0 0 1.51-1 1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>
              <svg v-else-if="group.icon === 'trending-up'" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="23 6 13.5 15.5 8.5 10.5 1 18"/><polyline points="17 6 23 6 23 12"/></svg>
              <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z"/></svg>
            </div>
            <span class="group-name">{{ group.name }}</span>
            <svg class="group-arrow" :class="{ rotated: expandedGroups.includes(groupIndex) }" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="6 9 12 15 18 9"/></svg>
          </div>
          <div v-show="expandedGroups.includes(groupIndex)" class="group-menu">
            <div
              v-for="item in group.children" :key="item.path"
              class="menu-item"
              :class="{ active: isActive(item), 'coming-soon': item.coming }"
              @click="navigateTo(item)"
              @mouseenter="preloadRoute(item.path)"
            >
              <div class="item-icon">
                <svg v-if="item.icon === 'bot'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>
                <svg v-else-if="item.icon === 'brain'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 4.5a2.5 2.5 0 0 0-4.96-.46 2.5 2.5 0 0 0-1.2 3.08 2.5 2.5 0 0 0 1.25 4.42 2.5 2.5 0 0 0 3.71 3.71 2.5 2.5 0 0 0 4.42-1.25 2.5 2.5 0 0 0 3.08-1.2A2.5 2.5 0 0 0 19.5 12a2.5 2.5 0 0 0-4.96-.46 2.5 2.5 0 0 0-1.2 3.08 2.5 2.5 0 0 0 1.25 4.42 2.5 2.5 0 0 0 3.71 3.71 2.5 2.5 0 0 0 4.42-1.25 2.5 2.5 0 0 0 3.08-1.2A2.5 2.5 0 0 0 22 12a2.5 2.5 0 0 0-4.96-.46"/></svg>
                <svg v-else-if="item.icon === 'lock'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>
                <svg v-else-if="item.icon === 'git-branch'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="6" y1="3" x2="6" y2="21"/><circle cx="18" cy="6" r="3"/><circle cx="6" cy="18" r="3"/><path d="M18 9a9 9 0 0 1-9 9"/></svg>
                <svg v-else-if="item.icon === 'smartphone'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="5" y="2" width="14" height="20" rx="2" ry="2"/><line x1="12" y1="18" x2="12.01" y2="18"/></svg>
                <svg v-else-if="item.icon === 'cpu-chip'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="4" y="4" width="16" height="16" rx="2" ry="2"/><rect x="9" y="9" width="6" height="6"/></svg>
                <svg v-else-if="item.icon === 'box'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/><polyline points="3.27 6.96 12 12.01 20.73 6.96"/></svg>
                <svg v-else-if="item.icon === 'send'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="22" y1="2" x2="11" y2="13"/><polygon points="22 2 15 22 11 13 2 9 22 2"/></svg>
                <svg v-else-if="item.icon === 'upload-cloud'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="16 16 12 12 8 16"/><line x1="12" y1="12" x2="12" y2="21"/><path d="M20.39 18.39A5 5 0 0 0 18 9h-1.26A8 8 0 1 0 3 16.3"/><polyline points="16 16 12 12 8 16"/></svg>
                <svg v-else-if="item.icon === 'database'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><ellipse cx="12" cy="5" rx="9" ry="3"/><path d="M3 5v14a9 3 0 0 0 18 0V5"/><path d="M3 12a9 3 0 0 0 18 0"/></svg>
                <svg v-else-if="item.icon === 'factory'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M2 20a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2V8l-7 5V8l-7 5V4a2 2 0 0 0-2-2H4a2 2 0 0 0-2 2Z"/><path d="M7 18h1"/><path d="M12 18h1"/><path d="M17 18h1"/></svg>
                <svg v-else-if="item.icon === 'trending-up'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="23 6 13.5 15.5 8.5 10.5 1 18"/><polyline points="17 6 23 6 23 12"/></svg>
                <svg v-else-if="item.icon === 'bar-chart-2'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="20" x2="18" y2="10"/><line x1="12" y1="20" x2="12" y2="4"/><line x1="6" y1="20" x2="6" y2="14"/></svg>
                <svg v-else-if="item.icon === 'pie-chart'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21.21 15.89A10 10 0 1 1 8 2.83"/><path d="M22 12A10 10 0 0 0 12 2v10z"/></svg>
                <svg v-else-if="item.icon === 'map-pin'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"/><circle cx="12" cy="10" r="3"/></svg>
                <svg v-else-if="item.icon === 'newspaper'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M4 22h16a2 2 0 0 0 2-2V4a2 2 0 0 0-2-2H8a2 2 0 0 0-2 2v16a2 2 0 0 1-2 2zm0 0a2 2 0 0 1-2-2v-9c0-1.1.9-2 2-2h2"/><line x1="8" y1="6" x2="16" y2="6"/><line x1="8" y1="10" x2="16" y2="10"/><line x1="8" y1="14" x2="12" y2="14"/></svg>
                <svg v-else-if="item.icon === 'trending-up'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="23 6 13.5 15.5 8.5 10.5 1 18"/><polyline points="17 6 23 6 23 12"/></svg>
                <svg v-else-if="item.icon === 'bar-chart-2'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="20" x2="18" y2="10"/><line x1="12" y1="20" x2="12" y2="4"/><line x1="6" y1="20" x2="6" y2="14"/></svg>
                <svg v-else-if="item.icon === 'pie-chart'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21.21 15.89A10 10 0 1 1 8 2.83"/><path d="M22 12A10 10 0 0 0 12 2v10z"/></svg>
                <svg v-else-if="item.icon === 'map-pin'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"/><circle cx="12" cy="10" r="3"/></svg>
                <svg v-else-if="item.icon === 'newspaper'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M4 22h16a2 2 0 0 0 2-2V4a2 2 0 0 0-2-2H8a2 2 0 0 0-2 2v16a2 2 0 0 1-2 2zm0 0a2 2 0 0 1-2-2v-9c0-1.1.9-2 2-2h2"/><line x1="8" y1="6" x2="16" y2="6"/><line x1="8" y1="10" x2="16" y2="10"/><line x1="8" y1="14" x2="12" y2="14"/></svg>
                <svg v-else width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"/><path d="M3 9h18"/></svg>
              </div>
              <span class="item-name">{{ item.name }}</span>
              <span v-if="item.coming" class="item-badge">即将上线</span>
            </div>
          </div>
        </div>
      </nav>

      <div class="sidebar-footer">
        <p class="version">AI 运维执行平台 V2.0</p>
        <p class="support">Agent 模式 · 多端适配</p>
      </div>
    </div>
  </aside>
</template>

<style scoped>
.sidebar {
  position: fixed; top: 60px; left: 0; width: 230px; height: calc(100vh - 60px);
  background: var(--bg-sidebar); backdrop-filter: blur(12px);
  border-right: 1px solid var(--border-color); overflow-y: auto; z-index: 99;
  transition: transform 0.3s, background-color 0.3s;
}
.sidebar-mobile { top: 56px; height: calc(100vh - 56px); }
.sidebar-hidden { transform: translateX(-100%); }
.sidebar-overlay { position: fixed; top: 56px; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.5); z-index: 98; }
.sidebar-header { display: flex; align-items: center; justify-content: space-between; padding: 16px 20px; border-bottom: 1px solid var(--border-color); }
.sidebar-title { font-size: 16px; font-weight: 600; color: var(--text-primary); }
.sidebar-close { background: rgba(245,63,63,0.1); border: 1px solid rgba(245,63,63,0.3); border-radius: 8px; color: #F53F3F; cursor: pointer; padding: 6px; }
.sidebar-content { padding: 16px 0; }
.sidebar-nav { padding: 0 8px; }
.sidebar-divider { height: 1px; background: var(--border-color); margin: 8px 12px 12px; opacity: 0.6; }

/* 首页入口 */
.home-entrance {
  display: flex; align-items: center; gap: 12px; padding: 14px; margin: 4px 0;
  border-radius: 12px; cursor: pointer; transition: all 0.2s;
  background: linear-gradient(135deg, rgba(22,93,255,0.1), rgba(15,76,208,0.05));
  border: 1px solid rgba(22,93,255,0.15);
}
.home-entrance:hover { border-color: rgba(22,93,255,0.35); transform: translateX(2px); }
.home-entrance.active {
  background: linear-gradient(135deg, #165DFF, #0F4CD0); color: #fff; border-color: transparent;
  box-shadow: 0 4px 12px rgba(22,93,255,0.3);
}
.home-entrance.active .home-desc, .home-entrance.active .home-arrow { color: rgba(255,255,255,0.7); }
.home-pulse {
  width: 10px; height: 10px; border-radius: 50%; background: #165DFF; flex-shrink: 0;
  animation: dot-pulse 2s infinite;
}
.home-entrance.active .home-pulse { background: #fff; }
@keyframes dot-pulse { 0%,100%{opacity:1;transform:scale(1)} 50%{opacity:0.5;transform:scale(1.3)} }
.home-info { flex: 1; min-width: 0; }
.home-label { font-size: 13px; font-weight: 700; display: block; }
.home-desc { font-size: 11px; opacity: 0.6; margin-top: 1px; }
.home-arrow { font-size: 14px; opacity: 0; transition: opacity 0.2s; }
.home-entrance:hover .home-arrow { opacity: 1; }

/* 菜单组 */
.menu-group { margin-bottom: 8px; }
.group-header {
  display: flex; align-items: center; padding: 10px 12px; cursor: pointer;
  border-radius: 8px; transition: all 0.2s; color: var(--text-secondary);
}
.group-header:hover { background: rgba(22,93,255,0.1); color: var(--text-primary); }
.group-icon { width: 16px; margin-right: 8px; }
.group-name { flex: 1; font-size: 12px; font-weight: 500; }
.group-arrow { color: var(--text-tertiary); transition: transform 0.2s; }
.group-arrow.rotated { transform: rotate(180deg); }
.group-menu { margin-top: 4px; }

.menu-item {
  display: flex; align-items: center; padding: 10px 12px 10px 36px;
  cursor: pointer; border-radius: 8px; transition: all 0.2s;
  color: var(--text-secondary); font-size: 13px;
}
.menu-item:hover { background: rgba(22,93,255,0.1); color: var(--text-primary); }
.menu-item.active { background: rgba(22,93,255,0.2); color: #165DFF; }
.menu-item.coming-soon { opacity: 0.5; cursor: not-allowed; }
.menu-item.coming-soon:hover { background: transparent; color: var(--text-secondary); }

.item-icon { width: 14px; margin-right: 8px; color: var(--text-tertiary); }
.menu-item.active .item-icon { color: #165DFF; }
.item-name { font-size: 13px; flex: 1; }
.item-badge { font-size: 9px; padding: 1px 6px; border-radius: 8px; background: rgba(22,93,255,0.1); color: #165DFF; font-weight: 600; }

/* 底部 */
.sidebar-footer { padding: 20px 16px; margin-top: 16px; border-top: 1px solid var(--border-color); }
.version, .support { font-size: 11px; color: var(--text-tertiary); margin-bottom: 4px; }

@media screen and (max-width: 768px) {
  .sidebar { width: 280px; box-shadow: 0 0 20px rgba(0,0,0,0.3); }
  .sidebar-footer { display: none; }
}
@media screen and (max-width: 480px) {
  .sidebar { width: 250px; }
  .menu-item { padding: 8px 12px 8px 32px; font-size: 12px; }
}
</style>