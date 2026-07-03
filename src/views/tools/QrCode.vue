<script setup>import { ref, computed } from 'vue';
import { ElMessage } from 'element-plus';
import GlassCard from '@/components/GlassCard.vue';
import QrcodeVue from 'qrcode-vue3';
import { useLogsStore } from '@/stores/logs';
const logsStore = useLogsStore();
const qrContent = ref('');
const qrSize = ref(200);
const isGenerating = ref(false);
const sizeOptions = [
 { value: 100, label: '100x100' },
 { value: 150, label: '150x150' },
 { value: 200, label: '200x200' },
 { value: 250, label: '250x250' },
 { value: 300, label: '300x300' },
 { value: 400, label: '400x400' }
];
const handleGenerate = () => {
 if (!qrContent.value.trim()) {
 ElMessage.warning('请输入二维码内容');
 return;
 }
 isGenerating.value = true;
 setTimeout(() => {
 isGenerating.value = false;
 }, 500);
 logsStore.addLog('生成', '二维码', `内容: ${qrContent.value.substring(0, 20)}${qrContent.value.length > 20 ? '...' : ''}`);
 ElMessage.success('二维码生成成功');
};
const handleDownload = () => {
 if (!qrContent.value.trim()) {
 ElMessage.warning('请先生成二维码');
 return;
 }
 const canvas = document.querySelector('.qrcode-canvas');
 if (canvas) {
 const url = canvas.toDataURL('image/png');
 const a = document.createElement('a');
 a.href = url;
 a.download = `qrcode_${Date.now()}.png`;
 a.click();
 logsStore.addLog('下载', '二维码', '下载二维码图片');
 ElMessage.success('二维码下载成功');
 }
};
const handleClear = () => {
 qrContent.value = '';
 logsStore.addLog('清空', '二维码', '清空内容');
};
const sampleContents = [
 'DEV-2024-001',
 'AUTH-LIC-2024001',
 'TASK-00123',
 'BATCH-A001'
];
const setSample = (content) => {
 qrContent.value = content;
};
</script>

<template>
  <div class="qr-code">
    <GlassCard title="二维码生成工具">
      <div class="qr-section">
        <div class="input-section">
          <div class="input-label">二维码内容</div>
          <textarea 
            v-model="qrContent"
            class="qr-input"
            placeholder="输入设备号、授权信息、任务ID等内容..."
            rows="3"
          ></textarea>
          
          <div class="sample-tags">
            <span class="sample-label">快速填入:</span>
            <button 
              v-for="sample in sampleContents" 
              :key="sample" 
              class="sample-tag"
              @click="setSample(sample)"
            >
              {{ sample }}
            </button>
          </div>
          
          <div class="size-control">
            <div class="size-label">二维码尺寸: {{ qrSize }}x{{ qrSize }}</div>
            <input 
              type="range" 
              v-model="qrSize"
              min="100"
              max="400"
              step="50"
              class="size-slider"
            />
            <div class="size-options">
              <button 
                v-for="opt in sizeOptions" 
                :key="opt.value"
                class="size-option"
                :class="{ 'active': qrSize === opt.value }"
                @click="qrSize = opt.value"
              >
                {{ opt.label }}
              </button>
            </div>
          </div>
          
          <div class="action-buttons">
            <button 
              class="generate-btn" 
              :disabled="isGenerating"
              @click="handleGenerate"
            >
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <rect x="3" y="3" width="18" height="18" rx="2" ry="2" />
                <path d="M3 9h18" />
                <path d="M9 21V3" />
              </svg>
              <span>{{ isGenerating ? '生成中...' : '生成二维码' }}</span>
            </button>
            <button 
              class="download-btn" 
              :disabled="!qrContent"
              @click="handleDownload"
            >
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
                <polyline points="7 10 12 15 17 10" />
                <line x1="12" y1="15" x2="12" y2="3" />
              </svg>
              <span>下载PNG</span>
            </button>
            <button class="clear-btn" @click="handleClear">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M3 6h18" />
                <path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6" />
                <path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2" />
              </svg>
              <span>清空</span>
            </button>
          </div>
        </div>
        
        <div class="preview-section">
          <div class="preview-label">实时预览</div>
          <div class="qr-container">
            <div v-if="qrContent" class="qr-wrapper">
              <QrcodeVue 
                :value="qrContent" 
                :size="qrSize"
                level="H"
                class="qrcode-canvas"
              />
            </div>
            <div v-else class="qr-placeholder">
              <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                <rect x="3" y="3" width="18" height="18" rx="2" ry="2" />
                <path d="M3 9h18" />
                <path d="M9 21V3" />
              </svg>
              <p>二维码预览区域</p>
            </div>
          </div>
          <div v-if="qrContent" class="qr-info">
            <span>内容: {{ qrContent }}</span>
            <span>尺寸: {{ qrSize }}x{{ qrSize }}</span>
          </div>
        </div>
      </div>
    </GlassCard>
  </div>
</template>

