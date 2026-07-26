# Coverage Ledger — §11.4.25 Full-Automation-Coverage

**Revision:** 2  
**Last modified:** 2026-07-26  
**Generated against:** working tree at `40416f9` + Completion Initiative pass  
**Method:** direct `find`/`grep` against the current working tree plus executed test runs; no guessing (§11.4.6).

---

## 0. Fleet-wide totals

| Metric | Count |
|---|---|
| Services under `services/*/` | 25 |
| Test files (`*_test.go`) | 267 |
| Test functions (`func Test...`) | 1161 |
| Services with integration tests | 18 |
| Services with stress tests | 25 (25 files) |
| Services with chaos tests | 25 (25 files) |
| Dedicated security test files | 6+ (auth, container-bridge, gateway SSRF, notification, vault IDOR) |
| Flutter client test files | 11 (8 real BLoC + api_client + widget + golden + 1 integration) |

**Change since rev 1 (c2718d2 → 40416f9 + this pass):**
- `*_test.go` files: 194 → 267 (+73)
- Test functions: 834 → 1161 (+327)
- Integration tests: 10 services → 18 services (15 new Podman-backed repo/handler integration tests; existing 3 services already had them)
- Stress/chaos: 1 service → 25 services (24 new files added, 1 existing auth-service retained)
- Stub/tautology tests: 0 remaining in default mode; `test/banks/*` generator and stubs removed

---

## 1. Per-service coverage table

Legend: **R**=real substantive assertions · **M**=moderate (real CRUD, in-memory or negative-path only) · **S**=shallow (constructor non-nil / not-connected error) · **X**=stub-bluff (tautology or unconditional `t.Skip`) · **A**=absent. Integration/Security/Stress/Chaos: **R**=present+real, **A**=absent.

| Service | Files | Funcs | Main | Model | Repo | Server | Handler | Extra | Integ. | Security | Stress | Chaos |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| ai-service | 10 | 34 | R | R | R(integ) | R | R(3) | — | R | A | R | R |
| analytics-service | 8 | 29 | R | R(2) | R(integ) | R | R(7) | — | R | A | R | R |
| audit-service | 8 | 36 | R | R(6) | R(integ) | R(4) | R(14) | — | R | A | R | R |
| auth-service | 12 | 45 | R(integ) | R(4) | R(5) | R(3) | R(3) | crypto R(4) | R(4 files) | partial | R(4) | R(3) |
| billing-service | 10 | 26 | R | R | R(integ) | R | R(3) | — | R(2 files) | A | R | R |
| collaboration-service | 8 | 33 | R | R(2) | R(integ) | R(2) | R(16) | — | R | A | R | R |
| config-service | 8 | 29 | R | R(2) | R(integ) | R(3) | R(8) | — | R | A | R | R |
| container-bridge-service | 16 | 39 | R | R(2) | R(integ) | R | R(8) | red_injection R(1) | R(1 file) | R(injection) | R | R |
| gateway-service | 8 | 41 | R | R | R | R | R(23+2) | resolve_path_params R(1) | R(2 funcs) | R(SSRF) | R | R |
| health-service | 8 | 35 | R(minimal) | R(5) | R(integ) | R(5) | R(11) | checker R(10) | R | A | R | R |
| helixtrack-bridge-service | 11 | 35 | R | R(2) | R(integ) | R | R(8) | coreclient R(2) | R | A | R | R |
| host-service | 8 | 26 | R(2) | R(4) | R(integ) | R(3) | R(4) | — | R | A | R | R |
| keychain-service | 11 | 47 | R | R | R(integ) | R | R(3) | crypto R(5) | R(1 file) | A | R | R |
| notification-service | 16 | 107 | R | R | R(12) | R(10) | R(14+3) | delivery R(25, 5 files) | R(4 files) | R(injection+ssrf+idor) | R | R |
| org-service | 8 | 22 | R | R(2) | R(integ) | R | R(6) | — | R | A | R | R |
| pki-service | 10 | 32 | R | R(4) | R(integ) | R | R(5) | crypto R(8) | R(1 file) | A | R | R |
| port-forward-service | 11 | 36 | R | R(2) | R(integ) | R | R(8+4) | forwarder R(5+3) | R(1 file) | A | R | R |
| recording-service | 8 | 25 | R | R(3) | R(integ) | R | R(8) | — | R | A | R | R |
| sftp-service | 8 | 25 | R | R(3) | R(integ) | R | R(8) | — | R | A | R | R |
| snippet-service | 8 | 23 | R | R(1) | R(integ) | R | R(8) | — | R | A | R | R |
| ssh-proxy-service | 10 | 34 | R(2) | R(4) | R(integ) | R(2) | R(3) | sshclient R(5), wshandler R(5) | R | A | R | R |
| terminal-service | 9 | 43 | R | R(3) | R(integ) | R(5) | R(9) | recorder R(13) | R | A | R | R |
| user-service | 9 | 21 | R | R | R(integ) | R | R(3) | — | R(1 file) | A | R | R |
| vault-service | 10 | 64 | R(49L) | R(2) | R(2)+R(integ,3) | R(9)+R(integ,7) | R(15) | — | R(2 files) | R(IDOR) | R | R |
| workspace-service | 8 | 21 | R | R(3) | R(integ) | R(1) | R(4) | — | R | A | R | R |

