<template>
  <div class="expense-list">
    <el-page-header :content="'消费记录 - ' + (groupStore.current?.name || '')" @back="router.push(`/groups/${groupId}`)" />
    <el-card shadow="never" style="margin-top: 16px">
      <div class="filter-bar">
        <el-select v-model="filters.category" placeholder="消费类别" clearable style="width: 140px" @change="handleFilter">
          <el-option v-for="opt in CategoryOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
        </el-select>
        <el-date-picker
          v-model="dateRange"
          type="daterange"
          value-format="YYYY-MM-DD"
          range-separator="至"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
          style="width: 260px"
          @change="handleFilter"
        />
        <el-button type="primary" @click="openCreate">添加消费</el-button>
        <el-button @click="handleExport">导出 CSV</el-button>
      </div>
      <DataTable :data="expenseStore.expenses" :loading="expenseStore.loading" :total="expenseStore.total" show-pagination @page-change="handlePage">
        <el-table-column prop="title" label="标题" min-width="130" show-overflow-tooltip />
        <el-table-column label="金额" width="110">
          <template #default="{ row }"><MoneyText :value="row.amount" /></template>
        </el-table-column>
        <el-table-column label="类别" width="90">
          <template #default="{ row }"><StatusBadge :status="row.category" kind="category" /></template>
        </el-table-column>
        <el-table-column prop="payer_name" label="付款人" width="100" />
        <el-table-column label="分摊方式" width="100">
          <template #default="{ row }"><SplitTypeTag :type="row.split_type" /></template>
        </el-table-column>
        <el-table-column prop="paid_at" label="消费时间" width="160" />
        <el-table-column label="状态" width="90">
          <template #default="{ row }"><StatusBadge :status="row.status" kind="expense" /></template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button size="small" text type="primary" @click="openDetail(row)">详情</el-button>
            <el-button v-if="row.status === 'active'" size="small" text type="danger" @click="handleRefund(row)">退款</el-button>
          </template>
        </el-table-column>
      </DataTable>
      <EmptyState v-if="!expenseStore.loading && expenseStore.expenses.length === 0" description="还没有消费记录，点击「添加消费」开始记账" />
    </el-card>

    <el-dialog v-model="detailVisible" title="消费记录详情" width="620px">
      <template v-if="currentExpense">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="标题">{{ currentExpense.title }}</el-descriptions-item>
          <el-descriptions-item label="金额"><MoneyText :value="currentExpense.amount" /></el-descriptions-item>
          <el-descriptions-item label="类别"><StatusBadge :status="currentExpense.category" kind="category" /></el-descriptions-item>
          <el-descriptions-item label="付款人">{{ currentExpense.payer_name }}</el-descriptions-item>
          <el-descriptions-item label="分摊方式"><SplitTypeTag :type="currentExpense.split_type" /></el-descriptions-item>
          <el-descriptions-item label="消费时间">{{ currentExpense.paid_at }}</el-descriptions-item>
          <el-descriptions-item label="小票" :span="2">
            <el-link v-if="currentExpense.receipt_url" :href="currentExpense.receipt_url" target="_blank">查看小票</el-link>
            <span v-else>-</span>
          </el-descriptions-item>
        </el-descriptions>
        <h4 style="margin: 16px 0 8px">分摊明细</h4>
        <el-table :data="currentExpense.shares" size="small" border>
          <el-table-column prop="nickname" label="成员" />
          <el-table-column label="应付金额">
            <template #default="{ row }"><MoneyText :value="row.share_amount" /></template>
          </el-table-column>
          <el-table-column label="占比">
            <template #default="{ row }">{{ (row.ratio * 100).toFixed(1) }}%</template>
          </el-table-column>
        </el-table>
        <div style="margin-top: 16px; text-align: right">
          <el-button v-if="currentExpense.status === 'active'" type="primary" @click="openEdit(currentExpense)">编辑</el-button>
        </div>
      </template>
    </el-dialog>

    <ExpenseFormDialog ref="formDialog" :members="groupStore.members" :expense="editingExpense" @saved="handleSaved" />
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { CategoryOptions } from '@/constants'
import { useGroupStore } from '@/stores/group'
import { useExpenseStore } from '@/stores/expense'
import type { ExpenseInfo, ExpensePayload } from '@/api/expense'
import { exportExpensesApi } from '@/api/expense'
import DataTable from '@/components/DataTable.vue'
import EmptyState from '@/components/EmptyState.vue'
import StatusBadge from '@/components/StatusBadge.vue'
import SplitTypeTag from '@/components/SplitTypeTag.vue'
import MoneyText from '@/components/MoneyText.vue'
import ExpenseFormDialog from '@/components/ExpenseFormDialog.vue'

const route = useRoute()
const router = useRouter()
const groupStore = useGroupStore()
const expenseStore = useExpenseStore()
const groupId = Number(route.params.id)

const filters = reactive<{ category?: string; start?: string; end?: string }>({
  category: (route.query.category as string) || undefined,
})
const dateRange = ref<[string, string] | null>(null)
const page = ref(1)
const detailVisible = ref(false)
const currentExpense = ref<ExpenseInfo | null>(null)
const editingExpense = ref<ExpenseInfo | null>(null)
const formDialog = ref<InstanceType<typeof ExpenseFormDialog>>()

onMounted(async () => {
  if (!groupStore.current) {
    await groupStore.fetchGroup(groupId)
  }
  await groupStore.fetchMembers(groupId)
  await loadExpenses()
})

watch(
  () => route.query.category,
  (v) => {
    filters.category = (v as string) || undefined
    if (v) loadExpenses()
  },
)

async function loadExpenses() {
  await expenseStore.fetchExpenses(groupId, { ...filters, page: page.value })
}

function handleFilter() {
  if (dateRange.value) {
    filters.start = dateRange.value[0]
    filters.end = dateRange.value[1]
  } else {
    filters.start = undefined
    filters.end = undefined
  }
  page.value = 1
  loadExpenses()
}

function handlePage(p: number) {
  page.value = p
  loadExpenses()
}

function openCreate() {
  editingExpense.value = null
  formDialog.value?.open()
}

function openEdit(expense: ExpenseInfo) {
  editingExpense.value = expense
  formDialog.value?.open(expense)
}

function openDetail(row: ExpenseInfo) {
  currentExpense.value = row
  detailVisible.value = true
}

async function handleSaved(payload: ExpensePayload) {
  if (editingExpense.value) {
    await expenseStore.updateExpense(editingExpense.value.id, { ...payload, group_id: groupId })
    ElMessage.success('消费记录已更新')
  } else {
    await expenseStore.createExpense({ ...payload, group_id: groupId })
    ElMessage.success('消费记录已创建')
  }
  await loadExpenses()
}

async function handleRefund(row: ExpenseInfo) {
  await expenseStore.refundExpense(row.id)
  ElMessage.success('已标记退款')
  await loadExpenses()
}

function handleExport() {
  const url = exportExpensesApi(groupId, { ...filters })
  const token = localStorage.getItem('aasplit_token') || ''
  window.open(url, '_blank')
}
</script>

<style scoped>
.filter-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}
</style>
