import {
  CheckOutlined,
  CloseOutlined,
  PlayCircleOutlined,
  ReloadOutlined,
  SearchOutlined,
} from '@ant-design/icons'
import {
  Alert,
  Button,
  Card,
  Collapse,
  DatePicker,
  Descriptions,
  Empty,
  Input,
  InputNumber,
  List,
  Space,
  Spin,
  Table,
  Tag,
  Timeline,
  Typography,
} from 'antd'
import dayjs from 'dayjs'
import { useEffect, useMemo, useState } from 'react'
import type { ReactNode } from 'react'
import {
  AgentActionProposal,
  AgentContextPack,
  AgentRunListItem,
  AgentTrajectory,
  JsonValue,
  api,
} from './api'
import { errorMessage } from './utils/labels'

type AgentPageProps = {
  connected: boolean
}

const defaultGoal = '根据最近 5 天学习记录，帮我安排今天的 Go 后端复习计划'

export function AgentPage({ connected }: AgentPageProps) {
  const [goal, setGoal] = useState(defaultGoal)
  const [targetDate, setTargetDate] = useState(todayISO())
  const [recentDays, setRecentDays] = useState(5)
  const [contextPack, setContextPack] = useState<AgentContextPack | null>(null)
  const [runs, setRuns] = useState<AgentRunListItem[]>([])
  const [trajectory, setTrajectory] = useState<AgentTrajectory | null>(null)
  const [pendingProposals, setPendingProposals] = useState<AgentActionProposal[]>([])
  const [loading, setLoading] = useState('')
  const [error, setError] = useState('')

  const request = useMemo(
    () => ({ goal: goal.trim(), target_date: targetDate, recent_days: recentDays }),
    [goal, targetDate, recentDays],
  )

  useEffect(() => {
    if (connected) {
      refreshRuns()
    }
  }, [connected])

  async function previewContext() {
    await runWithState('preview', async () => {
      const result = await api.previewAgentContext(request)
      setContextPack(result.context_pack)
    })
  }

  async function createRun() {
    await runWithState('run', async () => {
      const result = await api.createAgentRun(request)
      setTrajectory({
        run: result.run,
        context_snapshot: null,
        steps: result.steps,
        proposals: result.proposals ?? [],
      })
      await refreshRuns()
      await loadTrajectory(result.run.id)
    })
  }

  async function refreshRuns() {
    await runWithState('runs', async () => {
      const [runItems, proposals] = await Promise.all([
        api.listAgentRuns('', 20),
        api.listAgentActionProposals('pending'),
      ])
      setRuns(runItems)
      setPendingProposals(proposals)
    })
  }

  async function loadTrajectory(id: number) {
    await runWithState('trajectory', async () => {
      setTrajectory(await api.getAgentTrajectory(id))
    })
  }

  async function decideProposal(id: number, action: 'accept' | 'reject') {
    await runWithState(`${action}-${id}`, async () => {
      if (action === 'accept') {
        await api.acceptAgentActionProposal(id)
      } else {
        await api.rejectAgentActionProposal(id)
      }
      await refreshRuns()
      if (trajectory) {
        await loadTrajectory(trajectory.run.id)
      }
    })
  }

  async function runWithState(name: string, fn: () => Promise<void>) {
    setError('')
    setLoading(name)
    try {
      await fn()
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setLoading('')
    }
  }

  const shownProposals = trajectory?.proposals?.length ? trajectory.proposals : pendingProposals

  return (
    <div className="page-stack">
      <div className="page-header">
        <div>
          <Typography.Title level={2}>Agent</Typography.Title>
          <Typography.Text type="secondary">Harness console</Typography.Text>
        </div>
        <Button icon={<ReloadOutlined />} onClick={refreshRuns} loading={loading === 'runs'} disabled={!connected}>
          Refresh Runs
        </Button>
      </div>

      {error && <Alert type="error" showIcon message={error} />}

      <Card size="small" title="Goal">
        <Space direction="vertical" size="middle" style={{ width: '100%' }}>
          <Input.TextArea
            rows={3}
            value={goal}
            onChange={(event) => setGoal(event.target.value)}
            placeholder="例如：根据最近 5 天学习记录，帮我安排今天的 Go 后端复习计划"
          />
          <Space wrap>
            <DatePicker
              value={dayjs(targetDate)}
              onChange={(value) => setTargetDate(value?.format('YYYY-MM-DD') ?? todayISO())}
              allowClear={false}
            />
            <InputNumber
              min={1}
              max={14}
              value={recentDays}
              onChange={(value) => setRecentDays(Number(value ?? 5))}
              addonBefore="Recent days"
            />
            <Button
              icon={<SearchOutlined />}
              onClick={previewContext}
              loading={loading === 'preview'}
              disabled={!connected || !goal.trim()}
            >
              Preview Context
            </Button>
            <Button
              type="primary"
              icon={<PlayCircleOutlined />}
              onClick={createRun}
              loading={loading === 'run'}
              disabled={!connected || !goal.trim()}
            >
              Run Agent
            </Button>
          </Space>
        </Space>
      </Card>

      <div className="agent-grid">
        <Card size="small" title="Context Preview">
          {loading === 'preview' && <Spin />}
          {!contextPack && loading !== 'preview' && <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} />}
          {contextPack && <ContextPreview pack={contextPack} />}
        </Card>

        <Card size="small" title="Recent Runs">
          <Table
            size="small"
            rowKey="id"
            dataSource={runs}
            pagination={false}
            onRow={(record) => ({ onClick: () => loadTrajectory(record.id) })}
            columns={[
              { title: 'ID', dataIndex: 'id', width: 72 },
              { title: 'Goal', dataIndex: 'user_goal', ellipsis: true },
              { title: 'Date', dataIndex: 'target_date', width: 112 },
              { title: 'Status', dataIndex: 'status', width: 150, render: statusTag },
              { title: 'Pending', dataIndex: 'pending_proposal_count', width: 92 },
            ]}
          />
        </Card>
      </div>

      <Card size="small" title="Trajectory Detail">
        {!trajectory && <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} />}
        {trajectory && (
          <Space direction="vertical" size="middle" style={{ width: '100%' }}>
            <RunSummary trajectory={trajectory} />
            <ContextSnapshot trajectory={trajectory} />
            <StepsTimeline trajectory={trajectory} />
          </Space>
        )}
      </Card>

      <Card size="small" title="Proposals">
        <ProposalList
          proposals={shownProposals}
          loading={loading}
          onAccept={(id) => decideProposal(id, 'accept')}
          onReject={(id) => decideProposal(id, 'reject')}
        />
      </Card>
    </div>
  )
}

