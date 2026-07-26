# Implementation Plan: helix_terminator Completion Initiative

**Branch**: `helix-terminator-completion` | **Date**: 2026-07-26 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/helix-terminator-completion/spec.md`

## Summary

Bring `helix_terminator` from its current scaffolded state to a functionally complete, test-covered, security-hardened, and release-ready state. The work is organized into ten phases, sequenced to resolve foundational blockers first (submodules, gateway/service stubs, DB schemas), then harden security and tests, then infrastructure, client, observability, documentation, and release.

## Technical Context

**Language/Version**: Go 1.25.0 / Go workspaces 1.26.4; Flutter 3.24 (client); PostgreSQL 17.2; Protocol Buffers (proto3).

**Primary Dependencies**: Gorilla/mux, pgx/v5, jwt-go, gRPC, Prometheus, Jaeger, Loki, k6, Trivy, buf, protoc, Helm, Terraform, Kustomize, rootless Podman.

**Storage**: Per-service PostgreSQL schemas, Redis, Kafka/MSK.

**Testing**: Go test (unit, integration, stress, chaos), contract tests, test banks (`test/banks`), k6, Flutter tests (widget/bloc/integration), real DB in rootless podman.

**Target Platform**: Linux containers (Kubernetes), multi-platform Flutter client (iOS/Android/Web/Desktop per OpenDesign tokens).

**Project Type**: Multi-service Go backend + Flutter frontend + Helm/Terraform infrastructure.

**Performance Goals**: p99 < 200ms for API endpoints; SLO dashboard; capacity model.

**Constraints**: §11.4 anti-bluff, no fakes beyond unit tests, 4-layer coverage, real evidence per closure, constitution gates before commit, CI disabled per §11.4.156 unless operator re-enables with evidence.

**Scale/Scope**: 25 Go services, 24 proto services, ~38 DB tables (current) vs. ~121 claimed (verify), Flutter client with 278 widgets / 19 BLoCs.

## Constitution Check

- §1 / §1.1 — False-positive immunity: every gate must have a paired mutation proof; service-level gates need their own mutation proofs.
- §11.4 — Anti-bluff / captured evidence: each task must close with real command output, real test run, real DB delta, or rendered output.
- §11.4.25 — Full automation coverage: six-invariant coverage per feature; close 43 stub tests, add stress/chaos fleet-wide, add integration tests to 15 services.
- §11.4.27 — No fakes beyond unit tests: integration/e2e use real DB/Kafka/Redis; test banks must be real or removed, not silenced.
- §11.4.30 — No versioned build artifacts / `.gitignore`: pin image tags with SHA-256 digests in Helm; no build artifacts committed.
- §11.4.40 — Full-suite retest before release tag: release preparation includes clean-baseline retest.
- §11.4.52 — Autonomous validation: stress/chaos tests per service; evidence captured automatically.
- §11.4.83 — QA transcripts per feature: create `docs/qa/<feature>/` for every shipped feature.
- §12.10 — Continuation currency: update `docs/CONTINUATION.md` after every session/phase.

## Project Structure

### Documentation (this feature)

```text
specs/helix-terminator-completion/
├── spec.md              # Feature specification
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
└── tasks.md             # Phase 2 output (tasks generation)
```

### Source Code (repository root)

```text
services/                # 25 Go microservices
├── gateway-service/
├── auth-service/
├── user-service/
├── vault-service/
├── host-service/
├── ssh-proxy-service/
├── terminal-service/
├── keychain-service/
├── workspace-service/
├── config-service/
├── pki-service/
├── billing-service/
├── org-service/
├── notification-service/
├── audit-service/
├── health-service/
├── ai-service/
├── collaboration-service/
├── container-bridge-service/
├── helixtrack-bridge-service/
├── port-forward-service/
├── sftp-service/
├── recording-service/
├── snippet-service/
└── analytics-service/

