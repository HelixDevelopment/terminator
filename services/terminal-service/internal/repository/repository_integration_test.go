//go:build integration

package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/helixdevelopment/terminal-service/internal/model"
	"github.com/helixdevelopment/terminal-service/internal/repository"
	"github.com/helixdevelopment/terminal-service/internal/testutil"
)

func newRealRepo(t *testing.T) *repository.Repository {
	t.Helper()
	poolURL, available := testutil.StartTestPostgres(t)
	if !available {
		t.Skip("SKIP: podman not available — cannot run repository tests against a real database")
	}
	pool, err := pgxpool.New(context.Background(), poolURL)
	if err != nil {
		t.Fatalf("pgxpool.New failed: %v", err)
	}
	t.Cleanup(pool.Close)
	return repository.New(pool)
}

func TestCreateAndGetTerminalSession_RoundTrips(t *testing.T) {
	repo := newRealRepo(t)
	ctx := context.Background()
	now := time.Now().UTC()

	session := &model.TerminalSession{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		HostID:    uuid.New(),
		Status:    model.TerminalStatusActive,
		Cols:      80,
		Rows:      24,
		ShellType: "bash",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := repo.CreateSession(ctx, session); err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	got, err := repo.GetSessionByID(ctx, session.ID)
	if err != nil {
		t.Fatalf("GetSessionByID failed: %v", err)
	}
	if got.ID != session.ID || got.ShellType != session.ShellType {
		t.Fatalf("session round-trip mismatch: got %+v", got)
	}
}
