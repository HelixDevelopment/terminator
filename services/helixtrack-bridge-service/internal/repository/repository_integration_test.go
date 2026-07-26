//go:build integration

package repository_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/helixdevelopment/helixtrack-bridge-service/internal/model"
	"github.com/helixdevelopment/helixtrack-bridge-service/internal/repository"
	"github.com/helixdevelopment/helixtrack-bridge-service/internal/testutil"
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

func TestCreateAndGetBridge_RoundTrips(t *testing.T) {
	repo := newRealRepo(t)
	ctx := context.Background()

	bridge := &model.HelixTrackBridge{
		ID:            uuid.New(),
		IntegrationID: "test-integration",
		OrgID:         uuid.New(),
		Name:          "test-bridge",
		Status:        model.HelixTrackBridgeStatusActive,
		Config:        []byte(`{"key":"value"}`),
	}
	if err := repo.CreateBridge(ctx, bridge); err != nil {
		t.Fatalf("CreateBridge failed: %v", err)
	}

	got, err := repo.GetBridgeByID(ctx, bridge.ID)
	if err != nil {
		t.Fatalf("GetBridgeByID failed: %v", err)
	}
	if got.ID != bridge.ID || got.Name != bridge.Name {
		t.Fatalf("bridge round-trip mismatch: got %+v", got)
	}
}
