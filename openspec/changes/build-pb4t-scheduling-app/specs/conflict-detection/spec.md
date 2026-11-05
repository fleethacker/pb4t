## ADDED Requirements

### Requirement: Shared backend evaluation
The system SHALL use one deterministic backend evaluator for unsaved edits, saved working plans, agent candidates, generated schedules, and proposal application. It SHALL distinguish hard violations from soft preference warnings.

#### Scenario: Evaluate the same candidate through different callers
- **WHEN** the UI and a departmental agent evaluate the same candidate against the same planning snapshot
- **THEN** they receive the same hard-constraint result

### Requirement: Joint personnel feasibility
The evaluator SHALL find concrete qualified and available personnel assignments across overlapping events and required roles. A person SHALL fill at most one required role per event and SHALL NOT serve concurrent events or overlap a duty commitment. Named-participant requirements SHALL preserve the named person.

#### Scenario: Qualification counts conceal a collision
- **WHEN** two simultaneous events each need two Fire Party Leaders and only the same two people are qualified and available
- **THEN** the evaluator reports insufficient joint personnel coverage
- **AND** it does not declare both events feasible based on separate qualification counts

#### Scenario: Independent staffing allows overlap
- **WHEN** simultaneous events have distinct available qualified personnel and all other hard rules are satisfied
- **THEN** the evaluator returns a feasible set of concrete personnel assignments

#### Scenario: A personnel conflict does not mark an unrelated event
- **WHEN** two Monday drills cannot share the available qualified personnel and a Thursday steering drill is independently staffable outside their configured rest interval
- **THEN** the personnel conflict identifies the competing Monday group without including the Thursday event
- **AND** the Thursday event retains concrete assignments and has no unrelated personnel alert
- **AND** the overall plan remains infeasible until the genuine Monday conflict is resolved
- **AND** the event editor distinguishes this event's conflicts from other weekly-plan conflicts without bypassing plan validation when saving

#### Scenario: Report separate staffing groups independently
- **WHEN** independent groups have their own personnel shortages
- **THEN** each failed group receives its own personnel conflict with its relevant events, resources, and interval
- **AND** staffing groups include events linked by configured rest as well as actual overlap
- **AND** all roles of the same event are checked together
- **AND** agent revision feedback is minimized within each reported group and routed to that group's departments

### Requirement: Asset capacity enforcement
The evaluator SHALL ensure combined concurrent asset demand does not exceed available capacity and that requested assets are available for each event's full interval.

#### Scenario: Concurrent use exceeds capacity
- **WHEN** two overlapping events each require one simulator and only one simulator is available
- **THEN** the evaluator reports the shared asset, capacity, demand, and affected events

### Requirement: Time and incompatibility enforcement
The evaluator SHALL require the full event interval to fit applicable allowed windows, avoid excluded windows, and satisfy attached or shared incompatibility rules. Fixed ship operations SHALL participate in these checks.

#### Scenario: Event overlaps prohibited operations
- **WHEN** a drill tagged incompatible with flight quarters overlaps a fixed flight-quarters event
- **THEN** the evaluator reports the rule and both event intervals

### Requirement: Configured fatigue enforcement
The evaluator SHALL apply configured minimum-rest policies to actual personnel assignments using supplied fatigue-counting duty commitments and drill intervals. It SHALL use configured policy values rather than model-inferred standards.

#### Scenario: Insufficient rest after duty
- **WHEN** the only eligible person's prior duty ends too close to a proposed drill under the configured minimum-rest rule
- **THEN** that assignment is rejected and the feedback identifies the person, relevant intervals, and required rest

### Requirement: Explainable conflict results
Each conflict SHALL include its type, rule identifier where applicable, affected event identifiers, relevant personnel or asset identifiers, time interval, and a plain-language explanation. Unmet qualification demands SHALL include required count and eligible available personnel. Validity SHALL depend on backend checks rather than an agent's narrative.

#### Scenario: Show actionable personnel feedback
- **WHEN** an event cannot be staffed
- **THEN** the result explains the missing qualification coverage and which commitments restrict the candidate people
- **AND** the UI can link the explanation to the affected records
