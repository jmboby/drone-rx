# DroneRx Friction Log

Pain points encountered while building and distributing a Helm-based app with Replicated.

---

## Process — AI Agent Workflow

### Claude was too confident writing CI code from docs alone
**Problem:** Claude read Replicated CLI docs and `--help` output, then immediately wrote commands into GitHub Actions workflows without testing them locally first. Multiple commands had hidden requirements not obvious from docs (e.g., `--email` required with `--helm-install`, `--auto -y` ignoring `.replicated` config, `--id` deprecated in favour of positional args, stdout warnings corrupting kubeconfig redirects).
**Impact:** Each failure required waiting 5-10 minutes for a GH Actions run + CMX cluster to spin up, only to discover a simple flag issue. This happened repeatedly across multiple CI iterations.
**Resolution:** Established a rule: always run `replicated` CLI commands locally using the API token before embedding them in workflows. Use existing CMX clusters for testing (`replicated cluster ls`) instead of creating new ones.
**Time wasted:** ~2-3 hours across all CI iterations that could have been caught in seconds locally.
**Lesson:** Don't trust docs or help output alone. Run the actual command first, verify the output format, then write the workflow. This applies to any CLI tool being embedded in CI.

---

## Tier 0 — Build It

### CloudNativePG CRD chicken-and-egg
**Problem:** Including CNPG operator as a subchart and creating a Cluster CR in the same Helm release fails because Helm validates all manifests before applying any — CRDs don't exist yet.
**Error:** `no matches for kind "Cluster" in version "postgresql.cnpg.io/v1"`
**Resolution:** Made the Cluster CR a `post-install` hook so it's applied after the operator subchart registers CRDs.
**Time spent:** ~30 minutes figuring out the right approach.

### CNPG operator webhook timing
**Problem:** Even with the CR as a post-install hook, the operator webhook isn't ready when Helm fires the hook. The operator pod needs time to start and register endpoints.
**Error:** `failed calling webhook "mcluster.cnpg.io": no endpoints available for service "cnpg-webhook-service"`
**Resolution:** Added a wait Job (busybox + `nc -z`) as a post-install hook at weight 1, before the Cluster CR hook at weight 10. Polls the webhook service port until ready.
**Time spent:** Multiple iterations — tried kubectl wait (needed RBAC), then simplified to nc.

### Go status enum mismatch with Postgres
**Problem:** Go constant `StatusInFlight = "in_flight"` (underscore) didn't match Postgres enum `'in-flight'` (hyphen). Orders advanced from `placed` to `preparing` but never to `in-flight`.
**Error:** Silent failure — ticker logged errors but orders stayed stuck at `preparing`.
**Resolution:** Changed Go constant to match DB: `StatusInFlight = "in-flight"`.
**Time spent:** ~10 minutes — user noticed orders weren't progressing and reported it.

### Docker amd64 builds for CMX
**Problem:** Building Docker images on Apple Silicon (arm64) and pushing to GHCR, then pulling on CMX k3s clusters (amd64) fails with platform mismatch.
**Error:** `no match for platform in manifest: not found`
**Resolution:** `docker build --platform linux/amd64` for all CI builds.
**Lesson:** Always build for the target platform, not the dev machine.

---

## Tier 1 — Automate It

### GHCR package permissions for GitHub Actions
**Problem:** `GITHUB_TOKEN` in workflows can't push to GHCR packages that were originally created by manual `docker push`.
**Error:** `denied: permission_denied: write_package`
**Resolution:** Two steps: (1) Enable "Read and write permissions" in repo Settings → Actions → Workflow permissions. (2) Link existing GHCR packages to the repo in Package Settings → Repository Access.
**Time spent:** ~20 minutes across two separate permission issues.

