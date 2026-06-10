import { CalendarOutlined } from '@ant-design/icons'
import { Alert, DatePicker, Typography, message } from 'antd'
import dayjs, { Dayjs } from 'dayjs'
import { useEffect, useMemo, useState } from 'react'
import { api, DailyStats, DailyTask, Project, WeeklyStats } from '../../api'
import { DailyStatsCard } from './DailyStatsCard'
import { DashboardCalendar } from './DashboardCalendar'
import { FocusTimerCard } from './FocusTimerCard'
import { TodayTasksCard } from './TodayTasksCard'
import { WeeklySummaryCard } from './WeeklySummaryCard'

type DashboardPageProps = {
  connected: boolean
  openProjects: () => void
}

export type TimerAction = 'start' | 'pause' | 'resume' | 'finish'

export function DashboardPage({ connected, openProjects }: DashboardPageProps) {
  const [date, setDate] = useState<Dayjs>(dayjs())
  const [tasks, setTasks] = useState<DailyTask[]>([])
  const [projects, setProjects] = useState<Project[]>([])
  const [dailyStats, setDailyStats] = useState<DailyStats | null>(null)
  const [weeklyStats, setWeeklyStats] = useState<WeeklyStats | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const dateString = date.format('YYYY-MM-DD')
  const weekStart = date.subtract(6, 'day').format('YYYY-MM-DD')
  const projectNames = useMemo(() => new Map(projects.map((project) => [project.id, project.name])), [projects])
  const runningTask = tasks.find((task) => task.status === 'running') ?? null

  async function loadDashboard() {
    if (!connected) return
    setLoading(true)
    setError('')
    const results = await Promise.allSettled([
      api.listDailyTasks(dateString),
      api.getDailyStats(dateString),
      api.getWeeklyStats(weekStart, dateString),
    ])
    if (results[0].status === 'fulfilled') setTasks(results[0].value)
    if (results[1].status === 'fulfilled') setDailyStats(results[1].value)
    if (results[2].status === 'fulfilled') setWeeklyStats(results[2].value)
    const failures = results.flatMap((result, index) => {
      if (result.status === 'fulfilled') return []
      const labels = ['Daily tasks', 'Daily stats', 'Weekly stats']
      return [`${labels[index]} failed to load: ${errorMessage(result.reason)}`]
    })
    if (failures.length > 0) setError(failures.join(' '))
    setLoading(false)
  }

  async function loadProjects() {
    if (!connected) return
    try {
      setProjects(await api.getProjects())
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  async function runAction(task: DailyTask, action: TimerAction) {
    setLoading(true)
    setError('')
    try {
      if (action === 'start') await api.startTask(task.id)
      if (action === 'pause') await api.pauseTask(task.id)
      if (action === 'resume') await api.resumeTask(task.id)
      if (action === 'finish') await api.finishTask(task.id)
      await loadDashboard()
      message.success(`${action[0].toUpperCase()}${action.slice(1)} completed`)
    } catch (err) {
      const text = errorMessage(err)
      setError(text)
      message.error(text)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    if (connected) loadProjects()
  }, [connected])

  useEffect(() => {
    if (connected) loadDashboard()
  }, [connected, dateString])

  return (
    <div className="dashboard-page">
      <header className="dashboard-header">
        <div>
          <Typography.Title level={2}>Personal Timer Dashboard</Typography.Title>
          <Typography.Text type="secondary">Stay focused and track your progress.</Typography.Text>
        </div>
        <DatePicker
          value={date}
          suffixIcon={<CalendarOutlined />}
          onChange={(value) => value && setDate(value)}
          allowClear={false}
        />
      </header>

      {error && <Alert type="error" showIcon title={error} />}

      <div className="dashboard-grid">
        <TodayTasksCard
          date={dateString}
          tasks={tasks}
          projects={projects}
          projectNames={projectNames}
          loading={loading}
          connected={connected}
          openProjects={openProjects}
          runAction={runAction}
          refresh={loadDashboard}
        />
        <FocusTimerCard
          task={runningTask}
          projectName={runningTask ? projectNames.get(runningTask.project_id ?? -1) : undefined}
          dailyStats={dailyStats}
          loading={loading}
          runAction={runAction}
        />
        <div className="dashboard-side-stack">
          <DashboardCalendar date={date} setDate={setDate} />
          <DailyStatsCard stats={dailyStats} taskCount={tasks.length} loading={loading} />
        </div>
      </div>

      <WeeklySummaryCard stats={weeklyStats} loading={loading} />
    </div>
  )
}

function errorMessage(err: unknown) {
  return err instanceof Error ? err.message : typeof err === 'string' ? err : 'Unknown error'
}
