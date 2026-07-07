# S1 — golang-migrate startup wiring — captured evidence

Queue item #2 (GAP-02): real golang-migrate startup wiring for
auth-service, user-service, org-service.

All commands below were run with `GOMAXPROCS=2` and `-p 2` per the
host pid-constraint. `GOWORK=off` is required for every `go` command in
this worktree because the outer main-repo `go.work` (an ancestor
directory of this worktree's path) declares `use ./services/...`
relative to the MAIN repo root, which would otherwise misroute module
resolution into the sibling main-repo checkout instead of this
worktree's own copies.

## Summary / conclusion

- Implemented a real `migrations` package per service (`services/{auth,user,org}-service/migrations/migrate.go`)
  embedding each service's `.up.sql`/`.down.sql` via `embed.FS` and applying them at
  process startup with `github.com/golang-migrate/migrate/v4` v4.19.1, the
  `database/pgx/v5` driver, and the `source/iofs` source driver. Idempotent:
  a database already at the latest embedded version is a no-op
  (`migrate.ErrNoChange`), never an error. A "dirty" (partially-applied)
  schema is treated as a hard failure — never silently served against.
- Renamed each service's `migrations/001_init.sql` to `001_init.up.sql`
  (`git mv`, history preserved) and added a matching `001_init.down.sql`,
  since golang-migrate's `iofs`/`file` source drivers require the
  `{version}_{name}.{up|down}.{ext}` naming convention (confirmed against
  the current `source/parse.go` regex on GitHub — see section 0 below).
- Wired the runner into each service's real startup path:
  auth-service and org-service (`internal/server/server.go New()`, before
  `pgxpool.New`, non-fatal warn-and-fall-back-to-in-memory on failure,
  matching each service's pre-existing DB-trouble degrade philosophy);
  user-service (`cmd/user-service/main.go`, before `pgxpool.New`,
  `log.Fatalf` on failure, matching user-service's pre-existing fail-fast
  philosophy for DB connectivity trouble).
- Because all 25 helix_terminator services are configured (see
  `infrastructure/docker/compose/docker-compose.yml`) to share ONE
  PostgreSQL database, each service's migrations table is pinned to a
  distinct name (`x-migrations-table=<service>_schema_migrations`) via the
  DSN so services never clobber each other's migration version
  bookkeeping. **This is load-bearing, not decorative** — see the REAL
  FINDING in section 6 below.
- Real, captured, anti-bluff proof against actual rootless-podman
  PostgreSQL 17.2 (never mocked): fresh empty DB → real tables created by
  the real startup migration path → real second invocation proving
  idempotency (`-count=1` used deliberately to defeat `go test`'s result
  cache, which does NOT track `os.Getenv` reads as a cache key and would
  otherwise silently replay a stale PASS without re-touching the
  database — a genuine anti-bluff catch made and corrected during this
  run, see section 5b).
- `go build`, `go vet`, and the full existing test suite are clean for
  all three services (see sections 9). New unit tests
  (`toPGX5DSN` scheme rewriting + migrations-table pinning, empty-DSN
  guard) plus a real-infra integration test (`TestRun_Integration`, gated
  on `TEST_DATABASE_URL`, SKIPs honestly per §11.4.3 when no DB is
  available rather than faking a PASS) were added per service.
- **REAL FINDING requiring an operator/architecture decision (see
  section 6):** auth-service and user-service both migrate a table
  literally named `users` (different column sets — auth-service's is a
  credentials/MFA store, user-service's is a profile store) and both
  create an index literally named `idx_users_email`. Run against the
  SAME database (the docker-compose.yml topology), user-service's
  migration collides with auth-service's already-applied schema and
  fails with `relation "idx_users_email" already exists`. This is a
  genuine, pre-existing defect (never exercised before because
  migrations were never actually run at startup) — NOT something
  introduced by this change, and NOT something safely fixable by a
  scoped subagent without an architecture decision (separate
  databases/schemas per service vs. renaming/merging the colliding
  `users` tables across two services touches call sites beyond this
  queue item's file-scope lane). The dirty-schema guard in `Run()`
  correctly refused to leave a corrupted mid-migration state served —
  confirmed for real (`dirty=true` captured via psql, not simulated).

## 0. Research: golang-migrate v4.19.1 API (verified against GitHub source, not memory)

- Root package: `github.com/golang-migrate/migrate/v4`
- Postgres/pgx-v5 driver: `github.com/golang-migrate/migrate/v4/database/pgx/v5`
  (blank-imported for its `init()` registration of the `"pgx5"` scheme;
  confirmed via `database/pgx/v5/pgx.go` — `database.Register("pgx5", &db)`,
  `Open()` internally does `sql.Open("pgx/v5", ...)` via
  `github.com/jackc/pgx/v5/stdlib`).
- Embedded-FS source driver: `github.com/golang-migrate/migrate/v4/source/iofs`
  (`iofs.New(fs, "path")`, confirmed via `source/iofs/example_test.go`).
- Filename convention confirmed via `source/parse.go`:
  `^([0-9]+)_(.*)\.(down|up)\.(.*)$` — REQUIRES the `.up.`/`.down.` infix;
  the pre-existing `001_init.sql` files did NOT match this pattern.
- Latest release at implementation time: v4.19.1 (published 2025-11-29,
  confirmed via `https://api.github.com/repos/golang-migrate/migrate/releases/latest`).

## 1. Boot rootless podman PostgreSQL 17.2 (fresh, empty DB)

```
7d74a286d465f13cc2a0569825a29a03d079f1425e0bf9b1e7ac39eb45043382
```

## 2. Confirm DB is fresh/empty before any migration

```
Did not find any relations.
```

## 3. auth-service: run TestRun_Integration against the fresh DB (first run = applies migration)

```
=== RUN   TestRun_Integration
    migrate_test.go:109: migrations: applied pending migrations - schema now at version 1
    migrate_test.go:109: migrations: schema already up to date at version 1 (no-op)
--- PASS: TestRun_Integration (0.03s)
PASS
ok  	github.com/helixdevelopment/auth-service/migrations	0.031s
EXIT=0
```

## 4. psql \dt after auth-service migration — proves real tables created

```
                     List of relations
 Schema |              Name              | Type  |  Owner   
--------+--------------------------------+-------+----------
 public | auth_service_schema_migrations | table | postgres
 public | user_sessions                  | table | postgres
 public | users                          | table | postgres
(3 rows)


 version | dirty 
---------+-------
       1 | f
(1 row)

```

## 5. auth-service: SECOND, fully independent process invocation (idempotency across process restarts)

```
=== RUN   TestRun_Integration
    migrate_test.go:109: migrations: applied pending migrations - schema now at version 1
    migrate_test.go:109: migrations: schema already up to date at version 1 (no-op)
--- PASS: TestRun_Integration (0.03s)
PASS
ok  	github.com/helixdevelopment/auth-service/migrations	(cached)
EXIT=0
```

## 5b. CORRECTION: prior run said "(cached)" - Go test result caching does not
track os.Getenv("TEST_DATABASE_URL") as a cache key, so that "PASS" may have
been replayed output from the earlier run without touching the DB again. Per
§11.4.1/§11.4.50 anti-bluff, re-running forced with -count=1 for honest fresh evidence:

```
=== RUN   TestRun_Integration
    migrate_test.go:109: migrations: schema already up to date at version 1 (no-op)
    migrate_test.go:109: migrations: schema already up to date at version 1 (no-op)
--- PASS: TestRun_Integration (0.01s)
PASS
ok  	github.com/helixdevelopment/auth-service/migrations	0.018s
EXIT=0
```

## 6. user-service: run against the SAME shared DB (proves per-service
migrations-table isolation - all 25 services share one "helixterminator"-style
database per infrastructure/docker/compose/docker-compose.yml, so distinct
`x-migrations-table` names per service are load-bearing, not cosmetic)

```
=== RUN   TestRun_Integration
    migrate_test.go:89: Run() first invocation failed: migrations: apply: migration failed: relation "idx_users_email" already exists (column 0) in line 1: CREATE TABLE IF NOT EXISTS users (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            email VARCHAR(255) UNIQUE NOT NULL,
            display_name VARCHAR(255) NOT NULL,
            avatar_url TEXT,
            role VARCHAR(50) NOT NULL DEFAULT 'user',
            permissions TEXT[] NOT NULL DEFAULT '{}',
            org_id UUID,
            email_verified BOOLEAN NOT NULL DEFAULT FALSE,
            last_login_at TIMESTAMPTZ,
            created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
            updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
            deleted_at TIMESTAMPTZ
        );
        
        CREATE INDEX idx_users_email ON users(email);
        CREATE INDEX idx_users_org_id ON users(org_id);
        CREATE INDEX idx_users_role ON users(role);
        CREATE INDEX idx_users_deleted_at ON users(deleted_at) WHERE deleted_at IS NULL;
        
        CREATE TABLE IF NOT EXISTS user_profiles (
            user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
            bio TEXT,
            timezone VARCHAR(50) DEFAULT 'UTC',
            locale VARCHAR(10) DEFAULT 'en',
            preferences JSONB DEFAULT '{}',
            ssh_public_key TEXT,
            github_id VARCHAR(255),
            gitlab_id VARCHAR(255),
            updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
        );
        
        CREATE INDEX idx_user_profiles_github_id ON user_profiles(github_id);
        CREATE INDEX idx_user_profiles_gitlab_id ON user_profiles(gitlab_id);
        
        CREATE OR REPLACE FUNCTION update_updated_at_column()
        RETURNS TRIGGER AS $$
        BEGIN
            NEW.updated_at = NOW();
            RETURN NEW;
        END;
        $$ LANGUAGE plpgsql;
        
        CREATE TRIGGER users_updated_at
            BEFORE UPDATE ON users
            FOR EACH ROW
            EXECUTE FUNCTION update_updated_at_column();
        
        CREATE TRIGGER user_profiles_updated_at
            BEFORE UPDATE ON user_profiles
            FOR EACH ROW
            EXECUTE FUNCTION update_updated_at_column();
         (details: ERROR: relation "idx_users_email" already exists (SQLSTATE 42P07))
--- FAIL: TestRun_Integration (0.01s)
FAIL
FAIL	github.com/helixdevelopment/user-service/migrations	0.017s
FAIL
EXIT=1
```

### REAL FINDING: cross-service schema collision in the shared DB

auth-service and user-service both migrate a table literally named `users`
(different column sets - auth-service: credentials/MFA store; user-service:
profile store) and both create an index literally named `idx_users_email`.
infrastructure/docker/compose/docker-compose.yml points every service at the
SAME `helixterminator` database, so this is a real production-topology risk,
not a test artifact. This pre-dates this change (migrations were never run
for real before), so it is a newly-surfaced pre-existing defect, not a
regression introduced here. Fixing it (e.g. renaming tables/schemas) is an
architectural decision beyond this task scope (queue #2 = wire up the
runner) and out of my file-scope lane for other services. Flagged honestly
below rather than silently worked around.

```
--- state of shared DB after the collision (user_service migration failed partway) ---
                     List of relations
 Schema |              Name              | Type  |  Owner   
--------+--------------------------------+-------+----------
 public | auth_service_schema_migrations | table | postgres
 public | user_service_schema_migrations | table | postgres
 public | user_sessions                  | table | postgres
 public | users                          | table | postgres
(4 rows)


 version | dirty 
---------+-------
       1 | t
(1 row)

```
migrate-it-shared
stopped shared container (poisoned by the collision above); proceeding with fully isolated per-service containers below

## 7. Isolated per-service proof runs (fresh empty DB each)

### auth-service (port 15501)

```
66c2c69be63bac421461dba761a1e033e75cc5dc5c826fc12a544e722c150d66
```
### user-service (port 15502)

```
2ed1238e43dcff6460398015e6838f881c5b8128f902ada83ab2a3c7b00ebeae
```
### org-service (port 15503)

```
e32bcebce5ae1ca0de09fcaf7305b43e545c7d2949ce75075daebc6222b7dc92
```

#### auth-service — fresh DB, before-state \dt
```
Did not find any relations.
```

#### user-service — fresh DB, before-state \dt
```
Did not find any relations.
```

#### org-service — fresh DB, before-state \dt
```
Did not find any relations.
```

#### auth-service — TestRun_Integration (real first-apply + real idempotent second-run, -count=1 to defeat go test caching)
```
=== RUN   TestRun_Integration
    migrate_test.go:109: migrations: applied pending migrations - schema now at version 1
    migrate_test.go:109: migrations: schema already up to date at version 1 (no-op)
--- PASS: TestRun_Integration (0.03s)
PASS
ok  	github.com/helixdevelopment/auth-service/migrations	0.034s
EXIT=0
```

#### user-service — TestRun_Integration (real first-apply + real idempotent second-run, -count=1 to defeat go test caching)
```
=== RUN   TestRun_Integration
    migrate_test.go:109: migrations: applied pending migrations - schema now at version 1
    migrate_test.go:109: migrations: schema already up to date at version 1 (no-op)
--- PASS: TestRun_Integration (0.03s)
PASS
ok  	github.com/helixdevelopment/user-service/migrations	0.032s
EXIT=0
```

#### org-service — TestRun_Integration (real first-apply + real idempotent second-run, -count=1 to defeat go test caching)
```
=== RUN   TestRun_Integration
    migrate_test.go:109: migrations: applied pending migrations - schema now at version 1
    migrate_test.go:109: migrations: schema already up to date at version 1 (no-op)
--- PASS: TestRun_Integration (0.03s)
PASS
ok  	github.com/helixdevelopment/org-service/migrations	0.033s
EXIT=0
```

#### auth-service — psql \dt after migration (real tables proof)
```
                     List of relations
 Schema |              Name              | Type  |  Owner   
--------+--------------------------------+-------+----------
 public | auth_service_schema_migrations | table | postgres
 public | user_sessions                  | table | postgres
 public | users                          | table | postgres
(3 rows)


 version | dirty 
---------+-------
       1 | f
(1 row)

```

#### user-service — psql \dt after migration (real tables proof)
```
                     List of relations
 Schema |              Name              | Type  |  Owner   
--------+--------------------------------+-------+----------
 public | user_profiles                  | table | postgres
 public | user_service_schema_migrations | table | postgres
 public | users                          | table | postgres
(3 rows)


 version | dirty 
---------+-------
       1 | f
(1 row)

```

#### org-service — psql \dt after migration (real tables proof)
```
                     List of relations
 Schema |             Name              | Type  |  Owner   
--------+-------------------------------+-------+----------
 public | memberships                   | table | postgres
 public | org_service_schema_migrations | table | postgres
 public | organizations                 | table | postgres
 public | teams                         | table | postgres
(4 rows)


 version | dirty 
---------+-------
       1 | f
(1 row)

```

## 8. Cleanup — stop all podman containers started for this evidence run

```
migrate-it-auth-service
migrate-it-user-service
migrate-it-org-service
```

CONTAINER ID  IMAGE       COMMAND     CREATED     STATUS      PORTS       NAMES

## 9. Final sweep — go build / go vet / go test (unit, no infra) for all three services

### auth-service
```
-- go build -p 2 ./... --
build EXIT=0
-- go vet -p 2 ./... --
vet EXIT=0
-- go test -p 2 -count=1 ./... --
ok  	github.com/helixdevelopment/auth-service/cmd/auth-service	0.007s
ok  	github.com/helixdevelopment/auth-service/internal/crypto	0.329s
ok  	github.com/helixdevelopment/auth-service/internal/handler	0.007s
ok  	github.com/helixdevelopment/auth-service/internal/model	0.004s
ok  	github.com/helixdevelopment/auth-service/internal/repository	0.003s
ok  	github.com/helixdevelopment/auth-service/internal/server	0.007s
ok  	github.com/helixdevelopment/auth-service/migrations	0.005s
test EXIT=0
```

### user-service
```
-- go build -p 2 ./... --
build EXIT=0
-- go vet -p 2 ./... --
vet EXIT=0
-- go test -p 2 -count=1 ./... --
ok  	github.com/helixdevelopment/user-service/cmd/user-service	0.007s
ok  	github.com/helixdevelopment/user-service/internal/handler	0.007s
ok  	github.com/helixdevelopment/user-service/internal/model	0.002s
ok  	github.com/helixdevelopment/user-service/internal/repository	0.004s
ok  	github.com/helixdevelopment/user-service/internal/server	0.006s
ok  	github.com/helixdevelopment/user-service/migrations	0.004s
test EXIT=0
```

### org-service
```
-- go build -p 2 ./... --
build EXIT=0
-- go vet -p 2 ./... --
vet EXIT=0
-- go test -p 2 -count=1 ./... --
ok  	github.com/helixdevelopment/org-service/cmd/org-service	0.019s
ok  	github.com/helixdevelopment/org-service/internal/handler	0.010s
ok  	github.com/helixdevelopment/org-service/internal/model	0.003s
ok  	github.com/helixdevelopment/org-service/internal/repository	0.004s
ok  	github.com/helixdevelopment/org-service/internal/server	0.008s
ok  	github.com/helixdevelopment/org-service/migrations	0.004s
test EXIT=0
```
