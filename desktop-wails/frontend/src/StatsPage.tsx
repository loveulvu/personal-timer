import { BarChartOutlined, CheckCircleOutlined, ClockCircleOutlined } from '@ant-design/icons'
import { Alert, Button, Card, Col, DatePicker, Row, Space, Statistic, Table, Typography } from 'antd'
import dayjs from 'dayjs'
import { useState } from 'react'
import { api, DailyStats, WeeklyStats } from './api'
import { StatusTag } from './components/StatusTag'

type Props = { connected: boolean }

export function StatsPage({ connected }: Props) {
  const today = dayjs()
  const [dailyDate, setDailyDate] = useState(today)
  const [weeklyRange, setWeeklyRange] = useState<[dayjs.Dayjs, dayjs.Dayjs]>([today.subtract(6, 'day'), today])
  const [daily, setDaily] = useState<DailyStats | null>(null)
  const [weekly, setWeekly] = useState<WeeklyStats | null>(null)
  const [dailyLoading, setDailyLoading] = useState(false)
  const [weeklyLoading, setWeeklyLoading] = useState(false)
  const [dailyError, setDailyError] = useState('')
  const [weeklyError, setWeeklyError] = useState('')

  async function queryDaily() {
    setDailyLoading(true); setDailyError('')
    try { setDaily(await api.getDailyStats(dailyDate.format('YYYY-MM-DD'))) }
    catch (err) { setDailyError(errorMessage(err)) } finally { setDailyLoading(false) }
  }

  async function queryWeekly() {
    setWeeklyLoading(true); setWeeklyError('')
    try { setWeekly(await api.getWeeklyStats(weeklyRange[0].format('YYYY-MM-DD'), weeklyRange[1].format('YYYY-MM-DD'))) }
    catch (err) { setWeeklyError(errorMessage(err)) } finally { setWeeklyLoading(false) }
  }

  return (
    <div className="page-stack">
      <header className="section-header"><div><Typography.Title level={2}>Stats</Typography.Title><Typography.Text type="secondary">Review daily detail and compare a selected weekly range.</Typography.Text></div></header>
      <Card title="Daily Stats" extra={<Space><DatePicker value={dailyDate} onChange={(value) => value && setDailyDate(value)} /><Button type="primary" onClick={queryDaily} loading={dailyLoading} disabled={!connected}>Query</Button></Space>}>
        {dailyError && <Alert className="card-alert" type="error" showIcon title={dailyError} />}
        {daily && <>
          <StatCards totalMinutes={daily.total_minutes} completed={daily.completed_count} unfinished={daily.unfinished_count} />
          <Typography.Title level={5}>Tasks</Typography.Title>
          <Table
            rowKey="task_id"
            pagination={false}
            dataSource={daily.tasks}
            columns={[
              { title: 'Title', dataIndex: 'title', key: 'title' },
              { title: 'Status', dataIndex: 'status', key: 'status', render: (value: string) => <StatusTag value={value} /> },
              { title: 'Estimated minutes', dataIndex: 'estimated_minutes', key: 'estimated_minutes' },
              { title: 'Actual minutes', dataIndex: 'actual_minutes', key: 'actual_minutes' },
            ]}
          />
        </>}
      </Card>

      <Card title="Weekly Stats" extra={<Space><DatePicker.RangePicker value={weeklyRange} onChange={(value) => value?.[0] && value?.[1] && setWeeklyRange([value[0], value[1]])} /><Button type="primary" onClick={queryWeekly} loading={weeklyLoading} disabled={!connected}>Query</Button></Space>}>
        {weeklyError && <Alert className="card-alert" type="error" showIcon title={weeklyError} />}
        {weekly && <>
          <StatCards totalMinutes={weekly.total_minutes} completed={weekly.completed_count} unfinished={weekly.unfinished_count} />
          <Typography.Title level={5}>Days</Typography.Title>
          <Table rowKey="date" pagination={false} dataSource={weekly.days} columns={[
            { title: 'Date', dataIndex: 'date', key: 'date' },
            { title: 'Total minutes', dataIndex: 'total_minutes', key: 'total_minutes' },
            { title: 'Completed', dataIndex: 'completed_count', key: 'completed_count' },
            { title: 'Unfinished', dataIndex: 'unfinished_count', key: 'unfinished_count' },
          ]} />
          <Typography.Title level={5}>Projects</Typography.Title>
          <Table rowKey="project_id" pagination={false} dataSource={weekly.projects} columns={[
            { title: 'Project', dataIndex: 'project_name', key: 'project_name' },
            { title: 'Task count', dataIndex: 'task_count', key: 'task_count' },
            { title: 'Completed', dataIndex: 'completed_count', key: 'completed_count' },
            { title: 'Total minutes', dataIndex: 'total_minutes', key: 'total_minutes' },
          ]} />
        </>}
      </Card>
    </div>
  )
}

function StatCards({ totalMinutes, completed, unfinished }: { totalMinutes: number; completed: number; unfinished: number }) {
  return <Row gutter={[16, 16]} className="stat-row">
    <Col xs={24} md={8}><Card size="small"><Statistic title="Total minutes" value={totalMinutes} prefix={<ClockCircleOutlined />} /></Card></Col>
    <Col xs={24} md={8}><Card size="small"><Statistic title="Completed" value={completed} prefix={<CheckCircleOutlined />} /></Card></Col>
    <Col xs={24} md={8}><Card size="small"><Statistic title="Unfinished" value={unfinished} prefix={<BarChartOutlined />} /></Card></Col>
  </Row>
}

function errorMessage(err: unknown) { return err instanceof Error ? err.message : typeof err === 'string' ? err : 'Unknown error' }
