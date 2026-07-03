<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue';
import { useRouter } from 'vue-router';
import { ElMessage } from 'element-plus';

const router = useRouter();
const searchKeyword = ref('');
const showResults = ref(false);
const inputRef = ref(null);
const isComposing = false;

// 完整功能库（含路由和参数）
const allFeatures = [
  { name: '授权码转换工具', path: '/auth/converter', category: '授权管理', keywords: ['授权', '码', '转换', 'auth', 'converter', '设备码', '授权码'] },
  { name: '三方账号授权', path: '/auth/third-party', category: '授权管理', keywords: ['三方', '账号', '授权', '登录', '云系统', '智能制造', 'auth', '第三方'] },
  { name: '日志解密工具', path: '/log/decrypt', category: '日志处理专区', keywords: ['日志', '解密', 'log', 'decrypt'] },
  { name: '起爆器日志AI分析', path: '/log/analysis', category: '日志处理专区', keywords: ['起爆器', '日志', 'AI', '分析', 'log', 'analysis', '智能分析'] },
  { name: 'AI起爆数据查询', path: '/data/query', category: 'AI起爆数据专区', keywords: ['AI', '起爆', '数据', '查询', 'data', 'query', '起爆数据'] },
  { name: '全链路追溯', path: '/data/trace', category: 'AI起爆数据专区', keywords: ['全链路', '追溯', 'trace', '链路', '数据追溯'] },
  { name: '批次数据追溯', path: '/trace/analysis', category: 'AI起爆数据专区', keywords: ['批次', '数据', '追溯', 'batch', '批次号'] },
  { name: '智能制造系统', path: '/trace/factory-data', category: 'AI起爆数据专区', keywords: ['智能', '制造', '系统', '工厂', 'factory', 'iot'] },
  { name: '后台管理', path: '/admin', category: '系统管理', keywords: ['后台', '管理', 'admin', '系统管理'] },
  { name: '用户管理', path: '/admin/users', category: '系统管理', keywords: ['用户', '管理', 'user', '用户列表'] },
  { name: '权限管理', path: '/admin/permissions', category: '系统管理', keywords: ['权限', '管理', 'permission', '角色', 'role'] },
  { name: '功能管理', path: '/admin/features', category: '系统管理', keywords: ['功能', '管理', 'feature', '功能开关'] },
  { name: 'AI问题数据库', path: '/knowledge/base', category: '知识库&文档专区', keywords: ['AI', '问题', '数据库', 'knowledge', '问题库'] },
  { name: '起爆器版本手册', path: '/knowledge/manual', category: '知识库&文档专区', keywords: ['起爆器', '版本', '手册', 'manual', '文档'] },
  { name: '版本履历', path: '/knowledge/history', category: '知识库&文档专区', keywords: ['版本', '履历', 'history', '更新日志', '变更'] },
  { name: 'AI翻译工具', path: '/tools/translate', category: '通用辅助工具', keywords: ['AI', '翻译', 'translate', '多语言', '国际化'] },
  { name: '文档格式转换', path: '/tools/converter', category: '通用辅助工具', keywords: ['文档', '格式', '转换', 'converter', 'pdf', 'word'] },
  { name: '二维码生成工具', path: '/tools/qrcode', category: '通用辅助工具', keywords: ['二维码', '生成', 'qrcode', 'qr', '条形码'] },
  { name: 'AI运维智能助手', path: '/tools/assistant', category: 'AI运维智能助手', keywords: ['AI', '运维', '智能', '助手', 'assistant', '运维助手'] },
  { name: '云系统', path: '/tools/tester', category: 'AI运维智能助手', keywords: ['云', '系统', 'tester', '云平台', '云系统'] }
];

// 设备类型映射
const deviceTypeMap = {
  '一测': 'D',
  '二测': 'E',
  '成测': 'B',
  '发火电阻': 'R',
  '快检': 'K',
  '注码': 'C'
};

// 提取批次号（格式如 1-5901-d5-2-n-1, ABC-1234-XY, 等）
const extractBatchNumber = (text) => {
  // 匹配包含字母数字和连字符的批次号
  const patterns = [
    /批次号[：:为]?\s*([A-Za-z0-9][A-Za-z0-9\-_]+[A-Za-z0-9])/i,
    /batch[_\s]?no[：:为]?\s*([A-Za-z0-9][A-Za-z0-9\-_]+[A-Za-z0-9])/i,
    /([A-Za-z0-9]+(?:[-_][A-Za-z0-9]+){2,})/g  // 通用模式：至少3段
  ];

  for (const pattern of patterns) {
    const match = text.match(pattern);
    if (match) {
      return match[1] || match[0];
    }
  }
  return null;
};

