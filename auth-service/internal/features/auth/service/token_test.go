package auth_service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	core_domain "github.com/zosinkin/social_network/internal/core/domain"
	core_errors "github.com/zosinkin/social_network/internal/core/errors"
)

func TestService_RefreshAccessToken_Success(t *testing.T) {
	user := registeredUser(t, validPassword)
	stored := core_domain.NewRefreshToken(user.ID, time.Hour)

	refreshRepo := &fakeRefreshTokenRepo{
		getRefreshTokenFunc: func(ctx context.Context, tokenString string) (*core_domain.RefreshToken, error) {
			return &stored, nil
		},
	}
	userRepo := &fakeUserRepo{
		getUserByIDFunc: func(ctx context.Context, id uuid.UUID) (core_domain.AuthUser, error) {
			return user, nil
		},
	}
	secret := []byte("secret")
	svc := NewAuthService(userRepo, refreshRepo, secret, time.Minute, &fakePublisher{})

	accessToken, err := svc.RefreshAccessToken(context.Background(), stored.Token)
	if err != nil {
		t.Fatalf("RefreshAccessToken() unexpected error: %v", err)
	}

	claims := parseClaims(t, accessToken, secret)
	if claims["sub"] != user.ID.String() {
		t.Fatalf("RefreshAccessToken() token sub = %v, want %v", claims["sub"], user.ID.String())
	}
}

func TestService_RefreshAccessToken_NotFound(t *testing.T) {
	refreshRepo := &fakeRefreshTokenRepo{
		getRefreshTokenFunc: func(ctx context.Context, tokenString string) (*core_domain.RefreshToken, error) {
			return nil, errors.New("not found")
		},
	}
	svc := NewAuthService(&fakeUserRepo{}, refreshRepo, []byte("secret"), time.Minute, &fakePublisher{})

	_, err := svc.RefreshAccessToken(context.Background(), "unknown-token")
	if !errors.Is(err, core_errors.ErrInvalidToken) {
		t.Fatalf("RefreshAccessToken() error = %v, want %v", err, core_errors.ErrInvalidToken)
	}
}

func TestService_RefreshAccessToken_Revoked(t *testing.T) {
	revoked := core_domain.NewRefreshToken(uuid.New(), time.Hour)
	revoked.Revoked = true

	refreshRepo := &fakeRefreshTokenRepo{
		getRefreshTokenFunc: func(ctx context.Context, tokenString string) (*core_domain.RefreshToken, error) {
			return &revoked, nil
		},
	}
	svc := NewAuthService(&fakeUserRepo{}, refreshRepo, []byte("secret"), time.Minute, &fakePublisher{})

	_, err := svc.RefreshAccessToken(context.Background(), revoked.Token)
	if !errors.Is(err, core_errors.ErrInvalidToken) {
		t.Fatalf("RefreshAccessToken() error = %v, want %v", err, core_errors.ErrInvalidToken)
	}
}

func TestService_RefreshAccessToken_Expired(t *testing.T) {
	expired := core_domain.NewRefreshToken(uuid.New(), -time.Hour) // ttl в прошлом

	refreshRepo := &fakeRefreshTokenRepo{
		getRefreshTokenFunc: func(ctx context.Context, tokenString string) (*core_domain.RefreshToken, error) {
			return &expired, nil
		},
	}
	svc := NewAuthService(&fakeUserRepo{}, refreshRepo, []byte("secret"), time.Minute, &fakePublisher{})

	_, err := svc.RefreshAccessToken(context.Background(), expired.Token)
	if !errors.Is(err, core_errors.ErrExpiredToken) {
		t.Fatalf("RefreshAccessToken() error = %v, want %v", err, core_errors.ErrExpiredToken)
	}
}

func TestService_RefreshAccessToken_GetUserError(t *testing.T) {
	stored := core_domain.NewRefreshToken(uuid.New(), time.Hour)
	repoErr := errors.New("user gone")

	refreshRepo := &fakeRefreshTokenRepo{
		getRefreshTokenFunc: func(ctx context.Context, tokenString string) (*core_domain.RefreshToken, error) {
			return &stored, nil
		},
	}
	userRepo := &fakeUserRepo{
		getUserByIDFunc: func(ctx context.Context, id uuid.UUID) (core_domain.AuthUser, error) {
			return core_domain.AuthUser{}, repoErr
		},
	}
	svc := NewAuthService(userRepo, refreshRepo, []byte("secret"), time.Minute, &fakePublisher{})

	_, err := svc.RefreshAccessToken(context.Background(), stored.Token)
	if err == nil {
		t.Fatalf("RefreshAccessToken() expected error, got nil")
	}
}

