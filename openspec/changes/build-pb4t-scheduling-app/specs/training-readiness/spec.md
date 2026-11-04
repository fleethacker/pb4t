## ADDED Requirements

### Requirement: Configured training obligations
The system SHALL allow a planner to configure requirements by drill type, responsible department, explicit coverage window, required occurrence count, and source reference. Repeat intervals SHALL support generating subsequent coverage windows. Seeded policy values SHALL be visibly illustrative and SHALL NOT be labeled independently verified SFTRM mandates.

#### Scenario: Configure a periodic drill requirement
- **WHEN** a planner records two occurrences of a drill type within a specified coverage window
- **THEN** the system persists the count, window, department, and supplied source reference for planning

### Requirement: Separate completed and projected coverage
The system SHALL compute completed coverage from completion records and projected coverage from matching validated scheduled events. It SHALL display completed count, planned count, and remaining gap separately. Scheduling an event SHALL NOT create a completion record, and a completed linked event SHALL NOT also count as a future planned occurrence.

#### Scenario: A future drill covers a projected gap
- **WHEN** one completed occurrence and one validated future occurrence satisfy a two-occurrence requirement
- **THEN** the dashboard shows one completed occurrence, one planned occurrence, and no projected gap
- **AND** it does not show two completed occurrences

### Requirement: Account for obligations during weekly generation
For requirements due within the selected week, the system SHALL include enough missing desired drill occurrences to close projected gaps, accounting for matching existing events and completion history. Generated occurrences SHALL use the matching drill template's duration and constraints and participate in the same placement and staffing checks as manually entered events.

#### Scenario: Generate missing training
- **WHEN** a requirement due this week needs one more occurrence and a matching drill template is configured
- **THEN** generation includes one additional desired event with that template's constraints
- **AND** a complete feasible result places and staffs that event within the coverage window

### Requirement: Show unfulfilled obligations
The system SHALL expose each remaining training gap and any known constraints preventing its scheduling. A proposal with an unmet obligation due in the selected week SHALL NOT be described as a complete feasible weekly plan.

#### Scenario: Required training cannot be placed
- **WHEN** an outstanding drill has no found placement that satisfies its personnel and operating constraints
- **THEN** the application identifies the drill requirement, deadline, remaining count, and known scheduling blockers
