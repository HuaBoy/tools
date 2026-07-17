// 分页 composable
import { ref, computed } from 'vue'

export function usePagination(fetchFn, pageSize = 20) {
  const page = ref(1)
  const size = ref(pageSize)
  const total = ref(0)
  const list = ref([])
  const loading = ref(false)

  const totalPages = computed(() => Math.ceil(total.value / size.value))

  async function loadData(params = {}) {
    loading.value = true
    try {
      const res = await fetchFn({ page: page.value, size: size.value, ...params })
      list.value = res.data || res.list || []
      total.value = res.total || 0
    } finally {
      loading.value = false
    }
  }

  function goPage(p) { page.value = p; loadData() }
  function changeSize(s) { size.value = s; page.value = 1; loadData() }

  return { page, size, total, list, loading, totalPages, loadData, goPage, changeSize }
}