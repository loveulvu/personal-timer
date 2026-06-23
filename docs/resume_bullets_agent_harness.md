# Resume Bullets: Agent Harness

Project title:

`Personal Agent Harness Workbench | Go + Wails + MySQL + LLM Agent`

## Resume Bullets

- Built a local desktop Agent Harness Workbench with Go, Gin, Wails, React, TypeScript, and MySQL for structured study planning and agent workflow experiments.
- Designed a Tool Registry with explicit read/write risk levels, backend input validation, and disabled destructive tools.
- Implemented ContextPack construction from structured data including daily tasks, recent summaries, active memories, plan risk, constraints, and omitted sections.
- Added human-in-the-loop Action Proposal flow so model-requested write tools cannot directly mutate business tables.
- Implemented idempotent proposal accept/reject behavior for task creation and task completion actions.
- Persisted Agent Run trajectory data with runs, steps, context snapshots, tool inputs/outputs, statuses, and proposals for auditability.
- Built a minimal Wails Agent Console to preview context, launch runs, inspect steps, and accept or reject proposals.
- Added evaluation / replay checks for key harness behaviors without relying on a real external LLM.
- Modeled evidence-linked study memory in MySQL so long-term context is grounded in source records instead of free-form model claims.
- Kept deterministic rules and MySQL state as the source of truth, with LLM behavior constrained behind service boundaries.

## Short Version

- Built a Go + Wails + MySQL Agent Harness Workbench with tool registry, context packs, permission-guarded action proposals, trajectory logs, and eval/replay checks.

## Honest Scope

This is a local portfolio/workbench project, not a production multi-tenant agent platform. It focuses on architecture boundaries, safety, auditability, and evaluation before adding more autonomous behavior.
