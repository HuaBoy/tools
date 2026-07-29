<script setup>
import GlassCard from '@/components/GlassCard.vue'
import DeviceVersionTab from './DeviceVersionTab.vue'
</script>

<template>
  <div class="production-history">
    <GlassCard title="产线履历">
      <DeviceVersionTab />
    </GlassCard>
  </div>
</template>

<style scoped>
.production-history {
  max-width: 100%;
}
</style>
