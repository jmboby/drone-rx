# Tier 4 — Embedded Cluster v3 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Enable Embedded Cluster v3 distribution of DroneRx on bare VMs, supporting online/air-gap install, upgrade, and license entitlement gating.

**Architecture:** Add EC v3 config to the Replicated manifests, create a static v1beta3 preflight for EC's runner, wire `LicenseFieldValue` through the HelmChart CR into chart values and Go env vars for belt-and-suspenders entitlement gating, and update CI for EC releases. Phase 2 adds cert-manager + Cloudflare as an EC Helm extension.

**Tech Stack:** Embedded Cluster v3 (3.0.0-alpha-31+k8s-1.34), Replicated SDK, KOTS HelmChart CR, Troubleshoot v1beta3, GitHub Actions

**Spec:** `docs/superpowers/specs/2026-04-14-tier4-embedded-cluster-v3.md`

---

## Phase 1: EC Core (Rubric 4.1, 4.2, 4.3, 4.6, 4.7)

### Task 1: Create Embedded Cluster config

**Files:**
- Create: `replicated/embedded-cluster.yaml`

- [ ] **Step 1: Create the EmbeddedClusterConfig manifest**

```yaml
apiVersion: embeddedcluster.replicated.com/v1beta1
kind: Config
metadata:
  name: drone-rx
spec:
  version: "3.0.0-alpha-31+k8s-1.34"
```

This is the minimal EC config. EC's built-in OpenEBS handles storage. No Helm extensions yet (added in Phase 2).

- [ ] **Step 2: Verify it's picked up by the .replicated manifest glob**

Run: `cat .replicated`
Expected: `manifests: - ./replicated/*.yaml` — the new file is included automatically.

- [ ] **Step 3: Commit**

```bash
git add replicated/embedded-cluster.yaml
git commit -m "feat: add Embedded Cluster v3 config for VM installs"
```

---

### Task 2: Create static v1beta3 preflight for EC

**Files:**
- Create: `replicated/preflight-v1beta3.yaml`

EC v3's built-in preflight runner is broken when Helm templating is used in the spec (it reads YAML before the chart is templated). This preflight uses only static checks — no `{{ }}` expressions.

The existing v1beta2 preflight in `chart/templates/_preflight.tpl` is untouched and continues to work for Helm CLI installs.

- [ ] **Step 1: Create the v1beta3 preflight with static analyzers**

```yaml
apiVersion: troubleshoot.sh/v1beta3
kind: Preflight
metadata:
  name: drone-rx
spec:
  analyzers:
    # 3.1c: Cluster CPU capacity
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
    # 3.1c: Cluster memory capacity
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
    # 3.1d: Kubernetes version
    - clusterVersion:
        checkName: Kubernetes Version
        outcomes:
          - fail:
              when: "< 1.28.0"
              message: |
                Kubernetes version is not supported.
                DroneRx requires Kubernetes 1.28 or higher.
                Upgrade your cluster before installing.
          - warn:
              when: "< 1.30.0"
              message: Kubernetes version is supported but upgrading to 1.30+ is recommended.
          - pass:
              message: Kubernetes version is supported.
    # 3.1e: Distribution check
    - distribution:
        checkName: Kubernetes Distribution
        outcomes:
          - fail:
              when: "== docker-desktop"
              message: |
                docker-desktop is not a supported Kubernetes distribution.
                Supported distributions: EKS, GKE, AKS, RKE2, k3s, OpenShift, Embedded Cluster.
          - fail:
              when: "== microk8s"
              message: |
                microk8s is not a supported Kubernetes distribution.
                Supported distributions: EKS, GKE, AKS, RKE2, k3s, OpenShift, Embedded Cluster.
          - pass:
              message: Kubernetes distribution is supported.
    # 3.6: Default storage class
    - storageClass:
        checkName: Default Storage Class
        outcomes:
          - fail:
              when: "== false"
              message: |
                No default storage class is configured on this cluster.
                DroneRx requires a default storage class for PostgreSQL persistent volume claims.
          - pass:
              message: A default storage class is available.
```

- [ ] **Step 2: Commit**

```bash
git add replicated/preflight-v1beta3.yaml
git commit -m "feat: add static v1beta3 preflight for EC installer"
```

---

### Task 3: Update HelmChart CR with LicenseFieldValue and EC defaults

