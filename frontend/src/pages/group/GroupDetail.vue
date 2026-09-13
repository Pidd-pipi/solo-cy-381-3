<template>
  <div class="group-detail">
    <el-page-header :content="groupStore.current?.name || '群组详情'" @back="router.push('/')" />
    <el-card v-if="groupStore.current" shadow="never" class="group-detail__card">
      <div class="group-detail__info">
        <div>
          <h3>{{ groupStore.current.name }}</h3>
          <p class="group-detail__desc">{{ groupStore.current.description || '暂无描述' }}</p>
          <div class="group-detail__meta">
            <StatusBadge :status="groupStore.current.status" kind="group" />
            <span>{{ groupStore.current.member_count }} 位成员</span>
            <span>创建于 {{ groupStore.current.created_at }}</span>
          </div>
        </div>
        <div v-if="isManager" class="group-detail__actions">
          <el-button size="small" @click="editVisible = true">编辑</el-button>
          <el-button size="small" type="danger" plain @click="confirmArchive && confirmArchive.open()">归档</el-button>
        </div>
      </div>
    </el-card>
    <el-row :gutter="16" class="group-detail__nav">
      <el-col :span="6">
        <el-card shadow="never" class="nav-card" @click="router.push(`/groups/${groupId}/expenses`)">
          <el-icon :size="26" color="#409eff"><List /></el-icon>
          <span>消费记录</span>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never" class="nav-card" @click="router.push(`/groups/${groupId}/settlements`)">
          <el-icon :size="26" color="#67c23a"><Money /></el-icon>
          <span>智能结算</span>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never" class="nav-card" @click="router.push(`/groups/${groupId}/stats`)">
          <el-icon :size="26" color="#e6a23c"><TrendCharts /></el-icon>
          <span>数据统计</span>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never" class="nav-card" @click="router.push(`/groups/${groupId}/members`)">
          <el-icon :size="26" color="#909399"><UserFilled /></el-icon>
          <span>成员管理</span>
        </el-card>
      </el-col>
    </el-row>

    <el-dialog v-model="editVisible" title="编辑群组" width="460px">
      <el-form ref="editFormRef" :model="editForm" :rules="editRules" label-width="80px">
        <el-form-item label="群组名称" prop="name">
          <el-input v-model="editForm.name" maxlength="128" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="editForm.description" type="textarea" :rows="3" maxlength="512" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>

    <ConfirmDialog ref="confirmArchive" title="归档群组" message="归档后不可继续记账，确定归档该群组吗？" @confirm="handleArchive" />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { List, Money, TrendCharts, UserFilled } from '@element-plus/icons-vue'
import { useGroupStore } from '@/stores/group'
import { useAuthStore } from '@/stores/auth'
import StatusBadge from '@/components/StatusBadge.vue'
import ConfirmDialog from '@/components/ConfirmDialog.vue'

const route = useRoute()
const router = useRouter()
const groupStore = useGroupStore()
const authStore = useAuthStore()
const groupId = Number(route.params.id)

const editVisible = ref(false)
const saving = ref(false)
const editFormRef = ref<FormInstance>()
const confirmArchive = ref<InstanceType<typeof ConfirmDialog>>()
const editForm = reactive({ name: '', description: '' })
const editRules: FormRules = {
  name: [{ required: true, min: 1, max: 128, message: '请输入群组名称', trigger: 'blur' }],
}

const isManager = computed(() => {
  const g = groupStore.current
  if (!g) return false
  return g.owner_id === authStore.user?.id || authStore.isAdmin
})

onMounted(async () => {
  await groupStore.fetchGroup(groupId)
  editForm.name = groupStore.current?.name || ''
  editForm.description = groupStore.current?.description || ''
})

async function handleSave() {
  if (!editFormRef.value) return
  const valid = await editFormRef.value.validate().catch(() => false)
  if (!valid) return
  saving.value = true
  try {
    await groupStore.updateGroup(groupId, editForm)
    ElMessage.success('群组已更新')
    editVisible.value = false
  } finally {
    saving.value = false
  }
}

async function handleArchive() {
  await groupStore.archiveGroup(groupId)
  ElMessage.success('群组已归档')
  router.push('/')
}
</script>

<style scoped>
.group-detail__card {
  margin-top: 16px;
}
.group-detail__info {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}
.group-detail__info h3 {
  margin: 0 0 8px;
}
.group-detail__desc {
  color: #909399;
  margin: 0 0 8px;
}
.group-detail__meta {
  display: flex;
  align-items: center;
  gap: 16px;
  color: #909399;
  font-size: 13px;
}
.group-detail__nav {
  margin-top: 16px;
}
.nav-card {
  cursor: pointer;
  text-align: center;
  transition: transform 0.2s;
}
.nav-card:hover {
  transform: translateY(-2px);
}
.nav-card .el-card__body {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}
</style>
