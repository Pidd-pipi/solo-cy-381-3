// 消费记录模块 API
import { del, get, post, put } from '@/utils/request'
import type { PageData } from '@/utils/request'

export interface ShareInput {
  user_id: number
  ratio?: number
  amount?: number
}

export interface ExpenseShare {
  user_id: number
  username: string
  nickname: string
  share_amount: number
  ratio: number
  status: string
}

export interface ExpenseInfo {
  id: number
  group_id: number
  title: string
  amount: number
  category: string
  payer_id: number
  payer_name: string
  split_type: string
  paid_at: string
  receipt_url: string
  status: string
  created_by: number
  created_at: string
  shares: ExpenseShare[]
}

export interface ExpensePayload {
  group_id?: number
  title: string
  amount: number
  category: string
  payer_id: number
  split_type: string
  paid_at: string
  receipt_url?: string
  shares: ShareInput[]
}

export interface ExpenseQuery {
  page?: number
  page_size?: number
  category?: string
  start?: string
  end?: string
}

export function listExpensesApi(groupId: number, params: ExpenseQuery) {
  return get<PageData<ExpenseInfo>>(`/groups/${groupId}/expenses`, params)
}

export function createExpenseApi(payload: ExpensePayload) {
  return post<ExpenseInfo>('/groups/' + payload.group_id + '/expenses', payload)
}

export function getExpenseApi(id: number) {
  return get<ExpenseInfo>(`/expenses/${id}`)
}

export function updateExpenseApi(id: number, payload: ExpensePayload) {
  return put<{ message: string }>(`/expenses/${id}`, payload)
}

export function refundExpenseApi(id: number) {
  return del<{ message: string }>(`/expenses/${id}`)
}

export function exportExpensesApi(groupId: number, params: ExpenseQuery) {
  const query = new URLSearchParams()
  if (params.category) query.set('category', params.category)
  if (params.start) query.set('start', params.start)
  if (params.end) query.set('end', params.end)
  return `/groups/${groupId}/expenses/export?${query.toString()}`
}
