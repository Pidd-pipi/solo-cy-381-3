// 用户模块 API
import { get, put } from '@/utils/request'
import type { PageData } from '@/utils/request'

export interface UserInfo {
  id: number
  username: string
  nickname: string
  email: string
  avatar: string
  role: string
  status: string
  created_at: string
}

export interface UpdateProfilePayload {
  nickname: string
  email?: string
  avatar?: string
}

export function getMeApi() {
  return get<UserInfo>('/users/me')
}

export function updateProfileApi(payload: UpdateProfilePayload) {
  return put<UserInfo>('/users/me', payload)
}

export function changePasswordApi(payload: { old_password: string; new_password: string }) {
  return put<{ message: string }>('/users/me/password', payload)
}

export function listUsersApi(params: { page?: number; page_size?: number }) {
  return get<PageData<UserInfo>>('/users', params)
}

export function updateRoleApi(userId: number, role: string) {
  return put<{ message: string }>(`/users/${userId}/role`, { role })
}
