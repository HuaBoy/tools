<script setup>
import { ref, onMounted, onUnmounted } from 'vue';
import GlassCard from '@/components/GlassCard.vue';

const lang = ref('zh');
const iframeLoading = ref(true);
const iframeError = ref(false);
const iframeErrorMsg = ref('');
let loadTimer = null;

// 直接引用盛景可视化平台（外部大屏）地址
const bigScreenUrl = 'https://mp.holyview.cn:9443/bigScreen2/';

let observer = null;

// 中英文翻译映射表（长词优先匹配）
const zhToEnMap = [
  ['数据大屏', 'Data Dashboard'],
  ['生产完成率', 'Completion Rate'],
  ['生产数据', 'Production Data'],
  ['生产进度', 'Production Progress'],
  ['生产统计', 'Production Stats'],
  ['生产计划', 'Production Plan'],
  ['设备状态', 'Device Status'],
  ['合格率', 'Pass Rate'],
  ['良品率', 'Yield Rate'],
  ['不良品', 'Defect'],
  ['开机率', 'Uptime'],
  ['半成品', 'Semi-finished'],
  ['达成率', 'Achievement Rate'],
  ['本季度', 'This Quarter'],
  ['近7天', 'Last 7 Days'],
  ['近30天', 'Last 30 Days'],
  ['暂无数据', 'No Data'],
  ['条记录', 'records'],
  ['加载中', 'Loading'],
  ['成品', 'Finished Product'],
  ['原料', 'Raw Material'],
  ['出库', 'Outbound'],
  ['入库', 'Inbound'],
  ['库存', 'Inventory'],
  ['仓库', 'Warehouse'],
  ['物料', 'Material'],
  ['返工', 'Rework'],
  ['废品', 'Scrap'],
  ['故障', 'Fault'],
  ['报警', 'Alarm'],
  ['维护', 'Maintenance'],
  ['停机', 'Stopped'],
  ['待机', 'Idle'],
  ['运行', 'Running'],
  ['设备', 'Equipment'],
  ['人员', 'Personnel'],
  ['员工', 'Employee'],
  ['出勤', 'Attendance'],
  ['工时', 'Work Hours'],
  ['工单', 'Work Order'],
  ['班组', 'Shift'],
  ['工序', 'Process'],
  ['产线', 'Production Line'],
  ['产能', 'Capacity'],
  ['产量', 'Output'],
  ['生产', 'Production'],
  ['质量', 'Quality'],
  ['合格', 'Pass'],
  ['不合格', 'Fail'],
  ['统计', 'Statistics'],
  ['报表', 'Report'],
  ['趋势', 'Trend'],
  ['占比', 'Ratio'],
  ['排名', 'Rank'],
  ['目标', 'Target'],
  ['实际', 'Actual'],
  ['今日', 'Today'],
  ['昨日', 'Yesterday'],
  ['本周', 'This Week'],
  ['本月', 'This Month'],
  ['本年', 'This Year'],
  ['实时', 'Real-time'],
  ['总计', 'Total'],
  ['万元', '10K CNY'],
  ['摄氏度', '°C'],
  ['百分比', '%'],
  ['上午', 'AM'],
  ['下午', 'PM'],
  ['同比', 'YoY'],
  ['环比', 'MoM'],
  ['元', 'CNY'],
  ['小时', 'h'],
  ['分钟', 'min'],
  ['天', 'd'],
  ['月', 'month'],
  ['年', 'year'],
  ['台', ''],
  ['个', ''],
  ['次', ''],
  ['条', ''],
  ['日', ''],
  ['时', ':'],
  ['分', ':'],
  ['秒', 's'],
  ['０', '0'], ['１', '1'], ['２', '2'], ['３', '3'], ['４', '4'],
  ['５', '5'], ['６', '6'], ['７', '7'], ['８', '8'], ['９', '9'],
];

// 翻译文本节点
const translateTextNode = (textNode) => {
  let text = textNode.textContent;
  let changed = false;
  for (const [zh, en] of zhToEnMap) {
    if (text.includes(zh)) {
      text = text.split(zh).join(en);
      changed = true;
    }
  }
  if (changed) {
    textNode.textContent = text;
  }
};

// 递归翻译 DOM 树中所有文本节点
const translateDOM = (root) => {
  const walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT, null, false);
  const nodes = [];
  let node;
  while ((node = walker.nextNode())) {
    nodes.push(node);
  }
  nodes.forEach(translateTextNode);
};

// 注入翻译到 iframe（同源可访问 contentDocument；跨域外链时跳过）
const injectTranslation = () => {
  try {
    const iframe = document.querySelector('.big-screen-iframe');
    if (!iframe || !iframe.contentDocument || !iframe.contentDocument.body) return;

    const doc = iframe.contentDocument;

    // 先翻译已有内容
    translateDOM(doc.body);

    // 用 MutationObserver 监听后续动态变化
    if (observer) observer.disconnect();
    observer = new MutationObserver((mutations) => {
      mutations.forEach(mutation => {
        mutation.addedNodes.forEach(addedNode => {
          if (addedNode.nodeType === Node.TEXT_NODE) {
            translateTextNode(addedNode);
          } else if (addedNode.nodeType === Node.ELEMENT_NODE) {
            translateDOM(addedNode);
          }
        });
      });
    });
    observer.observe(doc.body, { childList: true, subtree: true, characterData: true });
  } catch (e) {
    // 跨域 iframe 无法访问 contentDocument，翻译注入跳过（不影响大屏显示）
  }
};

