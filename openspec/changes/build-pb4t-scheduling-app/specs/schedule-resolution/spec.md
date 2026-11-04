## ADDED Requirements

### Requirement: Generate a weekly plan
The system SHALL accept the selected week's desired events, fixed commitments, constraints, roster, assets, and training obligations and search for a feasible plan. It SHALL return event placements, concrete personnel assignments, validation results, and projected training coverage. Compatible simultaneous events SHALL remain permissible.

#### Scenario: Generate a conflict-free week
- **WHEN** a planner submits a small solvable dataset with competing departmental drills
- **THEN** the system proposes placements and assignments satisfying every supplied hard rule and in-scope training obligation
- **AND** fixed events retain their saved placements

### Requirement: Repair a requested create or move
The system SHALL support resolution anchored on a proposed new or moved event. It SHALL first consider personnel reassignment, then movement of other movable events. It SHALL keep the user's requested placement fixed within the anchored repair and separately label any alternative that changes that requested placement.

#### Scenario: Repair by moving another drill
- **WHEN** a requested event placement conflicts with a movable drill and that drill has another valid slot
- **THEN** the system can propose moving the other drill while retaining the requested placement
- **AND** the proposal identifies the affected drill and reason for the change

### Requirement: Preserve hard rules and minimize disruption
The system SHALL reject candidates violating hard constraints or mandatory coverage. Among complete feasible candidates it SHALL prefer fewer existing events moved, then smaller time deviation, then fewer soft preference violations. It SHALL NOT silently drop desired events, move fixed commitments, or relax hard constraints.

#### Scenario: Choose between feasible repairs
- **WHEN** one feasible repair moves one existing event and another moves three
- **THEN** the one-event repair ranks ahead of the three-event repair

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
