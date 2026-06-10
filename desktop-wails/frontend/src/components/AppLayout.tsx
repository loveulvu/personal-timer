import {
  BarChartOutlined,
  DashboardOutlined,
  FileTextOutlined,
  FolderOutlined,
  ReloadOutlined,
  FieldTimeOutlined,
} from '@ant-design/icons'
import { Alert, Button, Layout, Menu, Typography } from 'antd'
import { ReactNode } from 'react'
import { StartupStatus } from '../api'
import { StatusBar } from './StatusBar'

export type Page = 'dashboard' | 'projects' | 'stats' | 'summaries'

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
            <Typography.Text strong>Personal Timer</Typography.Text>
            <Typography.Text type="secondary">Study dashboard</Typography.Text>
          </div>
        </div>
        <Menu
          className="app-menu"
          mode="inline"
          selectedKeys={[page]}
          onClick={({ key }) => setPage(key as Page)}
          items={[
            { key: 'dashboard', icon: <DashboardOutlined />, label: 'Dashboard' },
            { key: 'projects', icon: <FolderOutlined />, label: 'Projects' },
            { key: 'stats', icon: <BarChartOutlined />, label: 'Stats' },
            { key: 'summaries', icon: <FileTextOutlined />, label: 'Summaries' },
          ]}
        />
        <div className="sidebar-footer">
          <Button block icon={<ReloadOutlined />} onClick={refresh}>Refresh status</Button>
          <StatusBar startup={startup} connected={connected} darkMode={darkMode} setDarkMode={setDarkMode} />
        </div>
      </Layout.Sider>

      <Layout className="main-shell">
        <Layout.Content className="app-content">
          {error && <Alert className="page-alert" type="error" showIcon title={error} />}
          {startup?.error && connected && (
            <Alert className="page-alert" type="warning" showIcon title={startup.error} />
          )}
          {!connected && (
            <Alert
              className="page-alert"
              type="error"
              showIcon
              title="Backend is not running. Please start backend-go first."
            />
          )}
          {children}
        </Layout.Content>
      </Layout>
    </Layout>
  )
}