**Files:**
- Modify: `replicated/dronerx-chart.yaml`

- [ ] **Step 1: Add LicenseFieldValue and EC-sensible defaults**

The current file:

```yaml
apiVersion: kots.io/v1beta2
kind: HelmChart
metadata:
  name: drone-rx
spec:
  chart:
    name: drone-rx
    chartVersion: $VERSION
  values:
    api:
      image:
        repository: images.littleroom.co.nz/proxy/drone-rx/ghcr.io/jmboby/dronerx-api
        tag: $VERSION
    frontend:
      image:
        repository: images.littleroom.co.nz/proxy/drone-rx/ghcr.io/jmboby/dronerx-frontend
        tag: $VERSION
    cloudnative-pg:
      imagePullSecrets:
        - name: enterprise-pull-secret
    nats:
      global:
        image:
          pullSecretNames:
            - enterprise-pull-secret
```

Replace with:

```yaml
apiVersion: kots.io/v1beta2
kind: HelmChart
metadata:
  name: drone-rx
spec:
  chart:
    name: drone-rx
    chartVersion: $VERSION
  values:
    api:
      image:
        repository: images.littleroom.co.nz/proxy/drone-rx/ghcr.io/jmboby/dronerx-api
        tag: $VERSION
      liveTrackingEnabled: repl{{ LicenseFieldValue "live_tracking_enabled" }}
    frontend:
      image:
        repository: images.littleroom.co.nz/proxy/drone-rx/ghcr.io/jmboby/dronerx-frontend
        tag: $VERSION
    ingress:
      enabled: false
    postgresql:
      enabled: true
    cloudnative-pg:
      imagePullSecrets:
        - name: enterprise-pull-secret
    nats:
      global:
        image:
          pullSecretNames:
            - enterprise-pull-secret
```

Changes:
- `api.liveTrackingEnabled`: Seeds entitlement from license at install time via `LicenseFieldValue`
- `ingress.enabled: false`: EC admin console proxies traffic, no ingress needed
- `postgresql.enabled: true`: Embedded DB by default on VMs

- [ ] **Step 2: Commit**

```bash
git add replicated/dronerx-chart.yaml
git commit -m "feat: add LicenseFieldValue and EC defaults to HelmChart CR"
```

---

### Task 4: Wire liveTrackingEnabled through chart values and configmap

**Files:**
- Modify: `chart/values.yaml` (add `api.liveTrackingEnabled`)
- Modify: `chart/values.schema.json` (add schema for new field)
- Modify: `chart/templates/configmap-api.yaml` (add env var)

- [ ] **Step 1: Add liveTrackingEnabled to values.yaml**

In `chart/values.yaml`, add after the `webhookURL` field (line 16):

```yaml
  liveTrackingEnabled: "true"
```

The value is a string because `LicenseFieldValue` returns strings. Default `"true"` means tracking is enabled on Helm CLI installs where the HelmChart CR doesn't apply.

- [ ] **Step 2: Add schema entry in values.schema.json**

In `chart/values.schema.json`, inside `properties.api.properties`, add after the `webhookURL` entry:

```json
        "liveTrackingEnabled": {
          "type": "string",
          "description": "Enable live drone tracking. Set via LicenseFieldValue on KOTS/EC installs.",
          "default": "true"
        }
```

- [ ] **Step 3: Add env var to configmap**

In `chart/templates/configmap-api.yaml`, add after the `REPLICATED_SDK_URL` line:

```yaml
  LIVE_TRACKING_ENABLED: {{ .Values.api.liveTrackingEnabled | quote }}
```

- [ ] **Step 4: Validate with helm lint**

Run: `helm dependency build chart/ && helm lint chart/`
Expected: No errors.

- [ ] **Step 5: Verify template output**

Run: `helm template drone-rx chart/ | grep -A1 LIVE_TRACKING`
Expected: `LIVE_TRACKING_ENABLED: "true"`

- [ ] **Step 6: Commit**

```bash
git add chart/values.yaml chart/values.schema.json chart/templates/configmap-api.yaml
git commit -m "feat: wire liveTrackingEnabled through chart values and configmap"
```

---

### Task 5: Add liveTrackingEnabled env var to Go config and SDK client

**Files:**
- Modify: `internal/config/config.go` (add field + loader)
- Modify: `internal/sdk/client.go` (add env fallback to IsFeatureEnabled)
- Modify: `internal/sdk/client_test.go` (add test for env fallback)
- Modify: `cmd/api/main.go` (pass config value to SDK client)

