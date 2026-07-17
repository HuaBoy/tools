<script setup>
import { ref, onMounted, onUnmounted, nextTick, watch } from 'vue';
import { io } from 'socket.io-client';

// ==================== 状态 ====================
const state = ref({
  socket: null,
  pairedDeviceId: null,
  peerConnection: null,
  dataChannel: null,
  isConnected: false,
  frameCount: 0,
  lastFpsTime: Date.now(),
  currentFps: 0,
  latency: 0,
  lastPingTime: 0,
  screenWidth: 0,
  screenHeight: 0,
});

const devices = ref([]);
const serverUrl = ref('http://10.51.22.46:9528');  // 直接连接信令服务器
const toastMsg = ref('');
const toastVisible = ref(false);
const navStatusClass = ref('offline');
const navStatusText = ref('未连接');
const showControlSection = ref(false);
const showPhoneFrame = ref(false);
const showPlaceholder = ref(true);
const showConnectionInfo = ref(false);
const isNativeMode = ref(false);

// ==================== Refs ====================
const phoneScreenRef = ref(null);
const remoteVideoRef = ref(null);
const remoteCanvasRef = ref(null);

// ==================== Toast ====================
let toastTimer = null;
function showToast(msg, duration = 2000) {
  toastMsg.value = msg;
  toastVisible.value = true;
  clearTimeout(toastTimer);
  toastTimer = setTimeout(() => {
    toastVisible.value = false;
  }, duration);
}

// ==================== 导航状态 ====================
function updateNavStatus(status, text) {
  navStatusClass.value = status;
  navStatusText.value = text;
}

// ==================== 设备列表渲染 ====================
function renderDeviceList(list) {
  devices.value = list || [];
}

// ==================== 连接设备 ====================
function connectToDevice(deviceId) {
  if (state.value.pairedDeviceId === deviceId) {
    showToast('已连接到该设备');
    return;
  }

  state.value.socket.emit('controller:connect-device', { deviceId }, (response) => {
    if (response.success) {
      state.value.pairedDeviceId = deviceId;
      const device = response.deviceInfo;
      if (device && device.nativeSocket) {
        isNativeMode.value = true;
        updateNavStatus('connecting', '等待画面传输...');
        showToast('已连接到设备（原生模式）');
      } else {
        isNativeMode.value = false;
        createPeerConnection(deviceId);
      }
    } else {
      showToast(response.error || '连接失败');
    }
  });
}

// ==================== WebRTC ====================
async function createPeerConnection(targetDeviceId) {
  if (state.value.peerConnection) {
    state.value.peerConnection.close();
    state.value.peerConnection = null;
  }

  const configuration = {
    iceServers: [
      { urls: 'stun:stun.l.google.com:19302' },
      { urls: 'stun:stun1.l.google.com:19302' },
    ],
  };

  const pc = new RTCPeerConnection(configuration);
  state.value.peerConnection = pc;

  pc.ontrack = (event) => {
    const video = remoteVideoRef.value;
    if (video && video.srcObject !== event.streams[0]) {
      video.srcObject = event.streams[0];
      video.onloadedmetadata = () => {
        state.value.screenWidth = video.videoWidth;
        state.value.screenHeight = video.videoHeight;
        updatePhoneFrameSize();
      };
      video.onplay = () => {
        showPlaceholder.value = false;
        showPhoneFrame.value = true;
        showConnectionInfo.value = true;
        showControlSection.value = true;
        updateNavStatus('online', '已连接');
        state.value.isConnected = true;
        startFpsCounter();
        startLatencyMonitor();
      };
    }
  };

  pc.onicecandidate = (event) => {
    if (event.candidate) {
      state.value.socket.emit('signal:ice-candidate', {
        targetId: targetDeviceId,
        candidate: event.candidate,
      });
    }
  };

  pc.onconnectionstatechange = () => {
    if (pc.connectionState === 'disconnected' || pc.connectionState === 'failed') {
      handleDisconnect();
    }
  };

  const dataChannel = pc.createDataChannel('control', { ordered: true });
  state.value.dataChannel = dataChannel;

  dataChannel.onopen = () => console.log('[DataChannel] 已打开');
  dataChannel.onclose = () => console.log('[DataChannel] 已关闭');

  try {
    const offer = await pc.createOffer();
    await pc.setLocalDescription(offer);
    state.value.socket.emit('signal:offer', {
      targetId: targetDeviceId,
      sdp: offer.sdp,
    });
  } catch (err) {
    console.error('[WebRTC] 创建 Offer 失败:', err);
    showToast('建立连接失败: ' + err.message);
  }
}

