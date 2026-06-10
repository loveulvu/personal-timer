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
    <Card className="dashboard-card focus-card" title="当前计时" loading={loading && !dailyStats}>
      {task ? (
        <>
          <Progress
            type="circle"
            percent={percent}
            size={210}
            strokeWidth={7}
            format={() => (
              <div className="focus-progress-label">
                <Typography.Text type="secondary">已专注</Typography.Text>
                <Typography.Title level={2}>{actualMinutes} 分钟</Typography.Title>
                <Typography.Text type="secondary">预计 {task.estimated_minutes} 分钟</Typography.Text>
              </div>
            )}
          />
          <div className="focus-task">
            <Typography.Title level={4}>{task.title}</Typography.Title>
            <Typography.Text type="secondary">{projectName ?? (task.project_id ? `项目 #${task.project_id}` : '未分配项目')}</Typography.Text>
          </div>
          <Space>
            <Button icon={<PauseOutlined />} onClick={() => runAction(task, 'pause')}>暂停</Button>
            <Button type="primary" icon={<StopOutlined />} onClick={() => runAction(task, 'finish')}>完成</Button>
          </Space>
        </>
      ) : (
        <div className="focus-empty">
          <Progress type="circle" percent={0} size={210} format={() => <Typography.Text type="secondary">暂无进行中任务</Typography.Text>} />
          <Typography.Title level={4}>当前没有进行中的任务</Typography.Title>
          <Typography.Text type="secondary">请从今日任务中开始一个任务</Typography.Text>
        </div>
      )}
    </Card>
  )
}
