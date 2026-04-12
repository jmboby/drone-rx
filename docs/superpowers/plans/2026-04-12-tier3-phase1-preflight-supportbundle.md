# Tier 3 Phase 1: Preflight & Support Bundle Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fill the existing empty `_preflight.tpl` and `_supportbundle.tpl` stubs with production-quality preflight checks (5 required) and support bundle collectors/analyzers satisfying bootcamp items 3.1a–3.6.

**Architecture:** Pure Helm template work — no Go or frontend code changes. The two stub templates get replaced with full specs. Two new values (`ingress.tls.cloudflare.enabled`) are added for conditional preflight checks. All specs use Helm template helpers for dynamic names and conditionals.

**Tech Stack:** Helm templates, troubleshoot.sh/v1beta2 API, busybox for run collectors

---

## File Map

| File | Action | Responsibility |
|------|--------|---------------|
| `chart/templates/_preflight.tpl` | Rewrite | All 5 preflight checks (3.1a–3.1e) |
| `chart/templates/_supportbundle.tpl` | Rewrite | All collectors and analyzers (3.2–3.6) |
| `chart/values.yaml` | Modify | Add `ingress.tls.cloudflare.enabled` |
| `chart/values.schema.json` | Modify | Add cloudflare schema under ingress.tls |

---

### Task 1: Add cloudflare.enabled to values and schema

**Files:**
- Modify: `chart/values.yaml:38-44`
- Modify: `chart/values.schema.json:162-178`

- [ ] **Step 1: Add cloudflare.enabled to values.yaml**

In `chart/values.yaml`, replace the `ingress.tls` block (lines 41-44):

```yaml
  tls:
    mode: "auto"
    secretName: "letsencrypt-cert"
    email: "jamesw@replicated.com"
```

with:

```yaml
  tls:
    mode: "auto"
    secretName: "letsencrypt-cert"
    email: "jamesw@replicated.com"
    cloudflare:
      enabled: false
```

- [ ] **Step 2: Add cloudflare schema to values.schema.json**

In `chart/values.schema.json`, inside the `tls.properties` object (after the `email` property block that ends around line 177), add a comma after the `email` closing brace and add:

```json
            "cloudflare": {
              "type": "object",
              "description": "Cloudflare DNS-01 solver for Let's Encrypt certs. Requires a Cloudflare API token Secret.",
              "properties": {
                "enabled": {
                  "type": "boolean",
                  "description": "Enable Cloudflare DNS-01 challenge for cert-manager"
                }
              }
            }
```

- [ ] **Step 3: Verify helm lint passes**

Run: `helm lint ./chart`
Expected: `1 chart(s) linted, 0 chart(s) failed`

- [ ] **Step 4: Verify the new value renders**

Run: `helm template test-release ./chart --show-only templates/ingress.yaml 2>&1 | head -5`
Expected: No errors (ingress template doesn't reference cloudflare yet, just ensuring values don't break anything)

- [ ] **Step 5: Commit**

```bash
git add chart/values.yaml chart/values.schema.json
git commit -m "feat: add ingress.tls.cloudflare.enabled for DNS-01 preflight check"
```

---

### Task 2: Implement preflight spec (`_preflight.tpl`)

**Files:**
- Rewrite: `chart/templates/_preflight.tpl`

- [ ] **Step 1: Replace the entire contents of `_preflight.tpl`**

Replace the full file with:

