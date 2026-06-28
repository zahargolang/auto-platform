package auth_service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	core_domain "github.com/zosinkin/social_network/internal/core/domain"
	core_errors "github.com/zosinkin/social_network/internal/core/errors"
	core_postgres_pool "github.com/zosinkin/social_network/internal/core/repository/postgres/pool"
	core_kafka "github.com/zosinkin/social_network/internal/core/transport/kafka"
)

const (
	validUsername = "john_doe"
	validPhone    = "+77001234567"
	validPassword = "Sup3rSecret!"
)

// phoneNotFound — fake-ответ "такого номера ещё нет", как его в реальности
// отдаёт репозиторий (ошибка, оборачивающая core_postgres_pool.ErrNoRows).
func phoneNotFoundErr() error {
	return fmt.Errorf("not found: %w", core_postgres_pool.ErrNoRows)
}

func TestService_Register_Success(t *testing.T) {
	var registeredArg core_domain.AuthUser
	var published core_kafka.Message

	userRepo := &fakeUserRepo{
		getUserByPhoneNumberFunc: func(ctx context.Context, phoneNumber string) (core_domain.AuthUser, error) {
			return core_domain.AuthUser{}, phoneNotFoundErr()
		},
		registerUserFunc: func(ctx context.Context, user core_domain.AuthUser) (core_domain.AuthUser, error) {
			registeredArg = user
			return user, nil
		},
	}
	publisher := &fakePublisher{
		publishFunc: func(ctx context.Context, message core_kafka.Message) error {
			published = message
			return nil
		},
	}

	svc := NewAuthService(userRepo, &fakeRefreshTokenRepo{}, []byte("secret"), time.Minute, publisher, testLogger())

	user, err := svc.Register(context.Background(), validUsername, validPhone, validPassword)
	if err != nil {
		t.Fatalf("Register() unexpected error: %v", err)
	}

	if user.Username != validUsername || user.PhoneNumber != validPhone {
		t.Fatalf("Register() unexpected user: %+v", user)
	}
	if user.PasswordHash == "" || user.PasswordHash == validPassword {
		t.Fatalf("Register() password was not hashed: %q", user.PasswordHash)
	}
	if registeredArg.Username != validUsername {
		t.Fatalf("RegisterUser() called with unexpected user: %+v", registeredArg)
	}
	if published.Topic != core_kafka.TopicUserRegistered {
		t.Fatalf("Publish() called with topic %q, want %q", published.Topic, core_kafka.TopicUserRegistered)
	}
	event, ok := published.Payload.(core_kafka.UserRegisterEvent)
	if !ok {
		t.Fatalf("Publish() payload type = %T, want core_kafka.UserRegisterEvent", published.Payload)
	}
	if event.Username != validUsername || event.PhoneNumber != validPhone {
		t.Fatalf("Publish() unexpected event: %+v", event)
	}
}

func TestService_Register_PhoneAlreadyInUse(t *testing.T) {
	userRepo := &fakeUserRepo{
		getUserByPhoneNumberFunc: func(ctx context.Context, phoneNumber string) (core_domain.AuthUser, error) {
			// err == nil означает "пользователь с этим номером найден".
			return core_domain.AuthUser{ID: core_domain.NewAuthUser("x", validPhone, "h").ID}, nil
		},
	}
	svc := NewAuthService(userRepo, &fakeRefreshTokenRepo{}, []byte("secret"), time.Minute, &fakePublisher{}, testLogger())

	_, err := svc.Register(context.Background(), validUsername, validPhone, validPassword)
	if !errors.Is(err, core_errors.ErrPhoneNumberUse) {
		t.Fatalf("Register() error = %v, want wrapped %v", err, core_errors.ErrPhoneNumberUse)
	}
}

