<script setup>import { ref, onMounted, watch, onUnmounted } from 'vue';
import * as echarts from 'echarts';
const props = defineProps({
 option: {
 type: Object,
 required: true
 },
 height: {
 type: String,
 default: '400px'
 },
 loading: {
 type: Boolean,
 default: false
 }
});
const chartRef = ref(null);
let chartInstance = null;
const initChart = () => {
 if (!chartRef.value)
 return;
 chartInstance = echarts.init(chartRef.value, 'dark');
 chartInstance.setOption(props.option);
};
const handleResize = () => {
 chartInstance?.resize();
};
watch(() => props.option, (newOption) => {
 if (chartInstance) {
 chartInstance.setOption(newOption, true);
 }
}, { deep: true });
watch(() => props.loading, (isLoading) => {
 if (chartInstance) {
 if (isLoading) {
 chartInstance.showLoading({
 text: '加载中...',
 color: '#165DFF',
 maskColor: 'rgba(13, 17, 23, 0.8)',
 textColor: '#fff'
 });
 }
 else {
 chartInstance.hideLoading();
 }
 }
});
onMounted(() => {
 initChart();
 window.addEventListener('resize', handleResize);
});
onUnmounted(() => {
 window.removeEventListener('resize', handleResize);
 chartInstance?.dispose();
});
defineExpose({
 refresh: () => {
 if (chartInstance) {
 chartInstance.resize();
 }
 }
});
</script>

<template>
  <div ref="chartRef" class="base-chart" :style="{ height }"></div>
</template>

<style scoped>
.base-chart {
  width: 100%;
  min-height: 300px;
}
</style>
