// 审计日志模块 API
import { get } from '@/utils/request'
import type { PageData } from '@/utils/request'

export interface AuditLogInfo {
  id: number
  user_id: number
  username: string
  action: string
  resource_type: string
  resource_id: string
  detail: string
  ip: string
  request_id: string
  created_at: string
}

export function listAuditLogsApi(params: { page?: number; page_size?: number; action?: string; user_id?: number }) {
  return get<PageData<AuditLogInfo>>('/audit-logs', params)
}