### prepare-cluster vs create-cluster
**Problem:** Started with `replicatedhq/replicated-actions/prepare-cluster` (all-in-one action). It doesn't properly handle image pull auth for `proxy.replicated.com` — creates auth for `registry.replicated.com` (OCI chart pull) but not for the proxy registry used by pod image pulls.
**Error:** `failed to authorize: failed to fetch anonymous token` on all proxied images.
**Resolution:** Switched to individual CLI commands: `replicated release create` → `replicated customer create` → `replicated cluster create` → `helm install`. The `helm-install` step with registry credentials handles auth correctly.
**Time spent:** ~2 hours across multiple debugging iterations.
**Lesson:** Use the `replicated` CLI, not `replicated-actions`. The CLI also supports Embedded Cluster which actions don't.

### replicated-actions vs CLI
**Problem:** After switching to CLI, used `--auto -y` flag which ignores the `.replicated` config file and defaults to looking for `./manifests` directory.
**Error:** `lstat ./manifests: no such file or directory`
**Resolution:** Remove `--auto -y` — the CLI reads `.replicated` automatically without it.
**Time spent:** ~15 minutes — was able to test locally with the CLI to reproduce.

### RBAC policy resource names are lowercase
**Problem:** The Vendor Portal documentation shows resource names in mixed case (e.g., `KOTS/app/*/read`) but the actual policy format requires lowercase.
**Error:** Various 403 errors on release creation, customer creation, cluster operations.
**Resolution:** Used `kots/` prefix (lowercase), not `KOTS/`. Also discovered that `kots/cluster/*/kubeconfig` is a separate permission from `kots/cluster/*`.
**Time spent:** ~45 minutes across multiple RBAC iterations.
**Lesson:** Start with broad permissions (`**/*` minus `team/**`), get CI working, then tighten.

### Helm-install customer requires email
**Problem:** `replicated customer create --helm-install` requires `--email` but this isn't obvious from the help text. The error only appears at runtime.
**Error:** `email is required for customers with helm install enabled`
**Resolution:** Add `--email "name@example.com"` to customer create.
**Lesson:** Test CLI commands locally before embedding in CI workflows.

### Channel slugs must be lowercase
**Problem:** Promote workflow used `Stable` (capitalized) but the API expects lowercase slugs.
**Error:** `Could not find channel with slug Stable or name undefined`
**Resolution:** Changed to `stable` in the workflow.
**Time spent:** 5 minutes.

### Helm-only release vs KOTS customer
**Problem:** `create-customer` defaults to KOTS install enabled. Our release is Helm-only, so assigning a KOTS customer to a Helm-only channel fails.
**Error:** `Cannot assign customer with KOTS install enabled to a channel with a helm-cli-only release`
**Resolution:** Add `--kots-install=false --helm-install` to customer create.
**Time spent:** 5 minutes once the error was clear.

### Version label must match chart version
**Problem:** `create-release` with `chart:` input requires the `version` label to match the version inside the packaged `.tgz`. We were using dynamic versions (`0.1.0-pr-9-sha`) but the chart was packaged as `0.1.0`.
**Error:** `Version label does not match any Helm charts in the release`
**Resolution:** Use the chart's native version from `Chart.yaml` for the release, bake dynamic tags into `values.yaml` via `sed` before packaging.
**Time spent:** ~15 minutes.

### `channel rm` needs ID not name
**Problem:** `replicated channel delete <name>` doesn't work — the CLI requires the channel ID.
**Error:** `archive app channel: Not found`
**Resolution:** Capture channel ID from the release create output and use `replicated channel rm <ID>`.
**Time spent:** 5 minutes.

### GitHub Actions default branch
**Problem:** The repo was created with `feat/phase1-build-it` as the default branch (from the initial push). The `promote.yaml` workflow with `workflow_dispatch` wasn't visible in the Actions UI because GitHub looks for workflows on the default branch.
**Error:** `workflow promote.yaml not found on the default branch`
**Resolution:** Changed default branch to `main` in repo Settings.
**Time spent:** 10 minutes.

---

