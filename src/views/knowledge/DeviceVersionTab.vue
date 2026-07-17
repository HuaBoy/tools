<script setup>
import { ref, computed, watch } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import * as XLSX from 'xlsx';
import { saveAs } from 'file-saver';
import deviceData from './deviceData.json';

// ==================== 本地持久化（刷新页面不丢失） ====================
const STORAGE_KEY = 'device_version_data_v1';
const loadDevices = () => {
  try {
    const saved = localStorage.getItem(STORAGE_KEY);
    if (saved) {
      const parsed = JSON.parse(saved);
      if (Array.isArray(parsed) && parsed.length) return parsed;
    }
  } catch (e) {}
  return JSON.parse(JSON.stringify(deviceData));
};
const persist = () => {
  try { localStorage.setItem(STORAGE_KEY, JSON.stringify(devices.value)); } catch (e) {}
};

// ==================== 设备分类（优先本地存储，否则内置 deviceData.json） ====================
const devices = ref(loadDevices());
// 任意数据变更自动持久化到浏览器本地存储
watch(devices, () => persist(), { deep: true });

// ==================== 产线选项 ====================
const productionLineOptions = [
  '创者', '瑞巽', '百奥-老', '百奥-新', '杉达激光打码', '汇维',
  '云南喆睿', '大华', '金奥博', '金源恒业', '弘腾', '祥禾', '兆镭凯'
];

// ==================== 拍平为设备台账列表 ====================
let uidSeq = 0;

const allVersions = computed(() => {
  const list = [];
  devices.value.forEach(device => {
    (device.versions || []).forEach(v => {
      list.push({
        uid: v.uid,
        deviceCategory: device.name,
        deviceColor: device.color,
        ...v
      });
    });
  });
  // 按出厂日期倒序
  return list.sort((a, b) => (a.factoryDate < b.factoryDate ? 1 : a.factoryDate > b.factoryDate ? -1 : 0));
});

// ==================== 筛选（按设备分类） ====================
const filterDevice = ref('');
const deviceOptions = computed(() => devices.value.map(d => d.name));

const filterPipeFactory = ref('');
const pipeFactoryOptions = computed(() => {
  const set = new Set();
  allVersions.value.forEach(v => { if (v.pipeFactory) set.add(v.pipeFactory); });
  return Array.from(set);
});

const filteredVersions = computed(() => {
  return allVersions.value.filter(v =>
    (!filterDevice.value || v.deviceCategory === filterDevice.value) &&
    (!filterPipeFactory.value || v.pipeFactory === filterPipeFactory.value)
  );
});

// ==================== 分页（默认每页 10 条） ====================
const currentPage = ref(1);
const pageSize = ref(10);
const totalPages = computed(() => Math.max(1, Math.ceil(filteredVersions.value.length / pageSize.value)));
const pagedVersions = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value;
  return filteredVersions.value.slice(start, start + pageSize.value);
});
// 筛选变化 / 数据减少导致越界时，回到有效页
watch([filterDevice, filterPipeFactory], () => { currentPage.value = 1; });
watch(totalPages, (tp) => { if (currentPage.value > tp) currentPage.value = tp; });

const resetFilters = () => {
  filterDevice.value = '';
  filterPipeFactory.value = '';
  currentPage.value = 1;
};

// ==================== 查看详情弹窗 ====================
const showDetailDialog = ref(false);
const detailVersion = ref(null);

const openDetail = (version) => {
  detailVersion.value = version;
  showDetailDialog.value = true;
};

const closeDetail = () => {
  showDetailDialog.value = false;
};

// ==================== 详情字段（过滤空值） ====================
const buildDetailFields = (v) => {
  const fields = [
    ['管厂', v.pipeFactory],
    ['设备名称', v.deviceName],
    ['设备型号', v.model],
    ['出厂编号', v.serial],
    ['出厂日期', v.factoryDate],
    ['产线', v.productionLine],
    ['车间名称', v.workshop],
    ['产线类型', v.lineType],
    ['工序名称', v.process],
    ['是否注码', v.isCoded],
    ['是否PLC信号', v.isPlc],
    ['上位机软件名称', v.upperSoftName],
    ['上位机软件版本号', v.upperSoftVer],
    ['下位机软件版本号', v.lowerSoftVer],
    ['台数', (v.count !== '' && v.count != null) ? v.count + ' 台' : ''],
    ['备注', v.remark],
    ['其他说明', v.otherDesc]
  ];
  return fields.map(f => [f[0], (f[1] === '' || f[1] == null) ? '暂无' : f[1]]);
};

// ==================== 新增 / 编辑设备台账 ====================
const showAddVersionDialog = ref(false);
const dialogMode = ref('add'); // 'add' | 'edit'
const editingUid = ref(null);
const versionForm = ref({
  deviceCategory: '',
  pipeFactory: '',
  deviceName: '',
  model: '',
  serial: '',
  factoryDate: '',
  productionLine: '',
  workshop: '',
  lineType: '',
  process: '',
  isCoded: '',
  isPlc: '',
  upperSoftName: '',
  upperSoftVer: '',
  lowerSoftVer: '',
  count: '',
  remark: '',
  otherDesc: '',
  features: []
});
const newFeatureType = ref('new');
const newFeatureText = ref('');

const blankForm = () => ({
  deviceCategory: deviceOptions.value[0] || '',
  pipeFactory: '',
  deviceName: '',
  model: '',
  serial: '',
  factoryDate: new Date().toISOString().slice(0, 10),
  productionLine: '',
  workshop: '',
  lineType: '',
  process: '',
  isCoded: '',
  isPlc: '',
  upperSoftName: '',
  upperSoftVer: '',
  lowerSoftVer: '',
  count: '',
  remark: '',
  otherDesc: '',
  features: []
});

