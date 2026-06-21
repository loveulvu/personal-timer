import { DeleteOutlined, EyeOutlined, ExperimentOutlined } from '@ant-design/icons'
import { Alert, Button, Card, DatePicker, Descriptions, Modal, Select, Space, Table, Tag, Typography, message } from 'antd'
import dayjs from 'dayjs'
import { useEffect, useState } from 'react'
import { api, GenerateSummaryResult, Summary, SummaryActionItem } from './api'
import { StatusTag } from './components/StatusTag'
import { errorMessage, valueLabel } from './utils/labels'

type Props = { connected: boolean }
type SummaryFilter = '' | 'daily' | 'weekly'
type AcceptStatus = { loading?: boolean; text?: string; error?: string }
type FeedbackStatus = { loading?: boolean; text?: string; error?: string }

export function SummariesPage({ connected }: Props) {
  const today = dayjs()
  const [dailyDate, setDailyDate] = useState(today)
  const [weeklyRange, setWeeklyRange] = useState<[dayjs.Dayjs, dayjs.Dayjs]>([today.subtract(6, 'day'), today])
  const [filter, setFilter] = useState<SummaryFilter>('')
  const [summaries, setSummaries] = useState<Summary[]>([])
  const [detail, setDetail] = useState<Summary | null>(null)
  const [generated, setGenerated] = useState<GenerateSummaryResult | null>(null)
  const [accepted, setAccepted] = useState<Record<string, AcceptStatus>>({})
  const [feedbackStatus, setFeedbackStatus] = useState<Record<string, FeedbackStatus>>({})
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  async function loadSummaries(selectedFilter = filter) {
    if (!connected) return
    setLoading(true); setError('')
    try { setSummaries(await api.getSummaries(selectedFilter)) }
    catch (err) { setError(errorMessage(err)) } finally { setLoading(false) }
  }

  async function testLLM() {
    setLoading(true); setError('')
    try { await api.testLLM(); message.success('LLM 连接正常') }
    catch (err) { const text = errorMessage(err); setError(text); message.error(text) } finally { setLoading(false) }
  }

  async function generate(action: () => Promise<GenerateSummaryResult>) {
    setLoading(true); setError(''); setGenerated(null)
    try { setGenerated(await action()); await loadSummaries(filter); message.success('总结生成成功') }
    catch (err) { const text = errorMessage(err); setError(text); message.error(text) } finally { setLoading(false) }
  }

  async function viewSummary(id: number) {
    setLoading(true); setError('')
    try { setDetail(await api.getSummary(id)) } catch (err) { setError(errorMessage(err)) } finally { setLoading(false) }
  }

  function confirmDelete(summary: Summary) {
    Modal.confirm({
      title: `确认删除${valueLabel(summary.summary_type)}总结 #${summary.id} 吗？`,
      okText: '删除', cancelText: '取消', okButtonProps: { danger: true },
      onOk: async () => {
        try { await api.deleteSummary(summary.id); if (detail?.id === summary.id) setDetail(null); await loadSummaries(filter); message.success('总结删除成功') }
        catch (err) { const text = errorMessage(err); setError(text); message.error(text) }
      },
    })
  }

  async function acceptActionItem(summaryId: number, itemIndex: number) {
    const key = acceptKey(summaryId, itemIndex)
    setAccepted((current) => ({ ...current, [key]: { loading: true } }))
    try {
      const result = await api.acceptSummaryActionItem(summaryId, itemIndex, dayjs().add(1, 'day').format('YYYY-MM-DD'))
      setAccepted((current) => ({ ...current, [key]: { text: result.already_exists ? '明日计划已存在' : '已采纳' } }))
    } catch (err) {
      setAccepted((current) => ({ ...current, [key]: { error: errorMessage(err) } }))
    }
  }

  async function submitFeedback(key: string, request: Parameters<typeof api.submitFeedback>[0], successText: string) {
    setFeedbackStatus((current) => ({ ...current, [key]: { loading: true } }))
    try {
      await api.submitFeedback(request)
      setFeedbackStatus((current) => ({ ...current, [key]: { text: successText } }))
      message.success('反馈已保存')
    } catch (err) {
      const text = errorMessage(err)
      setFeedbackStatus((current) => ({ ...current, [key]: { error: text } }))
      message.error(text)
    }
  }

  useEffect(() => { if (connected) loadSummaries(filter) }, [connected, filter])

  const columns = [
    { title: '编号', dataIndex: 'id', key: 'id', width: 70 },
    { title: '总结类型', dataIndex: 'summary_type', key: 'summary_type', render: (value: string) => <StatusTag value={value} /> },
    { title: '日期范围', key: 'range', render: (_: unknown, summary: Summary) => `${summary.start_date} 至 ${summary.end_date}` },
    { title: '创建时间', dataIndex: 'created_at', key: 'created_at', render: formatDate },
    { title: '内容预览', dataIndex: 'content', key: 'content', ellipsis: true, render: preview },
    { title: '操作', key: 'actions', render: (_: unknown, summary: Summary) => <Space><Button icon={<EyeOutlined />} onClick={() => viewSummary(summary.id)}>查看</Button><Button danger icon={<DeleteOutlined />} onClick={() => confirmDelete(summary)}>删除</Button></Space> },
  ]

  return <div className="page-stack">
    <header className="section-header"><div><Typography.Title level={2}>总结</Typography.Title><Typography.Text type="secondary">生成并查看每日或每周学习总结。</Typography.Text></div></header>
    {error && <Alert type="error" showIcon title={error} />}
    <div className="summary-tools">
      <Card title="LLM 测试"><Typography.Paragraph type="secondary">测试已配置的 LLM 连接，不会显示 API 密钥。</Typography.Paragraph><Button icon={<ExperimentOutlined />} onClick={testLLM} loading={loading} disabled={!connected}>测试 LLM</Button></Card>
      <Card title="生成每日总结"><Space orientation="vertical"><DatePicker value={dailyDate} onChange={(value) => value && setDailyDate(value)} /><Button type="primary" onClick={() => generate(() => api.generateDailySummary(dailyDate.format('YYYY-MM-DD')))} loading={loading} disabled={!connected}>生成每日总结</Button></Space></Card>
      <Card title="生成每周总结"><Space orientation="vertical"><DatePicker.RangePicker value={weeklyRange} onChange={(value) => value?.[0] && value?.[1] && setWeeklyRange([value[0], value[1]])} /><Button type="primary" onClick={() => generate(() => api.generateWeeklySummary(weeklyRange[0].format('YYYY-MM-DD'), weeklyRange[1].format('YYYY-MM-DD')))} loading={loading} disabled={!connected}>生成每周总结</Button></Space></Card>
    </div>

    {generated && <Card title={`已生成总结 #${generated.summary_id}`}><SummaryFeedback summaryId={generated.summary_id} feedbackStatus={feedbackStatus} onSubmit={submitFeedback} /><Typography.Paragraph className="content-block">{generated.content}</Typography.Paragraph><ActionItems summaryId={generated.summary_id} items={generated.action_items} accepted={accepted} feedbackStatus={feedbackStatus} onAccept={acceptActionItem} onFeedback={submitFeedback} /></Card>}

    <Card title="总结列表" extra={<Select value={filter} onChange={setFilter} style={{ width: 120 }} options={[{ value: '', label: '全部' }, { value: 'daily', label: '每日' }, { value: 'weekly', label: '每周' }]} />}>
      <Table rowKey="id" loading={loading} dataSource={summaries} columns={columns} pagination={{ pageSize: 8 }} />
    </Card>

    <Modal title={`总结详情 #${detail?.id ?? ''}`} open={detail !== null} onCancel={() => setDetail(null)} footer={<Button onClick={() => setDetail(null)}>关闭</Button>} width={760}>
      {detail && <>
        <Descriptions size="small" column={2} items={[
          { key: 'type', label: '总结类型', children: <StatusTag value={detail.summary_type} /> },
          { key: 'created', label: '创建时间', children: formatDate(detail.created_at) },
          { key: 'start', label: '开始日期', children: detail.start_date },
          { key: 'end', label: '结束日期', children: detail.end_date },
        ]} />
        <SummaryFeedback summaryId={detail.id} feedbackStatus={feedbackStatus} onSubmit={submitFeedback} />
        <Typography.Title level={5}>内容</Typography.Title>
        <pre className="content-block">{detail.content}</pre>
        <ActionItems summaryId={detail.id} items={detail.action_items} accepted={accepted} feedbackStatus={feedbackStatus} onAccept={acceptActionItem} onFeedback={submitFeedback} />
        {detail.source_data !== undefined && <><Typography.Title level={5}>源数据</Typography.Title><pre className="content-block">{formatSourceData(detail.source_data)}</pre></>}
      </>}
    </Modal>
  </div>
}

