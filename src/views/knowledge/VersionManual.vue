<script setup>import { ref, computed } from 'vue';
import { ElMessage } from 'element-plus';
import GlassCard from '@/components/GlassCard.vue';
import { useLogsStore } from '@/stores/logs';
const logsStore = useLogsStore();
const selectedVersion = ref('QB-2024-V1.0');
const searchQuery = ref('');
const isSearching = ref(false);
const versions = [
 { value: 'QB-2024-V1.0', label: 'QB-2024-V1.0 标准版' },
 { value: 'QB-2024-V1.1', label: 'QB-2024-V1.1 增强版' },
 { value: 'QB-2025-V2.0', label: 'QB-2025-V2.0 专业版' },
 { value: 'QB-2025-V2.1', label: 'QB-2025-V2.1 旗舰版' }
];
const manuals = {
 'QB-2024-V1.0': `起爆器版本手册 - QB-2024-V1.0

一、产品概述
本产品为工业起爆器控制系统，用于精确控制爆破作业中的起爆时序和能量输出。

二、技术参数
- 工作电压: 12V DC
- 最大输出电流: 5A
- 起爆通道数: 4通道
- 延时精度: ±1ms
- 工作温度: -20°C ~ +50°C

三、安装说明
1. 将起爆器固定在安全位置
2. 连接电源线和通信线
3. 安装天线并调整位置
4. 连接雷管引线

四、操作指南
1. 开机自检
2. 参数配置
3. 起爆测试
4. 执行起爆

五、故障排除
见知识库故障代码说明`,
 'QB-2024-V1.1': `起爆器版本手册 - QB-2024-V1.1

一、产品概述
增强版在标准版基础上增加了远程控制功能和数据加密传输。

二、技术参数
- 工作电压: 12V DC
- 最大输出电流: 8A
- 起爆通道数: 8通道
- 延时精度: ±0.5ms
- 工作温度: -30°C ~ +60°C
- 通信距离: 5km

三、新增功能
- 远程无线控制
- AES-256数据加密
- 实时状态监控
- 自动故障诊断

四、安装说明
1. 安装主控单元
2. 配置无线模块
3. 设置加密密钥
4. 连接传感器

五、操作指南
详见标准版操作指南`,
 'QB-2025-V2.0': `起爆器版本手册 - QB-2025-V2.0

一、产品概述
专业版集成AI智能分析功能，支持自动参数优化和故障预测。

二、技术参数
- 工作电压: 24V DC
- 最大输出电流: 10A
- 起爆通道数: 16通道
- 延时精度: ±0.1ms
- 工作温度: -40°C ~ +70°C
- 通信距离: 10km

三、AI功能特性
- 智能参数优化
- 故障预测预警
- 数据分析报告
- 远程诊断支持

四、系统要求
- 配套AI分析平台
- 高速网络连接
- 专业培训认证

五、安全注意事项
严格遵守爆破安全规范`,
 'QB-2025-V2.1': `起爆器版本手册 - QB-2025-V2.1

一、产品概述
旗舰版是本系列最高配置产品，支持全自动化起爆和智能决策系统。

二、技术参数
- 工作电压: 24V DC
- 最大输出电流: 15A
- 起爆通道数: 32通道
- 延时精度: ±0.05ms
- 工作温度: -50°C ~ +80°C
- 通信距离: 20km

三、旗舰功能
- 全自动起爆流程
- 智能决策系统
- 多设备协同控制
- 三维可视化监控
- 应急预案自动触发

四、高级配置
详见技术白皮书和专业培训文档

五、维护保养
定期校准和软件更新`
};
const currentManual = computed(() => manuals[selectedVersion.value] || '');
const handleVersionChange = () => {
 logsStore.addLog('切换', '版本手册', `切换版本: ${selectedVersion.value}`);
};
const handleSearch = async () => {
 if (!searchQuery.value.trim()) {
 ElMessage.warning('请输入搜索内容');
 return;
 }
 isSearching.value = true;
 await new Promise(resolve => setTimeout(resolve, 500));
 isSearching.value = false;
 const index = currentManual.value.toLowerCase().indexOf(searchQuery.value.toLowerCase());
 if (index !== -1) {
 ElMessage.success(`找到匹配内容，位于第 ${currentManual.value.substring(0, index).split('\n').length} 行`);
 }
 else {
 ElMessage.info('未找到匹配内容');
 }
 logsStore.addLog('搜索', '版本手册', `搜索: ${searchQuery.value}`);
};
const handleExport = () => {
 const content = currentManual.value;
 const blob = new Blob([content], { type: 'text/plain' });
 const url = URL.createObjectURL(blob);
 const a = document.createElement('a');
 a.href = url;
 a.download = `manual_${selectedVersion.value}.txt`;
 a.click();
 URL.revokeObjectURL(url);
 logsStore.addLog('导出', '版本手册', `导出手册: ${selectedVersion.value}`);
 ElMessage.success('手册导出成功');
};
</script>