## Tier 2 — Ship It with Helm

### Image proxy path format
**Problem:** Multiple iterations getting the proxy image path format right. Started with `proxy.replicated.com/proxy/app/docker.io/library/busybox`, then `proxy.replicated.com/proxy/app/library/busybox`, then `/anonymous/index.docker.io/library/busybox`.
**Error:** Various 400/404/401 errors on image pulls.
**Resolution:** The correct approach is to add all registries (including Docker Hub) as external registries in the Vendor Portal, then use `/proxy/<app-slug>/` for everything. Each registry needs credentials configured even for public images.
**Time spent:** ~2 hours across many iterations.
**Lesson:** Don't use `/anonymous/` path — add registries properly in Vendor Portal.

### imagePullSecrets needed everywhere
**Problem:** Added `imagePullSecrets` to deployments but forgot about hook Jobs (wait-for-cnpg, self-signed cert). These also need the `enterprise-pull-secret` to pull images through the proxy.
**Error:** `ErrImagePull` on hook job pods.
**Resolution:** Added `imagePullSecrets` helper include to ALL pod specs — deployments AND jobs.
**Time spent:** 15 minutes.

### NATS global.image.registry inconsistency
**Problem:** Set `global.image.registry` in NATS subchart values expecting it to apply to all images. The main `nats` container still used the default `nats:2.12.6-alpine` without the registry prefix.
**Error:** `pull access denied, repository does not exist`
**Resolution:** Use per-image `registry` overrides instead of `global.image.registry` — set `registry` on each of `nats`, `reloader`, and `natsBox` individually.
**Time spent:** 20 minutes.

### SDK metrics silently failing
**Problem:** Custom metrics weren't appearing in Vendor Portal. The `SendMetrics` function silently returned nil on all errors — no logging, no visibility.
**Resolution:** Added error logging to `SendMetrics`. Also added immediate send on startup (not just after first 5-minute interval) for faster verification.
**Lesson:** Never silently swallow errors in best-effort code. Log them.

### Stale releases on Unstable channel
**Problem:** Multiple commits pushed to main, some release workflows failed partway through. A failed workflow still created and promoted a release (with broken image paths) to Unstable. The later fix commit's release was overshadowed.
**Resolution:** Pushed an empty commit to trigger a fresh release from the correct state.
**Lesson:** Failed release workflows can leave stale releases on channels. Check what's actually on the channel, not just what CI reports.

### `--wait` deadlock with post-install hooks
**Problem:** Added `--wait --timeout 10m` to `helm install` in CI workflows. But the API pod has an init container waiting for the DB, which is created by a post-install hook. Helm `--wait` blocks until all pods are ready before running post-install hooks — creating a deadlock.
**Error:** `helm install` times out; pods stuck in `Init:0/1`, no Cluster CR created, no DB.
**Resolution:** Removed `--wait` from `helm install`. The smoke test step handles waiting for pod readiness separately.
**Time spent:** ~15 minutes debugging on CMX cluster.
**Lesson:** Never use `--wait` with post-install hooks that create resources other pods depend on.

### sed stopped matching after release-please changed tag values
**Problem:** The CI workflow used `sed -i "s|tag: \"latest\"|..."` to replace image tags. After adding release-please with `x-release-please-version` annotations, the tags in values.yaml changed from `"latest"` to `"1.0.0"`. The sed no longer matched anything.
**Error:** Images tried to pull with tag `1.0.0` which didn't exist in GHCR (only PR tags existed).
**Resolution:** Changed sed to match the annotation pattern: `sed -i "s|tag: \"[^\"]*\" # x-release-please-version|tag: \"${TAG}\" # x-release-please-version|g"`.
**Time spent:** ~10 minutes.
**Lesson:** When adding version management tools, check all sed/grep patterns that depend on the old format.

