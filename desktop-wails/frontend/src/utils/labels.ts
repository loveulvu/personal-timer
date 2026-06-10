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
