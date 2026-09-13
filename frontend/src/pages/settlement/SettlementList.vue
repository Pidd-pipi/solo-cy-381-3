<template>
  <div class="settlement-list">
    <el-page-header :content="'智能结算 - ' + (groupStore.current?.name || '')" @back="router.push(`/groups/${groupId}`)" />
    <el-card shadow="never" style="margin-top: 16px">
      <template #header>
        <div class="card-header">
          <span>结算建议（最小化转账次数）</span>
          <el-button type="primary" :loading="generating" @click="handleGenerate">生成结算建议</el-button>
        </div>
      </template>
      <DataTable :data="settlementStore.settlements" :loading="loading">
        <el-table-column label="转账方向" min-width="200">
          <template #default="{ row }">
            <span class="flow">
              <span class="flow__name">{{ row.from_name || '用户#' + row.from_user_id }}</span>
              <el-icon color="#f56c6c"><Right /></el-icon>
              <span class="flow__name">{{ row.to_name || '用户#' + row.to_user_id }}</span>
            </span>
          </template>
        </el-table-column>
        <el-table-column label="金额" width="130">
          <template #default="{ row }"><MoneyText :value="row.amount" /></template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }"><StatusBadge :status="row.status" kind="settlement" /></template>
        </el-table-column>
        <el-table-column prop="created_at" label="生成时间" width="170" />
        <el-table-column label="操作" width="110">
          <template #default="{ row }">
            <el-button v-if="row.status === 'pending'" size="small" type="success" text @click="handleSettle([row])">标记已结算</el-button>
          </template>
        </el-table-column>
      </DataTable>
      <EmptyState v-if="!loading && settlementStore.settlements.length === 0" description="暂无结算建议，点击右上角生成">
        <el-button type="primary" size="small" @click="handleGenerate">立即生成</el-button>
      </EmptyState>
    </el-card>

    <el-card shadow="never" style="margin-top: 16px">
      <template #header><span>成员净余额</span></template>
      <el-table :data="settlementStore.balances" border stripe>
        <el-table-column prop="nickname" label="成员" min-width="140" />
        <el-table-column prop="username" label="用户名" min-width="120" />
        <el-table-column label="净余额" min-width="140">
          <template #default="{ row }">
            <MoneyText :value="row.net_amount" :tone="row.net_amount >= 0 ? 'income' : 'expense'" />
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Right } from '@element-plus/icons-vue'
import { useGroupStore } from '@/stores/group'
import { useSettlementStore } from '@/stores/settlement'
import type { SettlementInfo } from '@/api/settlement'
import DataTable from '@/components/DataTable.vue'
import EmptyState from '@/components/EmptyState.vue'
import StatusBadge from '@/components/StatusBadge.vue'
import MoneyText from '@/components/MoneyText.vue'

const route = useRoute()
const router = useRouter()
const groupStore = useGroupStore()
const settlementStore = useSettlementStore()
const groupId = Number(route.params.id)

const loading = ref(false)
const generating = ref(false)

onMounted(async () => {
  loading.value = true
  try {
    if (!groupStore.current) {
      await groupStore.fetchGroup(groupId)
    }
    await settlementStore.fetchSettlements(groupId)
    await settlementStore.fetchBalances(groupId)
  } finally {
    loading.value = false
  }
})

async function handleGenerate() {
  generating.value = true
  try {
    await settlementStore.generate(groupId)
    await settlementStore.fetchBalances(groupId)
    ElMessage.success('结算建议已生成')
  } finally {
    generating.value = false
  }
}

async function handleSettle(items: SettlementInfo[]) {
  await settlementStore.settle(items.map((i) => i.id))
  await settlementStore.fetchSettlements(groupId)
  await settlementStore.fetchBalances(groupId)
  ElMessage.success('结算完成')
}
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.flow {
  display: flex;
  align-items: center;
  gap: 8px;
}
.flow__name {
  font-weight: 600;
}
</style>