// 提取设备类型
const extractDeviceType = (text) => {
  for (const [name, code] of Object.entries(deviceTypeMap)) {
    if (text.includes(name)) {
      return { name, code };
    }
  }
  return null;
};

// 智能解析命令
const parseCommand = (text) => {
  const result = {
    type: 'search',  // search / command
    targets: [],     // 目标功能列表
    params: {}       // 参数
  };

  const lowerText = text.toLowerCase();
  const batchNo = extractBatchNumber(text);
  const deviceType = extractDeviceType(text);

  if (batchNo) result.params.batchNo = batchNo;
  if (deviceType) result.params.deviceType = deviceType;

  // 检测是否为复合命令
  const isComposite = /并|并且|同时|再|然后|接着|，|,/.test(text);

  // 检测具体操作
  const hasQuery = /查询|筛选|搜索|查找|看看|查一下/.test(text);
  const hasDecrypt = /解密|decode|解开/.test(text);
  const hasAnalyze = /分析|analyze|ai/.test(lowerText);
  const hasImport = /导入|上传|打开|选择|读|文件/.test(text);
  const hasTrace = /追溯|链路|查询|看|查/.test(text);

  // 复合命令：导入日志文件，帮我解密并AI分析日志
  if (hasImport && (hasDecrypt || hasAnalyze)) {
    result.type = 'command';
    const targets = [];
    if (hasDecrypt) targets.push({ path: '/log/decrypt', params: { autoFile: 'open' }, action: 'decrypt' });
    if (hasAnalyze) targets.push({ path: '/log/analysis', params: {}, action: 'analyze' });
    result.targets = targets;
    return result;
  }

  // 批次号相关命令
  if (batchNo) {
    result.type = 'command';

    // 批次数据 + 一测/工厂
    if (deviceType || /一测|二测|成测|工厂|制造/.test(text)) {
      const targets = [];

      // 批次数据追溯
      if (hasTrace || /批次|追溯/.test(text)) {
        targets.push({
          path: '/trace/analysis',
          params: { batchNo, autoQuery: true },
          action: 'trace'
        });
      }

      // 智能制造/一测数据
      if (deviceType || /一测|工厂|制造/.test(text)) {
        const dt = deviceType || (text.includes('一测') ? { name: '一测', code: 'D' } : null);
        if (dt) {
          targets.push({
            path: '/trace/factory-data',
            params: { batchNo, deviceType: dt.code, deviceTypeName: dt.name, autoQuery: true },
            action: 'factory'
          });
        }
      }

      // AI起爆数据查询
      if (/AI|起爆|数据/.test(text)) {
        targets.push({
          path: '/data/query',
          params: { batchNo, autoQuery: true },
          action: 'ai-query'
        });
      }

      if (targets.length > 0) {
        result.targets = targets;
        return result;
      }
    }

    // 仅有批次号
    result.targets = [{
      path: '/trace/analysis',
      params: { batchNo, autoQuery: true },
      action: 'trace'
    }];
    return result;
  }

  // 设备类型查询
  if (deviceType && hasQuery) {
    result.type = 'command';
    result.targets = [{
      path: '/trace/factory-data',
      params: { deviceType: deviceType.code, deviceTypeName: deviceType.name, autoQuery: true },
      action: 'factory'
    }];
    return result;
  }

  // 解密相关
  if (hasDecrypt) {
    result.type = 'command';
    result.targets = [{ path: '/log/decrypt', params: { autoFile: 'open' }, action: 'decrypt' }];
    return result;
  }

  // AI分析相关
  if (hasAnalyze) {
    result.type = 'command';
    result.targets = [{ path: '/log/analysis', params: {}, action: 'analyze' }];
    return result;
  }

  return result;
};

