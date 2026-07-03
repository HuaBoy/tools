<script setup>
import { ref, computed } from 'vue';
import { ElMessage } from 'element-plus';
import GlassCard from '@/components/GlassCard.vue';
import FileUploader from '@/components/FileUploader.vue';
import BaseChart from '@/components/BaseChart.vue';
import { useLogsStore } from '@/stores/logs';
import CryptoJS from 'crypto-js';

const SECRET_KEY = '123456abc2345678';
const CIPHER_PATTERN = /##@@##([\s\S]*?)@@##@@/g;

const decryptLog = (content) => {
  return content.replace(CIPHER_PATTERN, (match, cipherText) => {
    const cleanedCipher = cipherText.replace(/\s/g, '');
    try {
      const key = CryptoJS.enc.Utf8.parse(SECRET_KEY);
      const decrypted = CryptoJS.AES.decrypt(cleanedCipher, key, {
        mode: CryptoJS.mode.ECB,
        padding: CryptoJS.pad.Pkcs7
      });
      return decrypted.toString(CryptoJS.enc.Utf8) || match;
    } catch (e) {
      return match;
    }
  });
};

const logsStore = useLogsStore();
const fileList = ref([]);
const isAnalyzing = ref(false);
const showAnalysis = ref(false);
const analysisProgress = ref(0);

const currentDate = new Date();
const startDate = ref('');
const endDate = ref('');

const analysisType = ref('all');

const parsedData = ref({
  currentDetection: [],
  detonatorDetection: [],
  detonationCommand: []
});

const faultDetails = ref([]);

const currentStats = computed(() => {
  const data = parsedData.value.currentDetection;
  return {
    total: data.length,
    abnormal: data.filter(d => d.isAbnormal).length,
    avgVoltage: data.length ? (data.reduce((sum, d) => sum + parseFloat(d.busVoltage || 0), 0) / data.length).toFixed(2) : 0,
    avgCurrent: data.length ? (data.reduce((sum, d) => sum + parseFloat(d.busCurrent || 0), 0) / data.length).toFixed(2) : 0
  };
});

const detonatorStats = computed(() => {
  const data = parsedData.value.detonatorDetection;
  return {
    total: data.length,
    success: data.filter(d => d.status === '正常').length,
    failed: data.filter(d => d.status !== '正常').length
  };
});

const parseFileName = (fileName) => {
  const parts = fileName.split('_');
  return {
    date: parts[0] || '',
    deviceCode: parts[1] || '',
    tenantCode: parts[2] ? parts[2].replace(/^0+/, '') : '',
    companyName: parts[3] || '',
    controllerCode: parts[4] || ''
  };
};

const parseLogContent = (content) => {
  const lines = content.split('\n');
  const currentDetection = [];
  const detonatorDetection = [];
  const detonationCommand = [];

  lines.forEach(line => {
    if (line.includes('网络授时') || line.includes('低压检测') || line.includes('实时检测电流电压')) {
      const timestampMatch = line.match(/(\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2})/);
      const voltageMatch = line.match(/总线电压[：:]\s*([\d.]+)\s*V/);
      const currentMatch = line.match(/总线电流[：:]\s*([\d.]+)\s*uA/);
      const statusMatch = line.match(/网络状态[：:]\s*(\S+)/);

      if (timestampMatch) {
        const voltage = voltageMatch ? parseFloat(voltageMatch[1]) : 0;
        const current = currentMatch ? parseFloat(currentMatch[1]) : 0;
        currentDetection.push({
          timestamp: timestampMatch[1],
          busVoltage: voltageMatch ? voltageMatch[1] : '0',
          busCurrent: currentMatch ? currentMatch[1] : '0',
          currentStatus: statusMatch ? statusMatch[1] : '',
          isAbnormal: voltage < 20 || current > 1000
        });
      }
    } else if (line.includes('[单发检测]') || line.includes('[连发检测]')) {
      const timestampMatch = line.match(/(\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2})/);
      const uidMatch = line.match(/UID[：:]\s*(\S+)/);
      const shellCodeMatch = line.match(/管壳码[：:]\s*(\S+)/);
      const currentMatch = line.match(/电流[：:]\s*([\d.]+)/);
      const voltageMatch = line.match(/电压[：:]\s*([\d.]+)/);
      const statusMatch = line.match(/状态[：:]\s*(\S+)/);

      detonatorDetection.push({
        timestamp: timestampMatch ? timestampMatch[1] : '',
        type: line.includes('[单发检测]') ? '单发检测' : '连发检测',
        uid: uidMatch ? uidMatch[1] : '',
        shellCode: shellCodeMatch ? shellCodeMatch[1] : '',
        current: currentMatch ? currentMatch[1] : '',
        voltage: voltageMatch ? voltageMatch[1] : '',
        status: statusMatch ? statusMatch[1] : '未知'
      });
    } else if (line.includes('起爆指令') || (line.includes('执行内容') && line.includes('命令'))) {
      const timestampMatch = line.match(/(\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2})/);
      const contentMatch = line.match(/执行内容[：:]\s*(.+?)(?=命令|$)/i);
      const commandMatch = line.match(/命令[：:]\s*(.+)/i);

      detonationCommand.push({
        timestamp: timestampMatch ? timestampMatch[1] : '',
        content: contentMatch ? contentMatch[1].trim() : '',
        command: commandMatch ? commandMatch[1].trim() : ''
      });
    }
  });

  return { currentDetection, detonatorDetection, detonationCommand };
};