function SummaryFeedback({ summaryId, feedbackStatus, onSubmit }: { summaryId: number; feedbackStatus: Record<string, FeedbackStatus>; onSubmit: (key: string, request: Parameters<typeof api.submitFeedback>[0], successText: string) => void }) {
  const key = `summary:${summaryId}`
  const status = feedbackStatus[key]
  return <Space wrap style={{ marginBottom: 12 }}>
    <Typography.Text type="secondary">总结反馈：</Typography.Text>
    <Button size="small" loading={status?.loading} onClick={() => onSubmit(key, { target_type: 'summary', target_id: summaryId, feedback_value: 'accurate' }, '已反馈：准确')}>准确</Button>
    <Button size="small" loading={status?.loading} onClick={() => onSubmit(key, { target_type: 'summary', target_id: summaryId, feedback_value: 'partially_accurate' }, '已反馈：部分准确')}>部分准确</Button>
    <Button size="small" danger loading={status?.loading} onClick={() => onSubmit(key, { target_type: 'summary', target_id: summaryId, feedback_value: 'inaccurate' }, '已反馈：不准确')}>不准确</Button>
    {status?.text && <Tag color="green">{status.text}</Tag>}
    {status?.error && <Typography.Text type="danger">{status.error}</Typography.Text>}
  </Space>
}

