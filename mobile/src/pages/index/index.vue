<template>
  <view class="page">
    <!-- AI 输入区 -->
    <view class="ai-banner">
      <view class="ai-pulse"></view>
      <text class="ai-title">AI 运维执行平台</text>
      <text class="ai-subtitle">说话，AI 帮你干活</text>
    </view>

    <view class="search-section">
      <textarea
        v-model="inputText"
        class="ai-textarea"
        placeholder="说出你的需求...&#10;例如：诊断批次号 B20240701-003"
        :maxlength="500"
        auto-height
      />
      <button class="send-btn" :disabled="!inputText.trim()" @tap="handleSend">
        <text class="send-icon">➤</text>
      </button>
    </view>

    <!-- 场景卡片 -->
    <view class="section">
      <text class="section-title">常用 AI 场景</text>
      <view class="scenario-grid">
        <view class="scenario-card" v-for="s in scenarios" :key="s.id" @tap="handleScenario(s)">
          <text class="scenario-icon">{{ s.icon }}</text>
          <view class="scenario-info">
            <text class="scenario-name">{{ s.title }}</text>
            <text class="scenario-desc">{{ s.desc }}</text>
          </view>
        </view>
      </view>
    </view>

    <!-- 快速工具 -->
    <view class="section">
      <text class="section-title">快速工具</text>
      <view class="tools-row">
        <view class="tool-item" v-for="t in tools" :key="t.name" @tap="handleTool(t)">
          <text class="tool-icon">{{ t.icon }}</text>
          <text class="tool-name">{{ t.name }}</text>
        </view>
      </view>
    </view>

    <!-- 硬件管理入口 -->
    <view class="hardware-entry" @tap="goHardware">
      <view class="hw-left">
        <text class="hw-icon">🔌</text>
        <view class="hw-info">
          <text class="hw-title">硬件管理</text>
          <text class="hw-desc">芯片 · 模块 · 固件 · BLE</text>
        </view>
      </view>
      <text class="hw-badge">NEW</text>
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue'

const inputText = ref('')

const scenarios = [
  { id: 'diagnose', icon: '🔍', title: '批次异常诊断', desc: 'AI自动查数据→分析→出结论' },
  { id: 'report', icon: '📋', title: '一键生成报告', desc: '自动生成质量分析报告' },
  { id: 'trace', icon: '🔗', title: '全链路追溯', desc: '追踪完整生产链路' },
  { id: 'log', icon: '📄', title: '日志智能分析', desc: '解密并AI分析日志' }
]

const tools = [
  { name: '扫码', icon: '📷', url: '/pages/tools/scan' },
  { name: '翻译', icon: '🌐', url: '/pages/tools/translate' },
  { name: '数据', icon: '📊', url: '/pages/data/index' },
  { name: '知识库', icon: '💡', url: '/pages/knowledge/index' }
]

function handleSend() {
  uni.showToast({ title: 'AI分析中...', icon: 'loading' })
  setTimeout(() => {
    uni.showToast({ title: '功能开发中', icon: 'none' })
  }, 1500)
}

function handleScenario(s) {
  uni.navigateTo({ url: '/pages/diagnosis/index' })
}

function handleTool(t) {
  uni.showToast({ title: t.name + ' 开发中', icon: 'none' })
}

function goHardware() {
  uni.switchTab({ url: '/pages/hardware/index' })
}
</script>

<style scoped>
.page { padding: 20rpx 30rpx; min-height: 100vh; }
.ai-banner { display: flex; flex-direction: column; align-items: center; padding: 40rpx 0; position: relative; }
.ai-pulse { width: 80rpx; height: 80rpx; border-radius: 50%; background: linear-gradient(135deg, #165DFF, #4080FF); margin-bottom: 20rpx; box-shadow: 0 0 40rpx rgba(22,93,255,0.3); animation: pulse 2s infinite; }
@keyframes pulse { 0%,100% { transform: scale(1); opacity: 1 } 50% { transform: scale(1.05); opacity: 0.8 } }
.ai-title { font-size: 40rpx; font-weight: 700; color: #1D2129; }
.ai-subtitle { font-size: 26rpx; color: #86909C; margin-top: 8rpx; }

.search-section { display: flex; align-items: flex-start; gap: 16rpx; background: #fff; border-radius: 24rpx; padding: 24rpx; box-shadow: 0 4rpx 20rpx rgba(0,0,0,0.06); margin: 20rpx 0 30rpx; }
.ai-textarea { flex: 1; font-size: 28rpx; min-height: 80rpx; line-height: 1.6; }
.send-btn { width: 80rpx; height: 80rpx; border-radius: 20rpx; background: linear-gradient(135deg, #165DFF, #0F4CD0); display: flex; align-items: center; justify-content: center; border: none; flex-shrink: 0; }
.send-btn[disabled] { opacity: 0.4; }
.send-icon { color: #fff; font-size: 32rpx; }

.section { margin-bottom: 30rpx; }
.section-title { font-size: 26rpx; font-weight: 600; color: #86909C; margin-bottom: 16rpx; display: block; }

.scenario-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 16rpx; }
.scenario-card { background: #fff; border-radius: 20rpx; padding: 24rpx; display: flex; flex-direction: column; gap: 12rpx; box-shadow: 0 2rpx 12rpx rgba(0,0,0,0.04); }
.scenario-icon { font-size: 40rpx; }
.scenario-name { font-size: 28rpx; font-weight: 600; color: #1D2129; }
.scenario-desc { font-size: 22rpx; color: #86909C; line-height: 1.4; }

.tools-row { display: flex; gap: 16rpx; }
.tool-item { flex: 1; background: #fff; border-radius: 20rpx; padding: 24rpx 16rpx; display: flex; flex-direction: column; align-items: center; gap: 8rpx; box-shadow: 0 2rpx 12rpx rgba(0,0,0,0.04); }
.tool-icon { font-size: 36rpx; }
.tool-name { font-size: 22rpx; color: #4E5969; }

.hardware-entry { display: flex; align-items: center; justify-content: space-between; background: linear-gradient(135deg, rgba(22,93,255,0.06), rgba(15,76,208,0.03)); border: 1px solid rgba(22,93,255,0.15); border-radius: 24rpx; padding: 28rpx; margin-top: 20rpx; }
.hw-left { display: flex; align-items: center; gap: 16rpx; }
.hw-icon { font-size: 44rpx; }
.hw-title { font-size: 28rpx; font-weight: 600; color: #1D2129; }
.hw-desc { font-size: 22rpx; color: #86909C; margin-top: 4rpx; display: block; }
.hw-badge { font-size: 20rpx; padding: 4rpx 16rpx; border-radius: 20rpx; background: #165DFF; color: #fff; font-weight: 600; }
</style>