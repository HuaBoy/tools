<script setup>
import { ref, onMounted, onUnmounted } from 'vue';
import { useRouter, useRoute } from 'vue-router';

const props = defineProps({
  visible: {
    type: Boolean,
    default: true
  }
});

const emit = defineEmits(['close']);

const router = useRouter();
const route = useRoute();
const isMobile = ref(false);

const menuGroups = [
  {
    name: '授权管理',
    icon: 'shield',
    children: [
      { name: '授权码转换工具', path: '/auth/converter', icon: 'key' },
      { name: '三方账号授权', path: '/auth/third-party', icon: 'user' }
    ]
  },
  {
    name: '日志处理专区',
    icon: 'file-text',
    children: [
      { name: '日志解密工具', path: '/log/decrypt', icon: 'lock' },
      { name: '起爆器日志AI分析', path: '/log/analysis', icon: 'brain' }
    ]
  },
  {
    name: 'AI起爆数据专区',
    icon: 'database',
    children: [
      { name: 'AI起爆数据查询', path: '/data/query', icon: 'search' },
      { name: '全链路追溯', path: '/data/trace', icon: 'git-branch' },
      { name: '批次数据追溯', path: '/trace/analysis', icon: 'bar-chart-2' },
      { name: '智能制造系统', path: '/trace/factory-data', icon: 'database' }
    ]
  },
  {
    name: '系统管理',
    icon: 'settings',
    children: [
      { name: '后台管理', path: '/admin', icon: 'shield' },
      { name: '用户管理', path: '/admin/users', icon: 'users' },
      { name: '权限管理', path: '/admin/permissions', icon: 'key' },
      { name: '功能管理', path: '/admin/features', icon: 'toggle' }
    ]
  },
  {
    name: '知识库&文档专区',
    icon: 'book',
    children: [
      { name: 'AI问题数据库', path: '/knowledge/base', icon: 'help-circle' },
      { name: '起爆器版本手册', path: '/knowledge/manual', icon: 'book-open' },
      { name: '版本履历', path: '/knowledge/history', icon: 'history' },
      { name: '录音管理', path: '/knowledge/audio', icon: 'mic' }
    ]
  },
  {
    name: '通用辅助工具',
    icon: 'wrench',
    children: [
      { name: 'AI翻译工具', path: '/tools/translate', icon: 'globe' },
      { name: '文档格式转换', path: '/tools/converter', icon: 'file-text' },
      { name: '二维码生成工具', path: '/tools/qrcode', icon: 'qr-code' }
    ]
  },
  {
    name: 'AI运维智能助手',
    icon: 'bot',
    children: [
      { name: 'AI运维智能助手', path: '/tools/assistant', icon: 'bot' },
      { name: '云系统', path: '/tools/tester', icon: 'terminal' }
    ]
  }
];

const expandedGroups = ref([0, 1, 2, 3, 4, 5]);

const toggleGroup = (index) => {
  const idx = expandedGroups.value.indexOf(index);
  if (idx > -1) {
    expandedGroups.value.splice(idx, 1);
  } else {
    expandedGroups.value.push(index);
  }
};

const isActive = (path) => {
  return route.path === path;
};

const navigateTo = (path) => {
  router.push(path);
  if (isMobile.value) {
    emit('close');
  }
};

const checkMobile = () => {
  isMobile.value = window.innerWidth <= 768;
};

onMounted(() => {
  checkMobile();
  window.addEventListener('resize', checkMobile);
});

onUnmounted(() => {
  window.removeEventListener('resize', checkMobile);
});
</script>

