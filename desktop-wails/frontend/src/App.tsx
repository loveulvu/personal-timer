import { useEffect, useMemo, useState } from 'react'
import { api, StartupStatus } from './api'
import { ProjectsPage } from './ProjectsPage'
import { StatsPage } from './StatsPage'
import { SummariesPage } from './SummariesPage'
import { TodayPage } from './TodayPage'

type Page = 'today' | 'projects' | 'stats' | 'summaries'

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
    <main className="app-shell">
      <header className="topbar">
        <div>
          <h1>Personal Study Timer</h1>
          <p>{title}</p>
        </div>
        <button type="button" onClick={checkBackend}>
          Refresh status
        </button>
      </header>

      <nav className="tabs" aria-label="Main navigation">
        <button
          type="button"
          className={page === 'today' ? 'active' : ''}
          onClick={() => setPage('today')}
        >
          Today
        </button>
        <button
          type="button"
          className={page === 'projects' ? 'active' : ''}
          onClick={() => setPage('projects')}
        >
          Projects
        </button>
        <button
          type="button"
          className={page === 'stats' ? 'active' : ''}
          onClick={() => setPage('stats')}
        >
          Stats
        </button>
        <button
          type="button"
          className={page === 'summaries' ? 'active' : ''}
          onClick={() => setPage('summaries')}
        >
          Summaries
        </button>
      </nav>

      <section className={`status ${connected ? 'ok' : 'error'}`}>
        <span>{connected ? 'backend connected' : 'backend disconnected'}</span>
        <span>version: {startup?.version?.version ?? '-'}</span>
        <span>database: {startup?.config?.database ?? '-'}</span>
        <span>llm: {startup?.config?.llm_configured ? 'configured' : 'not configured'}</span>
      </section>

      {error && <div className="message error">{error}</div>}
      {startup?.error && connected && <div className="message warning">{startup.error}</div>}
      {!connected && (
        <div className="message error">Backend is not running. Please start backend-go first.</div>
      )}

      {page === 'today' && (
        <TodayPage connected={connected} openProjects={() => setPage('projects')} />
      )}
      {page === 'projects' && <ProjectsPage connected={connected} />}
      {page === 'stats' && <StatsPage connected={connected} />}
      {page === 'summaries' && <SummariesPage connected={connected} />}
    </main>
  )
}

function errorMessage(err: unknown) {
  if (err instanceof Error) return err.message
  if (typeof err === 'string') return err
  return 'Unknown error'
}

export default App
