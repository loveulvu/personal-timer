import { FormEvent, useEffect, useMemo, useState } from 'react'
import { api, DailyTask, Project } from './api'

type TodayPageProps = {
  connected: boolean
  openProjects: () => void
}

type TaskForm = {
  projectId: string
  title: string
  estimatedMinutes: string
}

export function TodayPage({ connected, openProjects }: TodayPageProps) {
  const [date, setDate] = useState(todayString())
  const [tasks, setTasks] = useState<DailyTask[]>([])
  const [projects, setProjects] = useState<Project[]>([])
  const [form, setForm] = useState<TaskForm>({
    projectId: '',
    title: '',
    estimatedMinutes: '25',
  })
  const [loading, setLoading] = useState(false)
  const [projectsLoading, setProjectsLoading] = useState(false)
  const [error, setError] = useState('')
  const [projectsError, setProjectsError] = useState('')

  const projectNames = useMemo(
    () => new Map(projects.map((project) => [project.id, project.name])),
    [projects],
  )

  async function loadProjects() {
    if (!connected) return
    setProjectsLoading(true)
    setProjectsError('')
    try {
      const result = await api.getProjects()
      setProjects(result)
      setForm((current) => {
        const selectedStillExists = result.some(
          (project) => String(project.id) === current.projectId,
        )
        return {
          ...current,
          projectId: selectedStillExists ? current.projectId : String(result[0]?.id ?? ''),
        }
      })
    } catch (err) {
      setProjectsError(errorMessage(err))
    } finally {
      setProjectsLoading(false)
    }
  }

  async function loadTasks(selectedDate = date) {
    if (!connected) return
    setLoading(true)
    setError('')
    try {
      setTasks(await api.listDailyTasks(selectedDate))
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }

  async function createTask(event: FormEvent) {
    event.preventDefault()
    const projectId = Number(form.projectId)
    const estimatedMinutes = Number(form.estimatedMinutes)
    if (projects.length === 0) {
      setError('No projects yet. Create a project first.')
      return
    }
    if (!Number.isInteger(projectId) || projectId <= 0) {
      setError('Please select a project.')
      return
    }
    if (!form.title.trim()) {
      setError('title is required')
      return
    }
    if (!Number.isInteger(estimatedMinutes) || estimatedMinutes <= 0) {
      setError('estimated_minutes must be greater than 0')
      return
    }

    setLoading(true)
    setError('')
    try {
      await api.createDailyTask({
        project_id: projectId,
        task_date: date,
        title: form.title.trim(),
        estimated_minutes: estimatedMinutes,
      })
      setForm({ projectId: form.projectId, title: '', estimatedMinutes: '25' })
      await loadTasks(date)
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }

  async function runAction(task: DailyTask, action: 'start' | 'pause' | 'resume' | 'finish') {
    setLoading(true)
    setError('')
    try {
      if (action === 'start') await api.startTask(task.id)
      if (action === 'pause') await api.pauseTask(task.id)
      if (action === 'resume') await api.resumeTask(task.id)
      if (action === 'finish') await api.finishTask(task.id)
      await loadTasks(date)
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    if (connected) {
      loadProjects()
    }
  }, [connected])

  useEffect(() => {
    if (connected) {
      loadTasks(date)
    }
  }, [connected, date])

  return (
    <>
      {error && <div className="message error">{error}</div>}
      {projectsError && <div className="message error">{projectsError}</div>}

      <section className="content-grid">
        <section className="panel">
          <div className="panel-title">
            <h2>Today</h2>
            <label>
              Date
              <input
                type="date"
                value={date}
                onChange={(event) => setDate(event.target.value)}
              />
            </label>
          </div>

          {loading && <p className="muted">Loading...</p>}
          {!loading && tasks.length === 0 && <p className="muted">No tasks for this date.</p>}

          <div className="task-list">
            {tasks.map((task) => (
              <article key={task.id} className="task-row">
                <div>
                  <h3>{task.title}</h3>
                  <p>
                    status: {task.status} | estimate: {task.estimated_minutes} min | project:{' '}
                    {task.project_id ? projectNames.get(task.project_id) ?? `#${task.project_id}` : '-'}
                  </p>
                </div>
                <div className="actions">{renderActions(task, runAction)}</div>
              </article>
            ))}
          </div>
        </section>

        <aside className="panel">
          <h2>Create task</h2>
          {projectsLoading && <p className="muted">Loading projects...</p>}
          {!projectsLoading && !projectsError && projects.length === 0 && (
            <div className="empty-state">
              <p>No projects yet. Create a project first.</p>
              <button type="button" onClick={openProjects}>
                Go to Projects
              </button>
            </div>
          )}
          <form onSubmit={createTask} className="task-form">
            <label>
              project
              <select
                value={form.projectId}
                disabled={!connected || projects.length === 0}
                onChange={(event) => setForm({ ...form, projectId: event.target.value })}
              >
                <option value="">Select a project</option>
                {projects.map((project) => (
                  <option key={project.id} value={project.id}>
                    {project.name}
                  </option>
                ))}
              </select>
            </label>
            <label>
              task_date
              <input type="date" value={date} onChange={(event) => setDate(event.target.value)} />
            </label>
            <label>
              title
              <input
                value={form.title}
                onChange={(event) => setForm({ ...form, title: event.target.value })}
                placeholder="Read docs"
              />
            </label>
            <label>
              estimated_minutes
              <input
                type="number"
                min="1"
                value={form.estimatedMinutes}
                onChange={(event) => setForm({ ...form, estimatedMinutes: event.target.value })}
              />
            </label>
            <button
              type="submit"
              disabled={!connected || loading || projectsLoading || projects.length === 0}
            >
              Create
            </button>
          </form>
        </aside>
      </section>
    </>
  )
}

function renderActions(
  task: DailyTask,
  runAction: (task: DailyTask, action: 'start' | 'pause' | 'resume' | 'finish') => void,
) {
  if (task.status === 'planned') {
    return (
      <button type="button" onClick={() => runAction(task, 'start')}>
        Start
      </button>
    )
  }
  if (task.status === 'running') {
    return (
      <>
        <button type="button" onClick={() => runAction(task, 'pause')}>
          Pause
        </button>
        <button type="button" onClick={() => runAction(task, 'finish')}>
          Finish
        </button>
      </>
    )
  }
  if (task.status === 'paused') {
    return (
      <>
        <button type="button" onClick={() => runAction(task, 'resume')}>
          Resume
        </button>
        <button type="button" onClick={() => runAction(task, 'finish')}>
          Finish
        </button>
      </>
    )
  }
  return null
}

function todayString() {
  const now = new Date()
  const offset = now.getTimezoneOffset()
  const local = new Date(now.getTime() - offset * 60 * 1000)
  return local.toISOString().slice(0, 10)
}

function errorMessage(err: unknown) {
  if (err instanceof Error) return err.message
  if (typeof err === 'string') return err
  return 'Unknown error'
}
