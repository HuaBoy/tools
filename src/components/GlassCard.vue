<script setup>
defineProps({
  title: {
    type: String,
    default: ''
  },
  noPadding: {
    type: Boolean,
    default: false
  }
})
</script>

<template>
  <div class="glass-card-wrapper">
    <div v-if="title" class="card-header">
      <span class="title-text">{{ title }}</span>
    </div>
    <div class="card-body" :class="{ 'no-padding': noPadding }">
      <slot></slot>
    </div>
  </div>
</template>

<style scoped>
.glass-card-wrapper {
  background: var(--bg-card);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  box-shadow: var(--shadow-card);
  transition: all 0.3s ease;
  overflow-x: auto;
  overflow-y: visible;

  &:hover {
    border-color: var(--border-color-hover);
    box-shadow: var(--shadow-card-hover);
  }
}

.card-header {
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-color);
  background: rgba(22, 93, 255, 0.05);
}

.title-text {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

.card-body {
  padding: 20px;

  &.no-padding {
    padding: 0;
  }
}
</style>