import type { DailyTask } from '../api'

export function displayElapsedSeconds(task: DailyTask, nowMs: number): number {
  const persistedSeconds = Math.max(0, task.actual_seconds || 0)
  if (task.status !== 'running' || !task.current_session_started_at) return persistedSeconds

  const startedAtMs = Date.parse(task.current_session_started_at)
  if (!Number.isFinite(startedAtMs)) return persistedSeconds

  return persistedSeconds + Math.max(0, Math.floor((nowMs - startedAtMs) / 1000))
}

export function timerProgressPercent(elapsedSeconds: number, estimatedMinutes: number): number {
  if (estimatedMinutes <= 0) return 0
  return Math.min(100, Math.max(0, (elapsedSeconds / (estimatedMinutes * 60)) * 100))
}
