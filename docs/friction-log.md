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

### Claude claimed image contents without verifying
**Problem:** When switching the self-signed cert job away from alpine + runtime `apk add`, Claude confidently recommended `alpine/k8s:1.34.7` saying it "ships with kubectl + openssl pre-installed". User deployed; pod failed with `sh: openssl: not found`. alpine/k8s actually only has kubectl, helm, curl, jq — no openssl. Required another PR to split into an initContainer pattern (`alpine/openssl` for generation, `alpine/k8s` for kubectl apply).
**Resolution:** Before recommending any utility image, run `docker run --rm --entrypoint sh IMAGE:TAG -c "for t in tool1 tool2; do which $t || echo MISSING; done"` and paste the output into the commit / PR message. Saved to memory as `feedback_verify_image_contents.md`.
**Lesson:** Docker Hub descriptions and gut-feel aren't verification. Always `docker run` against the exact tag before writing it into a chart. Applies to every utility image swap, not just the cert-job.

### Misdiagnosing install-type regressions from the latest lint error
**Problem:** The vendor portal stopped listing helm-cli as an available install type for new releases. Claude jumped to the newest lint error (`config-is-invalid: readonly must be a bool`) and shipped PR #121 to fix it — availability did NOT return. Another full diagnostic round later, the real cause turned out to be a top-level KOTS-templated Secret (`replicated/cloudflare-api-token-secret.yaml`) introduced many releases earlier, which disqualified the release as helm-installable because plain Helm can't render `kots.io/when` / `repl{{ ConfigOption }}`. Fixed in PR #124.
**Resolution:** The correct debugging path: pull release metadata via the Replicated API (`replicated api get /v3/app/<id>/channel/<id>/releases`) and diff `installationTypes` across releases chronologically to find the exact transition sequence. Then `git log <last-good-tag>..<first-bad-tag>` to find the specific commit that flipped the state.
**Lesson:** Lint errors don't always correlate with availability transitions. For install-type-availability bugs, diff release metadata across the transition first. Saved to memory as `feedback_diff_release_metadata.md`.

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

## Tier 4 — Embedded Cluster v3

### v1beta3 preflight doesn't support KOTS template functions in EC v3 today
**Problem:** I added `ConfigOptionNotEquals`, `IsAirgap`, and similar KOTS template functions to `replicated/preflight-v1beta3.yaml` to gate conditional collectors (external DB FQDN check, Cloudflare API check). Install failed with `helm render: parse error: function "ConfigOptionNotEquals" not defined`.
**Root cause:** EC v3's built-in v1beta3 preflight runner renders via Helm templating (not KOTS template functions). `.Values` isn't accessible because the spec lives outside the chart, so even Helm-style conditionals don't work. The EC team documents this as a known limitation.
**Resolution:** Keep `replicated/preflight-v1beta3.yaml` limited to static cluster-level checks (K8s version, CPU, memory, distribution, storage class). Workload-specific conditional checks (external DB, Cloudflare) live in the chart's v1beta2 `_preflight.tpl` where Helm `.Values` work for Helm CLI installs.
**Time spent:** ~45 minutes chasing this through docs.replicated.com and reading the troubleshoot.sh source.
**Lesson:** v1beta3 preflight is "values-driven" by design but Replicated hasn't wired chart values into the EC v3 runner yet. Until then, cluster-level static checks are the only safe thing there.

