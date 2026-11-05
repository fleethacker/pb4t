## ADDED Requirements

### Requirement: Reusable UDS application package
The repository SHALL provide a Zarf package definition usable through the UDS CLI that includes the application image and existing Helm chart. It SHALL enable a UDS Core Package resource and retain the chart's single-instance persistent SQLite deployment. The package SHALL require an existing UDS Core Operator, tenant gateway, Keycloak, and Authservice installation rather than install another identity platform.

#### Scenario: Assemble the application archive
- **WHEN** a user builds the application image for the target architecture and creates the package
- **THEN** the archive contains the image and local application chart
- **AND** inspecting its manifests renders the Deployment, Service, persistent claim, and UDS Package

### Requirement: Platform-managed Keycloak access
The UDS Package SHALL configure an OpenID Connect client and enable Authservice for the application's Service and Pod selectors. It SHALL expose the complete Service through the tenant gateway, protecting both calendar and API without bypass routes. Its redirect URI SHALL use HTTPS with a non-root `/login` path handled by Authservice. The Service port and container target port SHALL both be specified.

#### Scenario: Access the shared planner
- **WHEN** an unauthenticated visitor opens the UDS application host or its API
- **THEN** the platform requires Keycloak authentication before application access
- **AND** an admitted authenticated user can access the shared planning workspace
- **AND** all admitted users have the same editing permissions

### Requirement: Configurable UDS access and hostname
The package SHALL support deployment-time overrides for the tenant subdomain, tenant domain, and optional allowed Keycloak groups using the chart's values structure. An empty group list SHALL omit the group restriction and allow authenticated realm users. A nonempty list SHALL require membership in any listed group. UDS mode SHALL require a ClusterIP Service. Ordinary Helm deployments SHALL leave the UDS Package disabled by default.

#### Scenario: Restrict a platform deployment
- **WHEN** a user supplies a custom hostname and allowed groups
- **THEN** the exposure and HTTPS redirect use that hostname
- **AND** Authservice's client configuration requires membership in one of those groups

#### Scenario: Preserve local operation
- **WHEN** the app runs natively, through Compose, or through the chart's default values
- **THEN** it does not require UDS Core or Keycloak

### Requirement: External credentials and bounded egress
The package SHALL not include Keycloak users, passwords, or client secrets. UDS Core SHALL manage client credentials. Optional OpenAI credentials SHALL use an existing Kubernetes Secret reference, with TLS egress to `api.openai.com:443` enabled only when that reference is configured. The deployment guide SHALL state platform prerequisites, configuration, persistent storage behavior, and live login verification steps.

#### Scenario: Enable optional AI
- **WHEN** a deployment references an existing OpenAI Secret
- **THEN** the Pod references its configured key and the UDS Package permits the corresponding TLS egress
- **AND** the archive and chart values contain no API-key value
