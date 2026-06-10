import { CheckCircleOutlined, PauseCircleOutlined, PlayCircleOutlined, PlusOutlined, StopOutlined } from '@ant-design/icons'
import { Alert, Button, Card, Empty, Form, Input, InputNumber, List, Modal, Progress, Select, Space, Typography, message } from 'antd'
import { useState } from 'react'
import { api, DailyTask, Project } from '../../api'
import { StatusTag } from '../../components/StatusTag'
import { TimerAction } from './DashboardPage'

type TodayTasksCardProps = {
  date: string
  tasks: DailyTask[]
  projects: Project[]
  projectNames: Map<number, string>
  loading: boolean
  connected: boolean
  openProjects: () => void
  runAction: (task: DailyTask, action: TimerAction) => void
  refresh: () => Promise<void>
}

type TaskForm = { projectId: number; title: string; estimatedMinutes: number }

export function TodayTasksCard({ date, tasks, projects, projectNames, loading, connected, openProjects, runAction, refresh }: TodayTasksCardProps) {
  const [modalOpen, setModalOpen] = useState(false)
  const [form] = Form.useForm<TaskForm>()
  const completed = tasks.filter((task) => task.status === 'completed').length
  const percent = tasks.length ? Math.round((completed / tasks.length) * 100) : 0

  async function createTask(values: TaskForm) {
    try {
      await api.createDailyTask({
        project_id: values.projectId,
        task_date: date,
        title: values.title.trim(),
        estimated_minutes: values.estimatedMinutes,
      })
      form.resetFields()
      setModalOpen(false)
      await refresh()
      message.success('Daily task created')
    } catch (err) {
      message.error(errorMessage(err))
    }
  }

  function openCreate() {
    if (projects.length === 0) {
      openProjects()
      return
    }
    form.setFieldsValue({ projectId: projects[0].id, estimatedMinutes: 25, title: '' })
    setModalOpen(true)
  }

  return (
    <Card
      className="dashboard-card tasks-card"
      title="Today's Tasks"
      extra={<Button type="text" icon={<PlusOutlined />} onClick={openCreate} disabled={!connected}>Add task</Button>}
      loading={loading && tasks.length === 0}
    >
      {projects.length === 0 && connected && (
        <Alert type="warning" showIcon title="Create a project before adding tasks." action={<Button onClick={openProjects}>Projects</Button>} />
      )}
      {tasks.length === 0 ? <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="No tasks for this date" /> : (
        <List
          className="task-list"
          dataSource={tasks}
          renderItem={(task) => (
            <List.Item actions={[<TaskActions key="actions" task={task} runAction={runAction} />]}>
              <List.Item.Meta
                avatar={<span className={`task-state-dot ${task.status}`}><CheckCircleOutlined /></span>}
                title={<Space wrap><Typography.Text strong delete={task.status === 'completed'}>{task.title}</Typography.Text><StatusTag value={task.status} /></Space>}
                description={`${task.project_id ? projectNames.get(task.project_id) ?? `Project #${task.project_id}` : 'Unassigned'} · ${task.estimated_minutes} min`}
              />
            </List.Item>
          )}
        />
      )}
      <div className="tasks-progress">
        <div><Typography.Text type="secondary">{completed} of {tasks.length} tasks completed</Typography.Text><Typography.Text type="secondary">{percent}%</Typography.Text></div>
        <Progress percent={percent} showInfo={false} size="small" />
      </div>

      <Modal title={`Add task · ${date}`} open={modalOpen} onCancel={() => setModalOpen(false)} footer={null}>
        <Form form={form} layout="vertical" onFinish={createTask}>
          <Form.Item name="projectId" label="Project" rules={[{ required: true }]}>
            <Select options={projects.map((project) => ({ value: project.id, label: project.name }))} />
          </Form.Item>
          <Form.Item name="title" label="Task title" rules={[{ required: true, whitespace: true }]}>
            <Input placeholder="What do you want to focus on?" />
          </Form.Item>
          <Form.Item name="estimatedMinutes" label="Estimated minutes" rules={[{ required: true, type: 'number', min: 1 }]}>
            <InputNumber min={1} />
          </Form.Item>
          <Button type="primary" htmlType="submit">Create task</Button>
        </Form>
      </Modal>
    </Card>
  )
}

function TaskActions({ task, runAction }: { task: DailyTask; runAction: (task: DailyTask, action: TimerAction) => void }) {
  if (task.status === 'planned') return <Button size="small" type="primary" icon={<PlayCircleOutlined />} onClick={() => runAction(task, 'start')}>Start</Button>
  if (task.status === 'running') return <Space size={4}><Button size="small" icon={<PauseCircleOutlined />} onClick={() => runAction(task, 'pause')}>Pause</Button><Button size="small" icon={<StopOutlined />} onClick={() => runAction(task, 'finish')}>Finish</Button></Space>
  if (task.status === 'paused') return <Space size={4}><Button size="small" type="primary" icon={<PlayCircleOutlined />} onClick={() => runAction(task, 'resume')}>Resume</Button><Button size="small" icon={<StopOutlined />} onClick={() => runAction(task, 'finish')}>Finish</Button></Space>
  return null
}

function errorMessage(err: unknown) {
  return err instanceof Error ? err.message : typeof err === 'string' ? err : 'Unknown error'
}
