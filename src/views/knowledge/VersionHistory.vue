<script setup>
import { ref, computed } from 'vue';
import { ElMessage } from 'element-plus';
import GlassCard from '@/components/GlassCard.vue';
import { useLogsStore } from '@/stores/logs';
import { getDeviceTypeByAppVersion, getAllDeviceTypes } from '@/utils/deviceType.js';

const logsStore = useLogsStore();

const selectedRegion = ref('domestic');
const selectedVersion = ref(null);
const compareVersion = ref(null);
const showCompare = ref(false);

// 存储已上传的版本（按区域分类）
const uploadedVersions = ref({
  domestic: [],
  overseas: []
});

// ==================== 手动新增版本 ====================
// 新增版本对话框
const showAddVersionDialog = ref(false);

// 新增版本表单
const versionForm = ref({
  region: 'domestic',         // 区域：domestic / overseas
  id: '',                     // 版本号
  date: new Date().toISOString().slice(0, 10), // 发布日期
  appVersion: '',             // 软件版本（APP版本）
  controllerVersion: '',      // 控制器版本
  usedDevice: '',             // 使用设备
  features: []                // 测试项列表
});

// 临时的测试项输入
const tempFeature = ref({
  type: 'new',
  text: ''
});

// 打开新增版本对话框
const openAddVersionDialog = () => {
  versionForm.value = {
    region: selectedRegion.value,
    id: '',
    date: new Date().toISOString().slice(0, 10),
    appVersion: '',
    controllerVersion: '',
    usedDevice: '',
    features: []
  };
  tempFeature.value = { type: 'new', text: '' };
  showAddVersionDialog.value = true;
};

// 添加测试项到表单
const addFeatureToForm = () => {
  if (!tempFeature.value.text.trim()) {
    ElMessage.warning('请输入测试项内容');
    return;
  }
  versionForm.value.features.push({
    type: tempFeature.value.type,
    text: tempFeature.value.text.trim()
  });
  tempFeature.value = { type: 'new', text: '' };
};

// 移除表单中的测试项
const removeFeatureFromForm = (index) => {
  versionForm.value.features.splice(index, 1);
};

// 监听 APP 版本变化，自动推荐使用设备
const onAppVersionChange = (val) => {
  if (val && !versionForm.value.usedDevice) {
    const deviceType = getDeviceTypeByAppVersion(val);
    if (deviceType && deviceType.name) {
      versionForm.value.usedDevice = deviceType.name;
    }
  }
};

// 保存新增的版本
const saveNewVersion = () => {
  // 验证必填
  if (!versionForm.value.id.trim()) {
    ElMessage.error('请输入版本号');
    return;
  }
  if (!versionForm.value.date) {
    ElMessage.error('请选择发布日期');
    return;
  }
  if (versionForm.value.features.length === 0) {
    ElMessage.error('请至少添加一个测试项');
    return;
  }

  // 检查版本号是否重复
  const existingList = versionForm.value.region === 'overseas'
    ? uploadedVersions.value.overseas
    : uploadedVersions.value.domestic;

  if (existingList.some(v => v.id === versionForm.value.id)) {
    ElMessage.error(`版本号 ${versionForm.value.id} 已存在`);
    return;
  }

  // 自动根据软件版本获取设备类型
  const deviceTypeByVersion = getDeviceTypeByAppVersion(versionForm.value.appVersion);
  const finalDevice = versionForm.value.usedDevice || deviceTypeByVersion.name;

  // 生成标题
  let title = '';
  if (versionForm.value.appVersion) {
    title = `APP ${versionForm.value.appVersion}`;
  }
  if (versionForm.value.controllerVersion) {
    title += title ? ` / 控制器 ${versionForm.value.controllerVersion}` : `控制器 ${versionForm.value.controllerVersion}`;
  }
  if (!title) {
    title = `版本 ${versionForm.value.id}`;
  }

  // 构建版本对象
  const newVersion = {
    id: versionForm.value.id,
    date: versionForm.value.date,
    appVersion: versionForm.value.appVersion,
    controllerVersion: versionForm.value.controllerVersion,
    usedDevice: finalDevice,
    deviceTypeColor: deviceTypeByVersion.color,
    title: title,
    features: [...versionForm.value.features],
    source: 'manual'
  };

  // 添加到列表（按区域）
  if (versionForm.value.region === 'overseas') {
    uploadedVersions.value.overseas.unshift(newVersion);
  } else {
    uploadedVersions.value.domestic.unshift(newVersion);
  }

  // 保存到 localStorage
  saveUploadedVersions();

  // 切换到对应区域
  selectedRegion.value = newVersion.region || 'domestic';
  selectedVersion.value = newVersion;

  showAddVersionDialog.value = false;
  ElMessage.success(`版本 ${newVersion.id} 已添加`);

  logsStore.addLog('新增', '版本履历', `新增版本: ${newVersion.id}`);
};

