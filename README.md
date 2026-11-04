# PB4T weekly planner

A local shipboard training calendar with personnel and asset conflict checks, explainable schedule repair, training coverage, and optional live departmental agents.

## Run locally

Install **Go 1.24+** and **Node.js 22.12+ with npm**, then run from this directory:

```sh
npm start
```

Open **http://localhost:8080** when the ready message appears. The command installs frontend dependencies when needed, builds React and Go, creates SQLite tables, and loads synthetic demo data into an empty database. The first run needs internet access for dependencies. No Docker, database service, or C compiler is needed. Later core planning runs require no external service. Stop with **Ctrl+C**.

Data lives in `./data/scheduler.db` and survives restarts. The included policies and personnel are illustrative, not verified SFTRM mandates.

To intentionally replace local planning data with the original synthetic dataset, stop the app and run:

```sh
npm run reset-demo
npm start
```

## Optional live AI

Core calendar editing, conflict checks, training coverage, and deterministic weekly generation work without a key. To enable real departmental negotiation and free-form constraint interpretation, copy `.env.example` to `.env`, set `OPENAI_API_KEY`, and restart. The key stays on the server. `OPENAI_MODEL` defaults to `gpt-5-mini`; use a Responses API model available to your account. Live AI needs internet access and API quota. API errors are displayed as failed runs; the application does not fabricate negotiation.

Optional settings in `.env` or environment variables:

| Setting | Default |
| --- | --- |
| `PORT` | `8080` |
| `DATABASE_PATH` | `./data/scheduler.db` |
| `SHIP_TIMEZONE` | `America/New_York` |
| `OPENAI_MODEL` | `gpt-5-mini` |

## Try the three planning flows

1. **Generate a week.** The initial calendar includes a personnel collision, competing simulator use, and a drill during incompatible flight quarters. Select **Generate week**, review proposed times, staffing, soft tradeoffs, and projected training. Select **Apply plan** to save it. Fixed ship operations stay in place. Without a key this is labeled deterministic scheduling; with a key, departments negotiate before the engine validates the plan.
2. **Add an event.** Select **Add event**, choose a reusable event template, and set its preferred start and allowed window. Department, point of contact, duration, tags, drill type, and constraints are copied from the template; expand **Event details and overrides** only when this occurrence needs different defaults. Attach existing constraints or create new ones. **Check placement** shows specific conflicts. Save a valid event, explicitly save a conflicting **draft**, or select **Resolve** to review alternatives. Contacts do not reserve personnel unless explicitly required.
3. **Move an event.** Drag a movable calendar event to another slot. The editor previews its proposed time. Select **Check placement**, inspect conflicts, and resolve them or save as draft. Closing the editor leaves the saved placement unchanged. Fixed events cannot be dragged.

**Training readiness** separates completion history from validated planned occurrences. Generation adds missing required occurrences from drill templates. Use **Record completion** in an event to create its completion record. A planned event is never automatically reported as completed.

**Planning inputs** maintains departments, qualifications, personnel and availability, duties, assets, constraints, event templates, requirements, and completion history. Event templates contain reusable defaults without dates. Each scheduled occurrence receives a copy; editing a template affects future occurrences. Existing events can be copied with **Save as new template**. Removing referenced records requires updating their attachments first. Empty availability means unavailable. Minimum rest uses only supplied duty/drill data and the configured policy.

Proposals are saved and can be reopened from **Saved proposals** after refresh. **Cancel proposal** leaves the calendar unchanged. If planning inputs change while a proposal is under review, generate a fresh proposal before applying.

## Implementation

- `cmd/scheduler`: local HTTP server and API
- `internal/planner`: typed domain data, joint staffing evaluation, training coverage, bounded scheduling search
- `internal/store`: SQLite entities, relationship tables, transactional writes, saved proposal snapshots
- `internal/agents`: server-side Responses API tool loop and departmental proposal rounds
- `web/src`: React/TypeScript calendar, event editor, planning forms, and proposal review
- `scripts/start.mjs`: one-command bootstrap and demo reset

The solver searches a seven-day horizon in 15-minute increments and favors fewer event moves, then smaller time deviations, then fewer preference violations. It searches new candidate slots on weekdays between 08:00 and 17:00, preserving requested placements when valid. Search is bounded, so a partial result is not proof of infeasibility or a globally optimal schedule. One person fills one role in an event. The hackathon app assumes one planner, one ship, and small synthetic inputs.

Build checks: `npm run build` (frontend typecheck and build), `go build ./...`, and `go vet ./...`. No automated test suite is included under the requested hackathon scope.

## Runtime logs

Backend timing logs go to the terminal's standard error stream. To capture startup, build, and backend logs together:

```sh
npm start 2>&1 | tee /tmp/pb4t.log
```

Stop with Ctrl+C and restart to load code changes. Resolution logs include `resolution run=...`, `department trace=run/department/round`, `openai trace=...`, and `solver run=...`. They report elapsed durations, HTTP status/request IDs, token counts, and tool names. Keys, prompts, tool arguments, tool results, and planning payloads are not logged. A long OpenAI request is visible immediately through its `phase=request` line and gets a duration when it completes. Frequent browser status checks are not logged.

Live week generation supplies context upfront and makes one structured proposal request per department per round, with up to four departments running concurrently. The coordinator merges and validates a complete round before requesting revisions from affected departments. With the four demo departments this uses four initial requests and at most four revision requests. The engine remains responsible for final hard-constraint validation.
