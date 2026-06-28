package auth_transport_http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	core_errors "github.com/zosinkin/social_network/internal/core/errors"
)

func TestRefreshToken_Success(t *testing.T) {
	svc := &fakeService{
		refreshAccessTokenFunc: func(ctx context.Context, refreshToken string) (string, error) {
			if refreshToken != "old-refresh-token" {
				t.Fatalf("RefreshAccessToken() called with unexpected token: %q", refreshToken)
			}
			return "new-access-token", nil
		},
	}
	h := NewAuthHTTPHandler(svc)

	c, w := newTestContext(`{"refresh_token":"old-refresh-token"}`)
	h.RefreshToken(c)

	if w.Code != http.StatusOK {
		t.Fatalf("RefreshToken() status = %d, want %d (body: %s)", w.Code, http.StatusOK, w.Body.String())
	}

	var resp RefreshResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("RefreshToken() invalid JSON body: %v", err)
	}
	if resp.Token != "new-access-token" {
		t.Fatalf("RefreshToken() token = %q, want %q", resp.Token, "new-access-token")
	}
}

func TestRefreshToken_InvalidJSON(t *testing.T) {
	h := NewAuthHTTPHandler(&fakeService{})

	c, w := newTestContext(`not-json`)
	h.RefreshToken(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("RefreshToken() status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	svc := &fakeService{
		refreshAccessTokenFunc: func(ctx context.Context, refreshToken string) (string, error) {
			return "", core_errors.ErrInvalidToken
		},
	}
	h := NewAuthHTTPHandler(svc)

	c, w := newTestContext(`{"refresh_token":"bad-token"}`)
	h.RefreshToken(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("RefreshToken() status = %d, want %d (body: %s)", w.Code, http.StatusUnauthorized, w.Body.String())
	}
}

func TestRefreshToken_ExpiredToken(t *testing.T) {
	svc := &fakeService{
		refreshAccessTokenFunc: func(ctx context.Context, refreshToken string) (string, error) {
			return "", core_errors.ErrExpiredToken
		},
	}
	h := NewAuthHTTPHandler(svc)

	c, w := newTestContext(`{"refresh_token":"expired-token"}`)
	h.RefreshToken(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("RefreshToken() status = %d, want %d (body: %s)", w.Code, http.StatusUnauthorized, w.Body.String())
	}
}

func TestRefreshToken_ServiceError(t *testing.T) {
	svc := &fakeService{
		refreshAccessTokenFunc: func(ctx context.Context, refreshToken string) (string, error) {
			return "", errors.New("db is down")
		},
	}
	h := NewAuthHTTPHandler(svc)

	c, w := newTestContext(`{"refresh_token":"some-token"}`)
	h.RefreshToken(c)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("RefreshToken() status = %d, want %d (body: %s)", w.Code, http.StatusInternalServerError, w.Body.String())
	}
}
