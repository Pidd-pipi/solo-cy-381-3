<!-- DataTable：通用数据表格容器（加载态 + 空态 + 分页） -->
<template>
  <div class="data-table">
    <el-table v-loading="loading" :data="data" border stripe>
      <slot />
      <template #empty>
        <el-empty description="暂无数据" :image-size="80" />
      </template>
    </el-table>
    <div v-if="showPagination && total > 0" class="data-table__pagination">
      <el-pagination
        layout="total, prev, pager, next"
        :total="total"
        :current-page="page"
        :page-size="pageSize"
        @current-change="handlePageChange"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    data: unknown[]
    loading?: boolean
    total?: number
    page?: number
    pageSize?: number
    showPagination?: boolean
  }>(),
  { loading: false, total: 0, page: 1, pageSize: 10, showPagination: false },
)

const emit = defineEmits<{ (e: 'page-change', page: number): void }>()

function handlePageChange(p: number) {
  emit('page-change', p)
}
</script>

<style scoped>
.data-table__pagination {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
</style>
