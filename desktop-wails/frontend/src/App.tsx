import { useEffect, useMemo, useState } from 'react'
import { api, StartupStatus } from './api'
import { AppLayout, Page } from './components/AppLayout'
import { ProjectsPage } from './ProjectsPage'
import { StatsPage } from './StatsPage'
import { SummariesPage } from './SummariesPage'
import { TodayPage } from './TodayPage'

function App() {
  const [page, setPage] = useState<Page>('today')
  const [startup, setStartup] = useState<StartupStatus | null>(null)
  const [error, setError] = useState('')

  const connected = startup?.connected === true
  const title = useMemo(() => {
    if (!startup) return 'Checking backend...'
    if (!startup.connected) return 'Backend disconnected'
    return `Backend connected ${startup.version?.version ?? ''}`
  }, [startup])

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

  return (
    <AppLayout
      page={page}
      setPage={setPage}
      startup={startup}
      connected={connected}
      title={title}
      error={error}
      refresh={checkBackend}
    >
      {page === 'today' && (
        <TodayPage connected={connected} openProjects={() => setPage('projects')} />
      )}
      {page === 'projects' && <ProjectsPage connected={connected} />}
      {page === 'stats' && <StatsPage connected={connected} />}
      {page === 'summaries' && <SummariesPage connected={connected} />}
    </AppLayout>
  )
}

function errorMessage(err: unknown) {
  if (err instanceof Error) return err.message
  if (typeof err === 'string') return err
  return 'Unknown error'
}

export default App