func TestService_ValidateToken_Success(t *testing.T) {
	user := registeredUser(t, validPassword)
	secret := []byte("secret")
	svc := NewAuthService(&fakeUserRepo{}, &fakeRefreshTokenRepo{}, secret, time.Minute, &fakePublisher{})

	tokenString, err := svc.generateAccessToken(&user)
	if err != nil {
		t.Fatalf("generateAccessToken() unexpected error: %v", err)
	}

	claims, err := svc.ValidateToken(tokenString)
	if err != nil {
		t.Fatalf("ValidateToken() unexpected error: %v", err)
	}
	if claims["sub"] != user.ID.String() {
		t.Fatalf("ValidateToken() sub = %v, want %v", claims["sub"], user.ID.String())
	}
	if claims["username"] != user.Username {
		t.Fatalf("ValidateToken() username = %v, want %v", claims["username"], user.Username)
	}
}

func TestService_ValidateToken_Expired(t *testing.T) {
	user := registeredUser(t, validPassword)
	secret := []byte("secret")
	// accessTokenTTL отрицательный — токен рождается уже просроченным.
	svc := NewAuthService(&fakeUserRepo{}, &fakeRefreshTokenRepo{}, secret, -time.Minute, &fakePublisher{})

	tokenString, err := svc.generateAccessToken(&user)
	if err != nil {
		t.Fatalf("generateAccessToken() unexpected error: %v", err)
	}

	_, err = svc.ValidateToken(tokenString)
	if !errors.Is(err, core_errors.ErrExpiredToken) {
		t.Fatalf("ValidateToken() error = %v, want %v", err, core_errors.ErrExpiredToken)
	}
}

func TestService_ValidateToken_WrongSecret(t *testing.T) {
	user := registeredUser(t, validPassword)
	signingSvc := NewAuthService(&fakeUserRepo{}, &fakeRefreshTokenRepo{}, []byte("secret-a"), time.Minute, &fakePublisher{})
	validatingSvc := NewAuthService(&fakeUserRepo{}, &fakeRefreshTokenRepo{}, []byte("secret-b"), time.Minute, &fakePublisher{})

	tokenString, err := signingSvc.generateAccessToken(&user)
	if err != nil {
		t.Fatalf("generateAccessToken() unexpected error: %v", err)
	}

	_, err = validatingSvc.ValidateToken(tokenString)
	if !errors.Is(err, core_errors.ErrInvalidToken) {
		t.Fatalf("ValidateToken() error = %v, want %v", err, core_errors.ErrInvalidToken)
	}
}

func TestService_ValidateToken_Garbage(t *testing.T) {
	svc := NewAuthService(&fakeUserRepo{}, &fakeRefreshTokenRepo{}, []byte("secret"), time.Minute, &fakePublisher{})

	_, err := svc.ValidateToken("not-a-jwt-at-all")
	if !errors.Is(err, core_errors.ErrInvalidToken) {
		t.Fatalf("ValidateToken() error = %v, want %v", err, core_errors.ErrInvalidToken)
	}
}

func TestService_ValidateToken_WrongSigningMethod(t *testing.T) {
	user := registeredUser(t, validPassword)
	svc := NewAuthService(&fakeUserRepo{}, &fakeRefreshTokenRepo{}, []byte("secret"), time.Minute, &fakePublisher{})

	// "none"-алгоритм — классическая JWT-атака, которую ValidateToken должен отбивать.
	token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{"sub": user.ID.String()})
	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("failed to build unsigned token: %v", err)
	}

	_, err = svc.ValidateToken(tokenString)
	if !errors.Is(err, core_errors.ErrInvalidToken) {
		t.Fatalf("ValidateToken() error = %v, want %v", err, core_errors.ErrInvalidToken)
	}
}