// 模糊搜索匹配（用于功能搜索）
const matchedFeatures = computed(() => {
  const keyword = searchKeyword.value.trim().toLowerCase();
  if (!keyword) return [];

  return allFeatures
    .map(feature => {
      let score = 0;
      const name = feature.name.toLowerCase();
      const category = feature.category.toLowerCase();

      if (name === keyword) score += 100;
      else if (name.includes(keyword)) score += 50;
      if (category.includes(keyword)) score += 20;
      for (const kw of feature.keywords) {
        const lowerKw = kw.toLowerCase();
        if (lowerKw === keyword) score += 30;
        else if (lowerKw.includes(keyword) || keyword.includes(lowerKw)) score += 10;
      }
      return { ...feature, score };
    })
    .filter(f => f.score > 0)
    .sort((a, b) => b.score - a.score)
    .slice(0, 8);
});

// 命令解析结果
const commandResult = computed(() => {
  const text = searchKeyword.value.trim();
  if (!text) return null;
  return parseCommand(text);
});

const hasResults = computed(() => matchedFeatures.value.length > 0);
const hasCommand = computed(() => commandResult.value && commandResult.value.type === 'command');

const groupedResults = computed(() => {
  const groups = {};
  matchedFeatures.value.forEach(feature => {
    if (!groups[feature.category]) groups[feature.category] = [];
    groups[feature.category].push(feature);
  });
  return groups;
});

// 处理执行命令
const handleExecuteCommand = () => {
  const cmd = commandResult.value;
  if (!cmd || !cmd.targets || cmd.targets.length === 0) return;

  const first = cmd.targets[0];
  const query = new URLSearchParams();
  Object.entries(first.params || {}).forEach(([k, v]) => {
    if (v !== null && v !== undefined) query.append(k, v);
  });

  const queryStr = query.toString();
  const url = queryStr ? `${first.path}?${queryStr}` : first.path;

  showResults.value = false;
  searchKeyword.value = '';
  router.push(url);
  ElMessage.success(`正在打开：${first.path}`);
};

const handleNavigate = (path) => {
  showResults.value = false;
  searchKeyword.value = '';
  router.push(path);
};

const handleKeydown = (e) => {
  if (e.key === 'Enter' && !e.isComposing) {
    if (hasCommand.value) {
      handleExecuteCommand();
    } else if (matchedFeatures.value.length > 0) {
      handleNavigate(matchedFeatures.value[0].path);
    }
  } else if (e.key === 'Escape') {
    showResults.value = false;
    inputRef.value?.blur();
  }
};

const handleClickOutside = (e) => {
  if (!e.target.closest('.search-container')) {
    showResults.value = false;
  }
};

// 示例命令
const exampleCommands = [
  { icon: '🔍', text: '请根据批次号1-5901-D5-2-N-1筛选数据' },
  { icon: '🏭', text: '根据批次号查询一测数据' },
  { icon: '🔓', text: '导入日志文件，帮我解密并AI分析日志' },
  { icon: '🔄', text: '批次号A001的全链路追溯' }
];

// 快速功能
const quickFeatures = [
  { name: '批次数据追溯', path: '/trace/analysis', icon: '📊' },
  { name: '智能制造系统', path: '/trace/factory-data', icon: '🏭' },
  { name: '日志解密', path: '/log/decrypt', icon: '🔓' },
  { name: 'AI起爆数据', path: '/data/query', icon: '🤖' },
  { name: 'AI翻译', path: '/tools/translate', icon: '🌐' },
  { name: '三方账号授权', path: '/auth/third-party', icon: '🔐' }
];

onMounted(() => {
  document.addEventListener('click', handleClickOutside);
  setTimeout(() => {
    inputRef.value?.focus();
  }, 100);
});

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside);
});
</script>

