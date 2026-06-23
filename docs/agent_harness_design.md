# Agent Harness Design

## 1. Project Goal

Personal Study Timer is moving from a local study timer and summary system toward a local desktop Agent Harness Workbench.

The goal is not to make the app an autonomous scheduler. The goal is to let a model reason over structured study data while the application keeps control of context selection, tool execution, permissions, logging, and evaluation.

Agent Phase 0 only defines the boundary. It does not implement runtime behavior.

## 2. Definition

In this project, Agent Harness means the engineering layer around a model. It is not the model itself, and it is not a LangChain-style framework.

The harness owns:

- Context Pack: selected structured data given to the model.
- Tool Registry: named operations the model may request.
- Agent Loop: the future run lifecycle around model calls and tool calls.
- Permission Guard: rules that separate safe reads from user-approved writes.
- Trajectory Log: durable records of runs, steps, context snapshots, and proposals.
- Memory Integration: controlled use of study memories and evidence.
- Evaluation / Replay: repeatable checks for harness behavior.

## 3. Existing System Mapping

Current system data already maps cleanly to Agent concepts:

- `daily_tasks` / `time_sessions` -> Environment State
- `generated_summaries` -> Historical Context
- `study_memories` / `study_memory_evidence` -> Long-term Memory
- `study_feedback` -> Human Feedback
- `summary_action_item_acceptances` -> Action Acceptance Signal
- plan risk / estimate preview -> Planning Signals

The harness should prefer these deterministic signals before asking a model to infer from raw text.

## 4. Agent Run Lifecycle

Future Agent Run flow:

1. User Goal
2. Build Context Pack
3. Model selects read tool or proposes write action
4. Harness executes safe read tools
5. Write tools become Action Proposal
6. User accepts or rejects
7. Save trajectory
8. Optional replay / evaluation

Read operations can be automatic. Write operations must stay proposals until the user confirms them.

## 5. Tool Registry Design

Future tools are grouped by risk:

- read
- write
- destructive

Rules:

- Read tools can execute automatically.
- Write tools can only create an action proposal.
- Destructive tools are not exposed for now.
- All tool input must be validated by the backend.
- All tool calls must eventually be written to `agent_steps`.

The registry should describe tool name, purpose, risk level, input schema, and output schema. It should not let the model call arbitrary code.

## 6. Context Pack Design

Future `ContextPack` includes:

- `user_goal`
- `target_date`
- `today_tasks`
- `recent_summaries`
- `memories`
- `plan_risk`
- `recent_action_items`
- `constraints`
- `omitted_sections`

Rules:

- Do not put all historical data into the prompt.
- Prefer structured data over raw text.
- Prefer active memories.
- Memory must be tied to evidence.
- Archived, wrong, or outdated memory should be omitted or heavily downweighted.

`omitted_sections` should make context truncation explicit so evaluation can see what was left out.

## 7. Permission Guard

The model cannot directly create, modify, or delete tasks.

Write flow:

1. Model requests a write tool.
2. Harness creates an action proposal.
3. User accepts or rejects the proposal.
4. Backend executes the real service only after acceptance.

Proposal acceptance must be idempotent. Repeated accept should not create duplicate tasks or duplicate side effects.

Rejecting a proposal must not execute any write operation.

## 8. Trajectory Log

Future persistence should record:

- `agent_runs`
- `agent_steps`
- `agent_context_snapshots`
- `agent_action_proposals`

This phase does not create these tables. It only defines the direction so later migrations can follow a stable shape.

The log should store tool inputs, tool outputs, proposal decisions, status, errors, and compact thought summaries. It must not store full chain-of-thought.

## 9. Evaluation / Replay

Initial evaluation cases:

- `task_creation_requires_confirmation`
- `read_today_tasks_no_confirmation`
- `memory_recall_used_for_plan`
- `reject_proposal_no_write`
- `repeated_accept_idempotent`

Replay should verify harness behavior, not model style. The important checks are permission boundaries, deterministic tool behavior, context selection, and proposal idempotency.

## 10. Current Phase Scope

Agent Phase 0 includes only:

- This design document.
- Backend base types under `backend-go/internal/agent`.

Agent Phase 0 does not:

- Add LangChain / LangGraph.
- Add Multi-Agent behavior.
- Add complex RAG.
- Allow Agent to directly write database rows.
- Save full chain-of-thought.
- Execute arbitrary shell commands.
- Modify or delete files through Agent tools.
- Connect an LLM.
- Add UI.
- Add database migrations.
- Implement Agent Run API.
- Refactor summary, memory, task, or timer logic.