const openAddVersionDialog = () => {
  dialogMode.value = 'add';
  editingUid.value = null;
  versionForm.value = blankForm();
  newFeatureType.value = 'new';
  newFeatureText.value = '';
  showAddVersionDialog.value = true;
};

const openEditDialog = (version) => {
  dialogMode.value = 'edit';
  editingUid.value = version.uid;
  versionForm.value = {
    deviceCategory: version.deviceCategory,
    pipeFactory: version.pipeFactory || '',
    deviceName: version.deviceName || '',
    model: version.model || '',
    serial: version.serial || '',
    factoryDate: version.factoryDate || '',
    productionLine: version.productionLine || '',
    workshop: version.workshop || '',
    lineType: version.lineType || '',
    process: version.process || '',
    isCoded: version.isCoded || '',
    isPlc: version.isPlc || '',
    upperSoftName: version.upperSoftName || '',
    upperSoftVer: version.upperSoftVer || '',
    lowerSoftVer: version.lowerSoftVer || '',
    count: version.count ?? '',
    remark: version.remark || '',
    features: (version.features || []).map(f => ({ ...f }))
  };
  newFeatureType.value = 'new';
  newFeatureText.value = '';
  showAddVersionDialog.value = true;
};

const addFeatureRow = () => {
  const text = newFeatureText.value.trim();
  if (!text) return;
  versionForm.value.features.push({ type: newFeatureType.value, text });
  newFeatureText.value = '';
};

const removeFeatureRow = (index) => {
  versionForm.value.features.splice(index, 1);
};

const buildRecord = (f, uid) => ({
  uid,
  pipeFactory: (f.pipeFactory || '').trim(),
  deviceName: (f.deviceName || '').trim() || f.deviceCategory,
  model: (f.model || '').trim(),
  serial: (f.serial || '').trim(),
  factoryDate: f.factoryDate || new Date().toISOString().slice(0, 10),
  productionLine: f.productionLine || '',
  workshop: (f.workshop || '').trim(),
  lineType: (f.lineType || '').trim(),
  process: (f.process || '').trim(),
  isCoded: f.isCoded || '',
  isPlc: f.isPlc || '',
  upperSoftName: (f.upperSoftName || '').trim(),
  upperSoftVer: (f.upperSoftVer || '').trim(),
  lowerSoftVer: (f.lowerSoftVer || '').trim(),
  count: f.count ?? '',
  remark: (f.remark || '').trim(),
  features: [...(f.features || [])]
});

const confirmSave = () => {
  const f = versionForm.value;
  if (!f.deviceCategory) {
    ElMessage.warning('请选择设备名称');
    return;
  }
  if (!f.serial.trim()) {
    ElMessage.warning('请填写出厂编号');
    return;
  }

  if (dialogMode.value === 'add') {
    const device = devices.value.find(d => d.name === f.deviceCategory);
    if (!device) {
      ElMessage.warning('未找到对应设备分类');
      return;
    }
    device.versions.unshift(buildRecord(f, 'dv' + (++uidSeq)));
    ElMessage.success(`设备 ${f.deviceCategory}（${f.serial.trim()}）已添加`);
  } else {
    let target = null;
    let dev = null;
    for (const d of devices.value) {
      const rec = d.versions.find(v => v.uid === editingUid.value);
      if (rec) { target = rec; dev = d; break; }
    }
    if (!target) {
      ElMessage.warning('未找到该设备记录');
      return;
    }
    const updated = buildRecord(f, editingUid.value);
    if (dev.name === f.deviceCategory) {
      Object.assign(target, updated);
    } else {
      dev.versions = dev.versions.filter(v => v.uid !== editingUid.value);
      const newDev = devices.value.find(d => d.name === f.deviceCategory);
      if (newDev) newDev.versions.unshift(updated);
    }
    ElMessage.success('已保存修改');
  }
  showAddVersionDialog.value = false;
};

