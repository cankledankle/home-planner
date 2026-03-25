package db_test

import (
	"testing"
	"time"

	"github.com/cankledankle/home-planner/internal/db"
	"github.com/google/uuid"
)

// TestHashToken_Deterministic verifies the same input always produces the same hash.
func TestHashToken_Deterministic(t *testing.T) {
	token := "test-token-abc123"
	h1 := db.HashToken(token)
	h2 := db.HashToken(token)
	if h1 != h2 {
		t.Fatalf("HashToken is not deterministic: got %q and %q", h1, h2)
	}
}

// TestHashToken_DifferentInputs verifies distinct tokens produce distinct hashes.
func TestHashToken_DifferentInputs(t *testing.T) {
	h1 := db.HashToken("token-a")
	h2 := db.HashToken("token-b")
	if h1 == h2 {
		t.Fatal("different tokens produced the same hash")
	}
}

// TestHashToken_EmptyString verifies an empty token hashes without panic.
func TestHashToken_EmptyString(t *testing.T) {
	h := db.HashToken("")
	if h == "" {
		t.Fatal("expected a non-empty hash for empty input")
	}
}

// TestHashToken_HexOutput verifies the hash is a 64-character hex string (SHA-256).
func TestHashToken_HexOutput(t *testing.T) {
	h := db.HashToken("any-input")
	if len(h) != 64 {
		t.Fatalf("expected 64-char hex hash, got %d chars: %q", len(h), h)
	}
	for _, c := range h {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Fatalf("hash contains non-hex character %q: %q", c, h)
		}
	}
}

// TestRefreshToken_RoundTrip stores a token and validates it (integration).
func TestRefreshToken_RoundTrip(t *testing.T) {
	store := newTestStore(t)

	userID := uuid.New()
	rawToken := uuid.NewString()
	tokenHash := db.HashToken(rawToken)
	expiry := time.Now().Add(time.Hour)

	// Store the token hash.
	if err := store.StoreRefreshToken(userID, tokenHash, expiry); err != nil {
		t.Fatalf("StoreRefreshToken: %v", err)
	}

	// Validate using the raw token (ValidateRefreshToken hashes internally).
	gotUserID, err := store.ValidateRefreshToken(rawToken)
	if err != nil {
		t.Fatalf("ValidateRefreshToken: %v", err)
	}
	if gotUserID != userID {
		t.Fatalf("expected userID %v, got %v", userID, gotUserID)
	}

	// Delete the token and confirm it is no longer valid.
	if err := store.DeleteRefreshTokenByValue(rawToken); err != nil {
		t.Fatalf("DeleteRefreshTokenByValue: %v", err)
	}
	if _, err := store.ValidateRefreshToken(rawToken); err == nil {
		t.Fatal("expected error after deletion, got nil")
	}
}

// TestRefreshToken_ExpiredIsRejected verifies an expired token fails validation (integration).
func TestRefreshToken_ExpiredIsRejected(t *testing.T) {
	store := newTestStore(t)

	userID := uuid.New()
	rawToken := uuid.NewString()
	tokenHash := db.HashToken(rawToken)
	expiry := time.Now().Add(-time.Second) // already expired

	if err := store.StoreRefreshToken(userID, tokenHash, expiry); err != nil {
		t.Fatalf("StoreRefreshToken: %v", err)
	}

	if _, err := store.ValidateRefreshToken(rawToken); err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
}