```gotemplate
{{- define "dronerx.preflight" -}}
apiVersion: troubleshoot.sh/v1beta2
kind: Preflight
metadata:
  name: {{ include "dronerx.fullname" . }}-preflight
spec:
  collectors:
    {{- /* 3.1a: External DB connectivity collector (conditional) */}}
    {{- if not .Values.postgresql.enabled }}
    - run:
        collectorName: dronerx-db-check
        image: images.littleroom.co.nz/anonymous/index.docker.io/library/busybox:1.36
        command: ["sh", "-c"]
        args:
          - |
            nc -zv {{ .Values.externalDatabase.host }} {{ .Values.externalDatabase.port }} 2>&1 && echo "connected" || echo "connection_failed"
    {{- end }}
    {{- /* 3.1b: Cloudflare API connectivity collector (conditional) */}}
    {{- if .Values.ingress.tls.cloudflare.enabled }}
    - run:
        collectorName: dronerx-cloudflare-check
        image: images.littleroom.co.nz/anonymous/index.docker.io/library/busybox:1.36
        command: ["sh", "-c"]
        args:
          - |
            nc -zv api.cloudflare.com 443 2>&1 && echo "connected" || echo "connection_failed"
    {{- end }}
  analyzers:
    {{- /* 3.1a: External DB connectivity analyzer (conditional) */}}
    {{- if not .Values.postgresql.enabled }}
    - textAnalyze:
        checkName: External Database Connectivity
        fileName: preflight/dronerx-db-check.txt
        regex: "connected"
        outcomes:
          - fail:
              when: "false"
              message: |
                Cannot connect to external database at {{ .Values.externalDatabase.host }}:{{ .Values.externalDatabase.port }}.
                Verify the host, port, and credentials are correct and the database is reachable from this cluster.
                Test manually: nc -zv {{ .Values.externalDatabase.host }} {{ .Values.externalDatabase.port }}
          - pass:
              when: "true"
              message: External database at {{ .Values.externalDatabase.host }}:{{ .Values.externalDatabase.port }} is reachable.
    {{- end }}
    {{- /* 3.1b: Cloudflare API connectivity analyzer (conditional) */}}
    {{- if .Values.ingress.tls.cloudflare.enabled }}
    - textAnalyze:
        checkName: Cloudflare API Connectivity
        fileName: preflight/dronerx-cloudflare-check.txt
        regex: "connected"
        outcomes:
          - fail:
              when: "false"
              message: |
                Cannot reach Cloudflare API at api.cloudflare.com:443.
                cert-manager DNS-01 challenges require outbound HTTPS access to the Cloudflare API.
                Ensure firewall and proxy rules allow outbound TCP to api.cloudflare.com on port 443.
          - pass:
              when: "true"
              message: Cloudflare API is reachable.
    {{- end }}
    {{- /* 3.1c: Cluster resource checks (always) */}}
    - nodeResources:
        checkName: Cluster CPU Capacity
        outcomes:
          - fail:
              when: "sum(cpuAllocatable) < 2"
              message: |
                Insufficient CPU: cluster has less than 2 allocatable cores.
                DroneRx requires at least 2 CPU cores across all nodes.
          - warn:
              when: "sum(cpuAllocatable) < 4"
              message: |
                Cluster has fewer than 4 CPU cores. Performance may be degraded under load.
                Recommended: 4+ cores for production workloads.
          - pass:
              message: Cluster has sufficient CPU capacity.
    - nodeResources:
        checkName: Cluster Memory Capacity
        outcomes:
          - fail:
              when: "sum(memoryAllocatable) < 4Gi"
              message: |
                Insufficient memory: cluster has less than 4 GiB allocatable.
                DroneRx requires at least 4 GiB of memory across all nodes.
          - warn:
              when: "sum(memoryAllocatable) < 8Gi"
              message: |
                Cluster has less than 8 GiB memory. Production workloads may be constrained.
          - pass:
              message: Cluster has sufficient memory.
    {{- /* 3.1d: Kubernetes version check (always) */}}
    - clusterVersion:
        checkName: Kubernetes Version
        outcomes:
          - fail:
              when: "< 1.28.0"
              message: |
                Kubernetes version is not supported.
                DroneRx requires Kubernetes 1.28 or higher.
                Upgrade your cluster before installing.
                See: https://kubernetes.io/docs/tasks/administer-cluster/cluster-upgrade/
          - warn:
              when: "< 1.30.0"
              message: Kubernetes version is supported but upgrading to 1.30+ is recommended.
          - pass:
              message: Kubernetes version is supported.
    {{- /* 3.1e: Distribution check (always) */}}
    - distribution:
        checkName: Kubernetes Distribution
        outcomes:
          - fail:
              when: "== docker-desktop"
              message: |
                docker-desktop is not a supported Kubernetes distribution.
                It is a local development environment and lacks the storage and networking
                capabilities required for production use.
                Supported distributions: EKS, GKE, AKS, RKE2, k3s, OpenShift.
          - fail:
              when: "== microk8s"
              message: |
                microk8s is not a supported Kubernetes distribution.
                It lacks enterprise storage and networking support required by DroneRx.
                Supported distributions: EKS, GKE, AKS, RKE2, k3s, OpenShift.
          - pass:
              message: Kubernetes distribution is supported.
{{- end }}
```

- [ ] **Step 2: Verify default values render (embedded path — no conditional collectors)**

Run: `helm template test-release ./chart 2>&1 | grep -A2 "kind: Preflight"`
Expected: Preflight resource renders with no errors.

