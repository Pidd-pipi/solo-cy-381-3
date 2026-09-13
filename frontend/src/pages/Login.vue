<template>
  <div class="auth-page">
    <el-card class="auth-card">
      <div class="auth-card__title">
        <el-icon :size="30" color="#409eff"><Money /></el-icon>
        <h2>AA 分账</h2>
        <p>朋友聚餐、合租、旅行的智能费用分摊与结算</p>
      </div>
      <el-form ref="formRef" :model="form" :rules="rules" label-position="top" @keyup.enter="handleLogin">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="form.username" placeholder="请输入用户名" :prefix-icon="User" />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input v-model="form.password" type="password" show-password placeholder="请输入密码" :prefix-icon="Lock" />
        </el-form-item>
        <el-button type="primary" class="auth-card__submit" :loading="loading" @click="handleLogin">登 录</el-button>
        <div class="auth-card__footer">
          还没有账号？<router-link to="/register">立即注册</router-link>
        </div>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { Lock, Money, User } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const formRef = ref<FormInstance>()
const loading = ref(false)

const form = reactive({ username: '', password: '' })
const rules: FormRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
}

async function handleLogin() {
  if (!formRef.value) return
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return
  loading.value = true
  try {
    await authStore.login(form.username, form.password)
    ElMessage.success('登录成功')
    const redirect = (route.query.redirect as string) || '/'
    router.push(redirect)
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
  width: 400px;
  padding: 12px 8px;
}
.auth-card__title {
  text-align: center;
  margin-bottom: 24px;
}
.auth-card__title h2 {
  margin: 8px 0 4px;
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