func TestService_Register_RepoErrorOtherThanNotFound(t *testing.T) {
	repoErr := errors.New("connection refused")
	userRepo := &fakeUserRepo{
		getUserByPhoneNumberFunc: func(ctx context.Context, phoneNumber string) (core_domain.AuthUser, error) {
			return core_domain.AuthUser{}, repoErr
		},
	}
	svc := NewAuthService(userRepo, &fakeRefreshTokenRepo{}, []byte("secret"), time.Minute, &fakePublisher{}, testLogger())

	_, err := svc.Register(context.Background(), validUsername, validPhone, validPassword)
	if err == nil {
		t.Fatalf("Register() expected error, got nil")
	}
	if errors.Is(err, core_errors.ErrPhoneNumberUse) {
		t.Fatalf("Register() should not report ErrPhoneNumberUse for an unrelated repo error")
	}
}

func TestService_Register_InvalidPassword(t *testing.T) {
	userRepo := &fakeUserRepo{
		getUserByPhoneNumberFunc: func(ctx context.Context, phoneNumber string) (core_domain.AuthUser, error) {
			return core_domain.AuthUser{}, phoneNotFoundErr()
		},
	}
	svc := NewAuthService(userRepo, &fakeRefreshTokenRepo{}, []byte("secret"), time.Minute, &fakePublisher{}, testLogger())

	_, err := svc.Register(context.Background(), validUsername, validPhone, "short")
	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("Register() error = %v, want wrapped %v", err, core_errors.ErrInvalidArgument)
	}
}

func TestService_Register_InvalidUserData(t *testing.T) {
	userRepo := &fakeUserRepo{
		getUserByPhoneNumberFunc: func(ctx context.Context, phoneNumber string) (core_domain.AuthUser, error) {
			return core_domain.AuthUser{}, phoneNotFoundErr()
		},
	}
	svc := NewAuthService(userRepo, &fakeRefreshTokenRepo{}, []byte("secret"), time.Minute, &fakePublisher{}, testLogger())

	// телефон без "+" не проходит AuthUser.Validate(), хотя сам пароль валиден.
	_, err := svc.Register(context.Background(), validUsername, "77001234567", validPassword)
	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("Register() error = %v, want wrapped %v", err, core_errors.ErrInvalidArgument)
	}
}

func TestService_Register_RegisterUserError(t *testing.T) {
	repoErr := errors.New("insert failed")
	userRepo := &fakeUserRepo{
		getUserByPhoneNumberFunc: func(ctx context.Context, phoneNumber string) (core_domain.AuthUser, error) {
			return core_domain.AuthUser{}, phoneNotFoundErr()
		},
		registerUserFunc: func(ctx context.Context, user core_domain.AuthUser) (core_domain.AuthUser, error) {
			return core_domain.AuthUser{}, repoErr
		},
	}
	svc := NewAuthService(userRepo, &fakeRefreshTokenRepo{}, []byte("secret"), time.Minute, &fakePublisher{}, testLogger())

	_, err := svc.Register(context.Background(), validUsername, validPhone, validPassword)
	if err == nil {
		t.Fatalf("Register() expected error, got nil")
	}
}

func TestService_Register_PublishError(t *testing.T) {
	publishErr := errors.New("kafka unavailable")
	userRepo := &fakeUserRepo{
		getUserByPhoneNumberFunc: func(ctx context.Context, phoneNumber string) (core_domain.AuthUser, error) {
			return core_domain.AuthUser{}, phoneNotFoundErr()
		},
		registerUserFunc: func(ctx context.Context, user core_domain.AuthUser) (core_domain.AuthUser, error) {
			return user, nil
		},
	}
	publisher := &fakePublisher{
		publishFunc: func(ctx context.Context, message core_kafka.Message) error {
			return publishErr
		},
	}
	svc := NewAuthService(userRepo, &fakeRefreshTokenRepo{}, []byte("secret"), time.Minute, publisher, testLogger())

	// В отличие от messenger-service, здесь публикация в Kafka — НЕ
	// best-effort: ошибка публикации проваливает весь Register(), хотя
	// пользователь уже создан в БД к этому моменту.
	_, err := svc.Register(context.Background(), validUsername, validPhone, validPassword)
	if err == nil {
		t.Fatalf("Register() expected error when publish fails, got nil")
	}
}
