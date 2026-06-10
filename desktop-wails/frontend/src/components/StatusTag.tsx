import { Tag } from 'antd'
import { valueLabel } from '../utils/labels'

const colors: Record<string, string> = {
  planned: 'default',
  running: 'processing',
  paused: 'warning',
  completed: 'success',
  cancelled: 'error',
  daily: 'blue',
  weekly: 'purple',
}

export function StatusTag({ value }: { value: string }) {
  return <Tag color={colors[value] ?? 'default'}>{valueLabel(value)}</Tag>
}
