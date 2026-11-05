## Why

Weekly PB4T planning must reconcile departmental drills with ship operations, scarce qualified personnel, shared assets, fatigue policies, and training deadlines. A shared calendar with explainable conflict detection and agent-assisted resolution can make those tradeoffs visible and produce a feasible weekly plan.

## What Changes

- Create a new full-stack application with a weekly calendar, a Go backend, SQLite persistence, and a React frontend.
- Make local use simple: with Go and Node.js/npm installed, `npm start` prepares dependencies, builds the application, initializes SQLite and demo data on first launch, and serves one localhost URL. No Docker, database server, or C compiler is required. Core calendar and scheduling features work without an OpenAI key; live agent features require a server-side key.
- Support event creation, editing, dragging, and removal, with a point of contact, department, duration, requested time, scheduling flexibility, and any number of constraints.
- Maintain qualified personnel rosters, availability and duty commitments, asset capacities, operational windows, reusable constraints, and training requirements.
- Detect specific personnel, asset, timing, incompatibility, and configured fatigue conflicts. Allow simultaneous events when their combined requirements can be satisfied.
- Generate a weekly schedule or repair a proposed change, keeping fixed events in place and preferring fewer changes to the existing calendar.
- Coordinate departmental OpenAI agents that inspect persisted planning data, negotiate alternatives, and explain proposals. A deterministic scheduling engine validates every proposed result.
- Preview proposed changes and their effects before the user applies them to the calendar.
- Track completed and planned training against user-supplied periodicity requirements; do not treat scheduled training as completed training or claim independently verified SFTRM compliance.
- Provide a small, synthetic shipboard demo dataset and manual walkthrough of the three requested scenarios.

## Capabilities

### New Capabilities

- `weekly-calendar`: View and manage the weekly calendar, preview edits, and review proposed schedule changes.
- `planning-data`: Persist departments, people, qualifications, duty commitments, assets, event requirements, and reusable typed constraints.
- `conflict-detection`: Evaluate candidate schedules and explain resource, personnel, timing, incompatibility, and fatigue conflicts.
- `schedule-resolution`: Generate feasible weekly plans and repair conflicts using explicit priorities and reviewable proposals.
- `department-agents`: Coordinate departmental agents through backend tools and retain an understandable negotiation summary.
- `training-readiness`: Track configured training obligations, completion history, and projected coverage from the weekly plan.
- `training-document-import`: Retain original training PDFs and review source-cited suggestions before creating training requirements, event templates, or constraints.
- `local-runtime`: Start the application locally with one command, automatic first-run setup, persistent data, and optional live AI configuration.

### Modified Capabilities

None. This is a new project with no existing application or specifications.

## Impact

- Introduces a frontend, backend API, relational schema, scheduling engine, and server-side OpenAI integration.
- Uses standard OpenSpec Markdown artifacts to guide implementation in Codex. `design.md` defines the proposed architecture and `tasks.md` defines the implementation sequence.
- Hackathon scope: one ship, one planning team, one ship-local display timezone, one week at a time, synthetic data, and a small scheduling problem. No automated tests are required; validate through manual demo scenarios.
- External calendar synchronization, authentication/role management, live collaborative editing, production deployment hardening, arbitrary executable rules, and comprehensive edge-case handling are outside this change.
- No application code, branch, commit, or deployment is created as part of preparing this proposal.
