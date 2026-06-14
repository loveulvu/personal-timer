import { CheckCircleOutlined, DeleteOutlined, EditOutlined, PauseCircleOutlined, PlayCircleOutlined, PlusOutlined, StopOutlined } from '@ant-design/icons'
import { Alert, Button, Card, Empty, Form, Input, InputNumber, List, Modal, Progress, Select, Space, Typography, message } from 'antd'
import { useState } from 'react'
import { api, DailyTask, Project } from '../../api'
import { StatusTag } from '../../components/StatusTag'
import { errorMessage } from '../../utils/labels'
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
  editCompletedTask: (task: DailyTask) => void
  deleteCompletedTask: (task: DailyTask) => void
  refresh: () => Promise<void>
}

type TaskForm = { projectId: number; title: string; estimatedMinutes: number }

export function TodayTasksCard({ date, tasks, projects, projectNames, loading, connected, openProjects, runAction, editCompletedTask, deleteCompletedTask, refresh }: TodayTasksCardProps) {
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
      message.success('每日任务创建成功')
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
      title="今日任务"
      extra={<Button type="text" icon={<PlusOutlined />} onClick={openCreate} disabled={!connected}>添加任务</Button>}
      loading={loading && tasks.length === 0}
    >
      {projects.length === 0 && connected && (
        <Alert type="warning" showIcon title="暂无项目，请先创建项目" action={<Button onClick={openProjects}>前往项目</Button>} />
      )}
      {tasks.length === 0 ? <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="当天暂无任务" /> : (
        <List
          className="task-list"
          dataSource={tasks}
          renderItem={(task) => (
            <List.Item actions={[<TaskActions key="actions" task={task} runAction={runAction} editCompletedTask={editCompletedTask} deleteCompletedTask={deleteCompletedTask} />]}>
              <List.Item.Meta
                avatar={<span className={`task-state-dot ${task.status}`}><CheckCircleOutlined /></span>}
                title={<Space wrap><Typography.Text strong delete={task.status === 'completed'}>{task.title}</Typography.Text><StatusTag value={task.status} /></Space>}
                description={<TaskDescription task={task} projectName={task.project_id ? projectNames.get(task.project_id) : undefined} />}
              />
            </List.Item>
          )}
        />
      )}
      <div className="tasks-progress">
        <div><Typography.Text type="secondary">已完成 {completed} / {tasks.length} 个任务</Typography.Text><Typography.Text type="secondary">{percent}%</Typography.Text></div>
        <Progress percent={percent} showInfo={false} size="small" />
      </div>

      <Modal title={`添加任务 · ${date}`} open={modalOpen} onCancel={() => setModalOpen(false)} footer={null}>
        <Form form={form} layout="vertical" onFinish={createTask}>
          <Form.Item name="projectId" label="项目" rules={[{ required: true }]}>
            <Select options={projects.map((project) => ({ value: project.id, label: project.name }))} />
          </Form.Item>
          <Form.Item name="title" label="任务标题" rules={[{ required: true, whitespace: true }]}>
            <Input placeholder="今天想专注完成什么？" />
          </Form.Item>
          <Form.Item name="estimatedMinutes" label="预计分钟数" rules={[{ required: true, type: 'number', min: 1 }]}>
            <InputNumber min={1} />
          </Form.Item>
          <Button type="primary" htmlType="submit">创建任务</Button>
        </Form>
      </Modal>
    </Card>
  )
}

function TaskDescription({ task, projectName }: { task: DailyTask; projectName?: string }) {
  return (
    <div>
      <div>{projectName ?? (task.project_id ? `项目 #${task.project_id}` : '未分配项目')} · 预计 {task.estimated_minutes} 分钟</div>
      {task.status === 'completed' && (
        <div>
          <div>完成备注：{task.finish_note ?? '暂无'}</div>
          <div>完成描述：{task.finish_description ?? '暂无'}</div>
          <div>实际时长：{Math.round(task.actual_seconds / 60)} 分钟</div>
        </div>
      )}
    </div>
  )
}

function TaskActions({ task, runAction, editCompletedTask, deleteCompletedTask }: { task: DailyTask; runAction: (task: DailyTask, action: TimerAction) => void; editCompletedTask: (task: DailyTask) => void; deleteCompletedTask: (task: DailyTask) => void }) {
  if (task.status === 'planned') return <Button size="small" type="primary" icon={<PlayCircleOutlined />} onClick={() => runAction(task, 'start')}>开始</Button>
  if (task.status === 'running') return <Space size={4}><Button size="small" icon={<PauseCircleOutlined />} onClick={() => runAction(task, 'pause')}>暂停</Button><Button size="small" icon={<StopOutlined />} onClick={() => runAction(task, 'finish')}>完成</Button></Space>
  if (task.status === 'paused') return <Space size={4}><Button size="small" type="primary" icon={<PlayCircleOutlined />} onClick={() => runAction(task, 'resume')}>继续</Button><Button size="small" icon={<StopOutlined />} onClick={() => runAction(task, 'finish')}>完成</Button></Space>
  if (task.status === 'completed') return <Space size={4}><Button size="small" icon={<EditOutlined />} onClick={() => editCompletedTask(task)}>编辑记录</Button><Button size="small" danger icon={<DeleteOutlined />} onClick={() => deleteCompletedTask(task)}>删除记录</Button></Space>
  return null
}
