//go:build integration

package repository_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/helixdevelopment/ai-service/internal/model"
	"github.com/helixdevelopment/ai-service/internal/repository"
	"github.com/helixdevelopment/ai-service/internal/testutil"
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

func TestCreateAndGetRequest_RoundTrips(t *testing.T) {
	repo := newRealRepo(t)
	ctx := context.Background()

	req := &model.AIRequest{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Prompt:    "What is Go?",
		Model:     "test-model",
		Status:    "completed",
		Response:  "Go is a programming language.",
		TokensUsed: 10,
	}
	if err := repo.CreateRequest(ctx, req); err != nil {
		t.Fatalf("CreateRequest failed: %v", err)
	}

	got, err := repo.GetRequestByID(ctx, req.ID)
	if err != nil {
		t.Fatalf("GetRequestByID failed: %v", err)
	}
	if got.ID != req.ID || got.Prompt != req.Prompt {
		t.Fatalf("request round-trip mismatch: got %+v", got)
	}
}
