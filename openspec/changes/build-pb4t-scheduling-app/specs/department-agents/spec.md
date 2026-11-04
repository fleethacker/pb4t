## ADDED Requirements

### Requirement: Departmental scheduling participants
When live AI is configured, the system SHALL invoke distinct departmental agents with department-specific goals for the configured departments. Agents SHALL receive the same persisted planning snapshot for each round and submit concrete scheduling preferences or proposals. Without an API key, the system SHALL keep deterministic planning available, clearly identify live agent features as unavailable, and SHALL NOT present simulated negotiation as real agent activity.

#### Scenario: Departments compete for a resource
- **WHEN** Engineering and Combat Systems both request the same shared asset during overlapping periods
- **THEN** their agents can inspect that shared asset's capacity and submit alternative allocations with department-specific explanations

### Requirement: One structured proposal per department per round
The backend SHALL supply week context, rosters, duties, assets, constraints, training requirements, prior proposal summaries, and relevant conflicts upfront. Each participating department SHALL make exactly one model request per round using a forced structured submission. The backend SHALL validate event ownership, duplicate placements, fixed events, and domain references. Agents SHALL NOT execute database queries or apply calendar changes.

#### Scenario: Submit without data-fetch round trips
- **WHEN** a department is asked for a proposal
- **THEN** all required planning context is present in its request
- **AND** the only available tool submits one proposal, with at most one preferred start per owned movable event

### Requirement: Bounded substantive negotiation
The coordinator SHALL collect initial departmental proposals concurrently with at most four active requests, merge placements at a round barrier, validate the combined candidate, and allow one revision round only for departments with actionable violations on movable events. Every department in a round SHALL see the same candidate and prior-round summaries. Joint personnel reports SHALL be narrowed to an infeasible staffing group before routing, rather than treating every rostered event as affected. Departmental preferences and revisions SHALL influence candidate ordering or soft scoring while hard rules remain authoritative.

#### Scenario: Revise a conflicting proposal
- **WHEN** an initial Weapons proposal conflicts with a fixed operational window
- **THEN** the Weapons agent receives the specific violation and can propose a supported alternative
- **AND** the final engine search uses the revised preferences without moving the fixed operation

### Requirement: Explain the negotiated outcome
The system SHALL provide a concise persisted summary of departmental requests, accepted concessions, resulting schedule changes, and unresolved blockers. Claims in the summary SHALL correspond to recorded proposals and validation results.

#### Scenario: Review a departmental concession
- **WHEN** Combat Systems accepts a later drill so Engineering can use a constrained earlier window
- **THEN** the proposal summary identifies the changed drill, agreed time, and relevant constraint
- **AND** it does not describe unresolved hard conflicts as resolved

#### Scenario: Unrelated department skips revision
- **WHEN** combined proposals conflict only over an asset used by two departments
- **THEN** those two departments receive the shared conflict for revision
- **AND** other departments make no second-round model request