// 切换语言
const switchLang = (newLang) => {
  lang.value = newLang;
  iframeLoading.value = true;
  iframeError.value = false;

  if (observer) {
    observer.disconnect();
    observer = null;
  }

  startLoadTimeout();

  if (newLang === 'en') {
    // 等待 iframe 重新加载后再注入
    setTimeout(() => {
      injectTranslation();
      iframeLoading.value = false;
    }, 800);
  } else {
    // 中文模式：重新加载 iframe 还原原始内容
    const iframe = document.querySelector('.big-screen-iframe');
    if (iframe) {
      iframe.src = iframe.src; // 触发重新加载
    }
    setTimeout(() => {
      iframeLoading.value = false;
    }, 500);
  }
};

// iframe 加载完成
const onIframeLoad = () => {
  if (loadTimer) { clearTimeout(loadTimer); loadTimer = null; }
  iframeLoading.value = false;
  iframeError.value = false;
  if (lang.value === 'en') {
    setTimeout(() => injectTranslation(), 300);
  }
};

// iframe 加载失败
const onIframeError = () => {
  if (loadTimer) { clearTimeout(loadTimer); loadTimer = null; }
  iframeLoading.value = false;
  iframeError.value = true;
  iframeErrorMsg.value = lang.value === 'zh'
    ? '数据大屏加载失败，请检查网络连接或联系管理员'
    : 'Failed to load dashboard. Please check network connection or contact administrator.';
};

// 超时兜底（15秒后强制显示）
const startLoadTimeout = () => {
  if (loadTimer) clearTimeout(loadTimer);
  loadTimer = setTimeout(() => {
    iframeLoading.value = false;
    iframeError.value = true;
    iframeErrorMsg.value = lang.value === 'zh'
      ? '数据大屏加载超时，请稍后重试'
      : 'Dashboard loading timeout, please try again later.';
  }, 15000);
};

onMounted(() => {
  startLoadTimeout();
});

onUnmounted(() => {
  if (observer) observer.disconnect();
  if (loadTimer) clearTimeout(loadTimer);
});
</script>

<template>
  <div class="big-screen-page">
    <GlassCard>
      <template #header>
        <div class="page-header">
          <div class="page-title-row">
            <h2 class="page-title">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <rect x="2" y="3" width="20" height="14" rx="2" ry="2" />
                <line x1="8" y1="21" x2="16" y2="21" />
                <line x1="12" y1="17" x2="12" y2="21" />
              </svg>
              {{ lang === 'zh' ? '数据大屏' : 'Data Dashboard' }}
            </h2>
            <div class="lang-toggle">
              <button class="lang-btn" :class="{ active: lang === 'zh' }" @click="switchLang('zh')">中文</button>
              <button class="lang-btn" :class="{ active: lang === 'en' }" @click="switchLang('en')">English</button>
            </div>
          </div>
        </div>
      </template>

      <div class="big-screen-container">
        <div v-if="iframeLoading" class="loading-overlay">
          <div class="loading-spinner">
            <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="#165DFF" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="spin">
              <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
            </svg>
            <p>{{ lang === 'zh' ? '数据大屏加载中...' : 'Loading dashboard...' }}</p>
          </div>
        </div>

        <div v-if="iframeError" class="error-overlay">
          <div class="error-content">
            <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="#F53F3F" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="12" r="10" />
              <line x1="12" y1="8" x2="12" y2="12" />
              <line x1="12" y1="16" x2="12.01" y2="16" />
            </svg>
            <p>{{ iframeErrorMsg }}</p>
            <button class="retry-btn" @click="switchLang(lang)">{{ lang === 'zh' ? '点击重试' : 'Retry' }}</button>
          </div>
        </div>

        <iframe
          :key="lang"
          :src="bigScreenUrl"
          class="big-screen-iframe"
          frameborder="0"
          allowfullscreen
          @load="onIframeLoad"
          @error="onIframeError"
        ></iframe>
      </div>
    </GlassCard>
  </div>
</template>

<style scoped>
.big-screen-page { min-height: 100%; }
.page-header { width: 100%; }
.page-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}
.page-title {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
}
.lang-toggle {
  display: flex;
  gap: 2px;
  background: rgba(22, 93, 255, 0.08);
  border-radius: 8px;
  padding: 3px;
}
.lang-btn {
  padding: 6px 16px;
  border: none;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  color: var(--text-secondary);
  background: transparent;
}
.lang-btn:hover { color: var(--text-primary); background: rgba(22, 93, 255, 0.08); }
.lang-btn.active { background: #165DFF; color: #FFFFFF; box-shadow: 0 1px 4px rgba(22, 93, 255, 0.3); }
.big-screen-container {
  position: relative;
  width: 100%;
  height: calc(100vh - 160px);
  min-height: 600px;
  border-radius: 8px;
  overflow: hidden;
  background: var(--bg-card);
}
.loading-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-card);
  z-index: 10;
}
.loading-spinner {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  color: var(--text-secondary);
  font-size: 14px;
}
.spin { animation: spin 1.5s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
.error-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-card);
  z-index: 10;
}
.error-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  color: var(--text-secondary);
  text-align: center;
  padding: 20px;
}
.error-content p {
  font-size: 14px;
  margin: 0;
  max-width: 400px;
}
.retry-btn {
  padding: 8px 24px;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  background: #165DFF;
  color: #FFFFFF;
  transition: all 0.2s;
}
.retry-btn:hover { background: #0D47A1; transform: translateY(-1px); }
.big-screen-iframe {
  width: 100%;
  height: 100%;
  border: none;
  background: #FFFFFF;
}
@media screen and (max-width: 768px) {
  .big-screen-container { height: calc(100vh - 200px); min-height: 400px; }
  .page-title { font-size: 15px; }
  .lang-btn { padding: 4px 10px; font-size: 12px; }
}
</style>
