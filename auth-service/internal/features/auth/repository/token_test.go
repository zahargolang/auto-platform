package auth_postgres_repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	core_domain "github.com/zosinkin/social_network/internal/core/domain"
	core_errors "github.com/zosinkin/social_network/internal/core/errors"
	core_postgres_pool "github.com/zosinkin/social_network/internal/core/repository/postgres/pool"
)

func TestRefreshTokenRepo_GetRefreshToken_Success(t *testing.T) {
	want := core_domain.NewRefreshToken(uuid.New(), time.Hour)
	pool := &fakePool{
		queryRowFunc: func(ctx context.Context, sql string, args ...any) core_postgres_pool.Row {
			return &fakeRow{values: []any{want.ID, want.UserID, want.Token, want.ExpiresAt, want.CreatedAt, want.Revoked}}
		},
	}
	repo := NewRefreshTokenRepo(pool)

	result, err := repo.GetRefreshToken(context.Background(), want.Token)
	if err != nil {
		t.Fatalf("GetRefreshToken() unexpected error: %v", err)
	}
	if result.ID != want.ID || result.UserID != want.UserID || result.Token != want.Token {
		t.Fatalf("GetRefreshToken() = %+v, want %+v", result, want)
	}
}

// TestRefreshTokenRepo_GetRefreshToken_NotFound проверяет, что "токен не
// найден" оборачивается через "%w" — errors.Is(err, core_errors.ErrNotFound)
// должен срабатывать, чтобы сервисный слой мог отличить эту ситуацию от
// любой другой ошибки БД.
func TestRefreshTokenRepo_GetRefreshToken_NotFound(t *testing.T) {
	pool := &fakePool{
		queryRowFunc: func(ctx context.Context, sql string, args ...any) core_postgres_pool.Row {
			return &fakeRow{err: core_postgres_pool.ErrNoRows}
		},
	}
	repo := NewRefreshTokenRepo(pool)

	_, err := repo.GetRefreshToken(context.Background(), "missing-token")
	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("GetRefreshToken() error = %v, want wrapped %v", err, core_errors.ErrNotFound)
	}
}

// TestRefreshTokenRepo_GetRefreshToken_OtherScanError проверяет, что любая
// ошибка скана, отличная от core_postgres_pool.ErrNoRows (например, обрыв
// соединения), действительно возвращается вызывающему коду, а не
// проглатывается молча.
func TestRefreshTokenRepo_GetRefreshToken_OtherScanError(t *testing.T) {
	scanErr := errors.New("connection reset by peer")
	pool := &fakePool{
		queryRowFunc: func(ctx context.Context, sql string, args ...any) core_postgres_pool.Row {
			return &fakeRow{err: scanErr}
		},
	}
	repo := NewRefreshTokenRepo(pool)

	_, err := repo.GetRefreshToken(context.Background(), "some-token")
	if err == nil {
		t.Fatalf("GetRefreshToken() expected error, got nil")
	}
}

func TestRefreshTokenRepo_CreateRefreshToken_Success(t *testing.T) {
	input := core_domain.NewRefreshToken(uuid.New(), time.Hour)
	pool := &fakePool{
		queryRowFunc: func(ctx context.Context, sql string, args ...any) core_postgres_pool.Row {
			return &fakeRow{values: []any{input.ID, input.UserID, input.Token, input.ExpiresAt, input.CreatedAt, input.Revoked}}
		},
	}
	repo := NewRefreshTokenRepo(pool)

	result, err := repo.CreateRefreshToken(context.Background(), input)
	if err != nil {
		t.Fatalf("CreateRefreshToken() unexpected error: %v", err)
	}
	if result != input {
		t.Fatalf("CreateRefreshToken() = %+v, want %+v", result, input)
	}
}

func TestRefreshTokenRepo_CreateRefreshToken_ScanError(t *testing.T) {
	pool := &fakePool{
		queryRowFunc: func(ctx context.Context, sql string, args ...any) core_postgres_pool.Row {
			return &fakeRow{err: errors.New("insert failed")}
		},
	}
	repo := NewRefreshTokenRepo(pool)

	_, err := repo.CreateRefreshToken(context.Background(), core_domain.NewRefreshToken(uuid.New(), time.Hour))
	if err == nil {
		t.Fatalf("CreateRefreshToken() expected error, got nil")
	}
}

func TestRefreshTokenRepo_RevokeRefreshToken_Success(t *testing.T) {
	pool := &fakePool{
		execFunc: func(ctx context.Context, sql string, args ...any) (core_postgres_pool.CommandTag, error) {
			return &fakeCommandTag{rows: 1}, nil
		},
	}
	repo := NewRefreshTokenRepo(pool)

	if err := repo.RevokeRefreshToken(context.Background(), "some-token"); err != nil {
		t.Fatalf("RevokeRefreshToken() unexpected error: %v", err)
	}
}

func TestRefreshTokenRepo_RevokeRefreshToken_ExecError(t *testing.T) {
	pool := &fakePool{
		execFunc: func(ctx context.Context, sql string, args ...any) (core_postgres_pool.CommandTag, error) {
			return nil, errors.New("update failed")
		},
	}
	repo := NewRefreshTokenRepo(pool)

	if err := repo.RevokeRefreshToken(context.Background(), "some-token"); err == nil {
		t.Fatalf("RevokeRefreshToken() expected error, got nil")
	}
}
