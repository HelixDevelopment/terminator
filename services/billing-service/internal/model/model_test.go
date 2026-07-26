package model_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/helixdevelopment/billing-service/internal/model"
)

func TestBillingPlanCreation(t *testing.T) {
	now := time.Now()
	plan := model.BillingPlan{
		ID:        uuid.New(),
		Name:      "Pro",
		Slug:      "pro",
		PriceCents: 1000,
		Currency:  "USD",
		Interval:  "month",
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if plan.Slug != "pro" {
		t.Fatalf("expected slug pro, got %s", plan.Slug)
	}
	if !plan.IsActive {
		t.Fatal("expected plan to be active")
	}
}

func TestSubscriptionCreation(t *testing.T) {
	sub := model.Subscription{
		ID:     uuid.New(),
		OrgID:  uuid.New(),
		PlanID: uuid.New(),
		Status: "active",
	}

	if sub.Status != "active" {
		t.Fatalf("expected status active, got %s", sub.Status)
	}
}

func TestCreateSubscriptionRequestValidation(t *testing.T) {
	req := model.CreateSubscriptionRequest{
		PlanID: "invalid-uuid",
	}
	if req.PlanID == "" {
		t.Fatal("plan id should not be empty")
	}
}
