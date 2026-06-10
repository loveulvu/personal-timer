import { PauseOutlined, StopOutlined } from '@ant-design/icons'
import { Button, Card, Progress, Space, Typography } from 'antd'
import { DailyStats, DailyTask } from '../../api'
import { TimerAction } from './DashboardPage'

type FocusTimerCardProps = {
  task: DailyTask | null
  projectName?: string
  dailyStats: DailyStats | null
  loading: boolean
  runAction: (task: DailyTask, action: TimerAction) => void
}

export function FocusTimerCard({ task, projectName, dailyStats, loading, runAction }: FocusTimerCardProps) {
  const actualMinutes = task ? dailyStats?.tasks?.find((item) => item.task_id === task.id)?.actual_minutes ?? 0 : 0
  const percent = task?.estimated_minutes ? Math.min(100, Math.round((actualMinutes / task.estimated_minutes) * 100)) : 0

  return (
    <Card className="dashboard-card focus-card" title="Focus Timer" loading={loading && !dailyStats}>
      {task ? (
        <>
          <Progress
            type="circle"
            percent={percent}
            size={210}
            strokeWidth={7}
            format={() => (
              <div className="focus-progress-label">
                <Typography.Text type="secondary">Focused</Typography.Text>
                <Typography.Title level={2}>{actualMinutes} min</Typography.Title>
                <Typography.Text type="secondary">of {task.estimated_minutes} min</Typography.Text>
              </div>
            )}
          />
          <div className="focus-task">
            <Typography.Title level={4}>{task.title}</Typography.Title>
            <Typography.Text type="secondary">{projectName ?? (task.project_id ? `Project #${task.project_id}` : 'Unassigned')}</Typography.Text>
          </div>
          <Space>
            <Button icon={<PauseOutlined />} onClick={() => runAction(task, 'pause')}>Pause</Button>
            <Button type="primary" icon={<StopOutlined />} onClick={() => runAction(task, 'finish')}>Finish</Button>
          </Space>
        </>
      ) : (
        <div className="focus-empty">
          <Progress type="circle" percent={0} size={210} format={() => <Typography.Text type="secondary">No active task</Typography.Text>} />
          <Typography.Title level={4}>No active task</Typography.Title>
          <Typography.Text type="secondary">Start a task from Today's Tasks.</Typography.Text>
        </div>
      )}
    </Card>
  )
}
