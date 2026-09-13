<template>
  <div class="group-stats">
    <el-page-header :content="'数据统计 - ' + (groupStore.current?.name || '')" @back="router.push(`/groups/${groupId}`)" />
    <el-row :gutter="16" style="margin-top: 16px">
      <el-col :span="6">
        <el-card shadow="never" class="stat-card">
          <div class="stat-card__label">累计消费</div>
          <div class="stat-card__value"><MoneyText :value="stats?.total_expense || 0" /></div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never" class="stat-card">
          <div class="stat-card__label">消费笔数</div>
          <div class="stat-card__value">{{ stats?.total_count || 0 }} 笔</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never" class="stat-card">
          <div class="stat-card__label">成员数</div>
          <div class="stat-card__value">{{ groupStore.members.length }} 人</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never" class="stat-card">
          <div class="stat-card__label">类别数</div>
          <div class="stat-card__value">{{ stats?.category_stats.length || 0 }} 类</div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="16" style="margin-top: 16px">
      <el-col :span="12">
        <el-card shadow="never">
          <template #header><span>月度消费趋势</span></template>
          <div ref="monthChartRef" class="chart"></div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card shadow="never">
          <template #header><span>类别占比</span></template>
          <div ref="categoryChartRef" class="chart"></div>
        </el-card>
      </el-col>
    </el-row>

    <el-card shadow="never" style="margin-top: 16px">
      <template #header><span>成员消费排行</span></template>
      <el-table :data="stats?.member_ranks || []" border stripe>
        <el-table-column type="index" label="#" width="60" />
        <el-table-column prop="nickname" label="成员" min-width="140" />
        <el-table-column prop="username" label="用户名" min-width="120" />
        <el-table-column label="已付" width="130">
          <template #default="{ row }"><MoneyText :value="row.paid" /></template>
        </el-table-column>
        <el-table-column label="应付" width="130">
          <template #default="{ row }"><MoneyText :value="row.owed" :tone="'expense'" /></template>
        </el-table-column>
        <el-table-column label="净额" width="130">
          <template #default="{ row }">
            <MoneyText :value="row.net" :tone="row.net >= 0 ? 'income' : 'expense'" />
          </template>
        </el-table-column>
      </el-table>
      <EmptyState v-if="!stats?.member_ranks?.length" description="暂无统计数据" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import * as echarts from 'echarts'
import { useGroupStore } from '@/stores/group'
import { getGroupStatsApi, type GroupStats } from '@/api/stats'
import { categoryText } from '@/utils/format'
import MoneyText from '@/components/MoneyText.vue'
import EmptyState from '@/components/EmptyState.vue'

const route = useRoute()
const router = useRouter()
const groupStore = useGroupStore()
const groupId = Number(route.params.id)

const stats = ref<GroupStats | null>(null)
const monthChartRef = ref<HTMLElement>()
const categoryChartRef = ref<HTMLElement>()

onMounted(async () => {
  if (!groupStore.current) {
    await groupStore.fetchGroup(groupId)
  }
  await groupStore.fetchMembers(groupId)
  stats.value = await getGroupStatsApi(groupId)
  renderCharts()
})

function renderCharts() {
  if (!stats.value) return
  if (monthChartRef.value) {
    const chart = echarts.init(monthChartRef.value)
    chart.setOption({
      tooltip: { trigger: 'axis' },
      grid: { left: 40, right: 20, top: 30, bottom: 30 },
      xAxis: { type: 'category', data: stats.value.monthly_stats.map((m) => m.month) },
      yAxis: { type: 'value' },
      series: [
        {
          name: '消费金额',
          type: 'bar',
          data: stats.value.monthly_stats.map((m) => m.amount),
          itemStyle: { color: '#409eff', borderRadius: [4, 4, 0, 0] },
        },
      ],
    })
    window.addEventListener('resize', () => chart.resize())
  }
  if (categoryChartRef.value) {
    const chart = echarts.init(categoryChartRef.value)
    chart.setOption({
      tooltip: { trigger: 'item' },
      legend: { bottom: 0 },
      series: [
        {
          name: '类别占比',
          type: 'pie',
          radius: ['40%', '65%'],
          data: stats.value.category_stats.map((c) => ({ name: categoryText(c.category), value: c.amount })),
        },
      ],
    })
    window.addEventListener('resize', () => chart.resize())
  }
}
</script>

<style scoped>
.stat-card__label {
  color: #909399;
  font-size: 13px;
}
.stat-card__value {
  font-size: 24px;
  font-weight: 700;
  margin-top: 8px;
}
.chart {
  height: 320px;
}
</style>
