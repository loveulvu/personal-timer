import { ReloadOutlined } from '@ant-design/icons'
import { Alert, Button, Card, Empty, Select, Space, Tag, Typography, message } from 'antd'
import { useEffect, useState } from 'react'
import { api, MemoryEvidenceItem, MemoryListItem, MemoryListStatusFilter } from './api'
import { errorMessage } from './utils/labels'

type Props = { connected: boolean }
type FeedbackValue = 'correct' | 'wrong' | 'outdated' | 'too_broad'

export function MemoriesPage({ connected }: Props) {
  const [status, setStatus] = useState<MemoryListStatusFilter>('active')
  const [memories, setMemories] = useState<MemoryListItem[]>([])
  const [loading, setLoading] = useState(false)
  const [feedbackLoading, setFeedbackLoading] = useState<Record<number, FeedbackValue | undefined>>({})
  const [expandedMemoryId, setExpandedMemoryId] = useState<number | null>(null)
  const [evidenceByMemory, setEvidenceByMemory] = useState<Record<number, MemoryEvidenceItem[]>>({})
  const [evidenceLoading, setEvidenceLoading] = useState<Record<number, boolean>>({})
  const [evidenceError, setEvidenceError] = useState<Record<number, string>>({})
  const [error, setError] = useState('')
  const [feedbackError, setFeedbackError] = useState('')

  async function loadMemories(selectedStatus = status) {
    if (!connected) return
    setLoading(true)
    setError('')
    try {
      setMemories(await api.listMemories(selectedStatus, '', 50))
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }

  async function submitMemoryFeedback(memory: MemoryListItem, feedbackValue: FeedbackValue) {
    setFeedbackLoading((current) => ({ ...current, [memory.id]: feedbackValue }))
    setFeedbackError('')
    try {
      await api.submitFeedback({
        target_type: 'memory',
        target_id: memory.id,
        feedback_value: feedbackValue,
      })
      message.success('反馈已保存')
      await loadMemories(status)
    } catch {
      const text = '反馈提交失败'
      setFeedbackError(text)
      message.error(text)
    } finally {
      setFeedbackLoading((current) => ({ ...current, [memory.id]: undefined }))
    }
  }

  async function toggleEvidence(memory: MemoryListItem) {
    if (expandedMemoryId === memory.id) {
      setExpandedMemoryId(null)
      return
    }
    setExpandedMemoryId(memory.id)
    if (evidenceByMemory[memory.id]) return
    setEvidenceLoading((current) => ({ ...current, [memory.id]: true }))
    setEvidenceError((current) => ({ ...current, [memory.id]: '' }))
    try {
      const items = await api.listMemoryEvidence(memory.id)
      setEvidenceByMemory((current) => ({ ...current, [memory.id]: items }))
    } catch {
      setEvidenceError((current) => ({ ...current, [memory.id]: '证据加载失败' }))
    } finally {
      setEvidenceLoading((current) => ({ ...current, [memory.id]: false }))
    }
  }

  useEffect(() => {
    if (connected) loadMemories(status)
  }, [connected, status])

  return (
    <div className="page-stack">
      <header className="section-header">
        <div>
          <Typography.Title level={2}>记忆管理</Typography.Title>
          <Typography.Text type="secondary">查看系统沉淀的长期学习规律，并反馈是否仍然准确。</Typography.Text>
        </div>
        <Space>
          <Select
            value={status}
            onChange={setStatus}
            style={{ width: 132 }}
            options={[
              { value: 'active', label: 'Active' },
              { value: 'archived', label: 'Archived' },
              { value: 'all', label: 'All' },
            ]}
          />
          <Button icon={<ReloadOutlined />} onClick={() => loadMemories(status)} loading={loading} disabled={!connected}>
            刷新
          </Button>
        </Space>
      </header>

      {error && <Alert type="error" showIcon title={error} />}
      {feedbackError && <Alert type="error" showIcon title={feedbackError} />}

      <Space direction="vertical" size="middle" style={{ width: '100%' }}>
        {memories.length === 0 && !loading && <Card><Empty description="暂无记忆" /></Card>}
        {memories.map((memory) => (
          <Card
            key={memory.id}
            loading={loading}
            title={<MemoryTitle memory={memory} />}
            extra={<Tag color={memory.status === 'active' ? 'green' : 'default'}>{memory.status}</Tag>}
          >
            <Space direction="vertical" size="small" style={{ width: '100%' }}>
              <Typography.Paragraph style={{ marginBottom: 4 }}>{memory.content}</Typography.Paragraph>
              <Space wrap>
                <Tag>{memory.memory_type}</Tag>
                <Tag>{memory.scope_type}</Tag>
                {memory.project_name && <Tag color="blue">{memory.project_name}</Tag>}
                <Typography.Text type="secondary">置信度 {formatPercent(memory.confidence)}</Typography.Text>
                <Typography.Text type="secondary">支持 {memory.support_count}</Typography.Text>
                <Typography.Text type="secondary">冲突 {memory.contradiction_count}</Typography.Text>
                <Typography.Text type="secondary">最近出现 {formatDate(memory.last_seen_at)}</Typography.Text>
              </Space>
              <Space wrap>
                <Typography.Text type="secondary">反馈：</Typography.Text>
                <Button size="small" loading={feedbackLoading[memory.id] === 'correct'} onClick={() => submitMemoryFeedback(memory, 'correct')}>正确</Button>
                <Button size="small" loading={feedbackLoading[memory.id] === 'wrong'} onClick={() => submitMemoryFeedback(memory, 'wrong')} danger>错误</Button>
                <Button size="small" loading={feedbackLoading[memory.id] === 'outdated'} onClick={() => submitMemoryFeedback(memory, 'outdated')}>已过期</Button>
                <Button size="small" loading={feedbackLoading[memory.id] === 'too_broad'} onClick={() => submitMemoryFeedback(memory, 'too_broad')}>太宽泛</Button>
                <Button size="small" onClick={() => toggleEvidence(memory)}>{expandedMemoryId === memory.id ? '收起证据' : '查看证据'}</Button>
              </Space>
              {expandedMemoryId === memory.id && (
                <EvidenceList
                  items={evidenceByMemory[memory.id] ?? []}
                  loading={evidenceLoading[memory.id]}
                  error={evidenceError[memory.id]}
                />
              )}
            </Space>
          </Card>
        ))}
      </Space>
    </div>
  )
}

function MemoryTitle({ memory }: { memory: MemoryListItem }) {
  return (
    <Space wrap>
      <Typography.Text strong>{memory.title}</Typography.Text>
      <Typography.Text type="secondary">#{memory.id}</Typography.Text>
    </Space>
  )
}

function formatPercent(value: number) {
  return `${Math.round(value * 100)}%`
}

function formatDate(value: string) {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString('zh-CN')
}

function EvidenceList({ items, loading, error }: { items: MemoryEvidenceItem[]; loading?: boolean; error?: string }) {
  if (loading) return <Typography.Text type="secondary">加载证据中...</Typography.Text>
  if (error) return <Alert type="error" showIcon title={error} />
  return (
    <Card size="small" title="支撑证据">
      {items.length === 0 && <Empty description="暂无证据记录" />}
      <Space direction="vertical" size="small" style={{ width: '100%' }}>
        {items.map((item) => (
          <div key={item.id}>
            <Space wrap>
              <Typography.Text>日期：{item.evidence_date}</Typography.Text>
              <Typography.Text>来源：{item.source_type}{item.source_id ? ` #${item.source_id}` : ''}</Typography.Text>
              <Typography.Text>权重：{item.weight}</Typography.Text>
              <Typography.Text type="secondary">创建：{formatDate(item.created_at)}</Typography.Text>
            </Space>
            <Typography.Paragraph style={{ margin: '6px 0 0' }}>摘录：{item.excerpt || '无'}</Typography.Paragraph>
          </div>
        ))}
      </Space>
    </Card>
  )
}
