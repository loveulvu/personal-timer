import {
  BarChartOutlined,
  CalendarOutlined,
  FileTextOutlined,
  FolderOutlined,
  ReloadOutlined,
} from '@ant-design/icons'
import { Alert, Button, Card, Layout, Menu, Space, Tag, Typography } from 'antd'
import { ReactNode } from 'react'
import { StartupStatus } from '../api'

export type Page = 'today' | 'projects' | 'stats' | 'summaries'

type AppLayoutProps = {
  page: Page
  setPage: (page: Page) => void
  startup: StartupStatus | null
  connected: boolean
  title: string
  error: string
  refresh: () => void
  children: ReactNode
}

export function AppLayout({
  page,
  setPage,
  startup,
  connected,
  title,
  error,
  refresh,
  children,
}: AppLayoutProps) {
  return (
    <Layout className="app-layout">
      <Layout.Header className="app-header">
        <div>
          <Typography.Title level={3}>Personal Study Timer</Typography.Title>
          <Typography.Text type="secondary">{title}</Typography.Text>
        </div>
        <Button icon={<ReloadOutlined />} onClick={refresh}>Refresh status</Button>
      </Layout.Header>

      <Menu
        className="app-menu"
        mode="horizontal"
        selectedKeys={[page]}
        onClick={({ key }) => setPage(key as Page)}
        items={[
          { key: 'today', icon: <CalendarOutlined />, label: 'Today' },
          { key: 'projects', icon: <FolderOutlined />, label: 'Projects' },
          { key: 'stats', icon: <BarChartOutlined />, label: 'Stats' },
          { key: 'summaries', icon: <FileTextOutlined />, label: 'Summaries' },
        ]}
      />

      <Layout.Content className="app-content">
        <Card size="small" className="status-card">
          <Space wrap>
            <Tag color={connected ? 'success' : 'error'}>
              {connected ? 'backend connected' : 'backend disconnected'}
            </Tag>
            <Tag>version: {startup?.version?.version ?? '-'}</Tag>
            <Tag color={startup?.config?.database === 'ok' ? 'success' : 'default'}>
              database: {startup?.config?.database ?? '-'}
            </Tag>
            <Tag color={startup?.config?.llm_configured ? 'success' : 'default'}>
              llm: {startup?.config?.llm_configured ? 'configured' : 'not configured'}
            </Tag>
          </Space>
        </Card>

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
  )
}