// ==================== 删除设备台账 ====================
const handleDelete = (version) => {
  ElMessageBox.confirm(
    `确认删除设备「${version.deviceName || version.deviceCategory}（${version.serial || '未填编号'}）」吗？删除后不可恢复。`,
    '删除确认',
    { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' }
  ).then(() => {
    for (const d of devices.value) {
      const idx = d.versions.findIndex(v => v.uid === version.uid);
      if (idx !== -1) {
        d.versions.splice(idx, 1);
        break;
      }
    }
    if (selectedVersionId.value === version.uid) selectedVersionId.value = null;
    ElMessage.success('已删除该设备记录');
  }).catch(() => {});
};

// ==================== xlsx 表格导入 ====================
const fileInputRef = ref(null);
const showImportResult = ref(false);
const importResult = ref({ success: 0, errors: [] });

// 表格列定义（表头中文 -> 字段 key），顺序按模板要求
const excelColumns = [
  { key: 'pipeFactory', label: '管厂' },
  { key: 'productionLine', label: '产线厂家' },
  { key: 'workshop', label: '车间名称' },
  { key: 'lineType', label: '产线类型' },
  { key: 'process', label: '工序名称' },
  { key: 'isCoded', label: '是否注码' },
  { key: 'isPlc', label: '是否PLC信号' },
  { key: 'upperSoftVer', label: '上位机软件版本号' },
  { key: 'upperSoftName', label: '上位机软件名称' },
  { key: 'lowerSoftVer', label: '下位机软件版本号' },
  { key: 'deviceCategory', label: '设备名称', required: true },
  { key: 'model', label: '设备型号' },
  { key: 'serial', label: '出厂编号', required: true },
  { key: 'factoryDate', label: '出厂日期' },
  { key: 'remark', label: '信息备注' },
  { key: 'otherDesc', label: '其他说明' }
];

const triggerImport = () => fileInputRef.value?.click();

const handleFileChange = (e) => {
  const file = e.target.files && e.target.files[0];
  if (!file) return;
  const reader = new FileReader();
  reader.onload = (ev) => {
    try {
      const wb = XLSX.read(ev.target.result, { type: 'array' });
      const firstSheetName = wb.SheetNames[0];
      const ws = wb.Sheets[firstSheetName];
      const rows = XLSX.utils.sheet_to_json(ws, { defval: '' });
      if (!rows.length) {
        ElMessage.warning('Excel 中没有可导入的数据行');
        return;
      }
      importDevices(rows);
    } catch (err) {
      ElMessage.error('解析 Excel 失败：' + (err && err.message ? err.message : err));
    } finally {
      e.target.value = '';
    }
  };
  reader.onerror = () => {
    ElMessage.error('文件读取失败，请重试');
    e.target.value = '';
  };
  reader.readAsArrayBuffer(file);
};

const importDevices = (rows) => {
  let success = 0;
  const errors = [];
  const validCategories = devices.value.map(d => d.name);
  const today = new Date().toISOString().slice(0, 10);

  rows.forEach((row, idx) => {
    const lineNo = idx + 2; // 第 1 行为表头
    const get = (label) => {
      const v = row[label];
      return v === undefined || v === null ? '' : String(v).trim();
    };
    const category = get('设备名称');
    const serial = get('出厂编号');
    if (!category) { errors.push(`第 ${lineNo} 行：缺少"设备名称"`); return; }
    if (!serial) { errors.push(`第 ${lineNo} 行：缺少"出厂编号"`); return; }
    // 识别到符号 "-" 时，自动匹配为「其他」分类
    const targetCategory = category.includes('-') ? '其他' : category;
    const device = devices.value.find(d => d.name === targetCategory);
    if (!device) {
      errors.push(`第 ${lineNo} 行：设备分类"${category}"不存在（可选：${validCategories.join('、')}）`);
      return;
    }
      const countRaw = get('台数');
      const rec = {
        uid: 'dv' + (++uidSeq),
        pipeFactory: get('管厂'),
        deviceName: category,
        deviceCategory: targetCategory,
        model: get('设备型号'),
        serial,
        factoryDate: get('出厂日期') || today,
        productionLine: get('产线厂家'),
        workshop: get('车间名称'),
        lineType: get('产线类型'),
        process: get('工序名称'),
        isCoded: get('是否注码'),
        isPlc: get('是否PLC信号'),
        upperSoftName: get('上位机软件名称'),
        upperSoftVer: get('上位机软件版本号'),
        lowerSoftVer: get('下位机软件版本号'),
        count: countRaw === '' ? '' : Number(countRaw),
        remark: get('信息备注'),
        otherDesc: get('其他说明'),
        features: []
      };
      const feat = get('变更内容');
      if (feat) rec.features.push({ type: 'new', text: feat });
    device.versions.unshift(rec);
    success++;
  });

  importResult.value = { success, errors };
  showImportResult.value = true;
};

// 下载 xlsx 导入模板
const downloadTemplate = () => {
  const header = excelColumns.map(c => c.label);
  const example = [
    '示例管厂',
    '创者', '总装车间', '装配线', '整机调试', '是', '否',
    'V2.3.1', 'BlastMaster PC', 'V1.8.0',
    '精密检测注码仪', 'XYS-2000', 'SN2026070001', '2026-07-01',
    '示例数据，请按此格式替换或追加', '首批量产'
  ];
  const ws = XLSX.utils.aoa_to_sheet([header, example]);
  // 设定列宽，便于阅读
  ws['!cols'] = excelColumns.map(() => ({ wch: 16 }));
  const wb = XLSX.utils.book_new();
  XLSX.utils.book_append_sheet(wb, ws, '设备台账');
  XLSX.writeFile(wb, '设备台账导入模板.xlsx');
};

// ==================== 导出 / 导入 独立 JSON 数据文件 ====================
// 导出：拍平为纯设备台账记录（去掉分类壳结构），清晰可读、便于迁移
const exportData = () => {
  const records = allVersions.value.map(v => {
    const { deviceColor, ...rest } = v;
    return rest;
  });
  if (!records.length) {
    ElMessage.warning('当前没有可导出的设备台账记录');
    return;
  }
  const data = JSON.stringify(records, null, 2);
  const blob = new Blob([data], { type: 'application/json;charset=utf-8' });
  const date = new Date().toISOString().slice(0, 10);
  saveAs(blob, `设备台账数据_${date}.json`);
  ElMessage.success(`已导出 ${records.length} 条设备台账记录`);
};

const jsonInputRef = ref(null);
const triggerImportJson = () => jsonInputRef.value?.click();

// 把拍平的台账记录按「设备名称」归位到对应分类，并自动持久化到本地
const importFlatRecords = (records) => {
  if (!records.length) {
    ElMessage.warning('文件中没有可导出的设备记录');
    return;
  }
  let success = 0;
  const errors = [];
  const validCategories = devices.value.map(d => d.name);
  // 先清空各分类现有记录，整体替换为导入数据
  devices.value.forEach(d => { d.versions = []; });
  records.forEach((rec, idx) => {
    const cat = (rec.deviceCategory || rec.deviceName || '').trim();
    if (!cat) { errors.push(`第 ${idx + 1} 条：缺少"设备名称"`); return; }
    // 识别到符号 "-" 时，自动匹配为「其他」分类
    const targetCategory = cat.includes('-') ? '其他' : cat;
    const device = devices.value.find(d => d.name === targetCategory);
    if (!device) {
      errors.push(`第 ${idx + 1} 条：设备名称"${cat}"不存在（可选：${validCategories.join('、')}）`);
      return;
    }
    const newRec = { ...rec };
    if (!newRec.uid) newRec.uid = 'dv' + (++uidSeq);
    device.versions.push(newRec);
    success++;
  });
  importResult.value = { success, errors };
  showImportResult.value = true;
};

const handleJsonChange = (e) => {
  const file = e.target.files && e.target.files[0];
  if (!file) return;
  const reader = new FileReader();
  reader.onload = (ev) => {
    try {
      const parsed = JSON.parse(ev.target.result);
      if (!Array.isArray(parsed)) throw new Error('文件格式不正确，应为 JSON 数组');
      // 兼容两种格式：① 旧版分类结构（含 versions 字段）；② 新版拍平台账记录（含设备名称/deviceCategory）
      const isDeviceStructure = parsed.length > 0 && parsed[0] &&
        Object.prototype.hasOwnProperty.call(parsed[0], 'versions');
      if (isDeviceStructure) {
        devices.value = parsed;
        const total = parsed.reduce((s, d) => s + (d.versions ? d.versions.length : 0), 0);
        ElMessage.success(`已导入 ${parsed.length} 个分类、共 ${total} 条设备数据`);
      } else {
        importFlatRecords(parsed);
      }
    } catch (err) {
      ElMessage.error('解析 JSON 失败：' + (err && err.message ? err.message : err));
    } finally {
      e.target.value = '';
    }
  };
  reader.onerror = () => {
    ElMessage.error('文件读取失败，请重试');
    e.target.value = '';
  };
  reader.readAsText(file);
};

// ==================== 特征类型 ====================
const getFeatureTypeLabel = (type) => {
  switch (type) {
    case 'new': return '新增';
    case 'improve': return '优化';
    case 'fix': return '修复';
    default: return type;
  }
};

const getFeatureTypeClass = (type) => {
  switch (type) {
    case 'new': return 'feature-new';
    case 'improve': return 'feature-improve';
    case 'fix': return 'feature-fix';
    default: return 'feature-other';
  }
};
</script>

<template>
  <div class="device-tab">
    <!-- ==================== 筛选栏 ==================== -->
    <div class="filter-bar">
      <div class="filter-group">
        <label class="filter-label">设备</label>
        <select v-model="filterDevice" class="filter-select">
          <option value="">全部设备</option>
          <option v-for="o in deviceOptions" :key="o" :value="o">{{ o }}</option>
        </select>
      </div>
      <div class="filter-group">
        <label class="filter-label">管厂</label>
        <select v-model="filterPipeFactory" class="filter-select">
          <option value="">全部管厂</option>
          <option v-for="o in pipeFactoryOptions" :key="o" :value="o">{{ o }}</option>
        </select>
      </div>
      <button class="filter-reset" @click="resetFilters" v-if="filterDevice || filterPipeFactory">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8" />
          <path d="M3 3v5h5" />
        </svg>
        <span>重置筛选</span>
      </button>
      <span class="filter-count">共 {{ filteredVersions.length }} 台设备</span>
      <div class="tool-btns">
        <button class="import-template-btn" @click="downloadTemplate">
          <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
            <polyline points="7 10 12 15 17 10" />
            <line x1="12" y1="15" x2="12" y2="3" />
          </svg>
          <span>下载模板</span>
        </button>
        <button class="import-xlsx-btn" @click="triggerImport">
          <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
            <polyline points="14 2 14 8 20 8" />
            <line x1="12" y1="18" x2="12" y2="12" />
            <line x1="9" y1="15" x2="15" y2="15" />
          </svg>
          <span>导入 xlsx</span>
        </button>
        <input ref="fileInputRef" type="file" accept=".xlsx,.xls" style="display:none" @change="handleFileChange" />
        <button class="import-template-btn" @click="exportData" title="导出全部设备台账为 JSON 文件">
          <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
            <polyline points="7 10 12 15 17 10" />
            <line x1="12" y1="15" x2="12" y2="3" />
          </svg>
          <span>导出数据</span>
        </button>
        <button class="import-template-btn" @click="triggerImportJson" title="从 JSON 文件导入设备台账">
          <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
            <polyline points="14 2 14 8 20 8" />
            <line x1="12" y1="18" x2="12" y2="12" />
            <line x1="9" y1="15" x2="15" y2="15" />
          </svg>
          <span>导入数据</span>
        </button>
        <input ref="jsonInputRef" type="file" accept=".json" style="display:none" @change="handleJsonChange" />
        <button class="add-version-btn" @click="openAddVersionDialog">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <line x1="12" y1="5" x2="12" y2="19" />
            <line x1="5" y1="12" x2="19" y2="12" />
          </svg>
          <span>新增版本</span>
        </button>
      </div>
    </div>

    <!-- 提示信息 -->
    <div class="upload-hint">
      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <circle cx="12" cy="12" r="10" />
        <line x1="12" y1="16" x2="12" y2="12" />
        <line x1="12" y1="8" x2="12.01" y2="8" />
      </svg>
      <span>
        支持「新增版本」手动登记，或点击「导入 xlsx」批量导入设备台账（先点「下载模板」获取表头格式）。导入按"设备名称"归入对应设备，必填"出厂编号"。
      </span>
    </div>

    <!-- ==================== 设备台账卡片 ==================== -->
    <div class="version-grid" v-if="filteredVersions.length">
      <div
        v-for="version in pagedVersions"
        :key="version.uid"
        class="version-card"
      >
        <div class="vc-top">
          <span class="version-id">{{ version.deviceName || '未命名设备' }}</span>
          <div class="card-actions">
            <button class="action-btn" title="编辑" @click="openEditDialog(version)">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" />
                <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" />
              </svg>
            </button>
            <button class="action-btn danger" title="删除" @click="handleDelete(version)">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <polyline points="3 6 5 6 21 6" />
                <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
                <line x1="10" y1="11" x2="10" y2="17" />
                <line x1="14" y1="11" x2="14" y2="17" />
              </svg>
            </button>
          </div>
        </div>





        <div class="version-info">
          <div class="info-row" v-if="version.pipeFactory"><span class="info-label">管厂</span><span class="info-value">{{ version.pipeFactory }}</span></div>
          <div class="info-row" v-if="version.productionLine"><span class="info-label">产线厂家</span><span class="info-value">{{ version.productionLine }}</span></div>
          <div class="info-row" v-if="version.workshop"><span class="info-label">车间名称</span><span class="info-value">{{ version.workshop }}</span></div>
          <div class="info-row" v-if="version.lineType"><span class="info-label">产线类型</span><span class="info-value">{{ version.lineType }}</span></div>
          <div class="info-row" v-if="version.process"><span class="info-label">工序名称</span><span class="info-value">{{ version.process }}</span></div>
        </div>

        <div class="version-meta">
          <span v-if="version.count !== '' && version.count != null" class="meta-badge">台数 {{ version.count }}</span>
          <span v-if="version.isCoded" class="meta-badge" :class="version.isCoded === '是' ? 'meta-yes' : 'meta-no'">注码 {{ version.isCoded }}</span>
          <span v-if="version.isPlc" class="meta-badge" :class="version.isPlc === '是' ? 'meta-yes' : 'meta-no'">PLC {{ version.isPlc }}</span>
        </div>

        <button class="vc-toggle" @click="openDetail(version)">查看详情</button>
      </div>
    </div>

    <!-- 分页 -->
    <div class="pagination" v-if="filteredVersions.length > pageSize">
      <span class="page-info">共 {{ filteredVersions.length }} 条 · 第 {{ currentPage }}/{{ totalPages }} 页</span>
      <div class="page-controls">
        <button class="page-btn" :disabled="currentPage === 1" @click="currentPage--">上一页</button>
        <button
          class="page-btn"
          v-for="p in totalPages"
          :key="p"
          :class="{ active: p === currentPage }"
          @click="currentPage = p"
        >{{ p }}</button>
        <button class="page-btn" :disabled="currentPage === totalPages" @click="currentPage++">下一页</button>
      </div>
      <select v-model="pageSize" class="page-size">
        <option :value="10">10 条/页</option>
        <option :value="20">20 条/页</option>
        <option :value="50">50 条/页</option>
      </select>
    </div>

    <!-- 空状态 -->
    <div class="version-empty" v-else>
      <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="#c9cdd4" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
        <circle cx="11" cy="11" r="8" />
        <line x1="21" y1="21" x2="16.65" y2="16.65" />
      </svg>
      <p v-if="filterDevice">「{{ filterDevice }}」暂无设备台账，请点击"新增版本"添加，或调整设备筛选</p>
      <p v-else>暂无设备台账，请点击右上角"新增版本"按钮登记设备信息</p>
    </div>

    <!-- ==================== 新增版本对话框 ==================== -->
    <Teleport to="body">
    <div v-if="showAddVersionDialog" class="add-version-modal" @click.self="showAddVersionDialog = false">
      <div class="add-version-dialog">
        <div class="dialog-header">
          <h3>{{ dialogMode === 'edit' ? '编辑设备版本' : '新增设备版本' }}</h3>
          <button class="close-btn" @click="showAddVersionDialog = false">×</button>
        </div>

        <div class="dialog-body">
          <div class="form-section-title">基础信息</div>
          <div class="form-row">
            <label class="form-label">管厂</label>
            <input v-model="versionForm.pipeFactory" class="form-input" placeholder="如 示例管厂" />
          </div>
          <div class="form-row">
            <label class="form-label required">设备名称</label>
            <select v-model="versionForm.deviceCategory" class="form-select-small">
              <option v-for="d in deviceOptions" :key="d" :value="d">{{ d }}</option>
            </select>
          </div>
          <div class="form-row two-col">
            <div>
              <label class="form-label">设备型号</label>
              <input v-model="versionForm.model" class="form-input" placeholder="如 XYS-2000" />
            </div>
            <div>
              <label class="form-label required">出厂编号</label>
              <input v-model="versionForm.serial" class="form-input" placeholder="如 SN2026070001" />
            </div>
          </div>
          <div class="form-row">
            <label class="form-label">出厂日期</label>
            <input v-model="versionForm.factoryDate" type="date" class="form-input" />
          </div>

          <div class="form-section-title">产线与车间</div>
          <div class="form-row">
            <label class="form-label">产线</label>
            <select v-model="versionForm.productionLine" class="form-select-small">
              <option value="">请选择产线</option>
              <option v-for="o in productionLineOptions" :key="o" :value="o">{{ o }}</option>
            </select>
          </div>
          <div class="form-row two-col">
            <div>
              <label class="form-label">车间名称</label>
              <input v-model="versionForm.workshop" class="form-input" placeholder="如 总装车间" />
            </div>
            <div>
              <label class="form-label">产线类型</label>
              <input v-model="versionForm.lineType" class="form-input" placeholder="如 装配线" />
            </div>
          </div>
          <div class="form-row">
            <label class="form-label">工序名称</label>
            <input v-model="versionForm.process" class="form-input" placeholder="如 整机调试" />
          </div>

          <div class="form-section-title">功能与软件</div>
          <div class="form-row two-col">
            <div>
              <label class="form-label">是否注码</label>
              <select v-model="versionForm.isCoded" class="form-select-small">
                <option value="">请选择</option>
                <option value="是">是</option>
                <option value="否">否</option>
              </select>
            </div>
            <div>
              <label class="form-label">是否PLC信号</label>
              <select v-model="versionForm.isPlc" class="form-select-small">
                <option value="">请选择</option>
                <option value="是">是</option>
                <option value="否">否</option>
              </select>
            </div>
          </div>
          <div class="form-row">
            <label class="form-label">上位机软件名称</label>
            <input v-model="versionForm.upperSoftName" class="form-input" placeholder="如 BlastMaster PC" />
          </div>
          <div class="form-row two-col">
            <div>
              <label class="form-label">上位机软件版本号</label>
              <input v-model="versionForm.upperSoftVer" class="form-input" placeholder="如 V2.3.1" />
            </div>
            <div>
              <label class="form-label">下位机软件版本号</label>
              <input v-model="versionForm.lowerSoftVer" class="form-input" placeholder="如 V1.8.0" />
            </div>
          </div>

          <div class="form-section-title">其他</div>
          <div class="form-row two-col">
            <div>
              <label class="form-label">台数</label>
              <input v-model="versionForm.count" type="number" min="0" class="form-input" placeholder="如 12" />
            </div>
            <div></div>
          </div>
          <div class="form-row">
            <label class="form-label">备注</label>
            <textarea v-model="versionForm.remark" class="form-textarea" rows="2" placeholder="补充说明..."></textarea>
          </div>
          <div class="form-row">
            <label class="form-label">其他说明</label>
            <textarea v-model="versionForm.otherDesc" class="form-textarea" rows="2" placeholder="其他补充说明..."></textarea>
          </div>

          <div class="form-section-title">变更内容（可选）</div>
          <div class="form-row">
            <div class="feature-input-row">
              <select v-model="newFeatureType" class="form-select-small">
                <option value="new">新增</option>
                <option value="improve">优化</option>
                <option value="fix">修复</option>
              </select>
              <input
                v-model="newFeatureText"
                class="form-input"
                placeholder="输入变更说明..."
                @keyup.enter="addFeatureRow"
              />
              <button class="add-feature-btn" @click="addFeatureRow">添加</button>
            </div>
            <div class="feature-list-form" v-if="versionForm.features.length">
              <div
                v-for="(f, i) in versionForm.features"
                :key="i"
                class="feature-list-item"
                :class="'type-' + f.type"
              >
                <span class="feature-type-tag">{{ getFeatureTypeLabel(f.type) }}</span>
                <span class="feature-text-form">{{ f.text }}</span>
                <button class="remove-feature-btn" @click="removeFeatureRow(i)">×</button>
              </div>
            </div>
            <div class="feature-empty-hint" v-else>暂未添加变更内容</div>
          </div>
        </div>

        <div class="dialog-footer">
          <button class="form-cancel" @click="showAddVersionDialog = false">取消</button>
          <button class="form-confirm" @click="confirmSave">{{ dialogMode === 'edit' ? '保存修改' : '确认新增' }}</button>
        </div>
      </div>
    </div>
    </Teleport>

    <!-- ==================== 导入结果弹窗 ==================== -->
    <Teleport to="body">
      <div v-if="showImportResult" class="add-version-modal" @click.self="showImportResult = false">
        <div class="add-version-dialog import-result-dialog">
          <div class="dialog-header">
            <h3>导入结果</h3>
            <button class="close-btn" @click="showImportResult = false">×</button>
          </div>
          <div class="dialog-body">
            <div class="import-summary">
              <span class="import-ok">成功导入 {{ importResult.success }} 条</span>
              <span v-if="importResult.errors.length" class="import-fail">失败 {{ importResult.errors.length }} 条</span>
              <span v-else class="import-ok-sub">全部成功</span>
            </div>
            <div v-if="importResult.errors.length" class="import-error-list">
              <div class="detail-subtitle">失败明细</div>
              <div v-for="(err, i) in importResult.errors" :key="i" class="import-error-item">
                {{ err }}
              </div>
            </div>
            <div v-else class="import-all-ok">
              所有数据均已成功导入设备台账，可在上方列表查看。
            </div>
          </div>
          <div class="dialog-footer">
            <button class="form-confirm" @click="showImportResult = false">知道了</button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- ==================== 设备详情弹窗 ==================== -->
    <Teleport to="body">
      <div v-if="showDetailDialog" class="add-version-modal" @click.self="closeDetail">
        <div class="add-version-dialog detail-modal">
          <div class="dialog-header">
            <h3>设备详情</h3>
            <button class="close-btn" @click="closeDetail">×</button>
          </div>
          <div class="dialog-body" v-if="detailVersion">
            <div class="detail-grid">
              <div class="detail-item" v-for="(f, i) in buildDetailFields(detailVersion)" :key="i">
                <span class="detail-label">{{ f[0] }}</span>
                <span class="detail-value">{{ f[1] }}</span>
              </div>
            </div>
            <div class="detail-changes" v-if="detailVersion.features && detailVersion.features.length">
              <div class="detail-subtitle">变更内容</div>
              <div class="feature-list">
                <div
                  v-for="(feature, index) in detailVersion.features"
                  :key="index"
                  class="feature-item"
                  :class="getFeatureTypeClass(feature.type)"
                >
                  <span class="feature-type">{{ getFeatureTypeLabel(feature.type) }}</span>
                  <span class="feature-text">{{ feature.text }}</span>
                </div>
              </div>
            </div>
          </div>
          <div class="dialog-footer">
            <button class="form-confirm" @click="closeDetail">关闭</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
/* ==================== 外层容器 ==================== */
.device-tab {
  max-width: 1500px;
  margin: 0 auto;
}

/* ==================== 筛选栏 ==================== */
.filter-bar {
  display: flex;
  align-items: center;
  gap: 14px;
  flex-wrap: wrap;
  padding: 14px 16px;
  background: #f7f9fc;
  border: 1px solid #ebedf0;
  border-radius: 12px;
  margin-bottom: 16px;
}

.filter-group {
  display: flex;
  align-items: center;
  gap: 8px;
}

.filter-label {
  font-size: 13px;
  color: #4e5969;
  font-weight: 500;
  white-space: nowrap;
}

.filter-select {
  min-width: 180px;
  padding: 8px 12px;
  border: 1px solid #dcdfe6;
  border-radius: 8px;
  font-size: 14px;
  color: #1d2129;
  background: white;
  cursor: pointer;
  transition: border-color 0.2s;
}

.filter-select:focus {
  outline: none;
  border-color: #165DFF;
  box-shadow: 0 0 0 3px rgba(22, 93, 255, 0.1);
}

.filter-reset {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 8px 14px;
  border: 1px solid #dcdfe6;
  border-radius: 8px;
  background: white;
  color: #4e5969;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}

.filter-reset:hover {
  border-color: #f53f3f;
  color: #f53f3f;
}

.filter-count {
  font-size: 13px;
  color: #86909c;
  margin-left: auto;
}

/* ==================== 工具按钮（导入 / 模板 / 新增） ==================== */
.tool-btns {
  display: flex;
  align-items: center;
  gap: 10px;
}

.import-template-btn,
.import-xlsx-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  border: 1px solid #c5d8ff;
  border-radius: 8px;
  background: white;
  color: #165DFF;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.import-template-btn {
  color: #4e5969;
  border-color: #dcdfe6;
}

.import-template-btn:hover {
  border-color: #4080ff;
  color: #4080ff;
  background: #eef4ff;
}

.import-xlsx-btn:hover {
  border-color: #4080ff;
  background: #eef4ff;
  box-shadow: 0 2px 6px rgba(22, 93, 255, 0.2);
}

/* ==================== 提示信息 ==================== */
.upload-hint {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 10px 14px;
  background: #f2f7ff;
  border: 1px solid #d6e4ff;
  border-radius: 8px;
  color: #4e5969;
  font-size: 13px;
  line-height: 1.6;
  margin-bottom: 16px;
}

.upload-hint svg {
  flex-shrink: 0;
  margin-top: 3px;
  color: #4080ff;
}

/* ==================== 卡片网格 ==================== */
.version-grid {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 16px;
  align-items: stretch;
}

@media (max-width: 1500px) {
  .version-grid { grid-template-columns: repeat(4, 1fr); }
}
@media (max-width: 1180px) {
  .version-grid { grid-template-columns: repeat(3, 1fr); }
}
@media (max-width: 860px) {
  .version-grid { grid-template-columns: repeat(2, 1fr); }
}
@media (max-width: 560px) {
  .version-grid { grid-template-columns: 1fr; }
}

/* ==================== 卡片 ==================== */
.version-card {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 16px;
  background: white;
  border: 1px solid #ebedf0;
  border-radius: 14px;
  overflow: hidden;
  transition: all 0.3s;
}

.version-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 4px;
  background: linear-gradient(90deg, #4080ff, #165DFF);
  opacity: 0;
  transition: opacity 0.3s;
}

.version-card:hover {
  transform: translateY(-4px);
  border-color: #c5d8ff;
  box-shadow: 0 12px 28px rgba(22, 93, 255, 0.15);
}

.version-card.selected {
  border-color: #4080ff;
  box-shadow: 0 8px 24px rgba(64, 128, 255, 0.18);
}

.version-card.selected::before {
  opacity: 1;
}

.vc-top {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 8px;
}

.vc-top .version-id {
  font-size: 18px;
  font-weight: 700;
  color: #1d2129;
  line-height: 1.3;
  word-break: break-all;
}

.card-actions {
  display: flex;
  gap: 6px;
  flex-shrink: 0;
}

.action-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border: 1px solid #e5e6eb;
  border-radius: 6px;
  background: #f7f9fc;
  color: #4e5969;
  cursor: pointer;
  transition: all 0.2s;
}

