<template>
  <span class="fw-badge" :class="'fw-' + level">
    <span class="fw-icon">⬆</span>
    固件可升级 v{{ current }} → v{{ latest }}
  </span>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  current: String,
  latest: String,
  level: { type: String, default: 'info' } // info / warn / critical
})
</script>

<style scoped>
.fw-badge {
  display: inline-flex; align-items: center; gap: 4px; font-size: 12px;
  padding: 4px 10px; border-radius: 6px;
}
.fw-info { background: rgba(22,93,255,0.08); color: #165DFF; }
.fw-warn { background: rgba(255,125,0,0.08); color: #FF7D00; }
.fw-critical { background: rgba(245,63,63,0.08); color: #F53F3F; }
.fw-icon { font-size: 10px; }
</style>