clients/flutter/         # Flutter client
infrastructure/          # Helm, Terraform, Kustomize, security configs
docs/                    # All docs, runbooks, QA transcripts, coverage ledger
tests/                   # Integration, e2e, contract, test banks
submodules/              # constitution, open-design, helixtrack-core, etc.
```

**Structure Decision**: The existing monorepo structure is preserved. The plan operates across all 25 backend services, the Flutter client, and the infrastructure directory.

## Complexity Tracking

No complexity violations. The existing multi-service layout is required by the architecture and is already in place.

---

## Phases, Tasks, and Subtasks

### Phase 0 — Research & Ground Truth

**Objective**: Resolve all unknowns, validate the current baseline, and produce `research.md`.

#### T0.1 Baseline verification

- Run `bash tests/verify_constitution_inheritance.sh` and `bash scripts/testing/meta_test_false_positive_proof.sh`; capture exit code and output.
- Run `go vet ./...` and `go test ./...` for every service; record which services pass/fail.
- Run migration apply against real PostgreSQL 17.2 in rootless podman for all services; capture schema diff.
- Verify the exact count of: stub tests, integration tests, stress/chaos tests, proto files, migrations.

#### T0.2 Scope reconciliation

- Cross-reference the completion report with `docs/research/mvp/final/implementation/DEVELOPMENT_PLAN.md` and `GAP_REGISTER.md`.
- Decide which gateway `notImplemented` routes are in MVP scope vs. out of scope; document decisions with reason.
- Identify operator-blocked items (Stripe TEST keys, Slack token, SMTP relay, etc.) and mark them explicitly.

#### T0.3 Toolchain inventory

- Confirm availability of: Go, podman, buf, protoc, k6, helm, terraform, kustomize, Flutter SDK.
- For missing host tools, document containerized fallback or honest SKIP.

**Deliverable**: `research.md` with decisions, rationale, alternatives, and operator-blocked items.

---

### Phase 1 — Foundation: Submodules, Gateway, and Service Stubs

**Objective**: Remove blockers and close the largest unimplemented route surface.

#### T1.1 Fix submodules

- Add missing nested submodule URL for `submodules/helixtrack-core/Website/Upstreamable` or remove the orphan.
- Add `submodules/open-design` gitlink to the parent index.
- Commit updated submodule pointers (`constitution`, `containers`, `helixllm`, `helixqa`).
- Re-run `git submodule update --init --recursive` and capture output.

#### T1.2 Close gateway unimplemented routes

For each `notImplemented` route in `services/gateway-service/internal/server/server.go`, either implement the upstream endpoint in the owning service, or remove the route from gateway with documented out-of-scope reason.

Affected routes:

- `GET /api/v1/users/me`
- `PATCH /api/v1/users/me`
- `GET /api/v1/users/me/sessions`
- `DELETE /api/v1/users/me/sessions/:sessionId`
- `GET /api/v1/users/me/preferences`
- `PATCH /api/v1/users/me/preferences`
- `GET /api/v1/vaults/:vaultId/items`
- `POST /api/v1/vaults/:vaultId/items`
- `POST /api/v1/vaults/:vaultId/share`
- `POST /api/v1/hosts/:hostId/connect`
- `GET /api/v1/sessions/:sessionId/terminal`
- `POST /api/v1/sessions/:sessionId/share`
- `GET /api/v1/sessions/:sessionId/sftp`
- `POST /api/v1/sessions/:sessionId/sftp/download`
- `POST /api/v1/sessions/:sessionId/sftp/upload`
- `GET /api/v1/sessions/:sessionId/tunnels`
- `POST /api/v1/sessions/:sessionId/tunnels`
- `POST /api/v1/snippets/:snippetId/execute`
- `GET /api/v1/recordings/:recordingId/playback`
- `POST /api/v1/recordings/:recordingId/export`
- `GET /api/v1/analytics/usage`
- `POST /api/v1/ai/autocomplete`
- `POST /api/v1/ai/explain`
- `GET /api/v1/billing/subscription`
- `GET /api/v1/billing/usage`
- `GET /api/v1/system/maintenance`

#### T1.3 Implement missing upstream service endpoints

- AI service: `/autocomplete`, `/explain`.
- Billing service: subscription and usage endpoints.
- Vault service: items list/create, vault share.
- Host service: connect.
- Terminal service: terminal stream (WebSocket proxy in gateway).
- Collaboration service: session share.
- SFTP service: per-session SFTP endpoints.
- Port-forward service: per-session tunnels endpoints.
- Snippet service: execute endpoint.
- Recording service: playback stream, export.
- Analytics service: usage endpoint.
- User service: `/me`, `/me/sessions`, `/me/preferences`.

#### T1.4 Gateway wiring

- Wire Prometheus metrics (replace `// TODO: integrate with Prometheus client`).
- Implement WebSocket proxy to terminal-service.
- Implement SSO redirect and callback handlers (or mark operator-blocked).

