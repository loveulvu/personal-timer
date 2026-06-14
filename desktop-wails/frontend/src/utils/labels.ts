const valueLabels: Record<string, string> = {
  planned: '计划中',
  running: '进行中',
  paused: '已暂停',
  completed: '已完成',
  cancelled: '已取消',
  daily: '每日',
  weekly: '每周',
  ok: '正常',
  error: '异常',
}

const errorMappings: Array<[string, string]> = [
  ['backend is not running. please start backend-go first.', '后端服务未启动，请先启动 backend-go'],
  ['summary already exists', '总结已存在'],
  ['llm is not configured', 'LLM 未配置，请检查 .env'],
  ['llm returned empty content', 'LLM 返回内容为空'],
  ['failed to fetch', '请求失败，请检查后端服务是否启动'],
  ['backend request timed out', '请求超时，请稍后重试'],
  ['finish_note is required', '完成备注不能为空'],
  ['finish_description is required', '完成描述不能为空'],
  ['task status must be completed', '只能编辑或删除已完成任务'],
  ['actual_minutes_override must be greater than or equal to 0', '实际时长不能小于 0'],
  ['actual_minutes_override and clear_actual_minutes_override cannot both be set', '实际时长设置与清空操作不能同时提交'],
  ['unknown error', '未知错误'],
]

export function valueLabel(value: string) {
  return valueLabels[value] ?? value
}

export function errorMessage(err: unknown) {
  const text = err instanceof Error ? err.message : typeof err === 'string' ? err : '未知错误'
  const lowerText = text.toLowerCase()
  const mapping = errorMappings.find(([source]) => lowerText.includes(source))
  return mapping?.[1] ?? text
}

export function timerActionLabel(action: 'start' | 'pause' | 'resume' | 'finish') {
  return {
    start: '开始',
    pause: '暂停',
    resume: '继续',
    finish: '完成',
  }[action]
}