// 从 localStorage 加载已保存的版本
const loadUploadedVersions = () => {
  try {
    const saved = localStorage.getItem('version_history_uploaded');
    if (saved) {
      const parsed = JSON.parse(saved);
      if (parsed && typeof parsed === 'object') {
        uploadedVersions.value = {
          domestic: Array.isArray(parsed.domestic) ? parsed.domestic : [],
          overseas: Array.isArray(parsed.overseas) ? parsed.overseas : []
        };
      }
    }
  } catch (e) {
    console.warn('加载本地版本失败:', e);
  }
};

// 保存到 localStorage
const saveUploadedVersions = () => {
  try {
    localStorage.setItem('version_history_uploaded', JSON.stringify(uploadedVersions.value));
  } catch (e) {
    console.warn('保存版本失败:', e);
  }
};

// 页面加载时读取
loadUploadedVersions();

const domesticVersions = [
  {
    id: 'v3.2.0',
    date: '2026-06-20',
    title: '新增AI智能分析功能',
    features: [
      { type: 'new', text: '新增AI日志分析功能，支持智能识别异常日志' },
      { type: 'new', text: '新增电压电流曲线图动态范围计算' },
      { type: 'new', text: '新增数据缩放、拖动、筛选功能' },
      { type: 'improve', text: '优化登录验证流程，支持2天免登录' },
      { type: 'fix', text: '修复曲线图顶部被遮挡问题' },
      { type: 'fix', text: '修复电流超过1500uA不显示问题' }
    ]
  },
  {
    id: 'v3.1.0',
    date: '2026-05-15',
    title: '多语言翻译支持',
    features: [
      { type: 'new', text: '新增AI翻译工具，支持8种语言翻译' },
      { type: 'new', text: '新增本地术语映射表，支持起爆器专业术语' },
      { type: 'new', text: '支持XML和TSV格式文件导入导出' },
      { type: 'improve', text: '优化翻译准确率，提升专业术语翻译质量' },
      { type: 'fix', text: '修复批量翻译时部分内容丢失问题' }
    ]
  },
  {
    id: 'v3.0.0',
    date: '2026-04-10',
    title: '知识库系统上线',
    features: [
      { type: 'new', text: '新增AI问题数据库，基于本地RAG系统' },
      { type: 'new', text: '支持向量匹配、阈值过滤、结果排序' },
      { type: 'new', text: '支持知识分类：硬件、固件、组网、授权、日志' },
      { type: 'new', text: '支持新增知识条目入库' },
      { type: 'improve', text: '优化检索算法，提升匹配准确度' }
    ]
  },
  {
    id: 'v2.5.0',
    date: '2026-03-05',
    title: 'API测试助手',
    features: [
      { type: 'new', text: '新增API测试助手，支持登录验证测试' },
      { type: 'new', text: '支持设备查询接口测试' },
      { type: 'new', text: '支持凭证本地保存，3天免登录' },
      { type: 'new', text: '支持双条件查询（deviceCode和controllerCode）' },
      { type: 'improve', text: '优化错误处理机制' }
    ]
  },
  {
    id: 'v2.0.0',
    date: '2026-01-20',
    title: '系统重构升级',
    features: [
      { type: 'new', text: '全新UI界面设计，玻璃态风格' },
      { type: 'new', text: '新增数据查询模块' },
      { type: 'new', text: '新增数据追踪模块' },
      { type: 'improve', text: '优化系统性能，提升响应速度' },
      { type: 'fix', text: '修复多处已知Bug' }
    ]
  }
];

