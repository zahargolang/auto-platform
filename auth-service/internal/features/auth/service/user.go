package auth_service

import (
	"context"
	"errors"
	"fmt"
	"time"

	core_domain "github.com/zosinkin/social_network/internal/core/domain"
	core_errors "github.com/zosinkin/social_network/internal/core/errors"
	core_postgres_pool "github.com/zosinkin/social_network/internal/core/repository/postgres/pool"
	core_kafka "github.com/zosinkin/social_network/internal/core/transport/kafka"
	"go.uber.org/zap"
)

// func (s *AuthService) Register(
// 	ctx context.Context,
// 	username string,
// 	phoneNumber string,
// 	password string,
// ) (core_domain.AuthUser, error) {
// 	op := "User.Service.CreateUser"

// 	_, err := s.userRepo.GetUserByPhoneNumber(ctx, phoneNumber)
// 	if err == nil {
// 		return core_domain.AuthUser{}, core_errors.ErrPhoneNumberUse
// 	}

// 	if !errors.Is(err, core_postgres_pool.ErrNoRows) {
// 		return core_domain.AuthUser{}, fmt.Errorf("%v: %v", op, err)
// 	}

// 	hashedPassword, err := HashPassword(password)
// 	if err != nil {
// 		return core_domain.AuthUser{}, fmt.Errorf("%v: %v", op, err)
// 	}
// 	user := core_domain.NewAuthUser(
// 		username,
// 		phoneNumber,
// 		hashedPassword,
// 	)

// 	registeredUser, err := s.userRepo.RegisterUser(ctx, user)
// 	if err != nil {
// 		return core_domain.AuthUser{}, fmt.Errorf("%v: %v", op, err)
// 	}

// 	return registeredUser, nil
// }


func (s *Service) Register(
	ctx context.Context,
	username string,
	phoneNumber string,
	password string,
) (core_domain.AuthUser, error) {
	op := "User.Service.CreateUser"

	_, err := s.userRepo.GetUserByPhoneNumber(ctx, phoneNumber)
	if err == nil {
		s.log.Debug("phone number already in use", zap.String("op", op))
		return core_domain.AuthUser{}, core_errors.ErrPhoneNumberUse
	}

	if !errors.Is(err, core_postgres_pool.ErrNoRows) {
		s.log.Error("get user by phone number error:", zap.String("op", op), zap.Error(err))
		return core_domain.AuthUser{}, fmt.Errorf("%v: %w", op, err)
	}

	if err := core_domain.ValidatePassword(password); err != nil {
		s.log.Debug("password invalid:", zap.String("op", op), zap.Error(err))
		return core_domain.AuthUser{}, fmt.Errorf("%v: %w", op, err)
	}

	hashedPassword, err := HashPassword(password)
	if err != nil {
		s.log.Error("password hash error:", zap.String("op", op), zap.Error(err))
		return core_domain.AuthUser{}, fmt.Errorf("%v: %w", op, err)
	}
	user := core_domain.NewAuthUser(
		username,
		phoneNumber,
		hashedPassword,
	)

	if err := user.Validate(); err != nil {
		s.log.Debug("invalid data:", zap.String("op", op), zap.Error(err))
		return core_domain.AuthUser{}, fmt.Errorf("%v: %w", op, err)
	}

	registeredUser, err := s.userRepo.RegisterUser(ctx, user)
	if err != nil {
		s.log.Error("register user error:", zap.String("op", op), zap.Error(err))
		return core_domain.AuthUser{}, fmt.Errorf("%v: %w", op, err)
	}


	//Create and publish Message
	event := core_kafka.UserRegisterEvent{
		ID: 			registeredUser.ID.String(),
		Username: 		registeredUser.Username,
		PhoneNumber: 	registeredUser.PhoneNumber,
	}
	
	message := core_kafka.NewMessage(
		core_kafka.TopicUserRegistered,
		event.ID,
		event,
	)
	err = s.publisher.Publish(
		ctx,
		message,
	)

	if err != nil {
		s.log.Error("Publish message error:", zap.String("op", op), zap.Error(err))
		return core_domain.AuthUser{}, fmt.Errorf("failed to publish message: %w", err)
	}

	return user, nil
}


func(s *Service) Login(
	ctx context.Context,
	phoneNumber string,
	password string,
) (string, error) {
	op := "AuthService.Service.Login"
	user, err := s.userRepo.GetUserByPhoneNumber(ctx, phoneNumber)
	if err != nil {
		s.log.Debug("get user by phone number error:", zap.String("op", op), zap.Error(err))
		return "", core_errors.ErrInvalidCredentials
	}

	if err := VerifyPassword(user.PasswordHash, password); err != nil {
		s.log.Debug("verify password error:", zap.String("op", op), zap.Error(err))
		return "", core_errors.ErrInvalidCredentials
	}

	token, err := s.generateAccessToken(&user)
	if err != nil {
		s.log.Error("generate access token error:", zap.String("op", op), zap.Error(err))
		return "", err
	}

	return token, nil
}


func (s *Service) LoginWithRefresh(
	ctx context.Context,
	phoneNumber string,
	password string,
	refreshTokenTTL time.Duration,
) (
	string,
	string,
	error,
) {
	op := "Auth.Service.LoginWithRefresh"

	user, err := s.userRepo.GetUserByPhoneNumber(ctx, phoneNumber)
	if err != nil {
		s.log.Error("Get user by phone number:", zap.String("op", op), zap.Error(err))
		return "", "", fmt.Errorf("%v: %w", op, core_errors.ErrInvalidCredentials)
	}

	if err := VerifyPassword(user.PasswordHash, password); err != nil {
		s.log.Error("Verify password error:", zap.String("op", op), zap.Error(err))
		return "", "", fmt.Errorf("%v: %w", op, core_errors.ErrInvalidCredentials)
	}

	accessToken, err := s.generateAccessToken(&user)
	if err != nil {
		s.log.Error("Generate access token error:", zap.String("op", op), zap.Error(err))
		return "", "", fmt.Errorf("%v: %w", op, err)
	}

	token := core_domain.NewRefreshToken(user.ID, refreshTokenTTL)

	refreshToken, err := s.refreshTokenRepo.CreateRefreshToken(ctx, token)
	if err != nil {
		s.log.Error("Create refresh token error:", zap.String("op", op), zap.Error(err))
		return "", "", fmt.Errorf("%v: %w", op, err)
	}

	return accessToken, refreshToken.Token, nil
}