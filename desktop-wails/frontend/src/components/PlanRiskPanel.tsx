import { Alert, Card, Space, Statistic, Tag, Typography } from 'antd'
import type { PlanRiskResponse } from '../api'
import { planRiskAlertType, planRiskTitle } from '../utils/planRiskDisplay'

type PlanRiskPanelProps = {
  risk: PlanRiskResponse | null
  loading: boolean
  error: string
}

export function PlanRiskPanel({ risk, loading, error }: PlanRiskPanelProps) {
  return (
    <Card title="今日计划风险" loading={loading}>
      {error && <Alert type="error" showIcon message={error} />}
      {!error && !risk && <Typography.Text type="secondary">暂无计划风险数据</Typography.Text>}
      {risk && (
        <Space direction="vertical" size={12} style={{ width: '100%' }}>
          <Alert
            type={planRiskAlertType(risk.risk_level)}
            showIcon
            message={<Space><span>{planRiskTitle(risk.risk_level)}</span><Tag>{risk.risk_level}</Tag></Space>}
            description={risk.reason}
          />
          <Space wrap>
            <Statistic title="今日计划" value={risk.planned_total_minutes} suffix="分钟" />
            <Statistic title="近期平均实际" value={risk.recent_avg_actual_minutes} suffix="分钟" />
            <Statistic title="近期有效天数" value={risk.recent_active_days} suffix="天" />
            <Statistic title="计划倍率" value={risk.plan_ratio} precision={2} suffix="x" />
          </Space>
          {risk.suggestions.length > 0 && (
            <Space direction="vertical" size={2}>
              {risk.suggestions.map((item) => <Typography.Text key={item}>• {item}</Typography.Text>)}
            </Space>
          )}
        </Space>
      )}
    </Card>
  )
}
