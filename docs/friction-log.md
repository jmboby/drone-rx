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

---

## General Observations

### What worked well
- The `.replicated` config file for release packaging — simple, declarative
- release-please for semver management — clean flow with Release PRs
- CloudNativePG as a subchart — once the webhook timing was solved, very clean
- CMX k3s clusters for CI testing — fast provisioning, realistic environment
- Replicated SDK for license gating — runtime queries with no-redeploy updates

### What could be improved
- **Documentation inconsistency** — RBAC resource names shown in mixed case in docs but require lowercase in config
- **prepare-cluster action** — doesn't handle image proxy auth, should probably be deprecated in favour of CLI
- **Proxy registry auth** — not obvious that `registry.replicated.com` (OCI chart pull) and `proxy.replicated.com` (image proxy) are different auth domains
- **Error messages** — many Replicated API errors are generic (403, 400) without indicating which specific permission is missing
- **CLI `--auto` flag** — confusing that it ignores `.replicated` config rather than enhancing it
