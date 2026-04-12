# Tier 3 Phase 1: Preflight Checks & Support Bundle Specs

**Goal:** Fill in the existing stub templates `_preflight.tpl` and `_supportbundle.tpl` with production-quality preflight checks and support bundle collectors/analyzers that satisfy bootcamp items 3.1a–3.6.

**Scope:** Pure Helm template work — no app code changes. Phase 2 (item 3.7, the `/admin` UI for support bundle generation) is a separate spec.

---

## Preflight Spec (`_preflight.tpl`)

### 3.1a — External Database Connectivity

**Conditional:** Only runs when `postgresql.enabled=false` (BYO external DB mode).

**Collector:** `run` collector using busybox to TCP-connect to `externalDatabase.host:externalDatabase.port`.

**Analyzer:** `textAnalyze` on the collector output. Regex matches `connected` for pass. Fail message names the host/port and tells the operator to verify credentials and network reachability.

### 3.1b — Cloudflare API Connectivity

**Conditional:** Only runs when `ingress.tls.cloudflare.enabled=true`.

**Collector:** `run` collector using busybox to TCP-connect to `api.cloudflare.com:443`.

**Analyzer:** `textAnalyze` on the collector output. Regex matches `connected` for pass. Fail message explains that cert-manager DNS-01 challenges require outbound HTTPS to the Cloudflare API.

### 3.1c — Cluster Resource Check

**Always runs.** Two `nodeResources` analyzers:

- **CPU:** fail when `sum(cpuAllocatable) < 2`, warn when `< 4`, pass otherwise. DroneRx is lightweight — 2 cores is the real minimum, 4 is recommended.
- **Memory:** fail when `sum(memoryAllocatable) < 4Gi`, warn when `< 8Gi`, pass otherwise.

### 3.1d — Kubernetes Version Check

**Always runs.** `clusterVersion` analyzer:

- fail when `< 1.28.0` — names minimum version, links to K8s upgrade docs.
- warn when `< 1.30.0` — recommends upgrade.
- pass otherwise.

### 3.1e — Distribution Check

**Always runs.** `distribution` analyzer:

- fail on `docker-desktop` — names it as unsupported, explains it lacks production storage/networking, lists supported distributions (EKS, GKE, AKS, RKE2, k3s, OpenShift).
- fail on `microk8s` — same pattern.
- pass otherwise.

---

## Support Bundle Spec (`_supportbundle.tpl`)

### 3.2 — Log Collectors (Per-Component)

Four log collectors, each with `maxLines` and `maxAge` limits:

| Collector | Selector | maxLines | maxAge | Conditional |
|-----------|----------|----------|--------|-------------|
| API logs | `app.kubernetes.io/name=<name>, component=api` | 10000 | 72h | Always |
| Frontend logs | `app.kubernetes.io/name=<name>, component=frontend` | 5000 | 48h | Always |
| CNPG Postgres logs | `cnpg.io/cluster=<fullname>-db` | 5000 | 48h | Only when `postgresql.enabled=true` |
| NATS logs | `app.kubernetes.io/name=<release>-nats` | 5000 | 48h | Always |

The CNPG selector uses `cnpg.io/cluster` label which CNPG applies to all managed pods. This is more reliable than matching on the operator's labels.

### 3.3 — Health Endpoint HTTP Collector + textAnalyze

**Collector:** `http` GET to `http://<fullname>-api.<namespace>.svc.cluster.local:8080/healthz` using in-cluster service DNS.

**Analyzer:** `textAnalyze` on the response body. Regex `"status":\s*"ok"` matches the health handler's JSON response `{"status":"ok","db":"ok","nats":"ok"}`. Fail message tells operator to check pod logs. Pass confirms app is healthy.

### 3.4 — Status Analyzers for All Workload Types

Two `deploymentStatus` analyzers (the app's workload types are all Deployments):

- **API deployment** (`<fullname>-api`): fail when `< 1` available replica. Message: "API is unavailable, users cannot access the application."
- **Frontend deployment** (`<fullname>-frontend`): fail when `< 1` available replica. Message: "Frontend is unavailable, users cannot load the web interface."

Note: CNPG manages its own pods via the Cluster CR, not a Deployment/StatefulSet we control. NATS is a StatefulSet from the subchart. We do NOT add status analyzers for subchart-managed workloads — the CNPG and NATS health is captured indirectly through the health endpoint check (3.3) and the known-failure textAnalyze (3.5).

### 3.5 — textAnalyze for Known App Failure Pattern

**Analyzer:** `textAnalyze` searching API log collector output for DB and NATS connection failures.

**Regex:** `database:.*|nats:.*|NATS not connected`

This matches the actual log output from `cmd/api/main.go`:
- `log.Fatalf("database: %v", err)` — DB connection failure at startup
- `log.Fatalf("nats: %v", err)` — NATS connection failure at startup
- `"NATS not connected"` — health check runtime detection

**Fail message:** Explains the failure mode (dependency connectivity), provides remediation steps (check DB/NATS pod status, verify connection strings, check network policies), and suggests relevant kubectl commands.

**Pass message:** "No database or NATS connection failures detected in recent logs."

### 3.6 — Storage Class and Node Readiness

Two analyzers:

- **`storageClass`**: fail when no default storage class. Message explains PVCs will fail and provides the kubectl patch command to set a default.
- **`nodeResources`**: fail when `nodeCondition(Ready) == False`. Message tells operator to check node status with `kubectl get nodes` and `kubectl describe node`.

---

## Values Changes

Add `cloudflare.enabled` under `ingress.tls` in both `values.yaml` and `values.schema.json`:

```yaml
ingress:
  tls:
    cloudflare:
      enabled: false
```

Schema addition:
```json
"cloudflare": {
  "type": "object",
  "properties": {
    "enabled": {
      "type": "boolean",
      "description": "Enable Cloudflare DNS-01 solver for Let's Encrypt certs. Requires a Cloudflare API token Secret."
    }
  }
}
```

---

## Files Changed

| File | Action |
|------|--------|
| `chart/templates/_preflight.tpl` | Rewrite (replace empty stub) |
| `chart/templates/_supportbundle.tpl` | Rewrite (replace empty stub) |
| `chart/values.yaml` | Add `ingress.tls.cloudflare.enabled` |
| `chart/values.schema.json` | Add cloudflare schema under ingress.tls |

---

## Testing Strategy

All verification is via `helm template` — no cluster needed.

- `helm template` with defaults: preflight renders 3 always-on checks (resources, version, distribution), no conditional collectors. Support bundle renders API/frontend/NATS log collectors, health HTTP collector, all analyzers.
- `helm template --set postgresql.enabled=false --set externalDatabase.host=... --set externalDatabase.password=...`: preflight gains DB connectivity collector + analyzer. Support bundle loses CNPG log collector.
- `helm template --set ingress.tls.cloudflare.enabled=true`: preflight gains Cloudflare connectivity collector + analyzer.
- `helm lint` passes on all combinations.