.action-btn:hover {
  border-color: #4080ff;
  color: #4080ff;
  background: #eef4ff;
}

.action-btn.danger:hover {
  border-color: #ff4d4f;
  color: #ff4d4f;
  background: #fff2f0;
}

.version-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.version-device {
  display: inline-flex;
  align-items: center;
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 4px;
  border: 1px solid transparent;
}

.version-line-tag {
  display: inline-flex;
  align-items: center;
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 4px;
  background: #f2f3f5;
  color: #4e5969;
  border: 1px solid #e5e6eb;
}

.version-pipe-tag {
  display: inline-flex;
  align-items: center;
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 4px;
  background: #fff7e6;
  color: #d4660a;
  border: 1px solid #ffe1ba;
  font-weight: 500;
}

.version-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.meta-badge {
  font-size: 12px;
  padding: 3px 8px;
  border-radius: 4px;
  background: #f2f3f5;
  color: #4e5969;
}

.meta-yes {
  background: #e8f8e8;
  color: #52c41a;
}

.meta-no {
  background: #fff2f0;
  color: #ff4d4f;
}

/* ==================== 卡片信息行（管厂/产线厂家/车间/产线类型/工序） ==================== */
.version-info {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 6px 16px;
  padding: 2px 0;
}

.info-row {
  display: flex;
  align-items: baseline;
  gap: 6px;
  font-size: 13px;
  min-width: 0;
}