**Deliverables**: Working gateway routes for all in-scope features; `research.md` updated with scope decisions.

---

### Phase 2 — Persistence & Contracts

**Objective**: Make every service persist to a real schema and expose real gRPC/OpenAPI contracts.

#### T2.1 Complete SQL migrations

- For each service, replace placeholder `migrations/001_init.up.sql` with a complete schema aligned with `docs/research/mvp/final/implementation/backend/<service>_db.sql`.
- Apply each migration to real PostgreSQL 17.2 in rootless podman; capture schema.
- Add rollback (`*.down.sql`) scripts where missing.

#### T2.2 Fix plaintext credential storage

- Encrypt `keychain-service` `private_key` and `passphrase` at rest.
- Add pre-store credential leak audit automation in CI.

#### T2.3 Define real gRPC services

- Replace `// TODO: define service RPCs and messages` in all 25 `.proto` files with real RPCs and messages per `docs/research/mvp/final/implementation/04-api-specification/`.
- Run `buf lint` and `protoc` compile for all 24 proto services; ensure no drift with Go structs.

#### T2.4 OpenAPI reconciliation

- Update `docs/research/mvp/final/implementation/api/openapi.yaml` to cover WebSocket streams and gRPC gateway mappings.
- Document honestly any gRPC/WS endpoints that remain uncovered.

**Deliverables**: Real per-service DB schemas, real proto contracts, updated OpenAPI spec.

---

### Phase 3 — Security Hardening

**Objective**: Close all CRITICAL/HIGH security gaps and satisfy constitution security clauses.

#### T3.1 Security tooling

- Populate `infrastructure/security/falco/README.md` with real Falco rules.
- Populate `infrastructure/security/cosign/README.md` with signing/verification policy.
- Populate `infrastructure/security/sealed-secrets/README.md` with SealedSecret templates.
- Expand `infrastructure/security/trivy/README.md` with CI scanning steps.

#### T3.2 Security runbooks

Write step-by-step procedures for:

- `docs/runbooks/certificate-rotation.md`
- `docs/runbooks/failover-procedure.md`
- `docs/runbooks/incident-response.md`
- `docs/runbooks/key-rotation.md`
- `docs/runbooks/postgres-pitr-restore.md`

Source procedures from `docs/research/mvp/final/implementation/05-security-zero-trust/README.md`.

#### T3.3 Helm / Kubernetes hardening

- Pin image tags + SHA-256 digests in `infrastructure/helm/helixterm/values.yaml`.
- Switch to `pullPolicy: Always` when digest-pinned.
- Convert umbrella `Chart.yaml` into per-service subcharts.
- Set HA replica counts per environment.
- Add fine-grained `NetworkPolicy` allow rules.
- Enforce Pod Security Standards (restricted) and consider Kyverno/OPA Gatekeeper.

#### T3.4 Auth middleware fleet-wide

- Wire shared JWT/SPIFFE/RBAC middleware into every service beyond `gateway-service` and `auth-service`.

#### T3.5 Replace `log.Fatalf` with graceful shutdown

- Audit `cmd/*/main.go` startup paths for `ai-service`, `billing-service`, `keychain-service`, `user-service`, `workspace-service`.
- Replace with structured logger + graceful shutdown.

#### T3.6 Health endpoint standardization

- Standardize all services on `/healthz/live` + `/healthz/ready` (or `/health/live` + `/health/ready`) with real dependency checks.

**Deliverables**: Security configs, populated runbooks, hardened Helm chart, auth middleware everywhere, graceful startup, standard health endpoints.

---

### Phase 4 — Test Coverage

