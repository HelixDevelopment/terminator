//go:build integration

package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/helixdevelopment/ssh-proxy-service/internal/model"
	"github.com/helixdevelopment/ssh-proxy-service/internal/repository"
	"github.com/helixdevelopment/ssh-proxy-service/internal/testutil"
)

func newRealRepo(t *testing.T) repository.Repository {
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
	return repository.NewPostgresRepository(pool)
}

func TestCreateAndGetSSHSession_RoundTrips(t *testing.T) {
	repo := newRealRepo(t)
	ctx := context.Background()
	now := time.Now().UTC()

	session := &model.SSHSession{
		ID:               uuid.New(),
		UserID:           uuid.New(),
		HostID:           uuid.New(),
		HostAddress:      "127.0.0.1",
		Username:         "test",
		AuthType:         "password",
		ConnectionStatus: model.StatusConnected,
		CreatedAt:        now,
	}
	if err := repo.CreateSession(ctx, session); err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	got, err := repo.GetSessionByID(ctx, session.ID)
	if err != nil {
		t.Fatalf("GetSessionByID failed: %v", err)
	}
	if got.ID != session.ID || got.Username != session.Username {
		t.Fatalf("session round-trip mismatch: got %+v", got)
	}
}
