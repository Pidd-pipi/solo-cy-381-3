// useAuth：路由守卫使用的认证组合式函数
import { computed } from 'vue'
import { useAuthStore } from '@/stores/auth'

export function useAuth() {
  const authStore = useAuthStore()
  const isLoggedIn = computed(() => authStore.isLoggedIn)
  const isAdmin = computed(() => authStore.isAdmin)
  const user = computed(() => authStore.user)
  return { isLoggedIn, isAdmin, user, login: authStore.login, logout: authStore.logout }
}
