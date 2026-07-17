<script setup>
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import GlassCard from '@/components/GlassCard.vue';
import { useLogsStore } from '@/stores/logs';

const logsStore = useLogsStore();

// ==================== 录音状态 ====================
const isRecording = ref(false);
const isPaused = ref(false);
const recordingDuration = ref(0); // 秒
const recordingStartTime = ref(null);
const transcript = ref(''); // 实时文稿
const interimTranscript = ref(''); // 临时（未确认）文稿
const isSpeechSupported = ref(true); // 浏览器是否支持 Web Speech API
const speechError = ref(''); // 语音识别错误信息

// 断句开关
const autoSentenceBreak = ref(true); // 是否自动断句
const autoSaveEnabled = ref(true); // 是否在停止时自动保存

// ==================== 录音列表 ====================
const recordings = ref([]);
const selectedRecording = ref(null);

// ==================== 录音相关对象 ====================
let mediaRecorder = null;
let audioChunks = [];
let audioBlob = null;
let audioUrl = ref('');
let timerInterval = null;
let recognition = null; // Web Speech API 实例

// 本地存储 key
const STORAGE_KEY = 'audio_recordings';

// ==================== 工具函数 ====================
const formatDuration = (seconds) => {
  if (!seconds || seconds < 0) return '00:00';
  const h = Math.floor(seconds / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  const s = seconds % 60;
  if (h > 0) {
    return `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`;
  }
  return `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`;
};

const formatDate = (timestamp) => {
  const d = new Date(timestamp);
  const pad = (n) => String(n).padStart(2, '0');
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`;
};

const formatFileSize = (bytes) => {
  if (!bytes) return '-';
  if (bytes < 1024) return bytes + ' B';
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
};

// ==================== 加载/保存录音列表 ====================
const saveRecordings = () => {
  try {
    // 仅保存元数据，不保存音频二进制（localStorage 限制）
    const metadata = recordings.value.map((r) => ({
      id: r.id,
      title: r.title,
      createdAt: r.createdAt,
      duration: r.duration,
      transcript: r.transcript,
      hasAudio: !!r.audioBase64
    }));
    localStorage.setItem(STORAGE_KEY + '_list', JSON.stringify(metadata));
  } catch (e) {
    console.warn('保存录音列表失败:', e);
  }
};

// 加载录音列表（仅元数据）
const loadRecordingsList = () => {
  try {
    const saved = localStorage.getItem(STORAGE_KEY + '_list');
    if (saved) {
      const parsed = JSON.parse(saved);
      if (Array.isArray(parsed)) {
        recordings.value = parsed;
      }
    }
  } catch (e) {
    console.warn('加载录音列表失败:', e);
  }
};

// ==================== 语音识别初始化 ====================
const initSpeechRecognition = () => {
  const SpeechRecognition = window.SpeechRecognition || window.webkitSpeechRecognition;
  if (!SpeechRecognition) {
    isSpeechSupported.value = false;
    return null;
  }

  const rec = new SpeechRecognition();
  rec.lang = 'zh-CN';
  rec.continuous = true;
  rec.interimResults = true;
  rec.maxAlternatives = 1;

  rec.onresult = (event) => {
    let interim = '';
    let final = '';
    for (let i = event.resultIndex; i < event.results.length; i++) {
      const result = event.results[i];
      if (result.isFinal) {
        final += result[0].transcript;
      } else {
        interim += result[0].transcript;
      }
    }
    if (final) {
      // 自动断句处理：在句末标点后添加换行
      let processed = final;
      if (autoSentenceBreak.value) {
        // 中文句末标点：。！？ 以及 英文 .!?
        processed = processed.replace(/([。！？!?\.])\s*/g, '$1\n');
        // 处理连续的换行
        processed = processed.replace(/\n{2,}/g, '\n');
      }
      transcript.value += processed;
    }
    interimTranscript.value = interim;
  };

  rec.onerror = (event) => {
    console.error('语音识别错误:', event.error);
    speechError.value = `语音识别错误: ${event.error}`;
    if (event.error === 'no-speech' || event.error === 'audio-capture') {
      // 这些错误不致命，继续录音
      return;
    }
    if (event.error === 'not-allowed') {
      ElMessage.error('麦克风权限被拒绝，请在浏览器设置中允许');
      stopRecording();
    }
  };

  rec.onend = () => {
    // 如果还在录音状态，自动重启识别
    if (isRecording.value && !isPaused.value) {
      try {
        rec.start();
      } catch (e) {
        console.warn('重启语音识别失败:', e);
      }
    }
  };

  return rec;
};

// ==================== 开始录音 ====================
const startRecording = async () => {
  try {
    speechError.value = '';

    // 1. 检查浏览器是否支持 getUserMedia
    if (!navigator.mediaDevices || !navigator.mediaDevices.getUserMedia) {
      // 尝试兼容性处理：在旧版浏览器中，getUserMedia 可能位于 navigator 对象上
      if (!navigator.getUserMedia) {
        throw new Error('浏览器不支持录音功能。请使用 Chrome、Edge 等现代浏览器，并确保使用 HTTPS 连接。');
      }
    }

    // 2. 检查安全上下文（HTTPS 或 localhost）
    const isSecureContext = window.isSecureContext || 
                          location.protocol === 'https:' || 
                          location.hostname === 'localhost' || 
                          location.hostname === '127.0.0.1';
    
    if (!isSecureContext) {
      console.warn('录音功能在非安全上下文中可能受限。建议使用 HTTPS 或 localhost 访问。');
    }

    // 3. 检查浏览器兼容性并请求麦克风权限
    let stream;
    try {
      // 首先尝试标准的 mediaDevices API
      stream = await navigator.mediaDevices.getUserMedia({ 
        audio: {
          echoCancellation: true,
          noiseSuppression: true,
          autoGainControl: true
        } 
      });
    } catch (mediaError) {
      // 如果标准 API 失败，尝试旧版 API（兼容性处理）
      console.warn('标准 getUserMedia API 失败，尝试旧版 API:', mediaError);
      
      if (navigator.getUserMedia) {
        // 旧版 API（Promise 包装）
        stream = await new Promise((resolve, reject) => {
          navigator.getUserMedia(
            { audio: true },
            (s) => resolve(s),
            (e) => reject(e)
          );
        });
      } else {
        throw mediaError; // 重新抛出原始错误
      }
    }

    // 3. 初始化 MediaRecorder
    mediaRecorder = new MediaRecorder(stream);
    audioChunks = [];

    mediaRecorder.ondataavailable = (e) => {
      if (e.data.size > 0) {
        audioChunks.push(e.data);
      }
    };

    mediaRecorder.onstop = () => {
      // 录音结束时，生成 blob
      audioBlob = new Blob(audioChunks, { type: 'audio/webm' });
      if (audioUrl.value) {
        URL.revokeObjectURL(audioUrl.value);
      }
      audioUrl.value = URL.createObjectURL(audioBlob);
      // 停止所有音频轨道
      stream.getTracks().forEach((track) => track.stop());
    };

    mediaRecorder.start();

    // 4. 启动 Web Speech API（实时转写）
    recognition = initSpeechRecognition();
    if (recognition) {
      try {
        recognition.start();
      } catch (e) {
        console.warn('启动语音识别失败:', e);
      }
    }

    // 5. 重置状态
    isRecording.value = true;
    isPaused.value = false;
    transcript.value = '';
    interimTranscript.value = '';
    recordingStartTime.value = Date.now();
    recordingDuration.value = 0;

    // 6. 启动计时器
    timerInterval = setInterval(() => {
      if (!isPaused.value) {
        recordingDuration.value = Math.floor((Date.now() - recordingStartTime.value) / 1000);
      }
    }, 1000);

    ElMessage.success('开始录音');
    logsStore.addLog('开始', '录音管理', '开始录音');
  } catch (e) {
    console.error('录音失败:', e);
    if (e.name === 'NotAllowedError') {
      ElMessage.error('麦克风权限被拒绝。请在浏览器设置中允许麦克风访问。');
    } else if (e.name === 'NotFoundError') {
      ElMessage.error('未找到麦克风设备。请确保麦克风已连接并正常工作。');
    } else if (e.name === 'NotReadableError') {
      ElMessage.error('无法访问麦克风。可能是其他应用正在使用麦克风。');
    } else if (e.name === 'SecurityError' || e.message.includes('security') || e.message.includes('secure')) {
      ElMessage.error('安全限制：请在 HTTPS 环境下使用录音功能，或使用 localhost 访问。\n当前环境：' + location.protocol + '//' + location.host);
    } else if (e.name === 'TypeError' && e.message.includes('getUserMedia')) {
      ElMessage.error('浏览器不支持录音功能。请使用最新版本的 Chrome、Edge、Firefox 或 Safari。');
    } else if (e.message.includes('不支持录音功能')) {
      ElMessage.error(e.message);
    } else {
      ElMessage.error('录音启动失败: ' + e.message);
    }
  }
};

// ==================== 暂停录音 ====================
const pauseRecording = () => {
  if (mediaRecorder && mediaRecorder.state === 'recording') {
    mediaRecorder.pause();
    isPaused.value = true;
    if (recognition) {
      try { recognition.stop(); } catch (e) {}
    }
    if (timerInterval) {
      clearInterval(timerInterval);
      timerInterval = null;
    }
  }
};

// ==================== 恢复录音 ====================
const resumeRecording = () => {
  if (mediaRecorder && mediaRecorder.state === 'paused') {
    mediaRecorder.resume();
    isPaused.value = false;
    recordingStartTime.value = Date.now() - recordingDuration.value * 1000;
    if (recognition) {
      try { recognition.start(); } catch (e) {}
    }
    timerInterval = setInterval(() => {
      if (!isPaused.value) {
        recordingDuration.value = Math.floor((Date.now() - recordingStartTime.value) / 1000);
      }
    }, 1000);
  }
};

// ==================== 停止录音 ====================
const stopRecording = () => {
  if (mediaRecorder && (mediaRecorder.state === 'recording' || mediaRecorder.state === 'paused')) {
    mediaRecorder.stop();
  }

  if (recognition) {
    try { recognition.stop(); } catch (e) {}
    recognition = null;
  }

  if (timerInterval) {
    clearInterval(timerInterval);
    timerInterval = null;
  }

  isRecording.value = false;
  isPaused.value = false;

  // 等待 mediaRecorder.onstop 完成（异步），再让用户保存
  setTimeout(() => {
    if (audioBlob) {
      // 停止时自动保存（不再弹窗命名，使用默认标题）
      if (autoSaveEnabled.value) {
        autoSaveRecording();
      } else {
        ElMessage.success('录音已停止，可以保存');
      }
    }
  }, 300);
};

// ==================== 工具：生成默认标题 ====================
const generateDefaultTitle = () => {
  const now = new Date();
  return `录音_${now.getFullYear()}${String(now.getMonth() + 1).padStart(2, '0')}${String(now.getDate()).padStart(2, '0')}_${String(now.getHours()).padStart(2, '0')}${String(now.getMinutes()).padStart(2, '0')}`;
};

// ==================== 内部：执行保存（使用给定标题） ====================
const performSave = async (title) => {
  if (!audioBlob) {
    ElMessage.warning('暂无录音可保存');
    return false;
  }
  // 兜底：若标题为空，使用默认标题
  const finalTitle = (title && title.trim()) ? title.trim() : generateDefaultTitle();
  const audioBase64 = await blobToBase64(audioBlob);
  const recording = {
    id: 'rec_' + Date.now() + '_' + Math.random().toString(36).slice(2, 8),
    title: finalTitle,
    createdAt: Date.now(),
    duration: recordingDuration.value,
    transcript: transcript.value.trim(),
    audioBase64: audioBase64,
    audioType: audioBlob.type || 'audio/webm',
    audioSize: audioBlob.size
  };
  recordings.value.unshift(recording);
  saveRecordings();
  logsStore.addLog('保存', '录音管理', `保存录音: ${finalTitle}`);
  // 释放当前 blob
  audioBlob = null;
  if (audioUrl.value) {
    URL.revokeObjectURL(audioUrl.value);
    audioUrl.value = '';
  }
  recordingDuration.value = 0;
  transcript.value = '';
  interimTranscript.value = '';
  return true;
};

// ==================== 手动保存（弹出输入框，命名后保存） ====================
const saveRecording = async () => {
  if (!audioBlob) {
    ElMessage.warning('暂无录音可保存');
    return;
  }
  const defaultTitle = generateDefaultTitle();
  try {
    const { value } = await ElMessageBox.prompt('请输入录音标题', '保存录音', {
      inputValue: defaultTitle,
      confirmButtonText: '保存',
      cancelButtonText: '取消',
      inputValidator: (val) => {
        if (!val || !val.trim()) {
          return '标题不能为空';
        }
        if (val.length > 50) {
          return '标题不能超过 50 个字符';
        }
        return true;
      }
    });
    const ok = await performSave(value);
    if (ok) {
      ElMessage.success('录音已保存');
    }
  } catch (e) {
    // 用户取消
  }
};

// ==================== 自动保存（停止录音后自动触发，使用默认标题） ====================
const autoSaveRecording = async () => {
  if (!audioBlob) {
    return;
  }
  // 仅在有内容（音频时长 > 0 或 文稿非空）时保存，避免保存空录音
  if (recordingDuration.value === 0 && !transcript.value.trim()) {
    return;
  }
  const title = generateDefaultTitle();
  const ok = await performSave(title);
  if (ok) {
    ElMessage.success(`已自动保存：${title}`);
  }
};

const blobToBase64 = (blob) => {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onloadend = () => resolve(reader.result);
    reader.onerror = reject;
    reader.readAsDataURL(blob);
  });
};

// ==================== 删除录音 ====================
const deleteRecording = async (recording) => {
  try {
    await ElMessageBox.confirm(`确定要删除录音"${recording.title}"吗？`, '确认删除', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消'
    });

    const index = recordings.value.findIndex((r) => r.id === recording.id);
    if (index >= 0) {
      recordings.value.splice(index, 1);
      saveRecordings();
      if (selectedRecording.value?.id === recording.id) {
        selectedRecording.value = null;
      }
      ElMessage.success('已删除');
    }
  } catch (e) {
    // 取消
  }
};

// ==================== 选中录音 ====================
const selectRecording = (recording) => {
  selectedRecording.value = recording;
  // 自动滚动到详情区
  nextTick(() => {
    const detailEl = document.querySelector('.recording-detail');
    if (detailEl) {
      detailEl.scrollIntoView({ behavior: 'smooth', block: 'start' });
    }
  });
};

// ==================== 重命名录音 ====================
const renameRecording = async (recording) => {
  try {
    const { value } = await ElMessageBox.prompt('请输入新标题', '重命名录音', {
      inputValue: recording.title,
      confirmButtonText: '保存',
      cancelButtonText: '取消'
    });
    if (value && value.trim()) {
      recording.title = value.trim();
      saveRecordings();
      ElMessage.success('已重命名');
    }
  } catch (e) {
    // 取消
  }
};

// ==================== 导出文稿 ====================
const exportTranscript = (recording) => {
  if (!recording.transcript) {
    ElMessage.warning('该录音无文稿内容');
    return;
  }
  const content = `${recording.title}\n录制时间: ${formatDate(recording.createdAt)}\n时长: ${formatDuration(recording.duration)}\n\n${recording.transcript}`;
  const blob = new Blob([content], { type: 'text/plain;charset=utf-8' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `${recording.title}_文稿.txt`;
  a.click();
  URL.revokeObjectURL(url);
  ElMessage.success('文稿已导出');
};

// ==================== 复制文稿 ====================
const copyTranscript = async (recording) => {
  if (!recording.transcript) {
    ElMessage.warning('该录音无文稿内容');
    return;
  }
  try {
    await navigator.clipboard.writeText(recording.transcript);
    ElMessage.success('文稿已复制到剪贴板');
  } catch (e) {
    ElMessage.error('复制失败');
  }
};

// ==================== 清除当前录音（不保存） ====================
const clearCurrent = () => {
  audioBlob = null;
  if (audioUrl.value) {
    URL.revokeObjectURL(audioUrl.value);
    audioUrl.value = '';
  }
  recordingDuration.value = 0;
  transcript.value = '';
  interimTranscript.value = '';
};

// 在光标位置插入换行
const insertLineBreak = () => {
  const textarea = document.querySelector('.transcript-textarea');
  if (!textarea) {
    transcript.value += '\n';
    return;
  }
  const start = textarea.selectionStart;
  const end = textarea.selectionEnd;
  const text = transcript.value;
  transcript.value = text.substring(0, start) + '\n' + text.substring(end);
  // 恢复光标位置
  nextTick(() => {
    textarea.focus();
    textarea.setSelectionRange(start + 1, start + 1);
  });
};

// 在文末添加换行
const addNewLine = () => {
  if (transcript.value && !transcript.value.endsWith('\n')) {
    transcript.value += '\n';
  } else {
    transcript.value += '\n';
  }
};

// ==================== 计算属性 ====================
const hasCurrentRecording = computed(() => audioBlob !== null);
const totalRecordings = computed(() => recordings.value.length);

// ==================== 生命周期 ====================
onMounted(() => {
  loadRecordingsList();
  
  // 检测语音识别支持
  if (!window.SpeechRecognition && !window.webkitSpeechRecognition) {
    isSpeechSupported.value = false;
  }
  
  // 检测媒体设备支持
  if (!navigator.mediaDevices || !navigator.mediaDevices.getUserMedia) {
    console.warn('浏览器不支持 mediaDevices API，录音功能不可用');
    // 可以在界面上显示警告
  }
});

onUnmounted(() => {
  if (timerInterval) {
    clearInterval(timerInterval);
  }
  if (recognition) {
    try { recognition.stop(); } catch (e) {}
  }
  if (mediaRecorder && (mediaRecorder.state === 'recording' || mediaRecorder.state === 'paused')) {
    try { mediaRecorder.stop(); } catch (e) {}
  }
  if (audioUrl.value) {
    URL.revokeObjectURL(audioUrl.value);
  }
});
</script>

<template>
  <div class="audio-recorder-page">
    <GlassCard>
      <div class="page-container">
        <!-- 顶部标题 -->
        <div class="page-header">
          <h2 class="page-title">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3z" />
              <path d="M19 10v2a7 7 0 0 1-14 0v-2" />
              <line x1="12" y1="19" x2="12" y2="23" />
              <line x1="8" y1="23" x2="16" y2="23" />
            </svg>
            录音管理
          </h2>
          <span class="recordings-count">共 {{ totalRecordings }} 条录音</span>
        </div>

        <!-- 录音控制面板 -->
        <div class="recording-panel" :class="{ 'is-recording': isRecording, 'is-paused': isPaused }">
          <!-- 录音状态指示 -->
          <div class="recording-status">
            <div class="status-indicator">
              <span v-if="!isRecording" class="status-dot idle"></span>
              <span v-else-if="isPaused" class="status-dot paused"></span>
              <span v-else class="status-dot active"></span>
            </div>
            <div class="status-text">
              <div v-if="!isRecording" class="status-label">就绪</div>
              <div v-else-if="isPaused" class="status-label">已暂停</div>
              <div v-else class="status-label">正在录音</div>
              <div class="status-time">{{ formatDuration(recordingDuration) }}</div>
            </div>
          </div>

          <!-- 录音按钮 -->
          <div class="recording-controls">
            <button
              v-if="!isRecording"
              class="record-btn start"
              @click="startRecording"
            >
              <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
                <circle cx="12" cy="12" r="6" />
              </svg>
              <span>开始录音</span>
            </button>

            <template v-else>
              <button
                v-if="!isPaused"
                class="record-btn pause"
                @click="pauseRecording"
              >
                <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
                  <rect x="6" y="4" width="4" height="16" />
                  <rect x="14" y="4" width="4" height="16" />
                </svg>
                <span>暂停</span>
              </button>

              <button
                v-else
                class="record-btn resume"
                @click="resumeRecording"
              >
                <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
                  <polygon points="5 3 19 12 5 21 5 3" />
                </svg>
                <span>继续</span>
              </button>

              <button class="record-btn stop" @click="stopRecording">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
                  <rect x="5" y="5" width="14" height="14" />
                </svg>
                <span>停止</span>
              </button>
            </template>

            <button
              v-if="hasCurrentRecording && !isRecording"
              class="record-btn save"
              @click="saveRecording"
            >
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z" />
                <polyline points="17 21 17 13 7 13 7 21" />
                <polyline points="7 3 7 8 15 8" />
              </svg>
              <span>另存为</span>
            </button>

            <button
              v-if="hasCurrentRecording && !isRecording"
              class="record-btn clear"
              @click="clearCurrent"
            >
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <polyline points="3 6 5 6 21 6" />
                <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
              </svg>
              <span>放弃</span>
            </button>
          </div>

          <!-- 录音预览 -->
          <div v-if="hasCurrentRecording" class="recording-preview">
            <div class="preview-label">录音预览：</div>
            <audio :src="audioUrl" controls class="audio-player"></audio>
          </div>
        </div>

        <!-- 语音识别提示 -->
        <div v-if="!isSpeechSupported" class="warning-hint">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" />
            <line x1="12" y1="9" x2="12" y2="13" />
            <line x1="12" y1="17" x2="12.01" y2="17" />
          </svg>
          <span>当前浏览器不支持实时语音识别（建议使用 Chrome / Edge）。您仍可以录音并手动编辑文稿。</span>
        </div>

        <!-- 实时文稿 -->
        <div class="transcript-panel">
          <div class="transcript-header">
            <span class="transcript-title">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
                <polyline points="14 2 14 8 20 8" />
                <line x1="16" y1="13" x2="8" y2="13" />
                <line x1="16" y1="17" x2="8" y2="17" />
              </svg>
              实时文稿
            </span>
            <div class="transcript-tools">
              <label class="auto-break-toggle" title="识别到句末标点（。！？）时自动换行">
                <input type="checkbox" v-model="autoSentenceBreak" />
                <span>自动断句</span>
              </label>
              <button class="tool-btn" @click="insertLineBreak" title="在光标处插入换行">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <polyline points="7 7 12 12 7 17" />
                  <line x1="13" y1="17" x2="20" y2="17" />
                </svg>
                <span>换行</span>
              </button>
              <button class="tool-btn" @click="addNewLine" title="在文末添加新行">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <line x1="12" y1="5" x2="12" y2="19" />
                  <line x1="5" y1="12" x2="19" y2="12" />
                </svg>
                <span>新行</span>
              </button>
              <span v-if="isRecording && !isPaused && isSpeechSupported" class="live-tag">
                <span class="live-dot"></span>
                实时转写中
              </span>
            </div>
          </div>
          <textarea
            v-model="transcript"
            class="transcript-textarea"
            placeholder="录音开始后，这里会实时显示语音转写的内容。&#10;您也可以手动编辑文稿内容。&#10;&#10;提示：&#10;- 勾选「自动断句」可在句末标点后自动换行&#10;- 勾选「停止时自动保存」停止录音后自动保存到本地&#10;- 点击「换行」按钮在光标处插入换行"
            rows="10"
          ></textarea>
          <div v-if="interimTranscript" class="interim-text">
            <span class="interim-label">识别中：</span>
            <span class="interim-content">{{ interimTranscript }}</span>
          </div>
        </div>

        <!-- 录音列表 -->
        <div class="recordings-section">
          <div class="section-header">
            <h3>历史录音</h3>
            <span class="section-count">{{ totalRecordings }} 条</span>
          </div>

          <div v-if="recordings.length === 0" class="empty-state">
            <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
              <path d="M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3z" />
              <path d="M19 10v2a7 7 0 0 1-14 0v-2" />
              <line x1="12" y1="19" x2="12" y2="23" />
              <line x1="8" y1="23" x2="16" y2="23" />
            </svg>
            <p>还没有录音</p>
            <span>点击上方"开始录音"按钮开始您的第一条录音</span>
          </div>

          <div v-else class="recordings-list">
            <div
              v-for="rec in recordings"
              :key="rec.id"
              class="recording-item"
              :class="{ selected: selectedRecording?.id === rec.id }"
              @click="selectRecording(rec)"
            >
              <div class="recording-icon">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3z" />
                  <path d="M19 10v2a7 7 0 0 1-14 0v-2" />
                </svg>
              </div>
              <div class="recording-info">
                <div class="recording-title">{{ rec.title }}</div>
                <div class="recording-meta">
                  <span>{{ formatDate(rec.createdAt) }}</span>
                  <span class="meta-divider">·</span>
                  <span>{{ formatDuration(rec.duration) }}</span>
                  <span v-if="rec.transcript" class="meta-divider">·</span>
                  <span v-if="rec.transcript" class="transcript-indicator">
                    <svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor">
                      <path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41L9 16.17z" />
                    </svg>
                    有文稿
                  </span>
                </div>
              </div>
              <div class="recording-actions" @click.stop>
                <button class="action-btn" @click="renameRecording(rec)" title="重命名">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M12 20h9" />
                    <path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z" />
                  </svg>
                </button>
                <button class="action-btn danger" @click="deleteRecording(rec)" title="删除">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <polyline points="3 6 5 6 21 6" />
                    <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
                  </svg>
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- 录音详情 -->
        <div v-if="selectedRecording" class="recording-detail">
          <div class="detail-header">
            <h3>{{ selectedRecording.title }}</h3>
            <div class="detail-actions">
              <button class="detail-btn" @click="copyTranscript(selectedRecording)">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
                  <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
                </svg>
                复制文稿
              </button>
              <button class="detail-btn" @click="exportTranscript(selectedRecording)">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
                  <polyline points="7 10 12 15 17 10" />
                  <line x1="12" y1="15" x2="12" y2="3" />
                </svg>
                导出文稿
              </button>
            </div>
          </div>

          <div class="detail-meta">
            <span>📅 {{ formatDate(selectedRecording.createdAt) }}</span>
            <span>⏱ {{ formatDuration(selectedRecording.duration) }}</span>
            <span>📦 {{ formatFileSize(selectedRecording.audioSize) }}</span>
          </div>

          <!-- 音频播放 -->
          <div v-if="selectedRecording.audioBase64" class="detail-audio">
            <audio :src="selectedRecording.audioBase64" controls class="audio-player full"></audio>
          </div>

          <!-- 文稿内容 -->
          <div class="detail-transcript">
            <div class="transcript-label">
              <span>文稿内容</span>
            </div>
            <div v-if="selectedRecording.transcript" class="transcript-content">
              <pre>{{ selectedRecording.transcript }}</pre>
            </div>
            <div v-else class="transcript-empty">
              该录音未包含文稿内容
            </div>
          </div>
        </div>
      </div>
    </GlassCard>
  </div>
</template>

<style scoped>
.audio-recorder-page {
  padding: 20px;
  min-height: 100%;
}

.page-container {
  max-width: 1000px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24px;
  padding-bottom: 16px;
  border-bottom: 1px solid #f0f0f0;
}

.page-title {
  display: flex;
  align-items: center;
  gap: 10px;
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: #1d2129;
}

.page-title svg {
  color: #165DFF;
}

.recordings-count {
  font-size: 13px;
  color: #86909c;
  background: #f2f3f5;
  padding: 4px 12px;
  border-radius: 12px;
}

/* ==================== 录音控制面板 ==================== */
.recording-panel {
  background: white;
  border: 1px solid #f0f0f0;
  border-radius: 12px;
  padding: 20px;
  margin-bottom: 16px;
  transition: all 0.3s;
}

.recording-panel.is-recording {
  border-color: #ff7d00;
  box-shadow: 0 0 0 3px rgba(255, 125, 0, 0.1);
}

.recording-panel.is-paused {
  border-color: #ffaa00;
  background: #fffbe6;
}

.recording-status {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 16px;
  padding-bottom: 16px;
  border-bottom: 1px solid #f5f5f5;
}

.status-indicator {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f5f5;
}

.status-dot {
  width: 14px;
  height: 14px;
  border-radius: 50%;
  background: #c9cdd4;
  display: block;
}

.status-dot.idle {
  background: #c9cdd4;
}

.status-dot.active {
  background: #f53f3f;
  animation: pulse 1.5s infinite;
}

.status-dot.paused {
  background: #ffaa00;
}

@keyframes pulse {
  0% {
    box-shadow: 0 0 0 0 rgba(245, 63, 63, 0.6);
  }
  70% {
    box-shadow: 0 0 0 12px rgba(245, 63, 63, 0);
  }
  100% {
    box-shadow: 0 0 0 0 rgba(245, 63, 63, 0);
  }
}

.status-text {
  flex: 1;
}

.status-label {
  font-size: 14px;
  font-weight: 500;
  color: #1d2129;
  margin-bottom: 2px;
}

.status-time {
  font-size: 20px;
  font-weight: 600;
  color: #165DFF;
  font-family: 'Courier New', monospace;
}

.recording-panel.is-recording .status-time {
  color: #f53f3f;
}

.recording-controls {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
  align-items: center;
}

.record-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 20px;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  color: white;
}

.record-btn:hover {
  transform: translateY(-1px);
}

.record-btn.start {
  background: linear-gradient(135deg, #f53f3f 0%, #ff7d00 100%);
  box-shadow: 0 2px 8px rgba(245, 63, 63, 0.3);
}

.record-btn.start:hover {
  box-shadow: 0 4px 16px rgba(245, 63, 63, 0.4);
}

.record-btn.pause,
.record-btn.resume {
  background: #ffaa00;
  box-shadow: 0 2px 8px rgba(255, 170, 0, 0.3);
}

.record-btn.stop {
  background: #4e5969;
  box-shadow: 0 2px 8px rgba(78, 89, 105, 0.3);
}

.record-btn.save {
  background: linear-gradient(135deg, #165DFF 0%, #4080FF 100%);
  box-shadow: 0 2px 8px rgba(22, 93, 255, 0.3);
}

.record-btn.clear {
  background: white;
  color: #4e5969;
  border: 1px solid #dcdfe6;
  box-shadow: none;
}

.record-btn.clear:hover {
  border-color: #f53f3f;
  color: #f53f3f;
}

.recording-preview {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid #f5f5f5;
  display: flex;
  align-items: center;
  gap: 12px;
}

.preview-label {
  font-size: 13px;
  color: #4e5969;
  white-space: nowrap;
}

.audio-player {
  flex: 1;
  height: 36px;
  max-width: 100%;
}

.audio-player.full {
  width: 100%;
  height: 40px;
}

/* ==================== 警告提示 ==================== */
.warning-hint {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  background: #fff7e6;
  border: 1px solid #ffd591;
  border-radius: 8px;
  color: #ad6800;
  font-size: 13px;
  margin-bottom: 16px;
}

.warning-hint svg {
  flex-shrink: 0;
}

/* ==================== 文稿面板 ==================== */
.transcript-panel {
  background: white;
  border: 1px solid #f0f0f0;
  border-radius: 12px;
  padding: 16px;
  margin-bottom: 20px;
}

.transcript-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10px;
}

.transcript-title {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  font-weight: 500;
  color: #1d2129;
}

.transcript-title svg {
  color: #165DFF;
}

.live-tag {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: #f53f3f;
  background: #ffece8;
  padding: 3px 10px;
  border-radius: 10px;
  font-weight: 500;
}

.live-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #f53f3f;
  animation: pulse-dot 1.5s infinite;
}

@keyframes pulse-dot {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.3; }
}

.transcript-textarea {
  width: 100%;
  border: 1px solid #e5e6eb;
  border-radius: 6px;
  padding: 10px 12px;
  font-size: 14px;
  line-height: 1.6;
  color: #1d2129;
  resize: vertical;
  font-family: inherit;
  box-sizing: border-box;
}

.transcript-textarea:focus {
  outline: none;
  border-color: #165DFF;
  box-shadow: 0 0 0 3px rgba(22, 93, 255, 0.1);
}

.interim-text {
  margin-top: 8px;
  padding: 8px 12px;
  background: #f5f5f5;
  border-radius: 6px;
  font-size: 13px;
  color: #86909c;
}

.interim-label {
  font-weight: 500;
  margin-right: 4px;
}

.interim-content {
  font-style: italic;
}

/* ==================== 文稿工具栏 ==================== */
.transcript-tools {
  display: flex;
  align-items: center;
  gap: 8px;
}

.auto-break-toggle {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: #4e5969;
  cursor: pointer;
  user-select: none;
}

.auto-break-toggle input[type="checkbox"] {
  cursor: pointer;
  accent-color: #165DFF;
}

.tool-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  background: #f2f3f5;
  color: #1d2129;
  border: none;
  border-radius: 4px;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s;
}

.tool-btn:hover {
  background: #e5e6eb;
  color: #165DFF;
}

/* ==================== 录音列表 ==================== */
.recordings-section {
  background: white;
  border: 1px solid #f0f0f0;
  border-radius: 12px;
  padding: 16px;
  margin-bottom: 20px;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
  padding-bottom: 10px;
  border-bottom: 1px solid #f5f5f5;
}

.section-header h3 {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  color: #1d2129;
}

.section-count {
  font-size: 12px;
  color: #86909c;
}

.empty-state {
  text-align: center;
  padding: 40px 20px;
  color: #86909c;
}

.empty-state svg {
  color: #c9cdd4;
  margin-bottom: 12px;
}

.empty-state p {
  margin: 0 0 4px 0;
  font-size: 14px;
  color: #4e5969;
  font-weight: 500;
}

.empty-state span {
  font-size: 13px;
}

.recordings-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.recording-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  border: 1px solid #f0f0f0;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  background: white;
}

.recording-item:hover {
  border-color: #165DFF;
  background: #f8faff;
}

.recording-item.selected {
  border-color: #165DFF;
  background: #e8f3ff;
}

.recording-icon {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  background: linear-gradient(135deg, #e8f3ff 0%, #f0f7ff 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #165DFF;
  flex-shrink: 0;
}

.recording-info {
  flex: 1;
  min-width: 0;
}

.recording-title {
  font-size: 14px;
  font-weight: 500;
  color: #1d2129;
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.recording-meta {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #86909c;
  flex-wrap: wrap;
}

.meta-divider {
  color: #c9cdd4;
}

.transcript-indicator {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  color: #00b42a;
}

.recording-actions {
  display: flex;
  gap: 4px;
  flex-shrink: 0;
}

.action-btn {
  width: 28px;
  height: 28px;
  border: none;
  background: transparent;
  border-radius: 4px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #86909c;
  transition: all 0.2s;
}

.action-btn:hover {
  background: #f2f3f5;
  color: #165DFF;
}

.action-btn.danger:hover {
  background: #ffe7e7;
  color: #f53f3f;
}

/* ==================== 录音详情 ==================== */
.recording-detail {
  background: white;
  border: 1px solid #f0f0f0;
  border-radius: 12px;
  padding: 20px;
}

.detail-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid #f5f5f5;
}

.detail-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #1d2129;
}

.detail-actions {
  display: flex;
  gap: 8px;
}

.detail-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 6px 12px;
  background: white;
  color: #4e5969;
  border: 1px solid #dcdfe6;
  border-radius: 6px;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}

.detail-btn:hover {
  border-color: #165DFF;
  color: #165DFF;
}

.detail-meta {
  display: flex;
  gap: 16px;
  margin-bottom: 16px;
  font-size: 13px;
  color: #4e5969;
  flex-wrap: wrap;
}

.detail-audio {
  margin-bottom: 16px;
  padding: 12px;
  background: #fafbfc;
  border-radius: 8px;
  border: 1px solid #f0f0f0;
}

.transcript-label {
  font-size: 13px;
  font-weight: 500;
  color: #4e5969;
  margin-bottom: 8px;
}

.transcript-content {
  padding: 12px 16px;
  background: #fafbfc;
  border-radius: 8px;
  border-left: 3px solid #165DFF;
  font-size: 14px;
  line-height: 1.8;
  color: #1d2129;
  white-space: pre-wrap;
  word-break: break-word;
}

.transcript-empty {
  padding: 16px;
  background: #f5f5f5;
  border-radius: 8px;
  text-align: center;
  color: #86909c;
  font-size: 13px;
}

@media (max-width: 768px) {
  .recording-status {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }

  .recording-controls {
    width: 100%;
  }

  .record-btn {
    flex: 1;
    justify-content: center;
  }
}
</style>