### ReplicatedImageName in EC alpha-31 doesn't strip custom-domain prefixes
**Problem:** With chart defaults in custom-domain proxy form (`images.littleroom.co.nz/proxy/drone-rx/...`), the HelmChart CR wrapped them with `ReplicatedImageName` / `ReplicatedImageRegistry` / `ReplicatedImageRepository`. EC online produced double-prefixed image URLs: `images.littleroom.co.nz/proxy/drone-rx/images.littleroom.co.nz/proxy/drone-rx/...`.
**Root cause:** EC alpha-31 source (`pkg/template/image_context.go` @ commit f845ba3) does NOT contain the "return unchanged if input already matches configured ProxyDomain" shortcut. That logic was added later (post-alpha-33) in EC main. Alpha-31 unconditionally prepends the proxy prefix.
**Resolution:** Pass `true` (noProxy) as the 2nd positional arg to every `ReplicatedImage*` call wrapping a value that's already in proxy-prefix form. Per the docs preview at `deploy-preview-3968--replicated-docs-upgrade.netlify.app/vendor/replicated-onboarding-air-gap`: *"The true parameter sets noProxy to true, indicating 'the image reference value in values.yaml already contains the proxy path prefix.'"* Airgap rewriting still works because the isAirgap branch runs before the noProxy check.
**Exception:** Traefik's defaults are upstream form (`docker.io` / `traefik`) — the function SHOULD prepend the proxy path there, so Traefik keeps the unadorned call.
**Time spent:** ~2 hours across multiple failed release + EC install cycles before finding the `true` pattern in the docs preview.
**Lesson:** Before assuming a template function "does the right thing", read the source at the version deployed. Function behavior evolves between alpha releases.

### SDK's `library/replicated-sdk-image` path isn't under `/proxy/<slug>/`
**Problem:** With `ReplicatedImageRepository` wrapping the SDK's repository, EC rendered `images.littleroom.co.nz/proxy/drone-rx/library/replicated-sdk-image:<tag>` — and got `pull access denied`. The SDK image lives at `<proxy-domain>/library/replicated-sdk-image`, NOT under `/proxy/<slug>/library/...`.
**Resolution:** Same `noProxy=true` fix — with `true`, the repository is returned unchanged and the final path matches where the SDK actually lives.
**Lesson:** Not all "Replicated-distributed" images follow the same `/proxy/<slug>/` path shape. The SDK is a special case because it's served from a common `/library/` namespace rather than per-customer proxy.

### Kubernetes default NodePort range (30000-32767) rejects :80/:443 — and EC already extends it
**Problem:** Traefik Service declared `nodePorts: {http: 80, https: 443}` but the service came up as `80:11473, 443:3170` — Kubernetes silently rewrote the ports because they're outside the default NodePort range.
**Resolution considered:** `hostNetwork: true` on Traefik pods (rejected on security grounds — NetworkPolicy bypass, compromised Traefik gets direct access to kubelet/metadata/localhost).
**Resolution attempted:** EC's `spec.unsupportedOverrides.k0s` with `api.extraArgs.service-node-port-range: "80-32767"`. Shipped this… then discovered EC already sets that exact flag by default in its k0s config ([pkg/k0s/config.go L167-169](https://github.com/replicatedhq/ec/blob/0ea20cf0eb442b136a223da13343164cbd873d83/pkg/k0s/config.go#L167-L169)). The override was a no-op.
**Real cause (found later):** The Traefik v3 chart accepts NodePort config at `ports.<entrypoint>.nodePort` (per-entrypoint), NOT at `service.nodePorts.{http,https}`. The latter path is silently ignored and k8s picks a random port from the extended range. See separate entry below.
**Lesson:** Check the upstream project's source/defaults before adding an override. And before concluding "k8s is rewriting my NodePort", verify the values actually reached the Service spec — `kubectl get svc -o yaml` will show whether the declared values made it through.

### Traefik v3 NodePort path is per-entrypoint, not `service.nodePorts`
**Problem:** After extending the NodePort range, Traefik's Service STILL came up with random NodePorts (`80:8043, 443:28326`). The HelmChart CR was setting `service.nodePorts: {http: 80, https: 443}`, which Traefik v3's chart does not recognize — the values were silently dropped at chart render time.
**Resolution:** Move to per-entrypoint config under `ports.web.nodePort` and `ports.websecure.nodePort`. Verified via `helm show values traefik/traefik` — the `ports.<entrypoint>.nodePort` field is explicitly schema-defined; `service.nodePorts` is not.
**Time spent:** Most of the earlier "NodePort range" debugging session was actually this bug — the range-extension override was unnecessary and the real bug was the wrong values path.
**Lesson:** When a values key silently has no effect, the first check should be `helm show values <chart>` on the actual version used — not assuming the chart accepts `service.nodePorts` just because it's a common pattern in other charts.

