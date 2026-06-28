package auth_service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	core_domain "github.com/zosinkin/social_network/internal/core/domain"
	core_errors "github.com/zosinkin/social_network/internal/core/errors"
	"go.uber.org/zap"
)


func (s *Service) generateAccessToken(user *core_domain.AuthUser) (string, error) {
	op := "Auth.Service.generateAccessToken"
	exppirationTime := time.Now().Add(s.accessTokenTTL)

	claims := jwt.MapClaims{
		"sub": 			user.ID.String(),
		"username": 	user.Username,
		"phoneNumber": 	user.PhoneNumber,
		"exp": 			exppirationTime.Unix(),
		"iat": 			time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		s.log.Error("Generate access token error", zap.String("op", op), zap.Error(err))
		return "", err
	}
	return tokenString, nil
}


func (s *Service) RefreshAccessToken(
	ctx context.Context,
	refreshTokenString string,
) (string, error) {
	op := "AuthService.Service.refreshAccessToken"
	token, err := s.refreshTokenRepo.GetRefreshToken(ctx, refreshTokenString)
	if err != nil {
		s.log.Debug("get refresh token error:", zap.String("op", op), zap.Error(err))
		return "", core_errors.ErrInvalidToken
	}

	if token.Revoked {
		s.log.Debug("token is revoked:", zap.String("op", op))
		return "", core_errors.ErrInvalidToken
	}
	
	if time.Now().After(token.ExpiresAt) {
		s.log.Debug("Token is Expired", zap.String("op", op), zap.Error(err))
		return "", core_errors.ErrExpiredToken
	}

	user, err := s.userRepo.GetUserByID(ctx, token.UserID)
	if err != nil {
		s.log.Error("get user by id error:", zap.String("op", op), zap.Error(err))
		return "", fmt.Errorf("%v: %w", op, err)
	}

	accessToken, err := s.generateAccessToken(&user)
	if err !=  nil {
		s.log.Error("generate access token error:", zap.String("op", op), zap.Error(err))
		return "", fmt.Errorf("%v: %w", op, err)
	}

	return accessToken, nil
}


func (s *Service) ValidateToken(tokenString string) (jwt.MapClaims, error) {
	op := "AuthService.Service.ValidateToken"
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			s.log.Debug("unexpected token signing method:", zap.String("op", op))
			return nil, core_errors.ErrInvalidToken
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			s.log.Debug("token is expired:", zap.String("op", op))
			return nil, core_errors.ErrExpiredToken
		}
		return nil, core_errors.ErrInvalidToken
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, core_errors.ErrInvalidToken 
}