- [ ] **Step 1: Write failing test for env var fallback**

In `internal/sdk/client_test.go`, add this test:

```go
func TestIsFeatureEnabled_EnvFallback(t *testing.T) {
	// SDK server that returns 500 (unreachable)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := NewClient(srv.URL)

	// No env override — SDK fails, returns false (fail closed)
	got := client.IsFeatureEnabled("live_tracking_enabled")
	if got != false {
		t.Errorf("expected false when SDK down and no override, got %v", got)
	}

	// With env override — SDK fails but env says enabled
	client.SetFeatureOverride("live_tracking_enabled", "true")
	got = client.IsFeatureEnabled("live_tracking_enabled")
	if got != true {
		t.Errorf("expected true with env override, got %v", got)
	}

	// With env override disabled
	client.SetFeatureOverride("live_tracking_enabled", "false")
	got = client.IsFeatureEnabled("live_tracking_enabled")
	if got != false {
		t.Errorf("expected false with env override false, got %v", got)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/jwilson/git/dronerx && go test ./internal/sdk/ -run TestIsFeatureEnabled_EnvFallback -v`
Expected: FAIL — `SetFeatureOverride` method does not exist.

- [ ] **Step 3: Add override support to SDK client**

In `internal/sdk/client.go`, update the `Client` struct and add `SetFeatureOverride`:

Add an `overrides` field to the Client struct:

```go
type Client struct {
	baseURL    string
	httpClient *http.Client
	overrides  map[string]string
}
```

Update `NewClient`:

```go
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		overrides: make(map[string]string),
	}
}
```

Add `SetFeatureOverride` method after `NewClient`:

```go
// SetFeatureOverride sets a static override for a feature field.
// Used as a fallback when the SDK API is unreachable (e.g., during startup).
// The value from the license at install time is passed via env var.
func (c *Client) SetFeatureOverride(fieldName, value string) {
	c.overrides[fieldName] = value
}
```

Update `IsFeatureEnabled` to check overrides on SDK failure. Replace lines 168-187:

```go
func (c *Client) IsFeatureEnabled(fieldName string) bool {
	field, err := c.GetLicenseField(fieldName)
	if err != nil {
		slog.Error("sdk: feature check failed", "field", fieldName, "error", err)
		// Fall back to static override from install-time license value
		if override, ok := c.overrides[fieldName]; ok {
			result := override == "true" || override == "1"
			slog.Debug("sdk feature check fallback", "field", fieldName, "enabled", result)
			return result
		}
		return false
	}
	var result bool
	switch v := field.Value.(type) {
	case bool:
		result = v
	case string:
		result = v == "true" || v == "1"
	case float64:
		result = v == 1
	default:
		result = false
	}
	slog.Debug("sdk feature check", "field", fieldName, "enabled", result)
	return result
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /Users/jwilson/git/dronerx && go test ./internal/sdk/ -run TestIsFeatureEnabled_EnvFallback -v`
Expected: PASS

- [ ] **Step 5: Run all SDK tests to check for regressions**

Run: `cd /Users/jwilson/git/dronerx && go test ./internal/sdk/ -v`
Expected: All tests PASS.

- [ ] **Step 6: Add LiveTrackingEnabled to Go config**

In `internal/config/config.go`, add `LiveTrackingEnabled` to the Config struct:

```go
type Config struct {
	Port                string
	DatabaseURL         string
	NATSUrl             string
	TickerInterval      int
	WebhookURL          string
	SDKUrl              string
	Namespace           string
	LiveTrackingEnabled string
}
```

And in `Load()`, add:

```go
		LiveTrackingEnabled: getEnv("LIVE_TRACKING_ENABLED", "true"),
```

- [ ] **Step 7: Pass config value to SDK client in main.go**

In `cmd/api/main.go`, after the line that creates the SDK client (`sdkClient := sdk.NewClient(cfg.SDKUrl)`), add:

```go
	if cfg.LiveTrackingEnabled != "" {
		sdkClient.SetFeatureOverride("live_tracking_enabled", cfg.LiveTrackingEnabled)
	}
```

- [ ] **Step 8: Run full build to verify compilation**

Run: `cd /Users/jwilson/git/dronerx && go build ./...`
Expected: No errors.

