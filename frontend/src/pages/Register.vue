<template>
  <div class="auth-page">
    <el-card class="auth-card">
      <div class="auth-card__title">
        <h2>注册 AA 分账</h2>
        <p>创建账号，开始记录与分摊</p>
      </div>
      <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="form.username" placeholder="3-64 个字符" :prefix-icon="User" />
        </el-form-item>
        <el-form-item label="昵称" prop="nickname">
          <el-input v-model="form.nickname" placeholder="如何称呼你" />
        </el-form-item>
        <el-form-item label="邮箱（可选）" prop="email">
          <el-input v-model="form.email" placeholder="用于找回密码" />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input v-model="form.password" type="password" show-password placeholder="至少 6 位" :prefix-icon="Lock" />
        </el-form-item>
        <el-button type="primary" class="auth-card__submit" :loading="loading" @click="handleRegister">注 册</el-button>
        <div class="auth-card__footer">
          已有账号？<router-link to="/login">去登录</router-link>
        </div>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { Lock, User } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()
const formRef = ref<FormInstance>()
const loading = ref(false)

const form = reactive({ username: '', nickname: '', email: '', password: '' })
const rules: FormRules = {
  username: [{ required: true, min: 3, max: 64, message: '用户名需 3-64 个字符', trigger: 'blur' }],
  nickname: [{ required: true, message: '请输入昵称', trigger: 'blur' }],
  email: [{ type: 'email', message: '邮箱格式不正确', trigger: 'blur' }],
  password: [{ required: true, min: 6, max: 72, message: '密码至少 6 位', trigger: 'blur' }],
}

async function handleRegister() {
  if (!formRef.value) return
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return
  loading.value = true
  try {
    await authStore.register({
      username: form.username,
      nickname: form.nickname,
      email: form.email || undefined,
      password: form.password,
    })
    ElMessage.success('注册成功，请登录')
    router.push('/login')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.auth-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #e8f1ff 0%, #f5f7fa 100%);
}
.auth-card {
  width: 420px;
  padding: 12px 8px;
}
.auth-card__title {
  text-align: center;
  margin-bottom: 24px;
}
.auth-card__title p {
  color: #909399;
  font-size: 13px;
}
.auth-card__submit {
  width: 100%;
  margin-top: 8px;
}
.auth-card__footer {
  margin-top: 16px;
  text-align: center;
  color: #909399;
  font-size: 13px;
}
</style>
