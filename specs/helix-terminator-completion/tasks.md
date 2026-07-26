# Tasks: helix_terminator Completion Initiative

**Input**: Design documents from `/specs/helix-terminator-completion/`

**Prerequisites**: plan.md (required), spec.md (required)

**Tests**: Test tasks are included because the completion initiative is coverage-driven and constitution-mandated.

**Organization**: Tasks are grouped by implementation phase to enable independent execution streams. Each phase is independently testable.

## Format: `[ID] [P?] [Phase] Description`

- **[P]**: Can run in parallel (different files / services, no dependencies)
- Include exact file paths in descriptions

---

## Phase 1: Setup & Ground Truth

**Purpose**: Establish verified baseline and resolve unknowns before touching production code.

- [ ] T001 Run constitution inheritance gate `tests/verify_constitution_inheritance.sh` and record evidence in `specs/helix-terminator-completion/research.md`
- [ ] T002 Run mutation proof `scripts/testing/meta_test_false_positive_proof.sh` and record evidence in `specs/helix-terminator-completion/research.md`
- [ ] T003 [P] Run `go vet ./...` and `go test ./...` for all 25 services; capture per-service pass/fail in `specs/helix-terminator-completion/research.md`
- [ ] T004 [P] Inventory exact counts of stub tests, integration tests, stress/chaos tests, proto files, and migrations; record in `specs/helix-terminator-completion/research.md`
- [ ] T005 Verify availability of Go, podman, buf, protoc, k6, helm, terraform, kustomize, Flutter SDK; document containerized fallback or SKIP in `specs/helix-terminator-completion/research.md`
- [ ] T006 [P] Reconcile gateway `notImplemented` routes against MVP scope; document in-scope / out-of-scope / operator-blocked decisions in `specs/helix-terminator-completion/research.md`
- [ ] T007 [P] Reconcile operator-blocked items (Stripe TEST keys, Slack token, SMTP relay, etc.) in `specs/helix-terminator-completion/research.md`

**Checkpoint**: `research.md` is complete and all unknowns are resolved or explicitly operator-blocked.

---

## Phase 2: Foundation — Submodules, Gateway, Service Stubs

**Purpose**: Remove foundational blockers and make the gateway route surface functional.

- [ ] T008 Fix `submodules/helixtrack-core/Website/Upstreamable` nested submodule URL or remove orphan; verify with `git submodule update --init --recursive`
- [ ] T009 Add `submodules/open-design` gitlink to parent index with `git add submodules/open-design`
- [ ] T010 Commit updated submodule pointers for `constitution`, `containers`, `helixllm`, `helixqa`
- [ ] T011 [P] Implement upstream `/me`, `/me/sessions`, `/me/preferences` endpoints in `services/user-service/` and wire into `services/gateway-service/internal/server/server.go`
- [ ] T012 [P] Implement vault items and share endpoints in `services/vault-service/` and wire into gateway
- [ ] T013 [P] Implement host connect endpoint in `services/host-service/` and wire into gateway
- [ ] T014 [P] Implement terminal WebSocket stream in `services/terminal-service/` and WebSocket proxy in gateway
- [ ] T015 [P] Implement session share endpoint in `services/collaboration-service/` and wire into gateway
- [ ] T016 [P] Implement SFTP per-session endpoints in `services/sftp-service/` and wire into gateway
- [ ] T017 [P] Implement tunnels endpoints in `services/port-forward-service/` and wire into gateway
- [ ] T018 [P] Implement snippet execute endpoint in `services/snippet-service/` and wire into gateway
- [ ] T019 [P] Implement recording playback/export in `services/recording-service/` and wire into gateway
- [ ] T020 [P] Implement analytics usage endpoint in `services/analytics-service/` and wire into gateway
- [ ] T021 [P] Implement AI autocomplete/explain in `services/ai-service/` and wire into gateway
- [ ] T022 [P] Implement billing subscription/usage in `services/billing-service/` and wire into gateway
- [ ] T023 [P] Implement system maintenance endpoint or document out-of-scope removal in `services/gateway-service/internal/server/server.go`
- [ ] T024 Wire Prometheus metrics in `services/gateway-service/internal/server/server.go:1006-1008`
- [ ] T025 Implement SSO redirect and callback in `services/gateway-service/internal/server/server.go:1023-1028` or mark operator-blocked

**Checkpoint**: All in-scope gateway routes return real responses; out-of-scope routes are removed with documented reasons.

---

## Phase 3: Persistence & Contracts

**Purpose**: Real schemas, real proto contracts, and accurate OpenAPI.