// ==================== 手机屏幕尺寸 ====================
function updatePhoneFrameSize() {
  const phoneScreen = phoneScreenRef.value;
  if (!phoneScreen) return;
  const maxWidth = window.innerWidth - 360;
  const maxHeight = window.innerHeight - 80;
  const aspectRatio = state.value.screenWidth / state.value.screenHeight;

  let width, height;
  if (maxWidth / maxHeight > aspectRatio) {
    height = Math.min(maxHeight * 0.85, 700);
    width = height * aspectRatio;
  } else {
    width = Math.min(maxWidth * 0.75, 400);
    height = width / aspectRatio;
  }

  phoneScreen.style.width = width + 'px';
  phoneScreen.style.height = height + 'px';
}

// ==================== 触控 ====================
function getTouchCoords(e) {
  const phoneScreen = phoneScreenRef.value;
  if (!phoneScreen) return { x: 0, y: 0 };
  const rect = phoneScreen.getBoundingClientRect();
  if (rect.width === 0 || rect.height === 0) return { x: 0, y: 0 };

  const sw = state.value.screenWidth || 540;
  const sh = state.value.screenHeight || 960;
  const scaleX = sw / rect.width;
  const scaleY = sh / rect.height;

  return {
    x: Math.round((e.clientX - rect.left) * scaleX),
    y: Math.round((e.clientY - rect.top) * scaleY),
  };
}

function sendTouchEvent(action, coords) {
  if (!state.value.pairedDeviceId) return;

  if (state.value.dataChannel && state.value.dataChannel.readyState === 'open') {
    state.value.dataChannel.send(JSON.stringify({
      type: 'touch',
      data: { action, x: coords.x, y: coords.y, pressure: 0.5 },
    }));
    return;
  }

  if (state.value.socket && state.value.socket.connected) {
    state.value.socket.emit('control:touch', {
      action,
      x: coords.x,
      y: coords.y,
      pressure: 0.5,
    });
  }
}

function showTouchIndicator(x, y) {
  const phoneScreen = phoneScreenRef.value;
  if (!phoneScreen) return;
  const indicator = document.createElement('div');
  indicator.className = 'touch-indicator';
  indicator.style.left = x + 'px';
  indicator.style.top = y + 'px';
  phoneScreen.appendChild(indicator);
  setTimeout(() => indicator.remove(), 300);
}

let isTouching = false;

function onMouseDown(e) {
  e.preventDefault();
  isTouching = true;
  const coords = getTouchCoords(e);
  sendTouchEvent('down', coords);
  const phoneScreen = phoneScreenRef.value;
  if (phoneScreen) {
    const rect = phoneScreen.getBoundingClientRect();
    showTouchIndicator(e.clientX - rect.left, e.clientY - rect.top);
  }
}

function onMouseMove(e) {
  if (!isTouching) return;
  const coords = getTouchCoords(e);
  sendTouchEvent('move', coords);
}

function onMouseUp(e) {
  if (!isTouching) return;
  isTouching = false;
  const coords = getTouchCoords(e);
  sendTouchEvent('up', coords);
}

function onWheel(e) {
  e.preventDefault();
  if (state.value.dataChannel && state.value.dataChannel.readyState === 'open') {
    state.value.dataChannel.send(JSON.stringify({
      type: 'scroll',
      data: { deltaX: e.deltaX, deltaY: e.deltaY, x: e.clientX, y: e.clientY },
    }));
  }
}

// ==================== 按键 ====================
function sendKey(key, code) {
  if (state.value.dataChannel && state.value.dataChannel.readyState === 'open') {
    state.value.dataChannel.send(JSON.stringify({ type: 'key', data: { key, code, action: 'down' } }));
    setTimeout(() => {
      if (state.value.dataChannel && state.value.dataChannel.readyState === 'open') {
        state.value.dataChannel.send(JSON.stringify({ type: 'key', data: { key, code, action: 'up' } }));
      }
    }, 50);
  }
}

