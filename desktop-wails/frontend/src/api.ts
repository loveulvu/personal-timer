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

export type EstimatePreviewRiskLevel = 'insufficient_data' | 'low' | 'medium' | 'high'

export type EstimatePreviewRequest = {
  project_id: number
  title?: string
  estimated_minutes: number
}

export type EstimatePreviewResponse = {
  project_id: number
  input_estimated_minutes: number
  sample_count: number
  avg_estimated_minutes: number
  avg_actual_minutes: number
  overrun_rate: number
  risk_level: EstimatePreviewRiskLevel
  suggested_minutes: number
  split_recommended: boolean
  reason: string
}

export type PlanRiskLevel = 'insufficient_data' | 'low' | 'medium' | 'high'

export type PlanRiskResponse = {
  date: string
  planned_total_minutes: number
  recent_avg_actual_minutes: number
  recent_active_days: number
  plan_ratio: number
  risk_level: PlanRiskLevel
  reason: string
  suggestions: string[]
}

export type FeedbackTargetType = 'summary' | 'action_item' | 'memory'

export type FeedbackRequest = {
  target_type: FeedbackTargetType
  target_id: number
  target_index?: number | null
  feedback_value: string
  feedback_note?: string
}

export type FeedbackResponse = {
  id: number
  target_type: FeedbackTargetType
  target_id: number
  target_index?: number | null
  feedback_value: string
  feedback_note: string
  created_at: string
}

export type MemoryStatus = 'active' | 'archived'

export type MemoryListStatusFilter = MemoryStatus | 'all'

export type MemoryListItem = {
  id: number
  memory_type: string
  scope_type: string
  project_id?: number | null
  project_name?: string | null
  title: string
  content: string
  confidence: number
  support_count: number
  contradiction_count: number
  status: MemoryStatus
  first_seen_at: string
  last_seen_at: string
  created_at: string
  updated_at: string
}