- [ ] T026 [P] Replace placeholder `migrations/001_init.up.sql` with complete schema per service, aligned with `docs/research/mvp/final/implementation/backend/<service>_db.sql` for all 25 services
- [ ] T027 [P] Add `migrations/001_init.down.sql` rollback scripts for all 25 services
- [ ] T028 Apply all migrations to PostgreSQL 17.2 in rootless podman; capture schema in `specs/helix-terminator-completion/research.md`
- [ ] T029 Encrypt `private_key` and `passphrase` storage in `services/keychain-service/` schema and repository
- [ ] T030 Add pre-store credential leak audit automation to CI or pre-commit
- [ ] T031 [P] Replace placeholder `// TODO: define service RPCs and messages` in all 25 `.proto` files with real RPCs and messages from `docs/research/mvp/final/implementation/04-api-specification/`
- [ ] T032 Run `buf lint` and `protoc` compile for all proto services; ensure no drift with Go structs
- [ ] T033 Update `docs/research/mvp/final/implementation/api/openapi.yaml` to cover WebSocket streams and gRPC gateway mappings
- [ ] T034 Document honestly any gRPC/WS endpoints remaining uncovered by OpenAPI

**Checkpoint**: All services have real schemas applied against PostgreSQL; proto and OpenAPI contracts are accurate.

---

## Phase 4: Security Hardening

**Purpose**: Close CRITICAL/HIGH security gaps and satisfy constitution security clauses.

- [ ] T035 [P] Populate `infrastructure/security/falco/README.md` with real Falco rules
- [ ] T036 [P] Populate `infrastructure/security/cosign/README.md` with signing/verification policy
- [ ] T037 [P] Populate `infrastructure/security/sealed-secrets/README.md` with SealedSecret templates
- [ ] T038 [P] Expand `infrastructure/security/trivy/README.md` with CI scanning steps
- [ ] T039 [P] Write `docs/runbooks/certificate-rotation.md` from `docs/research/mvp/final/implementation/05-security-zero-trust/README.md`
- [ ] T040 [P] Write `docs/runbooks/failover-procedure.md`
- [ ] T041 [P] Write `docs/runbooks/incident-response.md`
- [ ] T042 [P] Write `docs/runbooks/key-rotation.md`
- [ ] T043 [P] Write `docs/runbooks/postgres-pitr-restore.md`
- [ ] T044 Pin image tags + SHA-256 digests in `infrastructure/helm/helixterm/values.yaml`
- [ ] T045 Switch to `pullPolicy: Always` where digest-pinned in Helm values
- [ ] T046 Convert umbrella `infrastructure/helm/helixterm/Chart.yaml` into per-service subcharts
- [ ] T047 Set HA replica counts per environment in `values-staging.yaml` / `values-production.yaml`
- [ ] T048 Add fine-grained `NetworkPolicy` allow rules in `infrastructure/helm/helixterm/templates/`
- [ ] T049 Enforce Pod Security Standards (restricted) in Helm chart
- [ ] T050 [P] Wire shared JWT/SPIFFE/RBAC middleware into every service beyond gateway and auth
- [ ] T051 [P] Replace `log.Fatalf` with structured logger + graceful shutdown in `services/ai-service/`, `billing-service/`, `keychain-service/`, `user-service/`, `workspace-service/`
- [ ] T052 [P] Standardize all services on `/healthz/live` + `/healthz/ready` with real dependency checks

**Checkpoint**: Security runbooks and configs populated; Helm hardened; auth middleware fleet-wide; graceful startup and standard health endpoints.

---

## Phase 5: Test Coverage

**Purpose**: Eliminate stub tests and add integration/stress/chaos coverage fleet-wide.

