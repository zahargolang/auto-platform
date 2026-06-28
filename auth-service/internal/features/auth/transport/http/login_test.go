package auth_transport_http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	core_errors "github.com/zosinkin/social_network/internal/core/errors"
)

func TestLogin_Success(t *testing.T) {
	svc := &fakeService{
		loginWithRefreshFunc: func(ctx context.Context, phoneNumber, password string, refreshTokenTTL time.Duration) (string, string, error) {
			if phoneNumber != "+77001234567" || password != "Sup3rSecret!" {
				t.Fatalf("LoginWithRefresh() called with unexpected args: %q %q", phoneNumber, password)
			}
			return "access-token", "refresh-token", nil
		},
	}
	h := NewAuthHTTPHandler(svc)

	c, w := newTestContext(`{"phone_number":"+77001234567","password":"Sup3rSecret!"}`)
	h.Login(c)

	if w.Code != http.StatusOK {
		t.Fatalf("Login() status = %d, want %d (body: %s)", w.Code, http.StatusOK, w.Body.String())
	}

	var resp LoginResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Login() invalid JSON body: %v", err)
	}
	if resp.AccessToken != "access-token" || resp.RefreshToken != "refresh-token" {
		t.Fatalf("Login() response = %+v, want access/refresh tokens from service", resp)
	}
}

func TestLogin_InvalidJSON(t *testing.T) {
	h := NewAuthHTTPHandler(&fakeService{})

	c, w := newTestContext(`not-json`)
	h.Login(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Login() status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestLogin_InvalidCredentials(t *testing.T) {
	svc := &fakeService{
		loginWithRefreshFunc: func(ctx context.Context, phoneNumber, password string, refreshTokenTTL time.Duration) (string, string, error) {
			return "", "", core_errors.ErrInvalidCredentials
		},
	}
	h := NewAuthHTTPHandler(svc)

	c, w := newTestContext(`{"phone_number":"+77001234567","password":"Sup3rSecret!"}`)
	h.Login(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("Login() status = %d, want %d (body: %s)", w.Code, http.StatusUnauthorized, w.Body.String())
	}
}

func TestLogin_ServiceError(t *testing.T) {
	svc := &fakeService{
		loginWithRefreshFunc: func(ctx context.Context, phoneNumber, password string, refreshTokenTTL time.Duration) (string, string, error) {
			return "", "", errors.New("db is down")
		},
	}
	h := NewAuthHTTPHandler(svc)

	c, w := newTestContext(`{"phone_number":"+77001234567","password":"Sup3rSecret!"}`)
	h.Login(c)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("Login() status = %d, want %d (body: %s)", w.Code, http.StatusInternalServerError, w.Body.String())
	}
}
