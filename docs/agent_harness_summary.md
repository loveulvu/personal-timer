# Agent Harness Summary

This document summarizes the implemented Agent Harness after Phases A-G. The detailed design boundary is in `docs/agent_harness_design.md`.

## Tool Registry

The backend exposes named tools with schema and risk metadata.

- Read tools: `list_today_tasks`, `evaluate_plan_risk`, `recall_memories`.
- Write tools: `create_daily_task`, `finish_task`.
- Destructive tools are not registered.

Read tools can execute automatically. Write tools return proposal-like results and are not allowed to write business tables inside the agent loop.

## Context Pack

`ContextPack` is built from structured MySQL data:

- user goal and target date
- today tasks
- recent summaries
- active memories with evidence excerpts
- plan risk
- recent action items
- constraints
- omitted sections

It deliberately does not load all history. `omitted_sections` records capped ranges, omitted low-confidence memories, and excerpted content.

## Agent Run / Loop

An Agent Run creates a run record, builds context, calls a `ModelClient`, executes safe read tools, and stops on final answer, failure, or required confirmation.

Tests use fake model clients. The fallback model is deterministic and does not call an external LLM.

## Action Proposal

Write tools create `agent_action_proposals`.

- `pending`: waiting for user decision.
- `rejected`: no business write.
- `executed`: accepted and service executed.
- repeated accept returns the existing executed result.

The frontend never writes `daily_tasks` directly for agent actions.

## Trajectory Log

Trajectory data includes:

- `agent_runs`
- `agent_steps`
- `agent_context_snapshots`
- `agent_action_proposals`

The log stores tool inputs, tool outputs, statuses, errors, and short thought summaries. It does not store full chain-of-thought.

## Agent Console

The Wails Agent page can:

- preview context
- create an agent run
- list recent runs
- open trajectory detail
- inspect steps and JSON payloads
- show proposals
- accept or reject pending proposals
- show final answer, error, or requires-confirmation state

## Evaluation / Replay

Phase G adds fixed internal eval cases and read-only replay:

- read tool without confirmation
- write tool requires proposal
- reject proposal does not write
- repeated accept is idempotent
- context snapshot and constraints exist
- trajectory can be replayed without LLM

Replay summarizes existing trajectory data. It does not re-run the model or mutate state.

## Why Not LangChain / RAG

The project is testing harness boundaries more than model capability. The hard parts here are permission control, structured context, durable audit logs, idempotent proposals, and evaluation. A framework would not remove those responsibilities.

RAG is intentionally deferred. The current data is already structured in MySQL. A vector database only makes sense after importing large unstructured notes or interview-question corpora.

## Current Boundaries

- no multi-agent flow
- no destructive tools
- no arbitrary shell or file access
- no full chain-of-thought storage
- no automatic task writes from model output
- no vector search

## Later Work

- optional real `ModelClient` provider
- more eval cases
- screenshots in README
- RAG only for large imported documents