Counts in parens are test-function counts per layer, cited from per-file `grep -c "func Test"`. Integration tests are gated by `//go:build integration` and use rootless Podman PostgreSQL 17.2.

---

## 2. §11.4.85 Stress/chaos coverage

| Service | Stress | Chaos | Notes |
|---|---|---|---|
| auth-service | **R** (415L, 4 funcs) | **R** (404L, 3 funcs) | Real: sustained-load + concurrent-contention + boundary-conditions (stress); input-corruption + resource-exhaustion + boundary-conditions (chaos). Gated on `//go:build stress`/`//go:build chaos` tags. |
| All other 24 services | **R** (new `handler_stress_test.go`) | **R** (new `handler_chaos_test.go`) | Added in this pass. Each exercises the real HTTP handler layer under load and with corrupted/edge inputs. |

---

## 3. §11.4.27 Stub/tautology inventory

**Result:** 0 unconditional `t.Skip` stubs or `assert.True(t, true)` tautologies remain in default mode across the 25 services.

- `scripts/generate_test_banks.go` and all `test/banks/*` stubs were removed as bluff-free cleanup.
- Any remaining `t.Skip` calls are gated by environment conditions (e.g., `DATABASE_URL` absent, Podman unavailable) and include an honest §11.4.3-compliant reason.

---

## 4. §11.4.83 docs/qa/ status

**`docs/qa/` does not exist.** No end-to-end QA transcripts have been committed. Per §11.4.83, every feature that ships MUST carry a recorded e2e communication transcript under `docs/qa/<run-id>/`. This remains a release-blocking gap.

---

## 5. §11.4.25 Six-invariant coverage per service

| Invariant | Status | Notes |
|---|---|---|
| (1) Anti-bluff posture | IMPROVED | All default-mode tests are real; no tautologies or unconditional stubs remain. Stress/chaos added fleet-wide. |
| (2) Working capability end-to-end | IMPROVED | 18 services have integration tests; 15 new Podman-backed integration tests added in this pass. |
| (3) Working implementation matching docs | NOT ASSESSED | No `docs/qa/` transcripts exist to cross-reference. |
| (4) No open issues/bugs | NOT ASSESSED | No `docs/Issues.md` or workable-items DB exists. |
| (5) Full documentation | NOT ASSESSED | User manual entries not verified. |
| (6) Four-layer test floor | PARTIAL | Pre-build gates and integration pass; runtime paired-mutation layer verified only for constitution gate. |

---

## 6. Flutter client (`clients/flutter`)

11 test files total:

| File | Status | Evidence |
|---|---|---|
| `test/api_client_test.dart` | REAL | 7 assertions covering GET/POST/PUT/DELETE, headers, auth, and error handling |
| `test/auth_bloc_test.dart` | REAL | Full BLoC state-transition tests with `bloc_test` + `mocktail` |
| `test/collaboration_bloc_test.dart` | REAL | Full BLoC state-transition tests |
| `test/host_bloc_test.dart` | REAL | Full BLoC state-transition tests |
| `test/notification_bloc_test.dart` | REAL | Full BLoC state-transition tests |
| `test/terminal_bloc_test.dart` | REAL | Full BLoC state-transition tests |
| `test/vault_bloc_test.dart` | REAL | Full BLoC state-transition tests |
| `test/workspace_bloc_test.dart` | REAL | Full BLoC state-transition tests |
| `test/widget_test.dart` | REAL | Splash-screen render assertion |
| `test/terminal_view_golden_test.dart` | REAL | 149L, 3 testWidgets, §11.4.170-compliant golden+content-oracle+VT-engine |
| `integration_test/app_test.dart` | STUB (e2e) | Boot-only smoke, explicit `TODO: implement real e2e tests` |

**Evidence:** `flutter analyze` is clean; `flutter test` passes 69 tests.

**Gap:** 64 TODOs remain in `clients/flutter/lib/` (down from 65; one resolved by design-system wiring). The integration test is boot-only.

---

## 7. Release-blocking gaps (ordered by risk per §11.4.132)

1. **[CRITICAL] `docs/qa/` does not exist.** §11.4.83 requires e2e QA transcripts for every shipped feature.

2. **[CRITICAL] Operator-blocked submodule issues.** `submodules/helixtrack-core/Website/Upstreamable` nested submodule URL orphan blocks `git submodule update --init --recursive`; `submodules/open-design` is in `.gitignore` and lacks a gitlink.

3. **[HIGH] Gateway honest-501 routes remain.** 26 routes in `services/gateway-service/internal/server/server.go` return `501 Not Implemented` because no real upstream capability exists. These are documented as honest gaps, not stubs, but they still represent unimplemented functionality.

4. **[HIGH] Prometheus metrics stub in gateway.** `services/gateway-service/internal/server/server.go:1006-1008` returns a static placeholder for `/metrics` instead of real Prometheus metrics.

5. **[MEDIUM] SSO handlers stubbed.** `services/gateway-service/internal/server/server.go:1021-1029` returns 501 for SSO redirect and callback.

6. **[MEDIUM] Flutter integration test is boot-only.** Needs real end-to-end user journey tests.

7. **[LOW] ~64 Flutter lib TODOs remain.** Non-blocking for backend release but needed for client completeness.

---

## 8. Changes since rev 1 (c2718d2 → this pass)

| Change | Detail |
|---|---|
| +73 test files | 194 → 267 |
| +327 test functions | 834 → 1161 |
| Integration tests | 10 services → 18 services (15 new repository/handler integration tests) |
| Stress/chaos tests | 1 service → 25 services (+24 new `handler_stress_test.go` / `handler_chaos_test.go`) |
| Stub removal | 41 `t.Skip` + 2 tautology files eliminated; `test/banks/*` and `scripts/generate_test_banks.go` removed |
| Flutter tests | 8 stub BLoC tests replaced with real `bloc_test` + `mocktail` tests; `flutter analyze` clean; 69 tests pass |
| Infrastructure | `helm lint`, `helm template`, `terraform validate` pass |

---

## 9. Notes / honest boundaries

- Test **execution** was run for integration tests (15 services), `go vet` (all services), `helm lint`/`template`, `terraform validate`, Flutter `analyze`/`test`, and constitution gates.
- Stress/chaos tests are build-gated by `//go:build stress`/`//go:build chaos` and were not batch-executed in this pass; files exist and compile.
- Operator decisions required for T008 (nested submodule orphan) and T009 (`open-design` gitlink / `.gitignore`).