Run: `helm template test-release ./chart 2>&1 | grep -c "collectorName"`
Expected: `0` (no run collectors when postgresql.enabled=true and cloudflare.enabled=false)

Run: `helm template test-release ./chart 2>&1 | grep "checkName"`
Expected: 5 lines — Cluster CPU Capacity, Cluster Memory Capacity, Kubernetes Version, Kubernetes Distribution (appears once but has two fail outcomes)

- [ ] **Step 3: Verify external DB path adds the DB collector**

Run: `helm template test-release ./chart --set postgresql.enabled=false --set externalDatabase.host=ep-test.neon.tech --set externalDatabase.password=secret 2>&1 | grep "collectorName"`
Expected: `dronerx-db-check` appears.

Run: `helm template test-release ./chart --set postgresql.enabled=false --set externalDatabase.host=ep-test.neon.tech --set externalDatabase.password=secret 2>&1 | grep "checkName: External Database"`
Expected: `External Database Connectivity` appears.

- [ ] **Step 4: Verify cloudflare enabled adds the cloudflare collector**

Run: `helm template test-release ./chart --set ingress.tls.cloudflare.enabled=true 2>&1 | grep "collectorName"`
Expected: `dronerx-cloudflare-check` appears.

Run: `helm template test-release ./chart --set ingress.tls.cloudflare.enabled=true 2>&1 | grep "checkName: Cloudflare"`
Expected: `Cloudflare API Connectivity` appears.

- [ ] **Step 5: Verify both conditionals together**

Run: `helm template test-release ./chart --set postgresql.enabled=false --set externalDatabase.host=ep-test.neon.tech --set externalDatabase.password=secret --set ingress.tls.cloudflare.enabled=true 2>&1 | grep "collectorName"`
Expected: Both `dronerx-db-check` and `dronerx-cloudflare-check` appear.

- [ ] **Step 6: Run helm lint on all paths**

Run: `helm lint ./chart && helm lint ./chart --set postgresql.enabled=false --set externalDatabase.host=ep-test.neon.tech --set externalDatabase.password=secret --set ingress.tls.cloudflare.enabled=true`
Expected: Both pass with `0 chart(s) failed`.

- [ ] **Step 7: Commit**

```bash
git add chart/templates/_preflight.tpl
git commit -m "feat: implement all 5 required preflight checks (3.1a-3.1e)

- 3.1a: External DB connectivity (conditional on postgresql.enabled=false)
- 3.1b: Cloudflare API connectivity (conditional on ingress.tls.cloudflare.enabled)
- 3.1c: Cluster CPU and memory resource checks
- 3.1d: Kubernetes version >= 1.28
- 3.1e: Fail on docker-desktop and microk8s distributions"
```

---

### Task 3: Implement support bundle spec (`_supportbundle.tpl`)

**Files:**
- Rewrite: `chart/templates/_supportbundle.tpl`

- [ ] **Step 1: Replace the entire contents of `_supportbundle.tpl`**

Replace the full file with:

