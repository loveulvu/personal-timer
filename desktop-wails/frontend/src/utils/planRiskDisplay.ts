import type { PlanRiskLevel } from '../api'

export function planRiskTitle(riskLevel: PlanRiskLevel) {
  switch (riskLevel) {
    case 'insufficient_data':
      return '最近可用学习记录不足，暂时无法可靠判断今日计划风险。'
    case 'low':
      return '今日计划负载基本合理。'
    case 'medium':
      return '今日计划略高于近期平均水平，建议预留缓冲时间。'
    case 'high':
      return '今日计划明显高于近期平均水平，存在较高完不成风险。'
  }
}

export function planRiskAlertType(riskLevel: PlanRiskLevel): 'info' | 'success' | 'warning' | 'error' {
  if (riskLevel === 'high') return 'error'
  if (riskLevel === 'medium') return 'warning'
  if (riskLevel === 'low') return 'success'
  return 'info'
}
