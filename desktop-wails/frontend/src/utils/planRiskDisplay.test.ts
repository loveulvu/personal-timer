import assert from 'node:assert/strict'
import { test } from 'node:test'
import { planRiskAlertType, planRiskTitle } from './planRiskDisplay.ts'

test('plan risk display supports all risk levels', () => {
  assert.equal(planRiskAlertType('insufficient_data'), 'info')
  assert.equal(planRiskAlertType('low'), 'success')
  assert.equal(planRiskAlertType('medium'), 'warning')
  assert.equal(planRiskAlertType('high'), 'error')
  assert.match(planRiskTitle('insufficient_data'), /不足/)
  assert.match(planRiskTitle('low'), /合理/)
  assert.match(planRiskTitle('medium'), /缓冲/)
  assert.match(planRiskTitle('high'), /风险/)
})
