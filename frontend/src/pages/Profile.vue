<template>
  <div class="profile">
    <el-row :gutter="16">
      <el-col :span="10">
        <el-card shadow="never">
          <template #header><span>个人资料</span></template>
          <el-form ref="profileFormRef" :model="profileForm" :rules="profileRules" label-width="80px">
            <el-form-item label="用户名">
              <el-input :model-value="authStore.user?.username" disabled />
            </el-form-item>
            <el-form-item label="昵称" prop="nickname">
              <el-input v-model="profileForm.nickname" maxlength="64" />
            </el-form-item>
            <el-form-item label="邮箱" prop="email">
              <el-input v-model="profileForm.email" maxlength="128" />
            </el-form-item>
            <el-form-item label="头像 URL" prop="avatar">
              <el-input v-model="profileForm.avatar" maxlength="512" placeholder="头像图片地址（可选）" />
            </el-form-item>
            <el-form-item label="角色">
              <StatusBadge :status="authStore.user?.role || ''" kind="group" />
              <span style="margin-left: 8px">{{ roleText(authStore.user?.role || '') }}</span>
            </el-form-item>
            <el-button type="primary" :loading="saving" @click="handleSave">保存修改</el-button>
          </el-form>
        </el-card>
      </el-col>
      <el-col :span="10">
        <el-card shadow="never">
          <template #header><span>修改密码</span></template>
          <el-form ref="pwdFormRef" :model="pwdForm" :rules="pwdRules" label-width="90px">
            <el-form-item label="原密码" prop="old_password">
              <el-input v-model="pwdForm.old_password" type="password" show-password />
            </el-form-item>
            <el-form-item label="新密码" prop="new_password">
              <el-input v-model="pwdForm.new_password" type="password" show-password />
            </el-form-item>
            <el-button type="warning" :loading="savingPwd" @click="handleChangePwd">修改密码</el-button>
          </el-form>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { useAuthStore } from '@/stores/auth'
import { changePasswordApi } from '@/api/user'
import StatusBadge from '@/components/StatusBadge.vue'
import { roleText } from '@/utils/format'

const authStore = useAuthStore()
const profileFormRef = ref<FormInstance>()
const pwdFormRef = ref<FormInstance>()
const saving = ref(false)
const savingPwd = ref(false)

const profileForm = reactive({ nickname: '', email: '', avatar: '' })
const profileRules: FormRules = {
  nickname: [{ required: true, message: '请输入昵称', trigger: 'blur' }],
  email: [{ type: 'email', message: '邮箱格式不正确', trigger: 'blur' }],
}

const pwdForm = reactive({ old_password: '', new_password: '' })
const pwdRules: FormRules = {
  old_password: [{ required: true, min: 6, message: '请输入原密码', trigger: 'blur' }],
  new_password: [{ required: true, min: 6, message: '新密码至少 6 位', trigger: 'blur' }],
}

onMounted(() => {
  profileForm.nickname = authStore.user?.nickname || ''
  profileForm.email = authStore.user?.email || ''
  profileForm.avatar = authStore.user?.avatar || ''
})

async function handleSave() {
  if (!profileFormRef.value) return
  const valid = await profileFormRef.value.validate().catch(() => false)
  if (!valid) return
  saving.value = true
  try {
    await authStore.updateProfile(profileForm)
    ElMessage.success('资料已更新')
  } finally {
    saving.value = false
  }
}

async function handleChangePwd() {
  if (!pwdFormRef.value) return
  const valid = await pwdFormRef.value.validate().catch(() => false)
  if (!valid) return
  savingPwd.value = true
  try {
    await changePasswordApi(pwdForm)
    ElMessage.success('密码修改成功')
    pwdForm.old_password = ''
    pwdForm.new_password = ''
  } finally {
    savingPwd.value = false
  }
}
</script>