.info-label {
  color: #86909c;
  flex-shrink: 0;
}

.info-value {
  color: #1d2129;
  font-weight: 500;
  word-break: break-all;
}

/* ==================== 分页 ==================== */
.pagination {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 12px;
  margin-top: 18px;
  padding: 12px 16px;
  background: white;
  border: 1px solid #ebedf0;
  border-radius: 12px;
}

.page-info {
  font-size: 13px;
  color: #86909c;
}

.page-controls {
  display: flex;
  gap: 6px;
}

.page-btn {
  min-width: 32px;
  height: 32px;
  padding: 0 8px;
  border: 1px solid #e5e6eb;
  border-radius: 6px;
  background: #fff;
  color: #4e5969;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}

.page-btn:hover:not(:disabled) {
  border-color: #4080ff;
  color: #4080ff;
}

.page-btn.active {
  background: #4080ff;
  border-color: #4080ff;
  color: #fff;
}

.page-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.page-size {
  height: 32px;
  border: 1px solid #e5e6eb;
  border-radius: 6px;
  padding: 0 8px;
  font-size: 13px;
  color: #4e5969;
  background: #fff;
  cursor: pointer;
}

.vc-toggle {
  margin-top: auto;
  padding: 8px;
  border: 1px solid #e5e6eb;
  border-radius: 8px;
  background: #f7f9fc;
  color: #4080ff;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}

.vc-toggle:hover {
  border-color: #4080ff;
  background: #eef4ff;
}

.version-detail {
  margin: 4px -16px -16px;
  padding: 16px;
  background: #f7f9fc;
  border-top: 1px solid #eef0f3;
}

.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  font-weight: 600;
  color: #333;
}

.detail-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 8px;
}

