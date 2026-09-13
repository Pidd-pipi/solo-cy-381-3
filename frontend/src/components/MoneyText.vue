<!-- MoneyText：金额文本（¥ + 千分位两位小数） -->
<template>
  <span class="money-text" :class="colorClass">¥ {{ formatMoney(value) }}</span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { formatMoney } from '@/utils/format'

const props = withDefaults(
  defineProps<{ value: number | string | null | undefined; tone?: 'normal' | 'income' | 'expense' }>(),
  { tone: 'normal' },
)

const colorClass = computed(() => {
  if (props.tone === 'income') return 'money-text--income'
  if (props.tone === 'expense') return 'money-text--expense'
  return ''
})
</script>

<style scoped>
.money-text {
  font-variant-numeric: tabular-nums;
  font-weight: 600;
}
.money-text--income {
  color: #67c23a;
}
.money-text--expense {
  color: #f56c6c;
}
</style>
