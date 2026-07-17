<template>
  <view class="page">
    <view class="header">
      <text class="title">硬件管理</text>
      <text class="desc">芯片 · 模块 · 设备 · 固件</text>
    </view>

    <view class="card-list">
      <view class="card" @tap="goChip">
        <text class="card-icon">💾</text>
        <view class="card-info">
          <text class="card-title">芯片管理</text>
          <text class="card-desc">注册、绑定、状态追踪</text>
        </view>
        <text class="card-arrow">›</text>
      </view>
      <view class="card">
        <text class="card-icon">📦</text>
        <view class="card-info">
          <text class="card-title">模块管理</text>
          <text class="card-desc">功能模块注册与配置</text>
        </view>
        <text class="card-arrow">›</text>
      </view>
      <view class="card" @tap="goFirmware">
        <text class="card-icon">⬆️</text>
        <view class="card-info">
          <text class="card-title">固件升级</text>
          <text class="card-desc">OTA在线升级与回滚</text>
        </view>
        <text class="card-arrow">›</text>
      </view>
    </view>

    <view class="ble-section">
      <text class="section-title">🔵 蓝牙设备</text>
      <view class="ble-empty">
        <text class="ble-hint">点击「扫描设备」开始搜索附近的起爆器设备</text>
        <button class="scan-btn" @tap="scanBLE">扫描设备</button>
      </view>
    </view>
  </view>
</template>

<script setup>
function goChip() { uni.navigateTo({ url: '/pages/hardware/chip/index' }) }
function goFirmware() { uni.navigateTo({ url: '/pages/hardware/firmware/index' }) }

function scanBLE() {
  // #ifdef APP-PLUS
  uni.openBluetoothAdapter({
    success() {
      uni.startBluetoothDevicesDiscovery({
        success() { uni.showToast({ title: '正在扫描蓝牙设备...', icon: 'loading' }) }
      })
    },
    fail() { uni.showToast({ title: '请开启蓝牙', icon: 'none' }) }
  })
  // #endif
  // #ifdef MP-WEIXIN
  wx.openBluetoothAdapter({
    success() { wx.startBluetoothDevicesDiscovery({ success() { uni.showToast({ title: '正在扫描...', icon: 'loading' }) } }) },
    fail() { uni.showToast({ title: '请开启蓝牙', icon: 'none' }) }
  })
  // #endif
}
</script>

<style scoped>
.page { padding: 30rpx; }
.header { margin-bottom: 30rpx; }
.title { font-size: 40rpx; font-weight: 700; color: #1D2129; display: block; }
.desc { font-size: 26rpx; color: #86909C; margin-top: 8rpx; display: block; }

.card-list { display: flex; flex-direction: column; gap: 16rpx; margin-bottom: 30rpx; }
.card { display: flex; align-items: center; gap: 20rpx; background: #fff; border-radius: 20rpx; padding: 28rpx; box-shadow: 0 2rpx 12rpx rgba(0,0,0,0.04); }
.card-icon { font-size: 44rpx; }
.card-info { flex: 1; }
.card-title { font-size: 30rpx; font-weight: 600; color: #1D2129; display: block; }
.card-desc { font-size: 24rpx; color: #86909C; margin-top: 4rpx; }
.card-arrow { font-size: 36rpx; color: #C9CDD4; }

.ble-section { background: #fff; border-radius: 20rpx; padding: 28rpx; box-shadow: 0 2rpx 12rpx rgba(0,0,0,0.04); }
.section-title { font-size: 28rpx; font-weight: 600; color: #1D2129; display: block; margin-bottom: 20rpx; }
.ble-empty { text-align: center; padding: 40rpx 0; }
.ble-hint { font-size: 24rpx; color: #86909C; }
.scan-btn { margin-top: 20rpx; background: linear-gradient(135deg, #165DFF, #0F4CD0); color: #fff; border-radius: 40rpx; font-size: 28rpx; padding: 16rpx 48rpx; border: none; }
</style>