<template>
  <aside class="sidebar" :class="{ 'sidebar-mobile': isMobile, 'sidebar-hidden': isMobile && !visible }">
    <div class="sidebar-overlay" v-if="isMobile && visible" @click="emit('close')"></div>
    <div class="sidebar-content">
      <div class="sidebar-header" v-if="isMobile">
        <span class="sidebar-title">功能菜单</span>
        <button class="sidebar-close" @click="emit('close')">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <line x1="18" y1="6" x2="6" y2="18" />
            <line x1="6" y1="6" x2="18" y2="18" />
          </svg>
        </button>
      </div>
      
      <nav class="sidebar-nav">
        <!-- 首页快捷菜单（始终在顶部） -->
        <div class="home-menu" :class="{ active: isActive('/home') }" @click="navigateTo('/home')">
          <div class="home-icon">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z" />
              <polyline points="9 22 9 12 15 12 15 22" />
            </svg>
          </div>
          <span class="home-name">首页</span>
          <span class="home-badge">智能</span>
        </div>

        <div class="sidebar-divider"></div>

        <div
          v-for="(group, groupIndex) in menuGroups"
          :key="groupIndex"
          class="menu-group"
        >
          <div 
            class="group-header" 
            @click="toggleGroup(groupIndex)"
          >
            <div class="group-icon">
              <svg v-if="group.icon === 'shield'" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
              </svg>
              <svg v-else-if="group.icon === 'file-text'" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
                <polyline points="14 2 14 8 20 8" />
                <line x1="16" y1="13" x2="8" y2="13" />
                <line x1="16" y1="17" x2="8" y2="17" />
                <polyline points="10 9 9 9 8 9" />
              </svg>
              <svg v-else-if="group.icon === 'database'" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <ellipse cx="12" cy="5" rx="9" ry="3" />
                <path d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3" />
                <path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5" />
              </svg>
              <svg v-else-if="group.icon === 'book'" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z" />
                <path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z" />
              </svg>
              <svg v-else-if="group.icon === 'bot'" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />
              </svg>
              <svg v-else-if="group.icon === 'settings'" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <circle cx="12" cy="12" r="3" />
                <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09a1.65 1.65 0 0 0-1-1.51 1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09a1.65 1.65 0 0 0 1.51-1 1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z" />
              </svg>
              <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z" />
              </svg>
            </div>
            <span class="group-name">{{ group.name }}</span>
            <svg 
              class="group-arrow" 
              :class="{ 'rotated': expandedGroups.includes(groupIndex) }"
              width="14" 
              height="14" 
              viewBox="0 0 24 24" 
              fill="none" 
              stroke="currentColor" 
              stroke-width="2" 
              stroke-linecap="round" 
              stroke-linejoin="round"
            >
              <polyline points="6 9 12 15 18 9" />
            </svg>
          </div>
          
          <div 
            v-show="expandedGroups.includes(groupIndex)" 
            class="group-menu"
          >
            <div 
              v-for="item in group.children" 
              :key="item.path"
              class="menu-item"
              :class="{ 'active': isActive(item.path) }"
              @click="navigateTo(item.path)"
            >
              <div class="item-icon">
                <svg v-if="item.icon === 'key'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M21 2l-2 2m-7.61 7.61a5.5 5.5 0 1 1 7.778 7.778L12 17.778l-7.778 7.778a5.5 5.5 0 0 1 7.778-7.778L12 17.778l-7.778 7.778a5.5 5.5 0 0 1 7.778-7.778" />
                </svg>
                <svg v-else-if="item.icon === 'lock'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <rect x="3" y="11" width="18" height="11" rx="2" ry="2" />
                  <path d="M7 11V7a5 5 0 0 1 10 0v4" />
                </svg>
                <svg v-else-if="item.icon === 'brain'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M12 4.5a2.5 2.5 0 0 0-4.96-.46 2.5 2.5 0 0 0-1.2 3.08 2.5 2.5 0 0 0 1.25 4.42 2.5 2.5 0 0 0 3.71 3.71 2.5 2.5 0 0 0 4.42-1.25 2.5 2.5 0 0 0 3.08-1.2A2.5 2.5 0 0 0 19.5 12a2.5 2.5 0 0 0-4.96-.46 2.5 2.5 0 0 0-1.2 3.08 2.5 2.5 0 0 0 1.25 4.42 2.5 2.5 0 0 0 3.71 3.71 2.5 2.5 0 0 0 4.42-1.25 2.5 2.5 0 0 0 3.08-1.2A2.5 2.5 0 0 0 22 12a2.5 2.5 0 0 0-4.96-.46" />
                </svg>
                <svg v-else-if="item.icon === 'search'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <circle cx="11" cy="11" r="8" />
                  <line x1="21" y1="21" x2="16.65" y2="16.65" />
                </svg>
                <svg v-else-if="item.icon === 'git-branch'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <line x1="6" y1="3" x2="6" y2="21" />
                  <circle cx="18" cy="6" r="3" />
                  <circle cx="6" cy="18" r="3" />
                  <path d="M18 9a9 9 0 0 1-9 9" />
                </svg>
                <svg v-else-if="item.icon === 'help-circle'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <circle cx="12" cy="12" r="10" />
                  <path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3" />
                  <line x1="12" y1="17" x2="12.01" y2="17" />
                </svg>
                <svg v-else-if="item.icon === 'book-open'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z" />
                  <path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z" />
                </svg>
                <svg v-else-if="item.icon === 'globe'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <circle cx="12" cy="12" r="10" />
                  <line x1="2" y1="12" x2="22" y2="12" />
                  <path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z" />
                </svg>
                <svg v-else-if="item.icon === 'bar-chart-2'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <line x1="18" y1="20" x2="18" y2="10" />
                  <line x1="12" y1="20" x2="12" y2="4" />
                  <line x1="6" y1="20" x2="6" y2="16" />
                </svg>
                <svg v-else-if="item.icon === 'bot'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />
                </svg>
                <svg v-else-if="item.icon === 'shield'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
                </svg>
                <svg v-else-if="item.icon === 'users'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" />
                  <circle cx="9" cy="7" r="4" />
                  <path d="M23 21v-2a4 4 0 0 0-3-3.87" />
                  <path d="M16 3.13a4 4 0 0 1 0 7.75" />
                </svg>
                <svg v-else-if="item.icon === 'toggle'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <rect x="1" y="5" width="22" height="14" rx="7" ry="7" />
                  <circle cx="16" cy="12" r="3" />
                </svg>
                <svg v-else-if="item.icon === 'history'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M3 3v5h5" />
                  <path d="M3.05 13A9 9 0 1 0 6 5.3L3 8" />
                  <path d="M12 7v5l4 2" />
                </svg>
                <svg v-else-if="item.icon === 'mic'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3z" />
                  <path d="M19 10v2a7 7 0 0 1-14 0v-2" />
                  <line x1="12" y1="19" x2="12" y2="23" />
                  <line x1="8" y1="23" x2="16" y2="23" />
                </svg>
                <svg v-else width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <rect x="3" y="3" width="18" height="18" rx="2" ry="2" />
                  <path d="M3 9h18" />
                  <path d="M9 21V3" />
                </svg>
              </div>
              <span class="item-name">{{ item.name }}</span>
            </div>
          </div>
        </div>
      </nav>
      
      <div class="sidebar-footer">
        <p class="version">软件版本: V1.0.0</p>
        <p class="support">技术支持: AI起爆运维平台</p>
      </div>
    </div>
  </aside>