function onKeyDown(e) {
  if (!state.value.isConnected) return;
  if (state.value.dataChannel && state.value.dataChannel.readyState === 'open') {
    state.value.dataChannel.send(JSON.stringify({ type: 'key', data: { key: e.key, code: e.code, action: 'down' } }));
  }
}

function onKeyUp(e) {
  if (!state.value.isConnected) return;
  if (state.value.dataChannel && state.value.dataChannel.readyState === 'open') {
    state.value.dataChannel.send(JSON.stringify({ type: 'key', data: { key: e.key, code: e.code, action: 'up' } }));
  }
}

// ==================== FPS 统计 ====================
let fpsInterval = null;

function countFrame() {
  state.value.frameCount++;
  const now = Date.now();
  if (now - state.value.lastFpsTime >= 1000) {
    state.value.currentFps = state.value.frameCount;
    state.value.frameCount = 0;
    state.value.lastFpsTime = now;
  }
}

function startFpsCounter() {
  state.value.frameCount = 0;
  state.value.lastFpsTime = Date.now();
  if (fpsInterval) clearInterval(fpsInterval);
  fpsInterval = setInterval(() => {
    if (state.value.frameCount > 0) {
      state.value.currentFps = state.value.frameCount;
      state.value.frameCount = 0;
      state.value.lastFpsTime = Date.now();
    }
  }, 1000);
}

// ==================== 延迟监控 ====================
let latencyInterval = null;

function startLatencyMonitor() {
  if (latencyInterval) clearInterval(latencyInterval);
  latencyInterval = setInterval(() => {
    if (state.value.dataChannel && state.value.dataChannel.readyState === 'open') {
      state.value.lastPingTime = Date.now();
      state.value.dataChannel.send(JSON.stringify({ type: 'ping', timestamp: state.value.lastPingTime }));
    }
  }, 2000);
}

// ==================== 断开连接 ====================
function handleDisconnect() {
  state.value.isConnected = false;
  state.value.pairedDeviceId = null;

  if (state.value.peerConnection) {
    state.value.peerConnection.close();
    state.value.peerConnection = null;
  }

  const video = remoteVideoRef.value;
  if (video && video.srcObject) {
    video.srcObject.getTracks().forEach((t) => t.stop());
    video.srcObject = null;
  }

  state.value.dataChannel = null;

  const canvas = remoteCanvasRef.value;
  if (canvas) {
    const ctx = canvas.getContext('2d');
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    canvas.style.display = 'none';
  }
  if (video) video.style.display = 'none';

  showPhoneFrame.value = false;
  showPlaceholder.value = true;
  showConnectionInfo.value = false;
  showControlSection.value = false;
  isNativeMode.value = false;
  updateNavStatus('offline', '未连接');
  devices.value = [];
}

// ==================== 帧数据处理（原生APP） ====================
function handleFrameData(data) {
  if (data.deviceId !== state.value.pairedDeviceId) return;

  const canvas = remoteCanvasRef.value;
  const video = remoteVideoRef.value;
  if (!canvas) return;
  const ctx = canvas.getContext('2d');

  if (canvas.style.display === 'none') {
    if (video) video.style.display = 'none';
    canvas.style.display = 'block';
    showPlaceholder.value = false;
    showPhoneFrame.value = true;
    showConnectionInfo.value = true;
    showControlSection.value = true;
    updateNavStatus('online', '已连接（原生模式）');
    state.value.isConnected = true;
    state.value.screenWidth = data.width;
    state.value.screenHeight = data.height;
    updatePhoneFrameSize();
    startFpsCounter();
  }

  const img = new Image();
  img.onload = () => {
    canvas.width = data.width;
    canvas.height = data.height;
    ctx.drawImage(img, 0, 0);
    countFrame();
  };
  img.src = 'data:image/jpeg;base64,' + data.imageData;
}

