import { DeleteOutlined, EditOutlined, ReloadOutlined } from '@ant-design/icons'
import { Alert, Button, Card, Checkbox, Form, Input, Modal, Select, Space, Table, Tag, Typography, message } from 'antd'
import { useEffect, useState } from 'react'
import { api, Project, ProjectCategory, ProjectInput } from './api'
import { errorMessage } from './utils/labels'

type Props = { connected: boolean }

export function ProjectsPage({ connected }: Props) {
  const [projects, setProjects] = useState<Project[]>([])
  const [editing, setEditing] = useState<Project | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [createForm] = Form.useForm<ProjectInput>()
  const [editForm] = Form.useForm<ProjectInput>()

  async function loadProjects() {
    if (!connected) return
    setLoading(true); setError('')
    try { setProjects(await api.getProjects()) } catch (err) { setError(errorMessage(err)) } finally { setLoading(false) }
  }

  async function createProject(values: ProjectInput) {
    setLoading(true); setError('')
    try {
      await api.createProject({ ...values, name: values.name.trim(), is_fixed: values.is_fixed ?? false, category: values.category ?? 'study', include_in_summary: values.include_in_summary ?? true })
      createForm.resetFields(); await loadProjects(); message.success('项目创建成功')
    } catch (err) { setError(errorMessage(err)) } finally { setLoading(false) }
  }

  function startEditing(project: Project) {
    setEditing(project)
    editForm.setFieldsValue({ name: project.name, description: project.description, is_fixed: project.is_fixed, category: project.category ?? 'study', include_in_summary: project.include_in_summary ?? true })
  }

  async function updateProject() {
    if (!editing) return
    const values = await editForm.validateFields()
    setLoading(true); setError('')
    try {
      await api.updateProject(editing.id, { ...values, name: values.name.trim(), is_fixed: values.is_fixed ?? false, category: values.category ?? 'study', include_in_summary: values.include_in_summary ?? true })
      setEditing(null); await loadProjects(); message.success('项目更新成功')
    } catch (err) { setError(errorMessage(err)) } finally { setLoading(false) }
  }

  function confirmDelete(project: Project) {
    Modal.confirm({
      title: `确认删除项目“${project.name}”吗？`,
      content: '删除项目会保留历史任务，但会移除任务的项目关联。',
      okText: '删除', cancelText: '取消', okButtonProps: { danger: true },
      onOk: async () => {
        try { await api.deleteProject(project.id); await loadProjects(); message.success('项目删除成功') }
        catch (err) { const text = errorMessage(err); setError(text); message.error(text) }
      },
    })
  }

  useEffect(() => { if (connected) loadProjects() }, [connected])

  const columns = [
    { title: '编号', dataIndex: 'id', key: 'id', width: 70 },
    { title: '名称', dataIndex: 'name', key: 'name', render: (value: string, project: Project) => <Space><Typography.Text strong>{value}</Typography.Text>{project.is_fixed && <Tag color="blue">固定项目</Tag>}{!project.include_in_summary && <Tag color="orange">不纳入总结</Tag>}</Space> },
    { title: '分类', dataIndex: 'category', key: 'category', render: (value: ProjectCategory) => <Tag>{categoryLabel(value)}</Tag> },
    { title: '描述', dataIndex: 'description', key: 'description', render: (value: string) => value || <Typography.Text type="secondary">暂无描述</Typography.Text> },
    { title: '更新时间', dataIndex: 'updated_at', key: 'updated_at', render: formatDate },
    { title: '操作', key: 'actions', render: (_: unknown, project: Project) => <Space><Button icon={<EditOutlined />} onClick={() => startEditing(project)}>编辑</Button><Button danger icon={<DeleteOutlined />} onClick={() => confirmDelete(project)}>删除</Button></Space> },
  ]

  return (
    <div className="page-stack">
      <header className="section-header"><div><Typography.Title level={2}>项目管理</Typography.Title><Typography.Text type="secondary">管理每日任务所属的项目。</Typography.Text></div></header>
      <div className="projects-grid">
        <Card title="创建项目">
          {error && <Alert className="card-alert" type="error" showIcon title={error} />}
          <ProjectForm form={createForm} onFinish={createProject} loading={loading} submitLabel="创建" />
        </Card>
        <Card title="项目列表" extra={<Button icon={<ReloadOutlined />} onClick={loadProjects} loading={loading}>刷新</Button>}>
          <Alert className="card-alert" type="info" showIcon title="删除项目会保留历史任务，但会移除任务的项目关联。" />
          <Table rowKey="id" loading={loading} dataSource={projects} columns={columns} pagination={false} />
        </Card>
      </div>
      <Modal title={`编辑项目 #${editing?.id ?? ''}`} open={editing !== null} onOk={updateProject} okText="保存" cancelText="取消" confirmLoading={loading} onCancel={() => setEditing(null)}>
        <ProjectForm form={editForm} loading={loading} hideSubmit />
      </Modal>
    </div>
  )
}

function ProjectForm({ form, onFinish, loading, submitLabel, hideSubmit }: { form: ReturnType<typeof Form.useForm<ProjectInput>>[0]; onFinish?: (values: ProjectInput) => void; loading: boolean; submitLabel?: string; hideSubmit?: boolean }) {
  return <Form form={form} layout="vertical" initialValues={{ is_fixed: false, category: 'study', include_in_summary: true }} onFinish={onFinish}>
    <Form.Item name="name" label="名称" rules={[{ required: true, whitespace: true }]}><Input placeholder="例如：Go 后端学习" /></Form.Item>
    <Form.Item name="description" label="描述"><Input.TextArea rows={3} placeholder="填写项目说明" /></Form.Item>
    <Form.Item name="category" label="分类"><Select options={categoryOptions} /></Form.Item>
    <Form.Item name="include_in_summary" valuePropName="checked"><Checkbox>纳入学习总结</Checkbox></Form.Item>
    <Form.Item name="is_fixed" valuePropName="checked"><Checkbox>固定项目</Checkbox></Form.Item>
    {!hideSubmit && <Button type="primary" htmlType="submit" loading={loading}> {submitLabel} </Button>}
  </Form>
}

const categoryOptions: { value: ProjectCategory; label: string }[] = [
  { value: 'study', label: '学习' },
  { value: 'project', label: '项目推进' },
  { value: 'life', label: '生活' },
  { value: 'break', label: '休息' },
  { value: 'other', label: '其他' },
]

function categoryLabel(value: ProjectCategory) {
  return categoryOptions.find((option) => option.value === value)?.label ?? value
}

function formatDate(value: string) { const date = new Date(value); return Number.isNaN(date.getTime()) ? value : date.toLocaleString('zh-CN') }