const handleFilesSelected = (files) => {
  const newFiles = files.map(file => ({
    id: Date.now() + Math.random(),
    name: file.name,
    size: (file.size / 1024).toFixed(2) + ' KB',
    status: 'pending',
    file: file,
    parsedInfo: null,
    rawContent: ''
  }));
  fileList.value = [...fileList.value, ...newFiles];
  logsStore.addLog('上传', '起爆器日志AI分析', `上传文件: ${files.map(f => f.name).join(', ')}`);
};

const readFile = (file) => {
  return new Promise((resolve) => {
    const reader = new FileReader();
    reader.onload = (e) => resolve(e.target.result);
    reader.readAsText(file, 'UTF-8');
  });
};

const handleAnalysis = async () => {
  const pendingFiles = fileList.value.filter(f => f.status === 'pending');
  if (pendingFiles.length === 0) {
    ElMessage.warning('请先上传日志文件');
    return;
  }

  isAnalyzing.value = true;
  showAnalysis.value = false;
  analysisProgress.value = 0;

  parsedData.value = {
    currentDetection: [],
    detonatorDetection: [],
    detonationCommand: []
  };

  for (let i = 0; i < pendingFiles.length; i++) {
    const file = pendingFiles[i];
    file.status = 'analyzing';

    try {
        const rawContent = await readFile(file.file);
        file.rawContent = rawContent;
        file.parsedInfo = parseFileName(file.name);
        
        const decryptedContent = decryptLog(rawContent);
        const parsed = parseLogContent(decryptedContent);
        parsedData.value.currentDetection = [...parsedData.value.currentDetection, ...parsed.currentDetection];
        parsedData.value.detonatorDetection = [...parsedData.value.detonatorDetection, ...parsed.detonatorDetection];
        parsedData.value.detonationCommand = [...parsedData.value.detonationCommand, ...parsed.detonationCommand];
        
        file.status = 'completed';
    } catch (e) {
      file.status = 'failed';
    }

    analysisProgress.value = Math.round(((i + 1) / pendingFiles.length) * 100);
  }

  generateFaultDetails();
  isAnalyzing.value = false;
  showAnalysis.value = true;
  logsStore.addLog('分析', '起爆器日志AI分析', `完成AI分析，共处理 ${pendingFiles.length} 个文件`);
  ElMessage.success('AI解析完成');
};

const generateFaultDetails = () => {
  const faults = [];
  let id = 1;

  parsedData.value.currentDetection.forEach(item => {
    if (item.isAbnormal) {
      faults.push({
        id: id++,
        deviceId: fileList.value[0]?.parsedInfo?.deviceCode || '未知',
        timestamp: item.timestamp,
        faultType: '电流异常',
        level: parseFloat(item.busVoltage) < 15 || parseFloat(item.busCurrent) > 1500 ? '高' : '中',
        detail: `总线电压${item.busVoltage}V，电流${item.busCurrent}uA${parseFloat(item.busVoltage) < 20 ? '（电压过低）' : '（电流过高）'}`,
        solution: parseFloat(item.busVoltage) < 20 
          ? '检查供电系统，确认电压稳定，必要时更换电源模块' 
          : '检查网络负载，优化通信参数，降低总线负载'
      });
    }
  });

  parsedData.value.detonatorDetection.forEach(item => {
    if (item.status !== '正常') {
      faults.push({
        id: id++,
        deviceId: item.shellCode || '未知',
        timestamp: item.timestamp,
        faultType: '雷管检测异常',
        level: '高',
        detail: `${item.type}失败，UID:${item.uid}`,
        solution: '检查雷管连接，确认雷管完好，重新执行检测'
      });
    }
  });

  faultDetails.value = faults;
};

