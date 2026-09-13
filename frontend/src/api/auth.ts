// auth 相关 API
import { get, post } from '@/utils/request'
import type { UserInfo } from '@/api/user'

export interface LoginResult {
  token: string
  user: UserInfo
}

export interface RegisterPayload {
  username: string
  password: string
  nickname: string
  email?: string
}

export function loginApi(payload: { username: string; password: string }) {
  return post<LoginResult>('/auth/login', payload)
}

export function registerApi(payload: RegisterPayload) {
  return post<UserInfo>('/auth/register', payload)
}

export function logoutApi() {
  return get<Record<string, unknown>>('/users/me')
}