### Chart version vs release label in OCI registry
**Problem:** `helm install --version` needs the **chart version** from Chart.yaml, not the Replicated release label. We were passing the release label (e.g., `0.0.0-pr-17-xxx`) but the chart was packaged with version `1.0.0`.
**Error:** `FetchReference: not found` — chart not in registry at that version.
**Resolution:** Use `needs.build-and-push.outputs.version` (derived from tag/Chart.yaml) consistently, not the release label.
**Time spent:** ~20 minutes testing locally with CLI.

### GITHUB_TOKEN tags don't trigger other workflows
**Problem:** release-please creates a git tag using `GITHUB_TOKEN`. GitHub Actions security prevents tags created by `GITHUB_TOKEN` from triggering other workflows (to prevent infinite loops). So the Replicated Release workflow never triggered.
**Resolution:** Used `workflow_call` — release-please directly calls the Replicated Release workflow when a release is created. No PAT needed.
**Time spent:** ~15 minutes.

### SDK returns boolean license field values, not strings
**Problem:** The SDK returns `"value": true` (JSON boolean) for Boolean license fields, but our Go struct had `Value string`. Go's json decoder silently fails to decode a boolean into a string field — the value was always empty.
**Error:** `live_tracking_enabled` always returned false even when set to true in the license.
**Resolution:** Changed `LicenseField.Value` to `interface{}` with a type switch handling `bool`, `string`, and `float64`.
**Time spent:** ~20 minutes — had to query the SDK directly from inside the cluster to see the actual response format.

### SDK license info has no `isExpired` field
**Problem:** Our code checked `info.IsExpired` but the SDK `/license/info` endpoint doesn't have a top-level `isExpired` field. Expiry is in `entitlements.expires_at.value` as a date string.
**Error:** License always showed as valid even when expired.
**Resolution:** Added `Entitlements` map to `LicenseInfo` struct, with `IsExpired()` method that parses the `expires_at` entitlement date.
**Time spent:** ~10 minutes.

### SDK `licenseID` field casing mismatch
**Problem:** SDK returns `"licenseID"` (capital D) but our struct had `json:"licenseId"` (lowercase d). Go's JSON decoder is case-sensitive for struct tags.
**Resolution:** Fixed to `json:"licenseID"`.
**Time spent:** 2 minutes once spotted.
**Lesson:** Always test SDK responses by querying the actual endpoint — don't trust documentation for field names.

### CNPG Cluster CR data lost on helm upgrade
**Problem:** The CNPG Cluster CR had `helm.sh/hook: post-install,post-upgrade`. On upgrades, Helm re-ran the hook which recreated the cluster with `bootstrap.initdb`, wiping all data.
**Resolution:** Changed to `post-install` only. The Cluster CR persists after first install and CNPG manages it normally on upgrades.
**Time spent:** ~5 minutes.

### SDK `nameOverride` doesn't prepend release name
**Problem:** The SDK chart's `nameOverride` sets the deployment name **directly** — it does NOT prepend the Helm release name like most charts. So `nameOverride: "sdk"` gives a deployment called `sdk`, not `<release>-sdk`. You need `nameOverride: "drone-rx-sdk"` to get `drone-rx-sdk`.
**Resolution:** Set `nameOverride` to the full desired deployment name including the app prefix.
**Time spent:** ~15 minutes testing with `helm template`.
**Lesson:** Always verify subchart naming with `helm template` — don't assume standard Helm name prefix behavior.

---

## Tier 3 — Support It

### Dual database toggles caused silent inconsistency
**Problem:** The chart had two separate toggles — `cloudnativepg.enabled` (operator subchart) and `postgresql.enabled` (Cluster CR) — that always had to move together. The `databaseURL` helper checked one, all other templates checked the other. No valid combination existed where they differed.
**Error:** No error — just a confusing developer experience and risk of misconfiguration.
**Resolution:** Collapsed to single `postgresql.enabled` toggle. Changed Chart.yaml condition to `postgresql.enabled`, removed `cloudnativepg` from values/schema, simplified wait-for-cnpg-job condition.
**Time spent:** ~30 minutes to identify, plan, and implement.
**Lesson:** When a subchart exists solely to support one feature, tie its condition to the feature toggle, not a separate key.

