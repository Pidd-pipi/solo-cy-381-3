<template>
  <div class="audit-logs">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>操作审计日志</span>
          <el-select v-model="actionFilter" placeholder="按动作筛选" clearable style="width: 180px" @change="loadLogs">
            <el-option v-for="opt in actionOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
          </el-select>
        </div>
      </template>
      <DataTable :data="auditStore.logs" :loading="auditStore.loading" :total="auditStore.total" show-pagination @page-change="handlePage">
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="username" label="操作人" width="120" />
        <el-table-column prop="action" label="动作" width="180" show-overflow-tooltip />
        <el-table-column prop="resource_type" label="资源类型" width="120" />
        <el-table-column prop="resource_id" label="资源 ID" width="110" />
        <el-table-column prop="detail" label="详情" min-width="220" show-overflow-tooltip />
        <el-table-column prop="ip" label="IP" width="130" />
        <el-table-column prop="created_at" label="时间" width="170" />
      </DataTable>
      <EmptyState v-if="!auditStore.loading && auditStore.logs.length === 0" description="暂无审计日志" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useAuditStore } from '@/stores/audit'
import DataTable from '@/components/DataTable.vue'
import EmptyState from '@/components/EmptyState.vue'

const auditStore = useAuditStore()
const actionFilter = ref('')
const page = ref(1)

const actionOptions = [
  { value: 'group.create', label: '创建群组' },
  { value: 'group.update', label: '更新群组' },
  { value: 'group.archive', label: '归档群组' },
  { value: 'member.invite', label: '邀请成员' },
  { value: 'member.remove', label: '移除成员' },
  { value: 'expense.create', label: '创建消费' },
  { value: 'expense.update', label: '更新消费' },
  { value: 'expense.delete', label: '退款消费' },
  { value: 'settlement.generate', label: '生成结算' },
  { value: 'settlement.settle', label: '完成结算' },
]

onMounted(() => loadLogs())

async function loadLogs() {
  await auditStore.fetchLogs({ page: page.value, action: actionFilter.value || undefined })
}

function handlePage(p: number) {
  page.value = p
  loadLogs()
}
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
