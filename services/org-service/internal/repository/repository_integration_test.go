//go:build integration

package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/helixdevelopment/org-service/internal/model"
	"github.com/helixdevelopment/org-service/internal/repository"
	"github.com/helixdevelopment/org-service/internal/testutil"
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

func TestCreateAndGetOrg_RoundTrips(t *testing.T) {
	repo := newRealRepo(t)
	ctx := context.Background()
	now := time.Now().UTC()

	org := &model.Organization{
		ID:        uuid.New(),
		Name:      "Test Org",
		Slug:      "test-org",
		OwnerID:   uuid.New(),
		Plan:      model.PlanFree,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := repo.CreateOrg(ctx, org); err != nil {
		t.Fatalf("CreateOrg failed: %v", err)
	}

	got, err := repo.GetOrgByID(ctx, org.ID)
	if err != nil {
		t.Fatalf("GetOrgByID failed: %v", err)
	}
	if got.ID != org.ID || got.Slug != org.Slug {
		t.Fatalf("org round-trip mismatch: got %+v", got)
	}
}