<template>
  <div class="home-page">
    <div class="hero-section">
      <div class="hero-content">
        <h1 class="hero-title">
          <span class="title-icon">⚡</span>
          <span>起爆器运维智能工具集</span>
        </h1>
        <p class="hero-subtitle">输入指令或功能名称，AI 帮你自动匹配并跳转</p>

        <div class="search-container">
          <div class="search-box" :class="{ focused: showResults && searchKeyword }">
            <svg class="search-icon" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="11" cy="11" r="8" />
              <line x1="21" y1="21" x2="16.65" y2="16.65" />
            </svg>
            <textarea
              ref="inputRef"
              v-model="searchKeyword"
              class="search-input"
              placeholder="试试输入：请根据批次号1-5901-D5-2-N-1查询一测数据"
              rows="2"
              @focus="showResults = true"
              @input="showResults = true"
              @keydown="handleKeydown"
            ></textarea>
            <button
              v-if="searchKeyword"
              class="clear-btn"
              @click="searchKeyword = ''; inputRef?.focus()"
            >
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <line x1="18" y1="6" x2="6" y2="18" />
                <line x1="6" y1="6" x2="18" y2="18" />
              </svg>
            </button>
            <button
              v-else
              class="search-action-btn"
              title="搜索"
              @click="hasCommand ? handleExecuteCommand() : matchedFeatures.length > 0 && handleNavigate(matchedFeatures[0].path)"
            >
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <line x1="5" y1="12" x2="19" y2="12" />
                <polyline points="12 5 19 12 12 19" />
              </svg>
            </button>
          </div>

          <!-- 智能命令结果 -->
          <div v-if="showResults && hasCommand" class="command-result">
            <div class="command-header">
              <span class="command-icon">🤖</span>
              <span>AI 检测到智能指令，将自动跳转到：</span>
            </div>
            <div
              v-for="(target, index) in commandResult.targets"
              :key="index"
              class="command-target"
            >
              <div class="target-icon">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <polyline points="9 18 15 12 9 6" />
                </svg>
              </div>
              <div class="target-info">
                <div class="target-path">{{ target.path }}</div>
                <div class="target-params" v-if="target.params && Object.keys(target.params).length > 0">
                  <span
                    v-for="(value, key) in target.params"
                    :key="key"
                    class="param-chip"
                  >
                    {{ key }}: {{ value }}
                  </span>
                </div>
              </div>
            </div>
            <button class="execute-btn" @click="handleExecuteCommand">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <polygon points="5 3 19 12 5 21 5 3" />
              </svg>
              <span>执行指令</span>
              <kbd>Enter</kbd>
            </button>
          </div>

          <!-- 功能搜索结果 -->
          <div v-else-if="showResults && searchKeyword && !hasCommand" class="search-results">
            <div v-if="!hasResults" class="empty-results">
              <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                <circle cx="12" cy="12" r="10" />
                <line x1="12" y1="8" x2="12" y2="12" />
                <line x1="12" y1="16" x2="12.01" y2="16" />
              </svg>
              <p>未找到匹配的功能</p>
              <p class="empty-hint">试试其他关键词，或输入完整指令如"查询批次号XXX"</p>
            </div>

            <div v-else class="results-list">
              <div
                v-for="(group, category) in groupedResults"
                :key="category"
                class="result-group"
              >
                <div class="group-title">{{ category }}</div>
                <div
                  v-for="(feature, index) in group"
                  :key="feature.path + index"
                  class="result-item"
                  @click="handleNavigate(feature.path)"
                >
                  <div class="result-icon">
                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                      <polyline points="9 18 15 12 9 6" />
                    </svg>
                  </div>
                  <div class="result-info">
                    <div class="result-name">{{ feature.name }}</div>
                    <div class="result-path">{{ feature.path }}</div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 示例指令 -->
        <div v-if="!searchKeyword" class="example-commands">
          <div class="examples-title">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M12 2L2 7l10 5 10-5-10-5z" />
              <path d="M2 17l10 5 10-5" />
              <path d="M2 12l10 5 10-5" />
            </svg>
            <span>试试这些指令</span>
          </div>
          <div class="example-list">
            <div
              v-for="(cmd, index) in exampleCommands"
              :key="index"
              class="example-item"
              @click="searchKeyword = cmd.text; inputRef?.focus()"
            >
              <span class="example-icon">{{ cmd.icon }}</span>
              <span class="example-text">{{ cmd.text }}</span>
              <svg class="example-arrow" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <line x1="5" y1="12" x2="19" y2="12" />
                <polyline points="12 5 19 12 12 19" />
              </svg>
            </div>
          </div>
        </div>

        <!-- 快速功能 -->
        <div v-if="!searchKeyword" class="quick-features">
          <div class="quick-title">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2" />
            </svg>
            <span>快速访问</span>
          </div>
          <div class="quick-grid">
            <div
              v-for="feature in quickFeatures"
              :key="feature.path"
              class="quick-item"
              @click="handleNavigate(feature.path)"
            >
              <span class="quick-icon">{{ feature.icon }}</span>
              <span class="quick-name">{{ feature.name }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.home-page {
  width: 100%;
  min-height: calc(100vh - 100px);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
  box-sizing: border-box;
}

.hero-section {
  width: 100%;
  max-width: 820px;
}

.hero-content {
  text-align: center;
}

.hero-title {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  font-size: 36px;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0 0 12px 0;
  line-height: 1.2;
}

.title-icon {
  font-size: 42px;
  filter: drop-shadow(0 0 12px rgba(255, 200, 0, 0.5));
}

.hero-subtitle {
  font-size: 16px;
  color: var(--text-tertiary);
  margin: 0 0 36px 0;
}

.search-container {
  position: relative;
  width: 100%;
  margin-bottom: 32px;
}

.search-box {
  display: flex;
  align-items: flex-start;
  gap: 14px;
  background: var(--bg-card);
  border: 2px solid var(--border-color);
  border-radius: 18px;
  padding: 20px 22px;
  transition: all 0.2s;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.06);
  min-height: 80px;
}

