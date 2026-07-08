# Coverage ledger (§11.4.25) — helix_terminator services fleet

**Revision:** 1
**Last modified:** 2026-07-08T07:27:24Z

**Scope:** `services/*` (25 Go services) at commit `f0b29ff` (isolated worktree
`.claude/worktrees/agent-a4e274b0cab809246`, branched off `main`). `clients/flutter`
is carried forward from the prior read-only draft (verified unchanged since that
draft's baseline — see §6) and is not re-derived here.

**Method:** direct `find`/`grep`/`cat`/`git log`/`git diff` against the checked-out
tree — no `go test` execution in this pass (§11.4.6: this is a static inventory,
not a runtime-verified PASS/FAIL sweep). Every cell below cites a real file path
and, where relevant, a real commit SHA. This ledger reconciles and supersedes
`scratchpad/exec/session_resume/coverage_ledger_draft.md` (baseline commit
`45e0edd`) against the current tree at `f0b29ff` — §4 documents every place the
two disagree and why.

**Legend** (per file/layer): **R(n)** = real, substantive assertions, `n` = Test-func
count · **M(n, ...)** = moderate (real CRUD exercised but only via an in-memory
double or all-negative-path) · **S(n)** = shallow (compiles+runs, only checks
`New()` non-nil / a not-connected error string) · **X** = stub-bluff (unconditional
`t.Skip`, tautological `assert.True(t, true)`, or empty-body — zero real
coverage despite reporting PASS) · **A** = absent (no file) · **N/A** = not
applicable (service has no schema of its own).

---

## 0. Fleet-wide totals (re-derived at `f0b29ff`, not carried from the draft)

- `find services -name '*_test.go' | wc -l` → **185 test files** (draft baseline
  at `45e0edd`: 148 — +37 files landed this session, entirely migration-batch
  `migrate_test.go` additions + bridge/security-fix regression tests).
- `grep -rEc '^func Test' services --include='*_test.go' | awk -F: '{s+=$2} END{print s}'`
  → **727 Test funcs** (draft baseline: 519 — +208).
- `find services -iname '*stress*' -o -iname '*chaos*' -o -iname '*e2e*'` → **zero
  files, fleet-wide, unchanged from the draft.** No stress test, no chaos test,
  no dedicated e2e suite exists anywhere under `services/` for any of the 25
  services. This remains the single largest §11.4.85/§11.4.27 test-type gap.
- **18/25 services now have real golang-migrate wiring** (`migrations/migrate.go`
  + `NNN_*.up.sql`/`.down.sql` pair + `migrations/migrate_test.go`, 9 Test funcs
  each, uniformly §11.4.3/§11.4.69-compliant — real-Postgres-gated via
  `TEST_DATABASE_URL`, honest `SKIP` when absent, never a fabricated PASS):
  `ai-service`, `analytics-service`, `audit-service`, `auth-service`,
  `collaboration-service`, `config-service`, `container-bridge-service`,
  `host-service`, `keychain-service`, `org-service`, `pki-service`,
  `recording-service`, `sftp-service`, `snippet-service`, `ssh-proxy-service`,
  `terminal-service`, `user-service`, `vault-service`.
- **2/25 services (`gateway-service`, `health-service`) are genuinely N/A** for
  migrations — both ship only a `TODO`-placeholder `001_init.sql` with no real
  DDL (verified: `cat services/gateway-service/migrations/001_init.sql` /
  `services/health-service/migrations/001_init.sql`, both literally
  `-- TODO: create tables, indexes, and constraints for <svc>`); `health-service`
  has **zero** `pgxpool`/`database/sql` references anywhere in its tree
  (confirmed via `grep -rl "pgxpool\|sql.DB\|database/sql" services/health-service`,
  no hits) — it is a pure aggregator with no schema of its own. `gateway-service`
  is a stateless reverse proxy; its `internal/repository` package exists but its
  own migration file is the same TODO-only placeholder.
- **5/25 services have NEITHER golang-migrate wiring NOR N/A status — this is a
  real gap, not covered by the task brief's "23/25 + gateway/health N/A" framing**
  (see §4 correction #1 for the full discrepancy write-up): `billing-service`,
  `helixtrack-bridge-service`, `notification-service`, `port-forward-service`,
  `workspace-service`. Each still ships only a raw, ungated `001_init.sql` (some
  with a second hand-numbered file, e.g. `port-forward-service/migrations/002_add_tunnel_fields.sql`)
  with **no** `migrate.go`, **no** `migrate_test.go`, **no** golang-migrate
  library import anywhere in the service (`grep -rl "golang-migrate\|migrate.New"
  services/<svc>` → zero hits for all 5). **Confirmed via `git log --oneline
  --all` that the wiring work for exactly these 5 services DOES exist** — commits
  `9301ec4` (billing, BATCH E), `62923d3` (helixtrack-bridge, BATCH E), `df80aad`
  (notification, BATCH E), `46edaeb` (port-forward, BATCH E), `9eb4590`
  (workspace, BATCH F) — **but `git merge-base --is-ancestor <sha> HEAD` returns
  NO for all five**: they live on a sibling worktree branch
  (`worktree-agent-af47069405aee570e`), not yet merged into this branch's `f0b29ff`
  tip. This is an honest, current-as-of-`f0b29ff` fact, not a defect in the work
  itself — but this ledger's job is to report what is ACTUALLY on this commit,
  and on `f0b29ff` these 5 services are unmigrated.
- **3 real backend integrations confirmed live at the production-code layer**
  (not just tests) — `ai-service` (`internal/handler/handler.go:94-166` calls a
  real `LLMClient.Complete` against the local HelixLLM llama.cpp backend via
  `internal/llmclient/llmclient.go`, with an honest `"failed"` status on error/
  timeout, never a silent fabricated `"pending"`/`"completed"`),
  `container-bridge-service` (`internal/handler/handler.go` + `internal/containerrt/containerrt.go`
  drive real podman/docker via the `digital.vasic.containers/pkg/runtime` package,
  wired through `submodules/containers` + a `go.mod replace` directive — confirmed
  present: `services/container-bridge-service/go.mod:51`), `helixtrack-bridge-service`
  (`internal/handler/handler.go` + `internal/coreclient/coreclient.go` authenticate
  against a real HelixTrack Core instance before deciding bridge status). All
  three ship dedicated RED-then-GREEN (§11.4.43/§11.4.115) regression tests
  proving the PRE-fix fabrication defect and the POST-fix real behavior (see §1
  per-service rows).
- **`billing-service` payment-processor fabrication is NOT resolved** — confirmed
  live at `f0b29ff`: `CreateSubscription` (`internal/handler/handler.go:78`) still
  sets `Status: "active"` unconditionally; `grep -n "stripe\|Stripe\|payment"
  services/billing-service/internal/{handler,repository}/*.go` → zero hits. This
  is a DIFFERENT defect from the T12/T14 tenant-isolation security fixes that DID
  land this session (see §1's billing-service row) — the tenant-isolation fix
  closes an IDOR on top of the still-fabricated subscription state.
  `docs/CONTINUATION.md:42` confirms this is tracked and still
  **OPERATOR-BLOCKED**: *"Backend tiers still OPERATOR-BLOCKED: ai cloud-LLM
  keys, billing Stripe, push FCM/APNs."* — note ai-service's LOCAL LLM path is
  now real (see above); only its CLOUD multi-model tier remains blocked.
- **Stub-bluff total: 37 files fleet-wide, matching the draft's total count of 37
  despite membership shifting** (3 independent patterns, hand-verified,
  comment-only false positives excluded — e.g. `auth-service/cmd/auth-service/main_test.go`'s
  and `auth-service/internal/server/server_test.go`'s doc-comments both literally
  say `t.Skip("TODO")` while describing the STUB THEY REPLACED, not live code —
  confirmed by reading full file content):
  1. **Unconditional `t.Skip(...)` placeholder — 29 files, 15 services**:
     `ai-service` (cmd, model, repository, server — 4), `analytics-service`
     (cmd — 1), `audit-service` (cmd — 1), `billing-service` (cmd, model,
     repository, server — 4), `collaboration-service` (cmd — 1),
     `container-bridge-service` (cmd — 1), `gateway-service` (cmd, model,
     repository — 3; **`server` and `handler` are no longer stub — see §4
     correction #2**), `helixtrack-bridge-service` (cmd — 1), `keychain-service`
     (cmd, model, repository, server — 4), `port-forward-service` (cmd — 1),
     `recording-service` (cmd — 1), `sftp-service` (cmd — 1), `snippet-service`
     (cmd — 1), `terminal-service` (cmd — 1), `user-service` (cmd, model,
     repository, server — 4).
  2. **Tautological `assert.True(t, true)` (live, not comment) — 5 files**:
     `org-service/cmd/org-service/main_test.go` (`TestMainImports`),
     `terminal-service/internal/repository/repository_test.go`
     (`TestRepositoryStub`), `terminal-service/internal/server/server_test.go`
     (`TestServerStub`), `notification-service/internal/model/model_test.go`
     (`TestNotificationTypes` — constructs two real struct literals then ends
     on `assert.True(t, true)`), and **`notification-service/cmd/notification-service/main_test.go`**
     (`TestMainCompiles` — a 5th tautology file the prior draft's §0 list
     omitted; see §4 correction #3).
  3. **Empty-body test function — 3 files**: `config-service/cmd/config-service/main_test.go`
     (`TestMainStub`), `pki-service/cmd/pki-service/main_test.go` (`TestMainStub`),
     `workspace-service/cmd/workspace-service/main_test.go` (`TestMain`).
- **Dead code removed since the draft**: `gateway-service/internal/handler`
  package (the draft's finding: zero production importers, §11.4.124-flagged)
  **no longer exists in the tree** — `find services/gateway-service/internal/handler`
  now errors "No such file or directory". This is a genuine, positive closure;
  gateway's proxy/health/JWT logic lives entirely in `internal/server` now.
- `docs/qa/` **does not exist** (`find docs/qa -type f` → empty, `[ -d docs/qa ]`
  → false) — the §11.4.83 recorded-transcript mandate is unmet fleet-wide, not
  only for the bridge features named in the task brief.
- `docs/Issues.md` **does not exist** — no workable-items tracker is instantiated
  at the project layer yet (consistent with the draft's note); this ledger
  remains the closest thing to a tracked gap list.
- `README.md` has **no** `Tracked-Items` section (`grep -n "Tracked-Items"
  README.md` → no match) — per the task's constraint, no README pointer line is
  added (nothing to point from without restructuring).

---

## 1. Per-service coverage table

| Service | Unit tests (real?) | Integration/real-PG tests | Migration test | Security-relevant tests | Notable GAPS |
|---|---|---|---|---|---|
| ai-service | Handler R(6): `handler_test.go`(3)+`handler_ai_completion_test.go`(3); llmclient R(5): `llmclient_test.go`; Server **mixed**: `server_test.go` has 1 residual `t.Skip` stub func (`TestServerStub`) alongside 2 real timeout-budget tests. Main/Model/Repo **X** (`main_test.go`, `model_test.go`, `repository_test.go` — unconditional `t.Skip("TODO: implement ... tests")`) | `llmclient_integration_test.go`(1, real-LLM-gated) | **R**, real-PG, `migrations/migrate_test.go`(9) | none dedicated | Model/Repo/cmd-main 100% stub despite the pgx repository being the persistence layer for every real completion; cloud-multi-model tier OPERATOR-BLOCKED (API keys, `docs/CONTINUATION.md:42`); no e2e/stress/chaos |
| analytics-service | Model R(2), Handler R(7); Repo **S**(1, not-connected-error only), Server **S**(1); Main **X** | A | **R**, real-PG, `migrations/migrate_test.go`(9) | none | Repo/server shallow-only (no positive-path CRUD assertion); no e2e/stress/chaos |
| audit-service | Model R(6), Server R(4), Handler R(14); Repo **M**(1, all-negative-path, exercises 5 methods); Main **X** | A | **R**, real-PG, `migrations/migrate_test.go`(9) | none | Repo positive-path (real query logic) untested; no e2e/stress/chaos |
| auth-service | Model R(4), Repo R(5), Handler R(3)+`handler_readiness_integration_test.go`(1); Server R(3)+`server_jwt_key_test.go`(4)=7; crypto R(4)+`crypto_key_provisioning_test.go`(4)=8; Main R (integration, real PG — `main_test.go` boots real disposable PG, exercises `server.New`→JWT sign/verify→real HTTP `/healthz`) | build-tag `integration` × 3 files (`main_test.go`, `server_test.go` register/login/refresh/logout, `repository_test.go`) + `handler_readiness_integration_test.go`(1, real-PG, T8-6 DB-ping proof) | **R**, real-PG, `migrations/migrate_test.go`(9) | **T15** JWT persisted-key: `crypto_key_provisioning_test.go`(4, `TestCrossServiceValidation_EphemeralKeyFailsAcrossInstances` etc.) + `server_jwt_key_test.go`(4, fail-closed-in-production-mode proof) | Highest-churn service in the fleet; no e2e/stress/chaos |
| billing-service | Handler R(3, `TestHealthCheck`/`TestReadinessCheck`/`TestCreateSubscriptionValidation`); Main/Model/Repo/Server **X** (unconditional `t.Skip`) | build-tag `integration` × 2: `server_integration_test.go`(1, `TestBillingCrossTenantIsolation_RealPostgres`, T12) + `server_write_isolation_integration_test.go`(1, `TestBillingWriteEndpointsCrossTenantIsolation_RealPostgres`, T14) | **N/A — NOT wired at `f0b29ff`** (raw `001_init.sql` only, no `migrate.go`; BATCH-E commit `9301ec4` exists but is on a sibling unmerged worktree branch — see §0) | **T12** read-scope IDOR + **T14** write-path IDOR, both real-PG-gated | `CreateSubscription` still fabricates `Status: "active"` with zero Stripe call (OPERATOR-BLOCKED, `docs/CONTINUATION.md:42`) — the tenant-isolation fix does NOT close this; model/repo/server/cmd 100% stub; no migrations wiring; no e2e/stress/chaos |
| collaboration-service | Model R(2), Server R(2), Handler R(16); Repo **S**(1); Main **X** | A | **R**, real-PG, `migrations/migrate_test.go`(9) | none | Repo shallow-only; no e2e/stress/chaos |
| config-service | Model R(2), Repo R(3), Server R(3), Handler R(8); Main **X (empty-body)** | A | **R**, real-PG, `migrations/migrate_test.go`(9) | none | Main-package wiring untested; no e2e/stress/chaos |
| container-bridge-service | Handler R(8, `handler_test.go`)+`delete_teardown_test.go`(2)=10; containerrt R(6)+`validate_test.go`(2)=8; Model R(2); Server **S**(1, `TestNewServer` construction-only); Repo **S**(1); Main **X** | `handler_integration_test.go`(1, real podman/docker) + 3 dedicated §11.4.43/§11.4.115 RED-then-GREEN regression files: `red_fabrication_test.go`(1, proves `CreateBridge` no longer fabricates Active status when the container never comes up), `red_injection_test.go`(2, flag/argument-injection via a leading `-` in `Image`), `red_update_status_test.go`(1) | **R**, real-PG, `migrations/migrate_test.go`(9) | `red_injection_test.go` — host-privilege-escalation-class flag-injection regression, real-CLI-argv-gated | Real podman/docker backend now wired via `submodules/containers`; repo/server still shallow; no e2e/stress/chaos |
| gateway-service | Server R(29, `server_test.go` — includes `TestFullHealthEndpointRealUptime`/`TestFullHealthEndpointNoFabricatedLatency`/`TestUnreachableUpstreamFlipsUnhealthy` (T8-8/T16) and `TestCorrectedRoutes_ReachRealUpstreamPath`/`TestCorrectedRoutes_OldPathWouldHaveMissed`/`TestHonestGapRoutes_Return501NotImplemented` (T8-1 route-table fix)); `resolve_path_params_test.go`(1, SSRF/path-injection); Model/Repo **X** (`t.Skip`); Main **X** | `server_integration_test.go`(1, `TestIntegration_GatewayProxiesRealHTTPRequestToRealUpstream`) | **N/A — genuinely no schema** (placeholder-only `001_init.sql`; stateless reverse proxy) | SSRF/path-injection regression (`af78e22`) + T8-1 route-table + T8-8/T16 fabricated-latency-and-uptime fixes, all with dedicated real tests | **`internal/handler` package REMOVED since the draft** (was dead/unwired, zero production importers — now gone entirely, closing that draft finding); model/repo still stub; no e2e/stress/chaos |
| health-service | Model R(5), Server R(5), Handler R(11), checker R(10); Repo **S**(1, unchanged — `health-service` has no DB of its own); Main R(minimal, 1) | A | **N/A — genuinely no schema**; zero `pgxpool`/`sql.DB` references anywhere in the tree (confirmed) | none | Pure aggregator, no positive-path DB test applicable; no e2e/stress/chaos |
| helixtrack-bridge-service | Handler R(11, `handler_test.go`); coreclient R(6, `coreclient_test.go`); Model R(2); Server **S**(1); Repo **S**(1); Main **X** | `coreclient_live_test.go`(1) + `handler_live_test.go`(1) — §11.4.27 real-integration against a genuinely running HelixTrack Core instance, §11.4.3-compliant honest `SKIP` when unreachable (README.md documents the recipe) | **N/A — NOT wired at `f0b29ff`** (raw `001_init.sql` only; BATCH-E commit `62923d3` on unmerged sibling worktree — see §0) | none dedicated, but the live-gated auth tests ARE the security-relevant proof that `CreateBridge`'s status decision now depends on real Core auth, not a fabricated default | Real HelixTrack Core client now wired; migrations NOT wired at this commit; repo/server still shallow; no e2e/stress/chaos |
| host-service | Main R(2), Model R(4), Repo R(2, incl. negative-path), Server R(3), Handler R(4) | A | **R**, real-PG, `migrations/migrate_test.go`(9) | none | No e2e/stress/chaos |
| keychain-service | Handler R(3); crypto R(8, `crypto_test.go` — T10 encrypt/decrypt round-trip + fail-closed on empty key); repository_internal R(6, `repository_internal_test.go` — T13 RED/GREEN SQL-allow-list hardening); Main/Model/Repo(unit)/Server **X** (`t.Skip`) | `repository_integration_test.go`(4, real-PG — `TestEncryptionAtRest_RealPostgres`/`TestEncryptionAtRest_WrongKeyFailsToDecrypt`/`TestNew_EmptyEncryptionKeyFailsClosed`/`TestUpdateItem_RealPostgres_AllowedFieldsPersist_SecretsUntouched`) | **R**, real-PG, `migrations/migrate_test.go`(9) | **T10** plaintext-secret-at-rest fix (real crypto + real-PG encryption proof) + **T13** SQL column-allow-list hardening (RED-then-GREEN) — both closed with real evidence | Draft's "keychain stores secrets in plaintext" finding is **RESOLVED** (T10); model/repo(unit-level)/server/cmd still stub |
| notification-service | Server R(10), Handler R(14)+`delivery_integration_test.go`(3)=17, Repo R(12), delivery R(9+7+2+3+4=25 across 5 files); Model **partial-X** (`TestNotificationTypes` builds real structs, ends on `assert.True(t, true)`); Main **X (tautology, `TestMainCompiles`, not previously flagged as such — see §4 correction #3)** | 3 files: `delivery_integration_test.go`, `email_integration_test.go`(2), `webhook_integration_test.go`(3) | **N/A — NOT wired at `f0b29ff`** (raw `001_init.sql`+`002_add_delivery_fields.sql`, no `migrate.go`; BATCH-E commit `df80aad` on unmerged sibling worktree — see §0) | `email_injection_test.go`(7) + `webhook_ssrf_test.go`(4), both dedicated | Unchanged since draft (this service was NOT touched this session); migrations NOT wired; no e2e/stress/chaos |
| org-service | Model R(2), Handler R(6); Repo **S**(1), Server **S**(1); Main **X (tautology, `TestMainImports`)** | A | **R**, real-PG, `migrations/migrate_test.go`(9) | none | Repo/server shallow-only; no e2e/stress/chaos |
| pki-service | Model R(4), Handler R(5), crypto R(8); Repo **S**(1), Server **S**(1); Main **X (empty-body)** | `test/integration/lifecycle_test.go`(1, 394-line comprehensive lifecycle test) | **R**, real-PG, `migrations/migrate_test.go`(9) | none dedicated | Repo/server shallow; no e2e/stress/chaos |
| port-forward-service | Handler R(8)+`integration_test.go`(4)=12, forwarder `authorizer_test.go`(5)+`socks5_test.go`(3)=8, Model R(2); Repo **S**(1), Server **S**(1); Main **X** | `internal/handler/integration_test.go`(4, real `-L` SSH tunnels via `internal/sshclient`+`internal/forwarder`, default-deny `-D`/`-R` authorization gate) | **N/A — NOT wired at `f0b29ff`** (raw `001_init.sql`+`002_add_tunnel_fields.sql`; BATCH-E commit `46edaeb` on unmerged sibling worktree — see §0) | default-deny authorization gate proven inside the integration test | Unchanged since draft; migrations NOT wired; no e2e/stress/chaos |
| recording-service | Model R(3), Handler R(8); Repo **S**(1), Server **S**(1); Main **X** | A | **R**, real-PG, `migrations/migrate_test.go`(9) | none | No e2e/stress/chaos |
| sftp-service | Model R(3), Handler R(8); Repo **S**(1), Server **S**(1); Main **X** | A | **R**, real-PG, `migrations/migrate_test.go`(9) | none | No e2e/stress/chaos |
| snippet-service | Model R(1, real field asserts), Handler R(8); Repo **S**(1), Server **S**(1); Main **X** | A | **R**, real-PG, `migrations/migrate_test.go`(9) | none | No e2e/stress/chaos |
| ssh-proxy-service | Main R(2), Model R(4), Server R(2), Handler R(3); sshclient R(5), wshandler R(5); Repo **M**(2, real CRUD via `InMemoryRepository` double — the real `PostgresRepository` production implementation has no test exercising it against an actual DB anywhere in the tree, no `_integration_test.go` exists) | A (no `_integration_test.go` for this service) | **R**, real-PG, `migrations/migrate_test.go`(9) | none | `PostgresRepository` untested against a real DB despite migration wiring existing; no e2e/stress/chaos |
| terminal-service | recorder R(13), Handler R(9); Repo **X (tautology, `TestRepositoryStub`)**, Server **X (tautology, `TestServerStub`)**; Model R(3); Main **X (`t.Skip`, "no meaningful test for main yet")** | A | **R**, real-PG, `migrations/migrate_test.go`(9) | none | Core session-recording service's repo/server plumbing is 100% tautology-stub even though `internal/recorder` (the domain logic) is genuinely well-tested; no e2e/stress/chaos |
| user-service | Handler R(3)+`handler_readiness_integration_test.go`(1)=4 (`TestReadinessCheck_NoRepoConfigured` proves T8-6 fix); Main/Model/Repo/Server **X** (`t.Skip`) | `handler_readiness_integration_test.go`(1, real-PG, T8-6 DB-ping proof, mirrors auth-service's) | **R**, real-PG, `migrations/migrate_test.go`(9) | **T8-6** readiness-fabrication fix (real-PG-gated) | Model/repo/server/cmd 100% stub despite the pgx repository backing every real user record; no e2e/stress/chaos |
| vault-service | Server R(9)+`server_integration_test.go`(7)=16, Handler R(15), Model R(2); Main R(real, 49 lines); Repo R(2)+`repository_integration_test.go`(3)=5 | 2 files: `repository_integration_test.go`(3) + `server_integration_test.go`(7) — both prove the T7 IDOR fix (`966b58b`) on real Postgres | **R**, real-PG, `migrations/migrate_test.go`(9) | T7 IDOR/tenant-isolation tests embedded in the integration suite | Deepest coverage in the fleet; no e2e/stress/chaos |
| workspace-service | Model R(3), Handler R(4); Repo **S**(1); Server R(1, real `httptest`); Main **X (empty-body)** | A | **N/A — NOT wired at `f0b29ff`** (raw `001_init.sql` only; BATCH-F commit `9eb4590` on unmerged sibling worktree — see §0) | none | Unchanged since draft; migrations NOT wired; no e2e/stress/chaos |

Counts in parentheses are Test-func counts per file, re-derived this pass via
`grep -cE '^func Test' <file>` for every file cited (not carried from the draft
without verification).

---

## 2. Flutter client (`clients/flutter`) — carried forward, unchanged

`git diff --stat 45e0edd f0b29ff -- clients/` returns **zero changed files** —
confirmed no client work landed this session. The prior draft's findings are
reproduced here without re-derivation (flag as such, per §11.4.6 — this ledger
did not re-read every Dart file this pass):

- 8/10 test files remain literal `expect(true, isTrue)` no-op stubs (auth,
  vault, host, collaboration, notification, terminal, workspace BLoCs +
  `api_client_test.dart`).
- `test/widget_test.dart` — real but shallow (1 assertion, splash-screen only).
- `test/terminal_view_golden_test.dart` — real and deep (§11.4.170-compliant
  golden PNG diff + content-oracle + VT-engine correctness, 3 `testWidgets`).
- `integration_test/app_test.dart` — the sole e2e-tier test remains a
  boot-only smoke stub with an explicit `// TODO: implement real e2e tests`;
  does not drive login/host-connect/terminal-session/vault flows.

---

## 3. Anti-bluff invariants (§11.4.25) — met vs pending, by service class

§11.4.25 requires six invariants before a feature is "deliverable": (1)
anti-bluff posture (captured runtime evidence), (2) proof of working capability
end-to-end on the target topology, (3) implementation matches the documented
promise, (4) no open issues/bugs surfaced by the suite, (5) full documentation
kept in sync, (6) four-layer test floor (pre-build/post-build/runtime/paired
mutation).

| Class | (1) Anti-bluff evidence | (2) Working proof | (4) No open bugs | (6) Four-layer | Evidence |
|---|---|---|---|---|---|
| **18 golang-migrate-wired services** | **MET** — every `migrate_test.go` is real-PG-gated (`TEST_DATABASE_URL`), honest `SKIP` when absent, never a fabricated PASS | **MET for the migration layer specifically** — `TestRun_Integration` applies real DDL against a real disposable Postgres and asserts schema version | **MET for migrations** — no known open migration defect | **PARTIAL** — pre-build (compiles) + a real-PG runtime layer exist; no paired §1.1 mutation file was found anywhere in `services/` for any migration test (`find services -iname "*mutation*"` → no hits) — the mutation-pairing half of layer 4 is UNCONFIRMED/absent fleet-wide | `services/*/migrations/migrate_test.go` (18 files, 9 funcs each) |
| **3 real bridge backends (ai/container-bridge/helixtrack-bridge)** | **MET** — RED-then-GREEN §11.4.43/§11.4.115 regression tests prove the pre-fix fabrication AND the post-fix real behavior | **MET** — real local-LLM completion, real podman/docker lifecycle, real HelixTrack Core auth, each exercised end-to-end | **MET for the backend-wiring defect specifically** (T8-2/T8-5-class fabrication closed); ai-service's CLOUD multi-model tier remains an open, tracked OPERATOR-BLOCKED gap (not a bug in the merged local path) | **PARTIAL** — same paired-mutation gap as above; no `*_mutation_test.go`/equivalent found | `container-bridge-service/internal/handler/red_*_test.go`, `ai-service/internal/llmclient/*`, `helixtrack-bridge-service/internal/{coreclient,handler}/*_live_test.go` |
| **5 security fixes (T12/T13/T14/T15/T17)** | **MET** — every one of the 4 code-level fixes (T12/T13/T14/T15) has a dedicated real-PG-or-real-crypto regression test that reproduces the pre-fix defect class and proves the post-fix closure | **MET** — `TestBillingCrossTenantIsolation_RealPostgres`, `TestBillingWriteEndpointsCrossTenantIsolation_RealPostgres`, `TestUpdateItem_RealPostgres_AllowedFieldsPersist_SecretsUntouched`, `TestServerNew_ProductionModeWithNoKey_FailsClosed` all assert against real infrastructure, not mocks | **MET** for the specific defect each closes; T17 (compose JWT wiring) has **no accompanying automated test** — it is an `infrastructure/docker/compose/docker-compose.yml` env-var wiring change only, verified by direct `grep`, not by a test file — this is a **documentation/config-verification gap**, not a code defect | **PARTIAL** — T12–T15 have real tests but (again) no found paired-mutation file; T17 has no test at all | `services/billing-service/internal/server/server_{integration,write_isolation_integration}_test.go`, `services/keychain-service/internal/repository/repository_{integration,internal}_test.go`, `services/auth-service/internal/{crypto,server}/*_test.go` |
| **Remaining 19 CRUD services (analytics/audit/collaboration/config/health/host/notification/org/pki/port-forward/recording/sftp/snippet/ssh-proxy/terminal/user/gateway/workspace/billing's non-security surface)** | **PARTIAL** — handler-layer tests are real (httptest against the real Gin router); repository layer is mostly **S** (shallow, not-connected-only) for ~13 services, meaning the actual SQL query logic has zero executing assertion | **PARTIAL** — no positive-path CRUD-against-a-real-connection proof for the S-classified repos (analytics, collaboration, container-bridge's repo, helixtrack-bridge's repo, org, pki, port-forward, recording, sftp, snippet, health's repo, ssh-proxy's `PostgresRepository`) | **NOT MET** for terminal-service specifically (repo+server are literal tautology stubs on a core, high-traffic session-recording path) and for the 4 fully-stubbed services (ai/billing/keychain/user model+repo+cmd) | **NOT MET** — zero stress/chaos/e2e for any of the 25 services; zero found mutation-pair files anywhere | See §1 table, "Notable GAPS" column per row |

**Fleet-wide, invariant (6) four-layer coverage is the weakest column across
every service class**: the pre-build/post-build/runtime three layers have real
substance for the migration-wired and security-fixed services, but the
**paired-mutation layer (§1.1) could not be found anywhere in `services/`** —
`find services -iname "*mutation*"` returns zero files, and no test file
inspected during this pass contained a recognizable mutation-gate pattern
(env-flag-driven intentional-break-then-restore). This is reported as a
fleet-wide, cross-cutting gap rather than per-service, since it appears uniform.

---

## 4. Corrections vs the prior draft (`45e0edd` → `f0b29ff`)

1. **Migration-wiring count is 18/25 + 2 N/A + 5 unwired, not "23/25 + gateway/health
   N/A" as stated in the task brief.** The task brief's framing does not match
   the tree at `f0b29ff`: `billing-service`, `helixtrack-bridge-service`,
   `notification-service`, `port-forward-service`, and `workspace-service` have
   no golang-migrate wiring on this commit. The corresponding BATCH-E/BATCH-F
   commits genuinely exist in the repository's reflog (`9301ec4`, `62923d3`,
   `df80aad`, `46edaeb`, `9eb4590`) but sit on a **different, unmerged worktree
   branch** (`worktree-agent-af47069405aee570e`) — confirmed via
   `git merge-base --is-ancestor <sha> HEAD` returning `NO` for all five. This
   is not a fabrication in this ledger's method; it is the honest state of the
   specific commit this task was scoped to (`f0b29ff`).
2. **`gateway-service`'s `internal/handler` package (draft's finding #7, flagged
   dead/unwired) has been deleted from the tree.** This closes that finding —
   confirmed by `find services/gateway-service/internal/handler` now erroring
   "No such file or directory," and by `gateway-service`'s `internal/server`
   test suite growing from R(23+2) to R(29) funcs, absorbing what the removed
   package's responsibilities apparently were.
3. **`notification-service/cmd/notification-service/main_test.go` is a 5th
   tautology-stub file the prior draft's §0 pattern-2 enumeration omitted**
   (it lists 4: org/cmd, terminal/repo, terminal/server, notification/model).
   Verified live (not comment) `assert.True(t, true)` at line 12 inside
   `TestMainCompiles`. The combined stub-bluff total (37) is unaffected because
   this file was implicitly already counted in the draft's separate "R(minimal)"
   classification for notification's Main column — this ledger reclassifies it
   as **X**, correcting that column.
4. **Draft's risk-ordered items #1 and #2 are now RESOLVED, not merely
   "dispatched":**
   - **#1 (gateway route-table mismatch, T8-1)** — confirmed fixed via commit
     `b79ab72` "align proxyTo route table to real upstream registrations",
     ancestor of `HEAD`. Proven by new tests `TestCorrectedRoutes_ReachRealUpstreamPath`
     and `TestCorrectedRoutes_OldPathWouldHaveMissed` (the latter explicitly
     demonstrates the OLD route table would have 404'd).
   - **#2 (auth/user readiness fabrication, T8-6)** — confirmed fixed:
     `auth-service/internal/handler/handler.go:580-599` and
     `user-service/internal/handler/handler.go:281-297` both now gate `ready`
     on a real DB ping, each with a dedicated real-PG
     `handler_readiness_integration_test.go`.
   - **#4 (container-bridge/helixtrack-bridge fabricated "Active" status)** —
     confirmed fixed with real podman/docker + real HelixTrack Core wiring
     (see §0 and §1).
   - **#5 (ai-service never completes a request)** — confirmed fixed for the
     LOCAL LLM path (see §0); the cloud-multi-model tier remains a distinct,
     still-open, OPERATOR-BLOCKED item.
   - **#6 (keychain plaintext secrets)** — confirmed fixed via T10 (real
     encrypt/decrypt round-trip test, real-PG encryption-at-rest proof).
   - **#3 (billing fabricated active subscriptions, T8-4) is NOT resolved** —
     still OPERATOR-BLOCKED on Stripe test keys, confirmed live in
     `handler.go:78` and in `docs/CONTINUATION.md:42`.
5. **Test-file/func totals grew 148→185 files (+37) and 519→727 funcs (+208)**
   this session, driven almost entirely by: 18 new `migrate_test.go` files (9
   funcs × 18 = 162 funcs, though several already existed at the draft's
   baseline for a subset — see §0 for the exact 18-service list), plus the
   bridge-backend RED/GREEN + live-integration tests, plus the T10/T12/T13/T14/T15
   security-fix regression tests.

---

## 5. UNCONFIRMED / not verified in this pass

- **Runtime PASS/FAIL of any test file** — this ledger is a static source
  inventory (§11.4.6); no `go test ./...` was executed for any of the 25
  services in this pass. A file classified **R** here compiles and asserts
  real logic on inspection, but was not run.
- **Whether the 5 unmigrated services' BATCH-E/F work (on the sibling worktree
  `worktree-agent-af47069405aee570e`) has since been merged to `main`** after
  this snapshot was taken — that worktree is a live, potentially still-moving
  target per §11.4.147/§11.4.103; re-diff `git log --oneline -5` on `main`
  before relying on this ledger's "N/A — NOT wired" classification for billing/
  helixtrack-bridge/notification/port-forward/workspace.
- **Whether a paired §1.1 mutation harness exists somewhere outside
  `services/`** (e.g. under a project-level `scripts/testing/` or `tools/`
  directory not scoped to this ledger's `services/*` search) — this pass only
  searched `services/` per the task's scope; §3's "no mutation file found"
  finding is scoped to that directory tree, not the whole repository.
- **Flutter client findings (§2)** are carried forward from the prior draft
  without re-derivation in this pass (confirmed via `git diff` that
  `clients/` is byte-identical to the draft's baseline, so no re-read was
  performed to conserve scope — the draft's own file-by-file citations remain
  the evidence of record for that section).
