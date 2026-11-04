## ADDED Requirements

### Requirement: One-command local startup
With Go and Node.js/npm installed, the application SHALL start from the repository root using `npm start`. The command SHALL prepare project dependencies when needed, build the frontend and backend, and serve the UI and API at one localhost origin. It SHALL print the ready URL and SHALL NOT require Docker, a separate database server, a C compiler, manual SQL, or separate frontend/backend terminals. The default URL SHALL be `http://localhost:8080`.

#### Scenario: Start a fresh checkout
- **WHEN** a user with Go, Node.js/npm, and network access runs `npm start` from a fresh checkout
- **THEN** dependencies and builds are prepared and the application starts at the printed localhost URL
- **AND** the user can open the seeded weekly calendar without additional database setup

### Requirement: Automatic durable database initialization
The application SHALL create the SQLite directory and schema automatically, using `./data/scheduler.db` by default. It SHALL seed an empty database with synthetic planning data on first launch. Subsequent launches SHALL preserve saved changes and SHALL NOT duplicate or reload demo records over existing data.

#### Scenario: Restart after making changes
- **WHEN** a user changes an event, stops the application, and runs `npm start` again
- **THEN** the saved event change remains visible and demo records are not duplicated

### Requirement: Optional live AI configuration
The application SHALL start and provide calendar management, typed constraints, deterministic conflict evaluation, training coverage, and schedule generation without an OpenAI API key. It SHALL visibly mark live departmental negotiation and free-form constraint assistance as unavailable until configured. A documented optional `.env` SHALL supply the server-side key, and the application SHALL NOT expose that key to the browser or fabricate agent activity when it is absent.

#### Scenario: Use core planning without credentials
- **WHEN** a user starts the application without an OpenAI key
- **THEN** the user can view and edit events, detect conflicts, generate and review deterministic schedule proposals, and apply valid changes
- **AND** the UI clearly identifies live AI features as unconfigured

#### Scenario: Enable live agent features
- **WHEN** a user supplies a valid server-side OpenAI key through the documented configuration and starts the application
- **THEN** live departmental negotiation and free-form constraint assistance become available

### Requirement: Explicit demo reset and concise instructions
The application SHALL provide `npm run reset-demo` to explicitly restore synthetic data and SHALL document that it replaces local planning data. The root README SHALL document prerequisites, startup, the ready URL, the database location, optional AI configuration, and demo reset. Local credentials and database files SHALL be excluded from version control.

#### Scenario: Restore the demonstration dataset
- **WHEN** a user intentionally runs the documented demo reset command
- **THEN** synthetic planning data is restored
- **AND** ordinary startup remains non-destructive