**Objective**: Eliminate stub tests, add integration coverage to 15 services, and add stress/chaos fleet-wide.

#### T4.1 Convert/Remove stub tests

- Replace 41 `t.Skip` placeholders and 2 tautology tests with real assertions.
- Focus first on `ai-service`, `billing-service`, `keychain-service`, `user-service`, `gateway-service`, and `notification-service`.

#### T4.2 Integration tests

Add integration tests for the 15 services currently lacking them:

- analytics-service
- audit-service
- collaboration-service
- config-service
- health-service
- helixtrack-bridge-service
- host-service
- org-service
- recording-service
- sftp-service
- snippet-service
- ssh-proxy-service
- terminal-service
- workspace-service
- ai-service

Tests must exercise real internal packages or real DB-backed handlers.

#### T4.3 Stress tests

- Add `handler_stress_test.go` (or equivalent) to each of the 24 services without stress coverage, gated by `//go:build stress`.
- Include: sustained load, concurrent contention, boundary conditions.

#### T4.4 Chaos tests

- Add `handler_chaos_test.go` (or equivalent) to each of the 24 services without chaos coverage, gated by `//go:build chaos`.
- Include: input corruption, resource exhaustion, dependency failure.

#### T4.5 Test-bank reconciliation

- Fix `scripts/generate_test_banks.go` to generate real test bodies (not unused-import stubs).
- Regenerate `test/banks/unit/` and `test/banks/integration/`.
- Remove unrunnable banks (unit/a11y/compat) or rewrite to call real code.

**Deliverables**: Zero stub tests, integration tests in all 25 services, stress/chaos tests in all 25 services, real test banks.

---

### Phase 5 — Infrastructure & DevOps

**Objective**: Make Terraform, Helm, Kustomize, and CI/CD production-ready.

#### T5.1 Terraform validation

- Validate + plan all 6 modules and 3 environments via container.
- Reconcile modified modules: eks, elasticache, iam, msk, rds.

#### T5.2 Helm validation

- Run `helm lint` and `helm template` for the `helixterm` chart.
- Ensure `values-staging.yaml`, `values-production.yaml`, and base `values.yaml` are consistent.

#### T5.3 Kustomize validation

- Run `kustomize build` for base, dev, staging, and production overlays.

#### T5.4 Local stack

- Boot PostgreSQL, Redis, and Kafka via rootless podman compose with healthchecks.
- Verify all services can start against the local stack.

#### T5.5 CI/CD reconciliation

- Decide whether workflows remain disabled per §11.4.156 or are re-enabled with evidence.
- Update `.github/workflows/main.yml` and `nightly.yml` Go version to 1.25+.
- Fix non-existent paths: `test/banks/unit/...`, `test/banks/integration/...`, `test/e2e`, `test/integration`.
- Generate and commit missing test-bank directories OR remove the steps.

**Deliverables**: Validated Terraform/Helm/Kustomize, local stack boots, CI paths corrected.

---

### Phase 6 — Flutter Client

**Objective**: Replace stub tests, implement the core terminal view, and wire the design system.

#### T6.1 Toolchain

- Use rootless podman Flutter 3.24 image.
- Run `flutter pub get`, `flutter analyze` (target 0 issues), `flutter test`.

#### T6.2 Terminal emulator

- Implement real `terminal_view.dart`.
- Add required terminal/SSH package to `pubspec.yaml`.

#### T6.3 Burn down prod-source TODOs

- Replace ~30 placeholder stubs and ~110 demo wrappers with real implementations or remove.
- Add JSON serialization to the 4 models missing it.

#### T6.4 Design system wiring

- Wire `design_system/` tokens/theme into `lib/`.
- Implement light+dark variants.

#### T6.5 Bloc tests

- Replace 8 stub bloc tests with real tests.
- Achieve 19/19 BLoC coverage.

#### T6.6 Visual proof

- Generate host-rendered golden images for every screen×state×{light,dark}.
- Dual-validate with golden diff + OCR/vision oracle per §11.4.170.

#### T6.7 Integration tests

- Replace boot-only `integration_test/app_test.dart` with real end-to-end user journey tests.