### Hardcoded sslmode=disable breaks cloud Postgres
**Problem:** The `databaseURL` helper hardcoded `sslmode=disable` for both embedded and external DB paths. Neon (and most cloud Postgres providers) require `sslmode=require`.
**Resolution:** Added `externalDatabase.sslmode` field (default `require`), used it in the external DB path. Kept `sslmode=disable` for embedded CNPG (cluster-local traffic).
**Lesson:** Never hardcode connection parameters that differ between development and production.

### Mutual exclusion guard too strict for pre-configured values
**Problem:** Added a `fail` guard to prevent `postgresql.enabled=true` and `externalDatabase.host` being set simultaneously. But this prevented pre-configuring the Neon endpoint in values.yaml alongside the embedded default.
**Resolution:** Relaxed the guard to only enforce the essential check — `postgresql.enabled=false` requires a host. Having external DB values alongside embedded postgres is fine (pre-configuration, not conflict).
**Lesson:** Helm guards should prevent broken installs, not restrict how values are organized.

### Troubleshoot CRDs don't exist on Helm CLI installs
**Problem:** Added `preflight.yaml` and `supportbundle.yaml` as standalone `kind: Preflight` / `kind: SupportBundle` resources. CI Helm install failed because Troubleshoot CRDs don't exist on regular clusters — only in KOTS/EC environments.
**Error:** `resource mapping not found for name: "***-preflight" namespace: "" from "": no matches for kind "Preflight" in version "troubleshoot.sh/v1beta2"`
**Resolution:** Dual-mode rendering: (1) Standalone CRD resources gated on `.Capabilities.APIVersions.Has "troubleshoot.sh/v1beta2"` for KOTS/EC. (2) Kubernetes Secrets with `troubleshoot.sh/kind` labels for Helm CLI installs, discovered via `kubectl preflight --load-cluster-specs`.
**Time spent:** ~20 minutes debugging CI failure + researching Replicated docs.
**Lesson:** Troubleshoot specs need different delivery mechanisms for KOTS vs Helm CLI paths.

### Preflight run collector output paths differ from support bundle
**Problem:** `textAnalyze` fileName was set to `preflight/dronerx-db-check.txt` (based on reference docs), then changed to `dronerx-db-check/stdout.txt` (support bundle format). Neither matched the actual preflight output.
**Error:** `No matching files` — collectors ran successfully but analyzers couldn't find the output.
**Resolution:** Preflight `run` collectors write to `<collectorName>.log`, not `<collectorName>/stdout.txt` or `preflight/<collectorName>.txt`. Had to extract the actual preflight bundle and inspect file paths.
**Time spent:** ~30 minutes across two incorrect attempts.
**Lesson:** Always verify collector output paths by extracting a real bundle. Preflight and support bundle use different output path conventions for `run` collectors.

### Empty collectors section triggers default collection
**Problem:** When no conditional `run` collectors were active (embedded DB, no Cloudflare), the preflight spec rendered `collectors: []` (empty). The troubleshoot tool interprets this as "run default collectors" — gathering cluster info, pod logs, etc. — making preflights very slow.
**Resolution:** Only render the `collectors:` section when at least one conditional collector is active. When no collectors are needed, omit the section entirely.
**Time spent:** ~10 minutes.

### Faking --api-versions doesn't fully work for testing
**Problem:** Used `helm template --api-versions troubleshoot.sh/v1beta2` to render preflight specs for testing. The template rendered correctly but `kubectl preflight` didn't execute `run` collectors properly when fed the output.
**Resolution:** Use the Secret-based approach for local testing. Apply the chart (or just the preflight secret) to the cluster, then run `kubectl preflight` pointing at the rendered spec via `yq` extraction or `--show-only`.
**Time spent:** ~15 minutes.

