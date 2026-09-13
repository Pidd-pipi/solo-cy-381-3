// 结算建议模块 API
import { get, post } from '@/utils/request'

export interface SettlementInfo {
  id: number
  group_id: number
  from_user_id: number
  from_name: string
  to_user_id: number
  to_name: string
  amount: number
  status: string
  settled_at?: string
  created_at: string
}

export interface GroupBalance {
  user_id: number
  username: string
  nickname: string
  net_amount: number
}

export function generateSettlementsApi(groupId: number) {
  return post<{ list: SettlementInfo[]; total: number }>(`/groups/${groupId}/settlements/generate`)
}

export function listSettlementsApi(groupId: number) {
  return get<{ list: SettlementInfo[]; total: number }>(`/groups/${groupId}/settlements`)
}

export function listPendingSettlementsApi() {
  return get<{ list: SettlementInfo[]; total: number }>('/settlements/pending')
}

export function settleApi(settlementIds: number[]) {
  return post<{ message: string; affected: number }>('/settlements/settle', { settlement_ids: settlementIds })
}

export function listBalancesApi(groupId: number) {
  return get<{ list: GroupBalance[]; total: number }>(`/groups/${groupId}/balances`)
}
