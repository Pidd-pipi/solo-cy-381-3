<!-- SplitTypeTag：分摊方式标签 -->
<template>
  <StatusBadge :status="type" kind="split" />
</template>

<script setup lang="ts">
defineProps<{ type: string }>()
</script>
