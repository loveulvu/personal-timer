import {
  ApiOutlined,
  BarChartOutlined,
  DashboardOutlined,
  DatabaseOutlined,
  FieldTimeOutlined,
  FileTextOutlined,
  FolderOutlined,
  ReloadOutlined,
} from '@ant-design/icons'
import { Alert, Button, Layout, Menu, Typography } from 'antd'
import { ReactNode } from 'react'
import { StartupStatus } from '../api'
import { errorMessage } from '../utils/labels'
import { StatusBar } from './StatusBar'

export type Page = 'dashboard' | 'projects' | 'stats' | 'summaries' | 'memories' | 'agent'

type AppLayoutProps = {
  page: Page
  setPage: (page: Page) => void
  startup: StartupStatus | null
  connected: boolean
  error: string
  refresh: () => void
  darkMode: boolean
  setDarkMode: (enabled: boolean) => void
  children: ReactNode
}

export function AppLayout({
  page,
  setPage,
  startup,
  connected,
  error,
  refresh,
  darkMode,
  setDarkMode,
  children,
}: AppLayoutProps) {
  return (
    <Layout className={`app-layout ${darkMode ? 'is-dark' : ''}`}>
      <Layout.Sider className="app-sidebar" width={238} theme={darkMode ? 'dark' : 'light'}>
        <div className="app-brand">
          <span className="app-brand-icon"><FieldTimeOutlined /></span>
          <div>
            <Typography.Text strong>个人计时器</Typography.Text>
            <Typography.Text type="secondary">学习仪表盘</Typography.Text>
          </div>
        </div>
        <Menu
          className="app-menu"
          mode="inline"
          selectedKeys={[page]}
          onClick={({ key }) => setPage(key as Page)}
          items={[
            { key: 'dashboard', icon: <DashboardOutlined />, label: '仪表盘' },
            { key: 'projects', icon: <FolderOutlined />, label: '项目' },
            { key: 'stats', icon: <BarChartOutlined />, label: '统计' },
            { key: 'summaries', icon: <FileTextOutlined />, label: '总结' },
            { key: 'memories', icon: <DatabaseOutlined />, label: '记忆' },
            { key: 'agent', icon: <ApiOutlined />, label: 'Agent' },
          ]}
        />
        <div className="sidebar-footer">
          <Button block icon={<ReloadOutlined />} onClick={refresh}>刷新状态</Button>
          <StatusBar startup={startup} connected={connected} darkMode={darkMode} setDarkMode={setDarkMode} />
        </div>
      </Layout.Sider>

      <Layout className="main-shell">
        <Layout.Content className="app-content">
          {error && <Alert className="page-alert" type="error" showIcon title={error} />}
          {startup?.error && connected && (
            <Alert className="page-alert" type="warning" showIcon title={errorMessage(startup.error)} />
          )}
          {!connected && (
            <Alert
              className="page-alert"
              type="error"
              showIcon
              title="后端服务未启动，请先启动 backend-go"
            />
          )}
          {children}
        </Layout.Content>
      </Layout>
    </Layout>
  )
}
