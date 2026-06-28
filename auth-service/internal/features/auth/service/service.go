package auth_service

import (
	"context"
	"time"

	"github.com/google/uuid"
	core_domain "github.com/zosinkin/social_network/internal/core/domain"
	core_kafka "github.com/zosinkin/social_network/internal/core/transport/kafka"
)

type Service struct {
	userRepo         UserRepo
	refreshTokenRepo RefreshTokenRepo
	jwtSecret        []byte
	accessTokenTTL   time.Duration
	publisher        EventPublisher
}

type UserRepo interface {
	RegisterUser(
		ctx context.Context,
		user core_domain.AuthUser,
	) (core_domain.AuthUser, error)

	GetUserByPhoneNumber(
		ctx context.Context,
		phoneNumber string,
	) (core_domain.AuthUser, error)

	GetUserByID(
		ctx context.Context,
		id uuid.UUID,
	) (core_domain.AuthUser, error)
}

type RefreshTokenRepo interface {
	CreateRefreshToken(
		ctx context.Context,
		token core_domain.RefreshToken,
	) (core_domain.RefreshToken, error)

	GetRefreshToken(
		ctx context.Context,
		refreshTokenString string,
	) (*core_domain.RefreshToken, error)

	RevokeRefreshToken(
		ctx context.Context,
		tokenString string,
	) error
}

type EventPublisher interface {
	Publish(
		ctx context.Context,
		message core_kafka.Message,
	) error
}

func NewAuthService(
	userRepo UserRepo,
	refreshRepo RefreshTokenRepo,
	jwtSecret []byte,
	accesTokenTTL time.Duration,
	publisher EventPublisher,
) *Service {
	return &Service{
		userRepo:         userRepo,
		refreshTokenRepo: refreshRepo,
		jwtSecret:        jwtSecret,
		accessTokenTTL:   accesTokenTTL,
		publisher:        publisher,
	}
}
