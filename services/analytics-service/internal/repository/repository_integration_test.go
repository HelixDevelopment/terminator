//go:build integration

package repository_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/helixdevelopment/analytics-service/internal/model"
	"github.com/helixdevelopment/analytics-service/internal/repository"
	"github.com/helixdevelopment/analytics-service/internal/testutil"
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

func TestCreateAndGetEvent_RoundTrips(t *testing.T) {
	repo := newRealRepo(t)
	ctx := context.Background()

	event := &model.AnalyticsEvent{
		ID:        uuid.New(),
		EventType: "test_event",
		Payload:   []byte(`{"key":"value"}`),
	}
	if err := repo.CreateEvent(ctx, event); err != nil {
		t.Fatalf("CreateEvent failed: %v", err)
	}

	got, err := repo.GetEventByID(ctx, event.ID)
	if err != nil {
		t.Fatalf("GetEventByID failed: %v", err)
	}
	if got.ID != event.ID || got.EventType != event.EventType {
		t.Fatalf("event round-trip mismatch: got %+v", got)
	}
}

func TestListEvents_HonorsLimit(t *testing.T) {
	repo := newRealRepo(t)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		event := &model.AnalyticsEvent{
			ID:        uuid.New(),
			EventType: "list_event",
			Payload:   []byte(`{}`),
		}
		if err := repo.CreateEvent(ctx, event); err != nil {
			t.Fatalf("CreateEvent failed: %v", err)
		}
	}

	events, total, err := repo.ListEvents(ctx, nil, "list_event", 2, 0)
	if err != nil {
		t.Fatalf("ListEvents failed: %v", err)
	}
	if total < 5 {
		t.Fatalf("expected total >= 5, got %d", total)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events with limit=2, got %d", len(events))
	}
}
