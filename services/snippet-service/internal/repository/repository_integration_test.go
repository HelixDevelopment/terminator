//go:build integration

package repository_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/helixdevelopment/snippet-service/internal/model"
	"github.com/helixdevelopment/snippet-service/internal/repository"
	"github.com/helixdevelopment/snippet-service/internal/testutil"
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

func TestCreateAndGetSnippet_RoundTrips(t *testing.T) {
	repo := newRealRepo(t)
	ctx := context.Background()

	snippet := &model.Snippet{
		ID:        uuid.New(),
		CreatedBy: uuid.New(),
		Name:      "test-snippet",
		Content:   "echo hello",
		Language:  "bash",
		IsPublic:  true,
	}
	if err := repo.CreateSnippet(ctx, snippet); err != nil {
		t.Fatalf("CreateSnippet failed: %v", err)
	}

	got, err := repo.GetSnippetByID(ctx, snippet.ID)
	if err != nil {
		t.Fatalf("GetSnippetByID failed: %v", err)
	}
	if got.ID != snippet.ID || got.Content != snippet.Content {
		t.Fatalf("snippet round-trip mismatch: got %+v", got)
	}
}