```gotemplate
{{- define "dronerx.supportbundle" -}}
apiVersion: troubleshoot.sh/v1beta2
kind: SupportBundle
metadata:
  name: {{ include "dronerx.fullname" . }}-supportbundle
spec:
  collectors:
    {{- /* 3.2: Per-component log collectors */}}
    - logs:
        name: dronerx/api-logs
        selector:
          - app.kubernetes.io/name={{ include "dronerx.name" . }}
          - app.kubernetes.io/component=api
        namespace: {{ .Release.Namespace }}
        limits:
          maxLines: 10000
          maxAge: 72h
    - logs:
        name: dronerx/frontend-logs
        selector:
          - app.kubernetes.io/name={{ include "dronerx.name" . }}
          - app.kubernetes.io/component=frontend
        namespace: {{ .Release.Namespace }}
        limits:
          maxLines: 5000
          maxAge: 48h
    {{- if .Values.postgresql.enabled }}
    - logs:
        name: dronerx/postgres-logs
        selector:
          - cnpg.io/cluster={{ include "dronerx.fullname" . }}-db
        namespace: {{ .Release.Namespace }}
        limits:
          maxLines: 5000
          maxAge: 48h
    {{- end }}
    - logs:
        name: dronerx/nats-logs
        selector:
          - app.kubernetes.io/name={{ .Release.Name }}-nats
        namespace: {{ .Release.Namespace }}
        limits:
          maxLines: 5000
          maxAge: 48h
    {{- /* 3.3: Health endpoint HTTP collector */}}
    - http:
        collectorName: dronerx-health
        get:
          url: http://{{ include "dronerx.fullname" . }}-api.{{ .Release.Namespace }}.svc.cluster.local:{{ .Values.service.apiPort }}/healthz
  analyzers:
    {{- /* 3.3: Health endpoint analyzer */}}
    - textAnalyze:
        checkName: Application Health Endpoint
        fileName: http/dronerx-health.json
        regex: '"status":\s*"ok"'
        outcomes:
          - fail:
              when: "false"
              message: |
                Application health endpoint returned unhealthy status or is unreachable.
                The /healthz endpoint checks database and NATS connectivity.
                Check pod logs: kubectl logs -l app.kubernetes.io/component=api -n {{ .Release.Namespace }}
          - pass:
              when: "true"
              message: Application is healthy — database and NATS connections are working.
    {{- /* 3.4: Status analyzers for all app workloads */}}
    - deploymentStatus:
        name: {{ include "dronerx.fullname" . }}-api
        namespace: {{ .Release.Namespace }}
        outcomes:
          - fail:
              when: "< 1"
              message: |
                The DroneRx API deployment has no available replicas.
                The application is unavailable — users cannot place or track orders.
                Run: kubectl describe deployment {{ include "dronerx.fullname" . }}-api -n {{ .Release.Namespace }}
          - pass:
              message: DroneRx API is running.
    - deploymentStatus:
        name: {{ include "dronerx.fullname" . }}-frontend
        namespace: {{ .Release.Namespace }}
        outcomes:
          - fail:
              when: "< 1"
              message: |
                The DroneRx frontend deployment has no available replicas.
                Users cannot load the web interface.
                Run: kubectl describe deployment {{ include "dronerx.fullname" . }}-frontend -n {{ .Release.Namespace }}
          - pass:
              message: DroneRx frontend is running.
    {{- /* 3.5: Known failure pattern — DB and NATS connection errors */}}
    - textAnalyze:
        checkName: Database and NATS Connection Failures
        fileName: dronerx/api-logs/*/*
        regex: 'database:.*|nats:.*|NATS not connected'
        outcomes:
          - fail:
              when: "true"
              message: |
                Database or NATS connection failure detected in API logs.
                This indicates the API cannot reach its dependencies.
                Remediation:
                1. Check database pod status: kubectl get pods -l cnpg.io/cluster={{ include "dronerx.fullname" . }}-db -n {{ .Release.Namespace }}
                2. Check NATS pod status: kubectl get pods -l app.kubernetes.io/name={{ .Release.Name }}-nats -n {{ .Release.Namespace }}
                3. Verify DATABASE_URL in the API configmap: kubectl get configmap {{ include "dronerx.fullname" . }}-api-config -n {{ .Release.Namespace }} -o yaml
                4. Check for network policies blocking pod-to-pod traffic
          - pass:
              when: "false"
              message: No database or NATS connection failures detected in recent logs.
    {{- /* 3.6: Storage class and node readiness */}}
    - storageClass:
        checkName: Default Storage Class Present
        outcomes:
          - fail:
              when: "== false"
              message: |
                No default storage class is configured on this cluster.
                DroneRx requires a default storage class for PostgreSQL persistent volume claims.
                Configure a default storage class:
                  kubectl patch storageclass <name> -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}'
          - pass:
              message: A default storage class is available.
    - nodeResources:
        checkName: All Nodes Ready
        outcomes:
          - fail:
              when: "nodeCondition(Ready) == False"
              message: |
                One or more cluster nodes are not in the Ready state.
                Application pods may fail to schedule or run.
                Check node status: kubectl get nodes
                Check node events: kubectl describe node <node-name>
          - pass:
              message: All nodes are Ready.
{{- end }}
```

- [ ] **Step 2: Verify default values render (embedded path)**

Run: `helm template test-release ./chart 2>&1 | grep -c "name: dronerx/"`
Expected: `4` (api-logs, frontend-logs, postgres-logs, nats-logs)

Run: `helm template test-release ./chart 2>&1 | grep "collectorName: dronerx-health"`
Expected: 1 match.

Run: `helm template test-release ./chart 2>&1 | grep "checkName"`
Expected: 6 matches — Application Health Endpoint, two deploymentStatus (shown as `name:` not `checkName`... let me verify). Actually `deploymentStatus` doesn't use `checkName`. Let me count: Application Health Endpoint, Database and NATS Connection Failures, Default Storage Class Present, All Nodes Ready = 4 `checkName` matches.