const filterByDate = (data) => {
  if (!startDate.value && !endDate.value) return data;
  
  return data.filter(item => {
    const itemDate = new Date(item.timestamp);
    const start = startDate.value ? new Date(startDate.value) : null;
    const end = endDate.value ? new Date(endDate.value) : null;
    
    if (start && itemDate < start) return false;
    if (end) {
      const endOfDay = new Date(end);
      endOfDay.setDate(endOfDay.getDate() + 1);
      if (itemDate >= endOfDay) return false;
    }
    return true;
  });
};

const filteredCurrentData = computed(() => {
  let data = filterByDate(parsedData.value.currentDetection);
  if (analysisType.value === 'current') data = data.filter(d => d.isAbnormal);
  if (analysisType.value === 'normal') data = data.filter(d => !d.isAbnormal);
  return data;
});

const filteredDetonatorData = computed(() => {
  let data = filterByDate(parsedData.value.detonatorDetection);
  if (analysisType.value === 'current') data = data.filter(d => d.status !== '正常');
  if (analysisType.value === 'normal') data = data.filter(d => d.status === '正常');
  return data;
});

const filteredCommandData = computed(() => {
  return filterByDate(parsedData.value.detonationCommand);
});

const chartOption = computed(() => {
  let data = filteredCurrentData.value;
  
  if (data.length === 0) {
    const now = new Date();
    data = Array.from({ length: 12 }, (_, i) => {
      const time = new Date(now.getTime() - (11 - i) * 60000);
      return {
        timestamp: time.toISOString().replace('T', ' ').substring(0, 19),
        busVoltage: (24 + Math.random() * 2).toFixed(1),
        busCurrent: (500 + Math.random() * 300).toFixed(0),
        currentStatus: '正常',
        isAbnormal: false
      };
    });
  }
  
  const labels = data.map(d => {
    const ts = d.timestamp;
    return ts.includes(' ') ? ts.split(' ')[1] : ts;
  });
  
  const voltageArray = [];
  const currentArray = [];
  
  data.forEach(item => {
    voltageArray.push(parseFloat(item.busVoltage) || 0);
    currentArray.push(parseFloat(item.busCurrent) || 0);
  });

  const calculateAxisRange = (dataArray, defaultMin, defaultMax) => {
    if (dataArray.length === 0) {
      return { min: defaultMin, max: defaultMax };
    }
    
    const dataMin = Math.min(...dataArray);
    const dataMax = Math.max(...dataArray);
    const range = dataMax - dataMin;
    const padding = Math.max(range * 0.05, 0.1);
    
    let min = dataMin - padding;
    let max = dataMax + padding;
    
    return { min: Math.max(0, min), max: max };
  };

  const voltageRange = calculateAxisRange(voltageArray, 20, 30);
  const currentRange = calculateAxisRange(currentArray, 0, 1500);

  return {
    tooltip: {
      trigger: 'axis',
      backgroundColor: 'rgba(15, 23, 42, 0.9)',
      borderColor: 'rgba(22, 93, 255, 0.3)',
      textStyle: { color: '#94A3B8' }
    },
    legend: {
      data: ['总线电压(V)', '总线电流(uA)'],
      textStyle: { color: '#64748B' },
      top: 10,
      left: 'center'
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '18%',
      top: '22%',
      containLabel: true
    },
    dataZoom: [
      {
        type: 'slider',
        show: true,
        xAxisIndex: [0],
        bottom: '5%',
        start: 0,
        end: 100,
        height: 20,
        borderColor: 'rgba(22, 93, 255, 0.3)',
        fillerColor: 'rgba(22, 93, 255, 0.2)',
        handleStyle: { color: '#165DFF' },
        textStyle: { color: '#64748B' },
        dataBackground: {
          lineStyle: { color: 'rgba(22, 93, 255, 0.5)' },
          areaStyle: { color: 'rgba(22, 93, 255, 0.1)' }
        }
      },
      {
        type: 'inside',
        xAxisIndex: [0],
        start: 0,
        end: 100
      }
    ],
    toolbox: {
      feature: {
        dataZoom: {
          yAxisIndex: false,
          title: {
            zoom: '区域缩放',
            back: '还原'
          }
        },
        restore: { title: '还原' }
      },
      right: '3%',
      top: 10,
      iconStyle: { borderColor: '#165DFF' }
    },
    xAxis: {
      type: 'category',
      data: labels,
      axisLine: { lineStyle: { color: 'rgba(100, 116, 139, 0.3)' } },
      axisLabel: { color: '#64748B', rotate: 45 }
    },
    yAxis: [
      {
        type: 'value',
        name: '电压(V)',
        axisLine: { lineStyle: { color: 'rgba(100, 116, 139, 0.3)' } },
        axisLabel: { color: '#64748B' },
        splitLine: { lineStyle: { color: 'rgba(100, 116, 139, 0.1)' } },
        min: voltageRange.min,
        max: voltageRange.max
      },
      {
        type: 'value',
        name: '电流(uA)',
        axisLine: { lineStyle: { color: 'rgba(100, 116, 139, 0.3)' } },
        axisLabel: { color: '#64748B' },
        splitLine: { show: false },
        min: currentRange.min,
        max: currentRange.max
      }
    ],
    series: [
      {
        name: '总线电压(V)',
        type: 'line',
        data: voltageArray,
        itemStyle: { color: '#165DFF' },
        smooth: true,
        connectNulls: true
      },
      {
        name: '总线电流(uA)',
        type: 'line',
        yAxisIndex: 1,
        data: currentArray,
        itemStyle: { color: '#00B42A' },
        smooth: true,
        connectNulls: true
      }
    ]
  };
});