const overseasVersions = [
  {
    id: 'v3.2.0',
    date: '2026-06-25',
    title: 'International Release - AI Analysis',
    features: [
      { type: 'new', text: 'AI log analysis with intelligent anomaly detection' },
      { type: 'new', text: 'Dynamic range calculation for voltage/current charts' },
      { type: 'new', text: 'Multi-language support: English, Arabic, Russian, Japanese, Korean, French, German' },
      { type: 'new', text: 'Document format converter: Word, Excel, PDF, CSV conversion' },
      { type: 'improve', text: 'Optimized login flow with 2-day auto-login' },
      { type: 'fix', text: 'Fixed chart display issues for high current values' }
    ]
  },
  {
    id: 'v3.1.0',
    date: '2026-05-20',
    title: 'International Release - Translation Tool',
    features: [
      { type: 'new', text: 'AI Translation Tool with 8 language support' },
      { type: 'new', text: 'Professional terminology mapping for blasting equipment' },
      { type: 'new', text: 'XML and TSV file import/export support' },
      { type: 'improve', text: 'Enhanced translation accuracy for technical terms' },
      { type: 'fix', text: 'Fixed batch translation content loss issue' }
    ]
  },
  {
    id: 'v3.0.0',
    date: '2026-04-15',
    title: 'International Release - Knowledge Base',
    features: [
      { type: 'new', text: 'AI Question Database with local RAG system' },
      { type: 'new', text: 'Vector matching, threshold filtering, result sorting' },
      { type: 'new', text: 'Knowledge categories: Hardware, Firmware, Network, Authorization, Logs' },
      { type: 'new', text: 'New knowledge entry submission support' },
      { type: 'improve', text: 'Optimized search algorithm for better accuracy' }
    ]
  },
  {
    id: 'v2.5.0',
    date: '2026-03-10',
    title: 'International Release - API Tester',
    features: [
      { type: 'new', text: 'API Test Assistant for login verification' },
      { type: 'new', text: 'Device query API testing support' },
      { type: 'new', text: 'Credential local storage with 3-day auto-login' },
      { type: 'new', text: 'Dual condition query (deviceCode and controllerCode)' },
      { type: 'improve', text: 'Enhanced error handling mechanism' }
    ]
  },
  {
    id: 'v2.0.0',
    date: '2026-01-25',
    title: 'International Release - System Upgrade',
    features: [
      { type: 'new', text: 'New UI design with glassmorphism style' },
      { type: 'new', text: 'Data query module' },
      { type: 'new', text: 'Data trace module' },
      { type: 'improve', text: 'Performance optimization for faster response' },
      { type: 'fix', text: 'Multiple bug fixes' }
    ]
  }
];

const currentVersions = computed(() => {
  const baseList = selectedRegion.value === 'domestic' ? domesticVersions : overseasVersions;
  // 合并已上传的版本（去重，以 id 为准）
  const uploaded = uploadedVersions.value[selectedRegion.value] || [];
  const map = new Map();
  [...baseList, ...uploaded].forEach(v => {
    if (!map.has(v.id)) {
      map.set(v.id, v);
    }
  });
  // 按日期降序
  return Array.from(map.values()).sort((a, b) => (b.date || '').localeCompare(a.date || ''));
});

/**
 * 获取版本对应的设备类型（基于软件版本首字母）
 */
const getVersionDeviceType = (version) => {
  const appVer = version.appVersion || version.id;
  return getDeviceTypeByAppVersion(appVer);
};

/**
 * 获取版本对应的设备类型名称
 */
const getVersionDeviceTypeName = (version) => {
  return getVersionDeviceType(version).name;
};

/**
 * 获取所有支持的设备类型（用于筛选器）
 */
const supportedDeviceTypes = computed(() => {
  return getAllDeviceTypes();
});

const handleSelectVersion = (version) => {
  selectedVersion.value = version;
  logsStore.addLog('查看', '版本履历', `查看版本: ${version.id}`);
};

const handleCompareVersion = (version) => {
  if (!selectedVersion.value) {
    ElMessage.warning('请先选择一个版本作为基准');
    return;
  }
  if (selectedVersion.value.id === version.id) {
    ElMessage.warning('请选择不同的版本进行对比');
    return;
  }
  compareVersion.value = version;
  showCompare.value = true;
  logsStore.addLog('对比', '版本履历', `对比版本: ${selectedVersion.value.id} vs ${version.id}`);
};

const closeCompare = () => {
  showCompare.value = false;
  compareVersion.value = null;
};

const getFeatureTypeLabel = (type) => {
  switch (type) {
    case 'new': return '新增';
    case 'improve': return '优化';
    case 'fix': return '修复';
    case 'remove': return '移除';
    default: return '其他';
  }
};

const getFeatureTypeClass = (type) => {
  switch (type) {
    case 'new': return 'feature-new';
    case 'improve': return 'feature-improve';
    case 'fix': return 'feature-fix';
    case 'remove': return 'feature-remove';
    default: return 'feature-other';
  }
};

const compareFeatures = computed(() => {
  if (!selectedVersion.value || !compareVersion.value) return [];

  const baseFeatures = selectedVersion.value.features.map(f => f.text);
  const compareFeatures = compareVersion.value.features.map(f => f.text);

  const added = compareVersion.value.features.filter(f => !baseFeatures.includes(f.text));
  const removed = selectedVersion.value.features.filter(f => !compareFeatures.includes(f.text));
  const common = selectedVersion.value.features.filter(f => compareFeatures.includes(f.text));

  return {
    added,
    removed,
    common
  };
});

// ============ 手动新增版本（已实现） ============

/**
 * 删除已上传的版本
 */
