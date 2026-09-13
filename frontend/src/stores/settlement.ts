// 结算建议状态管理
import { defineStore } from 'pinia'
import {
  generateSettlementsApi,
  listBalancesApi,
  listPendingSettlementsApi,
  listSettlementsApi,
  settleApi,
  type GroupBalance,
  type SettlementInfo,
} from '@/api/settlement'

export const useSettlementStore = defineStore('settlement', {
  state: () => ({
    settlements: [] as SettlementInfo[],
    pending: [] as SettlementInfo[],
    balances: [] as GroupBalance[],
  }),
  actions: {
    async fetchSettlements(groupId: number) {
      const data = await listSettlementsApi(groupId)
      this.settlements = data.list
    },
    async generate(groupId: number) {
      const data = await generateSettlementsApi(groupId)
      this.settlements = data.list
    },
    async fetchPending() {
      const data = await listPendingSettlementsApi()
      this.pending = data.list
    },
    async settle(ids: number[]) {
      await settleApi(ids)
      await this.fetchPending()
    },
    async fetchBalances(groupId: number) {
      const data = await listBalancesApi(groupId)
      this.balances = data.list
    },
  },
})
