## Context

This is a greenfield hackathon project. The repository has no application code or existing integrations. The primary user is a PB4T planner coordinating Engineering, Weapons, Combat Systems, Damage Control, and ship operations.

The core object is a weekly plan containing fixed commitments and desired training events. Concurrent events are legitimate when people, assets, and operating conditions permit them. The application must support both generating a plan and explaining the effect of one calendar edit.

Use standard OpenSpec Markdown to guide implementation in Codex. Preparing these artifacts does not authorize implementation in this chat while explore mode remains active.

## Goals / Non-Goals

**Goals:**

- Deliver the three user journeys: generate a weekly plan, add an event with conflict feedback, and move an event with conflict feedback.
- Persist planning data and calendar changes across application restarts.
- Start locally with one command after installing Go and Node.js/npm, without requiring Docker, a database service, or a C compiler.
- Make every conflict traceable to specific events, resources, people, or configured rules.
- Demonstrate departmental agents making substantive proposals against shared persisted data.
- Produce a reviewable, deterministically validated schedule with explicit personnel assignments.
- Show configured training obligations and their projected coverage.

**Non-Goals:**

- Automated tests, extensive edge-case handling, production hardening, authentication, multi-ship support, external calendar sync, and live concurrent editing.
- Independently determining official SFTRM policy or certifying compliance.
- Arbitrary natural-language rule execution, a guaranteed global optimum, or a general-purpose distributed agent platform.

## Decisions

### 1. Architecture: one backend and one database

Use a React frontend with TypeScript, a Go HTTP backend, and SQLite. Keep scheduling, conflict checks, persistence, and agent orchestration in one backend process. Serve the frontend and API as one application for the demo; the backend owns a local SQLite database file and server-side OpenAI credentials. Configure a durable database file path so ordinary application restarts preserve saved planning data. Use a SQLite driver that does not require a C compiler. No separate database service is required.

```text
Browser
  Weekly calendar | Event editor | Conflict/resolution panel
                          |
                       HTTP API
                          |
                    Go application
       +------------------+------------------+
       |                  |                  |
  Planning service   Scheduling engine   Agent coordinator
       |                  ^                  |
       |                  |           Departmental agents
       |                  +------ typed backend tools
       |
   SQLite file
```