const exportToCSV = (type) => {
  let data, headers, filename;
  
  switch (type) {
    case 'current':
      data = filteredCurrentData.value;
      headers = ['时间戳', '总线电压(V)', '总线电流(uA)', '电流状态', '是否异常'];
      filename = '电流检测数据.csv';
      break;
    case 'detonator':
      data = filteredDetonatorData.value;
      headers = ['执行时间', '检测类型', 'UID', '管壳码', '电流', '电压', '雷管状态'];
      filename = '雷管检测数据.csv';
      break;
    case 'command':
      data = filteredCommandData.value;
      headers = ['发生时间', '执行内容', '命令'];
      filename = '起爆指令数据.csv';
      break;
    case 'fault':
      data = faultDetails.value;
      headers = ['设备编号', '发生时间', '故障类型', '风险等级', '故障详情', 'AI修复方案'];
      filename = '故障明细.csv';
      break;
    case 'all':
      const allContent = [
        ['起爆器日志AI分析数据导出'],
        [`导出时间: ${new Date().toLocaleString()}`],
        [''],
        ['电流检测数据'],
        ['时间戳', '总线电压(V)', '总线电流(uA)', '电流状态', '是否异常'],
        ...filteredCurrentData.value.map(d => [d.timestamp, d.busVoltage, d.busCurrent, d.currentStatus, d.isAbnormal ? '异常' : '正常']),
        [''],
        ['雷管检测数据'],
        ['执行时间', '检测类型', 'UID', '管壳码', '电流', '电压', '雷管状态'],
        ...filteredDetonatorData.value.map(d => [d.timestamp, d.type, d.uid, d.shellCode, d.current, d.voltage, d.status]),
        [''],
        ['起爆指令数据'],
        ['发生时间', '执行内容', '命令'],
        ...filteredCommandData.value.map(d => [d.timestamp, d.content, d.command]),
        [''],
        ['故障明细'],
        ['设备编号', '发生时间', '故障类型', '风险等级', '故障详情', 'AI修复方案'],
        ...faultDetails.value.map(f => [f.deviceId, f.timestamp, f.faultType, f.level, f.detail, f.solution])
      ];
      downloadCSV(allContent.map(row => row.join(',')).join('\n'), '起爆器日志AI分析_全部数据.csv');
      return;
  }
  
  const content = [headers.join(',')].concat(data.map(item => {
    if (type === 'current') return `${item.timestamp},${item.busVoltage},${item.busCurrent},${item.currentStatus},${item.isAbnormal ? '异常' : '正常'}`;
    if (type === 'detonator') return `${item.timestamp},${item.type},${item.uid},${item.shellCode},${item.current},${item.voltage},${item.status}`;
    return `${item.timestamp},${item.content},${item.command}`;
  })).join('\n');
  
  downloadCSV(content, filename);
};

