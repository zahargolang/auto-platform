package auth_postgres_repository

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	core_domain "github.com/zosinkin/social_network/internal/core/domain"
	core_postgres_pool "github.com/zosinkin/social_network/internal/core/repository/postgres/pool"
)

func TestUserRepo_RegisterUser_Success(t *testing.T) {
	wantID := uuid.New()
	pool := &fakePool{
		queryRowFunc: func(ctx context.Context, sql string, args ...any) core_postgres_pool.Row {
			return &fakeRow{values: []any{wantID, "john_doe", "+77001234567"}}
		},
	}
	repo := NewUsersRepo(pool)

	user := core_domain.NewAuthUser("john_doe", "+77001234567", "hash")
	result, err := repo.RegisterUser(context.Background(), user)
	if err != nil {
		t.Fatalf("RegisterUser() unexpected error: %v", err)
	}
	if result.ID != wantID || result.Username != "john_doe" || result.PhoneNumber != "+77001234567" {
		t.Fatalf("RegisterUser() = %+v, unexpected", result)
	}
	// RETURNING в запросе не включает password_hash — Scan его не трогает,
	// поэтому в результате он всегда пуст, даже если на входе был задан.
	if result.PasswordHash != "" {
		t.Fatalf("RegisterUser() PasswordHash = %q, want empty (not part of RETURNING)", result.PasswordHash)
	}
}

func TestUserRepo_RegisterUser_ScanError(t *testing.T) {
	pool := &fakePool{
		queryRowFunc: func(ctx context.Context, sql string, args ...any) core_postgres_pool.Row {
			return &fakeRow{err: errors.New("unique violation")}
		},
	}
	repo := NewUsersRepo(pool)

	_, err := repo.RegisterUser(context.Background(), core_domain.NewAuthUser("john_doe", "+77001234567", "hash"))
	if err == nil {
		t.Fatalf("RegisterUser() expected error, got nil")
	}
}

func TestUserRepo_GetUserByPhoneNumber_Success(t *testing.T) {
	wantID := uuid.New()
	pool := &fakePool{
		queryRowFunc: func(ctx context.Context, sql string, args ...any) core_postgres_pool.Row {
			return &fakeRow{values: []any{wantID, "john_doe", "+77001234567", "hashed-pw"}}
		},
	}
	repo := NewUsersRepo(pool)

	result, err := repo.GetUserByPhoneNumber(context.Background(), "+77001234567")
	if err != nil {
		t.Fatalf("GetUserByPhoneNumber() unexpected error: %v", err)
	}
	if result.PasswordHash != "hashed-pw" {
		t.Fatalf("GetUserByPhoneNumber() PasswordHash = %q, want %q", result.PasswordHash, "hashed-pw")
	}
}

func TestUserRepo_GetUserByPhoneNumber_NotFound(t *testing.T) {
	pool := &fakePool{
		queryRowFunc: func(ctx context.Context, sql string, args ...any) core_postgres_pool.Row {
			return &fakeRow{err: core_postgres_pool.ErrNoRows}
		},
	}
	repo := NewUsersRepo(pool)

	_, err := repo.GetUserByPhoneNumber(context.Background(), "+77001234567")
	if !errors.Is(err, core_postgres_pool.ErrNoRows) {
		t.Fatalf("GetUserByPhoneNumber() error = %v, want wrapped %v", err, core_postgres_pool.ErrNoRows)
	}
}

func TestUserRepo_GetUserByID_Success(t *testing.T) {
	wantID := uuid.New()
	pool := &fakePool{
		queryRowFunc: func(ctx context.Context, sql string, args ...any) core_postgres_pool.Row {
			return &fakeRow{values: []any{wantID, "john_doe", "+77001234567"}}
		},
	}
	repo := NewUsersRepo(pool)

	result, err := repo.GetUserByID(context.Background(), wantID)
	if err != nil {
		t.Fatalf("GetUserByID() unexpected error: %v", err)
	}
	if result.ID != wantID {
		t.Fatalf("GetUserByID() ID = %v, want %v", result.ID, wantID)
	}
}

// TestUserRepo_GetUserByID_NotFound проверяет, что ошибка "не найдено"
// оборачивается через "%w" — errors.Is(err, ErrNoRows) должен срабатывать,
// чтобы сервисный слой (RefreshAccessToken -> GetUserByID) мог отличить
// "юзера нет" от любой другой ошибки БД.
func TestUserRepo_GetUserByID_NotFound(t *testing.T) {
	pool := &fakePool{
		queryRowFunc: func(ctx context.Context, sql string, args ...any) core_postgres_pool.Row {
			return &fakeRow{err: core_postgres_pool.ErrNoRows}
		},
	}
	repo := NewUsersRepo(pool)

	_, err := repo.GetUserByID(context.Background(), uuid.New())
	if !errors.Is(err, core_postgres_pool.ErrNoRows) {
		t.Fatalf("GetUserByID() error = %v, want wrapped %v", err, core_postgres_pool.ErrNoRows)
	}
}
