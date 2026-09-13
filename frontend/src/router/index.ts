// 路由与守卫（JWT 认证 + 管理员 RBAC，横切关注点 1 前端触达层）
import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes: RouteRecordRaw[] = [
  { path: '/login', name: 'Login', component: () => import('@/pages/Login.vue'), meta: { public: true } },
  { path: '/register', name: 'Register', component: () => import('@/pages/Register.vue'), meta: { public: true } },
  {
    path: '/',
    component: () => import('@/layouts/MainLayout.vue'),
    children: [
      { path: '', name: 'Dashboard', component: () => import('@/pages/Dashboard.vue') },
      { path: 'groups/:id', name: 'GroupDetail', component: () => import('@/pages/group/GroupDetail.vue') },
      { path: 'groups/:id/members', name: 'GroupMembers', component: () => import('@/pages/group/GroupMembers.vue') },
      { path: 'groups/:id/expenses', name: 'ExpenseList', component: () => import('@/pages/expense/ExpenseList.vue') },
      { path: 'groups/:id/settlements', name: 'SettlementList', component: () => import('@/pages/settlement/SettlementList.vue') },
      { path: 'groups/:id/stats', name: 'GroupStats', component: () => import('@/pages/stats/GroupStats.vue') },
      { path: 'audit-logs', name: 'AuditLogs', component: () => import('@/pages/audit/AuditLogs.vue'), meta: { admin: true } },
      { path: 'profile', name: 'Profile', component: () => import('@/pages/Profile.vue') },
    ],
  },
  { path: '/:pathMatch(.*)*', redirect: '/' },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (!to.meta.public && !auth.isLoggedIn) {
    return { path: '/login', query: { redirect: to.fullPath } }
  }
  if (to.meta.admin && !auth.isAdmin) {
    return { path: '/', replace: true }
  }
  if (to.meta.public && auth.isLoggedIn && (to.name === 'Login' || to.name === 'Register')) {
    return { path: '/', replace: true }
  }
  return true
})

export default router
