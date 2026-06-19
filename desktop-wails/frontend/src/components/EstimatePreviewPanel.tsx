import { Alert, Button, Form, Space, Typography } from 'antd'
import type { FormInstance } from 'antd'
import { useEffect, useState } from 'react'
import { api } from '../api'
import type { EstimatePreviewResponse, EstimatePreviewRiskLevel } from '../api'

type EstimatePreviewPanelProps = {
  form: FormInstance
  connected: boolean
}

export function EstimatePreviewPanel({ form, connected }: EstimatePreviewPanelProps) {
  const projectId = Form.useWatch('projectId', form)
  const title = Form.useWatch('title', form)
  const estimatedMinutes = Form.useWatch('estimatedMinutes', form)
  const [estimatePreviewLoading, setEstimatePreviewLoading] = useState(false)
  const [estimatePreviewError, setEstimatePreviewError] = useState('')
  const [estimatePreviewResult, setEstimatePreviewResult] = useState<EstimatePreviewResponse | null>(null)

  useEffect(() => {
    setEstimatePreviewResult(null)
    setEstimatePreviewError('')
  }, [projectId, estimatedMinutes])

  async function preview() {
    if (!projectId || !estimatedMinutes || estimatedMinutes <= 0) return
    setEstimatePreviewLoading(true)
    setEstimatePreviewError('')
    try {
      setEstimatePreviewResult(await api.estimateTaskPreview({
        project_id: projectId,
        title: typeof title === 'string' ? title.trim() : '',
        estimated_minutes: estimatedMinutes,
      }))
    } catch {
      setEstimatePreviewResult(null)
      setEstimatePreviewError('估时预览失败，请稍后重试。')
    } finally {
      setEstimatePreviewLoading(false)
    }
  }

  function applySuggestedMinutes() {
    if (!estimatePreviewResult) return
    form.setFieldValue('estimatedMinutes', estimatePreviewResult.suggested_minutes)
  }

  return (
    <Space direction="vertical" size={8} style={{ width: '100%' }}>
      <Button
        onClick={preview}
        loading={estimatePreviewLoading}
        disabled={!connected || !projectId || !estimatedMinutes || estimatedMinutes <= 0}
      >
        估时预览
      </Button>
      {estimatePreviewError && <Alert type="error" showIcon message={estimatePreviewError} />}
      {estimatePreviewResult && (
        <Alert
          type={alertType(estimatePreviewResult.risk_level)}
          showIcon
          message={riskMessage(estimatePreviewResult.risk_level)}
          description={
            <Space direction="vertical" size={4}>
              <Typography.Text>{estimatePreviewResult.reason}</Typography.Text>
              <Typography.Text type="secondary">
                样本 {estimatePreviewResult.sample_count} 条 · 平均实际 {estimatePreviewResult.avg_actual_minutes} 分钟 · 建议 {estimatePreviewResult.suggested_minutes} 分钟
              </Typography.Text>
              {estimatePreviewResult.split_recommended && (
                <Typography.Text type="warning">该类任务历史平均耗时较长，建议拆分为多个子任务。</Typography.Text>
              )}
              {estimatePreviewResult.risk_level !== 'insufficient_data' && (
                <Button size="small" onClick={applySuggestedMinutes}>
                  采用建议估时
                </Button>
              )}
            </Space>
          }
        />
      )}
    </Space>
  )
}

function riskMessage(riskLevel: EstimatePreviewRiskLevel) {
  switch (riskLevel) {
    case 'insufficient_data':
      return '历史样本不足，暂时无法可靠判断估时偏差。'
    case 'low':
      return '当前估时基本合理。'
    case 'medium':
      return '当前估时可能偏低，建议参考历史平均实际耗时。'
    case 'high':
      return '当前估时明显偏低，建议提高估时或拆分任务。'
  }
}

function alertType(riskLevel: EstimatePreviewRiskLevel): 'info' | 'success' | 'warning' | 'error' {
  if (riskLevel === 'high') return 'error'
  if (riskLevel === 'medium') return 'warning'
  if (riskLevel === 'low') return 'success'
  return 'info'
}