</template>

<style scoped>
.sidebar {
  position: fixed;
  top: 60px;
  left: 0;
  width: 220px;
  height: calc(100vh - 60px);
  background: var(--bg-sidebar);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border-right: 1px solid var(--border-color);
  overflow-y: auto;
  z-index: 99;
  transition: transform 0.3s ease, background-color 0.3s ease, border-color 0.3s ease;
}

.sidebar-mobile {
  top: 56px;
  height: calc(100vh - 56px);
}

.sidebar-hidden {
  transform: translateX(-100%);
}

.sidebar-overlay {
  position: fixed;
  top: 56px;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 98;
}

.sidebar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-color);
}

.sidebar-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

.sidebar-close {
  background: rgba(245, 63, 63, 0.1);
  border: 1px solid rgba(245, 63, 63, 0.3);
  border-radius: 8px;
  color: #F53F3F;
  cursor: pointer;
  padding: 6px;
  transition: all 0.2s;

  &:hover {
    background: rgba(245, 63, 63, 0.2);
  }
}

.sidebar-content {
  padding: 16px 0;
}

.sidebar-nav {
  padding: 0 8px;
}

.sidebar-divider {
  height: 1px;
  background: var(--border-color);
  margin: 8px 12px 12px 12px;
  opacity: 0.6;
}

