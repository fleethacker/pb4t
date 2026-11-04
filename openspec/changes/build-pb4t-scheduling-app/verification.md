# Implementation verification

Date: 2026-11-04. Progress: 50/50 tasks complete. Implementation and live OpenAI verification are complete.

## Build and local operation

- Go backend builds and `go vet ./...` passes.
- Frontend TypeScript check and Vite production build pass through `npm run build` and `npm start`.
- Fresh startup with frontend dependencies absent: `npm start` installs them, builds both parts, initializes a new temporary SQLite database, and prints a working localhost URL.
- During the initial keyless check, core planning remained available and live features were explicitly labeled unconfigured. After local credentials were configured, the UI correctly enabled live features.
- Saved event placements and assignments survived process restart without duplicate seed data.
- `npm run reset-demo` restored the original seven-event synthetic calendar after the manual walkthrough.

## Manual API walkthrough

- Weekly generation returned a feasible eight-event plan with zero hard conflicts and zero configured training gaps. The added drill occurrence appears as an addition with no previous placement. Generation alone left the saved calendar unchanged.
- Applying a reviewed proposal persisted the complete placements and concrete personnel assignments.
- Creating an overlapping simulator drill returned asset and personnel conflicts. Explicit draft saving succeeded. Anchored repair retained the requested time, moved other events, and applied successfully.
- Moving a weapons drill into simulator use returned an asset conflict. Repair found a feasible plan retaining the requested placement. Leaving the proposal unapplied preserved the saved time.
- Independent concurrent events were accepted. Overlapping demands on the same two Fire Party Leaders were rejected by joint assignment.
- A fixed incompatible drill returned a partial result with unresolved work; applying it was rejected.
- Recording a damage-control completion changed coverage to one completed and one planned occurrence, without counting the completed linked event twice.

## Browser walkthrough

- Generated, reviewed, and applied a plan; calendar metrics updated to eight events, zero hard conflicts, and three of three obligations projected covered.
- Added a constrained event, viewed the specific simulator conflict, and explicitly saved a draft.
- Dragged a weapons event from 09:00 to 10:00; the editor displayed the proposed time and asset conflict. Closing retained the saved calendar position.
- Viewed separate completion/planned counts in Training readiness and editable qualifications/availability in Planning inputs.
- Inspected the narrow layout and desktop layout. Restored the original demo after walkthrough changes.

## Live OpenAI verification

- Loaded the key from the ignored local `.env` without printing it. Confirmed `.env` is ignored and the browser bundle contains no key configuration.
- Completed an actual live resolution run with all four departments: Combat Systems, Damage Control, Engineering, and Weapons. Recorded eight submissions across initial and revision rounds.
- The final live proposal was feasible with zero hard conflicts and zero configured training gaps. Verified its persisted results and department round summaries in the browser. The proposal remains saved for review without applying it to the calendar.
- Interpreted “This drill cannot overlap flight quarters” using the live API in a separate temporary demo. Confirmed the draft was not persisted before approval; after Confirm & attach, the rule was persisted as confirmed and the engine reported the expected flight-quarters incompatibility.
- The first live interpretation included an unrelated event alongside the correct flight tag. Tightened the prompt and normalized family restrictions to the tag alone. A second actual live request returned the intended flight-tag restriction without an unrelated event target. Go build and vet passed after that correction.
- Temporary confirmation checks did not add events or constraints to the main application database.

No automated test suite was added, per the requested hackathon scope. No branch, commit, push, or deployment was created.

## Reusable event templates

- Go build/vet and frontend typecheck/build pass after separating EventTemplate from dated Event occurrences.
- A manual model check verified training-gap generation retains template provenance, template/occurrence tag arrays are independent, and legacy dated template JSON decodes into timeless defaults. Loaded the existing application database successfully without resetting or changing its records.
- In an isolated local demo, selected a template in Add event, changed only its preferred start, and saved the inherited occurrence as a draft.
- Updated the template's duration and contact, verified the existing occurrence remained unchanged, and created a second occurrence with the new defaults and its own date.
- In the browser, authored a new reusable template with contact and constraint defaults without any date fields. It immediately appeared in the picker and populated the timing-focused occurrence form with details collapsed under overrides.
- Application data used for manual checks was isolated from the user's running app. No user server was stopped or restarted.

## Faster departmental negotiation

- Replayed the saved input snapshot from run `f1a39961e6d77c68` in an isolated local database. The model remained `gpt-5-mini`. The user's app and calendar were not restarted or changed.
- Baseline: 269.629 seconds, 43 OpenAI requests, eight department submissions, 0.722-second solver.
- Updated run `6d24d39ccc719327`: 51.738 seconds, six OpenAI requests/submissions, 0.737-second solver, feasible outcome, zero hard conflicts, zero training gaps. This is about 5.2 times faster in this single benchmark; live latency varies.
- All four initial department calls started concurrently. Merged validation requested a second round only from Combat Systems and Weapons. Damage Control and Engineering skipped unnecessary revisions.
- A controlled in-memory HTTP walkthrough, run with Go's race detector, verified exactly one forced submission request per department/round, four concurrent workers, identical per-round snapshots, combined-plan validation, stable history ordering, and targeted asset and personnel revisions. A forced request failure canceled the round without publishing a partial round. No races were reported.
- Go build/vet, frontend typecheck/build through the isolated startup, and strict OpenSpec validation passed.
- Failure outcomes use a separate persistence timeout so negotiation cancellation does not prevent the final failed status from being recorded.
- Manual benchmark artifacts and the controlled walkthrough are under the ignored `.cache/debug/` directory; no automated test suite was added under the hackathon scope.
