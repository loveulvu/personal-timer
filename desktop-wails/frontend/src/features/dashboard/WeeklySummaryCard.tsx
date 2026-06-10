import { CheckCircleOutlined, ClockCircleOutlined, FolderOutlined, WarningOutlined } from '@ant-design/icons'
import { Card, Empty, Statistic, Tag, Typography } from 'antd'
import dayjs from 'dayjs'
import { WeeklyStats } from '../../api'

type WeeklySummaryCardProps = {
  stats: WeeklyStats | null
  loading: boolean
}

export function WeeklySummaryCard({ stats, loading }: WeeklySummaryCardProps) {
  const maxMinutes = Math.max(...(stats?.days?.map((day) => day.total_minutes) ?? [0]), 1)

  return (
    <Card className="dashboard-card weekly-card" title="Weekly Summary" loading={loading && !stats}>
      {!stats ? <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="Weekly stats unavailable" /> : (
        <div className="weekly-content">
          <div className="weekly-chart">
            {stats.days.map((day) => (
              <div className="weekly-day" key={day.date}>
                <Typography.Text type="secondary">{day.total_minutes}m</Typography.Text>
                <div className="weekly-bar-track">
                  <div className="weekly-bar" style={{ height: `${Math.max(5, (day.total_minutes / maxMinutes) * 100)}%` }} />
                </div>
                <Typography.Text>{dayjs(day.date).format('ddd')}</Typography.Text>
              </div>
            ))}
          </div>
          <div className="weekly-metrics">
            <Statistic title="Total Minutes" value={stats.total_minutes} prefix={<ClockCircleOutlined />} />
            <Statistic title="Completed" value={stats.completed_count} prefix={<CheckCircleOutlined />} />
            <Statistic title="Unfinished" value={stats.unfinished_count} prefix={<WarningOutlined />} />
            <div className="weekly-projects">
              <Typography.Text type="secondary"><FolderOutlined /> Projects</Typography.Text>
              <div>{stats.projects.map((project) => <Tag key={project.project_id} color="blue">{project.project_name} · {project.total_minutes}m</Tag>)}</div>
            </div>
          </div>
        </div>
      )}
    </Card>
  )
}
