export type VersionInfo = {
  name: string
  version: string
  mode: string
}

export type ConfigStatus = {
  database: string
  llm_configured: boolean
  llm_base_url_configured: boolean
  llm_model_configured: boolean
  error?: string
}

export type StartupStatus = {
  connected: boolean
  version?: VersionInfo
  config?: ConfigStatus
  error?: string
}

export type DailyTask = {
  id: number
  project_id: number | null
  task_date: string
  title: string
  estimated_minutes: number
  status: 'planned' | 'running' | 'paused' | 'completed' | 'cancelled'
}

export type CreateDailyTaskRequest = {
  project_id: number
  task_date: string
  title: string
  estimated_minutes: number
}

export type Project = {
  id: number
  name: string
  description: string
  is_fixed: boolean
  created_at: string
  updated_at: string
}

export type ProjectInput = {
  name: string
  description: string
  is_fixed: boolean
}

export const api = {
  getStartupStatus: () => AppBindings.GetStartupStatus() as Promise<StartupStatus>,
  listDailyTasks: (date: string) => AppBindings.ListDailyTasks(date) as Promise<DailyTask[]>,
  createDailyTask: (request: CreateDailyTaskRequest) =>
    AppBindings.CreateDailyTask(request) as Promise<{ id: number }>,
  startTask: (id: number) => AppBindings.StartTask(id),
  pauseTask: (id: number) => AppBindings.PauseTask(id),
  resumeTask: (id: number) => AppBindings.ResumeTask(id),
  finishTask: (id: number) => AppBindings.FinishTask(id),
  getProjects: () => AppBindings.GetProjects() as Promise<Project[]>,
  getProject: (id: number) => AppBindings.GetProject(id) as Promise<Project>,
  createProject: (input: ProjectInput) =>
    AppBindings.CreateProject(input) as Promise<{ id: number }>,
  updateProject: (id: number, input: ProjectInput) => AppBindings.UpdateProject(id, input),
  deleteProject: (id: number) => AppBindings.DeleteProject(id),
}
import * as AppBindings from '../wailsjs/go/main/App'
