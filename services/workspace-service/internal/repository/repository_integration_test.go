//go:build integration

package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/helixdevelopment/workspace-service/internal/model"
	"github.com/helixdevelopment/workspace-service/internal/repository"
	"github.com/helixdevelopment/workspace-service/internal/testutil"
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

func TestCreateAndGetWorkspace_RoundTrips(t *testing.T) {
	repo := newRealRepo(t)
	ctx := context.Background()
	now := time.Now().UTC()

	workspace := &model.Workspace{
		ID:        uuid.New(),
		OrgID:     uuid.New(),
		UserID:    uuid.New(),
		Name:      "test-workspace",
		Color:     "#000000",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := repo.CreateWorkspace(ctx, workspace); err != nil {
		t.Fatalf("CreateWorkspace failed: %v", err)
	}

	got, err := repo.GetWorkspaceByID(ctx, workspace.ID)
	if err != nil {
		t.Fatalf("GetWorkspaceByID failed: %v", err)
	}
	if got.ID != workspace.ID || got.Name != workspace.Name {
		t.Fatalf("workspace round-trip mismatch: got %+v", got)
	}
}
