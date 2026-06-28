package auth_service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	core_domain "github.com/zosinkin/social_network/internal/core/domain"
	core_errors "github.com/zosinkin/social_network/internal/core/errors"
)

func (s *Service) generateAccessToken(user *core_domain.AuthUser) (string, error) {
	exppirationTime := time.Now().Add(s.accessTokenTTL)

	claims := jwt.MapClaims{
		"sub":         user.ID.String(),
		"username":    user.Username,
		"phoneNumber": user.PhoneNumber,
		"exp":         exppirationTime.Unix(),
		"iat":         time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
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
		return "", core_errors.ErrInvalidToken
	}

	if token.Revoked {
		return "", core_errors.ErrInvalidToken
	}

	if time.Now().After(token.ExpiresAt) {
		return "", core_errors.ErrExpiredToken
	}

	user, err := s.userRepo.GetUserByID(ctx, token.UserID)
	if err != nil {
		return "", fmt.Errorf("%v: %w", op, err)
	}

	accessToken, err := s.generateAccessToken(&user)
	if err != nil {
		return "", fmt.Errorf("%v: %w", op, err)
	}

	return accessToken, nil
}

func (s *Service) ValidateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, core_errors.ErrInvalidToken
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, core_errors.ErrExpiredToken
		}
		return nil, core_errors.ErrInvalidToken
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, core_errors.ErrInvalidToken
}
