<template>
  <div class="group-members">
    <el-page-header :content="'成员管理 - ' + (groupStore.current?.name || '')" @back="router.back()" />
    <el-card shadow="never" style="margin-top: 16px">
      <template #header>
        <div class="card-header">
          <span>群组成员（{{ groupStore.members.length }}）</span>
          <el-button v-if="isManager" type="primary" size="small" @click="inviteVisible = true">邀请成员</el-button>
        </div>
      </template>
      <DataTable :data="groupStore.members" :loading="loading">
        <el-table-column label="成员" min-width="180">
          <template #default="{ row }">
            <div class="member-cell">
              <el-avatar :size="30" :src="row.avatar || ''">{{ row.nickname.slice(0, 1) }}</el-avatar>
              <div>
                <div>{{ row.nickname }}</div>
                <div class="member-cell__username">@{{ row.username }}</div>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="角色" width="110">
          <template #default="{ row }">
            <StatusBadge :status="row.role === 'owner' ? 'active' : row.role" kind="group" />
            <span style="margin-left: 6px">{{ row.role === 'owner' ? '群主' : row.role === 'admin' ? '管理员' : '成员' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="joined_at" label="加入时间" width="170" />
        <el-table-column label="操作" width="110">
          <template #default="{ row }">
            <el-button v-if="isManager && row.role !== 'owner' && row.user_id !== myId" size="small" type="danger" text @click="handleRemove(row)">移除</el-button>
          </template>
        </el-table-column>
      </DataTable>
      <EmptyState v-if="!loading && groupStore.members.length === 0" description="暂无成员" />
    </el-card>

    <el-dialog v-model="inviteVisible" title="邀请成员" width="420px">
      <el-form label-width="80px" @submit.prevent>
        <el-form-item label="用户名">
          <el-input v-model="inviteUsername" placeholder="输入对方用户名" @keyup.enter="handleInvite" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="inviteVisible = false">取消</el-button>
        <el-button type="primary" :loading="inviting" @click="handleInvite">邀请</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useGroupStore } from '@/stores/group'
import { useAuthStore } from '@/stores/auth'
import DataTable from '@/components/DataTable.vue'
import EmptyState from '@/components/EmptyState.vue'
import StatusBadge from '@/components/StatusBadge.vue'

const route = useRoute()
const router = useRouter()
const groupStore = useGroupStore()
const authStore = useAuthStore()
const groupId = Number(route.params.id)

const loading = ref(false)
const inviteVisible = ref(false)
const inviteUsername = ref('')
const inviting = ref(false)
const myId = computed(() => authStore.user?.id || 0)

const isManager = computed(() => {
  const g = groupStore.current
  if (!g) return false
  return g.owner_id === authStore.user?.id || authStore.isAdmin
})

onMounted(async () => {
  loading.value = true
  try {
    if (!groupStore.current) {
      await groupStore.fetchGroup(groupId)
    }
    await groupStore.fetchMembers(groupId)
  } finally {
    loading.value = false
  }
})

async function handleInvite() {
  if (!inviteUsername.value.trim()) {
    ElMessage.warning('请输入用户名')
    return
  }
  inviting.value = true
  try {
    await groupStore.inviteMember(groupId, inviteUsername.value.trim())
    ElMessage.success('邀请成功')
    inviteVisible.value = false
    inviteUsername.value = ''
  } finally {
    inviting.value = false
  }
}

async function handleRemove(row: { user_id: number; nickname: string }) {
  try {
    await groupStore.removeMember(groupId, row.user_id)
    ElMessage.success('成员已移除')
  } catch {
    // 错误提示由拦截器统一处理
  }
}
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.member-cell {
  display: flex;
  align-items: center;
  gap: 10px;
}
.member-cell__username {
  color: #909399;
  font-size: 12px;
}
</style>