function ActionItems({ summaryId, items, accepted, feedbackStatus, onAccept, onFeedback }: { summaryId: number; items?: SummaryActionItem[] | null; accepted: Record<string, AcceptStatus>; feedbackStatus: Record<string, FeedbackStatus>; onAccept: (summaryId: number, itemIndex: number) => void; onFeedback: (key: string, request: Parameters<typeof api.submitFeedback>[0], successText: string) => void }) {
  const actionItems = normalizeActionItems(items)
  if (actionItems.length === 0) return null
  return <div>
    <Typography.Title level={5}>行动建议</Typography.Title>
    <Space direction="vertical" size="middle" style={{ width: '100%' }}>
      {actionItems.map((item, index) => {
        const status = accepted[acceptKey(summaryId, index)]
        const feedbackKey = `action_item:${summaryId}:${index}`
        const feedback = feedbackStatus[feedbackKey]
        return <div key={`${item.type}-${item.title}-${index}`}>
        <Space wrap>
          <Tag>{priorityLabel(item.priority)}</Tag>
          <Tag>{typeLabel(item.type)}</Tag>
          <Typography.Text strong>{item.title}</Typography.Text>
          {isAcceptableActionItem(item) && <Button size="small" loading={status?.loading} disabled={Boolean(status?.text)} onClick={() => onAccept(summaryId, index)}>
            {status?.loading ? '采纳中...' : '采纳到明日计划'}
          </Button>}
          {status?.text && <Tag color="green">{status.text}</Tag>}
          <Button size="small" loading={feedback?.loading} onClick={() => onFeedback(feedbackKey, { target_type: 'action_item', target_id: summaryId, target_index: index, feedback_value: 'useful' }, '已反馈：有用')}>有用</Button>
          <Button size="small" loading={feedback?.loading} onClick={() => onFeedback(feedbackKey, { target_type: 'action_item', target_id: summaryId, target_index: index, feedback_value: 'not_useful' }, '已反馈：没用')}>没用</Button>
          {feedback?.text && <Tag color="green">{feedback.text}</Tag>}
        </Space>
        {item.reason && <Typography.Paragraph style={{ margin: '6px 0 0' }}>原因：{item.reason}</Typography.Paragraph>}
        {(item.suggested_project || item.suggested_minutes) && <Typography.Text type="secondary">
          建议：{[item.suggested_project, item.suggested_minutes ? `${item.suggested_minutes} 分钟` : ''].filter(Boolean).join('，')}
        </Typography.Text>}
        {status?.error && <Typography.Paragraph type="danger" style={{ margin: '6px 0 0' }}>{status.error}</Typography.Paragraph>}
        {feedback?.error && <Typography.Paragraph type="danger" style={{ margin: '6px 0 0' }}>{feedback.error}</Typography.Paragraph>}
      </div>
      })}
    </Space>
  </div>
}

function normalizeActionItems(items: unknown): SummaryActionItem[] {
  if (Array.isArray(items)) return items.filter((item): item is SummaryActionItem => Boolean(item && typeof item === 'object' && 'title' in item))
  if (typeof items === 'string') {
    try { return normalizeActionItems(JSON.parse(items)) } catch { return [] }
  }
  return []
}

function priorityLabel(priority: string) {
  return ({ high: '高优先级', medium: '中优先级', low: '低优先级' } as Record<string, string>)[priority] ?? priority
}

function typeLabel(type: string) {
  return ({ schedule: '安排任务', consistency: '保持连续性', estimation: '估时校准', split_task: '拆分任务', focus_topic: '聚焦复习', cleanup: '清理数据' } as Record<string, string>)[type] ?? type
}

function isAcceptableActionItem(item: SummaryActionItem) {
  return Boolean(item.title && item.suggested_project && item.suggested_minutes && item.suggested_minutes > 0 && item.type !== 'cleanup')
}

function acceptKey(summaryId: number, itemIndex: number) {
  return `${summaryId}:${itemIndex}`
}

function preview(content: string) { return content.length > 140 ? `${content.slice(0, 140)}...` : content }
function formatSourceData(sourceData: unknown) { if (typeof sourceData === 'string') return sourceData; try { return JSON.stringify(sourceData, null, 2) } catch { return String(sourceData) } }
function formatDate(value: string) { const date = new Date(value); return Number.isNaN(date.getTime()) ? value : date.toLocaleString('zh-CN') }
