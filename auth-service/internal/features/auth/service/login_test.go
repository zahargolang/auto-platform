package auth_service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	core_domain "github.com/zosinkin/social_network/internal/core/domain"
	core_errors "github.com/zosinkin/social_network/internal/core/errors"
)

func registeredUser(t *testing.T, password string) core_domain.AuthUser {
	t.Helper()
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() unexpected error: %v", err)
	}
	return core_domain.NewAuthUser(validUsername, validPhone, hash)
}

func TestService_Login_Success(t *testing.T) {
	user := registeredUser(t, validPassword)
	userRepo := &fakeUserRepo{
		getUserByPhoneNumberFunc: func(ctx context.Context, phoneNumber string) (core_domain.AuthUser, error) {
			return user, nil
		},
	}
	secret := []byte("secret")
	svc := NewAuthService(userRepo, &fakeRefreshTokenRepo{}, secret, time.Minute, &fakePublisher{}, testLogger())

	token, err := svc.Login(context.Background(), validPhone, validPassword)
	if err != nil {
		t.Fatalf("Login() unexpected error: %v", err)
	}

	claims := parseClaims(t, token, secret)
	if claims["sub"] != user.ID.String() {
		t.Fatalf("Login() token sub = %v, want %v", claims["sub"], user.ID.String())
	}
}

func TestService_Login_UserNotFound(t *testing.T) {
	userRepo := &fakeUserRepo{
		getUserByPhoneNumberFunc: func(ctx context.Context, phoneNumber string) (core_domain.AuthUser, error) {
			return core_domain.AuthUser{}, phoneNotFoundErr()
		},
	}
	svc := NewAuthService(userRepo, &fakeRefreshTokenRepo{}, []byte("secret"), time.Minute, &fakePublisher{}, testLogger())

	_, err := svc.Login(context.Background(), validPhone, validPassword)
	if !errors.Is(err, core_errors.ErrInvalidCredentials) {
		t.Fatalf("Login() error = %v, want %v", err, core_errors.ErrInvalidCredentials)
	}
}

func TestService_Login_WrongPassword(t *testing.T) {
	user := registeredUser(t, validPassword)
	userRepo := &fakeUserRepo{
		getUserByPhoneNumberFunc: func(ctx context.Context, phoneNumber string) (core_domain.AuthUser, error) {
			return user, nil
		},
	}
	svc := NewAuthService(userRepo, &fakeRefreshTokenRepo{}, []byte("secret"), time.Minute, &fakePublisher{}, testLogger())

	_, err := svc.Login(context.Background(), validPhone, "WrongPassword1!")
	if !errors.Is(err, core_errors.ErrInvalidCredentials) {
		t.Fatalf("Login() error = %v, want %v", err, core_errors.ErrInvalidCredentials)
	}
}

func TestService_LoginWithRefresh_Success(t *testing.T) {
	user := registeredUser(t, validPassword)
	var createdToken core_domain.RefreshToken

	userRepo := &fakeUserRepo{
		getUserByPhoneNumberFunc: func(ctx context.Context, phoneNumber string) (core_domain.AuthUser, error) {
			return user, nil
		},
	}
	refreshRepo := &fakeRefreshTokenRepo{
		createRefreshTokenFunc: func(ctx context.Context, token core_domain.RefreshToken) (core_domain.RefreshToken, error) {
			createdToken = token
			return token, nil
		},
	}
	secret := []byte("secret")
	svc := NewAuthService(userRepo, refreshRepo, secret, time.Minute, &fakePublisher{}, testLogger())

	accessToken, refreshToken, err := svc.LoginWithRefresh(context.Background(), validPhone, validPassword, time.Hour)
	if err != nil {
		t.Fatalf("LoginWithRefresh() unexpected error: %v", err)
	}

	claims := parseClaims(t, accessToken, secret)
	if claims["sub"] != user.ID.String() {
		t.Fatalf("LoginWithRefresh() access token sub = %v, want %v", claims["sub"], user.ID.String())
	}
	if refreshToken != createdToken.Token {
		t.Fatalf("LoginWithRefresh() refresh token = %q, want %q", refreshToken, createdToken.Token)
	}
	if createdToken.UserID != user.ID {
		t.Fatalf("CreateRefreshToken() called with UserID = %v, want %v", createdToken.UserID, user.ID)
	}
}

func TestService_LoginWithRefresh_WrongCredentials(t *testing.T) {
	userRepo := &fakeUserRepo{
		getUserByPhoneNumberFunc: func(ctx context.Context, phoneNumber string) (core_domain.AuthUser, error) {
			return core_domain.AuthUser{}, phoneNotFoundErr()
		},
	}
	svc := NewAuthService(userRepo, &fakeRefreshTokenRepo{}, []byte("secret"), time.Minute, &fakePublisher{}, testLogger())

	_, _, err := svc.LoginWithRefresh(context.Background(), validPhone, validPassword, time.Hour)
	if !errors.Is(err, core_errors.ErrInvalidCredentials) {
		t.Fatalf("LoginWithRefresh() error = %v, want wrapped %v", err, core_errors.ErrInvalidCredentials)
	}
}

func TestService_LoginWithRefresh_CreateRefreshTokenError(t *testing.T) {
	user := registeredUser(t, validPassword)
	repoErr := errors.New("insert failed")

	userRepo := &fakeUserRepo{
		getUserByPhoneNumberFunc: func(ctx context.Context, phoneNumber string) (core_domain.AuthUser, error) {
			return user, nil
		},
	}
	refreshRepo := &fakeRefreshTokenRepo{
		createRefreshTokenFunc: func(ctx context.Context, token core_domain.RefreshToken) (core_domain.RefreshToken, error) {
			return core_domain.RefreshToken{}, repoErr
		},
	}
	svc := NewAuthService(userRepo, refreshRepo, []byte("secret"), time.Minute, &fakePublisher{}, testLogger())

	_, _, err := svc.LoginWithRefresh(context.Background(), validPhone, validPassword, time.Hour)
	if err == nil {
		t.Fatalf("LoginWithRefresh() expected error, got nil")
	}
}

// parseClaims разбирает access-токен тем же секретом, которым его подписали
// — самостоятельный, не зависящий от Service.ValidateToken способ проверить,
// что generateAccessToken действительно положил в claims то, что нужно.
func parseClaims(t *testing.T, tokenString string, secret []byte) jwt.MapClaims {
	t.Helper()

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		t.Fatalf("failed to parse generated token: %v", err)
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatalf("token claims have unexpected type %T", token.Claims)
	}
	return claims
}