**Deliverables**: Real terminal view, zero prod TODOs, design system wired, all bloc/integration tests real, visual proofs captured.

---

### Phase 7 — Observability

**Objective**: Wire real metrics, traces, logs, and alerts.

#### T7.1 Prometheus RED metrics

- Instrument all 25 services with request rate/error/duration metrics.
- Verify scrape endpoints and local Prometheus target health.

#### T7.2 Distributed tracing

- Propagate `X-Request-ID`/trace context across services.
- Verify spans in local Jaeger.

#### T7.3 Structured logging

- Ensure JSON structured logs with PII redaction.
- Verify Loki ingestion locally.

#### T7.4 Alert rules

- Add alert rules for error-rate, latency, and service-down.
- Fire-test alerts in local stack.

**Deliverables**: Live metrics, captured traces, structured logs, alert rules proven.

---

### Phase 8 — Documentation, Gap Closure & Coverage Ledger

**Objective**: Close documentation gaps, update ledgers, and sync exports.

#### T8.1 Close GAP-01..GAP-15

- Update `docs/research/mvp/final/implementation/GAP_REGISTER.md` with evidence per gap.
- Mark gaps closed, honestly deferred, or operator-blocked.

#### T8.2 Update reviews

- Refresh `docs/review/SECURITY_REVIEW.md`, `CODE_SCAFFOLD_REVIEW.md`, `OPENDESIGN_COMPLIANCE_REVIEW.md`.

#### T8.3 Coverage ledger

- Regenerate `docs/COVERAGE_LEDGER.md` after Phase 4.
- Ensure six-invariant coverage per service.

#### T8.4 QA transcripts

- Create `docs/qa/<feature>/` for every shipped feature with end-to-end communication transcripts per §11.4.83.

#### T8.5 Continuation document

- Update `docs/CONTINUATION.md` to revision 19+; reconcile parked items (HEL-012, HEL-013, etc.).

#### T8.6 Docs sync & exports

- Fix version-string drift (postgres/go/kafka/redis).
- Regenerate md/html/pdf/docx exports in sync.
- Render 30 mermaid + 8 drawio diagrams where renderers exist.

**Deliverables**: GAP_REGISTER closed, reviews refreshed, coverage ledger current, QA transcripts present, continuation current, synced exports.

---

### Phase 9 — Release Preparation

**Objective**: Verify everything together and publish a release.

#### T9.1 Full-suite retest from clean baseline

- Run all gates: `bash tests/verify_constitution_inheritance.sh`, `bash scripts/testing/meta_test_false_positive_proof.sh`, `bash tests/docs_consistency_gate.sh`, `make test`.
- Run integration, e2e, stress, and chaos suites.

#### T9.2 Production-readiness review

- Verify Helm/Terraform are production-ready.
- Confirm CI is compliant/disabled per §11.4.156.

#### T9.3 Release tag

- Create prefixed release tag per §11.4.151.
- Push to all upstreams; mirror tag on every owned submodule.
- Merge-onto-latest-main, no force-push.

#### T9.4 Continuation & changelog

- Update `docs/CONTINUATION.md`, `CHANGELOG.md`, and memory files.

**Deliverables**: All gates green, release tag published, submodules tagged, continuation/changelog updated.

---

## Key Rules

- Use absolute paths for filesystem operations; use project-relative paths for documentation references.
- ERROR on gate failures or unresolved clarifications.
- Every task closes with captured evidence.

## Done When

- [ ] `research.md` complete with all unknowns resolved and operator-blocked items documented.
- [ ] Gateway routes implemented or removed with honest scope reasons.
- [ ] All 25 services have real DB schemas, proto contracts, and OpenAPI coverage.
- [ ] Security runbooks, tooling configs, and Helm hardening complete.
- [ ] Zero stub tests; integration/stress/chaos tests in every service.
- [ ] Terraform/Helm/Kustomize validate; local stack boots.
- [ ] Flutter terminal view real, design system wired, visual proofs captured.
- [ ] Observability stack proven end-to-end.
- [ ] GAP_REGISTER, reviews, coverage ledger, QA transcripts, and continuation current.
- [ ] Full-suite retest green; release tag published; submodules tagged.
