import { ConfigProvider, theme } from 'antd'
import { useEffect, useState } from 'react'
import { api, StartupStatus } from './api'
import { AppLayout, Page } from './components/AppLayout'
import { DashboardPage } from './features/dashboard/DashboardPage'
import { ProjectsPage } from './ProjectsPage'
import { StatsPage } from './StatsPage'
import { SummariesPage } from './SummariesPage'

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
        setError(result.error)
      }
    } catch (err) {
      setStartup({
        connected: false,
        error: 'Backend is not running. Please start backend-go first.',
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
      </AppLayout>
    </ConfigProvider>
  )
}

function errorMessage(err: unknown) {
  if (err instanceof Error) return err.message
  if (typeof err === 'string') return err
  return 'Unknown error'
}

export default App
