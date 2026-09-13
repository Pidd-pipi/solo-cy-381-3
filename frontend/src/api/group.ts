// 群组模块 API
import { del, get, post, put } from '@/utils/request'
import type { PageData } from '@/utils/request'

export interface GroupInfo {
  id: number
  name: string
  description: string
  owner_id: number
  status: string
  member_count: number
  created_at: string
}

export interface MemberInfo {
  id: number
  group_id: number
  user_id: number
  username: string
  nickname: string
  avatar: string
  role: string
  invited_by: number
  joined_at: string
}

export function listMyGroupsApi(params: { page?: number; page_size?: number }) {
  return get<PageData<GroupInfo>>('/groups', params)
}

export function createGroupApi(payload: { name: string; description?: string }) {
  return post<GroupInfo>('/groups', payload)
}

export function getGroupApi(id: number) {
  return get<GroupInfo>(`/groups/${id}`)
}

export function updateGroupApi(id: number, payload: { name: string; description?: string }) {
  return put<{ message: string }>(`/groups/${id}`, payload)
}

export function archiveGroupApi(id: number) {
  return post<{ message: string }>(`/groups/${id}/archive`)
}

export function inviteMemberApi(groupId: number, username: string) {
  return post<{ message: string }>(`/groups/${groupId}/members`, { username })
}

export function listMembersApi(groupId: number) {
  return get<{ list: MemberInfo[]; total: number }>(`/groups/${groupId}/members`)
}

export function removeMemberApi(groupId: number, userId: number) {
  return del<{ message: string }>(`/groups/${groupId}/members/${userId}`)
}