.home-menu {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 14px;
  margin: 4px 0;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s;
  position: relative;
  background: linear-gradient(135deg, rgba(22, 93, 255, 0.08) 0%, rgba(15, 76, 208, 0.05) 100%);
  border: 1px solid rgba(22, 93, 255, 0.15);
  color: #165DFF;
  font-weight: 500;
}

.home-menu:hover {
  background: linear-gradient(135deg, rgba(22, 93, 255, 0.15) 0%, rgba(15, 76, 208, 0.1) 100%);
  border-color: rgba(22, 93, 255, 0.3);
  transform: translateX(2px);
}

.home-menu.active {
  background: linear-gradient(135deg, #165DFF 0%, #0F4CD0 100%);
  color: #FFFFFF;
  border-color: transparent;
  box-shadow: 0 2px 8px rgba(22, 93, 255, 0.3);
}

.home-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  flex-shrink: 0;
}

.home-name {
  flex: 1;
  font-size: 14px;
  font-weight: 600;
}

.home-badge {
  font-size: 10px;
  padding: 2px 6px;
  border-radius: 4px;
  background: rgba(255, 255, 255, 0.25);
  color: inherit;
  font-weight: 500;
}

.home-menu:not(.active) .home-badge {
  background: rgba(22, 93, 255, 0.15);
  color: #165DFF;
}

.menu-group {
  margin-bottom: 8px;
}

.group-header {
  display: flex;
  align-items: center;
  padding: 10px 12px;
  cursor: pointer;
  border-radius: 8px;
  transition: all 0.2s;
  color: var(--text-secondary);

  &:hover {
    background: rgba(22, 93, 255, 0.1);
    color: var(--text-primary);
  }
}

.group-icon {
  width: 16px;
  margin-right: 8px;
}

.group-name {
  flex: 1;
  font-size: 12px;
  font-weight: 500;
}

.group-arrow {
  color: var(--text-tertiary);
  transition: transform 0.2s;

  &.rotated {
    transform: rotate(180deg);
  }
}

.group-menu {
  margin-top: 4px;
}

.menu-item {
  display: flex;
  align-items: center;
  padding: 10px 12px 10px 36px;
  cursor: pointer;
  border-radius: 8px;
  transition: all 0.2s;
  color: var(--text-secondary);
  font-size: 13px;

  &:hover {
    background: rgba(22, 93, 255, 0.1);
    color: var(--text-primary);
  }

  &.active {
    background: rgba(22, 93, 255, 0.2);
    color: #165DFF;

    .item-icon {
      color: #165DFF;
    }
  }
}

.item-icon {
  width: 14px;
  margin-right: 8px;
  color: var(--text-tertiary);
}

.item-name {
  font-size: 13px;
}

.sidebar-footer {
  padding: 20px 16px;
  margin-top: 16px;
  border-top: 1px solid var(--border-color);
}

.version,
.support {
  font-size: 11px;
  color: var(--text-tertiary);
  margin-bottom: 4px;
}

@media screen and (max-width: 768px) {
  .sidebar {
    width: 280px;
    box-shadow: 0 0 20px rgba(0, 0, 0, 0.3);
  }

  .sidebar-footer {
    display: none;
  }
}

@media screen and (max-width: 480px) {
  .sidebar {
    width: 250px;
  }

  .group-name {
    font-size: 11px;
  }

  .menu-item {
    padding: 8px 12px 8px 32px;
    font-size: 12px;
  }
}
</style>