<style scoped>
.qr-code {
  max-width: 100%;
}

.qr-section {
  display: flex;
  gap: 40px;
  flex-wrap: wrap;
}

.input-section {
  flex: 1;
  min-width: 300px;
}

.input-label {
  font-size: 13px;
  color: var(--text-secondary);
  margin-bottom: 8px;
}

.qr-input {
  width: 100%;
  background: var(--bg-input);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 12px;
  color: var(--text-primary);
  font-size: 14px;
  resize: vertical;
  outline: none;
  transition: border-color 0.2s;
  font-family: 'Consolas', monospace;
  
  &:focus {
    border-color: rgba(22, 93, 255, 0.6);
  }
  
  &::placeholder {
    color: var(--text-tertiary);
  }
}

.sample-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 12px;
}

.sample-label {
  font-size: 12px;
  color: var(--text-tertiary);
  align-self: center;
}

.sample-tag {
  padding: 6px 12px;
  background: rgba(22, 93, 255, 0.1);
  border: 1px solid rgba(22, 93, 255, 0.3);
  border-radius: 6px;
  color: #165DFF;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s;
  
  &:hover {
    background: rgba(22, 93, 255, 0.2);
  }
}

.size-control {
  margin-top: 20px;
}

.size-label {
  font-size: 13px;
  color: var(--text-secondary);
  margin-bottom: 12px;
}

.size-slider {
  width: 100%;
  height: 6px;
  background: rgba(100, 116, 139, 0.3);
  border-radius: 3px;
  outline: none;
  -webkit-appearance: none;
  
  &::-webkit-slider-thumb {
    -webkit-appearance: none;
    width: 18px;
    height: 18px;
    background: #165DFF;
    border-radius: 50%;
    cursor: pointer;
    transition: transform 0.2s;
    
    &:hover {
      transform: scale(1.2);
    }
  }
}

.size-options {
  display: flex;
  gap: 8px;
  margin-top: 12px;
  flex-wrap: wrap;
}

.size-option {
  padding: 6px 12px;
  background: rgba(100, 116, 139, 0.2);
  border: 1px solid rgba(100, 116, 139, 0.3);
  border-radius: 6px;
  color: var(--text-secondary);
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s;
  
  &.active {
    background: rgba(22, 93, 255, 0.2);
    border-color: #165DFF;
    color: #165DFF;
  }
  
  &:hover {
    background: rgba(100, 116, 139, 0.3);
  }
}

.action-buttons {
  display: flex;
  gap: 10px;
  margin-top: 24px;
  flex-wrap: wrap;
}

.generate-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 20px;
  background: linear-gradient(135deg, #165DFF 0%, #0F4CD0 100%);
  border: none;
  border-radius: 8px;
  color: #FFFFFF;
  font-size: 14px;
  font-weight: 500;
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

.download-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 20px;
  background: rgba(0, 180, 42, 0.1);
  border: 1px solid rgba(0, 180, 42, 0.3);
  border-radius: 8px;
  color: #00B42A;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  
  &:hover:not(:disabled) {
    background: rgba(0, 180, 42, 0.2);
  }
  
  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
}

.clear-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 20px;
  background: rgba(100, 116, 139, 0.2);
  border: 1px solid rgba(100, 116, 139, 0.3);
  border-radius: 8px;
  color: #94A3B8;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
  
  &:hover {
    background: rgba(100, 116, 139, 0.3);
  }
}

.preview-section {
  flex-shrink: 0;
}

.preview-label {
  font-size: 13px;
  color: var(--text-secondary);
  margin-bottom: 12px;
}

.qr-container {
  width: 400px;
  height: 400px;
  background: var(--bg-input);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.qr-wrapper {
  display: flex;
  align-items: center;
  justify-content: center;
}

.qr-placeholder {
  text-align: center;
  color: var(--text-tertiary);
  
  svg {
    margin-bottom: 12px;
  }
  
  p {
    font-size: 13px;
  }
}

.qr-info {
  display: flex;
  gap: 16px;
  margin-top: 16px;
  
  span {
    font-size: 12px;
    color: var(--text-tertiary);
  }
}

@media screen and (max-width: 768px) {
  .qr-section {
    flex-direction: column;
    align-items: center;
    gap: 24px;
  }
  
  .input-section {
    min-width: 100%;
  }
  
  .action-buttons {
    justify-content: center;
  }
  
  .qr-container {
    width: 300px;
    height: 300px;
  }
}

@media screen and (max-width: 480px) {
  .action-buttons {
    flex-direction: column;
    align-items: stretch;
  }
  
  .generate-btn,
  .download-btn,
  .clear-btn {
    justify-content: center;
    padding: 10px 16px;
    font-size: 13px;
  }
  
  .qr-container {
    width: 250px;
    height: 250px;
  }
  
  .qr-info {
    flex-direction: column;
    gap: 8px;
  }
  
  .size-options {
    gap: 6px;
  }
  
  .size-option {
    padding: 4px 10px;
    font-size: 11px;
  }
}
</style>