const handleDeleteUploaded = (version) => {
  const list = uploadedVersions.value[selectedRegion.value];
  const index = list.findIndex(v => v.id === version.id);
  if (index >= 0) {
    list.splice(index, 1);
    if (selectedVersion.value?.id === version.id) {
      selectedVersion.value = null;
    }
    ElMessage.success('已删除');
  }
};
</script>

<template>
  <div class="version-manual">
    <GlassCard title="版本履历">
      <div class="version-container">
        <!-- 区域选择 + 上传按钮 -->
        <div class="toolbar">
          <div class="region-selector">
            <button
              class="region-btn"
              :class="{ active: selectedRegion === 'domestic' }"
              @click="selectedRegion = 'domestic'; selectedVersion = null"
            >
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <circle cx="12" cy="12" r="10" />
                <line x1="2" y1="12" x2="22" y2="12" />
                <path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z" />
              </svg>
              <span>国内版本</span>
            </button>
            <button
              class="region-btn"
              :class="{ active: selectedRegion === 'overseas' }"
              @click="selectedRegion = 'overseas'; selectedVersion = null"
            >
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <circle cx="12" cy="12" r="10" />
                <line x1="2" y1="12" x2="22" y2="12" />
                <path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z" />
              </svg>
              <span>海外版本</span>
            </button>
          </div>

          <button
            class="add-version-btn"
            @click="openAddVersionDialog"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
              <line x1="12" y1="5" x2="12" y2="19" />
              <line x1="5" y1="12" x2="19" y2="12" />
            </svg>
            <span>新增版本</span>
          </button>
        </div>

        <!-- 提示信息 -->
        <div class="upload-hint">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="10" />
            <line x1="12" y1="16" x2="12" y2="12" />
            <line x1="12" y1="8" x2="12.01" y2="8" />
          </svg>
          <span>
            点击"新增版本"按钮添加新版本。
            自动识别：<strong>软件版本</strong>（APP版本）/<strong>控制器版本</strong>（如 8.5-B）/<strong>使用设备</strong>/<strong>测试项</strong>
            <br />设备类型自动判断：<code>L</code>→小勇士、<code>I</code>→DT40、<code>Q</code>→全面屏、<code>K</code>→煤许、<code>T</code>→DT40+小勇士
          </span>
        </div>

        <!-- 版本列表 -->
        <div class="version-list">
          <div 
            v-for="version in currentVersions" 
            :key="version.id"
            class="version-item"
            :class="{ selected: selectedVersion?.id === version.id }"
          >
            <div class="version-header" @click="handleSelectVersion(version)">
              <div class="version-info">
                <span class="version-id">{{ version.id }}</span>
                <span class="version-date">{{ version.date }}</span>
                <span v-if="version.appVersion" class="version-app" title="APP 版本（软件版本）">
                  APP：{{ version.appVersion }}
                </span>
                <span v-if="version.controllerVersion" class="version-controller" title="控制器版本">
                  控制器：{{ version.controllerVersion }}
                </span>
                <span
                  v-if="version.usedDevice"
                  class="version-device"
                  :style="version.deviceTypeColor ? { background: version.deviceTypeColor + '15', color: version.deviceTypeColor, borderColor: version.deviceTypeColor + '40' } : {}"
                  title="使用设备"
                >
                  <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" style="margin-right: 2px;">
                    <rect x="2" y="3" width="20" height="14" rx="2" ry="2" />
                    <line x1="8" y1="21" x2="16" y2="21" />
                    <line x1="12" y1="17" x2="12" y2="21" />
                  </svg>
                  {{ version.usedDevice }}
                </span>
                <span v-if="version.source === 'uploaded'" class="uploaded-badge">已上传</span>
              </div>
              <div class="version-title">{{ version.title }}</div>
              <div class="version-count">
                <span class="count-new">{{ version.features.filter(f => f.type === 'new').length }} 新增</span>
                <span class="count-improve">{{ version.features.filter(f => f.type === 'improve').length }} 优化</span>
                <span class="count-fix">{{ version.features.filter(f => f.type === 'fix').length }} 修复</span>
                <button
                  v-if="version.source === 'uploaded'"
                  class="delete-version-btn"
                  @click.stop="handleDeleteUploaded(version)"
                  title="删除此上传版本"
                >
                  <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <polyline points="3 6 5 6 21 6" />
                    <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
                  </svg>
                </button>
              </div>
            </div>
            
            <div v-if="selectedVersion?.id === version.id" class="version-detail">
              <div class="detail-header">
                <span>版本详情</span>
                <button class="compare-btn" @click.stop="showCompare = true">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <line x1="18" y1="20" x2="18" y2="10" />
                    <line x1="12" y1="20" x2="12" y2="4" />
                    <line x1="6" y1="20" x2="6" y2="14" />
                  </svg>
                  <span>版本对比</span>
                </button>
              </div>

              <div class="feature-list">
                <div
                  v-for="(feature, index) in version.features"
                  :key="index"
                  class="feature-item"
                  :class="getFeatureTypeClass(feature.type)"
                >
                  <span class="feature-type">{{ getFeatureTypeLabel(feature.type) }}</span>
                  <span class="feature-text">{{ feature.text }}</span>
                </div>
              </div>
            </div>
            
            <button 
              v-if="selectedVersion && selectedVersion.id !== version.id"
              class="compare-action"
              @click="handleCompareVersion(version)"
            >
              与此版本对比
            </button>
          </div>
        </div>

        <!-- 版本对比弹窗 -->
        <div v-if="showCompare && compareVersion" class="compare-modal">
          <div class="compare-content">
            <div class="compare-header">
              <span>版本对比</span>
              <button class="close-btn" @click="closeCompare">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <line x1="18" y1="6" x2="6" y2="18" />
                  <line x1="6" y1="6" x2="18" y2="18" />
                </svg>
              </button>
            </div>
            
            <div class="compare-versions">
              <div class="compare-side">
                <div class="side-header">
                  <span class="side-version">{{ selectedVersion.id }}</span>
                  <span class="side-date">{{ selectedVersion.date }}</span>
                </div>
                <div class="side-title">{{ selectedVersion.title }}</div>
              </div>
              
              <div class="compare-arrow">
                <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M5 12h14" />
                  <polyline points="12 5 19 12 12 19" />
                </svg>
              </div>
              
              <div class="compare-side">
                <div class="side-header">
                  <span class="side-version">{{ compareVersion.id }}</span>
                  <span class="side-date">{{ compareVersion.date }}</span>
                </div>
                <div class="side-title">{{ compareVersion.title }}</div>
              </div>
            </div>
            
            <div class="compare-result">
              <div class="result-section">
                <div class="result-title">
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#52c41a" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <polyline points="20 6 9 17 4 12" />
                  </svg>
                  新增功能
                </div>
                <div class="result-list">
                  <div v-for="(f, i) in compareFeatures.added" :key="'add-'+i" class="result-item added">
                    {{ f.text }}
                  </div>
                  <div v-if="compareFeatures.added.length === 0" class="result-empty">无新增功能</div>
                </div>
              </div>
              
              <div class="result-section">
                <div class="result-title">
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#ff4d4f" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <line x1="18" y1="6" x2="6" y2="18" />
                    <line x1="6" y1="6" x2="18" y2="18" />
                  </svg>
                  移除功能
                </div>
                <div class="result-list">
                  <div v-for="(f, i) in compareFeatures.removed" :key="'remove-'+i" class="result-item removed">
                    {{ f.text }}
                  </div>
                  <div v-if="compareFeatures.removed.length === 0" class="result-empty">无移除功能</div>
                </div>
              </div>
              
              <div class="result-section">
                <div class="result-title">
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#4080ff" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <circle cx="12" cy="12" r="10" />
                    <line x1="12" y1="8" x2="12" y2="12" />
                    <line x1="12" y1="16" x2="12.01" y2="16" />
                  </svg>
                  共有功能
                </div>
                <div class="result-list">
                  <div v-for="(f, i) in compareFeatures.common" :key="'common-'+i" class="result-item common">
                    {{ f.text }}
                  </div>
                  <div v-if="compareFeatures.common.length === 0" class="result-empty">无共有功能</div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 新增版本对话框 -->
        <div v-if="showAddVersionDialog" class="add-version-modal" @click.self="showAddVersionDialog = false">
          <div class="add-version-dialog">
            <div class="dialog-header">
              <h3>新增版本</h3>
              <button class="close-btn" @click="showAddVersionDialog = false">×</button>
            </div>

            <div class="dialog-body">
              <!-- 区域选择 -->
              <div class="form-row">
                <label class="form-label">所属区域</label>
                <div class="region-radio">
                  <label class="radio-item">
                    <input
                      type="radio"
                      v-model="versionForm.region"
                      value="domestic"
                    />
                    <span>国内版本</span>
                  </label>
                  <label class="radio-item">
                    <input
                      type="radio"
                      v-model="versionForm.region"
                      value="overseas"
                    />
                    <span>海外版本</span>
                  </label>
                </div>
              </div>

              <!-- 版本号 -->
              <div class="form-row">
                <label class="form-label required">版本号</label>
                <input
                  type="text"
                  v-model="versionForm.id"
                  class="form-input"
                  placeholder="例如：v3.2.0、L.1.2.11T2"
                />
              </div>

              <!-- 发布日期 -->
              <div class="form-row">
                <label class="form-label required">发布日期</label>
                <input
                  type="date"
                  v-model="versionForm.date"
                  class="form-input"
                />
              </div>

              <!-- APP版本（软件版本） -->
              <div class="form-row">
                <label class="form-label">软件版本（APP版本）</label>
                <input
                  type="text"
                  v-model="versionForm.appVersion"
                  @input="onAppVersionChange(versionForm.appVersion)"
                  class="form-input"
                  placeholder="例如：L.1.2.11T2"
                />
              </div>

              <!-- 控制器版本 -->
              <div class="form-row">
                <label class="form-label">控制器版本</label>
                <input
                  type="text"
                  v-model="versionForm.controllerVersion"
                  class="form-input"
                  placeholder="例如：8.5-B"
                />
              </div>

              <!-- 使用设备 -->
              <div class="form-row">
                <label class="form-label">使用设备</label>
                <input
                  type="text"
                  v-model="versionForm.usedDevice"
                  class="form-input"
                  placeholder="留空则根据软件版本自动判断"
                />
              </div>

              <!-- 测试项 -->
              <div class="form-row">
                <label class="form-label required">测试项</label>
                <div class="feature-input-row">
                  <select v-model="tempFeature.type" class="form-select-small">
                    <option value="new">新增</option>
                    <option value="improve">优化</option>
                    <option value="fix">修复</option>
                  </select>
                  <input
                    type="text"
                    v-model="tempFeature.text"
                    class="form-input"
                    placeholder="测试项内容，回车快速添加"
                    @keyup.enter="addFeatureToForm"
                  />
                  <button class="add-feature-btn" @click="addFeatureToForm">
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                      <line x1="12" y1="5" x2="12" y2="19" />
                      <line x1="5" y1="12" x2="19" y2="12" />
                    </svg>
                    <span>添加</span>
                  </button>
                </div>

                <div v-if="versionForm.features.length > 0" class="feature-list-form">
                  <div
                    v-for="(f, i) in versionForm.features"
                    :key="i"
                    class="feature-list-item"
                    :class="'type-' + f.type"
                  >
                    <span class="feature-type-tag">{{
                      f.type === 'new' ? '新增' :
                      f.type === 'improve' ? '优化' :
                      f.type === 'fix' ? '修复' : f.type
                    }}</span>
                    <span class="feature-text-form">{{ f.text }}</span>
                    <button class="remove-feature-btn" @click="removeFeatureFromForm(i)">×</button>
                  </div>
                </div>
                <div v-else class="feature-empty-hint">还未添加测试项</div>
              </div>
            </div>

            <div class="dialog-footer">
              <button class="btn-cancel" @click="showAddVersionDialog = false">取消</button>
              <button class="btn-save" @click="saveNewVersion">保存</button>
            </div>
          </div>
        </div>
      </div>
    </GlassCard>
  </div>