.search-box:focus-within,
.search-box.focused {
  border-color: rgba(22, 93, 255, 0.6);
  box-shadow: 0 8px 32px rgba(22, 93, 255, 0.15);
  transform: translateY(-1px);
}

.search-icon {
  color: #94A3B8;
  flex-shrink: 0;
  margin-top: 4px;
}

.search-box:focus-within .search-icon {
  color: #165DFF;
}

.search-input {
  flex: 1;
  background: transparent;
  border: none;
  outline: none;
  font-size: 17px;
  color: var(--text-primary);
  font-weight: 500;
  min-width: 0;
  resize: none;
  font-family: inherit;
  line-height: 1.6;
  max-height: 120px;
}

.search-input::placeholder {
  color: #94A3B8;
  font-weight: 400;
}

.clear-btn {
  background: rgba(100, 116, 139, 0.1);
  border: none;
  border-radius: 50%;
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #64748B;
  cursor: pointer;
  transition: all 0.2s;
  flex-shrink: 0;
  margin-top: 2px;
}

.clear-btn:hover {
  background: rgba(245, 63, 63, 0.1);
  color: #F53F3F;
}

.search-action-btn {
  background: linear-gradient(135deg, #165DFF 0%, #0F4CD0 100%);
  border: none;
  border-radius: 50%;
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #FFFFFF;
  cursor: pointer;
  transition: all 0.2s;
  flex-shrink: 0;
  margin-top: 0;
}

.search-action-btn:hover {
  transform: scale(1.05);
  box-shadow: 0 4px 12px rgba(22, 93, 255, 0.4);
}

/* 智能命令结果 */
.command-result {
  position: absolute;
  top: calc(100% + 8px);
  left: 0;
  right: 0;
  background: var(--bg-card);
  border: 1px solid rgba(22, 93, 255, 0.3);
  border-radius: 18px;
  box-shadow: 0 12px 40px rgba(22, 93, 255, 0.15);
  z-index: 1000;
  text-align: left;
  overflow: hidden;
}

.command-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 14px 20px;
  background: linear-gradient(135deg, rgba(22, 93, 255, 0.08) 0%, rgba(15, 76, 208, 0.05) 100%);
  font-size: 13px;
  color: #165DFF;
  font-weight: 600;
  border-bottom: 1px solid rgba(22, 93, 255, 0.15);
}

.command-icon {
  font-size: 18px;
}

.command-target {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 20px;
  border-bottom: 1px solid var(--border-color);
}

.command-target:last-of-type {
  border-bottom: none;
}

.target-icon {
  width: 32px;
  height: 32px;
  background: rgba(22, 93, 255, 0.1);
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #165DFF;
  flex-shrink: 0;
}

.target-info {
  flex: 1;
  min-width: 0;
}

.target-path {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 4px;
  font-family: monospace;
}

.target-params {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.param-chip {
  font-size: 11px;
  padding: 2px 8px;
  background: rgba(22, 93, 255, 0.08);
  color: #165DFF;
  border-radius: 4px;
  font-family: monospace;
}

.execute-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  width: calc(100% - 40px);
  margin: 0 20px 20px 20px;
  padding: 12px 24px;
  background: linear-gradient(135deg, #165DFF 0%, #0F4CD0 100%);
  border: none;
  border-radius: 12px;
  color: #FFFFFF;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.execute-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(22, 93, 255, 0.4);
}

.execute-btn kbd {
  background: rgba(255, 255, 255, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.3);
  border-radius: 4px;
  padding: 1px 6px;
  font-size: 10px;
  font-family: monospace;
  margin-left: 4px;
}

/* 搜索结果 */
.search-results {
  position: absolute;
  top: calc(100% + 8px);
  left: 0;
  right: 0;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 18px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.12);
  max-height: 480px;
  overflow-y: auto;
  z-index: 1000;
  text-align: left;
}

