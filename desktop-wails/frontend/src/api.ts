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
  finish_note: string | null
  finish_description: string | null
  completed_at: string | null
  actual_seconds_override: number | null
  actual_seconds: number
  current_session_started_at: string | null
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
  category: ProjectCategory
  include_in_summary: boolean
  created_at: string
  updated_at: string
}

export type ProjectCategory = 'study' | 'project' | 'life' | 'break' | 'other'

export type ProjectInput = {
  name: string
  description: string
  is_fixed: boolean
  category: ProjectCategory
  include_in_summary: boolean
}

export type DailyTaskStat = {
  task_id: number
  title: string
  status: string
  estimated_minutes: number
  actual_seconds: number
  actual_minutes: number
}

export type DailyStats = {
  date: string
  total_seconds: number
  total_minutes: number
  completed_count: number
  unfinished_count: number
  tasks: DailyTaskStat[]
}

export type WeeklyDayStat = {
  date: string
  total_seconds: number
  total_minutes: number
  completed_count: number
  unfinished_count: number
}

export type WeeklyProjectStat = {
  project_id: number
  project_name: string
  task_count: number
  completed_count: number
  total_seconds: number
  total_minutes: number
}

export type WeeklyStats = {
  start_date: string
  end_date: string
  total_seconds: number
  total_minutes: number
  completed_count: number
  unfinished_count: number
  days: WeeklyDayStat[]
  projects: WeeklyProjectStat[]
}

export type GenerateSummaryResult = {
  summary_id: number
  content: string
}

export type FinishTaskRequest = {
  finish_note: string
  finish_description: string
}

export type UpdateCompletedTaskRequest = FinishTaskRequest & {
  actual_minutes_override?: number
  clear_actual_minutes_override?: boolean
}

export type Summary = {
  id: number
  summary_type: 'daily' | 'weekly'
  start_date: string
  end_date: string
  content: string
  source_data?: unknown
  created_at: string
}

export type LLMTestResponse = {
  status: string
  message: string
}

export const api = {
  getStartupStatus: () => AppBindings.GetStartupStatus() as Promise<StartupStatus>,
  listDailyTasks: (date: string) => AppBindings.ListDailyTasks(date) as Promise<DailyTask[]>,
  createDailyTask: (request: CreateDailyTaskRequest) =>
    AppBindings.CreateDailyTask(request) as Promise<{ id: number }>,
  startTask: (id: number) => AppBindings.StartTask(id),
  pauseTask: (id: number) => AppBindings.PauseTask(id),
  resumeTask: (id: number) => AppBindings.ResumeTask(id),
  finishTask: (id: number, input: FinishTaskRequest) => AppBindings.FinishTask(id, input),
  updateCompletedTask: (id: number, input: UpdateCompletedTaskRequest) =>
    AppBindings.UpdateCompletedTask(id, input),
  deleteCompletedTask: (id: number) => AppBindings.DeleteCompletedTask(id),
  getProjects: () => AppBindings.GetProjects() as Promise<Project[]>,
  getProject: (id: number) => AppBindings.GetProject(id) as Promise<Project>,
  createProject: (input: ProjectInput) =>
    AppBindings.CreateProject(input) as Promise<{ id: number }>,
  updateProject: (id: number, input: ProjectInput) => AppBindings.UpdateProject(id, input),
  deleteProject: (id: number) => AppBindings.DeleteProject(id),
  getDailyStats: (date: string) => AppBindings.GetDailyStats(date) as Promise<DailyStats>,
  getWeeklyStats: (startDate: string, endDate: string) =>
    AppBindings.GetWeeklyStats(startDate, endDate) as Promise<WeeklyStats>,
  generateDailySummary: (date: string) =>
    AppBindings.GenerateDailySummary(date) as Promise<GenerateSummaryResult>,
  generateWeeklySummary: (startDate: string, endDate: string) =>
    AppBindings.GenerateWeeklySummary(startDate, endDate) as Promise<GenerateSummaryResult>,
  getSummaries: (summaryType: string) =>
    AppBindings.GetSummaries(summaryType) as Promise<Summary[]>,
  getSummary: (id: number) => AppBindings.GetSummary(id) as Promise<Summary>,
  deleteSummary: (id: number) => AppBindings.DeleteSummary(id),
  testLLM: () => AppBindings.TestLLM() as Promise<LLMTestResponse>,
}
import * as AppBindings from '../wailsjs/go/main/App'