</template>

<style scoped>
.version-manual {
  padding: 20px;
}

.version-container {
  max-width: 900px;
  margin: 0 auto;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
  margin-bottom: 16px;
}

.upload-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 18px;
  border: none;
  border-radius: 8px;
  background: linear-gradient(135deg, #4080ff 0%, #165DFF 100%);
  color: white;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s;
  box-shadow: 0 2px 6px rgba(64, 128, 255, 0.3);
}

.upload-btn:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(64, 128, 255, 0.4);
}

.upload-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.upload-btn .spinning {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.upload-hint {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 10px 14px;
  background: #f0f7ff;
  border: 1px solid #d0e4ff;
  border-radius: 8px;
  color: #4080ff;
  font-size: 13px;
  line-height: 1.5;
  margin-bottom: 20px;
}

.upload-hint svg {
  flex-shrink: 0;
  margin-top: 2px;
}

.upload-hint code {
  background: white;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 12px;
  border: 1px solid #d0e4ff;
  color: #165DFF;
  font-family: 'Menlo', 'Monaco', monospace;
}

.version-app {
  font-size: 12px;
  padding: 2px 8px;
  background: #f0f7ff;
  color: #165DFF;
  border-radius: 4px;
  border: 1px solid #d0e4ff;
  font-weight: 500;
}

.version-controller {
  font-size: 12px;
  padding: 2px 8px;
  background: #fff7e6;
  color: #d4880c;
  border-radius: 4px;
  border: 1px solid #ffd591;
}

.version-device {
  display: inline-flex;
  align-items: center;
  font-size: 12px;
  padding: 2px 8px;
  background: #f0f5ff;
  color: #165DFF;
  border: 1px solid #adc8ff;
  border-radius: 4px;
  font-weight: 500;
}

.version-device svg {
  flex-shrink: 0;
}

.uploaded-badge {
  font-size: 11px;
  padding: 2px 8px;
  background: #f6ffed;
  color: #52c41a;
  border: 1px solid #b7eb8f;
  border-radius: 4px;
  font-weight: 500;
}

.delete-version-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border: none;
  border-radius: 4px;
  background: #fff2f0;
  color: #ff4d4f;
  cursor: pointer;
  margin-left: auto;
  transition: all 0.2s;
}