- [ ] **Step 9: Commit**

```bash
git add internal/config/config.go internal/sdk/client.go internal/sdk/client_test.go cmd/api/main.go
git commit -m "feat: add env var fallback for live_tracking_enabled entitlement"
```

---

### Task 6: Verify and update kots-app.yaml

**Files:**
- Modify: `replicated/kots-app.yaml`

- [ ] **Step 1: Review current kots-app.yaml**

Current content:

```yaml
apiVersion: kots.io/v1beta1
kind: Application
metadata:
  name: drone-rx
spec:
  title: DroneRx
  icon: https://raw.githubusercontent.com/jmboby/drone-rx/main/frontend/src/lib/assets/favicon.svg
  statusInformers:
    - deployment/dronerx-api
    - deployment/dronerx-frontend
  ports:
    - serviceName: dronerx-frontend
      servicePort: 3000
      localPort: 3000
      applicationUrl: "http://dronerx-frontend"
```

The icon and title are already correct (rubric 4.6). The `ports` section configures the EC admin console proxy — `serviceName` should match the actual frontend service name from the chart.

- [ ] **Step 2: Verify the frontend service name matches**

Run: `helm template drone-rx chart/ | grep 'name:.*frontend' | head -5`
Expected: Service name like `drone-rx-frontend`. If the Helm release name is `drone-rx`, the fullname helper produces `drone-rx-frontend`.

- [ ] **Step 3: Update ports to use the correct service name**

The `statusInformers` use `dronerx-api` and `dronerx-frontend` (without the hyphen in "drone"). But the Helm fullname helper uses the release name. On EC/KOTS installs, the release name comes from the HelmChart CR `metadata.name` which is `drone-rx`. So the actual resource names will be `drone-rx-api` and `drone-rx-frontend`.

Update `replicated/kots-app.yaml`:

```yaml
apiVersion: kots.io/v1beta1
kind: Application
metadata:
  name: drone-rx
spec:
  title: DroneRx
  icon: https://raw.githubusercontent.com/jmboby/drone-rx/main/frontend/src/lib/assets/favicon.svg
  statusInformers:
    - deployment/drone-rx-api
    - deployment/drone-rx-frontend
  ports:
    - serviceName: drone-rx-frontend
      servicePort: 3000
      localPort: 3000
      applicationUrl: "http://drone-rx-frontend"
```

**Note:** The Helm fullname template may vary. Verify with `helm template drone-rx chart/ | grep 'name: drone-rx-'` before committing. If the names don't use hyphens, adjust accordingly.

- [ ] **Step 4: Commit**

```bash
git add replicated/kots-app.yaml
git commit -m "fix: correct service and deployment names in kots-app.yaml for EC"
```

---

### Task 7: Update CI workflow for EC releases

**Files:**
- Modify: `.github/workflows/release.yaml`

The release pipeline already creates Replicated releases. EC builds are triggered automatically when the channel has Embedded Cluster enabled. The main change is ensuring the Unstable channel has EC enabled (a one-time Vendor Portal setting, not a CI change).

- [ ] **Step 1: Verify EC is enabled on the Unstable channel**

Run locally:
```bash
replicated channel ls --output json | jq '.[] | select(.name == "Unstable") | {name, embeddedClusterEnabled: .isEmbeddedClusterEnabled}'
```

If `embeddedClusterEnabled` is `false` or not present, enable it:
```bash
replicated channel enable-embedded-cluster Unstable
```

If that command doesn't exist, enable it via the Vendor Portal UI: Channels → Unstable → Settings → Enable Embedded Cluster.

- [ ] **Step 2: Add EC customer creation to test job**

In `.github/workflows/release.yaml`, in the `test-on-cmx` job, update the "Create test customer" step to also create an EC-enabled customer for future EC testing.

Add after the existing cleanup steps (at the end of `test-on-cmx`), add a new step:

```yaml
      - name: Verify EC build available
        continue-on-error: true
        run: |
          echo "Checking if EC build is available for this release..."
          replicated release ls --output json | jq '.[0] | {sequence, version, embeddedClusterArtifacts}'
```