.empty-results {
  padding: 60px 20px;
  text-align: center;
  color: var(--text-tertiary);
}

.empty-results svg {
  color: #CBD5E1;
  margin-bottom: 12px;
}

.empty-results p {
  margin: 0 0 4px 0;
  font-size: 14px;
}

.empty-hint {
  font-size: 12px;
  color: #94A3B8;
}

.results-list {
  padding: 8px 0;
}

.result-group {
  margin-bottom: 8px;
}

.result-group:last-child {
  margin-bottom: 0;
}

.group-title {
  padding: 10px 20px;
  font-size: 11px;
  font-weight: 600;
  color: var(--text-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  background: rgba(100, 116, 139, 0.04);
}

.result-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 20px;
  cursor: pointer;
  transition: background 0.15s;
}

.result-item:hover {
  background: rgba(22, 93, 255, 0.06);
}

.result-icon {
  width: 32px;
  height: 32px;
  background: rgba(22, 93, 255, 0.1);
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #165DFF;
  flex-shrink: 0;
}

.result-info {
  flex: 1;
  min-width: 0;
}

.result-name {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary);
  margin-bottom: 2px;
}

.result-path {
  font-size: 11px;
  color: var(--text-tertiary);
  font-family: monospace;
}

/* 示例指令 */
.example-commands {
  margin-top: 24px;
  text-align: left;
}

.examples-title,
.quick-title {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--text-tertiary);
  font-weight: 600;
  margin-bottom: 12px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.examples-title svg,
.quick-title svg {
  color: #94A3B8;
}

.example-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.example-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background: var(--bg-input);
  border: 1px solid var(--border-color);
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s;
  font-size: 13px;
}

.example-item:hover {
  background: rgba(22, 93, 255, 0.06);
  border-color: rgba(22, 93, 255, 0.3);
  transform: translateX(4px);
}

.example-icon {
  font-size: 18px;
  flex-shrink: 0;
}

.example-text {
  flex: 1;
  color: var(--text-secondary);
  text-align: left;
}

.example-arrow {
  color: var(--text-tertiary);
  flex-shrink: 0;
}

/* 快速功能 */
.quick-features {
  margin-top: 24px;
  text-align: left;
}

.quick-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  gap: 10px;
}

.quick-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 14px 16px;
  background: var(--bg-input);
  border: 1px solid var(--border-color);
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s;
}

.quick-item:hover {
  background: rgba(22, 93, 255, 0.06);
  border-color: rgba(22, 93, 255, 0.3);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(22, 93, 255, 0.1);
}

.quick-icon {
  font-size: 20px;
  flex-shrink: 0;
}

.quick-name {
  font-size: 13px;
  color: var(--text-primary);
  font-weight: 500;
}

/* 平板 */
@media screen and (max-width: 768px) {
  .home-page {
    padding: 20px 16px;
  }

  .hero-title {
    font-size: 26px;
  }

  .title-icon {
    font-size: 32px;
  }

  .hero-subtitle {
    font-size: 14px;
  }

  .search-box {
    padding: 16px 18px;
    min-height: 70px;
  }

  .search-input {
    font-size: 15px;
  }
}

/* 手机 */
@media screen and (max-width: 480px) {
  .home-page {
    padding: 16px 12px;
  }

  .hero-title {
    font-size: 22px;
    flex-direction: column;
    gap: 4px;
  }

  .title-icon {
    font-size: 28px;
  }

  .hero-subtitle {
    font-size: 13px;
    margin-bottom: 24px;
  }

  .search-box {
    padding: 14px 16px;
    min-height: 64px;
    gap: 10px;
  }

  .search-input {
    font-size: 14px;
  }

  .search-icon {
    width: 20px;
    height: 20px;
  }

  .command-header {
    padding: 10px 14px;
    font-size: 12px;
  }

  .command-target {
    padding: 10px 14px;
  }

  .target-path {
    font-size: 12px;
  }

  .execute-btn {
    margin: 0 14px 14px 14px;
    width: calc(100% - 28px);
    padding: 10px 16px;
    font-size: 13px;
  }

  .result-item {
    padding: 10px 14px;
  }

  .quick-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .quick-item {
    padding: 10px 12px;
  }

  .quick-name {
    font-size: 12px;
  }
}
</style>
