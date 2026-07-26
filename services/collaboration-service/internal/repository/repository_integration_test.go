//go:build integration

package repository_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/helixdevelopment/collaboration-service/internal/model"
	"github.com/helixdevelopment/collaboration-service/internal/repository"
	"github.com/helixdevelopment/collaboration-service/internal/testutil"
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

func TestCreateAndGetSession_RoundTrips(t *testing.T) {
	repo := newRealRepo(t)
	ctx := context.Background()

	session := &model.CollaborationSession{
		ID:        uuid.New(),
		HostID:    uuid.New(),
		CreatedBy: uuid.New(),
		Name:      "test-session",
		Status:    model.SessionStatusActive,
	}
	if err := repo.CreateSession(ctx, session); err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	got, err := repo.GetSessionByID(ctx, session.ID)
	if err != nil {
		t.Fatalf("GetSessionByID failed: %v", err)
	}
	if got.ID != session.ID || got.Name != session.Name {
		t.Fatalf("session round-trip mismatch: got %+v", got)
	}
}

func TestListSessions_HonorsLimit(t *testing.T) {
	repo := newRealRepo(t)
	ctx := context.Background()
	hostID := uuid.New()

	for i := 0; i < 5; i++ {
		session := &model.CollaborationSession{
			ID:        uuid.New(),
			HostID:    hostID,
			CreatedBy: uuid.New(),
			Name:      "test-session",
			Status:    model.SessionStatusActive,
		}
		if err := repo.CreateSession(ctx, session); err != nil {
			t.Fatalf("CreateSession failed: %v", err)
		}
	}

	sessions, total, err := repo.ListSessions(ctx, hostID, uuid.Nil, 2, 0)
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}
	if total < 5 {
		t.Fatalf("expected total >= 5, got %d", total)
	}
	if len(sessions) != 2 {
		t.Fatalf("expected 2 sessions with limit=2, got %d", len(sessions))
	}
}
