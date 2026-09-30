package service

import (
	"context"
	"testing"

	"github.com/davidlima/openspec-studio/backend/internal/domain"
)

func TestContextServiceReviewsAndDebts(t *testing.T) {
	ctx := context.Background()
	proj := newTestProject(t, 1, "demo")
	lister := &stubProjectLister{projects: []domain.Project{proj}}
	svc := NewContextService(lister)

	link := domain.ArtifactRef{Kind: "spec", Capability: "auth"}
	review, err := svc.CreateReview(ctx, 1, "Auth review", "Looks solid overall.", link)
	if err != nil {
		t.Fatalf("CreateReview: %v", err)
	}
	if review.Link.Capability != "auth" {
		t.Fatalf("expected link to round-trip, got %+v", review.Link)
	}

	reviews, err := svc.ListReviews(ctx, 1)
	if err != nil {
		t.Fatalf("ListReviews: %v", err)
	}
	if len(reviews) != 1 || reviews[0].ID != review.ID || reviews[0].Body != "Looks solid overall." {
		t.Fatalf("unexpected reviews: %+v", reviews)
	}

	debt, err := svc.CreateDebt(ctx, 1, "Refactor auth", "Needs cleanup.", link)
	if err != nil {
		t.Fatalf("CreateDebt: %v", err)
	}
	if debt.Status != domain.DebtStatusOpen {
		t.Fatalf("expected new debt to be open, got %q", debt.Status)
	}

	updated, err := svc.UpdateDebtStatus(ctx, 1, debt.ID, domain.DebtStatusResolved)
	if err != nil {
		t.Fatalf("UpdateDebtStatus: %v", err)
	}
	if updated.Status != domain.DebtStatusResolved {
		t.Fatalf("expected resolved status, got %q", updated.Status)
	}
	if updated.Link.Capability != "auth" {
		t.Fatalf("expected link preserved after status update, got %+v", updated.Link)
	}

	debts, err := svc.ListDebts(ctx, 1)
	if err != nil {
		t.Fatalf("ListDebts: %v", err)
	}
	if len(debts) != 1 || debts[0].Status != domain.DebtStatusResolved {
		t.Fatalf("unexpected debts: %+v", debts)
	}

	if err := svc.DeleteDebt(ctx, 1, debt.ID); err != nil {
		t.Fatalf("DeleteDebt: %v", err)
	}
	if debts, err := svc.ListDebts(ctx, 1); err != nil || len(debts) != 0 {
		t.Fatalf("expected debt deleted, got %+v (err=%v)", debts, err)
	}
}
