<template>
  <div class="data-table-wrap">
    <table class="data-table" v-if="data.length">
      <thead>
        <tr><th v-for="col in columns" :key="col.key" :style="{ width: col.width }">{{ col.label }}</th></tr>
      </thead>
      <tbody>
        <tr v-for="(row, i) in data" :key="i">
          <td v-for="col in columns" :key="col.key">
            <slot :name="'cell-' + col.key" :row="row" :value="row[col.key]">{{ row[col.key] }}</slot>
          </td>
        </tr>
      </tbody>
    </table>
    <div v-else class="data-empty">暂无数据</div>
  </div>
</template>

<script setup>
defineProps({
  columns: { type: Array, required: true },
  data: { type: Array, default: () => [] }
})
</script>

<style scoped>
.data-table-wrap { width: 100%; overflow-x: auto; }
.data-table { width: 100%; border-collapse: collapse; font-size: 13px; }
.data-table th { padding: 12px 14px; text-align: left; font-weight: 600; color: var(--text-secondary); background: rgba(22,93,255,0.04); border-bottom: 2px solid var(--border-color); white-space: nowrap; }
.data-table td { padding: 10px 14px; border-bottom: 1px solid var(--border-color); color: var(--text-primary); }
.data-table tr:hover td { background: rgba(22,93,255,0.02); }
.data-empty { text-align: center; padding: 40px; color: var(--text-tertiary); font-size: 14px; }
</style>