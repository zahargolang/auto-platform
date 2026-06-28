package core_domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewRefreshToken(t *testing.T) {
	userID := uuid.New()
	ttl := time.Hour

	before := time.Now()
	token := NewRefreshToken(userID, ttl)
	after := time.Now()

	if token.UserID != userID {
		t.Fatalf("UserID = %v, want %v", token.UserID, userID)
	}

	if token.Token != token.ID.String() {
		t.Fatalf("Token = %q, want it to equal ID.String() = %q", token.Token, token.ID.String())
	}

	if token.Revoked {
		t.Fatalf("Revoked = true, want false for a freshly created token")
	}

	if token.CreatedAt.Before(before) || token.CreatedAt.After(after) {
		t.Fatalf("CreatedAt = %v, want it to be between %v and %v", token.CreatedAt, before, after)
	}

	wantExpiresAt := token.CreatedAt.Add(ttl)
	if diff := token.ExpiresAt.Sub(wantExpiresAt); diff < -time.Second || diff > time.Second {
		t.Fatalf("ExpiresAt = %v, want approximately %v", token.ExpiresAt, wantExpiresAt)
	}
}

func TestNewRefreshToken_GeneratesDistinctIDs(t *testing.T) {
	userID := uuid.New()

	t1 := NewRefreshToken(userID, time.Hour)
	t2 := NewRefreshToken(userID, time.Hour)

	if t1.ID == t2.ID {
		t.Fatalf("expected distinct generated IDs, got the same: %v", t1.ID)
	}
	if t1.Token == t2.Token {
		t.Fatalf("expected distinct generated tokens, got the same: %v", t1.Token)
	}
}