function ContextPreview({ pack }: { pack: AgentContextPack }) {
  return (
    <Space direction="vertical" size="small" style={{ width: '100%' }}>
      <Descriptions size="small" column={2}>
        <Descriptions.Item label="Goal">{pack.user_goal}</Descriptions.Item>
        <Descriptions.Item label="Date">{pack.target_date}</Descriptions.Item>
        <Descriptions.Item label="Tasks">{pack.today_tasks.length}</Descriptions.Item>
        <Descriptions.Item label="Summaries">{pack.recent_summaries.length}</Descriptions.Item>
        <Descriptions.Item label="Memories">{pack.memories.length}</Descriptions.Item>
        <Descriptions.Item label="Action Items">{pack.recent_action_items.length}</Descriptions.Item>
      </Descriptions>
      {pack.today_tasks.length > 0 && (
        <List
          size="small"
          dataSource={pack.today_tasks}
          renderItem={(task) => (
            <List.Item>
              <Typography.Text>{task.title}</Typography.Text>
              <Typography.Text type="secondary">{task.status} · {task.estimated_minutes}m</Typography.Text>
            </List.Item>
          )}
        />
      )}
      {pack.plan_risk && <JsonCollapse title="Plan Risk" value={pack.plan_risk} />}
      <TagList title="Constraints" items={pack.constraints} />
      <TagList title="Omitted" items={pack.omitted_sections} />
    </Space>
  )
}

function RunSummary({ trajectory }: { trajectory: AgentTrajectory }) {
  const run = trajectory.run
  return (
    <Space direction="vertical" size="small" style={{ width: '100%' }}>
      <Descriptions size="small" column={3}>
        <Descriptions.Item label="Run">{run.id}</Descriptions.Item>
        <Descriptions.Item label="Status">{statusTag(run.status)}</Descriptions.Item>
        <Descriptions.Item label="Date">{run.target_date}</Descriptions.Item>
        <Descriptions.Item label="Created">{formatTime(run.created_at)}</Descriptions.Item>
        <Descriptions.Item label="Completed">{run.completed_at ? formatTime(run.completed_at) : '-'}</Descriptions.Item>
      </Descriptions>
      {run.final_answer && <Alert type="success" showIcon message="Final Answer" description={run.final_answer} />}
      {run.error_message && <Alert type="error" showIcon message="Error" description={run.error_message} />}
      {run.status === 'requires_confirmation' && (
        <Alert type="warning" showIcon message="Requires confirmation" />
      )}
    </Space>
  )
}

