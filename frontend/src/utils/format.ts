// format.ts 金额/日期/状态格式化工具
import dayjs from 'dayjs'
import {
  ExpenseCategory,
  GroupStatus,
  SettlementStatus,
  SplitType,
  ExpenseStatus,
} from '@/constants'

export function formatMoney(value: number | string | null | undefined): string {
  const num = Number(value || 0)
  return num.toFixed(2)
}

export function formatDateTime(value: string | null | undefined): string {
  if (!value) return '-'
  return dayjs(value).format('YYYY-MM-DD HH:mm:ss')
}

export function formatDate(value: string | null | undefined): string {
  if (!value) return '-'
  return dayjs(value).format('YYYY-MM-DD')
}

export function categoryText(value: string): string {
  const map: Record<string, string> = {
    [ExpenseCategory.DINING]: '餐饮',
    [ExpenseCategory.TRANSPORT]: '交通',
    [ExpenseCategory.LODGING]: '住宿',
    [ExpenseCategory.ENTERTAIN]: '娱乐',
    [ExpenseCategory.OTHER]: '其他',
  }
  return map[value] || value
}

export function splitTypeText(value: string): string {
  const map: Record<string, string> = {
    [SplitType.EQUAL]: '均摊',
    [SplitType.RATIO]: '按比例',
    [SplitType.AMOUNT]: '按金额',
  }
  return map[value] || value
}

export function groupStatusText(value: string): string {
  const map: Record<string, string> = {
    [GroupStatus.ACTIVE]: '进行中',
    [GroupStatus.ARCHIVED]: '已归档',
  }
  return map[value] || value
}

export function expenseStatusText(value: string): string {
  const map: Record<string, string> = {
    [ExpenseStatus.ACTIVE]: '有效',
    [ExpenseStatus.REFUNDED]: '已退款',
  }
  return map[value] || value
}

export function settlementStatusText(value: string): string {
  const map: Record<string, string> = {
    [SettlementStatus.PENDING]: '待结算',
    [SettlementStatus.SETTLED]: '已结算',
  }
  return map[value] || value
}

export function roleText(value: string): string {
  return value === 'admin' ? '管理员' : '普通用户'
}
