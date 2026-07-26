package model_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/helixdevelopment/notification-service/internal/model"
)

func TestNotificationCreation(t *testing.T) {
	now := time.Now()
	userID := uuid.New()

	n := model.Notification{
		ID:      uuid.New(),
		UserID:  userID,
		Type:    "info",
		Title:   "Test",
		Message: "Test message",
		Channel: "in_app",
		Status:  "pending",
		CreatedAt: now,
		UpdatedAt: now,
	}

	if n.UserID != userID {
		t.Fatalf("expected userID %s, got %s", userID, n.UserID)
	}
	if n.Type != "info" {
		t.Fatalf("expected type info, got %s", n.Type)
	}
}

func TestNotificationPreferenceValidation(t *testing.T) {
	userID := uuid.New()
	pref := model.NotificationPreference{
		UserID:  userID,
		Channel: "email",
		Enabled: true,
		Types:   []string{"info"},
	}

	if !pref.Enabled {
		t.Fatal("expected preference to be enabled")
	}
	if len(pref.Types) != 1 || pref.Types[0] != "info" {
		t.Fatalf("unexpected types: %v", pref.Types)
	}
}
