# Replicated Product Feedback — DroneRx Bootcamp

Compiled from building DroneRx end-to-end (Tiers 0-5: Helm chart → Replicated distribution → SDK → support bundles → KOTS config screen → EC v3). Ordered by product area. Every item is something a vendor hit during a realistic build-out.

---

## 🖥️ Vendor Portal

- **RBAC resource names docs show mixed case, API requires lowercase.** Docs say `KOTS/app/*/read`; real policy needs `kots/...`. Multiple 403 iterations before figuring it out. Also `kots/cluster/*/kubeconfig` is a separate permission from `kots/cluster/*`, which isn't obvious.
- **403 errors don't tell you which permission is missing.** Just "Not authorized". Tightening RBAC becomes guess-and-check.
- **Bundle Analysis view is incomplete.** The tarball on disk has JSON for statefulsets, clusterroles, clusterrolebindings, namespaces, persistentvolumes — but none of those render in the portal UI. Only deployments, services, pods, ingresses, PVCs, and secrets show. Have to download + extract to see the full picture.
- **"Custom domains → set as default" is release-gated.** Only releases promoted **after** flipping the default use the new domain. Not obvious from the UI — operators wondering why old releases still pull from `proxy.replicated.com`.

---

## 🧩 KOTS Admin Console

- **`readonly: true` is server-side-only — the UI doesn't visually lock the input.** Tested on both `type: bool` (checkbox remains tickable) and `type: text` (input remains editable). Rubric 4.7 asks for "hidden or locked" feature items — "locked" currently isn't UX-enforceable. Had to fall back to `when`-based hiding, which means we lose the upsell signal for non-entitled customers.
- **Config groups with zero visible items auto-hide entirely.** Surprising when all items in a group are `when`-gated off — the section title/description disappears too, not just the items.
- **`readonly` can't be templated.** Schema requires a literal bool. Blocked a natural pattern (`readonly: 'repl{{ not LicenseFieldValue ... }}'`) and caused `config-is-invalid` lint errors that silently disqualified releases from helm-cli availability. Took a release-metadata diff to pin down what was actually breaking.
- **Top-level KOTS-templated resources silently break helm-cli install-type availability.** A plain K8s Secret in `replicated/` using `kots.io/when` + `repl{{ ConfigOption }}` makes the release non-renderable by plain Helm, so Replicated drops `helm` from `installationTypes`. No lint warning calling this out; had to diff release metadata across 20 releases to find the commit that introduced it.
- **`when`-gated ConfigOptions return empty string, not the declared default.** If a config item is hidden, `ConfigOption "foo"` returns `""` rather than the `default:` you configured. Downstream consumers (HelmChart CR, templates, chart schema) break unexpectedly. Required defensive `| default "1"` everywhere.
- **YAML type inference on rendered templates is brittle.** Wrapping a template in single quotes turns `ParseInt`'s output into a YAML string and fails schema. Unquoting is required for numerics. Not documented; error surfaces only at install time as a cryptic schema violation.
- **ConfigMap-mounted env vars don't auto-propagate on upgrade.** This isn't Replicated-specific (it's K8s), but it bites hard in KOTS context because operators expect "change config, redeploy, it works". They need to know to add `checksum/config` annotations. Worth calling out in a best-practices guide for KOTS apps.
- **`apiTokenSecretName` / secret-named text fields lint as `config-option-password-type: Warning`.** False positive — the field holds a K8s resource **name**, not a credential. No way to suppress per-item.

---

## ⚙️ Embedded Cluster v3

- **`ReplicatedImageName` (alpha-31) doesn't strip custom-domain prefixes — silently double-prefixes.** Chart defaults in custom-domain form (`images.example.com/proxy/<slug>/...`) get wrapped to `images.example.com/proxy/<slug>/images.example.com/proxy/<slug>/...`. The "already-proxied, return unchanged" shortcut only exists in EC main, not in alpha-31 releases customers actually deploy. Workaround is the undocumented `true` (noProxy) 2nd arg — found only via a docs *preview* URL, not the public docs.
- **`ReplicatedImageRepository` mishandles SDK path.** `library/replicated-sdk-image` lives at `<domain>/library/...`, NOT under `/proxy/<slug>/library/...`. Without `noProxy=true` the function wrongly prepends and the SDK pod gets `pull access denied`. Not obvious from the docs.
- **`noProxy=true` flag is near-impossible to discover.** Current EC docs don't document it. Had to find it in a deploy-preview docs URL (`deploy-preview-3968--replicated-docs-upgrade.netlify.app`). Vendors using custom domain defaults have no realistic way to find the working pattern.
- **v1beta3 preflight runner in EC doesn't support KOTS template functions AND doesn't wire chart values through either.** So `ConfigOption`/`IsAirgap` fail with `function not defined`, and `.Values.x.enabled`-style conditionals also fail because `.Values` isn't populated for the runner. That leaves v1beta3 specs effectively static-cluster-check-only. Workload-specific checks have to stay in the chart's v1beta2 `_preflight.tpl` (Helm-CLI only). This is mentioned as a limitation in memory but not in public docs.
- **`--service-node-port-range=80-32767` is already the default, but isn't documented.** Added an `unsupportedOverrides.k0s` to extend it for Traefik :80/:443 — was a no-op. Found the EC default by reading `pkg/k0s/config.go`. Would save vendors hours to call this out in EC docs.
- **Airgap bundler only includes images from templates that render under `builder:`**. Conditionally-rendered templates (e.g. self-signed cert job gated on `tls.mode=self-signed`) get skipped unless the `builder:` overrides force them. Not obvious; silent "image missing in airgap" bugs.
- **Install-time lock-in on `unsupportedOverrides`.** `spec.api` / `spec.storage` can't be modified after first install — any mistake requires a full cluster rebuild. Documented, but the escape-hatch name ("unsupported") implies reversibility when it isn't.

---

## 📦 Replicated SDK

- **`licenseID` vs `LicenseID` casing mismatch.** The SDK's `/api/v1/license/info` returns `licenseID` (lowercase-i, capital-D). Easy to get wrong when reading docs; should match camelCase conventions or be documented with the exact casing.
- **`IsExpired` isn't a field — must be derived from `expiresAt`.** The `/license/info` payload has no boolean; clients have to parse the date and compare. Worth adding.
- **License fields return typed values, not strings.** Boolean license fields return JSON `true`/`false`, not `"true"`/`"false"`. Needs a `ParseBool` guard OR a type-aware client. Docs aren't explicit about this.
- **`nameOverride` on the SDK subchart doesn't prepend the release name.** Other Bitnami-style subcharts do. Trips up resource-name patterns across the chart.
- **SDK upload endpoint error messages are opaque.** `POST /api/v1/supportbundle` with anything wrong returns `400 Bad Request` with no body. Needed bytes of body + specific Content-Type (`application/gzip`) + Content-Length — figured out by trial and error.
- **`createPullSecret: true` is load-bearing and under-documented.** The SDK quietly creating `enterprise-pull-secret` is what makes the whole image-pull-secrets pattern work on helm-CLI installs, but that's only mentioned via a `values.yaml` comment: "If false, you must create a secret named enterprise-pull-secret yourself."

---

## 🛠️ Replicated CLI

- **`--auto -y` ignores `.replicated` config file and looks in `./manifests`.** Confusing — users expect it to enhance config, not bypass it. Remove `--auto -y` and CLI reads `.replicated` automatically. Docs don't flag this.
- **`replicated customer create --helm-install` requires `--email`, error only at runtime.** Not in the help text. Should surface as a required flag.
- **`replicated channel delete <name>` errors — must use ID.** `archive app channel: Not found` is cryptic. Accepting both name and ID (or erroring clearly) would save iterations.
- **Stdout warnings corrupt redirected output.** Piping `replicated cluster kubeconfig > kubeconfig.yaml` includes warning lines inline, breaking subsequent kubectl calls. Warnings should go to stderr.
- **`--kots-install` defaults to true, breaks helm-only channels.** Creating a KOTS-enabled customer against a Helm-only release fails with "Cannot assign customer with KOTS install enabled to a channel with a helm-cli-only release." Customer-create's defaults should match the release type or surface the mismatch earlier.

---

## 🔍 Troubleshoot (Preflight + Support Bundle)

- **Preflight `run` collectors write to `<collectorName>.log`, but reference docs show `preflight/<collectorName>.txt`.** Support bundles use a different path (`<collectorName>/stdout.txt`). Two different conventions, neither clearly documented. Required extracting a real bundle to find the right path.
- **Troubleshoot CRDs don't exist on helm-CLI clusters.** Standalone `kind: Preflight` resources in the chart fail with `no matches for kind "Preflight"` on vanilla K8s. The "Secret-with-specific-data-key" workaround isn't prominently documented.
- **Secret data key must be exactly `preflight-spec` or `support-bundle-spec`.** Not arbitrary filenames. Had to inspect the SDK's own support bundle Secret to discover this.
- **`support-bundle` CLI exits non-zero on warnings.** `set -e` wrapper scripts kill subsequent upload steps. Not documented.
- **`support-bundle` CLI `--auto-upload` targets `replicated.app` (cloud), not the in-cluster SDK.** Vendors who want bundles uploaded to their SDK pod (for viewing via the admin console) have to collect locally + POST separately. Not obvious; no docs.
- **`exec` collector silently fails when targeting its own pod.** If the support-bundle CLI runs inside a pod and tries to exec into the same pod, the collector returns empty with no error. Had to discover by testing.
- **`exec` collector RBAC requirements aren't surfaced.** Missing `pods/exec` + `create` permissions on the ServiceAccount cause silent empty output. Should fail loudly with a clear message.
- **`http` collector runs from inside the cluster, not the client machine.** Most docs imply client-side. Only fails when running locally without cluster DNS.
- **Empty `collectors: []` triggers default collection.** Rendering `collectors:` with no entries (because all conditionals were false) causes a very slow full cluster-info gather instead of a no-op. Vendors have to omit the `collectors:` section entirely when empty.

---

## 📚 Docs

- **`noProxy=true` parameter for `ReplicatedImage*` functions is undocumented** in the public EC docs. Documented only in a deploy-preview URL. This is the only working pattern for vendors who keep their custom domain in chart default values.
- **EC's default k0s config (including `service-node-port-range=80-32767`) isn't documented.** Requires reading `pkg/k0s/config.go` source. Vendors end up adding redundant `unsupportedOverrides.k0s` they don't need.
- **Proxy auth model isn't explicit.** `registry.replicated.com` (OCI chart pull) and `proxy.replicated.com` (image proxy) have different auth flows. Mixed up constantly — no single "which domain, when" doc.
- **`readonly` field docs say "applies uniformly across types"** but don't mention the UI-enforcement gap for `type: bool` / `type: text`. Vendors implement based on the docs, ship, then discover the admin console renders editable.
