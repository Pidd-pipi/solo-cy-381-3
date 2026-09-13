// 认证状态管理（JWT + 用户信息）
import { defineStore } from 'pinia'
import { loginApi, registerApi, type LoginResult, type RegisterPayload } from '@/api/auth'
import { getMeApi, updateProfileApi, type UpdateProfilePayload, type UserInfo } from '@/api/user'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem('aasplit_token') || '',
    user: JSON.parse(localStorage.getItem('aasplit_user') || 'null') as UserInfo | null,
  }),
  getters: {
    isLoggedIn: (state) => !!state.token,
    isAdmin: (state) => state.user?.role === 'admin',
    nickname: (state) => state.user?.nickname || state.user?.username || '',
  },
  actions: {
    setAuth(result: LoginResult) {
      this.token = result.token
      this.user = result.user
      localStorage.setItem('aasplit_token', result.token)
      localStorage.setItem('aasplit_user', JSON.stringify(result.user))
    },
    async login(username: string, password: string) {
      const result = await loginApi({ username, password })
      this.setAuth(result)
    },
    async register(payload: RegisterPayload) {
      await registerApi(payload)
    },
    async fetchMe() {
      const user = await getMeApi()
      this.user = user
      localStorage.setItem('aasplit_user', JSON.stringify(user))
    },
    async updateProfile(payload: UpdateProfilePayload) {
      const user = await updateProfileApi(payload)
      this.user = user
      localStorage.setItem('aasplit_user', JSON.stringify(user))
    },
    logout() {
      this.token = ''
      this.user = null
      localStorage.removeItem('aasplit_token')
      localStorage.removeItem('aasplit_user')
    },
  },
})