- [ ] T053 [P] Replace `t.Skip` placeholders in `services/ai-service/cmd/ai-service/main_test.go`, `internal/model/model_test.go`, `internal/repository/repository_test.go`, `internal/server/server_test.go`
- [ ] T054 [P] Replace `t.Skip` placeholders in `services/billing-service/cmd/billing-service/main_test.go`, `internal/model/model_test.go`, `internal/repository/repository_test.go`, `internal/server/server_test.go`
- [ ] T055 [P] Replace `t.Skip` placeholders in `services/keychain-service/cmd/keychain-service/main_test.go`, `internal/model/model_test.go`, `internal/repository/repository_test.go`
- [ ] T056 [P] Replace `t.Skip` placeholders in `services/user-service/cmd/user-service/main_test.go`, `internal/model/model_test.go`, `internal/repository/repository_test.go`, `internal/server/server_test.go`
- [ ] T057 [P] Replace `t.Skip` placeholders in `services/gateway-service/cmd/gateway-service/main_test.go`, `internal/model/model_test.go`, `internal/repository/repository_test.go`
- [ ] T058 [P] Replace 2 tautology tests in `services/notification-service/cmd/notification-service/main_test.go` and `internal/model/model_test.go`
- [ ] T059 [P] Replace remaining `t.Skip` placeholders across all other services
- [ ] T060 [P] Add integration tests for analytics-service, audit-service, collaboration-service, config-service, health-service, helixtrack-bridge-service, host-service, org-service, recording-service, sftp-service, snippet-service, ssh-proxy-service, terminal-service, workspace-service, ai-service
- [ ] T061 [P] Add `handler_stress_test.go` (gated `//go:build stress`) to each of the 24 services without stress coverage
- [ ] T062 [P] Add `handler_chaos_test.go` (gated `//go:build chaos`) to each of the 24 services without chaos coverage
- [ ] T063 Fix `scripts/generate_test_banks.go` to generate real test bodies
- [ ] T064 Regenerate `test/banks/unit/` and `test/banks/integration/` with real tests or remove unrunnable banks

**Checkpoint**: Zero stub tests; integration tests in all 25 services; stress/chaos tests in all 25 services; real test banks.

---

## Phase 6: Infrastructure & DevOps

**Purpose**: Make Terraform, Helm, Kustomize, and CI/CD production-ready.

- [ ] T065 Validate + plan all Terraform modules and environments via container
- [ ] T066 Reconcile modified Terraform modules: eks, elasticache, iam, msk, rds
- [ ] T067 Run `helm lint` and `helm template` for `infrastructure/helm/helixterm/`
- [ ] T068 Ensure `values-staging.yaml`, `values-production.yaml`, and base `values.yaml` are consistent
- [ ] T069 Run `kustomize build` for base, dev, staging, production overlays
- [ ] T070 Boot PostgreSQL, Redis, and Kafka via rootless podman compose with healthchecks
- [ ] T071 Verify all services start against the local stack
- [ ] T072 Decide CI re-enable vs. stay disabled per §11.4.156 and document with evidence
- [ ] T073 Update `.github/workflows/main.yml` and `.github/workflows/nightly.yml` Go version to 1.25+
- [ ] T074 Fix or remove references to `test/banks/unit/...`, `test/banks/integration/...`, `test/e2e`, `test/integration` in CI workflows

**Checkpoint**: Validated Terraform/Helm/Kustomize; local stack boots; CI paths corrected and compliant.

---

## Phase 7: Flutter Client

**Purpose**: Replace stub tests, implement core terminal view, and wire design system.

- [ ] T075 Use rootless podman Flutter 3.24 image; run `flutter pub get` in `clients/flutter/`
- [ ] T076 Run `flutter analyze` in `clients/flutter/` and drive to 0 issues
- [ ] T077 Implement real `clients/flutter/lib/terminal_view.dart` and add terminal/SSH package to `pubspec.yaml`
- [ ] T078 [P] Replace ~30 placeholder stubs and ~110 demo wrappers in `clients/flutter/lib/` with real implementations or remove
- [ ] T079 [P] Add JSON serialization to the 4 models missing it in `clients/flutter/lib/`
- [ ] T080 Wire `clients/flutter/design_system/` tokens/theme into `lib/` with light+dark variants
- [ ] T081 [P] Replace 8 stub BLoC tests in `clients/flutter/test/` with real tests
- [ ] T082 Generate host-rendered golden images for every screen×state×{light,dark} in `clients/flutter/test/`
- [ ] T083 Dual-validate golden images with diff + OCR/vision oracle per §11.4.170
- [ ] T084 Replace boot-only `clients/flutter/integration_test/app_test.dart` with real end-to-end user journey tests

**Checkpoint**: Real terminal view, zero prod TODOs, design system wired, all bloc/integration tests real, visual proofs captured.

---

## Phase 8: Observability

**Purpose**: Wire real metrics, traces, logs, and alerts.

- [ ] T085 [P] Instrument all 25 services with Prometheus RED metrics
- [ ] T086 Verify Prometheus scrape endpoints and target health locally
- [ ] T087 Propagate `X-Request-ID`/trace context across services and verify spans in local Jaeger
- [ ] T088 Ensure JSON structured logs with PII redaction in all 25 services
- [ ] T089 Verify Loki log ingestion locally
- [ ] T090 Add alert rules for error-rate, latency, and service-down; fire-test locally

**Checkpoint**: Live metrics, captured traces, structured logs, alert rules proven.

---

## Phase 9: Documentation, Gap Closure & Coverage Ledger

