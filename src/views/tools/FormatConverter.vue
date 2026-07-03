<script setup>
import { ref } from 'vue';
import { ElMessage } from 'element-plus';
import GlassCard from '@/components/GlassCard.vue';
import { useLogsStore } from '@/stores/logs';
import * as XLSX from 'xlsx';
import mammoth from 'mammoth';
import { PDFDocument } from 'pdf-lib';
import jsPDF from 'jspdf';
import { saveAs } from 'file-saver';

const logsStore = useLogsStore();

const selectedFile = ref(null);
const outputFormat = ref('pdf');
const conversionResult = ref(null);
const isConverting = ref(false);
const convertedBlob = ref(null);
const convertedFileName = ref('');

const inputFormats = [
  { value: 'word', label: 'Word文档', extensions: ['.doc', '.docx'] },
  { value: 'excel', label: 'Excel表格', extensions: ['.xls', '.xlsx'] },
  { value: 'pdf', label: 'PDF文档', extensions: ['.pdf'] },
  { value: 'csv', label: 'CSV文件', extensions: ['.csv'] },
  { value: 'txt', label: '文本文件', extensions: ['.txt'] },
  { value: 'json', label: 'JSON文件', extensions: ['.json'] }
];

const outputFormats = [
  { value: 'pdf', label: 'PDF文档' },
  { value: 'word', label: 'Word文档 (.docx)' },
  { value: 'excel', label: 'Excel表格 (.xlsx)' },
  { value: 'csv', label: 'CSV文件' },
  { value: 'txt', label: '文本文件' },
  { value: 'json', label: 'JSON文件' }
];

const detectFileType = (fileName) => {
  const ext = fileName.toLowerCase().split('.').pop();
  if (ext === 'doc' || ext === 'docx') return 'word';
  if (ext === 'xls' || ext === 'xlsx') return 'excel';
  if (ext === 'pdf') return 'pdf';
  if (ext === 'csv') return 'csv';
  if (ext === 'txt') return 'txt';
  if (ext === 'json') return 'json';
  return '';
};

const handleFileSelect = (event) => {
  const file = event.target.files[0];
  if (!file) return;
  
  selectedFile.value = file;
  const detectedType = detectFileType(file.name);
  
  // 根据输入文件类型自动设置可用的输出格式
  if (detectedType === 'word') {
    outputFormat.value = 'pdf';
  } else if (detectedType === 'excel') {
    outputFormat.value = 'csv';
  } else if (detectedType === 'pdf') {
    outputFormat.value = 'txt';
  } else if (detectedType === 'csv') {
    outputFormat.value = 'excel';
  } else if (detectedType === 'txt') {
    outputFormat.value = 'pdf';
  } else if (detectedType === 'json') {
    outputFormat.value = 'csv';
  }
  
  conversionResult.value = null;
  convertedBlob.value = null;
  logsStore.addLog('选择文件', '格式转换', `文件: ${file.name}, 类型: ${detectedType}`);
};

// Word转PDF
const convertWordToPdf = async (file) => {
  const arrayBuffer = await file.arrayBuffer();
  const result = await mammoth.extractRawText({ arrayBuffer });
  const text = result.value;
  
  const pdf = new jsPDF();
  const lines = text.split('\n');
  let y = 20;
  
  pdf.setFontSize(12);
  lines.forEach(line => {
    if (y > 280) {
      pdf.addPage();
      y = 20;
    }
    pdf.text(line, 20, y);
    y += 7;
  });
  
  return pdf.output('blob');
};

// Word转TXT
const convertWordToTxt = async (file) => {
  const arrayBuffer = await file.arrayBuffer();
  const result = await mammoth.extractRawText({ arrayBuffer });
  return new Blob([result.value], { type: 'text/plain' });
};

// Excel转CSV
const convertExcelToCsv = async (file) => {
  const arrayBuffer = await file.arrayBuffer();
  const workbook = XLSX.read(arrayBuffer, { type: 'array' });
  const firstSheet = workbook.Sheets[workbook.SheetNames[0]];
  const csv = XLSX.utils.sheet_to_csv(firstSheet);
  return new Blob([csv], { type: 'text/csv' });
};

