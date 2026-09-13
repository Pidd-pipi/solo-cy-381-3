// 审计日志状态管理
import { defineStore } from 'pinia'
import { listAuditLogsApi, type AuditLogInfo } from '@/api/audit'

export const useAuditStore = defineStore('audit', {
  state: () => ({
    logs: [] as AuditLogInfo[],
    total: 0,
    loading: false,
  }),
  actions: {
    async fetchLogs(params: { page?: number; page_size?: number; action?: string }) {
      this.loading = true
      try {
        const data = await listAuditLogsApi({ page_size: 50, ...params })
        this.logs = data.list
        this.total = data.total
      } finally {
        this.loading = false
      }
    },
  },
})