.delete-version-btn:hover {
  background: #ff4d4f;
  color: white;
}

.region-selector {
  display: flex;
  gap: 12px;
}

.region-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 24px;
  border: 1px solid #ddd;
  border-radius: 8px;
  background: white;
  cursor: pointer;
  transition: all 0.2s;
  font-size: 14px;
  font-weight: 500;
}

.region-btn:hover {
  border-color: #4080ff;
}

.region-btn.active {
  background: #4080ff;
  color: white;
  border-color: #4080ff;
}

.version-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.version-item {
  background: white;
  border-radius: 12px;
  border: 1px solid #eee;
  transition: all 0.3s;
}

.version-item.selected {
  border-color: #4080ff;
  box-shadow: 0 4px 12px rgba(64, 128, 255, 0.15);
}

.version-header {
  padding: 16px;
  cursor: pointer;
}

.version-info {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
}

.version-id {
  font-size: 18px;
  font-weight: 600;
  color: #4080ff;
}

.version-date {
  font-size: 13px;
  color: #999;
}

.version-title {
  font-size: 15px;
  color: #333;
  margin-bottom: 8px;
}

.version-count {
  display: flex;
  gap: 12px;
}

.count-new, .count-improve, .count-fix {
  font-size: 12px;
  padding: 4px 8px;
  border-radius: 4px;
}