**Purpose**: Close documentation gaps, update ledgers, and sync exports.

- [ ] T091 Update `docs/research/mvp/final/implementation/GAP_REGISTER.md` with evidence per GAP-01..GAP-15
- [ ] T092 Mark GAPs closed, honestly deferred, or operator-blocked
- [ ] T093 Refresh `docs/review/SECURITY_REVIEW.md`
- [ ] T094 Refresh `docs/review/CODE_SCAFFOLD_REVIEW.md`
- [ ] T095 Refresh `docs/review/OPENDESIGN_COMPLIANCE_REVIEW.md`
- [ ] T096 Regenerate `docs/COVERAGE_LEDGER.md` after Phase 5
- [ ] T097 Create `docs/qa/<feature>/` for every shipped feature with end-to-end transcripts per §11.4.83
- [ ] T098 Update `docs/CONTINUATION.md` to revision 19+; reconcile HEL-012, HEL-013, and parked items
- [ ] T099 Fix version-string drift (postgres/go/kafka/redis) across docs
- [ ] T100 Regenerate md/html/pdf/docx exports in sync
- [ ] T101 Render 30 mermaid + 8 drawio diagrams where renderers exist

**Checkpoint**: GAP_REGISTER closed, reviews refreshed, coverage ledger current, QA transcripts present, continuation current, synced exports.

---

## Phase 10: Release Preparation

**Purpose**: Verify everything together and publish a release.

- [ ] T102 Run full-suite retest from clean baseline: `bash tests/verify_constitution_inheritance.sh`, `bash scripts/testing/meta_test_false_positive_proof.sh`, `bash tests/docs_consistency_gate.sh`, `make test`
- [ ] T103 Run integration, e2e, stress, and chaos suites
- [ ] T104 Verify Helm/Terraform production-readiness
- [ ] T105 Confirm CI is compliant/disabled per §11.4.156
- [ ] T106 Create prefixed release tag per §11.4.151
- [ ] T107 Push tag to all upstreams and mirror on every owned submodule
- [ ] T108 Merge-onto-latest-main, no force-push
- [ ] T109 Update `docs/CONTINUATION.md`, `CHANGELOG.md`, and memory files

**Checkpoint**: All gates green, release tag published, submodules tagged, continuation/changelog updated.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies; produces `research.md`
- **Phase 2 (Foundation)**: Depends on Phase 1; unblocks gateway/service work
- **Phase 3 (Persistence & Contracts)**: Depends on Phase 2; needs service structure in place
- **Phase 4 (Security Hardening)**: Depends on Phase 3; needs real endpoints to secure
- **Phase 5 (Test Coverage)**: Depends on Phase 3 and Phase 4; needs real code and security in place
- **Phase 6 (Infrastructure & DevOps)**: Depends on Phase 3; can proceed in parallel with Phase 4-5
- **Phase 7 (Flutter Client)**: Depends on Phase 1; can proceed in parallel after research
- **Phase 8 (Observability)**: Depends on Phase 2 and Phase 6; needs services and infra stack
- **Phase 9 (Docs & Gap Closure)**: Depends on Phase 5 and Phase 7
- **Phase 10 (Release)**: Depends on all prior phases

### Parallel Opportunities

- All Phase 1 tasks marked [P] can run in parallel
- Phase 2 gateway/service endpoint tasks are mostly parallel per service
- Phase 3 migration/proto tasks are parallel per service
- Phase 4 security tasks are parallel across runbooks, Helm, middleware, startup, health
- Phase 5 test coverage tasks are parallel per service
- Phase 6 infra tasks are parallel across Terraform, Helm, Kustomize, local stack, CI
- Phase 7 Flutter tasks are parallel across terminal view, TODO burn-down, design system, tests
- Phase 8 observability tasks are parallel per service
- Phase 9 doc tasks are parallel per review/ledger/QA export

---

## Suggested Subagent Streams

Given the breadth, fan out into **5 concurrent streams** after Phase 1:

1. **Backend Completion Stream**: Phases 2, 3, 4 (gateway, services, schemas, protos, security)
2. **Backend Testing Stream**: Phase 5 (stub conversion, integration, stress, chaos, test banks)
3. **Infrastructure Stream**: Phase 6 (Terraform, Helm, Kustomize, local stack, CI)
4. **Flutter Client Stream**: Phase 7 (terminal view, BLoC tests, visual proofs, e2e)
5. **Docs & Closure Stream**: Phases 8, 9, 10 (observability wiring, gap register, coverage ledger, QA transcripts, release)

Each stream should own its ledger file under `specs/helix-terminator-completion/` and update `docs/CONTINUATION.md` after every session.
