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
    <Card className="dashboard-card daily-stats-card" title="Daily Stats" loading={loading && !stats}>
      <div className="daily-stat-grid">
        <Statistic title="Total Minutes" value={stats?.total_minutes ?? 0} prefix={<ClockCircleOutlined />} />
        <Statistic title="Completed" value={stats?.completed_count ?? 0} prefix={<CheckCircleOutlined />} />
        <Statistic title="Unfinished" value={stats?.unfinished_count ?? 0} prefix={<WarningOutlined />} />
        <Statistic title="Task Count" value={taskCount} prefix={<UnorderedListOutlined />} />
      </div>
    </Card>
  )
}
