//go:build integration

package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/helixdevelopment/host-service/internal/model"
	"github.com/helixdevelopment/host-service/internal/repository"
	"github.com/helixdevelopment/host-service/internal/testutil"
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

func TestCreateAndGetHost_RoundTrips(t *testing.T) {
	repo := newRealRepo(t)
	ctx := context.Background()
	now := time.Now().UTC()

	host := &model.Host{
		ID:               uuid.New(),
		UserID:           uuid.New(),
		OrgID:            uuid.New(),
		Name:             "test-host",
		Hostname:         "127.0.0.1",
		Port:             22,
		Username:         "root",
		AuthType:         model.AuthTypePassword,
		ConnectionStatus: model.StatusUnknown,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := repo.CreateHost(ctx, host); err != nil {
		t.Fatalf("CreateHost failed: %v", err)
	}

	got, err := repo.GetHostByID(ctx, host.ID)
	if err != nil {
		t.Fatalf("GetHostByID failed: %v", err)
	}
	if got.ID != host.ID || got.Hostname != host.Hostname {
		t.Fatalf("host round-trip mismatch: got %+v", got)
	}
}
