<!-- StatusBadge：通用状态徽标（状态枚举 → 颜色/文案，前后端枚举一致性） -->
<template>
  <el-tag :type="tagType" size="small" effect="light">{{ text }}</el-tag>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import {
  groupStatusText,
  settlementStatusText,
  expenseStatusText,
  splitTypeText,
  categoryText,
} from '@/utils/format'
import { GroupStatus, SettlementStatus, ExpenseStatus, SplitType } from '@/constants'

const props = defineProps<{
  status: string
  kind?: 'group' | 'settlement' | 'expense' | 'split' | 'category'
}>()

const tagType = computed<'primary' | 'success' | 'warning' | 'info' | 'danger'>(() => {
  const s = props.status
  switch (props.kind) {
    case 'group':
      return s === GroupStatus.ACTIVE ? 'success' : 'info'
    case 'settlement':
      return s === SettlementStatus.PENDING ? 'warning' : 'success'
    case 'expense':
      return s === ExpenseStatus.ACTIVE ? 'success' : 'danger'
    case 'split':
      return s === SplitType.EQUAL ? 'primary' : s === SplitType.RATIO ? 'warning' : 'success'
    default:
      return 'info'
  }
})

const text = computed(() => {
  switch (props.kind) {
    case 'group':
      return groupStatusText(props.status)
    case 'settlement':
      return settlementStatusText(props.status)
    case 'expense':
      return expenseStatusText(props.status)
    case 'split':
      return splitTypeText(props.status)
    case 'category':
      return categoryText(props.status)
    default:
      return props.status
  }
})
</script>
