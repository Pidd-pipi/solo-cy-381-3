// 消费记录状态管理
import { defineStore } from 'pinia'
import {
  createExpenseApi,
  listExpensesApi,
  refundExpenseApi,
  updateExpenseApi,
  type ExpenseInfo,
  type ExpensePayload,
  type ExpenseQuery,
} from '@/api/expense'

export const useExpenseStore = defineStore('expense', {
  state: () => ({
    expenses: [] as ExpenseInfo[],
    total: 0,
    loading: false,
  }),
  actions: {
    async fetchExpenses(groupId: number, query: ExpenseQuery) {
      this.loading = true
      try {
        const data = await listExpensesApi(groupId, { page_size: 50, ...query })
        this.expenses = data.list
        this.total = data.total
      } finally {
        this.loading = false
      }
    },
    async createExpense(payload: ExpensePayload) {
      const expense = await createExpenseApi(payload)
      return expense
    },
    async updateExpense(id: number, payload: ExpensePayload) {
      await updateExpenseApi(id, payload)
    },
    async refundExpense(id: number) {
      await refundExpenseApi(id)
    },
  },
})
