// 群组状态管理
import { defineStore } from 'pinia'
import {
  archiveGroupApi,
  createGroupApi,
  getGroupApi,
  inviteMemberApi,
  listMembersApi,
  listMyGroupsApi,
  removeMemberApi,
  updateGroupApi,
  type GroupInfo,
  type MemberInfo,
} from '@/api/group'

export const useGroupStore = defineStore('group', {
  state: () => ({
    groups: [] as GroupInfo[],
    total: 0,
    current: null as GroupInfo | null,
    members: [] as MemberInfo[],
    loading: false,
  }),
  actions: {
    async fetchGroups(page = 1, pageSize = 50) {
      this.loading = true
      try {
        const data = await listMyGroupsApi({ page, page_size: pageSize })
        this.groups = data.list
        this.total = data.total
      } finally {
        this.loading = false
      }
    },
    async fetchGroup(id: number) {
      this.current = await getGroupApi(id)
      return this.current
    },
    async createGroup(payload: { name: string; description?: string }) {
      const group = await createGroupApi(payload)
      await this.fetchGroups()
      return group
    },
    async updateGroup(id: number, payload: { name: string; description?: string }) {
      await updateGroupApi(id, payload)
      await this.fetchGroup(id)
    },
    async archiveGroup(id: number) {
      await archiveGroupApi(id)
      await this.fetchGroup(id)
      await this.fetchGroups()
    },
    async fetchMembers(groupId: number) {
      const data = await listMembersApi(groupId)
      this.members = data.list
    },
    async inviteMember(groupId: number, username: string) {
      await inviteMemberApi(groupId, username)
      await this.fetchMembers(groupId)
    },
    async removeMember(groupId: number, userId: number) {
      await removeMemberApi(groupId, userId)
      await this.fetchMembers(groupId)
    },
  },
})
