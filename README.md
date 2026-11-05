# PB4T weekly planner

A local shipboard training calendar with personnel and asset conflict checks, explainable schedule repair, training coverage, and optional live departmental agents.

## Run locally

With [mise](https://mise.jdx.dev/getting-started.html) installed, run from this directory:

```sh
mise trust
mise install
mise run start
```

The project pins Go 1.26.6 and Node.js 24.21.0. The start task uses the existing bootstrap script, including optional local `.env` settings. No shell activation is needed when using `mise run`.

| Command | Purpose |
| --- | --- |
| `mise run deps` | Install dependencies from the frontend lockfile and Go modules |
| `mise run build` | Install dependencies and build React plus `bin/scheduler` |
| `mise run check` | Build/typecheck the app and run Go vet |
| `mise run reset-demo` | Replace local planning data with the demo dataset; stop the app first |
| `mise run helm:lint` | Validate the Helm chart without deploying |
| `mise run uds:lint` | Validate the UDS package without deploying |

Helm 4.3.0 and UDS CLI 0.39.0 are installed only when their optional tasks are run. Use `mise tasks` to list available tasks. Machine-specific overrides can go in ignored `mise.local.toml` or `mise.<environment>.local.toml` files.

Without mise, install **Go 1.24+** and **Node.js 22.12+ with npm**, then run from this directory:

```sh
npm start
```

Open **http://localhost:8080** when the ready message appears. The command installs frontend dependencies when needed, builds React and Go, creates SQLite tables, and loads synthetic demo data into an empty database. The first run needs internet access for dependencies. No Docker, database service, or C compiler is needed. Later core planning runs require no external service. Stop with **Ctrl+C**.

Data lives in `./data/scheduler.db` and survives restarts. The included policies and personnel are illustrative, not verified SFTRM mandates.

The expanded demo contains 24 calendar events, 19 reusable templates, 20 constraints, 10 personnel, six shared assets, and six illustrative training requirements. The supplied Events and Constraints CSVs inspired flight operations, aloft work, HF radio events, CSTT training, DC fire, steering loss, diving, communications checks, EMCON, mass casualty, chow, and ammo movement. A radar tracking check demonstrates the opposing radar tags. Saturday includes a 09:00 communications check and 11:00 engineering review; Sunday includes a 10:00 DC fire drill and 13:00 CSTT drill. All seven days share the same calendar background. Added durations, staffing, resources, and weekly training frequencies are editable demo assumptions.

The CSV tag IDs map to readable tags: `0001` → `no-radios`, `0002` → `needs-radios`, `0003` → `no-radars`, `0004` → `needs-radars`, `0005` → `no-mess-decks`, and `0006` → `needs-mess-decks`. Restriction rules prohibit overlap with the corresponding needs tag; needs rules reserve the relevant asset. The initial week deliberately includes these conflicts for deterministic resolution.

Existing databases retain their saved data when the demo changes. To try the expanded sample alongside your current app, run this command and open **http://localhost:8085**:

```sh
PORT=8085 DATABASE_PATH=./data/expanded-demo.db npm start
```

To intentionally replace local planning data with the updated synthetic dataset, stop the app and run:

```sh
npm run reset-demo
npm start
```

## Run with Docker

With Docker and Docker Compose available and the Docker engine running:

```sh
docker compose up --build -d
```

Open **http://localhost:8080**. Go and Node.js are built inside the image, so neither is needed on the host. The named `planner-data` volume stores SQLite and the original uploaded PDFs. `docker compose down` stops the app while retaining this volume. Containers run as a non-root user with a read-only root filesystem and writable data/temporary storage.

Compose uses the existing optional `.env` for `OPENAI_API_KEY`, `OPENAI_MODEL`, `SHIP_TIMEZONE`, and the published `PORT`; keys and local databases are excluded from image builds. Without a key, deterministic planning remains available. To use another host port, run `PORT=8087 docker compose up --build -d` and open that port. View logs with `docker compose logs -f planner`.

For Kubernetes, use the [Helm chart instructions](deploy/helm/pb4t/README.md). The chart runs one instance with a persistent SQLite volume and can reference an existing OpenAI-key Secret.

For an existing UDS Core cluster, use the [UDS package instructions](deploy/uds/README.md). The package adds basic Keycloak sign-in through UDS Core's Authservice for the whole UI and API, with an optional group restriction. Authenticated users share the same planner and editing permissions.

## Optional live AI

Core calendar editing, conflict checks, training coverage, and deterministic weekly generation work without a key. To enable real departmental negotiation and free-form constraint interpretation, copy `.env.example` to `.env`, set `OPENAI_API_KEY`, and restart. The key stays on the server. `OPENAI_MODEL` defaults to `gpt-5-mini`; use a Responses API model available to your account. Live AI needs internet access and API quota. API errors are displayed as failed runs; the application does not fabricate negotiation.

Optional settings in `.env` or environment variables:

| Setting | Default |
| --- | --- |
| `PORT` | `8080` |
| `LISTEN_HOST` | `127.0.0.1` (containers use `0.0.0.0`) |
| `DATABASE_PATH` | `./data/scheduler.db` |
| `SHIP_TIMEZONE` | `America/New_York` |
| `OPENAI_MODEL` | `gpt-5-mini` |

## Try the three planning flows

1. **Generate a week.** The initial calendar includes a personnel collision, competing simulator use, and a drill during incompatible flight quarters. Select **Generate week**, review proposed times, staffing, soft tradeoffs, and projected training. The proposal overlays the current calendar: outlined cards retain original positions, filled cards show proposed moves/new events, and unchanged events appear once. Use **Hide proposal / Show proposal** to compare. Select **Apply plan** to save it. Fixed ship operations stay in place. Without a key this is labeled deterministic scheduling; with a key, departments negotiate before the engine validates the plan.
2. **Add an event.** Select **Add event**, choose a reusable event template, and set its preferred start and allowed window. Department, point of contact, duration, tags, drill type, and constraints are copied from the template; expand **Event details and overrides** only when this occurrence needs different defaults. Attach existing constraints or create new ones. **Check placement** shows specific conflicts. Save a valid event, explicitly save a conflicting **draft**, or select **Resolve** to review alternatives. Contacts do not reserve personnel unless explicitly required.
3. **Move an event.** Drag a movable calendar event to another slot. The editor previews its proposed time. Select **Check placement**, inspect conflicts, and resolve them or save as draft. Closing the editor leaves the saved placement unchanged. Fixed events cannot be dragged.

**Training readiness** separates completion history from validated planned occurrences. Generation adds missing required occurrences from drill templates. Use **Record completion** in an event to create its completion record. A planned event is never automatically reported as completed.

**Planning inputs** maintains departments, qualifications, personnel and availability, duties, assets, constraints, event templates, requirements, and completion history. Event templates contain reusable defaults without dates. Each scheduled occurrence receives a copy; editing a template affects future occurrences. Existing events can be copied with **Save as new template**. Removing referenced records requires updating their attachments first. Empty availability means unavailable. Minimum rest uses only supplied duty/drill data and the configured policy.

Personnel checks group events that can compete for the same eligible people through overlap or configured rest. A staffing problem in one group does not mark unrelated events elsewhere in the week. The event editor shows conflicts involving that event and identifies other weekly-plan conflicts separately; saving as scheduled still requires a valid plan.

Proposals are saved and can be reopened from **Saved proposals** after refresh. **Cancel proposal** leaves the calendar unchanged. If planning inputs change while a proposal is under review, generate a fresh proposal before applying.

Original-time outlines use department colors. Hover over a change in the right-hand proposal pane to highlight both its current and proposed positions; keyboard focus does the same. Other calendar blocks are subdued while inspecting that change.

Calendar cards prioritize the event name and adapt to their actual width and height. Short or narrow cards hide secondary text and use compact status indicators; taller, wider cards add time, department, and state labels. Department colors remain visible. Native hover tooltips and accessible names retain full event details, including any description, even when the card text is shortened.

**Describe a constraint** can return several independent drafts. For example, “needs a EOOW and a observer” produces separate one-person qualification drafts. The event/template being edited supplies the attachment context. Review, edit, and **Confirm & attach** each draft independently. Explanations appear separately from short rule names. Unknown qualification names stay unmapped for review; “observer” is not assumed to mean “Drill Evaluator.” Choose an existing qualification or create the intended qualification under Planning inputs.

## Implementation

- `cmd/scheduler`: local HTTP server and API
- `internal/planner`: typed domain data, joint staffing evaluation, training coverage, bounded scheduling search
- `internal/store`: SQLite entities, relationship tables, transactional writes, saved proposal snapshots
- `internal/agents`: server-side structured Responses API submissions, constraint drafts, and departmental proposal rounds
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

Live week generation supplies context upfront and makes one structured proposal request per department per round, with up to four departments running concurrently. The coordinator merges and validates a complete round before requesting revisions from affected departments. With the five departments in the expanded demo this uses five initial requests and at most five revision requests. The engine remains responsible for final hard-constraint validation.

## Import training documents

Open **Training documents → Import PDF**. The original PDF (up to 20 MB) is retained in the same local SQLite database and can be reopened from the library. With an OpenAI key configured, the app proposes training requirements, event templates, and constraints with source pages, excerpts, recommendation/requirement labels, and missing-detail notes. Analysis sends the PDF to the configured OpenAI service; no local PDF converter is required.

Review and edit each draft, supply missing department/duration/count/window information, and explicitly select **Create event template** or **Create training requirement**. Nothing is added to planning inputs or the calendar automatically. Recommended courses are advisory; the application does not infer training periodicities or certify compliance. Accepted records retain source links and can be edited under Planning inputs. Template-based events carry their source reference forward.

Use **Delete suggestion** inside **Review & edit** to remove an unwanted draft; **Undo** restores the saved draft. Deleted suggestions stay hidden after refresh/restart. The original PDF is retained. Already-created items are managed through Planning inputs.

Original PDFs remain available without a key and when analysis fails. Restart-interrupted analysis is marked retryable. Document records are retained when resetting the synthetic planning dataset; the original file lives in SQLite rather than a separate uploads directory.

Constraint suggestions use the existing personnel, asset, time-window, incompatibility, and rest rule types. Review the fields and scope, confirm the rule, then select **Create constraint**. Attached-event rules are available to select on templates/events after creation; they are not automatically attached. Account access, prerequisite knowledge/completion, and other unsupported conditions can be saved as **unenforced notes**. They never become active scheduling rules. Created constraints retain their original source-page link. To obtain constraint suggestions for an older analyzed import, import another copy; saved drafts and accepted records are not overwritten.

Only the uploaded PDF supplies source evidence; other imports and template defaults are excluded from extraction context. A bundled Go PDF reader matches literal excerpts to physical pages when text is extractable, including cover/advisory pages. Unmatched excerpts are clearly marked for review. This checks quotations and page locations, not whether the AI interpretation or source policy is correct/current. Doctrine-specific conditional triggers and procedural response timers remain subject to review rather than becoming invented calendar policies.

**Recheck citations** updates source checks locally without another AI request. Literal quotations shortened with ellipses can span adjacent PDF pages. If the AI paraphrased a quotation, use **Review & edit → Source citation** to correct its wording/page from the original; saving checks it again and clears resolved notices.
