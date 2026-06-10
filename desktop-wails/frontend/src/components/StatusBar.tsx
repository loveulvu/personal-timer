import { ApiOutlined, DatabaseOutlined, MoonOutlined, RobotOutlined } from '@ant-design/icons'
import { Switch, Tag, Typography } from 'antd'
import { ReactNode } from 'react'
import { StartupStatus } from '../api'
import { valueLabel } from '../utils/labels'

type StatusBarProps = {
  startup: StartupStatus | null
  connected: boolean
  darkMode: boolean
  setDarkMode: (enabled: boolean) => void
}

export function StatusBar({ startup, connected, darkMode, setDarkMode }: StatusBarProps) {
  const databaseOK = startup?.config?.database === 'ok'
  const llmConfigured = startup?.config?.llm_configured === true

  return (
    <div className="status-bar">
      <StatusRow icon={<ApiOutlined />} label="后端" value={connected ? '在线' : '离线'} ok={connected} />
      <StatusRow icon={<DatabaseOutlined />} label="数据库" value={databaseOK ? '已连接' : valueLabel(startup?.config?.database ?? '未知')} ok={databaseOK} />
      <StatusRow icon={<RobotOutlined />} label="LLM" value={llmConfigured ? '已配置' : '未配置'} ok={llmConfigured} />
      <div className="status-row">
        <span className="status-label"><MoonOutlined /><Typography.Text>深色模式</Typography.Text></span>
        <Switch size="small" checked={darkMode} onChange={setDarkMode} aria-label="切换深色模式" />
      </div>
    </div>
  )
}

function StatusRow({ icon, label, value, ok }: { icon: ReactNode; label: string; value: string; ok: boolean }) {
  return (
    <div className="status-row">
      <span className="status-label">{icon}<Typography.Text>{label}</Typography.Text></span>
      <Tag color={ok ? 'success' : 'default'}>{value}</Tag>
    </div>
  )
}