.count-new {
  background: #e8f8e8;
  color: #52c41a;
}

.count-improve {
  background: #e8f4f8;
  color: #4080ff;
}

.count-fix {
  background: #fff2f0;
  color: #ff4d4f;
}

.version-detail {
  padding: 16px;
  background: #fafafa;
  border-top: 1px solid #eee;
}

.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  font-weight: 600;
  color: #333;
}

.compare-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border: 1px solid #ddd;
  border-radius: 6px;
  background: white;
  cursor: pointer;
  font-size: 13px;
}

.compare-btn:hover {
  border-color: #4080ff;
  color: #4080ff;
}

.feature-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.feature-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 10px 12px;
  background: white;
  border-radius: 6px;
}

.feature-type {
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 4px;
  white-space: nowrap;
}

.feature-new .feature-type {
  background: #e8f8e8;
  color: #52c41a;
}

.feature-improve .feature-type {
  background: #e8f4f8;
  color: #4080ff;
}

.feature-fix .feature-type {
  background: #fff2f0;
  color: #ff4d4f;
}

.feature-remove .feature-type {
  background: #f5f5f5;
  color: #999;
}

.feature-text {
  font-size: 14px;
  color: #333;
}

.compare-action {
  margin: 12px 16px;
  padding: 8px 16px;
  border: 1px solid #4080ff;
  border-radius: 6px;
  background: white;
  color: #4080ff;
  cursor: pointer;
  font-size: 13px;
}

.compare-action:hover {
  background: #4080ff;
  color: white;
}

.compare-modal {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.compare-content {
  width: 90%;
  max-width: 800px;
  background: white;
  border-radius: 16px;
  padding: 24px;
  max-height: 80vh;
  overflow-y: auto;
}

.compare-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  font-size: 18px;
  font-weight: 600;
}

.close-btn {
  border: none;
  background: none;
  cursor: pointer;
  color: #999;
}

.close-btn:hover {
  color: #333;
}

.compare-versions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  margin-bottom: 24px;
}

.compare-side {
  flex: 1;
  padding: 16px;
  background: #f5f5f5;
  border-radius: 8px;
}

.side-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
}

.side-version {
  font-weight: 600;
  color: #4080ff;
}

.side-date {
  color: #999;
  font-size: 13px;
}

.side-title {
  font-size: 14px;
  color: #333;
}

.compare-arrow {
  color: #4080ff;
}

.compare-result {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.result-section {
  background: #fafafa;
  border-radius: 8px;
  padding: 16px;
}

.result-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  margin-bottom: 12px;
}

.result-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.result-item {
  padding: 8px 12px;
  background: white;
  border-radius: 4px;
  font-size: 14px;
}

.result-item.added {
  border-left: 3px solid #52c41a;
}

.result-item.removed {
  border-left: 3px solid #ff4d4f;
}

.result-item.common {
  border-left: 3px solid #4080ff;
}

.result-empty {
  color: #999;
  font-size: 13px;
  text-align: center;
  padding: 8px;
}