Go packages should reflect these responsibilities with concrete types and small interfaces only where callers need them. Keep transport and database details out of scheduling logic. Follow [Effective Go](https://go.dev/doc/effective_go) rather than introducing generic repository/service frameworks.

Alternative considered: separate agent services and a dedicated solver service. They add integration work without improving a small hackathon demonstration.

### 2. Shared model: requirements are different from assignments

An event can require two people with a qualification without choosing those people in advance. The scheduling engine must select actual qualified, available people when evaluating feasibility. A point of contact is a contact field; it only consumes personnel capacity if that person is also explicitly required or assigned.

| Entity | Essential information |
| --- | --- |
| Department | Name; agent responsibility |
| Person | Display name; department; availability windows |
| Qualification / person qualification | Qualification name; people holding it |
| Duty commitment | Person; start/end; description; whether it contributes to fatigue |
| Asset | Name; capacity; availability; optional department |
| Event | Title; department; contact name/reference; type/tags; duration; requested start; current start; allowed scheduling window; fixed flag; draft/scheduled/completed state |
| Constraint definition | Name; type; structured parameters; scope; human-readable explanation |
| Event constraint | Event-to-constraint attachment; multiple attachments allowed |
| Personnel assignment | Event; person; qualification role being filled |
| Event template | Name; drill type; department; point of contact; default duration; tags; constraint attachments; fatigue setting; no occurrence dates |
| Training requirement | Drill type; responsible department; explicit coverage-window start/end; required count; source reference; optional repeat interval for generating later windows |
| Completion | Drill type; department; completion time; linked event if present |
| Schedule proposal | Week; input snapshot; proposed placements and assignments; conflict report; changes; score; coverage; negotiation summary; status |

Persist core entities and relationships relationally in SQLite. Store type-specific constraint parameters and proposal snapshots/results as JSON text in TEXT columns, with encoding and decoding handled by Go. Constraints reference stable entity IDs, not names embedded in prose. Agents use the same backend services as the UI and never access raw SQL.

Availability and duty commitments are part of the input, not inferred from a person's department. Model fixed ship operations as fixed events and/or shared operating windows.

### 3. Constraints: a small typed rule catalog

Events can attach any number of existing constraints or create a new one in the editor. Shared operating and fatigue policies apply at plan scope. Initial rule types:

| Type | Example | Enforcement |
| --- | --- | --- |
| Qualified personnel | Two Fire Party Leaders | Assign two distinct qualified, available people |
| Named participant | A named evaluator must attend | Reserve that person for the event |
| Asset demand | One simulator or two radios | Concurrent demand cannot exceed capacity |
| Allowed/excluded window | Flight operations only in configured windows | Entire event must fit allowed time and avoid excluded time |
| Incompatibility | Flight quarters cannot overlap a damage-control drill | Match event IDs or event tags and prohibit overlap |
| Minimum rest | Configured rest after a fatigue-counting duty | Check assignments against supplied duty and drill intervals |

Hard rules are mandatory. Soft preferences are stored separately: preferred start time, avoiding changes to existing events, and departmental preferences. The solver never silently converts hard rules into preferences.

For this demo, one person fills at most one required role in an event and cannot be assigned to concurrent events. This keeps staffing semantics explicit. Seeded rosters and qualification demands must reflect that assumption.

Free-form text is an input convenience: an agent translates text into one or more supported typed constraints, the UI shows its interpretation, and the user confirms before attaching them. Unsupported interpretations remain explicitly unenforced notes and cannot qualify a proposal as fully validated until removed or replaced with a supported rule.

Alternative considered: ask the model to evaluate arbitrary prose directly. That would make identical calendar edits difficult to validate consistently.

### 4. Deterministic engine: one checker shared by all workflows

The engine receives the planning snapshot, candidate event placements, and staffing requirements. It returns concrete assignments or an explicit conflict report. It checks the full candidate plan when a person can be reassigned; event-by-event qualification counts alone are insufficient.

For example, two simultaneous drills can each have two eligible people but still compete for the same two people. The engine must find a joint assignment across all overlapping events and all required roles. Reuse that assignment search during both conflict checks and scheduling.

Each conflict includes rule type, constraint ID where applicable, affected event IDs, relevant person/asset IDs, the conflicting interval, and a plain-language explanation. For an unmet qualification demand, report required count and eligible available people. Hard conflicts and soft preference warnings are visually distinct.

Use the checker for event creation, moves, agent proposals, generated plans, and final proposal application. The frontend never implements a competing conflict engine.

### 5. Scheduling: search for a feasible plan, then improve it

For a small seeded dataset, implement a bounded backtracking search in Go over 15-minute start slots within a seven-day planning horizon. Event durations in the demo use 15-minute increments. Preserve fixed events and duty commitments. Schedule the most constrained events first, pruning candidates through the shared checker and searching personnel assignments jointly as needed.

Support two modes:

- **Generate week:** place all selected desired events and any missing event occurrences generated from configured training obligations.
- **Repair edit:** hold the user's requested create/move placement fixed for that proposal, first try personnel reassignment, then move the fewest other movable events. If necessary, offer a separately labeled alternative that changes the user's requested time.

Rank complete feasible plans lexicographically:

1. Fewer existing events moved.
2. Smaller total deviation from requested/current times.
3. Fewer soft preference violations, including departmental preferences.

Hard-rule satisfaction and required training coverage are feasibility gates, not penalties that can be traded away. A scheduled event counts once toward its matching drill-type/department requirement in the applicable explicit window.

Bound the search for demo responsiveness and report `feasible`, `partial`, or `no_solution_found`. Return placements and reasons for unscheduled work with partial results. A bounded search failing to find a plan does not prove mathematical infeasibility. Label successful output “recommended feasible plan” rather than claiming a proven global optimum.

Only a fully validated feasible proposal can be applied as a complete resolution. Users can change the inputs and retry when events remain unscheduled. Do not silently drop events or relax policies.

Alternative considered: introduce a separate optimization runtime immediately. Keep the engine boundary explicit so that can happen later, but avoid the extra language/service for the initial small demo.

### 6. Agents: parallel structured proposals with shared validation

When live AI is configured, the Go backend coordinates separate departmental model invocations. Each receives the round's frozen planning snapshot upfront: week events, departments, rosters, qualifications, duties, assets, constraints, training requirements, concise prior proposals, and actionable feedback. The only departmental tool is a forced `submit_department_proposal` call. Each participating department makes exactly one model request per round, with at most one preferred start for each owned movable event. The backend checks ownership, fixed/completed events, duplicate entries, and domain references. Calendar writes remain a separate user action.

Run at most four departmental requests concurrently. After collecting a complete round in deterministic department order, merge proposed placements into one shared candidate and validate it through the scheduling engine. Only then begin the revision round; every revising department sees the same merged candidate and prior-round summaries. Candidate changes never leak between concurrent calls in one round. Persist progress as completed-department counts and persist the merged validation results with proposal history.

Route direct conflicts only to owners of affected movable events. The current joint staffing checker reports all requested roles on failure, so derive a minimal infeasible staffing group before routing that feedback. This selects a demonstrated staffing blocker rather than treating every rostered event as responsible. At most one revision round is performed; the final deterministic search still checks every hard rule, including any additional unresolved staffing groups. Fixed requested placements in edit repair are locked in the negotiation snapshot.

Later revisions override earlier preferences for the same event. An empty revision retains the merged candidate rather than reverting previous placements. A failure cancels remaining calls in the round and records an explicit failed run. Outcome persistence has a separate short timeout so an exhausted negotiation deadline can still be saved.

Backend search uses the collected preferences for ranking and returns a fully validated result. Agents do not certify feasibility. Natural-language constraint assistance retains its separate confirmation-based tool workflow. Without a key, deterministic planning remains available and live features are labeled unavailable.

See [OpenAI function calling documentation](https://developers.openai.com/api/docs/guides/function-calling) for forced structured tools. Local logs record request durations, token counts, department/round traces, and solver timing without prompt or planning payloads.

### 7. Frontend: a calendar with an adjacent explanation panel

```text
+--------------------------------------------------------------------+
| PB4T     < Selected week >     [Add event] [Generate week]            |
| Department filters         Training coverage: covered / gaps         |
+-------------------------------------------+------------------------+
| Weekly calendar                           | Selection / feedback   |
|                                           |                        |
| Fixed operations                          | Event details          |
| Department-colored drills                 | Constraint chips       |
| Overlapping events visible side by side   | Specific conflicts     |
| Draft events use a distinct treatment     | Suggested alternatives |
|                                           |                        |
| Proposed moves shown before application  | [Resolve] [Apply plan] |
+-------------------------------------------+------------------------+
```

Event editor fields: name, department, point of contact, type, duration, requested start, allowed window, fixed/movable setting, and repeatable constraint rows. Each constraint row allows selecting an existing definition or creating a supported new one. Provide a free-form assist option with confirmation of the interpreted rule.

On create or drag, show the proposed placement immediately as a preview and request backend validation. A valid change can be saved. A conflicting change offers alternatives and an explicit “Save as draft” action; it cannot silently become a conflict-free scheduled event. Timed drafts appear in the working calendar and participate in subsequent whole-plan conflict evaluation and resolution. Drafts do not satisfy projected training coverage until placed and validated in a feasible plan.

The resolution panel shows before/after times, personnel reassignment, conflicts resolved, soft tradeoffs, and training coverage. Apply is a separate user action. Cancel leaves the saved plan unchanged. Fixed events remain visibly locked.

### 8. API and persistence contracts

Keep ordinary planning operations synchronous and agent-assisted resolution as a persisted run that the UI polls. No queue infrastructure is needed for the demo.

| API area | Operations |
| --- | --- |
| Week context | Fetch saved events, drafts, operating windows, and coverage for a week |
| Planning data | List/create/update/delete departments, people, qualifications, availability, duty commitments, assets, constraints, drill templates, requirements, and completion records |
| Events | Create/update/delete event records, attached constraints, and saved assignments |
| Validation | Evaluate an unsaved create/move or complete candidate plan; return conflicts, warnings, assignments, and coverage |
| Resolution runs | Start generation/repair; fetch progress, departmental proposals, and final proposal |
| Proposal application | Revalidate the proposed plan against saved inputs and atomically save placements and personnel assignments |
| Free-form assist | Translate text into supported constraint drafts for user confirmation |

SQLite is the source of truth for both browser and agents through the Go backend. Applying a proposal uses one transaction. Keep database transactions short; perform agent API calls and scheduling searches outside write transactions. Persist run status, the input snapshot, departmental proposal summaries, and the final result so it remains reviewable after refresh. Reloading the calendar must reproduce saved events, constraints, and assignments.

### 9. Training periodicities are configured data

Represent each requirement as a drill type, department, explicit coverage window, and required number of completed occurrences. A repeat interval can create subsequent windows, but the demo evaluates only obligations relevant to the selected week and their supplied history.

Show three different values: completed count, planned count, and remaining gap. Completion records alone establish completed coverage. A feasible future event can cover a projected gap but does not mark it complete. If an obligation is due within the selected week, generation creates enough missing desired occurrences to cover that obligation, accounting for existing matching events and completions without double counting.

The user supplies the authoritative source reference and periodicities. Seed values are labeled illustrative and are not represented as actual SFTRM policy. Baseline duty and completion history must include the intervals necessary to evaluate the demo's configured rest and coverage windows.

### 10. Local startup: one command and one URL

The primary local run command is `npm start` from the repository root. After Go and Node.js/npm are installed, that command prepares project dependencies when needed, builds the React frontend and Go backend, initializes the database schema, and starts the backend serving both the UI and API at `http://localhost:8080`. Print that URL when the application is ready. First-run dependency downloads and live AI require network access; an already prepared build must support core use without external services.

Use `./data/scheduler.db` as the default persistent SQLite file. Create the directory and schema automatically and load the synthetic demo dataset only when initializing an empty database. Do not overwrite saved planning changes or duplicate seed records on subsequent starts. Provide an explicit `npm run reset-demo` command to restore the synthetic dataset when the user chooses to reset it.

Provide a documented `.env.example` and automatically load an optional local `.env`. Keep `OPENAI_API_KEY` server-side; missing credentials must not prevent core local startup. Support optional port and database-path overrides, with working defaults. Exclude local credentials and database files from version control. Startup must not require manual SQL, separate frontend/backend terminal sessions, Docker, a cloud account, or a key for the core planning flows.

Document prerequisites, the start command, the ready URL, optional AI configuration, the database location, and the reset command in a short root README. A hot-reloading development command can be added later; it is not required for the initial runnable demo.

## Risks / Trade-offs

- [Search grows quickly with events and staffing choices] → Keep the demo small and bounded; report partial results honestly.
- [Natural language is ambiguous] → Confirm supported structured rules before enforcement.
- [Agents propose invalid tradeoffs] → Validate all proposals through the shared engine and retain hard-rule authority in backend data.
- [Training policy details are not supplied] → Use explicit user-configured windows with source references and visibly illustrative seed values.
- [A feasible calendar may still be undesirable to planners] → Show the full diff and allow review before applying.

## Migration Plan

There is no existing application or data to migrate. Implementation in Codex should automate SQLite schema setup and first-run synthetic data loading through `npm start`, and document the database file location. For the hackathon, explicitly resetting the demo database with `npm run reset-demo` is sufficient recovery; preserve applied changes during ordinary application restarts.

## Open Questions

These do not block the initial synthetic demo:

- What authoritative training requirement catalog and fatigue policy should replace illustrative values?
- Which real roster, qualification, and duty data sources would eventually feed the application?
- Should one person be permitted to fill multiple roles in a single event under future staffing policies?

Defaults for this proposal: one ship, one planner, a configured ship-local timezone, a 15-minute grid, small synthetic rosters, and the proposed Go/SQLite/React stack, built in Codex.

### Template-first event requests

Reusable event templates have a dedicated type rather than reusing the dated Event type. Each occurrence snapshots the template defaults and records its source template ID. Scheduling dates, windows, fixed placement, and completion state belong to the occurrence. Template edits do not rewrite previously created occurrences. Existing JSON template records remain readable because obsolete timing fields are ignored on decoding; no database reset is required.

Add event opens a template picker. The selected template fills the occurrence defaults. Timing is shown first and inherited details are available in an expandable overrides section. Planning inputs exposes Event templates with Edit and Schedule event actions; existing events can be copied into a new template. Training-gap generation uses the same template-to-occurrence conversion.

### Training document imports

Store original PDFs in a separate SQLite documents table with metadata and persisted suggestions. Keep these outside Data/scheduling snapshots so importing a reference alone does not alter a pending calendar proposal. Native Responses API PDF input avoids adding a local converter or Python dependency to startup. Extraction is a single forced structured call with explicit source-data boundaries and no action tools. Preserve source pages/excerpts and distinguish advisory recommendations from stated obligations; missing dates, counts, ownership, and durations require review.

Upload persists before analysis. The document library polls analysis status, pausing on errors. Failed or interrupted analysis retains the original and can be retried. Review edits persist independently. Explicit creation validates through the existing planning model and atomically saves the new record and its acceptance marker. Source references propagate to requirements/templates and events created from those templates. The original PDF is served as application/pdf from SQLite with inline filename and range support. Upload limit is 20 MB. Core retention works without a key; suggestions require the existing server-side OpenAI configuration.

Extraction retains duration evidence and leaves durations unsupported by evidence unspecified. Prompts distinguish total training duration from video seek offsets, and recommended prerequisite training from administrative access conditions. Requirement/template pairs identified as the same training activity are normalized to the requirement's drill type so projected coverage uses a shared activity identifier. These remain editable drafts requiring user review; the importer does not certify source accuracy or compliance.

Constraint suggestions are a third suggestion kind with a nested Rule draft and shared source evidence. One structured request returns separate training-item and constraint-suggestion arrays, avoiding irrelevant rule fields on every course. The extraction uses medium reasoning effort to preserve classifications and split compound restrictions into individual primitive rules. The existing rule schema and frontend rule-field component are shared with free-form assistance and planning inputs. Extraction sees existing resource IDs, removes invalid resource mappings, forces drafts unconfirmed and event-scoped, and keeps unsupported access/completion conditions as notes. Explicit creation validates and atomically saves acceptance, assigns a new rule ID, and preserves a source reference. Supported rules require user confirmation; notes are always stored unconfirmed. Event-scoped rules are attached separately through existing template/event editors. The deterministic evaluator is unchanged.

For doctrine imports, preserve currentness notices and distinguish actual training policy from operational response procedures and sample agreement language. Conditional triggers and calendar-based frequencies not faithfully represented by fixed-day recurrence remain review notes rather than invented dates. Conflicting thresholds are flagged for review. Importing an original reference never certifies that its policy is current.

Only the uploaded PDF is documentary evidence. The model's mapping context is a minimal index of department, qualification, person, asset, and event IDs/names, plus event tags. Prior templates, requirements, policies, source excerpts, durations, and availability are omitted to prevent cross-document contamination. Physical PDF page indices include cover/advisory pages rather than printed chapter-page numbering. Source excerpts and interpretations remain drafts that must be checked by the planner.

After extraction, a pure-Go PDF reader (`rsc.io/pdf`) matches normalized literal excerpts to physical pages and corrects page links when found. Normalization ignores case, whitespace, punctuation, and common fi/fl ligatures. Unmatched excerpts or unsupported/scanned text retain an explicit source-review notice; failure to read PDF text does not discard the original or the suggestion drafts. Matching checks quotation location, not policy interpretation, applicability, or compliance. No additional installed runtime or converter is required.

Literal ellipsis segments can match in order across adjacent pages, with the source link anchored to the first segment's page. Abbreviated/paraphrased wording is not fuzzy-matched. The library offers a local Recheck citations action, and uncreated drafts expose editable Source citation fields. Saving a draft checks its quotation and clears stale generated citation notices. The PDF/document identity stays fixed; AI-derived quotation text and page can be corrected. Unmatched excerpts use advisory styling rather than upload-error styling.

Review & edit offers Delete suggestion for uncreated drafts. Persist a deleted flag rather than removing the saved draft, filter it from visible cards/pending counts, and offer Undo for the most recent deletion. Undo restores the saved representation through the existing draft-update API. Deletion changes document metadata only. Accepted suggestions are protected from this action because their records are managed through Planning inputs; deleted drafts must be restored before creation.

### Compound free-form constraint descriptions

The event editor supplies its current event/template snapshot to interpretation, including unsaved records. One forced structured request returns independent ConstraintDraft values, each with a typed rule, the user's requested qualification term, and a separate explanation. The editor renders and confirms/cancels each draft independently. Rule names stay short; attachment context is separate from an incompatibility target (the prohibited other event/tag).

Qualification mappings are normalized against the catalog and the literal requested term, tolerating case, spacing/punctuation, and singular/plural differences. An observer is not inferred equivalent to a Drill Evaluator. Unknown roles retain empty mappings and counts for review instead of converting the entire compound request to a note. Invalid mappings are rejected on confirmation through existing domain validation. Interpretation itself neither persists rules nor attaches them. The former single-rule general tool loop is replaced by one forced submission request.

### Proposed calendar overlay

For feasible/partial proposals matching the displayed week, retain saved placements and add proposed changed/new placements in the same calendar grid. Original positions are outlined/hatched, proposed changes are filled and labeled, and unchanged events/fixed commitments appear once. Use separate render keys for current/proposed representations while preserving actual event IDs. The combined lane layout keeps simultaneous cards independently visible; the time range spans both layers, and department filters apply to both.

Hide/show toggles only the comparison layer. Preview cards display placement details without opening an editable event or starting a drag. Closing the proposal returns to the saved view; applying uses the existing validated API and removes the comparison after saved data reloads. Proposals and their overlay/panel are scoped to their week. The overlay helper is pure and never modifies the saved events or proposal records.

Original-position outlines use the event's department color. Proposal change rows share an event-ID highlight with the calendar on pointer hover or keyboard focus. Both original/proposed placements receive a department-colored ring, and unrelated cards are subdued. New occurrences highlight their single proposed placement. Hover takes precedence over keyboard focus; clearing both, closing the panel, or ending comparison clears the highlight. Highlighting does not alter layout, filter selections, or saved data.

The workspace heading leads directly into page content without the four global summary cards. Existing detailed planning views retain conflict, training, and proposal information; the calendar uses the freed vertical space.

Calendar buttons are CSS size-query containers with a name-first inner content layout. Rendered width and height determine whether time, department, and textual state labels fit, rather than using duration alone. Narrow cards hide time; the narrowest cards use a single primary state marker. Short cards retain up to two title lines with a compact marker, while tall cards allow three. Department borders and current/proposed styling remain intact. Full timing, department, contact, fixed/conflict/draft state, and any supplied description are retained in the accessible name and native title tooltip. This requires no layout observers, API changes, or changes to drag/edit/proposal behavior.

### Expanded synthetic planning examples

The seed includes all twelve event names from the supplied Events CSV plus a radar tracking check, preserving the earlier drill examples. Map CSV tag IDs to readable radio, radar, and mess-deck restriction/needs tags. Restriction rules target the opposing needs tag; needs rules reserve the relevant asset. Attach both tags and rules to events and templates so the existing deterministic evaluator handles incompatibility and resource capacity without AI. Add distinct Observer, Medical responder, Diver, Aloft supervisor, and Communications operator qualifications with synthetic personnel rather than aliasing unrelated roles.

All new durations, staffing, resource quantities, and training frequencies are illustrative defaults. The sample intentionally includes fixable tag conflicts alongside the existing personnel/simulator collisions. Seed occurrences link to their reusable templates, and template descriptions preserve the demo assumptions. This only changes fresh initialization and explicit resets; existing databases are never migrated or overwritten. A separate database and port provide an easy way to try the expanded sample alongside saved work.

All seven day columns use the same background; Saturday and Sunday have no separate visual treatment. Add Saturday communications and engineering occurrences and Sunday damage-control and combat-systems occurrences using existing template types, tags, and constraints. Weekend occurrences extend their allowed scheduling window through Sunday 17:00 rather than inheriting the weekday cutoff. Their valid requested weekend times remain candidates in the existing solver. The expanded seed contains 24 events and 19 templates. Calendar cards retain the committed adaptive content, native details tooltip, and direct editor interaction; the requested read-only popup was reverted before these changes.

### Scoped personnel conflict reporting

The former week-wide staffing search reported every staffed event when any assignment failed. Build independent groups from role slots instead: all roles for an event stay together, and different events connect only when an eligible person could be required in both during overlapping times or configured rest intervals. Apply the same competition predicate to grouping and assignment. Search each group jointly; retain concrete assignments for successful groups and emit a separate personnel conflict for each failed group, including only its events/resources and its actual time range. Qualification eligibility, duties, historical rest, and the one-person-per-role policy remain enforced.

Departmental feedback minimizes each personnel report only within its affected event IDs, preserving independent conflicts and revision routing. The event editor filters its conflict list to the selected event and reports the other plan-conflict count separately. Its save decision and backend validation continue to use the complete evaluation. No calendar data or constraint definitions are rewritten by this correction.

### Docker and Helm deployment

Build React with Node.js and a static Go binary in separate Docker stages; copy only the binary, frontend assets, and an owned empty data directory into a non-root distroless Debian 13 runtime with certificate roots. BuildKit target OS/architecture arguments support ARM64 and AMD64. Keep local `.env` files, databases, caches, and generated assets out of the build context. The image serves both UI and API from `/app` and persists SQLite at `/data/scheduler.db`. `LISTEN_HOST` defaults to localhost for native operation and is set to all container interfaces by Docker/Helm; use `net.JoinHostPort` to construct the listener address.

Compose publishes the service on host localhost, mounts a named data volume and writable temporary filesystem, and supplies optional runtime settings from the existing environment/`.env`. Its root filesystem is read-only. The Helm chart uses a one-replica Deployment with Recreate updates to avoid concurrent SQLite writers, a Service, HTTP startup/readiness/liveness probes, and a writable temporary volume. A retained ReadWriteOnce PVC is the default; an existing claim or explicit ephemeral volume is configurable. Pod UID/group/fsGroup match the image's non-root identity. Optional OpenAI credentials are referenced from an existing Secret rather than stored in values or generated manifests. Chart values have a schema, including rejection of unsupported scaling and inline API-key fields. No cluster is modified by the supplied offline validation commands.