---

## General Observations

### What worked well
- The `.replicated` config file for release packaging — simple, declarative
- release-please for semver management — clean flow with Release PRs
- CloudNativePG as a subchart — once the webhook timing was solved, very clean
- CMX k3s clusters for CI testing — fast provisioning, realistic environment
- Replicated SDK for license gating — runtime queries with no-redeploy updates
- Testing CLI commands locally before embedding in CI workflows — saved hours of debugging
- release-please for semver — clean flow, auto-CHANGELOG, version annotations in Chart.yaml and values.yaml
- workflow_call chaining — elegant solution for GITHUB_TOKEN tag limitation
- Named templates for troubleshoot specs — DRY pattern, same spec serves both KOTS CRD and Helm Secret delivery
- Extracting preflight bundles to verify file paths — saved debugging time vs guessing
- Single `postgresql.enabled` toggle — cleaner than two-toggle approach, less user confusion

### What could be improved
- **Documentation inconsistency** — RBAC resource names shown in mixed case in docs but require lowercase in config
- **prepare-cluster action** — doesn't handle image proxy auth, should probably be deprecated in favour of CLI
- **Proxy registry auth** — not obvious that `registry.replicated.com` (OCI chart pull) and `proxy.replicated.com` (image proxy) are different auth domains
- **Error messages** — many Replicated API errors are generic (403, 400) without indicating which specific permission is missing
- **CLI `--auto` flag** — confusing that it ignores `.replicated` config rather than enhancing it
- **Troubleshoot file path docs** — reference docs show `preflight/<collectorName>.txt` for run collectors but actual output is `<collectorName>.log`. Different format for preflights vs support bundles is not documented
- **No CRDs on Helm installs** — not obvious that Troubleshoot CRDs only exist in KOTS/EC environments. The Secret-based discovery pattern for Helm installs isn't prominently documented
- **Secret data key naming** — the key must be `support-bundle-spec` or `preflight-spec`, not arbitrary filenames like `support-bundle.yaml`. Not obvious from docs, had to inspect the SDK's own secret to discover
- **SDK upload endpoint** — `POST /api/v1/supportbundle` only accepts `application/gzip` with `Content-Length`, but the error messages on failure are just `400 Bad Request` with no detail about what's wrong
- **Busybox wget binary POST** — busybox `wget --post-file` doesn't correctly handle binary uploads to the SDK, returning 400. Requires `curl --data-binary` instead. Alpine images don't include curl by default
- **support-bundle CLI --auto-upload** — targets `replicated.app` (cloud), not the local SDK. Must collect locally and POST to the SDK endpoint separately. This is not documented anywhere obvious
- **support-bundle CLI exit codes** — exits non-zero when any analyzer has warnings/errors. Using `set -e` in wrapper scripts kills the upload step. Not documented
- **exec collector self-targeting** — exec collector silently fails when the support-bundle CLI runs inside the same pod it tries to exec into. No error, just empty output. Had to discover by testing
- **exec collector RBAC** — the SDK's own support bundle spec uses exec collectors to curl the SDK API. These fail silently without `pods/exec` and `create` permissions on the service account. No error messages indicating missing RBAC
- **http collector runs in-cluster** — contrary to initial assumptions, the http collector makes requests from inside the cluster, not client-side. Service DNS resolves correctly. Only fails when running `kubectl support-bundle` from a local machine that can't resolve cluster DNS
- **Vendor Portal Bundle Analysis gaps** — the Bundle Analysis view in the Vendor Portal doesn't display statefulsets, clusterroles, clusterrolebindings, namespaces, or persistentvolumes — even though the data exists as JSON files in the support bundle tar.gz. You have to download and extract the bundle to see these resources. Only deployments, services, pods, ingresses, PVCs, and secrets appear in the portal UI