// ==================== Socket.IO ====================
function initSocket() {
  const socket = io(serverUrl.value, { reconnection: true });
  state.value.socket = socket;

  socket.on('connect', () => {
    console.log('[Socket] 已连接');
    socket.emit('controller:join', {});
  });

  socket.on('controller:joined', (data) => {
    renderDeviceList(data.devices || []);
  });

  socket.on('device:list-update', (data) => {
    renderDeviceList(data.devices || []);
  });

  socket.on('device:offline', (data) => {
    if (state.value.pairedDeviceId === data.deviceId) {
      showToast('设备已离线');
      handleDisconnect();
    }
  });

  socket.on('frame:data', handleFrameData);

  socket.on('signal:answer', async (data) => {
    const pc = state.value.peerConnection;
    if (!pc) return;
    try {
      await pc.setRemoteDescription(new RTCSessionDescription({ type: 'answer', sdp: data.sdp }));
    } catch (err) {
      console.error('[信令] 处理 Answer 失败:', err);
    }
  });

  socket.on('signal:ice-candidate', async (data) => {
    const pc = state.value.peerConnection;
    if (!pc || !data.candidate) return;
    try {
      await pc.addIceCandidate(new RTCIceCandidate(data.candidate));
    } catch (err) {
      console.error('[ICE] 添加候选失败:', err);
    }
  });

  socket.on('disconnect', () => {
    updateNavStatus('offline', '服务器断开');
    handleDisconnect();
  });
}

// ==================== 快捷操作 ====================
function doHome() { sendKey('Home', 'Home'); showToast('主页'); }
function doBack() { sendKey('Backspace', 'Backspace'); showToast('返回'); }
function doRecent() { sendKey('AppSwitch', 'AppSwitch'); showToast('最近任务'); }
function doPower() { sendKey('Power', 'Power'); showToast('电源键'); }
function doScreenshot() {
  if (state.value.dataChannel && state.value.dataChannel.readyState === 'open') {
    state.value.dataChannel.send(JSON.stringify({ type: 'screenshot', data: {} }));
    showToast('截图请求已发送');
  }
}
function doVolumeUp() { sendKey('AudioVolumeUp', 'AudioVolumeUp'); showToast('音量+'); }
function doVolumeDown() { sendKey('AudioVolumeDown', 'AudioVolumeDown'); showToast('音量-'); }
function doRotate() {
  if (state.value.dataChannel && state.value.dataChannel.readyState === 'open') {
    state.value.dataChannel.send(JSON.stringify({ type: 'rotate', data: {} }));
    showToast('旋转屏幕');
  }
}
function doDisconnect() {
  handleDisconnect();
  showToast('已断开连接');
}

// ==================== 生命周期 ====================
onMounted(() => {
  nextTick(() => {
    initSocket();
  });
  window.addEventListener('mousemove', onMouseMove);
  window.addEventListener('mouseup', onMouseUp);
  window.addEventListener('keydown', onKeyDown);
  window.addEventListener('keyup', onKeyUp);
  window.addEventListener('resize', () => {
    if (state.value.isConnected) updatePhoneFrameSize();
  });
});

onUnmounted(() => {
  handleDisconnect();
  if (state.value.socket) state.value.socket.disconnect();
  window.removeEventListener('mousemove', onMouseMove);
  window.removeEventListener('mouseup', onMouseUp);
  window.removeEventListener('keydown', onKeyDown);
  window.removeEventListener('keyup', onKeyUp);
  if (fpsInterval) clearInterval(fpsInterval);
  if (latencyInterval) clearInterval(latencyInterval);
});
</script>

