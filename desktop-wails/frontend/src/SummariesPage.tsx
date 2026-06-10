import { FormEvent, useEffect, useState } from 'react'
import { api, GenerateSummaryResult, Summary } from './api'

type SummariesPageProps = {
  connected: boolean
}

type SummaryFilter = '' | 'daily' | 'weekly'

export function SummariesPage({ connected }: SummariesPageProps) {
  const today = todayString()
  const [dailyDate, setDailyDate] = useState(today)
  const [weeklyStart, setWeeklyStart] = useState(daysBefore(today, 6))
  const [weeklyEnd, setWeeklyEnd] = useState(today)
  const [filter, setFilter] = useState<SummaryFilter>('')
  const [summaries, setSummaries] = useState<Summary[]>([])
  const [detail, setDetail] = useState<Summary | null>(null)
  const [generated, setGenerated] = useState<GenerateSummaryResult | null>(null)
  const [message, setMessage] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  async function loadSummaries(selectedFilter = filter) {
    if (!connected) return
    setLoading(true)
    setError('')
    try {
      setSummaries(await api.getSummaries(selectedFilter))
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }

  async function testLLM() {
    setLoading(true)
    setError('')
    setMessage('')
    try {
      const result = await api.testLLM()
      setMessage(result.message || 'LLM connection works')
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }

  async function generateDaily(event: FormEvent) {
    event.preventDefault()
    await generate(() => api.generateDailySummary(dailyDate))
  }

  async function generateWeekly(event: FormEvent) {
    event.preventDefault()
    await generate(() => api.generateWeeklySummary(weeklyStart, weeklyEnd))
  }

  async function generate(action: () => Promise<GenerateSummaryResult>) {
    setLoading(true)
    setError('')
    setMessage('')
    setGenerated(null)
    try {
      setGenerated(await action())
      setMessage('Summary generated successfully.')
      await loadSummaries(filter)
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }

  async function viewSummary(id: number) {
    setLoading(true)
    setError('')
    try {
      setDetail(await api.getSummary(id))
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }

  async function deleteSummary(summary: Summary) {
    if (!window.confirm(`Delete ${summary.summary_type} summary #${summary.id}?`)) return
    setLoading(true)
    setError('')
    try {
      await api.deleteSummary(summary.id)
      if (detail?.id === summary.id) setDetail(null)
      await loadSummaries(filter)
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    if (connected) loadSummaries(filter)
  }, [connected, filter])

  return (
    <div className="page-stack">
      {error && <div className="message error">{error}</div>}
      {message && <div className="message success">{message}</div>}

      <section className="summary-tools">
        <section className="panel">
          <h2>LLM Test</h2>
          <p className="muted">Tests the configured LLM connection without exposing the API key.</p>
          <button type="button" onClick={testLLM} disabled={!connected || loading}>Test LLM</button>
        </section>

        <section className="panel">
          <h2>Generate Daily Summary</h2>
          <form className="stack-form" onSubmit={generateDaily}>
            <label>date<input type="date" value={dailyDate} onChange={(event) => setDailyDate(event.target.value)} /></label>
            <button type="submit" disabled={!connected || loading}>Generate Daily Summary</button>
          </form>
        </section>

        <section className="panel">
          <h2>Generate Weekly Summary</h2>
          <form className="stack-form" onSubmit={generateWeekly}>
            <label>start_date<input type="date" value={weeklyStart} onChange={(event) => setWeeklyStart(event.target.value)} /></label>
            <label>end_date<input type="date" value={weeklyEnd} onChange={(event) => setWeeklyEnd(event.target.value)} /></label>
            <button type="submit" disabled={!connected || loading}>Generate Weekly Summary</button>
          </form>
        </section>
      </section>

      {generated && (
        <section className="panel">
          <h2>Generated Summary #{generated.summary_id}</h2>
          <pre className="content-block">{generated.content}</pre>
        </section>
      )}

      <section className="panel">
        <div className="panel-title">
          <div>
            <h2>Summaries</h2>
            <p className="muted">Generated daily and weekly summaries.</p>
          </div>
          <label>
            type
            <select value={filter} onChange={(event) => setFilter(event.target.value as SummaryFilter)}>
              <option value="">all</option>
              <option value="daily">daily</option>
              <option value="weekly">weekly</option>
            </select>
          </label>
        </div>
        {loading && <p className="muted">Loading...</p>}
        {!loading && summaries.length === 0 && <p className="muted">No summaries found.</p>}
        <div className="summary-list">
          {summaries.map((summary) => (
            <article key={summary.id} className="summary-row">
              <div>
                <div className="project-heading">
                  <h3>{summary.summary_type} summary</h3>
                  <span className="project-id">#{summary.id}</span>
                </div>
                <p className="muted">
                  {summary.start_date} to {summary.end_date} | created: {formatDate(summary.created_at)}
                </p>
                <p className="summary-preview">{preview(summary.content)}</p>
              </div>
              <div className="actions">
                <button type="button" onClick={() => viewSummary(summary.id)} disabled={loading}>View</button>
                <button type="button" className="danger-button" onClick={() => deleteSummary(summary)} disabled={loading}>Delete</button>
              </div>
            </article>
          ))}
        </div>
      </section>

      {detail && (
        <section className="panel">
          <div className="panel-title">
            <h2>Summary Detail #{detail.id}</h2>
            <button type="button" className="secondary-button" onClick={() => setDetail(null)}>Close</button>
          </div>
          <p className="muted">{detail.summary_type} | {detail.start_date} to {detail.end_date}</p>
          <h3 className="section-heading">Content</h3>
          <pre className="content-block">{detail.content}</pre>
          {detail.source_data !== undefined && (
            <>
              <h3 className="section-heading">Source Data</h3>
              <pre className="content-block">{formatSourceData(detail.source_data)}</pre>
            </>
          )}
        </section>
      )}
    </div>
  )
}

function preview(content: string) {
  return content.length > 180 ? `${content.slice(0, 180)}...` : content
}

function formatSourceData(sourceData: unknown) {
  if (typeof sourceData === 'string') return sourceData
  try {
    return JSON.stringify(sourceData, null, 2)
  } catch {
    return String(sourceData)
  }
}

function formatDate(value: string) {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString()
}

function todayString() {
  const now = new Date()
  const offset = now.getTimezoneOffset()
  return new Date(now.getTime() - offset * 60 * 1000).toISOString().slice(0, 10)
}

function daysBefore(date: string, count: number) {
  const value = new Date(`${date}T00:00:00`)
  value.setDate(value.getDate() - count)
  const offset = value.getTimezoneOffset()
  return new Date(value.getTime() - offset * 60 * 1000).toISOString().slice(0, 10)
}

function errorMessage(err: unknown) {
  if (err instanceof Error) return err.message
  if (typeof err === 'string') return err
  return 'Unknown error'
}
