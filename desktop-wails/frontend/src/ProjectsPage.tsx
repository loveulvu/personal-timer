import { FormEvent, useEffect, useState } from 'react'
import { api, Project, ProjectInput } from './api'

type ProjectsPageProps = {
  connected: boolean
}

const emptyForm: ProjectInput = {
  name: '',
  description: '',
  is_fixed: false,
}

export function ProjectsPage({ connected }: ProjectsPageProps) {
  const [projects, setProjects] = useState<Project[]>([])
  const [createForm, setCreateForm] = useState<ProjectInput>(emptyForm)
  const [editingId, setEditingId] = useState<number | null>(null)
  const [editForm, setEditForm] = useState<ProjectInput>(emptyForm)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  async function loadProjects() {
    if (!connected) return
    setLoading(true)
    setError('')
    try {
      setProjects(await api.getProjects())
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }

  async function createProject(event: FormEvent) {
    event.preventDefault()
    if (!createForm.name.trim()) {
      setError('name is required')
      return
    }
    setLoading(true)
    setError('')
    try {
      await api.createProject({ ...createForm, name: createForm.name.trim() })
      setCreateForm(emptyForm)
      await loadProjects()
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }

  function startEditing(project: Project) {
    setEditingId(project.id)
    setEditForm({
      name: project.name,
      description: project.description,
      is_fixed: project.is_fixed,
    })
    setError('')
  }

  async function updateProject(event: FormEvent, id: number) {
    event.preventDefault()
    if (!editForm.name.trim()) {
      setError('name is required')
      return
    }
    setLoading(true)
    setError('')
    try {
      await api.updateProject(id, { ...editForm, name: editForm.name.trim() })
      setEditingId(null)
      await loadProjects()
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }

  async function deleteProject(project: Project) {
    const confirmed = window.confirm(
      `Delete "${project.name}"? Existing tasks will be kept but their project link will be removed.`,
    )
    if (!confirmed) return

    setLoading(true)
    setError('')
    try {
      await api.deleteProject(project.id)
      if (editingId === project.id) setEditingId(null)
      await loadProjects()
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

  return (
    <>
      {error && <div className="message error">{error}</div>}

      <section className="projects-grid">
        <aside className="panel">
          <h2>Create project</h2>
          <ProjectForm
            form={createForm}
            setForm={setCreateForm}
            onSubmit={createProject}
            submitLabel="Create"
            disabled={!connected || loading}
          />
        </aside>

        <section className="panel">
          <div className="panel-title">
            <div>
              <h2>Projects</h2>
              <p className="muted">
                Deleting a project will keep existing tasks but remove their project link.
              </p>
            </div>
            <button type="button" onClick={loadProjects} disabled={!connected || loading}>
              Refresh
            </button>
          </div>

          {loading && <p className="muted">Loading...</p>}
          {!loading && projects.length === 0 && <p className="muted">No projects yet.</p>}

          <div className="project-list">
            {projects.map((project) => (
              <article key={project.id} className="project-row">
                {editingId === project.id ? (
                  <ProjectForm
                    form={editForm}
                    setForm={setEditForm}
                    onSubmit={(event) => updateProject(event, project.id)}
                    submitLabel="Save"
                    disabled={loading}
                    onCancel={() => setEditingId(null)}
                  />
                ) : (
                  <>
                    <div className="project-heading">
                      <h3>{project.name}</h3>
                      <span className="project-id">#{project.id}</span>
                      {project.is_fixed && <span className="fixed-badge">fixed</span>}
                    </div>
                    <p>{project.description || 'No description'}</p>
                    <p className="muted">
                      created: {formatDate(project.created_at)} | updated:{' '}
                      {formatDate(project.updated_at)}
                    </p>
                    <div className="actions">
                      <button type="button" onClick={() => startEditing(project)} disabled={loading}>
                        Edit
                      </button>
                      <button
                        type="button"
                        className="danger-button"
                        onClick={() => deleteProject(project)}
                        disabled={loading}
                      >
                        Delete
                      </button>
                    </div>
                  </>
                )}
              </article>
            ))}
          </div>
        </section>
      </section>
    </>
  )
}

type ProjectFormProps = {
  form: ProjectInput
  setForm: (form: ProjectInput) => void
  onSubmit: (event: FormEvent) => void
  submitLabel: string
  disabled: boolean
  onCancel?: () => void
}

function ProjectForm({
  form,
  setForm,
  onSubmit,
  submitLabel,
  disabled,
  onCancel,
}: ProjectFormProps) {
  return (
    <form className="project-form" onSubmit={onSubmit}>
      <label>
        name
        <input
          value={form.name}
          onChange={(event) => setForm({ ...form, name: event.target.value })}
          placeholder="Go backend"
          required
        />
      </label>
      <label>
        description
        <textarea
          value={form.description}
          onChange={(event) => setForm({ ...form, description: event.target.value })}
          placeholder="Go backend learning"
          rows={3}
        />
      </label>
      <label className="checkbox-label">
        <input
          type="checkbox"
          checked={form.is_fixed}
          onChange={(event) => setForm({ ...form, is_fixed: event.target.checked })}
        />
        fixed project
      </label>
      <div className="actions form-actions">
        {onCancel && (
          <button type="button" className="secondary-button" onClick={onCancel} disabled={disabled}>
            Cancel
          </button>
        )}
        <button type="submit" disabled={disabled}>
          {submitLabel}
        </button>
      </div>
    </form>
  )
}

function formatDate(value: string) {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString()
}

function errorMessage(err: unknown) {
  if (err instanceof Error) return err.message
  if (typeof err === 'string') return err
  return 'Unknown error'
}
