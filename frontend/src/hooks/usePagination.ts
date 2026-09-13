// usePagination：列表分页组合式函数
import { computed, ref } from 'vue'

export function usePagination(defaultPageSize = 10) {
  const page = ref(1)
  const pageSize = ref(defaultPageSize)
  const total = ref(0)

  const pagination = computed(() => ({
    current: page.value,
    pageSize: pageSize.value,
    total: total.value,
  }))

  function handlePageChange(p: number) {
    page.value = p
  }

  function handleSizeChange(size: number) {
    pageSize.value = size
    page.value = 1
  }

  return { page, pageSize, total, pagination, handlePageChange, handleSizeChange }
}