<template>
  <div class="remote-phone-control">
    <!-- Toast -->
    <div class="toast" :class="{ show: toastVisible }">{{ toastMsg }}</div>

    <!-- 主布局 -->
    <div class="main-container">
      <!-- 左侧面板 -->
      <div class="side-panel">
        <!-- 服务器配置 -->
        <div class="panel-section">
          <div class="panel-title">⚙️ 服务器配置</div>
          <div class="config-row">
            <input
              v-model="serverUrl"
              class="input-server"
              placeholder="信令服务器地址"
              @blur="() => { handleDisconnect(); state.value.socket?.disconnect(); nextTick(initSocket); }"
            />
          </div>
        </div>

        <!-- 在线设备 -->
        <div class="panel-section">
          <div class="panel-title">📋 在线设备 ({{ devices.length }})</div>
          <div class="device-list">
            <div v-if="devices.length === 0" class="no-devices">暂无在线设备</div>
            <div
              v-for="d in devices"
              :key="d.id"
              class="device-item"
              :class="{ selected: state.pairedDeviceId === d.id }"
              @click="connectToDevice(d.id)"
            >
              <span class="device-icon">📱</span>
              <div>
                <div class="device-name">{{ d.name }}</div>
                <div class="device-model">{{ d.model }}</div>
              </div>
              <span class="device-status online">在线</span>
            </div>
          </div>
        </div>

        <!-- 快捷操作 -->
        <div v-show="showControlSection" class="panel-section">
          <div class="panel-title">🎮 快捷操作</div>
          <div class="control-buttons">
            <button class="ctrl-btn" @click="doHome"><span class="ctrl-icon">🏠</span> 主页</button>
            <button class="ctrl-btn" @click="doBack"><span class="ctrl-icon">⬅️</span> 返回</button>
            <button class="ctrl-btn" @click="doRecent"><span class="ctrl-icon">⬜</span> 最近任务</button>
            <button class="ctrl-btn" @click="doPower"><span class="ctrl-icon">🔒</span> 锁屏</button>
            <button class="ctrl-btn" @click="doScreenshot"><span class="ctrl-icon">📸</span> 截图</button>
            <button class="ctrl-btn" @click="doVolumeUp"><span class="ctrl-icon">🔊</span> 音量+</button>
            <button class="ctrl-btn" @click="doVolumeDown"><span class="ctrl-icon">🔉</span> 音量-</button>
            <button class="ctrl-btn" @click="doRotate"><span class="ctrl-icon">🔄</span> 旋转</button>
          </div>
          <div style="margin-top:12px;">
            <button class="btn btn-danger" style="width:100%;" @click="doDisconnect">断开连接</button>
          </div>
        </div>
      </div>

      <!-- 右侧屏幕显示 -->
      <div class="screen-area">
        <!-- 空状态 -->
        <div v-show="showPlaceholder" class="screen-placeholder">
          <div class="icon">📱</div>
          <div class="title">等待设备连接</div>
          <div class="subtitle">
            1. 在手机上打开 APP 并启动投屏<br>
            2. 设备出现在左侧列表中后点击连接<br>
            3. 连接成功后即可查看和控制手机屏幕
          </div>
        </div>

        <!-- 手机屏幕 -->
        <div v-show="showPhoneFrame" class="phone-frame">
          <div class="phone-notch"></div>
          <div
            ref="phoneScreenRef"
            class="phone-screen"
            @mousedown="onMouseDown"
            @wheel="onWheel"
          >
            <video ref="remoteVideoRef" autoplay playsinline muted style="display:none;"></video>
            <canvas ref="remoteCanvasRef" style="display:none; width:100%; height:100%;"></canvas>
          </div>
        </div>

        <!-- 连接信息 -->
        <div v-show="showConnectionInfo" class="connection-overlay">
          <span>FPS: <span class="fps">{{ state.currentFps || '--' }}</span></span>
          <span>延迟: <span class="latency">{{ state.latency || '--' }}</span>ms</span>
          <span>{{ state.screenWidth }}x{{ state.screenHeight }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.remote-phone-control {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #0d1117;
  color: #c9d1d9;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
}

* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

/* Toast */
.toast {
  position: fixed;
  top: 70px;
  left: 50%;
  transform: translateX(-50%);
  background: rgba(0,0,0,0.9);
  color: #fff;
  padding: 10px 24px;
  border-radius: 20px;
  font-size: 13px;
  z-index: 1000;
  opacity: 0;
  transition: opacity 0.3s;
  pointer-events: none;
}
.toast.show { opacity: 1; }

/* 主布局 */
.main-container {
  display: flex;
  flex: 1;
  height: calc(100vh - 120px);
}

/* 左侧面板 */
.side-panel {
  width: 300px;
  min-width: 300px;
  background: #161b22;
  border-right: 1px solid #30363d;
  display: flex;
  flex-direction: column;
  overflow-y: auto;
}

.panel-section {
  padding: 16px;
  border-bottom: 1px solid #30363d;
}

.panel-title {
  font-size: 12px;
  text-transform: uppercase;
  color: #8b949e;
  font-weight: 600;
  letter-spacing: 0.5px;
  margin-bottom: 12px;
}

.config-row {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.input-server {
  background: #0d1117;
  border: 1px solid #30363d;
  border-radius: 8px;
  padding: 8px 12px;
  color: #f0f6fc;
  font-size: 13px;
  outline: none;
  transition: border-color 0.2s;
}
.input-server:focus {
  border-color: #58a6ff;
  box-shadow: 0 0 0 3px rgba(88,166,255,0.15);
}

/* 设备列表 */
.device-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.device-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  background: #0d1117;
  border: 1px solid #30363d;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.15s;
}
.device-item:hover {
  border-color: #58a6ff;
  background: #161b22;
}
.device-item.selected {
  border-color: #3fb950;
  background: rgba(63,185,80,0.08);
}
.device-item .device-icon { font-size: 20px; }
.device-item .device-name { font-size: 14px; color: #f0f6fc; }
.device-item .device-model { font-size: 11px; color: #8b949e; }
.device-item .device-status {
  margin-left: auto;
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 10px;
  background: rgba(63,185,80,0.15);
  color: #3fb950;
}

.no-devices {
  text-align: center;
  padding: 20px;
  color: #484f58;
  font-size: 13px;
}

/* 控制按钮 */
.control-buttons {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
}

.ctrl-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  padding: 12px 8px;
  background: #21262d;
  border: 1px solid #30363d;
  border-radius: 8px;
  color: #c9d1d9;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.15s;
}
.ctrl-btn:hover { background: #30363d; }
.ctrl-btn:active { transform: scale(0.97); }
.ctrl-btn .ctrl-icon { font-size: 22px; }

.btn {
  padding: 10px 20px;
  border-radius: 8px;
  border: 1px solid #30363d;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}
.btn-danger {
  background: #da3633;
  color: #fff;
  border-color: rgba(240,246,252,0.1);
}
.btn-danger:hover { background: #f85149; }

/* 右侧屏幕显示区 */
.screen-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: #010409;
  position: relative;
  overflow: hidden;
}

.screen-placeholder {
  text-align: center;
  color: #484f58;
}
.screen-placeholder .icon {
  font-size: 80px;
  margin-bottom: 16px;
  opacity: 0.3;
}
.screen-placeholder .title {
  font-size: 20px;
  color: #8b949e;
  margin-bottom: 8px;
}
.screen-placeholder .subtitle {
  font-size: 14px;
  line-height: 1.6;
}

/* 手机屏幕容器 */
.phone-frame {
  position: relative;
  background: #1c1c1e;
  border-radius: 24px;
  padding: 12px;
  box-shadow:
    0 0 0 2px #333,
    0 0 0 4px #1c1c1e,
    0 0 0 6px #222,
    0 20px 60px rgba(0,0,0,0.5);
}

.phone-notch {
  position: absolute;
  top: 12px;
  left: 50%;
  transform: translateX(-50%);
  width: 100px;
  height: 20px;
  background: #1c1c1e;
  border-radius: 0 0 14px 14px;
  z-index: 2;
}

.phone-screen {
  background: #000;
  border-radius: 14px;
  overflow: hidden;
  position: relative;
  cursor: crosshair;
}

.phone-screen video {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: contain;
  pointer-events: none;
}

/* 触摸指示器 */
:deep(.touch-indicator) {
  position: absolute;
  width: 36px;
  height: 36px;
  border: 2px solid rgba(88,166,255,0.8);
  border-radius: 50%;
  pointer-events: none;
  transform: translate(-50%, -50%);
  animation: touchPulse 0.3s ease-out forwards;
  z-index: 10;
}

@keyframes touchPulse {
  0% { transform: translate(-50%, -50%) scale(0.5); opacity: 1; }
  100% { transform: translate(-50%, -50%) scale(1.5); opacity: 0; }
}

/* 连接信息浮层 */
.connection-overlay {
  position: absolute;
  bottom: 20px;
  left: 50%;
  transform: translateX(-50%);
  background: rgba(22,27,34,0.9);
  backdrop-filter: blur(10px);
  border: 1px solid #30363d;
  border-radius: 12px;
  padding: 8px 16px;
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 12px;
  color: #8b949e;
}

.connection-overlay .fps { color: #3fb950; font-weight: 600; }
.connection-overlay .latency { color: #d29922; font-weight: 600; }

/* 响应式 */
@media (max-width: 768px) {
  .main-container { flex-direction: column; }
  .side-panel {
    width: 100%;
    min-width: 100%;
    max-height: 300px;
    border-right: none;
    border-bottom: 1px solid #30363d;
  }
  .phone-frame {
    margin: 10px;
    padding: 8px;
    border-radius: 16px;
  }
  .phone-screen { border-radius: 8px; }
}
</style>
