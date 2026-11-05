# PB4T UDS package

Packages the existing application image and Helm chart for a cluster already running **UDS Core with its Operator, tenant gateway, Keycloak, and Authservice**. UDS Core creates the Keycloak client and protects the whole calendar and API through Authservice. The app continues to use one shared planner and persistent SQLite database; all admitted users have the same editing permissions.

Default address: **https://pb4t.uds.dev**. Use an existing Keycloak realm account. This package does not install Keycloak, create users, or include passwords/client secrets. Native startup, Compose, and ordinary Helm deployments retain their existing behavior; authentication is provided by the UDS gateway and service mesh.

## Build and deploy

From the repository root, with Docker running and the [UDS CLI](https://github.com/defenseunicorns/uds-cli) installed:

```sh
# Choose the target cluster architecture: arm64 or amd64.
docker build --platform linux/arm64 -t docker.io/library/pb4t-planner:0.1.0 .
mkdir -p .cache/uds-packages
uds zarf package create deploy/uds --confirm --architecture arm64 \
  --output .cache/uds-packages

# Verify the selected cluster before deploying.
kubectl config current-context
uds zarf package deploy .cache/uds-packages/zarf-package-pb4t-arm64-0.1.0.tar.zst --confirm
kubectl --namespace pb4t get package pb4t
kubectl --namespace pb4t rollout status deployment/pb4t
```

For AMD64, replace both architecture arguments and the archive filename with `amd64`. Zarf includes the locally built image in the archive, so no external image registry push is needed. The local image tag produces a Zarf lint warning about digest pinning. When releasing a new image/package, update the image tag in `values.yaml`, the image reference in `zarf.yaml`, and the package version together.

Wait for the Package to become Ready. Open the HTTPS address and sign in with Keycloak. A matching DNS entry and trusted gateway certificate must already be configured by the platform. Visiting the protected host without a session should require login, including direct requests to `/api/context`; an authenticated session should load the calendar and allow editing. Persistent storage requires the cluster's default StorageClass, or an existing claim configured through the chart values.

## Hostname and access

The `uds.domain` value must match UDS Core's tenant domain. The `uds.host` value is its application subdomain. Create a local override file, for example `.cache/pb4t-uds-values.yaml`:

```yaml
uds:
  host: training
  domain: platform.example.test
  allowedGroups:
    - /PB4T/Planners
```

Use the exact Keycloak group paths configured by your platform. An omitted or empty `allowedGroups` admits any authenticated realm user. A nonempty list requires membership in at least one listed group. No department-level permissions are implemented.

Preview and deploy the same overrides:

```sh
uds zarf package inspect manifests \
  .cache/uds-packages/zarf-package-pb4t-arm64-0.1.0.tar.zst \
  --values .cache/pb4t-uds-values.yaml
uds zarf package deploy .cache/uds-packages/zarf-package-pb4t-arm64-0.1.0.tar.zst \
  --values .cache/pb4t-uds-values.yaml --confirm
```

Override files use the Helm values structure directly. Keep `uds.enabled: true` and `service.type: ClusterIP` for this deployment. Keep the image reference matching the image included in the archive. The Keycloak redirect URI is constructed as `https://<host>.<domain>/login`, which Authservice handles at the mesh boundary. No application callback route or client secret is needed.

## Optional live AI

Core planning works without an OpenAI key. To enable live AI, first provision a Secret named `pb4t-openai` in namespace `pb4t`, with a key named `OPENAI_API_KEY`. Add this reference to your override file:

```yaml
openai:
  existingSecret: pb4t-openai
```

Use `openai.key` for a different Secret key name. The UDS Package permits outbound TLS to `api.openai.com:443` only when this Secret reference is configured. Credentials stay out of chart values and the package. Other outbound connections remain subject to UDS Core's network policy. See the [Helm guide](../helm/pb4t/README.md) for storage, model, timezone, resource, and log settings.

## Validate without a cluster

```sh
helm lint deploy/helm/pb4t --strict
uds zarf dev lint deploy/uds
uds zarf dev inspect manifests deploy/uds
uds zarf dev inspect manifests deploy/uds --values .cache/pb4t-uds-values.yaml
```

Validated with UDS CLI 0.39.0 (embedded Zarf 0.87.0), Helm 4.3.0, and the UDS Core 1.14.3 Package CR schema. ARM64 and AMD64 archives can be built from the same source. Manifest/archive checks do not verify a live Keycloak login; complete the browser/API check above on your UDS cluster.

The integration follows UDS Core's [Authservice guide](https://docs.defenseunicorns.com/core/v1-15/how-to-guides/identity--authorization/protect-apps-with-authservice/) and [Package CR reference](https://docs.defenseunicorns.com/core/v1-15/reference/operator--crds/packages-v1alpha1-cr/). The exposure, Service, Pod, and Authservice selectors match, and the exposure records both Service port 80 and container port 8080. There are no authentication bypass paths.