### alpine/k8s doesn't ship with openssl
**Problem:** Assumed `alpine/k8s:1.34.7` was an all-in-one utility image with openssl + kubectl. Swapped the self-signed cert job from `alpine:3.19` (which was doing a runtime `apk add openssl kubectl` that fails in airgap) to alpine/k8s. Pods then failed with `sh: openssl: not found` — alpine/k8s only ships kubectl, helm, curl, jq.
**Resolution:** Split the Job into an `initContainer` (`alpine/openssl:3.5.6`) that writes `tls.key`/`tls.crt` to an emptyDir, followed by a main container (`alpine/k8s:1.34.7`) that applies the Secret via kubectl.
**Lesson:** Always run `docker run --rm --entrypoint sh IMAGE -c "which tool1 tool2 tool3"` against utility images before committing to them. Saved to memory so I don't repeat it.

### Hardcoded chart-template images don't rewrite for airgap
**Problem:** The self-signed cert job originally hardcoded `images.littleroom.co.nz/proxy/drone-rx/index.docker.io/library/alpine:3.19` in the template. On airgap, containerd tried to pull from the custom domain and timed out (no egress).
**Resolution:** Move the image to `.Values.selfSignedCert.image` and wrap it with `ReplicatedImageName ... true` in the HelmChart CR so airgap rewrites it to the local registry.
**Additional step:** The HelmChart CR `builder:` section must force-render the conditional template (`ingress.tls.mode=self-signed`) during airgap bundle builds, or the bundler doesn't scan the cert job's image refs.
**Lesson:** Every image in the chart must either (a) flow through a HelmChart CR value wrapped in `ReplicatedImage*` or (b) be impossible to reach during airgap. No hardcoded image refs in templates.

### Subcharts need pull secrets wired via subchart-specific value paths
**Problem:** After moving images from the `/anonymous/` proxy path to authenticated `/proxy/` paths, helm-cli installs failed to pull nats/cnpg/cloudnative-pg images with `failed to fetch anonymous token: 400 Bad Request`. The `enterprise-pull-secret` existed in the namespace (the Replicated SDK creates it automatically via `createPullSecret: true`), but the subchart pods weren't referencing it.
**Root cause:** The main dronerx chart's `_helpers.tpl:dronerx.imagePullSecrets` only applies to our own Deployments/Jobs. Subchart pods (nats StatefulSet, cnpg operator Deployment, CNPG-managed postgres StatefulSet) use the subchart's own values paths (`nats.global.image.pullSecretNames`, `cloudnative-pg.imagePullSecrets`, `postgres-cluster.yaml:spec.imagePullSecrets`). The chart defaults for these were `[]`.
**Resolution:** Default each subchart pull-secret list to `[{name: enterprise-pull-secret}]` in `chart/values.yaml`. Pass `imagePullSecrets` through the CNPG Cluster CR spec explicitly (CNPG auto-derives the StatefulSet from Cluster, so the Cluster spec is the only injection point).
**Lesson:** The "enterprise-pull-secret" pattern is load-bearing for every subchart separately. Audit every Pod/StatefulSet/DaemonSet/Deployment source in the rendered chart (`helm template | grep kind:`) to confirm each has an imagePullSecrets reference.

### `tag: latest` override in the HelmChart CR broke airgap image pulls
**Problem:** The HelmChart CR set `api.image.tag: latest` and `frontend.image.tag: latest`, overriding the chart defaults (which release-please keeps in sync with the real semver via `x-release-please-version`). The airgap bundler rendered the chart with builder defaults (so it saw `:1.19.7`) and pushed `dronerx-api:1.19.7` + `dronerx-frontend:1.19.7` into the local registry. At install time the CR override flipped tags to `:latest` → pods tried to pull `<LocalRegistry>/drone-rx/dronerx-api:latest` which doesn't exist in the bundle → `NotFound`.
**Error:** `Failed to pull image "10.244.128.11:5000/drone-rx/dronerx-api:latest": not found`.
**Resolution:** Delete the `tag: latest` lines from the HelmChart CR. Chart defaults flow through and always match whatever release-please bumped.
**Lesson:** Don't override the chart default tag in the HelmChart CR unless you're also overriding it in `builder:` — the bundler and the runtime need to see the same tag. Simplest: let the chart default (semver-managed) be authoritative.

