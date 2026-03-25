package db_test

import (
	"context"
	"testing"

	"github.com/cankledankle/home-planner/internal/db"
)

// TestPlan_CRUDCycle covers create, read, update, soft-delete, and restore (integration).
func TestPlan_CRUDCycle(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	// Seed a user for the created_by / updated_by fields.
	user, err := store.CreateUser(ctx, "Test User", "testplan@example.com", "password123", "editor")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	t.Cleanup(func() { store.DeleteUser(ctx, user.ID) })

	// Create
	beds := 3
	baths := 2
	heatedSF := 1800
	plan, err := store.CreatePlan(ctx, db.CreatePlanInput{
		Name:      "Test Plan Alpha",
		Beds:      &beds,
		Baths:     &baths,
		HeatedSF:  &heatedSF,
		CreatedBy: user.ID,
	})
	if err != nil {
		t.Fatalf("CreatePlan: %v", err)
	}
	if plan.ID == "" {
		t.Fatal("expected non-empty plan ID")
	}
	if plan.Slug == "" {
		t.Fatal("expected non-empty slug")
	}
	if plan.Status != "incomplete" {
		t.Fatalf("expected status 'incomplete', got %q", plan.Status)
	}

	// Read
	fetched, err := store.GetPlanByID(ctx, plan.ID)
	if err != nil {
		t.Fatalf("GetPlanByID: %v", err)
	}
	if fetched == nil {
		t.Fatal("expected plan, got nil")
	}
	if fetched.Name != plan.Name {
		t.Fatalf("name mismatch: %q vs %q", fetched.Name, plan.Name)
	}

	// Read by slug
	bySlug, err := store.GetPlanBySlug(ctx, plan.Slug)
	if err != nil {
		t.Fatalf("GetPlanBySlug: %v", err)
	}
	if bySlug == nil || bySlug.ID != plan.ID {
		t.Fatal("GetPlanBySlug returned wrong plan")
	}

	// Update
	newBeds := 4
	updated, err := store.UpdatePlan(ctx, plan.ID, db.UpdatePlanInput{
		Name:      "Test Plan Beta",
		Beds:      &newBeds,
		Baths:     &baths,
		HeatedSF:  &heatedSF,
		UpdatedBy: user.ID,
	})
	if err != nil {
		t.Fatalf("UpdatePlan: %v", err)
	}
	if updated.Name != "Test Plan Beta" {
		t.Fatalf("expected updated name %q, got %q", "Test Plan Beta", updated.Name)
	}
	if updated.Beds == nil || *updated.Beds != 4 {
		t.Fatalf("expected beds=4, got %v", updated.Beds)
	}

	// Soft delete
	if err := store.SoftDeletePlan(ctx, plan.ID); err != nil {
		t.Fatalf("SoftDeletePlan: %v", err)
	}
	gone, err := store.GetPlanByID(ctx, plan.ID)
	if err != nil {
		t.Fatalf("GetPlanByID after delete: %v", err)
	}
	if gone != nil {
		t.Fatal("expected nil after soft delete, plan still visible")
	}

	// Restore
	if err := store.RestorePlan(ctx, plan.ID); err != nil {
		t.Fatalf("RestorePlan: %v", err)
	}
	restored, err := store.GetPlanByID(ctx, plan.ID)
	if err != nil {
		t.Fatalf("GetPlanByID after restore: %v", err)
	}
	if restored == nil {
		t.Fatal("expected plan after restore, got nil")
	}

	// Cleanup: soft delete again so FK constraints don't block user deletion
	store.SoftDeletePlan(ctx, plan.ID)
}

// TestPlan_SlugDeduplication verifies two plans with the same name get unique slugs (integration).
func TestPlan_SlugDeduplication(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	user, err := store.CreateUser(ctx, "Slug Test User", "slug-test@example.com", "password123", "editor")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	t.Cleanup(func() { store.DeleteUser(ctx, user.ID) })

	p1, err := store.CreatePlan(ctx, db.CreatePlanInput{Name: "Duplicate Slug Plan", CreatedBy: user.ID})
	if err != nil {
		t.Fatalf("CreatePlan 1: %v", err)
	}
	p2, err := store.CreatePlan(ctx, db.CreatePlanInput{Name: "Duplicate Slug Plan", CreatedBy: user.ID})
	if err != nil {
		t.Fatalf("CreatePlan 2: %v", err)
	}

	if p1.Slug == p2.Slug {
		t.Fatalf("expected unique slugs, both got %q", p1.Slug)
	}

	store.SoftDeletePlan(ctx, p1.ID)
	store.SoftDeletePlan(ctx, p2.ID)
}
