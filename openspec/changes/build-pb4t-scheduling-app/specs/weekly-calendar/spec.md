## ADDED Requirements

### Requirement: Weekly planning view
The application SHALL display saved events and timed drafts in a seven-day calendar using the configured ship-local timezone. It SHALL display concurrent events side by side, identify departments, distinguish drafts, and mark fixed events.

#### Scenario: View overlapping events
- **WHEN** a planner opens a week containing simultaneous Engineering and Combat Systems events
- **THEN** both events appear at their saved times with department identification
- **AND** overlap alone does not imply a conflict

### Requirement: Event management with multiple constraints
The application SHALL allow a planner to create, edit, and delete an event with a name, department, point of contact, drill type, duration, requested time, allowed scheduling window, fixed/movable setting, and zero or more constraint attachments. It SHALL allow selecting existing constraints or creating supported new constraints within the editor.

#### Scenario: Create an event with several constraints
- **WHEN** a planner enters an event requiring two qualified people, one asset, and an allowed time window
- **THEN** the editor retains all three constraints and submits them with the proposed event for validation

### Requirement: Validate before saving an event placement
The application SHALL preview an event create or move and obtain backend conflict feedback before saving it as scheduled. A conflicting placement SHALL offer explicit draft saving and resolution actions. Draft placements SHALL remain visible and participate in working-plan conflict evaluation, but SHALL NOT count as validated planned training coverage.

#### Scenario: Add a conflicting event
- **WHEN** a planner proposes an event that competes with an existing event for a required named participant
- **THEN** the application identifies the participant and affected events before saving
- **AND** the planner can request alternatives or explicitly save a visible conflicting draft

#### Scenario: Move an event into a conflict
- **WHEN** a planner drags an event into a time with insufficient asset capacity
- **THEN** the application previews the new position and displays the asset conflict
- **AND** the saved event retains its previous position unless the planner explicitly saves a draft or applies a valid alternative

### Requirement: Review schedule changes before application
The application SHALL present proposed event moves, personnel assignment changes, resolved conflicts, remaining issues, soft tradeoffs, and projected training coverage before the planner applies a proposal. Canceling SHALL preserve the saved calendar.

#### Scenario: Review and apply a feasible proposal
- **WHEN** a planner reviews a feasible proposal and selects Apply
- **THEN** the backend revalidates and saves the placements and assignments
- **AND** the calendar displays the resulting saved plan

#### Scenario: Cancel a proposal
- **WHEN** a planner cancels the proposed changes
- **THEN** the saved events and assignments remain unchanged

### Requirement: Schedule events from reusable templates
The Add event action SHALL offer a template picker. Choosing a template SHALL populate its name, department, contact, duration, drill type, tags, constraints, and fatigue setting. The occurrence editor SHALL emphasize preferred start and allowed scheduling window, with inherited details available as optional overrides. Planners SHALL be able to create and edit reusable definitions under Event templates and copy an existing event into a new template.

#### Scenario: Minimal event-specific entry
- **WHEN** a planner selects an existing template in Add event
- **THEN** the occurrence inherits the template defaults without requiring them to be re-entered
- **AND** the planner can set timing, optionally override details, and use the usual validation, draft saving, and resolution actions

#### Scenario: Define a reusable activity
- **WHEN** a planner creates an event template
- **THEN** the editor requests reusable defaults without calendar dates
- **AND** the saved definition becomes available in the event picker
