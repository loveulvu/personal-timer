import { DeleteOutlined, EditOutlined, ReloadOutlined } from '@ant-design/icons'
import { Alert, Button, Card, Checkbox, Form, Input, Modal, Space, Table, Tag, Typography, message } from 'antd'
import { useEffect, useState } from 'react'
import { api, Project, ProjectInput } from './api'

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
      await api.createProject({ ...values, name: values.name.trim(), is_fixed: values.is_fixed ?? false })
      createForm.resetFields(); await loadProjects(); message.success('Project created')
    } catch (err) { setError(errorMessage(err)) } finally { setLoading(false) }
  }

  function startEditing(project: Project) {
    setEditing(project)
    editForm.setFieldsValue({ name: project.name, description: project.description, is_fixed: project.is_fixed })
  }

  async function updateProject() {
    if (!editing) return
    const values = await editForm.validateFields()
    setLoading(true); setError('')
    try {
      await api.updateProject(editing.id, { ...values, name: values.name.trim(), is_fixed: values.is_fixed ?? false })
      setEditing(null); await loadProjects(); message.success('Project updated')
    } catch (err) { setError(errorMessage(err)) } finally { setLoading(false) }
  }

  function confirmDelete(project: Project) {
    Modal.confirm({
      title: `Delete "${project.name}"?`,
      content: 'Existing tasks will be kept but their project link will be removed.',
      okText: 'Delete', okButtonProps: { danger: true },
      onOk: async () => {
        try { await api.deleteProject(project.id); await loadProjects(); message.success('Project deleted') }
        catch (err) { const text = errorMessage(err); setError(text); message.error(text) }
      },
    })
  }

  useEffect(() => { if (connected) loadProjects() }, [connected])

  const columns = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 70 },
    { title: 'Name', dataIndex: 'name', key: 'name', render: (value: string, project: Project) => <Space><Typography.Text strong>{value}</Typography.Text>{project.is_fixed && <Tag color="blue">fixed</Tag>}</Space> },
    { title: 'Description', dataIndex: 'description', key: 'description', render: (value: string) => value || <Typography.Text type="secondary">No description</Typography.Text> },
    { title: 'Updated', dataIndex: 'updated_at', key: 'updated_at', render: formatDate },
    { title: 'Actions', key: 'actions', render: (_: unknown, project: Project) => <Space><Button icon={<EditOutlined />} onClick={() => startEditing(project)}>Edit</Button><Button danger icon={<DeleteOutlined />} onClick={() => confirmDelete(project)}>Delete</Button></Space> },
  ]

  return (
    <div className="projects-grid">
      <Card title="Create project">
        {error && <Alert className="card-alert" type="error" showIcon title={error} />}
        <ProjectForm form={createForm} onFinish={createProject} loading={loading} submitLabel="Create" />
      </Card>
      <Card title="Projects" extra={<Button icon={<ReloadOutlined />} onClick={loadProjects} loading={loading}>Refresh</Button>}>
        <Alert className="card-alert" type="info" showIcon title="Deleting a project keeps existing tasks but removes their project link." />
        <Table rowKey="id" loading={loading} dataSource={projects} columns={columns} pagination={false} />
      </Card>
      <Modal title={`Edit project #${editing?.id ?? ''}`} open={editing !== null} onOk={updateProject} confirmLoading={loading} onCancel={() => setEditing(null)}>
        <ProjectForm form={editForm} loading={loading} hideSubmit />
      </Modal>
    </div>
  )
}

function ProjectForm({ form, onFinish, loading, submitLabel, hideSubmit }: { form: ReturnType<typeof Form.useForm<ProjectInput>>[0]; onFinish?: (values: ProjectInput) => void; loading: boolean; submitLabel?: string; hideSubmit?: boolean }) {
  return <Form form={form} layout="vertical" initialValues={{ is_fixed: false }} onFinish={onFinish}>
    <Form.Item name="name" label="Name" rules={[{ required: true, whitespace: true }]}><Input placeholder="Go backend" /></Form.Item>
    <Form.Item name="description" label="Description"><Input.TextArea rows={3} placeholder="Go backend learning" /></Form.Item>
    <Form.Item name="is_fixed" valuePropName="checked"><Checkbox>Fixed project</Checkbox></Form.Item>
    {!hideSubmit && <Button type="primary" htmlType="submit" loading={loading}> {submitLabel} </Button>}
  </Form>
}

function formatDate(value: string) { const date = new Date(value); return Number.isNaN(date.getTime()) ? value : date.toLocaleString() }
function errorMessage(err: unknown) { return err instanceof Error ? err.message : typeof err === 'string' ? err : 'Unknown error' }
