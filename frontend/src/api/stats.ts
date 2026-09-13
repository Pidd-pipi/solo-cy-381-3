// 数据统计模块 API
import { get } from '@/utils/request'

export interface CategoryStat {
  category: string
  amount: number
  count: number
}

export interface MonthlyStat {
  month: string
  amount: number
  count: number
}

export interface MemberRank {
  user_id: number
  username: string
  nickname: string
  paid: number
  owed: number
  net: number
}

export interface GroupStats {
  category_stats: CategoryStat[]
  monthly_stats: MonthlyStat[]
  member_ranks: MemberRank[]
  total_expense: number
  total_count: number
}

export function getGroupStatsApi(groupId: number) {
  return get<GroupStats>(`/groups/${groupId}/stats`)
}
