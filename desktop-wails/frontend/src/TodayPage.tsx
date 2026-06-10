import { PauseCircleOutlined, PlayCircleOutlined, StopOutlined } from '@ant-design/icons'
import { Alert, Button, Calendar, Card, DatePicker, Form, Input, InputNumber, Select, Space, Table, Typography, message } from 'antd'
import dayjs, { Dayjs } from 'dayjs'
import { useEffect, useMemo, useState } from 'react'
import { api, DailyTask, Project } from './api'
import { StatusTag } from './components/StatusTag'
import { errorMessage, timerActionLabel } from './utils/labels'

type Props = { connected: boolean; openProjects: () => void }
type TaskForm = { projectId: number; title: string; estimatedMinutes: number }

export function TodayPage({ connected, openProjects }: Props) {
  const [date, setDate] = useState(todayString())
  const [tasks, setTasks] = useState<DailyTask[]>([])
  const [projects, setProjects] = useState<Project[]>([])
  const [loading, setLoading] = useState(false)
  const [projectsLoading, setProjectsLoading] = useState(false)
  const [error, setError] = useState('')
  const [form] = Form.useForm<TaskForm>()
  const projectNames = useMemo(() => new Map(projects.map((p) => [p.id, p.name])), [projects])

  async function loadProjects() {
    if (!connected) return
    setProjectsLoading(true)
    try {
      const result = await api.getProjects()
      setProjects(result)
      const selected = form.getFieldValue('projectId')
      if (!result.some((project) => project.id === selected)) form.setFieldValue('projectId', result[0]?.id)
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setProjectsLoading(false)
    }
  }

  async function loadTasks(selectedDate = date) {
    if (!connected) return
    setLoading(true)
    setError('')
    try {
      setTasks(await api.listDailyTasks(selectedDate))
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }

  async function createTask(values: TaskForm) {
    if (projects.length === 0) return setError('暂无项目，请先创建项目')
    setLoading(true)
    setError('')
    try {
      await api.createDailyTask({
        project_id: values.projectId,
        task_date: date,
        title: values.title.trim(),
        estimated_minutes: values.estimatedMinutes,
      })
      form.setFieldsValue({ title: '', estimatedMinutes: 25 })
      await loadTasks(date)
      message.success('每日任务创建成功')
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }

  async function runAction(task: DailyTask, action: 'start' | 'pause' | 'resume' | 'finish') {
    setLoading(true)
    setError('')
    try {
      if (action === 'start') await api.startTask(task.id)
      if (action === 'pause') await api.pauseTask(task.id)
      if (action === 'resume') await api.resumeTask(task.id)
      if (action === 'finish') await api.finishTask(task.id)
      await loadTasks(date)
      message.success(`${timerActionLabel(action)}操作成功`)
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { if (connected) loadProjects() }, [connected])
  useEffect(() => { if (connected) loadTasks(date) }, [connected, date])

  const columns = [
    { title: '任务标题', dataIndex: 'title', key: 'title' },
    { title: '状态', dataIndex: 'status', key: 'status', render: (value: string) => <StatusTag value={value} /> },
    { title: '预计分钟数', dataIndex: 'estimated_minutes', key: 'estimated_minutes', render: (value: number) => `${value} 分钟` },
    { title: '项目', dataIndex: 'project_id', key: 'project_id', render: (value: number | null) => value ? projectNames.get(value) ?? `#${value}` : '-' },
    { title: '计时操作', key: 'actions', render: (_: unknown, task: DailyTask) => <TimerControls task={task} runAction={runAction} /> },
  ]

  return (
    <div className="page-stack">
      {error && <Alert type="error" showIcon title={error} />}
      <div className="today-grid">
        <Card title="日历">
          <Calendar fullscreen={false} value={dayjs(date)} onSelect={(value) => setDate(value.format('YYYY-MM-DD'))} />
        </Card>
        <Card title="创建每日任务">
          {projects.length === 0 && !projectsLoading && (
            <Alert type="warning" showIcon title="暂无项目，请先创建项目" action={<Button onClick={openProjects}>前往项目</Button>} />
          )}
          <Form form={form} layout="vertical" initialValues={{ estimatedMinutes: 25 }} onFinish={createTask}>
            <Form.Item name="projectId" label="项目" rules={[{ required: true }]}>
              <Select loading={projectsLoading} disabled={!connected || projects.length === 0} options={projects.map((project) => ({ value: project.id, label: project.name }))} />
            </Form.Item>
            <Form.Item label="任务日期">
              <DatePicker value={dayjs(date)} onChange={(value) => value && setDate(value.format('YYYY-MM-DD'))} />
            </Form.Item>
            <Form.Item name="title" label="任务标题" rules={[{ required: true, whitespace: true }]}>
              <Input placeholder="例如：阅读文档" />
            </Form.Item>
            <Form.Item name="estimatedMinutes" label="预计分钟数" rules={[{ required: true, type: 'number', min: 1 }]}>
              <InputNumber min={1} />
            </Form.Item>
            <Button type="primary" htmlType="submit" loading={loading} disabled={!connected || projects.length === 0}>创建</Button>
          </Form>
        </Card>
      </div>
      <Card title={<Space><Typography.Text strong>每日任务</Typography.Text><DatePicker value={dayjs(date)} onChange={(value: Dayjs | null) => value && setDate(value.format('YYYY-MM-DD'))} /></Space>}>
        <Table rowKey="id" loading={loading} dataSource={tasks} columns={columns} pagination={false} locale={{ emptyText: '当天暂无任务' }} />
      </Card>
    </div>
  )
}

function TimerControls({ task, runAction }: { task: DailyTask; runAction: (task: DailyTask, action: 'start' | 'pause' | 'resume' | 'finish') => void }) {
  if (task.status === 'planned') return <Button icon={<PlayCircleOutlined />} onClick={() => runAction(task, 'start')}>开始</Button>
  if (task.status === 'running') return <Space><Button icon={<PauseCircleOutlined />} onClick={() => runAction(task, 'pause')}>暂停</Button><Button icon={<StopOutlined />} onClick={() => runAction(task, 'finish')}>完成</Button></Space>
  if (task.status === 'paused') return <Space><Button icon={<PlayCircleOutlined />} onClick={() => runAction(task, 'resume')}>继续</Button><Button icon={<StopOutlined />} onClick={() => runAction(task, 'finish')}>完成</Button></Space>
  return null
}

function todayString() { return dayjs().format('YYYY-MM-DD') }