<template>
  <div class="version-manual">
    <GlassCard title="起爆器版本手册">
      <div class="toolbar">
        <div class="version-select">
          <label class="select-label">选择版本</label>
          <select 
            v-model="selectedVersion" 
            class="version-dropdown"
            @change="handleVersionChange"
          >
            <option v-for="v in versions" :key="v.value" :value="v.value">
              {{ v.label }}
            </option>
          </select>
        </div>
        
        <div class="search-wrap">
          <input 
            v-model="searchQuery"
            type="text" 
            class="search-input"
            placeholder="全文搜索..."
            @keyup.enter="handleSearch"
          />
          <button 
            class="search-btn" 
            :disabled="isSearching"
            @click="handleSearch"
          >
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="11" cy="11" r="8" />
              <line x1="21" y1="21" x2="16.65" y2="16.65" />
            </svg>
            <span>全文AI检索</span>
          </button>
        </div>
        
        <button class="export-btn" @click="handleExport">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
            <polyline points="7 10 12 15 17 10" />
            <line x1="12" y1="15" x2="12" y2="3" />
          </svg>
          <span>导出PDF手册</span>
        </button>
      </div>
    </GlassCard>
    
    <GlassCard title="手册内容" style="margin-top: 20px;">
      <div class="manual-content">
        <pre>{{ currentManual }}</pre>
      </div>
    </GlassCard>
  </div>
</template>

<style scoped>
.version-manual {
  max-width: 100%;
}

.toolbar {
  display: flex;
  gap: 20px;
  flex-wrap: wrap;
  align-items: center;
}

.version-select {
  display: flex;
  align-items: center;
  gap: 10px;
}

.select-label {
  font-size: 13px;
  color: var(--text-secondary);
}

.version-dropdown {
  background: var(--bg-input);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 10px 16px;
  color: var(--text-primary);
  font-size: 14px;
  outline: none;
  cursor: pointer;
  min-width: 200px;
  
  &:focus {
    border-color: rgba(22, 93, 255, 0.6);
  }
  
  option {
    background: var(--bg-base);
    color: var(--text-primary);
  }
}

.search-wrap {
  flex: 1;
  display: flex;
  gap: 10px;
  min-width: 200px;
}

.search-input {
  flex: 1;
  background: var(--bg-input);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 10px 14px;
  color: var(--text-primary);
  font-size: 14px;
  outline: none;
  
  &:focus {
    border-color: rgba(22, 93, 255, 0.6);
  }
  
  &::placeholder {
    color: var(--text-tertiary);
  }
}

.search-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 20px;
  background: rgba(54, 207, 201, 0.1);
  border: 1px solid rgba(54, 207, 201, 0.3);
  border-radius: 8px;
  color: #36CFC9;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
  
  &:hover:not(:disabled) {
    background: rgba(54, 207, 201, 0.2);
  }
  
  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
}

.export-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 20px;
  background: rgba(0, 180, 42, 0.1);
  border: 1px solid rgba(0, 180, 42, 0.3);
  border-radius: 8px;
  color: #00B42A;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
  
  &:hover {
    background: rgba(0, 180, 42, 0.2);
  }
}

.manual-content {
  max-height: 600px;
  overflow-y: auto;
  padding: 16px;
  background: var(--bg-input);
  border-radius: 8px;
  
  pre {
    font-family: 'Consolas', 'Monaco', monospace;
    font-size: 14px;
    color: var(--text-secondary);
    white-space: pre-wrap;
    word-break: break-all;
    margin: 0;
    line-height: 1.8;
  }
}

@media screen and (max-width: 768px) {
  .toolbar {
    flex-direction: column;
    gap: 16px;
    align-items: stretch;
  }
  
  .version-select {
    flex-direction: column;
    align-items: flex-start;
  }
  
  .version-dropdown {
    width: 100%;
    min-width: 100%;
  }
  
  .search-wrap {
    min-width: 100%;
  }
  
  .search-btn,
  .export-btn {
    justify-content: center;
  }
  
  .manual-content pre {
    font-size: 13px;
  }
}

@media screen and (max-width: 480px) {
  .search-btn,
  .export-btn {
    padding: 8px 16px;
    font-size: 12px;
  }
  
  .version-dropdown {
    padding: 8px 12px;
    font-size: 13px;
  }
  
  .manual-content {
    padding: 12px;
  }
  
  .manual-content pre {
    font-size: 12px;
  }
}
</style>
