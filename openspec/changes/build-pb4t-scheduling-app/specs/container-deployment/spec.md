## ADDED Requirements

### Requirement: Self-contained container image
The project SHALL provide a multi-stage Docker build containing the Go server and built React assets. The runtime SHALL run as a non-root user, serve the UI and API on one port, and store SQLite at `/data/scheduler.db`. Image builds SHALL exclude local credentials, databases, caches, and generated host assets. The image SHALL provide certificate roots for optional HTTPS API requests.

#### Scenario: Run without host build tools
- **WHEN** a user builds and starts the image with a writable data volume
- **THEN** the UI and API are available without Go or Node.js on the host
- **AND** an empty database initializes the synthetic planning dataset
- **AND** core planning works without an OpenAI key

### Requirement: Persistent local Compose operation
The repository SHALL provide a Compose configuration that builds the image, publishes its port on host localhost, and mounts a named data volume. It SHALL allow runtime timezone, model, optional API key, and host port configuration. Logs SHALL use standard output/error. Container replacement SHALL preserve saved planning and document data through the data volume.

#### Scenario: Replace the container
- **WHEN** a user recreates the Compose service using the same volume
- **THEN** the saved SQLite records remain available without reseeding over them

### Requirement: Configurable listening address
The backend SHALL use `LISTEN_HOST` to configure its listening address. Native startup SHALL default to `127.0.0.1`; container configurations SHALL use `0.0.0.0` so published ports and cluster Services reach it.

#### Scenario: Retain native startup behavior
- **WHEN** no listening host is configured
- **THEN** the server listens on localhost at its configured port

### Requirement: Single-instance Helm deployment
The project SHALL provide a Helm chart with a one-replica Deployment, `Recreate` updates, a Service, health probes, and writable data/temporary volumes. Persistence SHALL be enabled by default using a `ReadWriteOnce` claim. It SHALL support an existing claim, configurable storage class/capacity, and explicit ephemeral storage. Chart-created claims SHALL be retained on uninstall by default.

#### Scenario: Upgrade a persistent deployment
- **WHEN** Helm updates the deployment image
- **THEN** the old instance is stopped before its replacement starts
- **AND** the replacement mounts the same SQLite data volume

#### Scenario: Configure optional live AI
- **WHEN** a user supplies an existing Secret name and key through chart values
- **THEN** the container obtains its OpenAI key through a Secret reference
- **AND** secret values are not embedded in chart values or the image
- **AND** omitting the Secret leaves core planning available without live AI
