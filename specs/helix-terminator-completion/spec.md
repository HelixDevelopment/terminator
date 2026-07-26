# helix_terminator Completion Initiative

| Field | Value |
|---|---|
| Feature ID | `HELIX-COMPLETION-2026` |
| Date | 2026-07-25 |
| Authority | docs/research/completion/report/2026.07.25.md, docs/COVERAGE_LEDGER.md, docs/CONTINUATION.md, docs/research/mvp/final/implementation/DEVELOPMENT_PLAN.md, docs/research/mvp/final/implementation/GAP_REGISTER.md |
| Status | ACTIVE — planning |

## Summary

Close the implementation gap between the documented MVP and the current repository state. The completion report inventories every unimplemented route, stub test, empty migration, placeholder proto, empty runbook, incomplete Helm chart, disabled CI workflow, and Flutter stub. This initiative turns those findings into an executable, phased plan with evidence captured for each closure.

## Scope

- Entire Go microservice fleet (25 services).
- gRPC / OpenAPI contracts.
- SQL migrations and persistence schemas.
- Security, runbooks, and infrastructure (Helm/Terraform/NetworkPolicy).
- CI/CD workflows and test banks.
- Flutter client and its test suite.
- Documentation, QA transcripts, and coverage ledger.

## Invariants

- Anti-bluff: every task closes with captured physical evidence (test run, real DB delta, rendered output, etc.).
- Red-first: a failing test must reproduce the defect before the fix.
- No fakes beyond unit tests: integration/stress/chaos tests exercise real systems.
- Constitution gates run before any commit.

## Actors

- **Operator / SRE**: deploys the system, validates infrastructure, and reviews runbooks.
- **Backend Engineer**: implements and tests Go microservice features.
- **Mobile / Frontend Engineer**: implements and tests the Flutter client and design system.
- **Security Engineer**: validates zero-trust controls, mTLS, secrets management, and audit controls.
- **End User**: interacts with the Flutter client to manage SSH sessions, hosts, workspaces, vaults, recordings, snippets, and billing.

## User Stories

### US1 — Service completion
As a backend engineer, I want every Go service to have real (non-stub) tests, real persistence, and real endpoint wiring so that the fleet can be deployed with confidence.

### US2 — Gateway route coverage
As an operator, I want the gateway to route to real upstream endpoints where an upstream exists, or return an honest 501 with documented evidence where it does not, so that clients never receive fabricated 200s from non-existent features.

### US3 — Security hardening
As a security engineer, I want mTLS, RBAC, network policies, pinned images, sealed secrets, and runbooks in place so that the production deployment meets zero-trust and compliance requirements.

### US4 — Infrastructure readiness
As an operator, I want valid Terraform, Helm, and Kustomize configurations with production-ready values so that the system can be deployed to staging and production.

### US5 — Flutter client completion
As a mobile engineer, I want the Flutter client to analyze cleanly, pass real widget/BLoC tests, and implement the terminal view and design system so that end users can interact with the platform.

### US6 — Observability
As an SRE, I want Prometheus metrics, distributed traces, structured logs, and alerts wired into every service so that production health is visible and actionable.

### US7 — Documentation and release
As a release engineer, I want QA transcripts, an accurate coverage ledger, a closed gap register, and a tagged release so that the deployment is auditable and reproducible.

## Functional Requirements

- FR1: Every Go service must build, pass `go vet`, and pass all default tests.
- FR2: Every service with persistence must have a real integration test backed by PostgreSQL 17.2 in rootless Podman.
- FR3: Gateway routes must proxy to real upstream endpoints where they exist; otherwise return a documented 501.
- FR4: All 25 `.proto` files must define real RPCs and messages matching the OpenAPI contract.
- FR5: Every service migration must apply cleanly to PostgreSQL 17.2 and include a rollback script.
- FR6: Security runbooks must cover certificate rotation, failover, incident response, key rotation, and Postgres PITR restore.
- FR7: Helm chart must use digest-pinned images, NetworkPolicy rules, and Pod Security Standards.
- FR8: Flutter client must pass `flutter analyze`, real widget/BLoC tests, and golden tests.
- FR9: Every service must expose Prometheus RED metrics, structured JSON logs, and propagate trace context.
- FR10: QA transcripts must exist for every shipped feature before release.

## Success Criteria

- SC1: 100% of services pass `go test ./...` in default mode.
- SC2: 100% of services with persistence pass integration tests with `-tags integration`.
- SC3: 100% of services have real stress and chaos tests that compile and run under their build tags.
- SC4: `helm lint`, `helm template`, `terraform validate`, and `kustomize build` all pass.
- SC5: `flutter analyze` reports 0 issues and `flutter test` passes.
- SC6: All security runbooks are populated and reviewed.
- SC7: GAP_REGISTER.md has no open critical/high gaps at release time.
- SC8: A release tag is created and pushed, and the constitution gates pass on the tagged commit.

## Edge Cases

- A service has no persistence layer: integration test can be skipped with an honest documented reason.
- An upstream service lacks a capability needed by a gateway route: gateway returns 501 with evidence, never a fabricated 200.
- External secrets (Stripe, Slack, SMTP) are unavailable: tests skip with §11.4.3-compliant reason, not tautologies.
- Tool not on host (helm, terraform, kustomize, k6, protoc): use containerized fallback and capture evidence.

## Assumptions

- PostgreSQL 17.2 can be started via rootless Podman for integration tests.
- The operator may provide external secrets later; their absence does not block shipping core functionality.
- The existing monorepo structure and service boundaries are preserved.

## Clarifications

### Session 2026-07-26

- Q: Should the 26 gateway `notImplemented` routes be implemented or kept as honest 501s? → A: Implement only those whose upstream service already exposes a matching capability; keep the rest as documented 501s until the upstream service is extended.
- Q: Should external-secret integrations (Stripe, Slack, SMTP) be faked or skipped? → A: Skip with an honest §11.4.3 reason; never fake the external service in tests.
- Q: What is the minimum client target for Flutter? → A: Flutter 3.24.0 as pinned in `clients/flutter/pubspec.yaml`; build for all OpenDesign platforms (iOS, Android, Web, Desktop).