Corrected check:
Run: `helm template test-release ./chart 2>&1 | grep -E "checkName|deploymentStatus" | grep -v "#"`
Expected: Lines for all 6 analyzers.

- [ ] **Step 3: Verify external DB path drops postgres log collector**

Run: `helm template test-release ./chart --set postgresql.enabled=false --set externalDatabase.host=ep-test.neon.tech --set externalDatabase.password=secret 2>&1 | grep -c "name: dronerx/"`
Expected: `3` (api-logs, frontend-logs, nats-logs — no postgres-logs)

Run: `helm template test-release ./chart --set postgresql.enabled=false --set externalDatabase.host=ep-test.neon.tech --set externalDatabase.password=secret 2>&1 | grep "postgres-logs"`
Expected: No output.

- [ ] **Step 4: Verify health endpoint URL is constructed correctly**

Run: `helm template test-release ./chart 2>&1 | grep "url: http://"`
Expected: `url: http://test-release-drone-rx-api.default.svc.cluster.local:8080/healthz`

- [ ] **Step 5: Verify deployment status analyzer names**

Run: `helm template test-release ./chart 2>&1 | grep -A1 "deploymentStatus"`
Expected: Two blocks with `name: test-release-drone-rx-api` and `name: test-release-drone-rx-frontend`.

- [ ] **Step 6: Run helm lint on both paths**

Run: `helm lint ./chart && helm lint ./chart --set postgresql.enabled=false --set externalDatabase.host=ep-test.neon.tech --set externalDatabase.password=secret`
Expected: Both pass.

- [ ] **Step 7: Commit**

```bash
git add chart/templates/_supportbundle.tpl
git commit -m "feat: implement support bundle collectors and analyzers (3.2-3.6)

- 3.2: Per-component log collectors (API, frontend, CNPG, NATS) with limits
- 3.3: HTTP health endpoint collector + textAnalyze
- 3.4: Deployment status analyzers for API and frontend
- 3.5: textAnalyze for DB/NATS connection failure patterns in API logs
- 3.6: Storage class and node readiness analyzers"
```

---

### Task 4: Full integration verification

- [ ] **Step 1: Render full chart with defaults and spot-check both specs**

Run: `helm template test-release ./chart 2>&1 > /tmp/dronerx-embedded.yaml && echo "Rendered $(wc -l < /tmp/dronerx-embedded.yaml) lines"`
Expected: No errors, large output.

Run: `grep -c "kind: Preflight" /tmp/dronerx-embedded.yaml`
Expected: `1`

Run: `grep -c "kind: SupportBundle" /tmp/dronerx-embedded.yaml`
Expected: `1`

- [ ] **Step 2: Render external DB + Cloudflare path**

Run: `helm template test-release ./chart --set postgresql.enabled=false --set externalDatabase.host=ep-test.neon.tech --set externalDatabase.password=secret --set ingress.tls.cloudflare.enabled=true 2>&1 > /tmp/dronerx-external.yaml && echo "Rendered $(wc -l < /tmp/dronerx-external.yaml) lines"`
Expected: No errors.

Run: `grep "collectorName: dronerx-" /tmp/dronerx-external.yaml`
Expected: 3 matches — `dronerx-db-check`, `dronerx-cloudflare-check`, `dronerx-health`

Run: `grep "postgres-logs" /tmp/dronerx-external.yaml`
Expected: No output (CNPG logs excluded on external path).

- [ ] **Step 3: Verify mutual exclusion guard still works**

Run: `helm template test-release ./chart --set externalDatabase.host=ep-test.neon.tech 2>&1 | head -3`
Expected: Error containing "mutually exclusive"

- [ ] **Step 4: Run helm lint on all value combinations**

Run: `helm lint ./chart`
Expected: Pass.

Run: `helm lint ./chart --set postgresql.enabled=false --set externalDatabase.host=ep-test.neon.tech --set externalDatabase.password=secret`
Expected: Pass.

Run: `helm lint ./chart --set ingress.tls.cloudflare.enabled=true`
Expected: Pass.

Run: `helm lint ./chart --set postgresql.enabled=false --set externalDatabase.host=ep-test.neon.tech --set externalDatabase.password=secret --set ingress.tls.cloudflare.enabled=true`
Expected: Pass.

- [ ] **Step 5: Run Go tests to verify no regressions**

Run: `go test ./...`
Expected: All tests pass.