.detail-item {
  display: flex;
  gap: 10px;
  font-size: 13px;
  line-height: 1.5;
}

.detail-label {
  flex-shrink: 0;
  width: 96px;
  color: #86909c;
}

.detail-value {
  color: #1d2129;
  word-break: break-all;
}

.detail-changes {
  margin-top: 14px;
}

.detail-subtitle {
  font-size: 13px;
  font-weight: 600;
  color: #4e5969;
  margin-bottom: 8px;
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

.feature-other .feature-type {
  background: #f5f5f5;
  color: #999;
}

.feature-text {
  font-size: 14px;
  color: #333;
}

/* ==================== 列表空状态 ==================== */
.version-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 60px 20px;
  color: #86909c;
  background: #fafbfc;
  border: 1px dashed #e5e6eb;
  border-radius: 12px;
}

.version-empty p {
  margin: 0;
  font-size: 14px;
}

/* ==================== 新增版本按钮 ==================== */
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

/* ==================== 新增版本对话框 ==================== */
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
  width: 620px;
  max-width: 92vw;
  max-height: 88vh;
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

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 14px 20px;
  border-top: 1px solid #f0f0f0;
}

.form-section-title {
  font-size: 13px;
  font-weight: 600;
  color: #165DFF;
  margin: 18px 0 10px;
  padding-left: 8px;
  border-left: 3px solid #165DFF;
}