export type MemoryEvidenceItem = {
  id: number
  memory_id: number
  source_type: string
  source_id?: number | null
  evidence_date: string
  excerpt?: string | null
  weight: number
  created_at: string
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

export type SummaryActionItem = {
  type: string
  priority: string
  title: string
  reason?: string
  suggested_project?: string
  suggested_minutes?: number
  source?: string
}

export type GenerateSummaryResult = {
  summary_id: number
  content: string
  action_items?: SummaryActionItem[] | null
}

export type AcceptActionItemResult = {
  summary_id: number
  item_index: number
  target_date: string
  target_task_id?: number | null
  created: boolean
  already_exists: boolean
  acceptance_status: 'accepted' | 'already_exists'
  task?: DailyTask
  message?: string
}

export type ActionItemAcceptance = {
  id: number
  summary_id: number
  item_index: number
  target_date: string
  target_task_id?: number | null
  status: 'accepted' | 'already_exists'
  created_at: string
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
  action_items?: SummaryActionItem[] | null
  created_at: string
}

export type LLMTestResponse = {
  status: string
  message: string
}

export type JsonValue =
  | null
  | boolean
  | number
  | string
  | JsonValue[]
  | { [key: string]: JsonValue }

export type AgentContextPreviewRequest = {
  goal: string
  target_date: string
  recent_days: number
}

export type AgentContextPreviewResponse = {
  context_pack: AgentContextPack
}

export type AgentContextPack = {
  user_goal: string
  target_date: string
  today_tasks: AgentContextTask[]
  recent_summaries: AgentContextSummary[]
  memories: AgentContextMemory[]
  plan_risk?: JsonValue
  recent_action_items: AgentContextActionItem[]
  constraints: string[]
  omitted_sections: string[]
}

export type AgentContextTask = {
  id: number
  project_id?: number | null
  project_name?: string
  task_date: string
  title: string
  estimated_minutes: number
  actual_minutes: number
  status: string
}

export type AgentContextSummary = {
  id: number
  summary_type: string
  start_date: string
  end_date: string
  content_excerpt: string
  action_items_excerpt?: string
  created_at: string
}

export type AgentContextMemory = {
  id: number
  memory_type: string
  scope_type: string
  title: string
  content: string
  confidence: number
  support_count: number
  contradiction_count: number
  evidence_count: number
  evidence_excerpt?: string
  status: string
}

export type AgentContextActionItem = {
  summary_id: number
  item_index: number
  content: string
  accepted: boolean
  target_date?: string
  target_task_id?: number | null
}

export type AgentRunRequest = AgentContextPreviewRequest

export type AgentRun = {
  id: number
  user_goal: string
  target_date: string
  status: string
  final_answer?: string
  pending_actions?: JsonValue
  error_message?: string
  created_at: string
  completed_at?: string | null
}

export type AgentRunListItem = {
  id: number
  user_goal: string
  target_date: string
  status: string
  final_answer_excerpt?: string
  proposal_count: number
  pending_proposal_count: number
  step_count: number
  created_at: string
  completed_at?: string | null
}

export type AgentStep = {
  id: number
  run_id: number
  step_index: number
  step_type: string
  tool_name?: string
  tool_input?: JsonValue
  tool_output?: JsonValue
  thought_summary?: string
  status: string
  error_message?: string
  created_at: string
}

export type AgentActionProposal = {
  id: number
  run_id: number
  step_id: number
  tool_name: string
  action_type: string
  payload?: JsonValue
  risk_level: string
  status: string
  result?: JsonValue
  error_message?: string
  target_entity_type?: string
  target_entity_id?: number | null
  created_at: string
  decided_at?: string | null
  executed_at?: string | null
}

export type AgentContextSnapshot = {
  id: number
  run_id: number
  context_pack: AgentContextPack
  token_estimate: number
  omitted_sections: string[]
  created_at: string
}

export type AgentRunResponse = {
  run: AgentRun
  steps: AgentStep[]
  proposals?: AgentActionProposal[]
}

export type AgentTrajectory = {
  run: AgentRun
  context_snapshot?: AgentContextSnapshot | null
  steps: AgentStep[]
  proposals: AgentActionProposal[]
}

export const api = {
  getStartupStatus: () => AppBindings.GetStartupStatus() as Promise<StartupStatus>,
  listDailyTasks: (date: string) => AppBindings.ListDailyTasks(date) as Promise<DailyTask[]>,
  createDailyTask: (request: CreateDailyTaskRequest) =>
    AppBindings.CreateDailyTask(request) as Promise<{ id: number }>,
  estimateTaskPreview: (request: EstimatePreviewRequest) =>
    AppBindings.EstimateTaskPreview({ ...request, title: request.title ?? '' }) as Promise<EstimatePreviewResponse>,
  getPlanRisk: (date?: string) =>
    AppBindings.GetPlanRisk(date ?? '') as Promise<PlanRiskResponse>,
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
  acceptSummaryActionItem: (summaryId: number, itemIndex: number, targetDate: string) =>
    AppBindings.AcceptSummaryActionItem(summaryId, itemIndex, targetDate) as unknown as Promise<AcceptActionItemResult>,
  listActionItemAcceptances: (summaryId: number) =>
    actionItemBindings.ListActionItemAcceptances(summaryId),
  submitFeedback: (request: FeedbackRequest) =>
    AppBindings.SubmitFeedback({
      target_type: request.target_type,
      target_id: request.target_id,
      feedback_value: request.feedback_value,
      feedback_note: request.feedback_note ?? '',
      ...(request.target_index == null ? {} : { target_index: request.target_index }),
    }) as Promise<FeedbackResponse>,
  listMemories: (status?: MemoryListStatusFilter, memoryType?: string, limit?: number) =>
    memoryBindings.ListMemories(status ?? '', memoryType ?? '', limit ?? 0),
  listMemoryEvidence: (memoryId: number) => memoryBindings.ListMemoryEvidence(memoryId),
  testLLM: () => AppBindings.TestLLM() as Promise<LLMTestResponse>,
  previewAgentContext: (request: AgentContextPreviewRequest) =>
    parseBinding<AgentContextPreviewResponse>(
      AppBindings.PreviewAgentContext(request.goal, request.target_date, request.recent_days),
    ),
  createAgentRun: (request: AgentRunRequest) =>
    parseBinding<AgentRunResponse>(
      AppBindings.CreateAgentRun(request.goal, request.target_date, request.recent_days),
    ),
  listAgentRuns: (status?: string, limit?: number) =>
    parseBinding<AgentRunListItem[]>(AppBindings.ListAgentRuns(status ?? '', limit ?? 20)),
  getAgentRun: (id: number) => parseBinding<AgentRunResponse>(AppBindings.GetAgentRun(id)),
  getAgentTrajectory: (id: number) =>
    parseBinding<AgentTrajectory>(AppBindings.GetAgentTrajectory(id)),
  listAgentActionProposals: (status?: string) =>
    parseBinding<AgentActionProposal[]>(AppBindings.ListAgentActionProposals(status ?? '')),
  acceptAgentActionProposal: (id: number) =>
    parseBinding<AgentActionProposal>(AppBindings.AcceptAgentActionProposal(id)),
  rejectAgentActionProposal: (id: number) =>
    parseBinding<AgentActionProposal>(AppBindings.RejectAgentActionProposal(id)),
}
import * as AppBindings from '../wailsjs/go/main/App'

const memoryBindings = AppBindings as unknown as {
  ListMemories: (status: string, memoryType: string, limit: number) => Promise<MemoryListItem[]>
  ListMemoryEvidence: (memoryId: number) => Promise<MemoryEvidenceItem[]>
}

const actionItemBindings = AppBindings as unknown as {
  ListActionItemAcceptances: (summaryId: number) => Promise<ActionItemAcceptance[]>
}

async function parseBinding<T>(promise: Promise<string>): Promise<T> {
  return JSON.parse(await promise) as T
}
