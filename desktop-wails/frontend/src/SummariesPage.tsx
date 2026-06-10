import { DeleteOutlined, EyeOutlined, ExperimentOutlined } from '@ant-design/icons'
import { Alert, Button, Card, DatePicker, Descriptions, Modal, Select, Space, Table, Typography, message } from 'antd'
import dayjs from 'dayjs'
import { useEffect, useState } from 'react'
import { api, GenerateSummaryResult, Summary } from './api'
import { StatusTag } from './components/StatusTag'

type Props = { connected: boolean }
type SummaryFilter = '' | 'daily' | 'weekly'

export function SummariesPage({ connected }: Props) {
  const today = dayjs()
  const [dailyDate, setDailyDate] = useState(today)
  const [weeklyRange, setWeeklyRange] = useState<[dayjs.Dayjs, dayjs.Dayjs]>([today.subtract(6, 'day'), today])
  const [filter, setFilter] = useState<SummaryFilter>('')
  const [summaries, setSummaries] = useState<Summary[]>([])
  const [detail, setDetail] = useState<Summary | null>(null)
  const [generated, setGenerated] = useState<GenerateSummaryResult | null>(null)
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
    try { const result = await api.testLLM(); message.success(result.message || 'LLM connection works') }
    catch (err) { const text = errorMessage(err); setError(text); message.error(text) } finally { setLoading(false) }
  }

  async function generate(action: () => Promise<GenerateSummaryResult>) {
    setLoading(true); setError(''); setGenerated(null)
    try { setGenerated(await action()); await loadSummaries(filter); message.success('Summary generated successfully') }
    catch (err) { const text = errorMessage(err); setError(text); message.error(text) } finally { setLoading(false) }
  }

  async function viewSummary(id: number) {
    setLoading(true); setError('')
    try { setDetail(await api.getSummary(id)) } catch (err) { setError(errorMessage(err)) } finally { setLoading(false) }
  }

  function confirmDelete(summary: Summary) {
    Modal.confirm({
      title: `Delete ${summary.summary_type} summary #${summary.id}?`,
      okText: 'Delete', okButtonProps: { danger: true },
      onOk: async () => {
        try { await api.deleteSummary(summary.id); if (detail?.id === summary.id) setDetail(null); await loadSummaries(filter); message.success('Summary deleted') }
        catch (err) { const text = errorMessage(err); setError(text); message.error(text) }
      },
    })
  }

  useEffect(() => { if (connected) loadSummaries(filter) }, [connected, filter])

  const columns = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 70 },
    { title: 'Type', dataIndex: 'summary_type', key: 'summary_type', render: (value: string) => <StatusTag value={value} /> },
    { title: 'Range', key: 'range', render: (_: unknown, summary: Summary) => `${summary.start_date} to ${summary.end_date}` },
    { title: 'Created', dataIndex: 'created_at', key: 'created_at', render: formatDate },
    { title: 'Content preview', dataIndex: 'content', key: 'content', ellipsis: true, render: preview },
    { title: 'Actions', key: 'actions', render: (_: unknown, summary: Summary) => <Space><Button icon={<EyeOutlined />} onClick={() => viewSummary(summary.id)}>View</Button><Button danger icon={<DeleteOutlined />} onClick={() => confirmDelete(summary)}>Delete</Button></Space> },
  ]

  return <div className="page-stack">
    <header className="section-header"><div><Typography.Title level={2}>Summaries</Typography.Title><Typography.Text type="secondary">Generate and review daily or weekly study reflections.</Typography.Text></div></header>
    {error && <Alert type="error" showIcon title={error} />}
    <div className="summary-tools">
      <Card title="LLM Test"><Typography.Paragraph type="secondary">Tests the configured LLM connection without exposing the API key.</Typography.Paragraph><Button icon={<ExperimentOutlined />} onClick={testLLM} loading={loading} disabled={!connected}>Test LLM</Button></Card>
      <Card title="Generate Daily Summary"><Space orientation="vertical"><DatePicker value={dailyDate} onChange={(value) => value && setDailyDate(value)} /><Button type="primary" onClick={() => generate(() => api.generateDailySummary(dailyDate.format('YYYY-MM-DD')))} loading={loading} disabled={!connected}>Generate Daily Summary</Button></Space></Card>
      <Card title="Generate Weekly Summary"><Space orientation="vertical"><DatePicker.RangePicker value={weeklyRange} onChange={(value) => value?.[0] && value?.[1] && setWeeklyRange([value[0], value[1]])} /><Button type="primary" onClick={() => generate(() => api.generateWeeklySummary(weeklyRange[0].format('YYYY-MM-DD'), weeklyRange[1].format('YYYY-MM-DD')))} loading={loading} disabled={!connected}>Generate Weekly Summary</Button></Space></Card>
    </div>

    {generated && <Card title={`Generated Summary #${generated.summary_id}`}><Typography.Paragraph className="content-block">{generated.content}</Typography.Paragraph></Card>}

    <Card title="Summaries" extra={<Select value={filter} onChange={setFilter} style={{ width: 120 }} options={[{ value: '', label: 'all' }, { value: 'daily', label: 'daily' }, { value: 'weekly', label: 'weekly' }]} />}>
      <Table rowKey="id" loading={loading} dataSource={summaries} columns={columns} pagination={{ pageSize: 8 }} />
    </Card>

    <Modal title={`Summary Detail #${detail?.id ?? ''}`} open={detail !== null} onCancel={() => setDetail(null)} footer={<Button onClick={() => setDetail(null)}>Close</Button>} width={760}>
      {detail && <>
        <Descriptions size="small" column={2} items={[
          { key: 'type', label: 'Type', children: <StatusTag value={detail.summary_type} /> },
          { key: 'created', label: 'Created', children: formatDate(detail.created_at) },
          { key: 'start', label: 'Start date', children: detail.start_date },
          { key: 'end', label: 'End date', children: detail.end_date },
        ]} />
        <Typography.Title level={5}>Content</Typography.Title>
        <pre className="content-block">{detail.content}</pre>
        {detail.source_data !== undefined && <><Typography.Title level={5}>Source Data</Typography.Title><pre className="content-block">{formatSourceData(detail.source_data)}</pre></>}
      </>}
    </Modal>
  </div>
}

function preview(content: string) { return content.length > 140 ? `${content.slice(0, 140)}...` : content }
function formatSourceData(sourceData: unknown) { if (typeof sourceData === 'string') return sourceData; try { return JSON.stringify(sourceData, null, 2) } catch { return String(sourceData) } }
function formatDate(value: string) { const date = new Date(value); return Number.isNaN(date.getTime()) ? value : date.toLocaleString() }
function errorMessage(err: unknown) { return err instanceof Error ? err.message : typeof err === 'string' ? err : 'Unknown error' }
