## ADDED Requirements

### Requirement: Retain the original training document
The application SHALL accept PDF uploads up to 20 MB and persist the original bytes, filename, import timestamp, size, and SHA-256 hash in SQLite before analysis. The document SHALL remain available through the training document library and an original PDF link after refresh or restart. Uploading a document SHALL NOT create planning records automatically.

#### Scenario: Import and reopen an original
- **WHEN** a planner imports a training PDF and later restarts the application
- **THEN** the document remains listed and its original PDF can be opened unchanged
- **AND** no requirements or templates have been created without user action

### Requirement: Extract advisory suggestions from untrusted source content
With server-side AI configured, the application SHALL analyze the PDF and persist editable training requirement and event-template suggestions with a 1-based source page, supporting excerpt, source basis, and missing-detail notes. Document contents SHALL be treated as reference data rather than executable instructions. The importer SHALL NOT follow document links, contact recipients, execute commands, or create accounts. Recommended training SHALL NOT be represented as a mandated periodicity. Unknown departments, durations, counts, and recurrence SHALL remain unspecified for user review.

#### Scenario: Import recommended courses
- **WHEN** a PDF recommends courses with some durations but no recurring training policy
- **THEN** the application presents advisory suggestions and identifies missing planning details
- **AND** it does not invent recurring requirements, dates, personnel demands, or asset constraints

#### Scenario: Distinguish duration from a video timestamp
- **WHEN** the source says to fast forward to a timestamp without stating total duration
- **THEN** the suggestion leaves duration unspecified for user review
- **AND** an extracted nonzero duration requires supporting source evidence

### Requirement: Explicit review and creation
The planner SHALL be able to edit and persist a suggestion draft, select a department and applicable constraints, and explicitly create an event template or training requirement. Creation SHALL validate the resulting record and atomically record acceptance so retrying the same creation does not duplicate it. Missing mandatory record fields SHALL produce actionable feedback. Created records SHALL remain editable through existing planning views.

#### Scenario: Create a reviewed requirement
- **WHEN** the planner supplies a department, count, and coverage window and selects Create training requirement
- **THEN** the requirement is persisted with those user-defined values
- **AND** the suggestion is marked created without creating a calendar event

#### Scenario: Match a prerequisite to its reusable activity
- **WHEN** extraction suggests both a requirement and template for the same prerequisite activity
- **THEN** both suggestions use the same drill type for projected coverage

### Requirement: Preserve source references on accepted records
Created requirements and event templates SHALL retain their source document ID, source page, and excerpt. Events requested from an imported template SHALL inherit its source reference and description. The planning views SHALL provide links back to the retained original PDF.

#### Scenario: Trace a scheduled activity to its source
- **WHEN** a planner creates an event from an imported template
- **THEN** its details include a link to the original source PDF page

#### Scenario: Match physical pages with an added advisory cover
- **WHEN** an extracted excerpt matches text on a physical PDF page that differs from the model's page number
- **THEN** the source link uses the locally matched physical page, including advisory/cover pages in the count
- **AND** excerpts that cannot be matched locally are visibly marked for source review

#### Scenario: Check an elided quote across two pages
- **WHEN** a citation uses ellipses to omit text and its literal segments appear in order on one or two adjacent PDF pages
- **THEN** the checker matches those segments and links to the page where the quotation begins
- **AND** paraphrased or reordered text is not treated as a literal source match

#### Scenario: Correct a citation without rerunning AI
- **WHEN** a planner rechecks a saved document or edits a draft's Source citation and saves it
- **THEN** the application checks the citation locally without an API key or another AI request
- **AND** obsolete citation notices are cleared after a successful match
- **AND** remaining mismatches appear as advisory review notices with editable wording/page fields
- **AND** the retained original PDF remains fixed as the source document

### Requirement: Retention without AI and recoverable analysis
Without an API key, the application SHALL save and serve uploaded PDFs and explain how to enable suggestions. Failed or interrupted analyses SHALL retain the original PDF and offer retry. Document polling SHALL pause after a request error rather than loop indefinitely.

#### Scenario: Restart during analysis
- **WHEN** the app restarts while an import is being analyzed
- **THEN** the original is retained and analysis is marked interrupted with a retry action

### Requirement: Review source-supported constraint suggestions
The importer SHALL also suggest source-cited constraint drafts using the existing qualified-personnel, named-person, asset, allowed/excluded-window, incompatibility, minimum-rest, and unenforced-note types. Resource mappings SHALL reference existing planning IDs or remain unspecified for user review. Missing operational policies SHALL NOT be invented. Unsupported conditions, including account access and prerequisite knowledge/completion, SHALL be explicitly labeled unenforced notes. New constraint drafts SHALL be unconfirmed and default to attached-event scope, without modifying scheduling rules or attachments.

#### Scenario: Delete an unwanted suggestion
- **WHEN** the planner selects Delete suggestion in an uncreated item's Review & edit editor
- **THEN** the suggestion disappears and its pending review count decreases, including after refresh
- **AND** Undo restores the saved draft
- **AND** the original PDF and existing planning records remain unchanged
- **AND** already-created suggestions must be managed through Planning inputs rather than deleted through draft review

#### Scenario: Review and create an operational rule
- **WHEN** a source states that a drill cannot overlap flight quarters
- **THEN** the importer offers an unconfirmed incompatibility draft using the existing flight tag with a page and excerpt
- **AND** the planner can edit, save, explicitly confirm, and create it as a reusable constraint
- **AND** it affects attached events or templates only after explicit attachment, unless the planner chooses entire-plan scope

#### Scenario: Retain a condition that cannot be evaluated
- **WHEN** a source states that an account is needed to access training
- **THEN** the importer may suggest a source-cited unenforced constraint note
- **AND** saving the note never makes it an enforced rule or a training-completion requirement

#### Scenario: Enforce a confirmed imported rule deterministically
- **WHEN** the planner creates and attaches a confirmed imported incompatibility constraint to a template or event
- **THEN** the existing deterministic evaluator reports matching overlaps without further AI interpretation
- **AND** the constraint remains traceable to its original PDF page

#### Scenario: Keep earlier imports out of the source evidence
- **WHEN** earlier templates or documents contain durations, counts, or source excerpts that the new PDF does not state
- **THEN** extraction does not reuse those values as evidence for the new document
- **AND** the mapping context contains only reference IDs, names, and event tags

#### Scenario: Distinguish doctrine procedures from scheduling policy
- **WHEN** doctrine contains a currency warning, historical deadlines, conditional training triggers, sample agreement language, and fire-response timers
- **THEN** drafts preserve that context for review rather than inventing current calendar dates or fixed-day recurrence
- **AND** a maximum firefighter work/rotation interval is not represented as a minimum-rest scheduling rule