### CNPG deployment naming differs between helm-CLI and KOTS/EC installs
**Problem:** Support bundle analyzers hardcoded `drone-rx-cloudnative-pg` as the CNPG operator deployment name. That works for helm-CLI installs where the operator is a subchart of drone-rx (release name `drone-rx` prefixes the subchart's fullname). On EC/KOTS the operator is installed via a separate HelmChart CR (`cnpg-operator-chart.yaml`) and KOTS uses the chart name (`cloudnative-pg`) as the Helm release name — the upstream fullname helper collapses release==chartname to just `cloudnative-pg`. Bundle failed with `deployment "drone-rx-cloudnative-pg" was not found`.
**Resolution:** Gate the `deploymentStatus` analyzer on `.Values.cnpgOperator.managed`. True (helm-CLI subchart path) → check `<release>-cloudnative-pg`. False (KOTS/EC separate-release path) → check bare `cloudnative-pg`.
**Lesson:** KOTS HelmChart CRs name the Helm release after `spec.chart.name`, not the CR's `metadata.name`. Any analyzer / support-bundle check that references subchart-managed resource names needs the same conditional you use for the chart dependency itself.

### ConfigMap-mounted env vars don't auto-propagate on `helm upgrade`
**Problem:** Reported as "EC v3 isn't applying config changes on upgrade". Operator unticked `light_mode_enabled` / `admin_link_visible` in the KOTS config screen and redeployed — UI didn't update. Pods kept their cached env vars until something restarted them.
**Root cause:** Kubernetes doesn't automatically restart pods when a referenced ConfigMap's content changes (applies to `envFrom` / `valueFrom: configMapKeyRef` / mounted volumes without subPath). Not EC-specific.
**Resolution:** Add a `checksum/config` annotation to the api Deployment's pod template derived from the rendered `configmap-api.yaml`:
```yaml
template:
  metadata:
    annotations:
      checksum/config: {{ include (print $.Template.BasePath "/configmap-api.yaml") . | sha256sum }}
```
When the ConfigMap's content changes, the annotation changes → Helm treats it as a pod-spec diff → rolling restart → new pods read new env vars.
**Lesson:** Any deployment that consumes a ConfigMap via env should have this checksum annotation, otherwise "I changed config, nothing happened" bugs are baked in.

### `dronerx.imagePullSecrets` helper emitted duplicate entries on KOTS
**Problem:** After defaulting the top-level `imagePullSecrets: [{name: enterprise-pull-secret}]` in values.yaml (to fix helm-cli), KOTS installs produced a Kubernetes warning: `spec.template.spec.imagePullSecrets[1].name: duplicate name "enterprise-pull-secret"`.
**Root cause:** The helper had two independent producers: (1) if `global.replicated.dockerconfigjson` is set (KOTS injects this), emit `- name: enterprise-pull-secret`; (2) iterate `.Values.imagePullSecrets` and emit each. With both active, `enterprise-pull-secret` appeared twice.
**Resolution:** Build a dedup'd list in the helper (`has .name $names` check before append), then emit once.
**Lesson:** If a helper has two code paths that can produce the same item, dedupe explicitly. K8s tolerates duplicate imagePullSecret entries but emits a warning that clutters logs.

---

## Tier 5 — Config Screen

### KOTS config schema rejects templated `readonly`
**Problem:** Implemented license-gated toggles (live_tracking_enabled, light_mode_enabled) with `readonly: 'repl{{ not (LicenseFieldValue ... | ParseBool) }}'` — the intent was to grey out features the customer isn't entitled to, while still displaying them for marketing/discovery. Release lint reported:
```
config-is-invalid  failed to decode config content: json: cannot unmarshal string
                   into Go struct field ConfigItem.spec.groups.items.readonly of type bool
```
**Impact:** The schema error invalidated the release for helm-cli / existing-cluster installs, leaving only EC available. (The vendor portal UI tolerates some config errors for EC rendering but not for helm-cli.)
**Resolution:** Drop the templated `readonly`. Keep the `default` templated from `LicenseFieldValue` so entitled users start with the feature on. Entitlement is still enforced at runtime by the HelmChart CR `and`-guard: `repl{{ and (ConfigOptionEquals "X" "1") (LicenseFieldValue "X" | ParseBool) }}` — operators without entitlement can toggle the config but the final value stays false.
**Lesson:** KOTS config schema fields typed as `bool`, `int`, `string` must receive literal values, not templated strings. Use templated `default` for initial value and runtime `and`-guards for entitlement enforcement.

### Top-level KOTS-templated Secret kills helm-cli install availability
**Problem:** Long-running regression — since release 1.19.1 (seq 184), the vendor portal stopped listing `helm` under `installationTypes` for new releases. Only KOTS + EC remained. I initially misdiagnosed as the `readonly` schema error above; fixing that didn't restore helm availability.
**Debugging approach:** Pulled release metadata for the Unstable channel via `replicated api get /v3/app/<id>/channel/<id>/releases` and diff'd `installationTypes` across 20 releases chronologically. `helm` was present through seq 177 (1.18.11), absent starting seq 184 (1.19.1). `git log v1.19.0..v1.19.1` surfaced commit `c33f13f` which added `replicated/cloudflare-api-token-secret.yaml` — a top-level plain Kubernetes Secret using `kots.io/when` and `repl{{ ConfigOption }}` in stringData.
**Root cause:** Replicated's release validator classifies a release as helm-installable only if ALL resources can be rendered by plain Helm. A top-level K8s resource using KOTS template syntax disqualifies the release for helm-cli because Helm has no way to process `kots.io/when` or `repl{{ }}`.
**Resolution:** Move the Secret into `chart/templates/` as a Helm-conditional template. Plumb the token value through `.Values.ingress.tls.cloudflare.apiToken` — KOTS HelmChart CR populates it from the ConfigOption; helm-cli users pass `--set ingress.tls.cloudflare.apiToken=xxx`. Net token exposure is identical (still ends up in the same Secret).
**Time spent:** ~1.5 hours including the misdiagnosed PR.
**Lesson:** When install-type availability regresses, diff release metadata across the transition (`replicated api get .../releases`) to find the actual breaking commit. Lint errors don't necessarily correlate with availability transitions. Saved to memory as a debugging playbook.

### `when`-gated ConfigOptions render as empty string on external-mode installs
**Problem:** EC online install with `database_type=external` failed during helm install:
```
at '/postgresql/instances': minimum: got 0, want 1
```
**Root cause:** `postgres_instances` is `when`-gated to embedded DB (`ConfigOptionEquals "database_type" "embedded"`). On external installs the field is hidden. `ConfigOption "postgres_instances"` returns `""` (not the field's `default: "1"`). `ParseInt ""` = 0 — fails the chart's `values.schema.json` rule `postgresql.instances minimum: 1`. The postgres Cluster CR is gated on `postgresql.enabled=false` and never actually renders, but schema validation runs before template conditionals.
**Resolution:** Defensive `| default "1"` (sprig) in the HelmChart CR so empty values get sane literals. Only kicks in when the field is hidden. Helm-CLI wasn't affected because it uses chart defaults (instances: 1).
**Lesson:** `when`-gated config fields return empty strings, NOT their declared default, when hidden. Any consumer of such a field must handle empty explicitly.

### Quoting the rendered template flips the YAML type
**Problem:** After adding `| default "1" | ParseInt` for postgres_instances, wrapped the whole template in single quotes for visual consistency: `'repl{{ ... | ParseInt }}'`. Next install failed:
```
at '/postgresql/instances': got string, want integer
```
**Root cause:** Single quotes make the YAML value a string literal. The template rendered `1` but YAML parsed it as the string `"1"`, not integer `1`.
**Resolution:** Drop the outer quotes. `storage.size` keeps quotes because it's supposed to be a string (`"1Gi"`).
**Lesson:** YAML type inference on rendered templates is subtle. Quote only string-typed values; leave numerics, booleans, and null unquoted so YAML parses them as their native types.

### RandomString + `default` for install-stable generated secrets
**Problem:** Rubric 5.2 asks for an auto-generated DB password that survives upgrades. Templating `value:` with `RandomString` re-generates on every config render (password churns). Templating `default:` with `RandomString` evaluates once and caches — the intended behavior.
**Resolution:** `default: 'repl{{ RandomString 32 }}'` with `type: password`. KOTS stores the rendered string as the config value; subsequent renders see the cached value. Chart creates a basic-auth Secret from this value and points CNPG's `bootstrap.initdb.secret.name` at it.
**Lesson:** For auto-generated values that must persist across upgrades, use `default:` (evaluate-on-first-render-then-cache), not `value:` (re-evaluate every render).

### `readonly: true` is server-side-only — the KOTS admin UI doesn't visually lock the input
**Problem:** Rubric 4.7 asks for feature items to be "hidden or locked" when the license lacks the entitlement. We tried the locked path first with a two-item swap pattern (editable item `when: license=true`, placeholder item `readonly: true, when: license=false`). Operator reported the locked placeholder still rendered as a clickable checkbox — they could tick it. Swapped the placeholder to `type: text` with `value: "🔒 Locked — license upgrade required"` + `readonly: true` — still rendered as an ordinary editable text field showing `0` that the user could type into.
**Root cause:** KOTS admin console enforces `readonly` **server-side only** (rejects saves), but doesn't visually disable the input on either `type: bool` or `type: text`. Users see an editable widget even though their edits don't persist.
**Resolution:** Rubric 4.7 permits "hidden OR locked". Since "locked" isn't visually enforced, switch to **hidden** via `when: LicenseFieldValue=true` — non-entitled operators don't see the item at all. Drop the placeholder.
**Time spent:** Two PRs (#133, #135) trying different flavors of "locked" before accepting that hidden is the only reliable option today.
**Lesson:** Don't trust that `readonly` in admin UI means visually-disabled — it only means "save rejected". For "locked" UX you need a type the UI genuinely can't interact with (none exists for these types today).

### KOTS config groups with no visible items are automatically hidden
**Problem:** After switching to `when`-based hiding for the license-gated `live_tracking_enabled` item, the entire "Features" group disappeared from the config screen for non-entitled licenses — group title, description, and all. I'd wanted the group heading + description to remain visible as an upsell signal.
**Root cause:** KOTS auto-hides config groups that contain zero visible items. Useful UX default in most cases (no empty sections), but removes any "this feature exists, upgrade to access" messaging that the group description was carrying.
**Resolution:** Accept the behavior. Marketing / README / sales outreach is the right place for upsell — not the config screen. Alternative if the upsell signal is critical: add a single always-visible informational item to the group, but be aware it will render as editable (see prior entry).
**Lesson:** Config groups aren't persistent section headers — they only render if at least one child item renders.

### Rubric 4.7 vs 5.1 taxonomy — separate the paying features from plain toggles
**Problem:** Initially put both `live_tracking_enabled` (license-gated) and `light_mode_enabled` (UI preference) under the license-gated three-layer pattern. That over-gated the theme toggle — it's not a paying feature and had no business on the license path.
**Resolution:** Split clearly:
- **Rubric 4.7** (license entitlement, hidden/locked): `live_tracking_enabled` — via `when: LicenseFieldValue=true`. Tier 6's `terraform` will follow the same pattern.
- **Rubric 5.1** (≥2 non-trivial plain config features): `light_mode_enabled` + `admin_link_visible` — plain bool toggles, no license involvement. Each observably changes app behavior (theme toggle appears/disappears in header; Admin link appears/disappears).
Introduced a new `/api/config/ui` endpoint to cleanly separate UI toggles from license-status responses.
**Lesson:** Before adding license gating to a field, ask: "does this feature need to be revenue-protected, or is it just an operator preference?" Only revenue-protective features belong on the license path.

### Regex validation on text fields
**Problem:** `tls_email` is used by Let's Encrypt during cert issuance; a malformed address silently fails the issuance later. No compile-time check catches it.
**Resolution:** Add `validation.regex.pattern` + `message` to the config item. Similar pattern applied to `webhook_url` (allow blank or http(s) URL).
**Note:** The config schema linter emits `config-option-password-type` warnings on text fields whose name contains "secret"/"password"/"token" — for `tls_existing_secret_name` (which holds a K8s resource NAME, not the cert bytes) this is a false positive. Left as a warning; not acting on it.

---

## Tier 4/5 — CI / Release automation friction

### GITHUB_TOKEN-pushed release-please tags don't trigger downstream workflows
**Problem:** Release-please opened a bump PR. After merging it, the "Replicated Release" workflow didn't run automatically — the tag was created by `github-actions[bot]` (via default `GITHUB_TOKEN`), and GitHub's anti-recursion rule prevents those pushes from firing downstream workflows. The release workflow run was showing `action_required` and blocked behind the repo's "first-time-contributor" approval gate.
**Resolution:** Create a fine-grained PAT scoped to this repo with `Contents: RW + Pull requests: RW`, store as `RELEASE_PLEASE_TOKEN`, pass into `googleapis/release-please-action`. PRs and tags are then authored by the user, triggering downstream runs normally.
**Lesson:** `secrets.GITHUB_TOKEN` is never going to cascade — this applies to any workflow that relies on "merge → tag → build".

### Release pipeline ran twice per release-please merge
**Problem:** Every release-please merge fired both the "Release Please" workflow (which called `release.yaml` via `workflow_call`) AND the "Replicated Release" workflow (triggered by the tag push). Same build/promote/attach ran twice.
**Root cause:** The `workflow_call` chain from release-please.yaml was a legacy workaround for the GITHUB_TOKEN tag-push recursion limitation. With the PAT in place (previous entry), the tag push now triggers release.yaml directly.
**Resolution:** Drop the `replicated-release` job from `release-please.yaml`. `release.yaml` stays the single entry point via `on.push.tags: ['v*.*.*']`. Manual `git tag && git push --tags` still works as an emergency release path.
**Lesson:** Workarounds accrete. Re-check for redundancy whenever the root constraint is lifted.

### Release-please silently skipped a release because of merge-commit parse failures
**Problem:** Merged a PR that should have cut v1.19.13 (a `fix:` commit fixing the ConfigMap-checksum rollout). No release PR appeared. release-please log said:
```
commit could not be parsed: 3d5b4cd... Merge pull request #137 from jmboby/fix/config-change-rollout
error message: Error: unexpected token ' ' at 1:6, valid tokens [(, !, :]
commits: 0
✔ No commits for path: ., skipping
```
release-please tried to parse the **merge commit message** (`"Merge pull request #137 from..."`) as a conventional commit. "Merge" isn't a valid type, parser chokes, considers zero parseable commits, skips the release. The actual `Fix(chart):` commit inside the PR got dropped during commit-splitting.
**Resolution:**
1. Immediate: manually tagged `v1.19.13` (release.yaml fires on `v*.*.*` tag push) to unblock.
2. Permanent: switched the repo to **Squash-and-merge**. Each merged PR becomes a single commit on main using the PR title as the commit message. As long as PR titles follow conventional-commits (`fix: ...`, `feat: ...`) release-please parses them cleanly every time, and the merge-commit garbage problem disappears.
**Lesson:** Any repo that uses release-please + GitHub's default "Create a merge commit" strategy has this latent failure mode. Switch to squash-merge early — it also makes history cleaner (no internal PR commits polluting main).

### PR workflow ran on release-please-only bump PRs
**Problem:** Release-please bump PRs (only touching `CHANGELOG.md` + `.release-please-manifest.json`) were running the full lint/build/cmx pipeline. Pure waste — the code was already validated on the feature PR.
**Resolution:** Add `paths-ignore` to `pr.yaml`'s `pull_request:` trigger for exactly those two files. GitHub only skips when ALL changed paths match the ignore list, so regular PRs that also edit CHANGELOG still run CI.
**Lesson:** `paths-ignore` is safer than branch-name filtering for "skip bot PRs" because it doesn't couple to bot-specific branch naming.

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