.form-section-title:first-child {
  margin-top: 0;
}

.form-cancel {
  padding: 8px 20px;
  border: 1px solid #dcdfe6;
  border-radius: 8px;
  background: white;
  color: #4e5969;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
}

.form-cancel:hover {
  border-color: #c9cdd4;
  color: #1d2129;
}

.form-confirm {
  padding: 8px 20px;
  border: none;
  border-radius: 8px;
  background: linear-gradient(135deg, #165DFF 0%, #4080FF 100%);
  color: white;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  box-shadow: 0 2px 6px rgba(22, 93, 255, 0.25);
}

.form-confirm:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(22, 93, 255, 0.35);
}

.form-row {
  margin-bottom: 14px;
}

.form-row.two-col {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  margin-bottom: 14px;
}

.form-row.two-col > div {
  min-width: 0;
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
.form-select-small,
.form-textarea {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #dcdfe6;
  border-radius: 6px;
  font-size: 14px;
  color: #1d2129;
  transition: border-color 0.2s;
  box-sizing: border-box;
  font-family: inherit;
}

.form-textarea {
  resize: vertical;
  line-height: 1.5;
}

.form-input:focus,
.form-select-small:focus,
.form-textarea:focus {
  outline: none;
  border-color: #165DFF;
  box-shadow: 0 0 0 3px rgba(22, 93, 255, 0.1);
}

.feature-input-row {
  display: flex;
  gap: 8px;
  align-items: center;
}

.feature-input-row .form-select-small {
  width: 90px;
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
}

/* ==================== 导入结果弹窗 ==================== */
.import-result-dialog {
  width: 520px;
}

/* ==================== 设备详情弹窗 ==================== */
.detail-modal {
  width: 660px;
  max-height: 84vh;
  display: flex;
  flex-direction: column;
}

.detail-modal .dialog-body {
  overflow-y: auto;
}

.detail-modal .detail-grid {
  grid-template-columns: 1fr 1fr;
}

.detail-modal .detail-label {
  width: 100px;
}

.import-summary {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 14px;
  background: #f2f7ff;
  border: 1px solid #d6e4ff;
  border-radius: 8px;
  margin-bottom: 14px;
  font-size: 14px;
  font-weight: 600;
}

.import-ok {
  color: #00b42a;
}

.import-fail {
  color: #f53f3f;
}

.import-ok-sub {
  color: #86909c;
  font-weight: 500;
}

.import-error-list {
  max-height: 320px;
  overflow-y: auto;
  padding: 4px;
  background: #fafbfc;
  border: 1px solid #f0f0f0;
  border-radius: 6px;
}

.import-error-item {
  padding: 8px 10px;
  font-size: 13px;
  color: #4e5969;
  line-height: 1.5;
  border-bottom: 1px dashed #eef0f3;
}

.import-error-item:last-child {
  border-bottom: none;
}

.import-all-ok {
  padding: 16px;
  text-align: center;
  color: #4e5969;
  font-size: 13px;
  background: #fafbfc;
  border-radius: 6px;
}
</style>
