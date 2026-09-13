<template>
  <el-container class="main-layout">
    <el-aside width="220px" class="main-layout__aside">
      <div class="main-layout__logo">
        <el-icon :size="22" color="#409eff"><Money /></el-icon>
        <span>AA 分账</span>
      </div>
      <el-menu :default-active="activeMenu" router class="main-layout__menu">
        <el-menu-item index="/">
          <el-icon><HomeFilled /></el-icon>
          <span>工作台</span>
        </el-menu-item>
        <el-menu-item v-if="authStore.isAdmin" index="/audit-logs">
          <el-icon><Document /></el-icon>
          <span>审计日志</span>
        </el-menu-item>
        <el-menu-item index="/profile">
          <el-icon><User /></el-icon>
          <span>个人中心</span>
        </el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header class="main-layout__header">
        <div class="main-layout__header-left">
          <el-badge :value="pendingCount" :hidden="pendingCount === 0" :max="99">
            <el-button text @click="goPending">待结算提醒</el-button>
          </el-badge>
        </div>
        <el-dropdown @command="handleCommand">
          <span class="main-layout__user">
            <el-avatar :size="30" :src="authStore.user?.avatar || ''">{{ (authStore.user?.nickname || authStore.user?.username || '?').slice(0, 1) }}</el-avatar>
            <span>{{ authStore.user?.nickname || authStore.user?.username }}</span>
            <el-icon><ArrowDown /></el-icon>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="profile">个人中心</el-dropdown-item>
              <el-dropdown-item command="logout" divided>退出登录</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </el-header>
      <el-main class="main-layout__main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowDown, Document, HomeFilled, Money, User } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import { useSettlementStore } from '@/stores/settlement'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const settlementStore = useSettlementStore()

const activeMenu = computed(() => {
  const gid = route.params.id
  if (gid) {
    return route.path
  }
  return route.path
})

const pendingCount = computed(() => settlementStore.pending.length)

async function loadPending() {
  if (authStore.isLoggedIn) {
    await settlementStore.fetchPending().catch(() => {})
  }
}

watch(
  () => authStore.isLoggedIn,
  (v) => {
    if (v) loadPending()
  },
)

onMounted(() => {
  loadPending()
})

function goPending() {
  router.push('/')
}

function handleCommand(cmd: string) {
  if (cmd === 'logout') {
    authStore.logout()
    router.push('/login')
  } else if (cmd === 'profile') {
    router.push('/profile')
  }
}
</script>

<style scoped>
.main-layout {
  min-height: 100vh;
}
.main-layout__aside {
  background: #fff;
  border-right: 1px solid #e4e7ed;
}
.main-layout__logo {
  display: flex;
  align-items: center;
  gap: 8px;
  height: 60px;
  padding: 0 20px;
  font-size: 18px;
  font-weight: 700;
  color: #303133;
}
.main-layout__menu {
  border-right: none;
}
.main-layout__header {
  background: #fff;
  border-bottom: 1px solid #e4e7ed;
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.main-layout__user {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  color: #303133;
}
.main-layout__main {
  padding: 24px;
}
</style>