This is a validation step — it confirms the release includes EC artifacts. Full EC VM testing can be added later once the basic flow is validated manually.

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/release.yaml
git commit -m "ci: add EC build verification step to release workflow"
```

---

### Task 8: Local validation and helm lint

- [ ] **Step 1: Run helm lint**

Run: `helm dependency build chart/ && helm lint chart/`
Expected: No errors.

- [ ] **Step 2: Run helm template and spot-check**

Run:
```bash
helm template drone-rx chart/ --set api.liveTrackingEnabled="false" | grep -A2 LIVE_TRACKING
```
Expected: `LIVE_TRACKING_ENABLED: "false"`

- [ ] **Step 3: Run Go tests**

Run: `cd /Users/jwilson/git/dronerx && go test ./... -v`
Expected: All tests pass.

- [ ] **Step 4: Verify all replicated manifests are valid YAML**

Run:
```bash
for f in replicated/*.yaml; do echo "--- $f ---"; python3 -c "import yaml; yaml.safe_load(open('$f'))" && echo "OK" || echo "FAIL"; done
```
Expected: All files OK. Note: `dronerx-chart.yaml` will fail YAML parse because of `repl{{ }}` template — this is expected and correct for KOTS templating.

---

## Phase 2: cert-manager + Cloudflare TLS

### Task 9: Add cert-manager as EC Helm extension

**Files:**
- Modify: `replicated/embedded-cluster.yaml`

- [ ] **Step 1: Add cert-manager extension**

Replace `replicated/embedded-cluster.yaml` with:

```yaml
apiVersion: embeddedcluster.replicated.com/v1beta1
kind: Config
metadata:
  name: drone-rx
spec:
  version: "3.0.0-alpha-31+k8s-1.34"
  extensions:
    helmCharts:
      - chart:
          name: cert-manager
          chartVersion: "v1.17.1"
        releaseName: cert-manager
        namespace: cert-manager
        weight: 10
        values: |
          crds:
            enabled: true
          image:
            pullPolicy: IfNotPresent
```

Weight 10 ensures cert-manager installs before the app chart. `crds.enabled: true` installs CRDs. `pullPolicy: IfNotPresent` is required for air-gap compatibility.

- [ ] **Step 2: Commit**

```bash
git add replicated/embedded-cluster.yaml
git commit -m "feat: add cert-manager as EC Helm extension for TLS support"
```

---

### Task 10: Wire Cloudflare TLS toggle through HelmChart CR for EC

**Files:**
- Modify: `replicated/dronerx-chart.yaml`

- [ ] **Step 1: Add Cloudflare/TLS overrides to HelmChart CR values**

In `replicated/dronerx-chart.yaml`, update the `ingress` section in `spec.values`:

```yaml
    ingress:
      enabled: false
      tls:
        mode: "self-signed"
```

The default for EC is `ingress.enabled: false` (admin console proxies) with `self-signed` TLS. When a user wants real TLS on EC, they can override these values via the KOTS Config screen (Tier 5) or by editing the HelmChart CR to set `ingress.enabled: true` and `tls.mode: "auto"` with Cloudflare enabled.

For now, this sets the safe EC defaults. The full Cloudflare wiring (with KOTS Config conditional) is deferred to Tier 5 when `repl{{ ConfigOptionEquals }}` template functions become available.

- [ ] **Step 2: Commit**

```bash
git add replicated/dronerx-chart.yaml
git commit -m "feat: set self-signed TLS defaults for EC installs"
```

---

## Manual Testing Checklist (post-implementation)

These are performed manually on CMX or real VMs after all code changes are merged.

- [ ] **4.1 — Fresh EC install:** Create CMX VM with `replicated cluster create --distribution embedded-cluster-v3 --version 3.0.0-alpha-31 --ttl 2h`. Download and run installer. Verify all pods Running and app opens in browser.
- [ ] **4.2 — Upgrade:** Install release N, create drone delivery data, promote release N+1, trigger upgrade in admin console. Verify data persists.
- [ ] **4.3 — Air-gap:** Build air-gap bundle from Vendor Portal, transfer to isolated VM, install offline. Verify all pods Running.
- [ ] **4.6 — Icon and name:** During EC install, verify "DroneRx" title and SVG icon appear in the admin console.
- [ ] **4.7 — Entitlement gating:** Install with license where `live_tracking_enabled: false`. Navigate to order tracking — verify 403/premium message. Update license field to `true` in Vendor Portal — verify tracking works without redeploy.
- [ ] **Phase 2 — TLS:** EC install with cert-manager extension. Enable Cloudflare + ingress. Verify real TLS cert issued.
