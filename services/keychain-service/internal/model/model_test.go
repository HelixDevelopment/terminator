package model_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/helixdevelopment/keychain-service/internal/model"
)

func TestKeychainItemCreation(t *testing.T) {
	now := time.Now()
	userID := uuid.New()

	item := model.KeychainItem{
		ID:         uuid.New(),
		UserID:     userID,
		Name:       "test-key",
		Type:       model.KeyTypeSSH,
		PublicKey:  "ssh-rsa AAAAB3NzaC1 test",
		PrivateKey: "private-key-material",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if item.Type != model.KeyTypeSSH {
		t.Fatalf("expected type ssh, got %s", item.Type)
	}
	if item.UserID != userID {
		t.Fatalf("expected userID %s, got %s", userID, item.UserID)
	}
}

func TestValidKeyTypes(t *testing.T) {
	types := model.ValidKeyTypes()
	if len(types) != 5 {
		t.Fatalf("expected 5 valid key types, got %d", len(types))
	}
}
