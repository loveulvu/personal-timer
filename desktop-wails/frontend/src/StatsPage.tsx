import { FormEvent, useState } from 'react'
import { api, DailyStats, WeeklyStats } from './api'

type StatsPageProps = {
  connected: boolean
}

export function StatsPage({ connected }: StatsPageProps) {
  const today = todayString()
  const [dailyDate, setDailyDate] = useState(today)
  const [weeklyStart, setWeeklyStart] = useState(daysBefore(today, 6))
  const [weeklyEnd, setWeeklyEnd] = useState(today)
  const [daily, setDaily] = useState<DailyStats | null>(null)
  const [weekly, setWeekly] = useState<WeeklyStats | null>(null)
  const [dailyLoading, setDailyLoading] = useState(false)
  const [weeklyLoading, setWeeklyLoading] = useState(false)
  const [dailyError, setDailyError] = useState('')
  const [weeklyError, setWeeklyError] = useState('')

  async function queryDaily(event: FormEvent) {
    event.preventDefault()
    setDailyLoading(true)
    setDailyError('')
    try {
      setDaily(await api.getDailyStats(dailyDate))
    } catch (err) {
      setDailyError(errorMessage(err))
    } finally {
      setDailyLoading(false)
    }
  }

  async function queryWeekly(event: FormEvent) {
    event.preventDefault()
    setWeeklyLoading(true)
    setWeeklyError('')
    try {
      setWeekly(await api.getWeeklyStats(weeklyStart, weeklyEnd))
    } catch (err) {
      setWeeklyError(errorMessage(err))
    } finally {
      setWeeklyLoading(false)
    }
  }

  return (
    <div className="page-stack">
      <section className="panel">
        <h2>Daily Stats</h2>
        <form className="inline-form" onSubmit={queryDaily}>
          <label>
            date
            <input type="date" value={dailyDate} onChange={(event) => setDailyDate(event.target.value)} />
          </label>
          <button type="submit" disabled={!connected || dailyLoading}>
            Query daily stats
          </button>
        </form>
        {dailyError && <div className="message error">{dailyError}</div>}
        {dailyLoading && <p className="muted">Loading daily stats...</p>}
        {daily && (
          <>
            <StatCards
              totalMinutes={daily.total_minutes}
              completed={daily.completed_count}
              unfinished={daily.unfinished_count}
            />
            <h3 className="section-heading">Tasks</h3>
            <Table
              headers={['Title', 'Status', 'Estimated minutes', 'Actual minutes']}
              rows={daily.tasks.map((task) => [
                task.title,
                task.status,
                task.estimated_minutes,
                task.actual_minutes,
              ])}
              empty="No tasks for this date."
            />
          </>
        )}
      </section>

      <section className="panel">
        <h2>Weekly Stats</h2>
        <form className="inline-form" onSubmit={queryWeekly}>
          <label>
            start_date
            <input
              type="date"
              value={weeklyStart}
              onChange={(event) => setWeeklyStart(event.target.value)}
            />
          </label>
          <label>
            end_date
            <input
              type="date"
              value={weeklyEnd}
              onChange={(event) => setWeeklyEnd(event.target.value)}
            />
          </label>
          <button type="submit" disabled={!connected || weeklyLoading}>
            Query weekly stats
          </button>
        </form>
        {weeklyError && <div className="message error">{weeklyError}</div>}
        {weeklyLoading && <p className="muted">Loading weekly stats...</p>}
        {weekly && (
          <>
            <StatCards
              totalMinutes={weekly.total_minutes}
              completed={weekly.completed_count}
              unfinished={weekly.unfinished_count}
            />
            <h3 className="section-heading">Days</h3>
            <Table
              headers={['Date', 'Total minutes', 'Completed', 'Unfinished']}
              rows={weekly.days.map((day) => [
                day.date,
                day.total_minutes,
                day.completed_count,
                day.unfinished_count,
              ])}
              empty="No daily stats in this range."
            />
            <h3 className="section-heading">Projects</h3>
            <Table
              headers={['Project', 'Task count', 'Completed', 'Total minutes']}
              rows={weekly.projects.map((project) => [
                project.project_name,
                project.task_count,
                project.completed_count,
                project.total_minutes,
              ])}
              empty="No project stats in this range."
            />
          </>
        )}
      </section>
    </div>
  )
}

function StatCards({
  totalMinutes,
  completed,
  unfinished,
}: {
  totalMinutes: number
  completed: number
  unfinished: number
}) {
  return (
    <div className="stat-cards">
      <div><strong>{totalMinutes}</strong><span>total minutes</span></div>
      <div><strong>{completed}</strong><span>completed</span></div>
      <div><strong>{unfinished}</strong><span>unfinished</span></div>
    </div>
  )
}

function Table({
  headers,
  rows,
  empty,
}: {
  headers: string[]
  rows: Array<Array<string | number>>
  empty: string
}) {
  if (rows.length === 0) return <p className="muted">{empty}</p>
  return (
    <div className="table-wrap">
      <table>
        <thead><tr>{headers.map((header) => <th key={header}>{header}</th>)}</tr></thead>
        <tbody>
          {rows.map((row, rowIndex) => (
            <tr key={rowIndex}>{row.map((cell, cellIndex) => <td key={cellIndex}>{cell}</td>)}</tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

function todayString() {
  const now = new Date()
  const offset = now.getTimezoneOffset()
  return new Date(now.getTime() - offset * 60 * 1000).toISOString().slice(0, 10)
}

function daysBefore(date: string, count: number) {
  const value = new Date(`${date}T00:00:00`)
  value.setDate(value.getDate() - count)
  const offset = value.getTimezoneOffset()
  return new Date(value.getTime() - offset * 60 * 1000).toISOString().slice(0, 10)
}

function errorMessage(err: unknown) {
  if (err instanceof Error) return err.message
  if (typeof err === 'string') return err
  return 'Unknown error'
}
