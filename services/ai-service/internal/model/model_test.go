package model_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/helixdevelopment/ai-service/internal/model"
)

func TestAIRequestCreation(t *testing.T) {
	userID := uuid.New()
	now := time.Now()

	req := model.AIRequest{
		ID:          uuid.New(),
		UserID:      userID,
		Prompt:      "What is the capital of France?",
		Model:       "test-model",
		MaxTokens:   100,
		Temperature: 0.5,
		Status:      "pending",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if req.UserID != userID {
		t.Fatalf("expected userID %s, got %s", userID, req.UserID)
	}
	if req.Status != "pending" {
		t.Fatalf("expected status pending, got %s", req.Status)
	}
}

func TestCreateAIRequestValidation(t *testing.T) {
	req := model.CreateAIRequest{
		Prompt:      "Hello",
		Model:       "test-model",
		MaxTokens:   100,
		Temperature: 0.7,
	}

	if req.Prompt != "Hello" {
		t.Fatalf("unexpected prompt: %s", req.Prompt)
	}
	if req.Temperature < 0 || req.Temperature > 2 {
		t.Fatalf("temperature out of range: %f", req.Temperature)
	}
}
