# Changelog

## [1.15.1](https://github.com/jmboby/drone-rx/compare/v1.15.0...v1.15.1) (2026-04-15)


### Bug Fixes

* CNPG and NATS reloader image paths for EC v3 ([df52edd](https://github.com/jmboby/drone-rx/commit/df52edd7122d169d0bdd266a0a5d8cac10f0b088))
* CNPG and NATS reloader image paths for EC v3 ([41c87f1](https://github.com/jmboby/drone-rx/commit/41c87f1ff93b55d8d72b9e1bfed45d0c0e494dec))
* sync chart version with release version in PR workflow ([b80d302](https://github.com/jmboby/drone-rx/commit/b80d302c9453d36e6e3423550f5f3fa7415b9606))
* use channel slug instead of name in PR helm install ([e921d1a](https://github.com/jmboby/drone-rx/commit/e921d1ab14546d93491cac99823a36381895080c))

## [1.15.0](https://github.com/jmboby/drone-rx/compare/v1.14.0...v1.15.0) (2026-04-15)


### Features

* split image fields into registry/repository for EC v3 ([f2f38a7](https://github.com/jmboby/drone-rx/commit/f2f38a7c5bcb1fbc4a3daedcf178bac6d2c9f505))
* split image fields into registry/repository for EC v3 ([16bbd95](https://github.com/jmboby/drone-rx/commit/16bbd95f61c6affcefd16a66c14dd8a3bbdfc264))

## [1.14.0](https://github.com/jmboby/drone-rx/compare/v1.13.0...v1.14.0) (2026-04-15)


### Features

* add PostgreSQL instances to KOTS config screen ([fdde19f](https://github.com/jmboby/drone-rx/commit/fdde19f416c8f3d4a817ac5bbcd0699204758216))

## [1.13.0](https://github.com/jmboby/drone-rx/compare/v1.12.1...v1.13.0) (2026-04-15)


### Features

* replace Svelte logo with DroneRx drone icon ([d5135bf](https://github.com/jmboby/drone-rx/commit/d5135bfab9cd91b36afce9d7fb522355a80dd141))
* replace Svelte logo with DroneRx drone icon ([fa2a6ad](https://github.com/jmboby/drone-rx/commit/fa2a6ad8578aa14fa5624bc5db49f8232058f6c5))


### Bug Fixes

* use app's actual drone icon for favicon ([f2a1cf9](https://github.com/jmboby/drone-rx/commit/f2a1cf9a7dc76ad9c73feb5245ef20c6044f3455))

## [1.12.1](https://github.com/jmboby/drone-rx/compare/v1.12.0...v1.12.1) (2026-04-15)


### Bug Fixes

* quote LicenseFieldValue as string, remove port minimum ([9538774](https://github.com/jmboby/drone-rx/commit/953877476e911a4aaca01996749e967c567ff05c))
* run dev-release from repo root ([df4223a](https://github.com/jmboby/drone-rx/commit/df4223a3db1100d825e178ec1b96786e853cbc77))
* run dev-release script from repo root ([2c680cc](https://github.com/jmboby/drone-rx/commit/2c680cc16b3ddaddd57a67fe5e99629c438d10b7))
* use latest image tag in dev-release script ([24d746a](https://github.com/jmboby/drone-rx/commit/24d746a210256c06d2d6df4f768837488ea33ec1))

## [1.12.0](https://github.com/jmboby/drone-rx/compare/v1.11.2...v1.12.0) (2026-04-15)


### Features

* add dev-release script for manual EC testing ([7638a83](https://github.com/jmboby/drone-rx/commit/7638a8393f6f6fbd2186e03bbbd2612702d7b587))


### Bug Fixes

* remove cert-manager extension from EC config ([5f6d359](https://github.com/jmboby/drone-rx/commit/5f6d3595d0398bd8bc4de57f6ad5f56355822f4d))
* remove cert-manager extension to unblock EC testing ([f4acbd8](https://github.com/jmboby/drone-rx/commit/f4acbd8051abdb365abcc3d551b9ec05732f1b3a))

## [1.11.2](https://github.com/jmboby/drone-rx/compare/v1.11.1...v1.11.2) (2026-04-14)


### Bug Fixes

* revert to helmCharts format for EC cert-manager extension ([7900fd4](https://github.com/jmboby/drone-rx/commit/7900fd499b2b4b34629da828d9fb7b595875517f))
* revert to helmCharts format for EC extensions ([42d848b](https://github.com/jmboby/drone-rx/commit/42d848b2ee2a09c84647577076364f83169d6e31))

## [1.11.1](https://github.com/jmboby/drone-rx/compare/v1.11.0...v1.11.1) (2026-04-14)


### Bug Fixes

* correct EC Helm extension format for cert-manager ([4a50dd9](https://github.com/jmboby/drone-rx/commit/4a50dd9b05145cab70e4ba7293a5c5d6a5d030ba))
* correct EC Helm extension format for cert-manager ([e7326e8](https://github.com/jmboby/drone-rx/commit/e7326e828e2af5d5ad257bc58807433499d6e74b))

## [1.11.0](https://github.com/jmboby/drone-rx/compare/v1.10.0...v1.11.0) (2026-04-14)


### Features

* add minimal KOTS Config for EC admin console ([81fe6b3](https://github.com/jmboby/drone-rx/commit/81fe6b38ff95135c7ed9115ea215569d267ff81a))
* add minimal KOTS Config for EC admin console ([037bd24](https://github.com/jmboby/drone-rx/commit/037bd247840ae0b3d8638231f86c6e2a5b955185))

## [1.10.0](https://github.com/jmboby/drone-rx/compare/v1.9.5...v1.10.0) (2026-04-14)


### Features

* add cert-manager as EC Helm extension for TLS support ([535b28a](https://github.com/jmboby/drone-rx/commit/535b28a90a4a97bfb3a278e2ab60dc9efd5b97ea))
* add Embedded Cluster v3 config for VM installs ([33364b6](https://github.com/jmboby/drone-rx/commit/33364b683723842f0634889330e420607be64827))
* add env var fallback for live_tracking_enabled entitlement ([05aa0e4](https://github.com/jmboby/drone-rx/commit/05aa0e458c030a55d9157e6e454204eedca57d22))
* add LicenseFieldValue and EC defaults to HelmChart CR ([9291cfa](https://github.com/jmboby/drone-rx/commit/9291cfa8ed4716193992032492b91cfb9f903876))
* add static v1beta3 preflight for EC installer ([9c2803e](https://github.com/jmboby/drone-rx/commit/9c2803e868847f8756b3d38102280c1792095085))
* set self-signed TLS defaults for EC installs ([20d6907](https://github.com/jmboby/drone-rx/commit/20d69071010c4c9351f1ea4387ed5a6c058bd345))
* Tier 4 — Embedded Cluster v3 + cert-manager TLS ([ea0d1e1](https://github.com/jmboby/drone-rx/commit/ea0d1e1774e7dac1b5a1ffdee7bfe033863cce10))
* wire liveTrackingEnabled through chart values and configmap ([4f61b4b](https://github.com/jmboby/drone-rx/commit/4f61b4b6c6c818e1c61e3d76cb9c6b4838184310))


### Bug Fixes

* correct service and deployment names in kots-app.yaml for EC ([e4c254f](https://github.com/jmboby/drone-rx/commit/e4c254f2c7dc33e4ff37a7b0549be1f324a7f67d))

## [1.9.5](https://github.com/jmboby/drone-rx/compare/v1.9.4...v1.9.5) (2026-04-14)


### Bug Fixes

* support bundle collection and upload from /admin ([0b6244a](https://github.com/jmboby/drone-rx/commit/0b6244af5a728202ef6735bee7cedb46a0e00e80))
* support bundle upload, health collector, RBAC, and NATS status ([098042e](https://github.com/jmboby/drone-rx/commit/098042e2d1ea2d40dfa7911e79c714b8de3590f3))

## [1.9.4](https://github.com/jmboby/drone-rx/compare/v1.9.3...v1.9.4) (2026-04-13)


### Bug Fixes

* broaden 3.5 failure pattern to catch SDK, DB, and NATS errors ([39b8c95](https://github.com/jmboby/drone-rx/commit/39b8c958cc3b8799eaf40fcb9c557d4fdaf9d6b7))
* broaden support bundle 3.5 analyzer to catch SDK, DB, and NATS errors ([6e7094b](https://github.com/jmboby/drone-rx/commit/6e7094b31a7e91bd25466ab9b8ab3a94ed437662))

## [1.9.3](https://github.com/jmboby/drone-rx/compare/v1.9.2...v1.9.3) (2026-04-13)


### Bug Fixes

* correct Content-Length when uploading support bundle to SDK ([634f023](https://github.com/jmboby/drone-rx/commit/634f0235865019bcabac53cd32fc360ee8ad2d04))
* ensure correct Content-Length when uploading bundle to SDK ([0471d1d](https://github.com/jmboby/drone-rx/commit/0471d1d3f0817f496199ec89bf6ec7acba2e8614))

## [1.9.2](https://github.com/jmboby/drone-rx/compare/v1.9.1...v1.9.2) (2026-04-13)


### Bug Fixes

* correct health endpoint analyzer file path in support bundle ([38d5f4e](https://github.com/jmboby/drone-rx/commit/38d5f4e9ccb35aefec2ee52be98e8369f8076ce2))
* scope DB/NATS failure analyzer to api.log only ([b04702e](https://github.com/jmboby/drone-rx/commit/b04702ea330652e708fe8000a7e317fe3c8bbfe2))
* scope gitignore api pattern to root build artifact only ([e781590](https://github.com/jmboby/drone-rx/commit/e7815908de27ac5dcaec387edc10b8973cc147d2))
* upload support bundle to local SDK + gitignore scope ([afd9657](https://github.com/jmboby/drone-rx/commit/afd96577563efb78aea8c3d145c32d27cb975bf7))
* upload support bundle to local SDK, not replicated.app ([07db982](https://github.com/jmboby/drone-rx/commit/07db98223a815bebfe512de210820b90e1c0ea8e))

## [1.9.1](https://github.com/jmboby/drone-rx/compare/v1.9.0...v1.9.1) (2026-04-13)


### Bug Fixes

* use exec collector for health check (kubectl runs client-side) ([5b294b0](https://github.com/jmboby/drone-rx/commit/5b294b0dc34472eeedf439ca49eef6f31ac948ad))

## [1.9.0](https://github.com/jmboby/drone-rx/compare/v1.8.0...v1.9.0) (2026-04-13)


### Features

* add /admin page with support bundle generation button ([036ec54](https://github.com/jmboby/drone-rx/commit/036ec543f0996916410313db17780e75e34f569c))
* add admin handler for support bundle generation ([a973f25](https://github.com/jmboby/drone-rx/commit/a973f25f1c1a6e852bae16b27a9cfeb9e22794b9))
* add RBAC for support bundle collection and POD_NAMESPACE env ([adb651d](https://github.com/jmboby/drone-rx/commit/adb651d866457305957f5af24a0f57c911676af6))
* add support-bundle CLI to API container image ([6000896](https://github.com/jmboby/drone-rx/commit/6000896ad2a4b914725ee2930062388b8afd3f6b))
* register POST /api/admin/support-bundle route ([085b54c](https://github.com/jmboby/drone-rx/commit/085b54ccd66fe7cc633d1b17cde320b2fa8b9c9d))
* Tier 3 Phase 2 — /admin page with support bundle generation ([96adb0a](https://github.com/jmboby/drone-rx/commit/96adb0ac10f2d5f81521d6275d82250ed9df0f86))


### Bug Fixes

* use correct troubleshoot release version (v0.126.1) ([72ab013](https://github.com/jmboby/drone-rx/commit/72ab013f122fb0532cec4e1cfb0a3de8edfc3e18))

## [1.8.0](https://github.com/jmboby/drone-rx/compare/v1.7.0...v1.8.0) (2026-04-12)


### Features

* add default storage class check to preflight ([df42dd2](https://github.com/jmboby/drone-rx/commit/df42dd2cfe984406ea7268e429caed398625ad96))
* wrap preflight/supportbundle specs in Secrets for Helm installs ([4f1e406](https://github.com/jmboby/drone-rx/commit/4f1e4062afad4056ddf3c32508c4c97371b39f7f))


### Bug Fixes

* correct health endpoint file path and tighten failure regex ([0ff7a79](https://github.com/jmboby/drone-rx/commit/0ff7a79bc79e6161b2316fe0b72b235dcdc8b73e))
* correct preflight run collector output paths to .log format ([0d4c508](https://github.com/jmboby/drone-rx/commit/0d4c5084792f2c0660706b539cb1696d59f7bb7b))
* correct run collector output paths and omit empty collectors ([a0f575d](https://github.com/jmboby/drone-rx/commit/a0f575dc0791f46140cdc8dfe79683eba5f2085f))
* simplify to Secret-only delivery, remove CRD-gated path ([830593d](https://github.com/jmboby/drone-rx/commit/830593d83f9da8120f81fef701a7d07afee60a57))
* Tier 3 Phase 1 — Secret wrappers, file paths, and collector fixes ([4bca2c6](https://github.com/jmboby/drone-rx/commit/4bca2c67f73b9ceface82f11f7954a8e7146c688))
* use correct Secret data keys for troubleshoot spec discovery ([da543cf](https://github.com/jmboby/drone-rx/commit/da543cf15b0394e2a609ccb1c6d4b6af596425db))

## [1.7.0](https://github.com/jmboby/drone-rx/compare/v1.6.2...v1.7.0) (2026-04-12)


### Features

* Tier 3 Phase 1 — preflight checks and support bundle specs ([f0ba4ec](https://github.com/jmboby/drone-rx/commit/f0ba4ec46cb44062304e9e4d02d2e2929cc29db5))


### Bug Fixes

* gate preflight/supportbundle on Troubleshoot CRD availability ([4301c07](https://github.com/jmboby/drone-rx/commit/4301c0766ad2225ada5c8b3ae070490a045fc8c6))

## [1.6.2](https://github.com/jmboby/drone-rx/compare/v1.6.1...v1.6.2) (2026-04-12)


### Bug Fixes

* opt into Node.js 24 for release-please action ([d1d052c](https://github.com/jmboby/drone-rx/commit/d1d052c33ee67f2a8dd112a343e150108249397c))
* opt into Node.js 24 for release-please action (Node 20 deprecated) ([a30e925](https://github.com/jmboby/drone-rx/commit/a30e925010b80e410d92fa13728a779a28ad001c))

## [1.6.1](https://github.com/jmboby/drone-rx/compare/v1.6.0...v1.6.1) (2026-04-12)


### Bug Fixes

* add sslmode to externalDatabase values, clarify toggle comments ([9408bd4](https://github.com/jmboby/drone-rx/commit/9408bd484e79dda58e9b7635f8d17a274ce0d771))
* collapse cloudnativepg.enabled into postgresql.enabled ([fad04ff](https://github.com/jmboby/drone-rx/commit/fad04fff83fd6f8b4f538f99651329927ef2fd3e))
* consolidate DB toggles into single postgresql.enabled ([a690f1f](https://github.com/jmboby/drone-rx/commit/a690f1f32f384b39c9b3fe4abbb187235c89cb7d))
* update schema — add sslmode enum, clarify DB toggle descriptions ([1d87b56](https://github.com/jmboby/drone-rx/commit/1d87b5603f431749ec929ebf871c94d70bb448c1))

## [1.6.0](https://github.com/jmboby/drone-rx/compare/v1.5.0...v1.6.0) (2026-04-12)


### Features

* add structured JSON logging to frontend ([856e0c7](https://github.com/jmboby/drone-rx/commit/856e0c78d1a5dc27e5729ebaabf8437035b1ef49))
* add structured JSON logging to frontend — page requests and API proxy ([49b865b](https://github.com/jmboby/drone-rx/commit/49b865bdd1a508b9d5282a8266dc60970ed4d405))

## [1.5.0](https://github.com/jmboby/drone-rx/compare/v1.4.0...v1.5.0) (2026-04-11)


### Features

* add structured JSON logging with slog across all components ([f81dbd2](https://github.com/jmboby/drone-rx/commit/f81dbd20a38593a9f3c28f118b4575a2950db208))
* add structured JSON logging with slog across all components ([857257d](https://github.com/jmboby/drone-rx/commit/857257d0d95fd79750f032a2e412e0fd9162d611))

## [1.4.0](https://github.com/jmboby/drone-rx/compare/v1.3.0...v1.4.0) (2026-04-11)


### Features

* polish premium tracking animations — accessibility, browser compat, visual refinement ([8b31ddc](https://github.com/jmboby/drone-rx/commit/8b31ddcbf50a46b78f5468700d2a97da47389624))
* replace full-screen confetti with localized sparkle burst from delivered icon ([264bd5d](https://github.com/jmboby/drone-rx/commit/264bd5d28f29ec07b104356704e501008daecc6b))


### Bug Fixes

* drone flight path now animates correctly using initial total ETA ([8eba7e0](https://github.com/jmboby/drone-rx/commit/8eba7e0861ca8239fe87e1e42132c6355de83c2d))
* faster ticker, animated flight path, polished premium tracking ([b3183ea](https://github.com/jmboby/drone-rx/commit/b3183ea237c7ed6621522d09a3206e9498c93d2c))
* reduce ticker interval from 10s to 5s for faster order fulfilment ([e7c0804](https://github.com/jmboby/drone-rx/commit/e7c0804843173c2c168ce6826a7819eb6f27328d))

## [1.3.0](https://github.com/jmboby/drone-rx/compare/v1.2.3...v1.3.0) (2026-04-11)


### Features

* add premium live tracking animations — flight path, countdown, transitions, confetti ([45ac5d4](https://github.com/jmboby/drone-rx/commit/45ac5d43fe40f9d98e5023d49bd42c00fc66a3c8))
* replace license expiry banner with full-screen blocking modal overlay ([07590b3](https://github.com/jmboby/drone-rx/commit/07590b3536530e7cd5f272bca2bf8803e70aaebe))


### Bug Fixes

* DB persistence + premium tracking animations + license expiry modal ([6804006](https://github.com/jmboby/drone-rx/commit/6804006b5d77244dc5a44ba2dcb48e83fb8459e6))
* remove post-upgrade hook from CNPG Cluster CR to preserve data across upgrades ([a999ca9](https://github.com/jmboby/drone-rx/commit/a999ca930f714daac10c8b0a66808421b05e81ea))

## [1.2.3](https://github.com/jmboby/drone-rx/compare/v1.2.2...v1.2.3) (2026-04-11)


### Bug Fixes

* add tag_name to gh-release action, remove push-only gate on attach-chart ([ad8c663](https://github.com/jmboby/drone-rx/commit/ad8c66301009959486ccebfe70c69a8bea0d3282))
* handle boolean license field values from SDK (was only matching strings) ([cc35390](https://github.com/jmboby/drone-rx/commit/cc35390cd21b2d5c9cc60bc45522e1397dad618d))
* parse SDK license info correctly — field casing, expiry from entitlements ([038cae9](https://github.com/jmboby/drone-rx/commit/038cae9483a4667683fadeeb75ee029389d989d9))
* SDK response parsing, GH release tag, license field types ([a622481](https://github.com/jmboby/drone-rx/commit/a6224816dd3f9ecd38ba824dd0b921ea2171e67f))

## [1.2.2](https://github.com/jmboby/drone-rx/compare/v1.2.1...v1.2.2) (2026-04-10)


### Bug Fixes

* use build output version for helm install in release workflow ([43c497c](https://github.com/jmboby/drone-rx/commit/43c497cf46d6e8c607cdb6055477f25454c10980))
* use version from build output for helm install, not Chart.yaml in repo ([436e81e](https://github.com/jmboby/drone-rx/commit/436e81e76f3d65fd5c8a467b1f93d87f03081ef0))

## [1.2.1](https://github.com/jmboby/drone-rx/compare/v1.2.0...v1.2.1) (2026-04-10)


### Bug Fixes

* verify release chain end-to-end ([217a2f3](https://github.com/jmboby/drone-rx/commit/217a2f3b6677b3e4f15f17edf181d014e3fc798c))
* verify release chain end-to-end ([6529f8f](https://github.com/jmboby/drone-rx/commit/6529f8f4b4a7f14f4b0b6d0983cbb4a60a143865))

## [1.2.0](https://github.com/jmboby/drone-rx/compare/v1.1.0...v1.2.0) (2026-04-10)


### Features

* chain release-please → Replicated Release via workflow_call ([dd2d4bd](https://github.com/jmboby/drone-rx/commit/dd2d4bd73be0cd77a6be2efe42208804eb45a228))
* chain release-please to Replicated Release via workflow_call ([5ba6529](https://github.com/jmboby/drone-rx/commit/5ba65292e75219f50a0c8a194aa64ae59abb3512))

## [1.1.0](https://github.com/jmboby/drone-rx/compare/v1.0.1...v1.1.0) (2026-04-10)


### Features

* add release-please version annotations to Chart.yaml and values.yaml ([f7f2db5](https://github.com/jmboby/drone-rx/commit/f7f2db565a275be5cd13697eb22b10beb9ca9186))


### Bug Fixes

* add --email flag to customer create (required for helm-install customers) ([efab481](https://github.com/jmboby/drone-rx/commit/efab48118947fb0b546e43d43cbf27300c1caf35))
* match x-release-please-version annotation in sed tag replacement ([2a7ed28](https://github.com/jmboby/drone-rx/commit/2a7ed281c7d6bfaf8ff377ad21c7887c126b68fc))
* remove --auto flag from CLI (uses .replicated config), fix channel cleanup to use ID ([f1cf5a4](https://github.com/jmboby/drone-rx/commit/f1cf5a4890bdd7de10afd239bcf8bd3cd630b585))
* remove --wait from helm install to prevent deadlock with post-install hooks ([53f1e39](https://github.com/jmboby/drone-rx/commit/53f1e395097c46f6ab8c96f0128d6ea7c9a758db))
* use --output-path for kubeconfig instead of stdout redirect ([82ffa42](https://github.com/jmboby/drone-rx/commit/82ffa42383f6ca370bf8f2f3d6ba1f880d0a5899))
* use chart version from Chart.yaml for helm install (not release label) ([840874b](https://github.com/jmboby/drone-rx/commit/840874bf9d1f5a89737dfd290028cd8d39d44726))
* use positional args for CLI commands (--id is deprecated) ([f3bc34e](https://github.com/jmboby/drone-rx/commit/f3bc34eb653994bea0c780a16039e96aa3618c7f))

## [1.0.1](https://github.com/jmboby/drone-rx/compare/v1.0.0...v1.0.1) (2026-04-09)


### Bug Fixes

* **images:** Fix Replicated SDK image location ([62904b6](https://github.com/jmboby/drone-rx/commit/62904b61843f3a086deee8d903bc1c5c5a948d6b))
* improve custom metrics — add logging, delivery stats, immediate send on startup ([a1c184b](https://github.com/jmboby/drone-rx/commit/a1c184b0eef486c1ded418b766fd4180fe211f2c))
* improve custom metrics with delivery stats and error logging ([0d9f2c9](https://github.com/jmboby/drone-rx/commit/0d9f2c92af11e60ac85876435a0b96a30785b465))
* proxy all remaining images through custom domain ([419a966](https://github.com/jmboby/drone-rx/commit/419a9662d1f8c26e5397d62af24b889d7d0c8e2b))
* proxy all remaining images through custom domain ([f9f87e6](https://github.com/jmboby/drone-rx/commit/f9f87e6d99903eae65c2c503532aafdd292c9f18))
* send custom metrics every 30s instead of 5 minutes ([ed0387c](https://github.com/jmboby/drone-rx/commit/ed0387ce25631e97ed5f7eb52ebf1a714a1e1673))
* use per-image registry overrides for NATS instead of global ([9d83809](https://github.com/jmboby/drone-rx/commit/9d8380923f2e2eb1ea18a58707a1e19f3c03c7ef))

## 1.0.0 (2026-04-08)


### Features

* add .replicated config file, KOTS manifests, and CLI-based release creation ([fece5c4](https://github.com/jmboby/drone-rx/commit/fece5c453a25672c90251135210b19199d95c999))
* add amd64 build, push, and helm-package targets to Makefile ([cd7b3ff](https://github.com/jmboby/drone-rx/commit/cd7b3ff61657c20498b1b9e3f7a12920355eab51))
* add background metrics sender with order count by status ([ea08d55](https://github.com/jmboby/drone-rx/commit/ea08d55979c7e54a5d17a14799b9821d12ba54aa))
* add config package with env var parsing ([297d3fc](https://github.com/jmboby/drone-rx/commit/297d3fcf14ac752380b69cf25b72220262ddaf05))
* add database connection package with pgx pool ([476487c](https://github.com/jmboby/drone-rx/commit/476487c353f841a8fa2885b7bdf5e59e929c84e2))
* add database migrations with medicines seed data and orders schema ([3af617d](https://github.com/jmboby/drone-rx/commit/3af617d3d634f7c8867e8df2b83b8d7ef349eed2))
* add ETA calculation for order delivery estimation ([5949b70](https://github.com/jmboby/drone-rx/commit/5949b705517eb6aaedb4c62f58587d1cf30d4205))
* add frontend types, API client, and cart store ([cf3643f](https://github.com/jmboby/drone-rx/commit/cf3643ffbc7d50ca31e49fceb9fccc26fec8857c))
* add health endpoint with DB and NATS connectivity checks ([2494081](https://github.com/jmboby/drone-rx/commit/24940810f751b4421195bbe0343760c67214543f))
* add Helm chart foundation with CloudNativePG and NATS subcharts ([767741d](https://github.com/jmboby/drone-rx/commit/767741d7ef84908cb9c814122c38cc9ae057b22e))
* add Helm templates for API and frontend deployments, services, and PostgreSQL cluster ([907eab9](https://github.com/jmboby/drone-rx/commit/907eab9daaca6100e37789c8cc77bc3be21d7eb3))
* add imagePullSecrets support to API and frontend deployments ([84acfc4](https://github.com/jmboby/drone-rx/commit/84acfc4feff853f1708f6f28797190ba0a4bf9c6))
* add ingress with 3 TLS modes (auto, manual, self-signed) ([da9a9e4](https://github.com/jmboby/drone-rx/commit/da9a9e4feea10d48bfe6467e56e76334e1c7af47))
* add license expiry warning banner and premium tracking badge ([49f61c2](https://github.com/jmboby/drone-rx/commit/49f61c2627f5364604a51191e6aa5a1be96c158d))
* add license status and updates check API handlers ([8e7026b](https://github.com/jmboby/drone-rx/commit/8e7026bcd7271b4538a8849ac10dca2073065263))
* add Makefile with build, lint, test, and clean targets ([6bb3fd7](https://github.com/jmboby/drone-rx/commit/6bb3fd7cc0aa4a83df982fad5a2c7e2d0c2511ad))
* add manual promote workflow for Beta/Stable channel promotion ([096e2af](https://github.com/jmboby/drone-rx/commit/096e2afd15fe11d98949db6d87d2412254bc814d))
* add medicine browsing page with category filtering and cart ([2fa3a80](https://github.com/jmboby/drone-rx/commit/2fa3a80917305ad2dd59cf0bde507a4296449238))
* add medicine list and get-by-id HTTP handlers ([408e9c7](https://github.com/jmboby/drone-rx/commit/408e9c7d8646c0e31dd513b1a715dba1575784b5))
* add medicine model with list and get-by-id queries ([7b507f0](https://github.com/jmboby/drone-rx/commit/7b507f0b44ba01422dcf4698fce43f9519999cb4))
* add multi-stage Dockerfiles for API and frontend ([e1238b4](https://github.com/jmboby/drone-rx/commit/e1238b46fc3d67d49b29cef5bf25844e5f2179f4))
* add NATS event publisher for order status updates ([143665a](https://github.com/jmboby/drone-rx/commit/143665af68bd3e2eb3b66d300103f15ae6957634))
* add order create, get, and list HTTP handlers ([b169cc7](https://github.com/jmboby/drone-rx/commit/b169cc7d45e45ac1cbe610b567330b4fedabfe2b))
* add order form page with cart summary and delivery details ([f3a83be](https://github.com/jmboby/drone-rx/commit/f3a83be781a103d8c70d09beba51a1d5cef1b946))
* add order history page with name-based search ([6337d21](https://github.com/jmboby/drone-rx/commit/6337d214db4db150cf9a9f54174d854ab7b04ee3))
* add order model with status progression, validation, and DB queries ([cd98bf4](https://github.com/jmboby/drone-rx/commit/cd98bf4698b9f2187e4433d12bb9a94e04ae17c7))
* add order status page with live tracking and status stepper ([91fa5fc](https://github.com/jmboby/drone-rx/commit/91fa5fcff3cbc7427051d5d8a16980cd948bb3f3))
* add PR workflow with lint, test, GHCR push, and CMX smoke test ([6203d57](https://github.com/jmboby/drone-rx/commit/6203d57a58f24e95b3ddec6d373ef3a585b53278))
* add preflight and support bundle placeholder templates ([073867c](https://github.com/jmboby/drone-rx/commit/073867c2a9d0a7e61cdee9f7afdc62d2325a586b))
* add release workflow — build, push, create release, promote to Unstable, test on CMX ([7319944](https://github.com/jmboby/drone-rx/commit/7319944d8684fd01634fc225dbfb0cd9362cecaa))
* add release-please for automated semver from conventional commits ([7c9810b](https://github.com/jmboby/drone-rx/commit/7c9810bda1a644e2736ad916add978f3f375eb0f))
* add Replicated SDK client with license, metrics, and updates API ([68e9adf](https://github.com/jmboby/drone-rx/commit/68e9adf4d79e36335d0460d1ed643be7687ed1f5))
* add Replicated SDK subchart and proxy registry image pull ([f3dbdde](https://github.com/jmboby/drone-rx/commit/f3dbddeca10271fb4f50e2fc4106bfd360081be3))
* add state machine ticker that auto-advances order statuses ([4775c56](https://github.com/jmboby/drone-rx/commit/4775c563b018907be1afe0a666a10c726cc1884b))
* add update banner, license status check, and tracking license gate in frontend ([d56606c](https://github.com/jmboby/drone-rx/commit/d56606c3d65447001f34ef6e1e0a931683192a04))
* add values.schema.json for Helm values validation ([49edc94](https://github.com/jmboby/drone-rx/commit/49edc94c3055bc82fd50839ad0ff59ff06a4a238))
* add webhook notifier for delivery events ([c9d03d4](https://github.com/jmboby/drone-rx/commit/c9d03d4a0bad61a9babdeb68816cf9151c6e4df2))
* add WebSocket tracking handler with NATS subscription relay ([f6a0e3e](https://github.com/jmboby/drone-rx/commit/f6a0e3ef392d21e633087fe34ce8a8e759c4c026))
* gate WebSocket live tracking by license entitlement via SDK ([9f2ac9e](https://github.com/jmboby/drone-rx/commit/9f2ac9e8dd050b60b8bb484bc5e845c5d884668e))
* initialize Go project with minimal health endpoint ([197d79a](https://github.com/jmboby/drone-rx/commit/197d79a400660658aa9382ee2e80706a8b75441d))
* proxy all images through custom domain images.littleroom.co.nz ([0a81a6d](https://github.com/jmboby/drone-rx/commit/0a81a6dd229ab14ee22e232f24870bf3b942d124))
* redesign frontend with Aerial Pharmacy aesthetic and custom drone icon ([dd91fd9](https://github.com/jmboby/drone-rx/commit/dd91fd9deb62a90121c75e14ee9621697d5e0d5a))
* scaffold SvelteKit frontend with TypeScript, Tailwind, and adapter-node ([5c9dfb3](https://github.com/jmboby/drone-rx/commit/5c9dfb3b9e5191a68035e197f260f1cb4eaf0857))
* set ghcr-creds as default imagePullSecrets in values.yaml ([7dd60ca](https://github.com/jmboby/drone-rx/commit/7dd60ca67a8c336c933dae2c4b97e98299ebcb76))
* use Replicated proxy registry for image pulls in CI workflows ([b863b20](https://github.com/jmboby/drone-rx/commit/b863b2067931928f93d09411b5e891d1197ec37d))
* wire SDK client, metrics sender, license and updates routes into main.go ([cbdb81c](https://github.com/jmboby/drone-rx/commit/cbdb81c1c18ae17adc771168acf10db940362ccf))
* wire up all components in main.go with graceful shutdown ([9d8a267](https://github.com/jmboby/drone-rx/commit/9d8a267182633b99c96587acff947dd684897bc4))


### Bug Fixes

* add aria-labels to icon-only back navigation links ([276a588](https://github.com/jmboby/drone-rx/commit/276a58859a3cd5366150573aa71c8e6feebc3b0f))
* add CNPG operator wait job to resolve webhook timing issue ([9a9f29a](https://github.com/jmboby/drone-rx/commit/9a9f29a0f20fc761107b590b133e9c8bea3d78f6))
* add helm repo add before dependency build in CI workflows ([de70a74](https://github.com/jmboby/drone-rx/commit/de70a741940fdf834ff2d46f697613bd9834b033))
* add imagePullSecrets config for CNPG and NATS subcharts, inject via HelmChart CR ([10828f9](https://github.com/jmboby/drone-rx/commit/10828f9bbc8ea2704c65f559a99c22793def42a7))
* add imagePullSecrets to hook jobs for proxied image pulls ([d00b29f](https://github.com/jmboby/drone-rx/commit/d00b29f23b8371756d9b6e659c1621d5b49eee96))
* align StatusInFlight constant with Postgres enum value ([9c37ea7](https://github.com/jmboby/drone-rx/commit/9c37ea7fe0413ed504d9ae45cd44f19fef7ff75c))
* bake PR image tags into values.yaml before packaging chart ([d43a1d3](https://github.com/jmboby/drone-rx/commit/d43a1d30a31e38d36c051038e85cba808ede0b49))
* CNPG credentials, amd64 builds, ticker interval, design tweaks ([46225f9](https://github.com/jmboby/drone-rx/commit/46225f9ae6d9a7fad03ed9241fc5d545ce436ee8))
* correct proxy image path format — remove docker.io prefix ([c3c470b](https://github.com/jmboby/drone-rx/commit/c3c470b54fe135b75db334d0958e20d98c5e5571))
* disable KOTS install on test customers (Helm-only release) ([ce09f7d](https://github.com/jmboby/drone-rx/commit/ce09f7dd918867de943562205cc7d0ed18c08be7))
* disable preflight checks in CI (placeholder spec causes JSON parse error) ([611fcf2](https://github.com/jmboby/drone-rx/commit/611fcf2209f9819179f1a4cbe7249d90dbf8045c))
* match promote version label to chart version (0.1.0) ([c2b30ae](https://github.com/jmboby/drone-rx/commit/c2b30ae2a0786a034b38df0c4f43073391cc1d0f))
* override imagePullSecrets to empty in CI (no ghcr-creds in CMX clusters) ([dbe753e](https://github.com/jmboby/drone-rx/commit/dbe753e48328df22c316d1f8a0f21d9e9dd43f10))
* simplify CNPG wait job — use busybox nc instead of kubectl, no RBAC needed ([3cc420a](https://github.com/jmboby/drone-rx/commit/3cc420ad748d7d95a0530768c5f0c1cb5353e198))
* update k3s version from 1.29 to 1.34 (1.29 no longer supported) ([aa48654](https://github.com/jmboby/drone-rx/commit/aa486549cef2ed8575d0ab17f85972c77ef9a00b))
* use /anonymous/ path for public images through custom proxy domain ([5aac947](https://github.com/jmboby/drone-rx/commit/5aac947e52decc456cbc560a6ff959075156a9c0))
* use chart native version (0.1.0) instead of dynamic version in prepare-cluster ([c12d257](https://github.com/jmboby/drone-rx/commit/c12d257c3ae7617af6ad956ffe8d363f5dc6a306))
* use label selectors instead of hardcoded deployment names in smoke tests ([48935fc](https://github.com/jmboby/drone-rx/commit/48935fc48bf6d4295563a47764b429a1e1e05b9a))
* use lowercase channel slugs in promote workflow ([07725e4](https://github.com/jmboby/drone-rx/commit/07725e4510e857102376c1489d78c17fd5cbc55a))