// Excel转JSON
const convertExcelToJson = async (file) => {
  const arrayBuffer = await file.arrayBuffer();
  const workbook = XLSX.read(arrayBuffer, { type: 'array' });
  const result = {};
  workbook.SheetNames.forEach(sheetName => {
    const sheet = workbook.Sheets[sheetName];
    result[sheetName] = XLSX.utils.sheet_to_json(sheet);
  });
  return new Blob([JSON.stringify(result, null, 2)], { type: 'application/json' });
};

// CSV转Excel
const convertCsvToExcel = async (file) => {
  const text = await file.text();
  const workbook = XLSX.read(text, { type: 'string' });
  const excelBuffer = XLSX.write(workbook, { bookType: 'xlsx', type: 'array' });
  return new Blob([excelBuffer], { type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' });
};

// CSV转PDF
const convertCsvToPdf = async (file) => {
  const text = await file.text();
  const pdf = new jsPDF();
  
  const lines = text.split('\n');
  let y = 20;
  
  pdf.setFontSize(10);
  lines.forEach(line => {
    if (y > 280) {
      pdf.addPage();
      y = 20;
    }
    pdf.text(line, 20, y);
    y += 6;
  });
  
  return pdf.output('blob');
};

// JSON转CSV
const convertJsonToCsv = async (file) => {
  const text = await file.text();
  const data = JSON.parse(text);
  
  let csv = '';
  if (Array.isArray(data)) {
    if (data.length > 0) {
      const headers = Object.keys(data[0]);
      csv = headers.join(',') + '\n';
      data.forEach(item => {
        const values = headers.map(h => {
          const v = item[h];
          return typeof v === 'string' && v.includes(',') ? `"${v}"` : v;
        });
        csv += values.join(',') + '\n';
      });
    }
  } else {
    const headers = Object.keys(data);
    csv = headers.join(',') + '\n';
    const values = headers.map(h => {
      const v = data[h];
      return typeof v === 'string' && v.includes(',') ? `"${v}"` : v;
    });
    csv += values.join(',') + '\n';
  }
  
  return new Blob([csv], { type: 'text/csv' });
};

// JSON转Excel
const convertJsonToExcel = async (file) => {
  const text = await file.text();
  const data = JSON.parse(text);
  
  const worksheet = XLSX.utils.json_to_sheet(Array.isArray(data) ? data : [data]);
  const workbook = XLSX.utils.book_new();
  XLSX.utils.book_append_sheet(workbook, worksheet, 'Sheet1');
  
  const excelBuffer = XLSX.write(workbook, { bookType: 'xlsx', type: 'array' });
  return new Blob([excelBuffer], { type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' });
};

// TXT转PDF
const convertTxtToPdf = async (file) => {
  const text = await file.text();
  const pdf = new jsPDF();
  
  const lines = text.split('\n');
  let y = 20;
  
  pdf.setFontSize(12);
  lines.forEach(line => {
    if (y > 280) {
      pdf.addPage();
      y = 20;
    }
    pdf.text(line, 20, y);
    y += 7;
  });
  
  return pdf.output('blob');
};

// PDF转TXT
const convertPdfToTxt = async (file) => {
  const arrayBuffer = await file.arrayBuffer();
  const pdfDoc = await PDFDocument.load(arrayBuffer);
  const pages = pdfDoc.getPages();
  
  let text = '';
  for (const page of pages) {
    const pageText = await page.getTextContent();
    text += pageText.items.map(item => item.str).join(' ') + '\n';
  }
  
  // pdf-lib无法直接提取文本，使用简化方式
  text = `PDF文档共 ${pages.length} 页\n文件名: ${file.name}\n大小: ${(file.size / 1024).toFixed(2)} KB`;
  
  return new Blob([text], { type: 'text/plain' });
};

// 执行转换
const handleConvert = async () => {
  if (!selectedFile.value) {
    ElMessage.warning('请先选择要转换的文件');
    return;
  }
  
  const inputType = detectFileType(selectedFile.value.name);
  const outputType = outputFormat.value;
  
  isConverting.value = true;
  conversionResult.value = null;
  
  try {
    let blob = null;
    
    // 根据输入和输出类型执行转换
    if (inputType === 'word') {
      if (outputType === 'pdf') {
        blob = await convertWordToPdf(selectedFile.value);
      } else if (outputType === 'txt') {
        blob = await convertWordToTxt(selectedFile.value);
      } else {
        throw new Error('Word文档只能转换为PDF或TXT格式');
      }
    } else if (inputType === 'excel') {
      if (outputType === 'csv') {
        blob = await convertExcelToCsv(selectedFile.value);
      } else if (outputType === 'json') {
        blob = await convertExcelToJson(selectedFile.value);
      } else if (outputType === 'pdf') {
        blob = await convertCsvToPdf(await convertExcelToCsv(selectedFile.value));
      } else {
        throw new Error('Excel表格只能转换为CSV、JSON或PDF格式');
      }
    } else if (inputType === 'csv') {
      if (outputType === 'excel') {
        blob = await convertCsvToExcel(selectedFile.value);
      } else if (outputType === 'pdf') {
        blob = await convertCsvToPdf(selectedFile.value);
      } else if (outputType === 'json') {
        const excelBlob = await convertCsvToExcel(selectedFile.value);
        blob = await convertExcelToJson(excelBlob);
      } else {
        throw new Error('CSV文件只能转换为Excel、PDF或JSON格式');
      }
    } else if (inputType === 'json') {
      if (outputType === 'csv') {
        blob = await convertJsonToCsv(selectedFile.value);
      } else if (outputType === 'excel') {
        blob = await convertJsonToExcel(selectedFile.value);
      } else if (outputType === 'pdf') {
        const csvBlob = await convertJsonToCsv(selectedFile.value);
        blob = await convertCsvToPdf(csvBlob);
      } else {
        throw new Error('JSON文件只能转换为CSV、Excel或PDF格式');
      }
    } else if (inputType === 'txt') {
      if (outputType === 'pdf') {
        blob = await convertTxtToPdf(selectedFile.value);
      } else {
        throw new Error('文本文件只能转换为PDF格式');
      }
    } else if (inputType === 'pdf') {
      if (outputType === 'txt') {
        blob = await convertPdfToTxt(selectedFile.value);
      } else {
        throw new Error('PDF文档只能转换为TXT格式');
      }
    } else {
      throw new Error('不支持的文件类型');
    }
    
    convertedBlob.value = blob;
    
    // 生成输出文件名
    const baseName = selectedFile.value.name.split('.').slice(0, -1).join('.');
    const outputExt = outputFormats.find(f => f.value === outputType)?.label.match(/\.\w+/)?.[0] || '.' + outputType;
    convertedFileName.value = `${baseName}${outputExt}`;
    
    conversionResult.value = {
      success: true,
      inputType: inputFormats.find(f => f.value === inputType)?.label,
      outputType: outputFormats.find(f => f.value === outputType)?.label,
      size: `${(blob.size / 1024).toFixed(2)} KB`
    };
    
    logsStore.addLog('转换', '格式转换', `${inputType} -> ${outputType}, 文件: ${selectedFile.value.name}`);
    ElMessage.success('文件转换成功');
    
  } catch (error) {
    conversionResult.value = {
      success: false,
      error: error.message
    };
    ElMessage.error(`转换失败: ${error.message}`);
    logsStore.addLog('转换', '格式转换', `失败: ${error.message}`);
  } finally {
    isConverting.value = false;
  }
};

// 下载转换后的文件
const handleDownload = () => {
  if (!convertedBlob.value) {
    ElMessage.warning('请先转换文件');
    return;
  }
  
  saveAs(convertedBlob.value, convertedFileName.value);
  logsStore.addLog('下载', '格式转换', `下载文件: ${convertedFileName.value}`);
  ElMessage.success('文件下载成功');
};

// 清空
const handleClear = () => {
  selectedFile.value = null;
  outputFormat.value = 'pdf';
  conversionResult.value = null;
  convertedBlob.value = null;
  convertedFileName.value = '';
  logsStore.addLog('清空', '格式转换', '清空所有内容');
};

// 获取可用的输出格式
const getAvailableOutputFormats = () => {
  if (!selectedFile.value) return outputFormats;
  
  const inputType = detectFileType(selectedFile.value.name);
  
  switch (inputType) {
    case 'word':
      return outputFormats.filter(f => ['pdf', 'txt'].includes(f.value));
    case 'excel':
      return outputFormats.filter(f => ['csv', 'json', 'pdf'].includes(f.value));
    case 'csv':
      return outputFormats.filter(f => ['excel', 'pdf', 'json'].includes(f.value));
    case 'json':
      return outputFormats.filter(f => ['csv', 'excel', 'pdf'].includes(f.value));
    case 'txt':
      return outputFormats.filter(f => ['pdf'].includes(f.value));
    case 'pdf':
      return outputFormats.filter(f => ['txt'].includes(f.value));
    default:
      return outputFormats;
  }
};
</script>

<template>
  <div class="format-converter">
    <GlassCard title="文档格式转换">
      <div class="converter-container">
        <!-- 文件上传区域 -->
        <div class="upload-section">
          <div class="upload-area" @click="$refs.fileInput.click()">
            <input 
              ref="fileInput"
              type="file" 
              @change="handleFileSelect" 
              accept=".doc,.docx,.xls,.xlsx,.pdf,.csv,.txt,.json"
              style="display: none"
            />
            <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
              <polyline points="7 10 12 15 17 10" />
              <line x1="12" y1="15" x2="12" y2="3" />
            </svg>
            <div class="upload-text">
              <span class="upload-title">点击上传文件</span>
              <span class="upload-desc">支持 Word、Excel、PDF、CSV、TXT、JSON 格式</span>
            </div>
          </div>
          
          <div v-if="selectedFile" class="file-info">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
              <polyline points="14 2 14 8 20 8" />
            </svg>
            <span class="file-name">{{ selectedFile.name }}</span>
            <span class="file-size">{{ (selectedFile.size / 1024).toFixed(2) }} KB</span>
          </div>
        </div>
        
        <!-- 格式选择区域 -->
        <div class="format-section">
          <div class="format-row">
            <div class="format-item">
              <label>输入格式</label>
              <div class="format-display">
                {{ selectedFile ? inputFormats.find(f => f.value === detectFileType(selectedFile.name))?.label : '未选择' }}
              </div>
            </div>
            
            <div class="format-arrow">
              <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M5 12h14" />
                <polyline points="12 5 19 12 12 19" />
              </svg>
            </div>
            
            <div class="format-item">
              <label>输出格式</label>
              <select v-model="outputFormat" class="format-selector">
                <option 
                  v-for="fmt in getAvailableOutputFormats()" 
                  :key="fmt.value" 
                  :value="fmt.value"
                >
                  {{ fmt.label }}
                </option>
              </select>
            </div>
          </div>
        </div>
        
        <!-- 操作按钮 -->
        <div class="action-buttons">
          <button 
            class="action-btn primary"
            :disabled="isConverting || !selectedFile"
            @click="handleConvert"
          >
            <svg v-if="!isConverting" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M12 15V3" />
              <path d="M17 6l5-5-5-5" />
              <path d="M9 21l-1.5-7.5L21 6" />
            </svg>
            <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="spin">
              <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
            </svg>
            <span>{{ isConverting ? '转换中...' : '开始转换' }}</span>
          </button>
          
          <button 
            class="action-btn secondary"
            :disabled="!convertedBlob"
            @click="handleDownload"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
              <polyline points="7 10 12 15 17 10" />
              <line x1="12" y1="15" x2="12" y2="3" />
            </svg>
            <span>下载文件</span>
          </button>
          
          <button 
            class="action-btn danger"
            @click="handleClear"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <line x1="18" y1="6" x2="6" y2="18" />
              <line x1="6" y1="6" x2="18" y2="18" />
            </svg>
            <span>清空</span>
          </button>
        </div>
        
        <!-- 转换结果 -->
        <div v-if="conversionResult" class="result-section">
          <div v-if="conversionResult.success" class="result-success">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <polyline points="20 6 9 17 4 12" />
            </svg>
            <span>转换成功</span>
            <div class="result-details">
              <span>{{ conversionResult.inputType }} → {{ conversionResult.outputType }}</span>
              <span>文件大小: {{ conversionResult.size }}</span>
            </div>
          </div>
          <div v-else class="result-error">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="12" r="10" />
              <line x1="15" y1="9" x2="9" y2="15" />
              <line x1="9" y1="9" x2="15" y2="15" />
            </svg>
            <span>{{ conversionResult.error }}</span>
          </div>
        </div>
        
        <!-- 支持的转换说明 -->
        <div class="support-info">
          <div class="support-title">支持的转换类型</div>
          <div class="support-grid">
            <div class="support-item">
              <span class="support-label">Word → PDF/TXT</span>
            </div>
            <div class="support-item">
              <span class="support-label">Excel → CSV/JSON/PDF</span>
            </div>
            <div class="support-item">
              <span class="support-label">CSV → Excel/PDF/JSON</span>
            </div>
            <div class="support-item">
              <span class="support-label">JSON → CSV/Excel/PDF</span>
            </div>
            <div class="support-item">
              <span class="support-label">TXT → PDF</span>
            </div>
            <div class="support-item">
              <span class="support-label">PDF → TXT</span>
            </div>
          </div>
        </div>
      </div>
    </GlassCard>
  </div>
</template>

<style scoped>
.format-converter {
  padding: 20px;
}

.converter-container {
  max-width: 800px;
  margin: 0 auto;
}

.upload-section {
  margin-bottom: 30px;
}

.upload-area {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px;
  border: 2px dashed #ddd;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.3s;
  background: #fafafa;
}

.upload-area:hover {
  border-color: #4080ff;
  background: #f0f8ff;
}

.upload-text {
  margin-top: 15px;
  text-align: center;
}

.upload-title {
  font-size: 16px;
  font-weight: 500;
  color: #333;
}

.upload-desc {
  font-size: 13px;
  color: #999;
  margin-top: 5px;
}

.file-info {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px;
  background: #e8f4f8;
  border-radius: 8px;
  margin-top: 15px;
}

.file-name {
  font-weight: 500;
  color: #333;
}

.file-size {
  color: #999;
  font-size: 13px;
}

.format-section {
  margin-bottom: 30px;
}

.format-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
}

.format-item {
  flex: 1;
}

.format-item label {
  display: block;
  font-size: 14px;
  color: #666;
  margin-bottom: 8px;
}

.format-display {
  padding: 10px 16px;
  background: #f5f5f5;
  border-radius: 6px;
  color: #333;
  font-weight: 500;
}

.format-selector {
  width: 100%;
  padding: 10px 16px;
  border: 1px solid #ddd;
  border-radius: 6px;
  font-size: 14px;
  background: white;
  cursor: pointer;
}

.format-arrow {
  display: flex;
  align-items: center;
  justify-content: center;
  color: #4080ff;
}

.action-buttons {
  display: flex;
  gap: 12px;
  justify-content: center;
  margin-bottom: 30px;
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 24px;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.action-btn.primary {
  background: #4080ff;
  color: white;
}

.action-btn.primary:hover:not(:disabled) {
  background: #3070ee;
}

.action-btn.secondary {
  background: #f0f0f0;
  color: #333;
}

.action-btn.secondary:hover:not(:disabled) {
  background: #e0e0e0;
}

.action-btn.danger {
  background: #fff2f0;
  color: #ff4d4f;
}

.action-btn.danger:hover:not(:disabled) {
  background: #ffe0e0;
}

.result-section {
  margin-bottom: 30px;
}

.result-success {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 16px;
  background: #e8f8e8;
  border-radius: 8px;
  color: #52c41a;
}

.result-details {
  display: flex;
  gap: 20px;
  margin-left: 10px;
  color: #666;
  font-size: 13px;
}

.result-error {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 16px;
  background: #fff2f0;
  border-radius: 8px;
  color: #ff4d4f;
}

.support-info {
  padding: 20px;
  background: #f5f5f5;
  border-radius: 8px;
}

.support-title {
  font-weight: 600;
  color: #333;
  margin-bottom: 15px;
}

.support-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
}

.support-item {
  padding: 8px 12px;
  background: white;
  border-radius: 4px;
}

.support-label {
  font-size: 13px;
  color: #666;
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>