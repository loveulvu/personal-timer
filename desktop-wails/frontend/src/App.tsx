import { ConfigProvider, theme } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import { useEffect, useState } from 'react'
import { AgentPage } from './AgentPage'
import { api, StartupStatus } from './api'
import { AppLayout, Page } from './components/AppLayout'
import { DashboardPage } from './features/dashboard/DashboardPage'
import { MemoriesPage } from './MemoriesPage'
import { ProjectsPage } from './ProjectsPage'
import { StatsPage } from './StatsPage'
import { SummariesPage } from './SummariesPage'
import { errorMessage } from './utils/labels'

function App() {
  const [page, setPage] = useState<Page>('dashboard')
  const [startup, setStartup] = useState<StartupStatus | null>(null)
  const [error, setError] = useState('')
  const [darkMode, setDarkMode] = useState(() => localStorage.getItem('personal-timer-theme') === 'dark')

  const connected = startup?.connected === true

  async function checkBackend() {
    setError('')
    try {
      const result = await api.getStartupStatus()
      setStartup(result)
      if (!result.connected && result.error) {
        setError(errorMessage(result.error))
      }
    } catch (err) {
      setStartup({
        connected: false,
        error: '后端服务未启动，请先启动 backend-go',
      })
      setError(errorMessage(err))
    }
  }

  useEffect(() => {
    checkBackend()
  }, [])

  function changeTheme(enabled: boolean) {
    setDarkMode(enabled)
    localStorage.setItem('personal-timer-theme', enabled ? 'dark' : 'light')
  }

  return (
    <ConfigProvider
      locale={zhCN}
      theme={{
        algorithm: darkMode ? theme.darkAlgorithm : theme.defaultAlgorithm,
        token: {
          colorPrimary: '#1677ff',
          borderRadius: 12,
          fontFamily: '"Segoe UI", Inter, system-ui, sans-serif',
        },
      }}
    >
      <AppLayout
        page={page}
        setPage={setPage}
        startup={startup}
        connected={connected}
        error={error}
        refresh={checkBackend}
        darkMode={darkMode}
        setDarkMode={changeTheme}
      >
        {page === 'dashboard' && (
          <DashboardPage connected={connected} openProjects={() => setPage('projects')} />
        )}
        {page === 'projects' && <ProjectsPage connected={connected} />}
        {page === 'stats' && <StatsPage connected={connected} />}
        {page === 'summaries' && <SummariesPage connected={connected} />}
        {page === 'memories' && <MemoriesPage connected={connected} />}
        {page === 'agent' && <AgentPage connected={connected} />}
      </AppLayout>
    </ConfigProvider>
  )
}

export default App
