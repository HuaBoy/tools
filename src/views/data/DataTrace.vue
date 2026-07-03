<script setup>import { ref, computed } from 'vue';
import { ElMessage } from 'element-plus';
import GlassCard from '@/components/GlassCard.vue';
import BaseChart from '@/components/BaseChart.vue';
import { useLogsStore } from '@/stores/logs';
const logsStore = useLogsStore();
const searchId = ref('');
const isTracing = ref(false);
const traceResult = ref(null);
const traceDetail = ref('');
const chartOption = computed(() => {
 if (!traceResult.value) {
 return {};
 }
 return {
 tooltip: {
 trigger: 'item',
 backgroundColor: 'rgba(15, 23, 42, 0.9)',
 borderColor: 'rgba(22, 93, 255, 0.3)',
 textStyle: {
 color: '#94A3B8'
 }
 },
 series: [{
 type: 'graph',
 layout: 'force',
 data: [
 { 
 name: '授权记录', 
 category: 0,
 symbolSize: 60,
 itemStyle: {
 color: '#165DFF',
 borderColor: '#36CFC9',
 borderWidth: 2
 }
 },
 { 
 name: '加密日志', 
 category: 1,
 symbolSize: 50,
 itemStyle: {
 color: '#36CFC9',
 borderColor: '#165DFF',
 borderWidth: 2
 }
 },
 { 
 name: '起爆任务', 
 category: 2,
 symbolSize: 50,
 itemStyle: {
 color: '#00B42A',
 borderColor: '#165DFF',
 borderWidth: 2
 }
 },
 { 
 name: '故障记录', 
 category: 3,
 symbolSize: 50,
 itemStyle: {
 color: traceResult.value.hasFault ? '#F53F3F' : '#FF7D00',
 borderColor: '#165DFF',
 borderWidth: 2
 }
 }
 ],
 links: [
 { source: '授权记录', target: '加密日志', lineStyle: { color: '#165DFF', width: 2 } },
 { source: '加密日志', target: '起爆任务', lineStyle: { color: '#36CFC9', width: 2 } },
 { source: '起爆任务', target: '故障记录', lineStyle: { color: traceResult.value.hasFault ? '#F53F3F' : '#FF7D00', width: 2 } }
 ],
 categories: [
 { name: '授权' },
 { name: '日志' },
 { name: '任务' },
 { name: '故障' }
 ],
 force: {
 repulsion: 300,
 gravity: 0.1,
 edgeLength: [100, 200]
 }
 }]
 };
});
const handleTrace = async () => {
 if (!searchId.value.trim()) {
 ElMessage.warning('请输入设备ID或任务ID');
 return;
 }
 isTracing.value = true;
 await new Promise(resolve => setTimeout(resolve, 1500));
 traceResult.value = {
 hasFault: searchId.value.includes('ERR'),
 licenseKey: 'LIC2026-QB-' + Math.floor(Math.random() * 1000).toString().padStart(3, '0'),
 logFile: 'LOG_' + new Date().toISOString().split('T')[0] + '.dat',
 taskId: 'TASK-' + Math.floor(Math.random() * 10000).toString().padStart(4, '0'),
 faultCode: searchId.value.includes('ERR') ? 'E003' : '无'
 };
 traceDetail.value = `【设备ID】${searchId.value}\n` +
 `【授权记录】授权码: ${traceResult.value.licenseKey}，授权状态: 有效\n` +
 `【加密日志】文件名: ${traceResult.value.logFile}，创建时间: ${new Date().toLocaleString('zh-CN')}\n` +
 `【起爆任务】任务ID: ${traceResult.value.taskId}，状态: ${searchId.value.includes('ERR') ? '异常' : '正常'}\n` +
 `【故障记录】故障代码: ${traceResult.value.faultCode}\n\n` +
 `全链路追溯完成，共经过 4 个节点。${searchId.value.includes('ERR') ? '检测到异常节点，请关注故障记录。' : '所有节点状态正常。'}`;
 isTracing.value = false;
 logsStore.addLog('追溯', '数据追溯', `设备/任务ID: ${searchId.value}`);
 ElMessage.success('数据追溯完成');
};
</script>

