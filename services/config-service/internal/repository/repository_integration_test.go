//go:build integration

package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/helixdevelopment/config-service/internal/model"
	"github.com/helixdevelopment/config-service/internal/repository"
	"github.com/helixdevelopment/config-service/internal/testutil"
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

func TestCreateAndGetConfig_RoundTrips(t *testing.T) {
	repo := newRealRepo(t)
	ctx := context.Background()
	now := time.Now().UTC()

	config := &model.Config{
		ID:        uuid.New(),
		Scope:     "global",
		Key:       "test_key",
		Value:     "test_value",
		ValueType: "string",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := repo.CreateConfig(ctx, config); err != nil {
		t.Fatalf("CreateConfig failed: %v", err)
	}

	got, err := repo.GetConfigByID(ctx, config.ID)
	if err != nil {
		t.Fatalf("GetConfigByID failed: %v", err)
	}
	if got.ID != config.ID || got.Value != config.Value {
		t.Fatalf("config round-trip mismatch: got %+v", got)
	}
}

func TestGetConfigByKey_RoundTrips(t *testing.T) {
	repo := newRealRepo(t)
	ctx := context.Background()
	now := time.Now().UTC()

	config := &model.Config{
		ID:        uuid.New(),
		Scope:     "global",
		Key:       "feature_flag",
		Value:     "true",
		ValueType: "bool",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := repo.CreateConfig(ctx, config); err != nil {
		t.Fatalf("CreateConfig failed: %v", err)
	}

	got, err := repo.GetConfigByKey(ctx, "global", nil, "feature_flag")
	if err != nil {
		t.Fatalf("GetConfigByKey failed: %v", err)
	}
	if got.Key != "feature_flag" {
		t.Fatalf("expected key feature_flag, got %s", got.Key)
	}
}
