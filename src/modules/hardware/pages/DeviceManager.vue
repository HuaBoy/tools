<template>
  <div class="coming-page">
    <div class="coming-card">
      <div class="coming-icon">🔌</div>
      <h2>设备管理</h2>
      <p>统一管理所有起爆器硬件设备，支持蓝牙连接、状态监控、远程诊断</p>
      <p class="coming-hint">该模块正在开发中，将通过 APP 端（Android APK / 微信小程序）提供完整硬件管理能力</p>
      <div class="coming-features">
        <div class="feature-tag">芯片管理</div>
        <div class="feature-tag">模块管理</div>
        <div class="feature-tag">BLE 蓝牙</div>
        <div class="feature-tag">固件 OTA</div>
        <div class="feature-tag">MQTT 通信</div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.coming-page { display: flex; align-items: center; justify-content: center; min-height: calc(100vh - 120px); padding: 40px; }
.coming-card { text-align: center; max-width: 560px; padding: 48px 40px; background: var(--bg-card); border: 1px solid var(--border-color); border-radius: 20px; }
.coming-icon { font-size: 64px; margin-bottom: 20px; }
h2 { font-size: 22px; font-weight: 700; color: var(--text-primary); margin: 0 0 12px; }
p { font-size: 14px; color: var(--text-secondary); margin: 0 0 8px; line-height: 1.7; }
.coming-hint { font-size: 12px; color: var(--text-tertiary); margin-top: 14px; }
.coming-features { display: flex; gap: 8px; justify-content: center; flex-wrap: wrap; margin-top: 20px; }
.feature-tag { font-size: 11px; padding: 4px 12px; border-radius: 20px; background: rgba(22,93,255,0.08); color: #165DFF; font-weight: 500; }
</style>