const downloadCSV = (content, filename) => {
  const blob = new Blob(['\uFEFF' + content], { type: 'text/csv;charset=utf-8;' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  a.click();
  URL.revokeObjectURL(url);
  logsStore.addLog('导出', '起爆器日志AI分析', `导出${filename}`);
  ElMessage.success(`导出成功`);
};

const removeFile = (fileId) => {
  fileList.value = fileList.value.filter(f => f.id !== fileId);
};

const clearAllFiles = () => {
  fileList.value = [];
  showAnalysis.value = false;
  parsedData.value = { currentDetection: [], detonatorDetection: [], detonationCommand: [] };
  faultDetails.value = [];
};

const getLevelClass = (level) => {
  switch (level) {
    case '高': return 'level-high';
    case '中': return 'level-medium';
    case '低': return 'level-low';
    default: return 'level-medium';
  }
};
</script>

<template>
  <div class="log-analysis">
    <GlassCard title="起爆器日志AI分析">
      <FileUploader @files-selected="handleFilesSelected" />
      
      <div class="filter-section">
        <span class="filter-label">分析类型:</span>
        <div class="filter-options">
          <label class="filter-radio">
            <input type="radio" v-model="analysisType" value="all" />
            <span>全部数据</span>
          </label>
          <label class="filter-radio">
            <input type="radio" v-model="analysisType" value="current" />
            <span>仅异常</span>
          </label>
          <label class="filter-radio">
            <input type="radio" v-model="analysisType" value="normal" />
            <span>仅正常</span>
          </label>
        </div>
        <button 
          class="analyze-btn" 
          :class="{ 'pulse-animation': isAnalyzing }"
          :disabled="isAnalyzing || fileList.length === 0"
          @click="handleAnalysis"
        >
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M21 12a9 9 0 0 0-9-9 9.75 9.75 0 0 0-6.74 2.74L3 8" />
            <path d="M3 3v5h5" />
            <path d="M3 16a9 9 0 0 0 9 9 9.75 9.75 0 0 0 6.74-2.74L21 16" />
            <path d="M16 21h5v-5" />
          </svg>
          <span>{{ isAnalyzing ? 'AI解析中...' : '启动AI解析' }}</span>
        </button>
      </div>

      <div v-if="isAnalyzing || analysisProgress > 0" class="progress-area">
        <div class="progress-label">解析进度</div>
        <div class="progress-bar">
          <div class="progress-fill" :style="{ width: analysisProgress + '%' }"></div>
        </div>
        <div class="progress-text">{{ analysisProgress }}%</div>
      </div>
      
      <div v-if="fileList.length > 0" class="file-list">
        <div v-for="file in fileList" :key="file.id" class="file-item">
          <div class="file-icon" :class="file.status">
            <svg v-if="file.status === 'pending'" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
              <polyline points="14 2 14 8 20 8" />
            </svg>
            <svg v-else-if="file.status === 'analyzing'" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="spin">
              <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
              <path d="M9 12l2 2 4-4" />
            </svg>
            <svg v-else-if="file.status === 'failed'" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10" />
              <line x1="15" y1="9" x2="9" y2="15" />
              <line x1="9" y1="9" x2="15" y2="15" />
            </svg>
            <svg v-else width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" />
              <polyline points="22 4 12 14.01 9 11.01" />
            </svg>
          </div>
          <div class="file-info">
            <span class="file-name">{{ file.name }}</span>
            <span class="file-size">{{ file.size }}</span>
            <span v-if="file.parsedInfo" class="file-detail">
              {{ file.parsedInfo.date }} | {{ file.parsedInfo.deviceCode }} | {{ file.parsedInfo.companyName }}
            </span>
          </div>
          <span class="file-status" :class="file.status">
            {{ file.status === 'pending' ? '待分析' : file.status === 'analyzing' ? '分析中' : file.status === 'failed' ? '分析失败' : '已完成' }}
          </span>
          <button class="remove-btn" @click="removeFile(file.id)">×</button>
        </div>
      </div>
    </GlassCard>
    
    <div v-if="showAnalysis" class="analysis-results">
      <GlassCard title="AI分析统计">
        <div class="date-filter">
          <span class="filter-label">日期筛选：</span>
          <input type="date" v-model="startDate" class="date-input" placeholder="开始日期" />
          <span class="filter-separator">至</span>
          <input type="date" v-model="endDate" class="date-input" placeholder="结束日期" />
          <button v-if="startDate || endDate" class="clear-date-btn" @click="startDate = ''; endDate = ''">清除筛选</button>
        </div>
        
        <div class="stats-grid">
          <div class="stat-card">
            <div class="stat-value">{{ currentStats.total }}</div>
            <div class="stat-label">电流检测记录</div>
          </div>
          <div class="stat-card warning">
            <div class="stat-value">{{ currentStats.abnormal }}</div>
            <div class="stat-label">电流异常</div>
          </div>
          <div class="stat-card">
            <div class="stat-value">{{ currentStats.avgVoltage }}V</div>
            <div class="stat-label">平均电压</div>
          </div>
          <div class="stat-card">
            <div class="stat-value">{{ currentStats.avgCurrent }}uA</div>
            <div class="stat-label">平均电流</div>
          </div>
          <div class="stat-card success">
            <div class="stat-value">{{ detonatorStats.success }}</div>
            <div class="stat-label">雷管检测成功</div>
          </div>
          <div class="stat-card danger">
            <div class="stat-value">{{ detonatorStats.failed }}</div>
            <div class="stat-label">雷管检测失败</div>
          </div>
        </div>
      </GlassCard>
      
      <GlassCard title="电压/电流曲线图" style="margin-top: 20px;">
        <div class="chart-wrapper">
          <BaseChart :option="chartOption" height="300px" />
          <button class="export-chart-btn" @click="exportToCSV('all')">导出全部数据</button>
        </div>
      </GlassCard>
      
      <GlassCard title="电流检测数据" style="margin-top: 20px;">
        <div class="section-header">
          <span>共 {{ filteredCurrentData.length }} 条记录</span>
          <button class="export-btn small" @click="exportToCSV('current')">导出CSV</button>
        </div>
        <div class="table-container">
          <table>
            <thead>
              <tr>
                <th>时间戳</th>
                <th>总线电压(V)</th>
                <th>总线电流(uA)</th>
                <th>电流状态</th>
                <th>状态</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(item, index) in filteredCurrentData" :key="index" :class="{ abnormal: item.isAbnormal }">
                <td>{{ item.timestamp }}</td>
                <td>{{ item.busVoltage }}</td>
                <td>{{ item.busCurrent }}</td>
                <td>{{ item.currentStatus }}</td>
                <td><span :class="['status-badge', item.isAbnormal ? 'danger' : 'success']">{{ item.isAbnormal ? '异常' : '正常' }}</span></td>
              </tr>
            </tbody>
          </table>
          <div v-if="filteredCurrentData.length === 0" class="empty-table">暂无数据</div>
        </div>
      </GlassCard>
      
      <GlassCard title="雷管检测数据" style="margin-top: 20px;">
        <div class="section-header">
          <span>共 {{ filteredDetonatorData.length }} 条记录</span>
          <button class="export-btn small" @click="exportToCSV('detonator')">导出CSV</button>
        </div>
        <div class="table-container">
          <table>
            <thead>
              <tr>
                <th>执行时间</th>
                <th>检测类型</th>
                <th>UID</th>
                <th>管壳码</th>
                <th>电流</th>
                <th>电压</th>
                <th>雷管状态</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(item, index) in filteredDetonatorData" :key="index" :class="{ abnormal: item.status !== '正常' }">
                <td>{{ item.timestamp }}</td>
                <td>{{ item.type }}</td>
                <td>{{ item.uid }}</td>
                <td>{{ item.shellCode }}</td>
                <td>{{ item.current }}</td>
                <td>{{ item.voltage }}</td>
                <td><span :class="['status-badge', item.status === '正常' ? 'success' : 'danger']">{{ item.status }}</span></td>
              </tr>
            </tbody>
          </table>
          <div v-if="filteredDetonatorData.length === 0" class="empty-table">暂无数据</div>
        </div>
      </GlassCard>
      
      <GlassCard title="起爆指令数据" style="margin-top: 20px;">
        <div class="section-header">
          <span>共 {{ filteredCommandData.length }} 条记录</span>
          <button class="export-btn small" @click="exportToCSV('command')">导出CSV</button>
        </div>
        <div class="table-container">
          <table>
            <thead>
              <tr>
                <th>发生时间</th>
                <th>执行内容</th>
                <th>命令</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(item, index) in filteredCommandData" :key="index">
                <td>{{ item.timestamp }}</td>
                <td>{{ item.content }}</td>
                <td>{{ item.command }}</td>
              </tr>
            </tbody>
          </table>
          <div v-if="filteredCommandData.length === 0" class="empty-table">暂无数据</div>
        </div>
      </GlassCard>
      
      <GlassCard title="AI识别故障明细" style="margin-top: 20px;">
        <div class="section-header">
          <span>共 {{ faultDetails.length }} 条故障</span>
          <button class="export-btn small" @click="exportToCSV('fault')">导出CSV</button>
        </div>
        <div class="table-container">
          <table>
            <thead>
              <tr>
                <th>设备编号</th>
                <th>发生时间</th>
                <th>故障类型</th>
                <th>风险等级</th>
                <th>故障详情</th>
                <th>AI修复方案</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="detail in faultDetails" :key="detail.id">
                <td>{{ detail.deviceId }}</td>
                <td>{{ detail.timestamp }}</td>
                <td>{{ detail.faultType }}</td>
                <td><span class="level-tag" :class="getLevelClass(detail.level)">{{ detail.level }}</span></td>
                <td>{{ detail.detail }}</td>
                <td class="solution-cell">{{ detail.solution }}</td>
              </tr>
            </tbody>
          </table>
          <div v-if="faultDetails.length === 0" class="empty-table">未检测到故障</div>
        </div>
      </GlassCard>
    </div>
  </div>
</template>

<style scoped>
.log-analysis {
  max-width: 100%;
}

.filter-section {
  display: flex;
  align-items: center;
  gap: 20px;
  margin-top: 16px;
  padding: 16px;
  background: var(--bg-input);
  border-radius: 8px;
  flex-wrap: wrap;
}

.filter-label {
  font-size: 13px;
  color: var(--text-secondary);
}

.filter-options {
  display: flex;
  gap: 16px;
}

.filter-radio {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--text-secondary);
  cursor: pointer;
  
  input[type="radio"] {
    accent-color: #165DFF;
  }
}

.analyze-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 24px;
  background: linear-gradient(135deg, #165DFF 0%, #0F4CD0 100%);
  border: none;
  border-radius: 8px;
  color: #FFFFFF;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s;
  margin-left: auto;
  
  &:hover:not(:disabled) {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(22, 93, 255, 0.3);
  }
  
  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
}

.progress-area {
  margin-top: 16px;
  padding: 16px;
  background: var(--bg-input);
  border-radius: 8px;
}

.progress-label {
  font-size: 13px;
  color: var(--text-secondary);
  margin-bottom: 8px;
}

.progress-bar {
  height: 8px;
  background: rgba(100, 116, 139, 0.3);
  border-radius: 4px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: linear-gradient(90deg, #165DFF, #36CFC9);
  border-radius: 4px;
  transition: width 0.3s ease;
}

.progress-text {
  font-size: 12px;
  color: var(--text-tertiary);
  margin-top: 8px;
  text-align: right;
}

.file-list {
  margin-top: 16px;
}

.file-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background: var(--bg-input);
  border-radius: 8px;
  margin-bottom: 8px;
}

.file-icon {
  flex-shrink: 0;
  
  &.pending {
    color: #FF7D00;
  }
  
  &.analyzing {
    color: #165DFF;
  }
  
  &.completed {
    color: #00B42A;
  }
  
  &.failed {
    color: #F53F3F;
  }
  
  .spin {
    animation: spin 1s linear infinite;
  }
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.file-info {
  flex: 1;
  min-width: 0;
}

.file-name {
  font-size: 13px;
  color: var(--text-primary);
  display: block;
}

.file-size {
  font-size: 12px;
  color: var(--text-tertiary);
}

.file-detail {
  font-size: 11px;
  color: var(--text-tertiary);
  display: block;
  margin-top: 4px;
}

.file-status {
  font-size: 12px;
  padding: 4px 10px;
  border-radius: 4px;
  
  &.pending {
    background: rgba(255, 125, 0, 0.1);
    color: #FF7D00;
  }
  
  &.analyzing {
    background: rgba(22, 93, 255, 0.1);
    color: #165DFF;
  }
  
  &.completed {
    background: rgba(0, 180, 42, 0.1);
    color: #00B42A;
  }

  &.failed {
    background: rgba(245, 63, 63, 0.1);
    color: #F53F3F;
  }
}

.remove-btn {
  background: none;
  border: none;
  font-size: 18px;
  color: var(--text-tertiary);
  cursor: pointer;
  padding: 0 8px;
  
  &:hover {
    color: #F53F3F;
  }
}

.analysis-results {
  margin-top: 20px;
}

.date-filter {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.filter-label {
  font-size: 13px;
  color: var(--text-secondary);
}

.filter-separator {
  color: var(--text-tertiary);
}

.date-input {
  padding: 8px 12px;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  font-size: 13px;
  background: var(--bg-card);
  color: var(--text-primary);
  
  &:focus {
    outline: none;
    border-color: #165DFF;
  }
}

.clear-date-btn {
  padding: 8px 12px;
  background: rgba(245, 63, 63, 0.1);
  border: 1px solid rgba(245, 63, 63, 0.3);
  border-radius: 6px;
  color: #F53F3F;
  font-size: 12px;
  cursor: pointer;
  
  &:hover {
    background: rgba(245, 63, 63, 0.2);
  }
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(6, 1fr);
  gap: 16px;
}

.stat-card {
  background: var(--bg-input);
  padding: 16px;
  border-radius: 8px;
  text-align: center;
  
  &.warning {
    background: rgba(255, 125, 0, 0.1);
  }
  
  &.success {
    background: rgba(0, 180, 42, 0.1);
  }
  
  &.danger {
    background: rgba(245, 63, 63, 0.1);
  }
}

.stat-value {
  font-size: 24px;
  font-weight: 700;
  color: var(--text-primary);
}

.stat-label {
  font-size: 12px;
  color: var(--text-tertiary);
  margin-top: 4px;
}

.chart-wrapper {
  position: relative;
}

.export-chart-btn {
  position: absolute;
  top: 0;
  right: 0;
  padding: 8px 16px;
  background: rgba(0, 180, 42, 0.1);
  border: 1px solid rgba(0, 180, 42, 0.3);
  border-radius: 6px;
  color: #00B42A;
  font-size: 13px;
  cursor: pointer;
  
  &:hover {
    background: rgba(0, 180, 42, 0.2);
  }
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  
  span {
    font-size: 13px;
    color: var(--text-secondary);
  }
}

.export-btn {
  padding: 6px 12px;
  background: rgba(0, 180, 42, 0.1);
  border: 1px solid rgba(0, 180, 42, 0.3);
  border-radius: 4px;
  color: #00B42A;
  font-size: 12px;
  cursor: pointer;
  
  &:hover {
    background: rgba(0, 180, 42, 0.2);
  }
}

.table-container {
  overflow-x: auto;
}

.table-container table {
  width: 100%;
  border-collapse: collapse;
}

.table-container th,
.table-container td {
  padding: 12px;
  text-align: left;
  font-size: 13px;
  border-bottom: 1px solid var(--border-color);
}

.table-container th {
  background: var(--bg-input);
  color: var(--text-tertiary);
  font-weight: 500;
}

.table-container tbody tr:hover {
  background: rgba(22, 93, 255, 0.05);
}

.table-container tbody tr.abnormal {
  background: rgba(245, 63, 63, 0.1);
}

.status-badge {
  font-size: 12px;
  padding: 4px 10px;
  border-radius: 4px;
  
  &.success {
    background: rgba(0, 180, 42, 0.1);
    color: #00B42A;
  }
  
  &.danger {
    background: rgba(245, 63, 63, 0.1);
    color: #F53F3F;
  }
}

.level-tag {
  font-size: 12px;
  padding: 4px 10px;
  border-radius: 4px;
  font-weight: 500;
  
  &.level-high {
    background: rgba(245, 63, 63, 0.1);
    color: #F53F3F;
  }
  
  &.level-medium {
    background: rgba(255, 125, 0, 0.1);
    color: #FF7D00;
  }
  
  &.level-low {
    background: rgba(0, 180, 42, 0.1);
    color: #00B42A;
  }
}

.solution-cell {
  max-width: 300px;
  color: var(--text-secondary);
}

.empty-table {
  text-align: center;
  padding: 40px;
  color: var(--text-tertiary);
}

@media screen and (max-width: 768px) {
  .stats-grid {
    grid-template-columns: repeat(3, 1fr);
  }
  
  .filter-section {
    flex-direction: column;
    align-items: flex-start;
  }
  
  .analyze-btn {
    margin-left: 0;
    width: 100%;
    justify-content: center;
  }
}

@media screen and (max-width: 480px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
  
  .table-container th,
  .table-container td {
    padding: 8px;
    font-size: 12px;
  }
}
</style>