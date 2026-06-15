import assert from 'node:assert/strict'
import test from 'node:test'
import type { DailyTask } from '../api.ts'
import { displayElapsedSeconds, timerProgressPercent } from './timerDisplay.ts'

const startedAt = '2026-06-15T10:00:00+08:00'
const startedAtMs = Date.parse(startedAt)

function task(overrides: Partial<DailyTask>): DailyTask {
  return {
    id: 1,
    project_id: 1,
    task_date: '2026-06-15',
    title: 'test',
    estimated_minutes: 30,
    status: 'running',
    finish_note: null,
    finish_description: null,
    completed_at: null,
    actual_seconds_override: null,
    actual_seconds: 120,
    current_session_started_at: startedAt,
    ...overrides,
  }
}

test('running task elapsed time increases from current session started_at', () => {
  assert.equal(displayElapsedSeconds(task({}), startedAtMs + 10_000), 130)
})

test('paused and completed tasks do not increase', () => {
  assert.equal(displayElapsedSeconds(task({ status: 'paused' }), startedAtMs + 10_000), 120)
  assert.equal(displayElapsedSeconds(task({ status: 'completed' }), startedAtMs + 10_000), 120)
})

test('progress is capped at 100 percent', () => {
  assert.equal(timerProgressPercent(31 * 60, 30), 100)
})

test('missing or invalid started_at falls back to persisted seconds', () => {
  assert.equal(displayElapsedSeconds(task({ current_session_started_at: null }), startedAtMs + 10_000), 120)
  assert.equal(displayElapsedSeconds(task({ current_session_started_at: 'invalid' }), startedAtMs + 10_000), 120)
})