<template>
  <div class="data-trace">
    <GlassCard title="AI数据追溯">
      <div class="search-section">
        <div class="search-input-wrap">
          <input 
            v-model="searchId"
            type="text" 
            class="search-input"
            placeholder="输入设备ID或任务ID"
          />
          <button 
            class="trace-btn" 
            :disabled="isTracing"
            @click="handleTrace"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <polyline points="1 4 1 10 7 10" />
              <path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10" />
            </svg>
            <span>{{ isTracing ? '溯源中...' : '一键溯源' }}</span>
          </button>
        </div>
        <p class="search-hint">示例: DEV-2024-001 或 TASK-001 或 DEV-ERR-001(含故障)</p>
      </div>
    </GlassCard>
    
    <GlassCard title="全链路拓扑图" style="margin-top: 20px;">
      <div v-if="traceResult" class="chart-wrapper">
        <BaseChart :option="chartOption" :loading="isTracing" height="400px" />
      </div>
      <div v-else class="empty-chart">
        <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <line x1="6" y1="3" x2="6" y2="21" />
          <circle cx="18" cy="6" r="3" />
          <circle cx="6" cy="18" r="3" />
          <path d="M18 9a9 9 0 0 1-9 9" />
        </svg>
        <p>输入设备ID或任务ID开始溯源</p>
      </div>
    </GlassCard>
    
    <GlassCard title="溯源链路详情" style="margin-top: 20px;">
      <div v-if="traceDetail" class="detail-content">
        <pre>{{ traceDetail }}</pre>
      </div>
      <div v-else class="empty-detail">
        <p>溯源结果将在此显示...</p>
      </div>
    </GlassCard>
  </div>
</template>

<style scoped>
.data-trace {
  max-width: 100%;
}

.search-section {
  padding: 20px;
  background: var(--bg-input);
  border-radius: 8px;
}

.search-input-wrap {
  display: flex;
  gap: 12px;
}

.search-input {
  flex: 1;
  background: var(--bg-input);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 14px 16px;
  color: var(--text-primary);
  font-size: 14px;
  outline: none;
  transition: border-color 0.2s;
  
  &:focus {
    border-color: rgba(22, 93, 255, 0.6);
  }
  
  &::placeholder {
    color: var(--text-tertiary);
  }
}

.trace-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 14px 32px;
  background: linear-gradient(135deg, #165DFF 0%, #0F4CD0 100%);
  border: none;
  border-radius: 8px;
  color: #FFFFFF;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  
  &:hover:not(:disabled) {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(22, 93, 255, 0.3);
  }
  
  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
}

.search-hint {
  font-size: 12px;
  color: var(--text-tertiary);
  margin-top: 12px;
}

.chart-wrapper {
  width: 100%;
}

.empty-chart {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px;
  color: var(--text-tertiary);
  
  svg {
    margin-bottom: 16px;
  }
}

.detail-content {
  padding: 16px;
  background: var(--bg-input);
  border-radius: 8px;
  
  pre {
    font-family: 'Consolas', 'Monaco', monospace;
    font-size: 13px;
    color: var(--text-secondary);
    white-space: pre-wrap;
    word-break: break-all;
    margin: 0;
    line-height: 1.8;
  }
}

.empty-detail {
  text-align: center;
  padding: 40px;
  color: var(--text-tertiary);
}

@media screen and (max-width: 768px) {
  .search-section {
    padding: 16px;
  }
  
  .search-input-wrap {
    flex-direction: column;
  }
  
  .search-input {
    padding: 12px 14px;
  }
  
  .trace-btn {
    justify-content: center;
    padding: 12px 24px;
  }
  
  .detail-content pre {
    font-size: 12px;
  }
}

@media screen and (max-width: 480px) {
  .trace-btn {
    padding: 10px 20px;
    font-size: 13px;
  }
  
  .search-hint {
    font-size: 11px;
  }
  
  .empty-chart {
    padding: 40px 20px;
  }
}
</style>
