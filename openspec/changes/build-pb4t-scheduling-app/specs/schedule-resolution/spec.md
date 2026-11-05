## ADDED Requirements

### Requirement: Generate a weekly plan
The system SHALL accept the selected week's desired events, fixed commitments, constraints, roster, assets, and training obligations and search for a feasible plan. It SHALL return event placements, concrete personnel assignments, validation results, and projected training coverage. Compatible simultaneous events SHALL remain permissible.

#### Scenario: Generate a conflict-free week
- **WHEN** a planner submits a small solvable dataset with competing departmental drills
- **THEN** the system proposes placements and assignments satisfying every enforced hard rule and in-scope training obligation
- **AND** fixed events retain their saved placements

### Requirement: Repair a requested create or move
The event detail sidebar SHALL resolve a proposed new or moved event by searching placements for that event only. It SHALL preserve every other event's placement and metadata, including draft status, and SHALL NOT add missing training occurrences. AI negotiation SHALL only suggest placements for the selected event. Existing training coverage gaps SHALL remain visible without requiring unrelated events to be added during this repair. Weekly resolution SHALL continue to search the full plan and enforce mandatory coverage.

#### Scenario: Repair by moving only the selected event
- **WHEN** a planner selects Resolve in an event's detail sidebar and the event has another valid slot
- **THEN** the proposal can move only the selected event while retaining every other event unchanged
- **AND** if no feasible placement is found, the result reports blockers without falling back to moving other events
- **AND** unrelated weekly conflicts do not block finding or applying a placement that satisfies the selected event's constraints
- **AND** unrelated conflicts remain visible and staffing records for unrelated unresolved events are retained

### Requirement: Preserve hard rules and minimize disruption
The system SHALL reject candidates violating hard constraints or mandatory coverage. Among complete feasible candidates it SHALL first minimize changed placements by tier, comparing Tier 1 (highest priority), then Tier 2, Tier 3, and Tier 4 (lowest priority). Changed placements SHALL be measured against saved starts for existing events and requested starts for new events. Equal tier move counts SHALL prefer fewer existing events moved, then smaller time deviation, then fewer soft preference violations. Department agents SHALL use the same tier priority guidance. It SHALL NOT silently drop desired events, move fixed commitments, or relax hard constraints.

#### Scenario: Choose between feasible repairs
- **WHEN** one feasible repair moves one existing event and another moves three of the same tier
- **THEN** the one-event repair ranks ahead of the three-event repair

#### Scenario: Preserve higher-priority events
- **WHEN** one feasible weekly repair moves a Tier 1 event and another retains it by moving multiple Tier 4 events
- **THEN** the repair moving Tier 4 events ranks ahead of the repair moving the Tier 1 event
- **AND** higher-priority events may still move when required to find a feasible plan

### Requirement: Report search outcomes honestly
The system SHALL bound search for the small hackathon dataset and report `feasible`, `partial`, or `no_solution_found`. Partial results SHALL identify unscheduled events and blocking constraints. A bounded search that finds no complete plan SHALL NOT claim proof of infeasibility or global optimality.

#### Scenario: No complete plan is found
- **WHEN** the search cannot place every required event within its search bound
- **THEN** the result identifies unresolved work and known blockers
- **AND** it does not present the result as a complete conflict-free resolution

### Requirement: Persist and apply reviewable proposals
The system SHALL persist the resolution input snapshot, result, event and assignment differences, validation output, and agent summary. A complete resolution SHALL be applied only after user selection and backend revalidation, with placements and assignments saved atomically.

#### Scenario: Review a proposal after refreshing
- **WHEN** a completed resolution run is reopened after a page refresh
- **THEN** its proposed changes and explanations remain available without having modified the saved calendar

#### Scenario: Apply a validated complete result
- **WHEN** the planner applies a feasible proposal that passes backend revalidation
- **THEN** all proposed placements and assignments are saved together
- **AND** formerly draft events included in the complete plan become scheduled

### Requirement: Preserve ignored occurrences during resolution
Generation and repair SHALL retain ignored events, their complete saved metadata, and their saved placements. Ignored events SHALL not appear in resolution change suggestions or be automatically promoted from draft to scheduled. Proposal application SHALL reject changes to a saved ignored event. Direct anchored resolution of an ignored event SHALL be rejected with guidance to save it directly. An explicit user-selected waiver SHALL not count as silent relaxation of a hard constraint.

#### Scenario: Generate and apply with ignored events
- **WHEN** a plan contains ignored events and the planner generates and applies a proposal
- **THEN** ignored events remain unchanged on the calendar
- **AND** they are absent from proposed changes
- **AND** the summary acknowledges waived conflicts instead of claiming those events were staffed and validated
