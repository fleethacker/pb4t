## ADDED Requirements

### Requirement: Persistent planning records
The system SHALL persist departments, people, qualifications, availability windows, duty commitments, assets, events, personnel assignments, constraints, drill templates, training requirements, completion records, and resolution proposals in a durable SQLite database file managed by the Go backend. It SHALL NOT require a separate database service. Event templates SHALL specify a name, drill type, department, point of contact, default duration, tags, constraints, and fatigue setting. Templates SHALL NOT store occurrence dates, requested starts, scheduling windows, fixed placements, or completion status. The frontend SHALL provide basic forms or lists for maintaining these planning inputs.

#### Scenario: Reload saved planning data
- **WHEN** a planner saves roster and event changes and the application restarts
- **THEN** the calendar and planning-data views reload those changes from persistence

### Requirement: Personnel qualifications and availability
The system SHALL associate people with departments, qualifications, availability windows, and timed duty commitments. Qualification demands SHALL reference qualification identifiers and required counts. Named-participant requirements SHALL reference specific people. A point of contact alone SHALL NOT reserve that person's time.

#### Scenario: Identify eligible personnel
- **WHEN** an event requires two Fire Party Leaders
- **THEN** the engine receives the people holding that qualification together with their availability and duty commitments
- **AND** an unqualified contact person is not counted toward the requirement

### Requirement: Asset capacity and availability
The system SHALL persist each asset's capacity and availability and each event's required quantity of that asset.

#### Scenario: Configure a shared training space
- **WHEN** a planner creates a training space with capacity one and attaches a quantity-one demand to two events
- **THEN** both events reference the same persisted asset and its shared capacity

### Requirement: Reusable typed constraints
The system SHALL support reusable constraints for qualified personnel, named participants, asset demand, allowed and excluded windows, event incompatibility, and configured minimum rest. Events SHALL support any number of attachments. Shared operating and fatigue policies SHALL apply at plan scope. Hard constraints and soft preferences SHALL be represented distinctly.

#### Scenario: Attach an existing incompatibility rule
- **WHEN** a planner selects a saved rule prohibiting overlap with flight-quarters events
- **THEN** the event stores a reference to that rule and the checker enforces its structured parameters

### Requirement: Confirm free-form rule interpretation
The system SHALL allow the planner to enter a free-form constraint and receive a proposed interpretation using supported rule types. The planner MUST confirm the interpretation before it becomes enforceable. Unsupported or unconfirmed text SHALL be labeled unenforced and SHALL prevent the affected plan from being described as fully validated.

#### Scenario: Convert a natural-language restriction
- **WHEN** a planner enters “This drill cannot overlap flight quarters”
- **THEN** the application displays the proposed incompatibility rule and matched event tag
- **AND** the rule becomes enforceable only after the planner confirms it

#### Scenario: Retain an unsupported note honestly
- **WHEN** the system cannot map a submitted constraint to a supported type
- **THEN** it identifies the text as unenforced and asks for a supported replacement or removal before full validation

### Requirement: Synthetic demonstration data
The application SHALL provide a repeatable synthetic dataset containing multiple departments, overlapping qualifications, duty commitments, shared assets, fixed operating events, movable drills, configured rest policies, and illustrative training obligations. Illustrative policy values SHALL be labeled as demo values.

#### Scenario: Start the demo
- **WHEN** a planner loads the synthetic dataset
- **THEN** the calendar includes an intentional personnel conflict, an asset conflict, and a resolvable training-coverage gap

### Requirement: Independent template occurrences
The system SHALL copy template defaults into each new occurrence and retain its source template ID. Subsequent template edits SHALL affect future occurrences without changing existing calendar events. Existing stored templates SHALL remain readable without resetting the database.

#### Scenario: Reuse and edit a template
- **WHEN** a planner schedules two occurrences from a template and later edits its default duration or contact
- **THEN** the two occurrences retain their copied values and independent dates
- **AND** subsequent occurrences receive the updated defaults
