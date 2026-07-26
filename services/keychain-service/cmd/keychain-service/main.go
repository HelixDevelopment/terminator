package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/helixdevelopment/keychain-service/internal/repository"
	"github.com/helixdevelopment/keychain-service/internal/server"
	"github.com/helixdevelopment/keychain-service/migrations"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8086"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5432/helixterminator?sslmode=disable"
	}

	// Encryption-at-rest key for private_key + passphrase (§11.4.10 / T10).
	// Fail-closed: the service MUST NOT start without it — there is no
	// silent plaintext fallback.
	encKey := os.Getenv("KEYCHAIN_ENCRYPTION_KEY")
	if encKey == "" {
		log.Printf("KEYCHAIN_ENCRYPTION_KEY environment variable is required")
		os.Exit(1)
	}

	// Apply pending schema migrations before opening the steady-state pool.
	// keychain-service already fails fast on DB connectivity trouble (see
	// pgxpool.New below), so a migration failure (including a dirty schema
	// state) is fatal here too - never serve against an unmigrated schema.
	// Scope note: this touches SCHEMA INIT ONLY - it never reads, writes,
	// or reasons about the private_key/passphrase ciphertext handled by
	// internal/repository + internal/crypto (T10 encryption-at-rest).
	version, merr := migrations.Run(databaseURL, log.Default())
	if merr != nil {
		log.Printf("failed to apply database migrations: %v", merr)
		os.Exit(1)
	}
	log.Printf("database migrations applied - schema version %d", version)

	// Use the same schema-scoped connection URL the migrator applied
	// (search_path=migrations.Schema) so the steady-state pool's
	// unqualified "keychain_items" queries resolve against the schema
	// migrations.Run just migrated, not the shared database's default
	// "public" schema (schema-per-service, GAP-01).
	poolURL, perr := migrations.ConnectionURL(databaseURL)
	if perr != nil {
		log.Printf("failed to build schema-scoped connection URL: %v", perr)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, poolURL)
	if err != nil {
		log.Printf("failed to connect to database: %v", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Printf("failed to ping database: %v", err)
		pool.Close()
		os.Exit(1)
	}
	log.Println("database connection established")

	repo, err := repository.New(pool, encKey)
	if err != nil {
		log.Printf("failed to initialize repository: %v", err)
		pool.Close()
		os.Exit(1)
	}
	srv := server.New(repo)

	log.Printf("keychain-service starting on port %s", port)
	if err := srv.Run(":" + port); err != nil {
		log.Printf("server failed: %v", err)
		return
	}
}
