# Codex build brief: PB4T calendar conflict and resolution app

This directory is a standard OpenSpec `spec-driven` change for a net-new project to be built in Codex. Use the entire directory, including the seven capability specs. Start with `proposal.md`, then `design.md`, `specs/*/spec.md`, and `tasks.md`.

## Build brief

Build a weekly PB4T planning application in Codex with a React/TypeScript frontend, Go backend, SQLite persistence, and server-side OpenAI departmental agents. The backend manages a durable local database file; no separate database service is required. Follow the requirements and acceptance scenarios in this change. Use the design as the architecture and the unchecked tasks as the build sequence.

Local startup is a deliverable: with Go and Node.js/npm installed, `npm start` must prepare dependencies, build the app, initialize SQLite and demo data automatically on first launch, and serve one localhost URL. No Docker, separate database server, or C compiler is required. Core planning works without an OpenAI key; optional server-side configuration enables real agent negotiation and free-form assistance. Ordinary restarts preserve data; `npm run reset-demo` explicitly restores the demo. The root README documents the implemented local commands.

The application must support:

1. Enter desired events, constraints, and training requirements; generate and review a feasible weekly calendar.
2. Add an event with a point of contact and any number of constraints; receive specific conflict feedback and alternatives.
3. Move an event on the calendar; receive conflict feedback and review proposed repairs before applying them.

Overlapping events are allowed when all requirements can be met. Check actual personnel assignments across shared qualification pools, combined asset capacity, operating windows, incompatible evolutions, and configured rest policies. Agents negotiate concrete departmental preferences and alternatives; the deterministic engine checks feasibility. Keep planned and completed training distinct.

Use a small synthetic shipboard dataset and illustrative policy values. This is a hackathon build: automated tests and comprehensive edge-case handling are out of scope. Use manual walkthroughs in tasks.md to demonstrate the core flows. Do not silently remove events, relax hard constraints, or claim a proven optimum or independently verified SFTRM compliance.

Follow applicable repository instructions. Do not auto-commit or auto-push. Do not create a branch without a supplied Linear issue; if branch creation is requested later, use `timlee/{LINEAR-ISSUE-NUMBER}`. No Linear issue is supplied by this handoff.

## Artifact map

| Artifact | Purpose |
| --- | --- |
| proposal.md | Problem, scope, and capability inventory |
| design.md | Architecture, data model, engine, agents, UI, and API boundaries |
| specs/weekly-calendar/spec.md | Calendar interaction and proposal review |
| specs/planning-data/spec.md | Persisted inputs and constraint authoring |
| specs/conflict-detection/spec.md | Shared deterministic feasibility checks |
| specs/schedule-resolution/spec.md | Weekly generation and local repair |
| specs/department-agents/spec.md | Tool-driven departmental negotiation |
| specs/training-readiness/spec.md | Configured obligations and training coverage |
| specs/local-runtime/spec.md | One-command local startup, initialization, and optional AI configuration |
| tasks.md | Ordered implementation and manual demonstration checklist |

## Scope of this planning chat

The user invoked the apply workflow to authorize implementation in Codex. Track completed implementation and remaining manual verification in tasks.md.
