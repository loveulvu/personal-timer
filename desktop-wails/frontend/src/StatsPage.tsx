import { BarChartOutlined, CheckCircleOutlined, ClockCircleOutlined } from '@ant-design/icons'
import { Alert, Button, Card, Col, DatePicker, Row, Space, Statistic, Table, Typography } from 'antd'
import dayjs from 'dayjs'
import { useState } from 'react'
import { api, DailyStats, WeeklyStats } from './api'
import { StatusTag } from './components/StatusTag'
import { errorMessage } from './utils/labels'

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
      <header className="section-header"><div><Typography.Title level={2}>统计</Typography.Title><Typography.Text type="secondary">查看每日明细与指定日期范围的每周统计。</Typography.Text></div></header>
      <Card title="每日统计" extra={<Space><DatePicker value={dailyDate} onChange={(value) => value && setDailyDate(value)} /><Button type="primary" onClick={queryDaily} loading={dailyLoading} disabled={!connected}>查询</Button></Space>}>
        {dailyError && <Alert className="card-alert" type="error" showIcon title={dailyError} />}
        {daily && <>
          <StatCards totalMinutes={daily.total_minutes} completed={daily.completed_count} unfinished={daily.unfinished_count} />
          <Typography.Title level={5}>任务明细</Typography.Title>
          <Table
            rowKey="task_id"
            pagination={false}
            dataSource={daily.tasks}
            columns={[
              { title: '任务标题', dataIndex: 'title', key: 'title' },
              { title: '状态', dataIndex: 'status', key: 'status', render: (value: string) => <StatusTag value={value} /> },
              { title: '预计分钟数', dataIndex: 'estimated_minutes', key: 'estimated_minutes' },
              { title: '实际分钟数', dataIndex: 'actual_minutes', key: 'actual_minutes' },
            ]}
          />
        </>}
      </Card>

      <Card title="每周统计" extra={<Space><DatePicker.RangePicker value={weeklyRange} onChange={(value) => value?.[0] && value?.[1] && setWeeklyRange([value[0], value[1]])} /><Button type="primary" onClick={queryWeekly} loading={weeklyLoading} disabled={!connected}>查询</Button></Space>}>
        {weeklyError && <Alert className="card-alert" type="error" showIcon title={weeklyError} />}
        {weekly && <>
          <StatCards totalMinutes={weekly.total_minutes} completed={weekly.completed_count} unfinished={weekly.unfinished_count} />
          <Typography.Title level={5}>日期列表</Typography.Title>
          <Table rowKey="date" pagination={false} dataSource={weekly.days} columns={[
            { title: '日期', dataIndex: 'date', key: 'date' },
            { title: '总分钟数', dataIndex: 'total_minutes', key: 'total_minutes' },
            { title: '完成数量', dataIndex: 'completed_count', key: 'completed_count' },
            { title: '未完成数量', dataIndex: 'unfinished_count', key: 'unfinished_count' },
          ]} />
          <Typography.Title level={5}>项目汇总</Typography.Title>
          <Table rowKey="project_id" pagination={false} dataSource={weekly.projects} columns={[
            { title: '项目', dataIndex: 'project_name', key: 'project_name' },
            { title: '任务数', dataIndex: 'task_count', key: 'task_count' },
            { title: '完成数量', dataIndex: 'completed_count', key: 'completed_count' },
            { title: '总分钟数', dataIndex: 'total_minutes', key: 'total_minutes' },
          ]} />
        </>}
      </Card>
    </div>
  )
}

function StatCards({ totalMinutes, completed, unfinished }: { totalMinutes: number; completed: number; unfinished: number }) {
  return <Row gutter={[16, 16]} className="stat-row">
    <Col xs={24} md={8}><Card size="small"><Statistic title="总分钟数" value={totalMinutes} prefix={<ClockCircleOutlined />} /></Card></Col>
    <Col xs={24} md={8}><Card size="small"><Statistic title="完成数量" value={completed} prefix={<CheckCircleOutlined />} /></Card></Col>
    <Col xs={24} md={8}><Card size="small"><Statistic title="未完成数量" value={unfinished} prefix={<BarChartOutlined />} /></Card></Col>
  </Row>
}