function ContextSnapshot({ trajectory }: { trajectory: AgentTrajectory }) {
  const snapshot = trajectory.context_snapshot
  if (!snapshot) {
    return <Alert type="info" showIcon message="No context snapshot" />
  }
  return (
    <Space direction="vertical" size="small" style={{ width: '100%' }}>
      <Descriptions size="small" column={3}>
        <Descriptions.Item label="Snapshot">{snapshot.id}</Descriptions.Item>
        <Descriptions.Item label="Token Estimate">{snapshot.token_estimate}</Descriptions.Item>
        <Descriptions.Item label="Created">{formatTime(snapshot.created_at)}</Descriptions.Item>
      </Descriptions>
      <TagList title="Omitted" items={snapshot.omitted_sections} />
      <JsonCollapse title="Context Pack" value={snapshot.context_pack as unknown as JsonValue} />
    </Space>
  )
}

function StepsTimeline({ trajectory }: { trajectory: AgentTrajectory }) {
  return (
    <Timeline
      items={trajectory.steps.map((step) => ({
        color: step.status === 'failed' ? 'red' : 'blue',
        children: (
          <Space direction="vertical" size={4} style={{ width: '100%' }}>
            <Space wrap>
              <Tag>{step.step_index}</Tag>
              <Typography.Text strong>{step.step_type}</Typography.Text>
              {statusTag(step.status)}
              {step.tool_name && <Tag>{step.tool_name}</Tag>}
            </Space>
            {step.thought_summary && <Typography.Text>{step.thought_summary}</Typography.Text>}
            {step.error_message && <Alert type="error" showIcon message={step.error_message} />}
            <Collapse
              size="small"
              items={[
                step.tool_input ? { key: 'input', label: 'tool_input_json', children: <JsonPre value={step.tool_input} /> } : null,
                step.tool_output ? { key: 'output', label: 'tool_output_json', children: <JsonPre value={step.tool_output} /> } : null,
              ].filter(Boolean) as { key: string; label: string; children: ReactNode }[]}
            />
          </Space>
        ),
      }))}
    />
  )
}

function ProposalList({
  proposals,
  loading,
  onAccept,
  onReject,
}: {
  proposals: AgentActionProposal[]
  loading: string
  onAccept: (id: number) => void
  onReject: (id: number) => void
}) {
  if (proposals.length === 0) {
    return <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} />
  }
  return (
    <List
      dataSource={proposals}
      renderItem={(proposal) => (
        <List.Item
          actions={[
            <Button
              key="accept"
              icon={<CheckOutlined />}
              onClick={() => onAccept(proposal.id)}
              loading={loading === `accept-${proposal.id}`}
              disabled={proposal.status !== 'pending'}
            >
              Accept
            </Button>,
            <Button
              key="reject"
              icon={<CloseOutlined />}
              onClick={() => onReject(proposal.id)}
              loading={loading === `reject-${proposal.id}`}
              disabled={proposal.status !== 'pending'}
            >
              Reject
            </Button>,
          ]}
        >
          <List.Item.Meta
            title={
              <Space wrap>
                <Typography.Text strong>#{proposal.id}</Typography.Text>
                <Typography.Text>{proposal.tool_name}</Typography.Text>
                <Tag>{proposal.action_type}</Tag>
                <Tag color={proposal.risk_level === 'write' ? 'orange' : undefined}>{proposal.risk_level}</Tag>
                {statusTag(proposal.status)}
              </Space>
            }
            description={
              <Space direction="vertical" size="small" style={{ width: '100%' }}>
                <JsonCollapse title="payload_json" value={proposal.payload ?? null} />
                {proposal.result && <JsonCollapse title="result_json" value={proposal.result} />}
                {proposal.error_message && <Alert type="error" showIcon message={proposal.error_message} />}
              </Space>
            }
          />
        </List.Item>
      )}
    />
  )
}

function JsonCollapse({ title, value }: { title: string; value: JsonValue }) {
  return <Collapse size="small" items={[{ key: title, label: title, children: <JsonPre value={value} /> }]} />
}

function JsonPre({ value }: { value: JsonValue }) {
  return <pre className="json-block">{JSON.stringify(value, null, 2)}</pre>
}

function TagList({ title, items }: { title: string; items: string[] }) {
  if (items.length === 0) {
    return null
  }
  return (
    <Space wrap>
      <Typography.Text type="secondary">{title}</Typography.Text>
      {items.map((item) => <Tag key={item}>{item}</Tag>)}
    </Space>
  )
}

function statusTag(value: string) {
  const color =
    value === 'completed' || value === 'executed'
      ? 'green'
      : value === 'failed'
        ? 'red'
        : value === 'requires_confirmation' || value === 'pending'
          ? 'orange'
          : value === 'rejected'
            ? 'default'
            : 'blue'
  return <Tag color={color}>{value}</Tag>
}

function todayISO() {
  return dayjs().format('YYYY-MM-DD')
}

function formatTime(value: string) {
  return value ? dayjs(value).format('YYYY-MM-DD HH:mm') : '-'
}
