<template>
  <div class="task-progress">
    <div class="task-bar">
      <div class="task-fill" :style="{ width: percent + '%' }" :class="status"></div>
    </div>
    <span class="task-label">{{ label }}</span>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  current: { type: Number, default: 0 },
  total: { type: Number, default: 1 },
  label: { type: String, default: '' },
  status: { type: String, default: 'running' }
})

const percent = computed(() => props.total > 0 ? Math.round(props.current / props.total * 100) : 0)
</script>

<style scoped>
.task-progress { display: flex; align-items: center; gap: 10px; }
.task-bar { flex: 1; height: 6px; background: rgba(22,93,255,0.1); border-radius: 3px; overflow: hidden; }
.task-fill { height: 100%; border-radius: 3px; transition: width 0.5s; }
.task-fill.running { background: linear-gradient(90deg, #165DFF, #4080FF); animation: shimmer 1.5s infinite; }
@keyframes shimmer { 0% { opacity: 0.7 } 50% { opacity: 1 } 100% { opacity: 0.7 } }
.task-fill.done { background: #00B42A; }
.task-fill.failed { background: #F53F3F; }
.task-label { font-size: 12px; color: var(--text-tertiary); white-space: nowrap; }
</style>