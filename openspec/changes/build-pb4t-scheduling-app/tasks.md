## 1. Application foundation

- [x] 1.1 Create the new Go backend and React/TypeScript frontend following the boundaries in design.md.
- [x] 1.2 Configure SQLite with a durable database file path, server-side OpenAI credentials, ship-local display timezone, and local application startup without a separate database service.
- [x] 1.3 Add SQLite schema setup and a repeatable synthetic dataset with a small solvable scheduling problem and labeled illustrative policies; store structured constraint parameters and proposal snapshots as JSON text.
- [x] 1.4 Provide root `npm start` to prepare project dependencies when needed, build both application parts, initialize SQLite/demo data on first launch, and serve the frontend and API at one printed localhost URL. Require only Go and Node.js/npm, with no Docker, database server, or C compiler.
- [x] 1.5 Provide optional `.env` loading, `.env.example`, working port/database defaults, and a clearly labeled core planning mode that starts without an OpenAI key; keep credentials and local database files out of version control.
- [x] 1.6 Provide explicit `npm run reset-demo`; preserve existing data and avoid duplicate seeding on ordinary starts.

## 2. Persisted planning inputs

- [x] 2.1 Implement persistence and basic API operations for departments, people, qualifications, availability, and duty commitments.
- [x] 2.2 Implement assets with capacity and availability, reusable typed constraints, and shared operating/fatigue policies.
- [x] 2.3 Implement events, multiple constraint attachments, fixed/movable settings, draft/scheduled/completed states, and personnel assignments.
- [x] 2.4 Implement drill templates, training requirements with explicit coverage windows, and completion history.
- [x] 2.5 Add simple frontend views for maintaining the planning inputs used by the demo.

## 3. Deterministic conflict evaluation

- [x] 3.1 Implement event window, excluded-window, and event-incompatibility evaluation against the full planning context.
- [x] 3.2 Implement combined asset-capacity and availability evaluation for concurrent events.
- [x] 3.3 Implement joint assignment of distinct qualified available people to required roles, honoring named participants and duty commitments.
- [x] 3.4 Add configured minimum-rest evaluation to personnel assignment using supplied duty and drill history.
- [x] 3.5 Expose candidate-plan validation with structured conflict explanations, soft warnings, and concrete feasible assignments.

## 4. Training coverage and schedule search

- [x] 4.1 Compute completed, planned, and missing training coverage without counting a completed linked event twice.
- [x] 4.2 Generate missing desired drill occurrences from templates for configured obligations due in the selected week.
- [x] 4.3 Implement bounded weekly search over 15-minute start slots, preserving fixed events and reusing the conflict and staffing engine.
- [x] 4.4 Implement anchored create/move repair with personnel reassignment first, then changes to other movable events, and separately labeled alternate requested times.
- [x] 4.5 Rank feasible plans by events moved, time deviation, and soft preference violations; return honest partial/no-solution-found outcomes with unresolved work.
- [x] 4.6 Persist proposal snapshots, placements, assignments, differences, coverage, and validation output; revalidate and atomically apply complete feasible proposals.

## 5. Calendar and review interface

- [x] 5.1 Build the weekly department-colored calendar with visible concurrent events, fixed commitments, drafts, and department filters.
- [x] 5.2 Build the event editor with contact, duration, time, allowed window, flexibility, and repeatable select-existing/create-new constraint rows.
- [x] 5.3 Connect create and drag previews to backend validation; show specific conflicts and support explicit draft saving.
- [x] 5.4 Add Generate Week and Resolve actions with progress and proposed placements displayed before application.
- [x] 5.5 Build the proposal review panel with before/after times, assignment changes, tradeoffs, coverage, Apply, and Cancel.
- [x] 5.6 Add the training coverage summary and detailed outstanding-obligation view.

## 6. Departmental agent coordination

- [x] 6.1 Implement typed agent tools for week context, rosters, assets, candidate evaluation, and department proposal submission.
- [x] 6.2 Define department-specific agent goals and run the initial proposal round against one shared planning snapshot.
- [x] 6.3 Route engine-identified conflicts to affected agents for one revision round and feed resulting preferences into scheduling.
- [x] 6.4 Persist and display concise departmental requests, concessions, proposed changes, and unresolved blockers tied to actual engine results.
- [x] 6.5 Add free-form constraint assistance that proposes supported structured rules for confirmation and visibly marks unenforced text.

## 7. Manual hackathon walkthrough