/* ==================== 新增版本对话框 ==================== */
.add-version-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background: linear-gradient(135deg, #165DFF 0%, #4080FF 100%);
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  box-shadow: 0 2px 6px rgba(22, 93, 255, 0.25);
}

.add-version-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(22, 93, 255, 0.35);
}

.add-version-btn:active {
  transform: translateY(0);
}

.add-version-modal {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  backdrop-filter: blur(2px);
}

.add-version-dialog {
  background: white;
  border-radius: 12px;
  width: 560px;
  max-width: 90vw;
  max-height: 85vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.2);
  animation: dialog-in 0.2s ease-out;
}

@keyframes dialog-in {
  from {
    opacity: 0;
    transform: scale(0.95) translateY(-10px);
  }
  to {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}

.dialog-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid #f0f0f0;
}

.dialog-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #1d2129;
}

.close-btn {
  background: none;
  border: none;
  font-size: 24px;
  line-height: 1;
  color: #86909c;
  cursor: pointer;
  padding: 0;
  width: 28px;
  height: 28px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.close-btn:hover {
  background: #f5f5f5;
  color: #1d2129;
}

.dialog-body {
  padding: 20px;
  overflow-y: auto;
  flex: 1;
}

.form-row {
  margin-bottom: 16px;
}

.form-label {
  display: block;
  font-size: 13px;
  color: #4e5969;
  margin-bottom: 6px;
  font-weight: 500;
}

.form-label.required::before {
  content: '*';
  color: #f53f3f;
  margin-right: 4px;
}

.form-input,
.form-select-small {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #dcdfe6;
  border-radius: 6px;
  font-size: 14px;
  color: #1d2129;
  transition: border-color 0.2s;
  box-sizing: border-box;
}

.form-input:focus,
.form-select-small:focus {
  outline: none;
  border-color: #165DFF;
  box-shadow: 0 0 0 3px rgba(22, 93, 255, 0.1);
}

.region-radio {
  display: flex;
  gap: 20px;
  padding: 6px 0;
}

.radio-item {
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
  font-size: 14px;
  color: #1d2129;
}

.radio-item input[type="radio"] {
  cursor: pointer;
  accent-color: #165DFF;
}

.feature-input-row {
  display: flex;
  gap: 8px;
  align-items: center;
}

.feature-input-row .form-select-small {
  width: 80px;
  flex-shrink: 0;
}

.feature-input-row .form-input {
  flex: 1;
}

.add-feature-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 8px 12px;
  background: #f2f3f5;
  color: #1d2129;
  border: none;
  border-radius: 6px;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
  flex-shrink: 0;
}

.add-feature-btn:hover {
  background: #e5e6eb;
}

.feature-list-form {
  margin-top: 10px;
  display: flex;
  flex-direction: column;
  gap: 6px;
  max-height: 200px;
  overflow-y: auto;
  padding: 4px;
  background: #fafbfc;
  border-radius: 6px;
  border: 1px solid #f0f0f0;
}

.feature-list-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 10px;
  background: white;
  border-radius: 4px;
  font-size: 13px;
  color: #1d2129;
}

.feature-type-tag {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 10px;
  font-size: 11px;
  font-weight: 500;
  flex-shrink: 0;
}

.type-new .feature-type-tag {
  background: #e8f3ff;
  color: #165DFF;
}

.type-improve .feature-type-tag {
  background: #e8ffea;
  color: #00b42a;
}

.type-fix .feature-type-tag {
  background: #fff3e8;
  color: #ff7d00;
}

.feature-text-form {
  flex: 1;
  word-break: break-all;
}

.remove-feature-btn {
  background: none;
  border: none;
  color: #86909c;
  font-size: 18px;
  line-height: 1;
  cursor: pointer;
  padding: 0;
  width: 22px;
  height: 22px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: all 0.2s;
}

.remove-feature-btn:hover {
  background: #ffe7e7;
  color: #f53f3f;
}

.feature-empty-hint {
  margin-top: 10px;
  padding: 16px;
  text-align: center;
  color: #86909c;
  font-size: 13px;
  background: #fafbfc;
  border-radius: 6px;
  border: 1px dashed #e5e6eb;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 12px 20px;
  border-top: 1px solid #f0f0f0;
  background: #fafbfc;
  border-radius: 0 0 12px 12px;
}

.btn-cancel,
.btn-save {
  padding: 8px 20px;
  border-radius: 6px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
  border: 1px solid transparent;
}

.btn-cancel {
  background: white;
  color: #4e5969;
  border-color: #dcdfe6;
}

.btn-cancel:hover {
  border-color: #165DFF;
  color: #165DFF;
}

.btn-save {
  background: #165DFF;
  color: white;
}

.btn-save:hover {
  background: #0e42d2;
}
</style>