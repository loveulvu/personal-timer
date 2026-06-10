import { CheckCircleOutlined, ClockCircleOutlined, UnorderedListOutlined, WarningOutlined } from '@ant-design/icons'
import { Card, Statistic } from 'antd'
import { DailyStats } from '../../api'

type DailyStatsCardProps = {
  stats: DailyStats | null
  taskCount: number
  loading: boolean
}

export function DailyStatsCard({ stats, taskCount, loading }: DailyStatsCardProps) {
  return (
    <Card className="dashboard-card daily-stats-card" title="今日统计" loading={loading && !stats}>
      <div className="daily-stat-grid">
        <Statistic title="总分钟数" value={stats?.total_minutes ?? 0} prefix={<ClockCircleOutlined />} />
        <Statistic title="已完成任务" value={stats?.completed_count ?? 0} prefix={<CheckCircleOutlined />} />
        <Statistic title="未完成任务" value={stats?.unfinished_count ?? 0} prefix={<WarningOutlined />} />
        <Statistic title="任务数" value={taskCount} prefix={<UnorderedListOutlined />} />
      </div>
    </Card>
  )
}
