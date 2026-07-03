<script setup>
import { ref } from 'vue';
import { ElMessage } from 'element-plus';
import GlassCard from '@/components/GlassCard.vue';
import FileUploader from '@/components/FileUploader.vue';
import { useLogsStore } from '@/stores/logs';
import CryptoJS from 'crypto-js';

const logsStore = useLogsStore();
const fileList = ref([]);
const decryptProgress = ref(0);
const isDecrypting = ref(false);

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

const readFile = (file) => {
  return new Promise((resolve) => {
    const reader = new FileReader();
    reader.onload = (e) => resolve(e.target.result);
    reader.readAsText(file, 'UTF-8');
  });
};

const handleFilesSelected = (files) => {
  const newFiles = files.map(file => ({
    id: Date.now() + Math.random(),
    name: file.name,
    size: (file.size / 1024).toFixed(2) + ' KB',
    status: 'pending',
    progress: 0,
    rawContent: '',
    decryptedContent: '',
    file: file
  }));
  fileList.value = [...fileList.value, ...newFiles];
  logsStore.addLog('上传', '日志解密', `上传文件: ${files.map(f => f.name).join(', ')}`);
};

const handleBatchDecrypt = async () => {
  const pendingFiles = fileList.value.filter(f => f.status === 'pending');
  if (pendingFiles.length === 0) {
    ElMessage.warning('没有待解密的文件');
    return;
  }

  isDecrypting.value = true;
  decryptProgress.value = 0;

  for (let i = 0; i < pendingFiles.length; i++) {
    const file = pendingFiles[i];
    file.status = 'decrypting';

    try {
      const rawContent = await readFile(file.file);
      const decryptedContent = decryptLog(rawContent);
      file.rawContent = rawContent;
      file.decryptedContent = decryptedContent;
      file.status = 'decrypted';
    } catch (e) {
      file.status = 'failed';
    }

    decryptProgress.value = Math.round(((i + 1) / pendingFiles.length) * 100);
  }

  isDecrypting.value = false;
  logsStore.addLog('解密', '日志解密', `批量解密完成，共 ${pendingFiles.length} 个文件`);
  ElMessage.success(`解密完成，共 ${pendingFiles.length} 个文件`);
};

