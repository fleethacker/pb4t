# PB4T Helm chart

Deploy the same Docker image as one application instance with persistent SQLite storage. A `Recreate` upgrade stops the old instance before starting its replacement. The chart includes a Service, startup/readiness/liveness probes, a non-root container, and writable `/data` and `/tmp` volumes. No external database is needed.

## Build and install

Use a registry accessible to your cluster, replacing the example registry below:

```sh
docker buildx build --platform linux/amd64,linux/arm64 \
  -t registry.example.com/team/pb4t-planner:0.1.0 --push .
helm upgrade --install pb4t ./deploy/helm/pb4t \
  --namespace pb4t --create-namespace \
  --set image.repository=registry.example.com/team/pb4t-planner \
  --set-string image.tag=0.1.0
kubectl --namespace pb4t rollout status deployment/pb4t
kubectl --namespace pb4t port-forward service/pb4t 8080:80
```

Open **http://localhost:8080**. For a local cluster, load your locally built image into its runtime instead of pushing it, and use that image's repository/tag. The default pull policy is `IfNotPresent`. Set `imagePullSecrets` when your registry requires authentication.

## Storage

Persistence is enabled by default. The chart requests a 1 GiB `ReadWriteOnce` claim using the cluster's default StorageClass. SQLite includes planning data, proposals, original PDFs, and document suggestions. Normal upgrades/restarts retain the data; the application seeds only an empty database.

The chart-created claim is named `<release>-data` and is retained on uninstall by default. To reuse a retained claim or one managed outside Helm, set `persistence.existingClaim`. Its storage must support ownership through group `65532`, which is supplied as the Pod's `fsGroup`.

| Value | Default | Purpose |
| --- | --- | --- |
| `persistence.enabled` | `true` | Use persistent storage; `false` uses a Pod-lifetime temporary volume |
| `persistence.existingClaim` | empty | Mount an existing claim instead of creating one |
| `persistence.size` | `1Gi` | Requested capacity for a new claim |
| `persistence.storageClass` | `null` | Cluster default; `""` explicitly requests no class |
| `persistence.retain` | `true` | Keep the chart-created claim on uninstall |

The Deployment always has one replica. Scale-out requires changing the application's SQLite persistence model.

## Optional OpenAI configuration

Core planning works without credentials. For live AI, reference an existing Secret in the release namespace containing the `OPENAI_API_KEY` key:

```sh
helm upgrade --install pb4t ./deploy/helm/pb4t \
  --namespace pb4t \
  --set image.repository=registry.example.com/team/pb4t-planner \
  --set-string image.tag=0.1.0 \
  --set openai.existingSecret=pb4t-openai
```

Use `openai.key` if the Secret's key has a different name. API-key values are not stored in chart values or baked into the image. Restart the Deployment after rotating an environment-backed Secret.

Other settings include `config.timezone` (`America/New_York`), `config.openaiModel` (`gpt-5-mini`), `service.type` (`ClusterIP`), `service.port` (`80`), and `resources`. Logs go to standard output/error and can be viewed with `kubectl --namespace pb4t logs deployment/pb4t`.

## UDS Core and Keycloak

The [UDS package](../../uds/README.md) bundles this chart and its image for an existing UDS Core installation. It enables `uds.enabled`, which emits a `uds.dev/v1alpha1` Package to expose the Service through the tenant gateway and protect all UI/API routes with Keycloak and Authservice. Ordinary chart installations leave this disabled.

| Value | Default | Purpose |
| --- | --- | --- |
| `uds.enabled` | `false` | Create the UDS Package; requires UDS Core |
| `uds.host` | `pb4t` | Tenant gateway subdomain |
| `uds.domain` | `uds.dev` | UDS Core tenant domain for the HTTPS login callback |
| `uds.allowedGroups` | `[]` | Allow any authenticated user, or require any listed Keycloak group |

UDS deployments require `service.type: ClusterIP`. Supplying `openai.existingSecret` also permits TLS egress to `api.openai.com:443`. Authservice manages login and its `/login` callback; the application has no separate identity store or per-user permissions.

## Validate without deploying

```sh
helm lint ./deploy/helm/pb4t --strict
helm template pb4t ./deploy/helm/pb4t --namespace pb4t
```