- [x] 7.1 Demonstrate scenario 1: enter desired events and requirements, generate a complete feasible week, inspect coverage, and apply the result. Verify live agent concessions separately in task 7.8.
- [x] 7.2 Demonstrate scenario 2: add an event with personnel, asset, and timing constraints; inspect conflicts and apply a valid alternative or explicitly save a draft.
- [x] 7.3 Demonstrate scenario 3: drag an event into conflict, inspect the explanation, review a repair, and apply or cancel it.
- [x] 7.4 Manually confirm independent overlapping events are allowed, shared qualified-person pools are evaluated jointly, and saved data survives restart.
- [x] 7.5 Manually confirm an unresolved requirement is shown honestly and a planned drill is not shown as completed training.
- [x] 7.6 Document local startup, synthetic-data loading, and the three demo flows. No automated tests or comprehensive edge-case work are required for this hackathon scope.
- [x] 7.7 Manually confirm a fresh checkout starts with `npm start` after installing only Go and Node.js/npm, without an OpenAI key; verify automatic setup, the printed URL, core planning features, restart persistence, and explicit demo reset.
- [x] 7.8 With a server-side OpenAI key configured, manually confirm live departmental negotiation and free-form constraint assistance are available and clearly distinguished from keyless core planning.

## 8. Reusable event templates

- [x] 8.1 Separate timeless event templates from dated occurrences, retaining compatibility with stored template records.
- [x] 8.2 Copy template defaults and provenance into independently scheduled occurrences, including training-gap generation.
- [x] 8.3 Add template selection to Add event and emphasize timing with optional inherited-detail overrides.
- [x] 8.4 Provide reusable template editing, direct scheduling, and copying an event into a new template.
- [x] 8.5 Manually verify template persistence, two independent occurrences, future-only template edits, and the browser scheduling flow.

## 9. Faster departmental negotiation

- [x] 9.1 Supply planning context upfront and force exactly one structured proposal request per participating department and round.
- [x] 9.2 Run up to four departments concurrently against an immutable round snapshot and merge placements before validation.
- [x] 9.3 Route actionable shared conflicts to affected movable-event owners and skip unrelated revisions.
- [x] 9.4 Verify concurrency, round isolation, merging, targeted revisions, cancellation, and a timed live feasible result.

## 10. Training document imports

- [x] 10.1 Persist original PDF bytes and metadata separately from planning snapshots, with original-file access and interrupted-analysis recovery.
- [x] 10.2 Analyze PDFs through one structured, source-cited AI request, treating source text as untrusted data and leaving unknown policy values for review.
- [x] 10.3 Add the document library, import progress, editable persisted suggestion drafts, and explicit validated creation of requirements/templates.
- [x] 10.4 Preserve source-page references and descriptions through templates, requirements, and template-based events.
- [x] 10.5 Manually verify the sample PDF, draft edits, record creation, original-byte retention, no automatic planning changes, and restart persistence.

## 11. Document constraint suggestions

- [x] 11.1 Extract source-cited, unconfirmed typed constraint drafts alongside training requirements/templates, preserving unsupported conditions as unenforced notes.
- [x] 11.2 Provide editable rule review and explicit validated creation, preserving source references and existing template attachment workflows.
- [x] 11.3 Manually verify extraction, draft persistence, confirmation, note behavior, deterministic enforcement after attachment, and builds.

## 12. Delete import suggestions

- [x] 12.1 Add persistent suggestion deletion with Undo in Review & edit, retaining originals and protecting already-created planning records.
- [x] 12.2 Manually verify delete/Undo, refresh persistence, empty review state, and builds.

## 13. CIM citation review fixes

- [x] 13.1 Match ordered literal ellipsis segments across adjacent PDF pages, preserving review status for genuinely paraphrased excerpts.
- [x] 13.2 Add local citation rechecking, editable source citations, advisory wording, and clear stale citation notices after correction.
- [x] 13.3 Reproduce the saved CIM warnings, verify the correction flow, and repair the affected saved citations with a backup.

## 14. Compound constraint descriptions

- [x] 14.1 Return independent short-name rule drafts with separate guidance, event/template context, and conservative qualification mapping.
- [x] 14.2 Review and confirm individual drafts without losing other roles from the description.
- [x] 14.3 Verify the reported phrase, mapping safeguards, incompatibility regression, and browser confirmation flow.

## 15. Proposed calendar overlay

- [x] 15.1 Overlay proposed changes on current placements with distinct styling, one card for unchanged events, and comparison controls.
- [x] 15.2 Verify moved/new events, shared times, filtering, proposal lifecycle, and preservation of saved calendar data.

## 16. Linked proposal highlighting

- [x] 16.1 Use department colors for original-position outlines and link proposal-row hover/keyboard focus to both calendar placements.
- [x] 16.2 Verify colors, moved/new event highlights, pointer/focus clearing, and saved-data preservation.

## 17. Streamlined workspace header

- [x] 17.1 Remove the four top dashboard cards, their unused component/styles, and use the freed space for the calendar.

## 18. Expanded sample planning data

- [x] 18.1 Expand the synthetic roster, resources, constraints, templates, calendar, and illustrative requirements using the supplied CSV examples.
- [x] 18.2 Verify CSV tag mappings, deterministic conflicts, complete generation, persistence, and template defaults without changing saved user data.

## 19. Adaptive calendar card content

- [x] 19.1 Prioritize event names and adapt secondary metadata to each card's rendered width and height, preserving accessible details and status indicators.
- [x] 19.2 Verify short, narrow, roomy, fixed, draft, conflicting, and proposal cards in an isolated calendar.