const exportSingleFile = (file) => {
  if (file.status !== 'decrypted' || !file.decryptedContent) {
    ElMessage.warning('该文件尚未解密');
    return;
  }
  
  const blob = new Blob([file.decryptedContent], { type: 'text/plain;charset=utf-8;' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = '解密_' + file.name;
  a.click();
  URL.revokeObjectURL(url);
  
  file.exported = true;
  logsStore.addLog('导出', '日志解密', `导出文件: 解密_${file.name}`);
  ElMessage.success(`导出成功: 解密_${file.name}`);
};

const exportAllFiles = async () => {
  const decryptedFiles = fileList.value.filter(f => f.status === 'decrypted' && f.decryptedContent);
  if (decryptedFiles.length === 0) {
    ElMessage.warning('没有已解密的文件');
    return;
  }
  
  const JSZip = await import('jszip');
  const zip = new JSZip();
  
  decryptedFiles.forEach(file => {
    zip.file('解密_' + file.name, file.decryptedContent);
    file.exported = true;
  });
  
  const blob = await zip.generateAsync({ type: 'blob' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = '测试工具_解密结果.zip';
  a.click();
  URL.revokeObjectURL(url);
  
  logsStore.addLog('导出', '日志解密', `批量导出ZIP: 测试工具_解密结果.zip`);
  ElMessage.success(`导出成功: 测试工具_解密结果.zip`);
};

const clearAllFiles = () => {
  fileList.value = [];
  decryptProgress.value = 0;
  logsStore.addLog('清空', '日志解密', '清空文件列表');
};

const removeFile = (fileId) => {
  fileList.value = fileList.value.filter(f => f.id !== fileId);
};
</script>

<template>
  <div class="log-decrypt">
    <GlassCard title="日志解密工具">
      <FileUploader @files-selected="handleFilesSelected" />
      
      <div class="action-area">
        <button 
          class="action-btn decrypt" 
          :disabled="isDecrypting || fileList.filter(f => f.status === 'pending').length === 0"
          @click="handleBatchDecrypt"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
            <path d="M9 12l2 2 4-4" />
          </svg>
          <span>{{ isDecrypting ? '解密中...' : '批量解密' }}</span>
        </button>
        <button 
          class="action-btn export" 
          :disabled="isDecrypting || fileList.filter(f => f.status === 'decrypted').length === 0"
          @click="exportAllFiles"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
            <polyline points="7 10 12 15 17 10" />
            <line x1="12" y1="15" x2="12" y2="3" />
          </svg>
          <span>导出全部ZIP</span>
        </button>
        <button 
          class="action-btn clear" 
          :disabled="fileList.length === 0"
          @click="clearAllFiles"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="3 6 5 6 21 6" />
            <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
          </svg>
          <span>清空列表</span>
        </button>
      </div>
      
      <div v-if="isDecrypting || decryptProgress > 0" class="progress-area">
        <div class="progress-label">解密进度</div>
        <div class="progress-bar">
          <div class="progress-fill" :style="{ width: decryptProgress + '%' }"></div>
        </div>
        <div class="progress-text">{{ decryptProgress }}%</div>
      </div>
      
      <div class="file-count">文件总数: {{ fileList.length }} 个 | 已解密: {{ fileList.filter(f => f.status === 'decrypted').length }} 个</div>
    </GlassCard>
    
    <GlassCard title="文件列表" style="margin-top: 20px;">
      <div v-if="fileList.length === 0" class="empty-state">
        <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
          <polyline points="14 2 14 8 20 8" />
          <line x1="16" y1="13" x2="8" y2="13" />
          <line x1="16" y1="17" x2="8" y2="17" />
          <polyline points="10 9 9 9 8 9" />
        </svg>
        <p>暂无上传的文件</p>
        <p class="hint">支持拖拽或点击上传 .log / .txt 文件</p>
      </div>
      
      <div v-else class="files-container">
        <div v-for="file in fileList" :key="file.id" class="file-item">
          <div class="file-icon" :class="file.status">
            <svg v-if="file.status === 'pending'" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
              <polyline points="14 2 14 8 20 8" />
            </svg>
            <svg v-else-if="file.status === 'decrypting'" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="spin">
              <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
              <path d="M9 12l2 2 4-4" />
            </svg>
            <svg v-else-if="file.status === 'failed'" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="12" r="10" />
              <line x1="15" y1="9" x2="9" y2="15" />
              <line x1="9" y1="9" x2="15" y2="15" />
            </svg>
            <svg v-else width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" />
              <polyline points="22 4 12 14.01 9 11.01" />
            </svg>
          </div>
          <div class="file-info">
            <div class="file-name">{{ file.name }}</div>
            <div class="file-size">{{ file.size }}</div>
          </div>
          <div class="file-status">
            <span v-if="file.status === 'pending'" class="status-tag pending">待解密</span>
            <span v-else-if="file.status === 'decrypting'" class="status-tag decrypting">解密中</span>
            <span v-else-if="file.status === 'failed'" class="status-tag failed">解密失败</span>
            <span v-else class="status-tag decrypted">已解密</span>
          </div>
          <div class="file-actions">
            <button 
              v-if="file.status === 'decrypted'" 
              class="export-single-btn" 
              @click="exportSingleFile(file)"
              title="下载解密后的文件"
            >
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
                <polyline points="7 10 12 15 17 10" />
                <line x1="12" y1="15" x2="12" y2="3" />
              </svg>
            </button>
            <button class="remove-btn" @click="removeFile(file.id)" title="删除文件">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <line x1="18" y1="6" x2="6" y2="18" />
                <line x1="6" y1="6" x2="18" y2="18" />
              </svg>
            </button>
          </div>
        </div>
      </div>
    </GlassCard>
  </div>
</template>

<style scoped>
.log-decrypt {
  max-width: 100%;
}

.action-area {
  display: flex;
  gap: 12px;
  margin-top: 20px;
  flex-wrap: wrap;
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 20px;
  border: none;
  border-radius: 8px;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
  
  &.decrypt {
    background: linear-gradient(135deg, #165DFF 0%, #0F4CD0 100%);
    color: #FFFFFF;
    
    &:hover:not(:disabled) {
      transform: translateY(-2px);
      box-shadow: 0 4px 12px rgba(22, 93, 255, 0.3);
    }
  }
  
  &.export {
    background: rgba(0, 180, 42, 0.1);
    border: 1px solid rgba(0, 180, 42, 0.3);
    color: #00B42A;
    
    &:hover:not(:disabled) {
      background: rgba(0, 180, 42, 0.2);
    }
  }
  
  &.clear {
    background: rgba(245, 63, 63, 0.1);
    border: 1px solid rgba(245, 63, 63, 0.3);
    color: #F53F3F;
    
    &:hover:not(:disabled) {
      background: rgba(245, 63, 63, 0.2);
    }
  }
  
  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
}

.progress-area {
  margin-top: 20px;
  padding: 16px;
  background: rgba(15, 23, 42, 0.5);
  border-radius: 8px;
}

.progress-label {
  font-size: 13px;
  color: #94A3B8;
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
  color: #64748B;
  margin-top: 8px;
  text-align: right;
}

.file-count {
  margin-top: 12px;
  font-size: 12px;
  color: #64748B;
}

.empty-state {
  text-align: center;
  padding: 40px;
  color: #64748B;
  
  svg {
    margin-bottom: 16px;
  }
  
  .hint {
    font-size: 12px;
    margin-top: 8px;
  }
}

.files-container {
  max-height: 400px;
  overflow-y: auto;
}

.file-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  border-bottom: 1px solid rgba(22, 93, 255, 0.1);
  
  &:hover {
    background: rgba(22, 93, 255, 0.05);
  }
}

.file-icon {
  flex-shrink: 0;
  
  &.pending {
    color: #FF7D00;
  }
  
  &.decrypting {
    color: #165DFF;
  }
  
  &.decrypted {
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
  color: #FFFFFF;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-size {
  font-size: 12px;
  color: #64748B;
}

.file-status {
  flex-shrink: 0;
}

.status-tag {
  font-size: 12px;
  padding: 4px 10px;
  border-radius: 4px;
  
  &.pending {
    background: rgba(255, 125, 0, 0.1);
    color: #FF7D00;
  }
  
  &.decrypting {
    background: rgba(22, 93, 255, 0.1);
    color: #165DFF;
  }
  
  &.decrypted {
    background: rgba(0, 180, 42, 0.1);
    color: #00B42A;
  }
  
  &.failed {
    background: rgba(245, 63, 63, 0.1);
    color: #F53F3F;
  }
}

.file-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

.export-single-btn {
  background: rgba(0, 180, 42, 0.1);
  border: 1px solid rgba(0, 180, 42, 0.3);
  color: #00B42A;
  border-radius: 4px;
  padding: 6px;
  cursor: pointer;
  transition: all 0.2s;
  
  &:hover {
    background: rgba(0, 180, 42, 0.2);
  }
}

.remove-btn {
  background: none;
  border: none;
  color: #64748B;
  cursor: pointer;
  padding: 6px;
  transition: color 0.2s;
  
  &:hover {
    color: #F53F3F;
  }
}

@media screen and (max-width: 768px) {
  .action-area {
    flex-direction: column;
  }
  
  .action-btn {
    justify-content: center;
  }
  
  .file-item {
    flex-wrap: wrap;
  }
  
  .file-actions {
    margin-top: 8px;
  }
}
</style>