<template>
  <div class="dashboard">
    <el-row :gutter="16">
      <el-col :span="16">
        <el-card shadow="never">
          <template #header>
            <div class="card-header">
              <span>我的群组</span>
              <el-button type="primary" size="small" @click="createVisible = true">创建群组</el-button>
            </div>
          </template>
          <DataTable :data="groupStore.groups" :loading="groupStore.loading">
            <el-table-column label="群组名称" min-width="160">
              <template #default="{ row }">
                <el-link type="primary" @click="goGroup(row.id)">{{ row.name }}</el-link>
              </template>
            </el-table-column>
            <el-table-column prop="description" label="描述" min-width="160" show-overflow-tooltip />
            <el-table-column label="成员" width="80">
              <template #default="{ row }">{{ row.member_count }} 人</template>
            </el-table-column>
            <el-table-column label="状态" width="90">
              <template #default="{ row }">
                <StatusBadge :status="row.status" kind="group" />
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="创建时间" width="160" />
            <el-table-column label="操作" width="120">
              <template #default="{ row }">
                <el-button size="small" type="primary" text @click="goGroup(row.id)">进入</el-button>
              </template>
            </el-table-column>
          </DataTable>
          <EmptyState v-if="!groupStore.loading && groupStore.groups.length === 0" description="还没有群组，点击右上角创建" />
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="never">
          <template #header><span>待结算提醒</span></template>
          <div v-if="settlementStore.pending.length === 0" class="dashboard__ok">
            <el-icon color="#67c23a" :size="28"><CircleCheck /></el-icon>
            <span>当前没有待结算项</span>
          </div>
          <div v-for="item in settlementStore.pending" :key="item.id" class="dashboard__pending-item">
            <div class="dashboard__pending-main">
              <span class="dashboard__pending-amount"><MoneyText :value="item.amount" :tone="item.from_user_id === myId ? 'expense' : 'income'" /></span>
              <span class="dashboard__pending-desc">
                {{ item.from_user_id === myId ? '你应转给' : item.to_name + ' 应转给你' }}
              </span>
            </div>
            <el-button size="small" type="primary" text @click="goSettle(item.group_id)">去结算</el-button>
          </div>
        </el-card>
        <el-card shadow="never" style="margin-top: 16px">
          <template #header><span>快捷入口</span></template>
          <el-button v-for="c in CategoryOptions" :key="c.value" class="dashboard__quick" size="small" @click="quickAdd(c.value)">
            {{ c.label }}
          </el-button>
        </el-card>
      </el-col>
    </el-row>

    <el-dialog v-model="createVisible" title="创建群组" width="460px">
      <el-form ref="createFormRef" :model="createForm" :rules="createRules" label-width="80px">
        <el-form-item label="群组名称" prop="name">
          <el-input v-model="createForm.name" placeholder="例如：周末聚餐小分队" maxlength="128" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="createForm.description" type="textarea" :rows="3" placeholder="群组用途说明（可选）" maxlength="512" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">取消</el-button>
        <el-button type="primary" :loading="creating" @click="handleCreate">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { CircleCheck } from '@element-plus/icons-vue'
import { CategoryOptions } from '@/constants'
import { useGroupStore } from '@/stores/group'
import { useSettlementStore } from '@/stores/settlement'
import { useAuthStore } from '@/stores/auth'
import DataTable from '@/components/DataTable.vue'
import EmptyState from '@/components/EmptyState.vue'
import StatusBadge from '@/components/StatusBadge.vue'
import MoneyText from '@/components/MoneyText.vue'

const router = useRouter()
const groupStore = useGroupStore()
const settlementStore = useSettlementStore()
const authStore = useAuthStore()

const myId = computed(() => authStore.user?.id || 0)
const createVisible = ref(false)
const creating = ref(false)
const createFormRef = ref<FormInstance>()
const createForm = reactive({ name: '', description: '' })
const createRules: FormRules = {
  name: [{ required: true, min: 1, max: 128, message: '请输入群组名称', trigger: 'blur' }],
}

onMounted(async () => {
  await groupStore.fetchGroups()
  await settlementStore.fetchPending()
})

function goGroup(id: number) {
  router.push(`/groups/${id}`)
}

function goSettle(groupId: number) {
  router.push(`/groups/${groupId}/settlements`)
}

function quickAdd(category: string) {
  const group = groupStore.groups[0]
  if (!group) {
    ElMessage.warning('请先创建群组')
    return
  }
  router.push({ path: `/groups/${group.id}/expenses`, query: { category } })
}

async function handleCreate() {
  if (!createFormRef.value) return
  const valid = await createFormRef.value.validate().catch(() => false)
  if (!valid) return
  creating.value = true
  try {
    const group = await groupStore.createGroup(createForm)
    ElMessage.success('群组创建成功')
    createVisible.value = false
    createForm.name = ''
    createForm.description = ''
    router.push(`/groups/${group.id}`)
  } finally {
    creating.value = false
  }
}
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.dashboard__ok {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #909399;
  padding: 12px 0;
}
.dashboard__pending-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 0;
  border-bottom: 1px dashed #ebeef5;
}
.dashboard__pending-item:last-child {
  border-bottom: none;
}
.dashboard__pending-main {
  display: flex;
  flex-direction: column;
}
.dashboard__pending-desc {
  color: #909399;
  font-size: 12px;
}
.dashboard__quick {
  margin: 4px 8px 4px 0;
}
</style>
