<!-- EmptyState：通用空状态占位 -->
<template>
  <div class="empty-state">
    <el-empty :description="description" :image-size="100">
      <slot v-if="$slots.default" />
    </el-empty>
  </div>
</template>

<script setup lang="ts">
defineProps<{ description?: string }>()
</script>

<style scoped>
.empty-state {
  padding: 24px 0;
}
</style>
