package auth_service

import (
	"context"

	"github.com/google/uuid"
	core_domain "github.com/zosinkin/social_network/internal/core/domain"
	core_kafka "github.com/zosinkin/social_network/internal/core/transport/kafka"
)


// fakeUserRepo — ручная fake-реализация UserRepo: каждый тест задаёт только
// то поле-функцию, которое ему нужно для конкретного сценария.
type fakeUserRepo struct {
	registerUserFunc         func(ctx context.Context, user core_domain.AuthUser) (core_domain.AuthUser, error)
	getUserByPhoneNumberFunc func(ctx context.Context, phoneNumber string) (core_domain.AuthUser, error)
	getUserByIDFunc          func(ctx context.Context, id uuid.UUID) (core_domain.AuthUser, error)
}

func (f *fakeUserRepo) RegisterUser(ctx context.Context, user core_domain.AuthUser) (core_domain.AuthUser, error) {
	return f.registerUserFunc(ctx, user)
}

func (f *fakeUserRepo) GetUserByPhoneNumber(ctx context.Context, phoneNumber string) (core_domain.AuthUser, error) {
	return f.getUserByPhoneNumberFunc(ctx, phoneNumber)
}

func (f *fakeUserRepo) GetUserByID(ctx context.Context, id uuid.UUID) (core_domain.AuthUser, error) {
	return f.getUserByIDFunc(ctx, id)
}

// fakeRefreshTokenRepo — fake RefreshTokenRepo.
type fakeRefreshTokenRepo struct {
	createRefreshTokenFunc func(ctx context.Context, token core_domain.RefreshToken) (core_domain.RefreshToken, error)
	getRefreshTokenFunc    func(ctx context.Context, tokenString string) (*core_domain.RefreshToken, error)
	revokeRefreshTokenFunc func(ctx context.Context, tokenString string) error
}

func (f *fakeRefreshTokenRepo) CreateRefreshToken(ctx context.Context, token core_domain.RefreshToken) (core_domain.RefreshToken, error) {
	return f.createRefreshTokenFunc(ctx, token)
}

func (f *fakeRefreshTokenRepo) GetRefreshToken(ctx context.Context, tokenString string) (*core_domain.RefreshToken, error) {
	return f.getRefreshTokenFunc(ctx, tokenString)
}

func (f *fakeRefreshTokenRepo) RevokeRefreshToken(ctx context.Context, tokenString string) error {
	if f.revokeRefreshTokenFunc == nil {
		return nil
	}
	return f.revokeRefreshTokenFunc(ctx, tokenString)
}

// fakePublisher — по умолчанию (без заданного publishFunc) считает
// публикацию успешной, чтобы тестам, не проверяющим Kafka, не нужно было
// задавать функцию вовсе.
type fakePublisher struct {
	publishFunc func(ctx context.Context, message core_kafka.Message) error
}

func (f *fakePublisher) Publish(ctx context.Context, message core_kafka.Message) error {
	if f.publishFunc == nil {
		return nil
	}
	return f.publishFunc(ctx, message)
}
