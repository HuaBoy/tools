<template>
  <div class="device-card" :class="{ offline: !online }">
    <div class="device-status-dot" :class="{ online }"></div>
    <div class="device-info">
      <div class="device-name">{{ device.name || device.model }}</div>
      <div class="device-id">{{ device.serialNo || device.id }}</div>
      <div class="device-meta" v-if="device.chipModel">
        <span class="chip-tag">{{ device.chipModel }}</span>
        <span class="fw-tag">固件 v{{ device.firmwareVersion }}</span>
      </div>
    </div>
    <div class="device-actions">
      <slot name="actions"></slot>
    </div>
  </div>
</template>

<script setup>
defineProps({
  device: { type: Object, required: true },
  online: { type: Boolean, default: true }
})
</script>

<style scoped>
.device-card {
  display: flex; align-items: center; gap: 14px; padding: 14px 18px;
  border: 1px solid var(--border-color); border-radius: 12px;
  background: var(--bg-card); transition: all 0.2s;
}
.device-card:hover { border-color: rgba(22,93,255,0.3); box-shadow: 0 2px 8px rgba(22,93,255,0.08); }
.device-card.offline { opacity: 0.5; }
.device-status-dot { width: 10px; height: 10px; border-radius: 50%; background: #ccc; flex-shrink: 0; }
.device-status-dot.online { background: #00B42A; }
.device-info { flex: 1; min-width: 0; }
.device-name { font-size: 14px; font-weight: 600; color: var(--text-primary); }
.device-id { font-size: 12px; color: var(--text-tertiary); margin-top: 2px; }
.device-meta { display: flex; gap: 6px; margin-top: 4px; }
.chip-tag { font-size: 11px; padding: 1px 6px; border-radius: 4px; background: rgba(22,93,255,0.1); color: #165DFF; }
.fw-tag { font-size: 11px; padding: 1px 6px; border-radius: 4px; background: rgba(0,180,42,0.1); color: #00B42A; }
.device-actions { display: flex; gap: 6px; flex-shrink: 0; }
</style>