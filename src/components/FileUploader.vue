<script setup>import { ref } from 'vue';
import { ElMessage } from 'element-plus';
const props = defineProps({
 multiple: {
 type: Boolean,
 default: true
 },
 accept: {
 type: String,
 default: '.log,.txt'
 }
});
const emit = defineEmits(['files-selected']);
const isDragging = ref(false);
const fileInput = ref(null);
const handleDragEnter = (e) => {
 e.preventDefault();
 isDragging.value = true;
};
const handleDragLeave = (e) => {
 e.preventDefault();
 isDragging.value = false;
};
const handleDrop = (e) => {
 e.preventDefault();
 isDragging.value = false;
 const files = Array.from(e.dataTransfer.files);
 processFiles(files);
};
const handleClick = () => {
 fileInput.value?.click();
};
const handleFileChange = (e) => {
 const files = Array.from(e.target.files);
 processFiles(files);
};
const processFiles = (files) => {
 const validFiles = files.filter(file => {
 const ext = file.name.split('.').pop().toLowerCase();
 return ['log', 'txt', 'dat'].includes(ext);
 });
 if (validFiles.length !== files.length) {
 ElMessage.warning('部分文件格式不支持，仅支持 .log, .txt, .dat 格式');
 }
 if (validFiles.length > 0) {
 emit('files-selected', validFiles);
 }
};
</script>

<template>
  <div 
    class="upload-area"
    :class="{ 'is-dragging': isDragging }"
    @dragenter="handleDragEnter"
    @dragleave="handleDragLeave"
    @drop="handleDrop"
    @click="handleClick"
  >
    <input 
      ref="fileInput"
      type="file" 
      :multiple="multiple"
      :accept="accept"
      class="file-input"
      @change="handleFileChange"
    />
    <div class="upload-icon">
      <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
        <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
        <polyline points="17 8 12 3 7 8" />
        <line x1="12" y1="3" x2="12" y2="15" />
      </svg>
    </div>
    <p class="upload-text">拖拽加密日志文件到此处，或点击上传</p>
    <p class="upload-hint">支持 .log, .txt, .dat 格式</p>
  </div>
</template>

<style scoped>
.upload-area {
  border: 2px dashed rgba(22, 93, 255, 0.4);
  border-radius: 12px;
  padding: 40px;
  text-align: center;
  cursor: pointer;
  transition: all 0.3s ease;
  background: rgba(22, 93, 255, 0.03);

  &:hover {
    border-color: rgba(22, 93, 255, 0.8);
    background: rgba(22, 93, 255, 0.08);
  }

  &.is-dragging {
    border-color: #165DFF;
    background: rgba(22, 93, 255, 0.15);
    transform: scale(1.02);
  }
}

.file-input {
  display: none;
}

.upload-icon {
  color: #165DFF;
  margin-bottom: 16px;
}

.upload-text {
  font-size: 14px;
  color: var(--text-secondary);
  margin-bottom: 8px;
}

.upload-hint {
  font-size: 12px;
  color: var(--text-tertiary);
}
</style>
