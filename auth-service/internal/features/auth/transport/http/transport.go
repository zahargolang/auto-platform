package auth_transport_http

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	core_domain "github.com/zosinkin/social_network/internal/core/domain"
	core_middleware "github.com/zosinkin/social_network/internal/core/transport/http/middleware"
	auth_service "github.com/zosinkin/social_network/internal/features/auth/service"

	_ "github.com/zosinkin/social_network/docs" // swagger docs
)

type AuthHTTPHandler struct {
	authService Service
}

func NewAuthHTTPHandler(
	service Service,
) *AuthHTTPHandler {
	return &AuthHTTPHandler{
		authService: service,
	}
}

type Service interface {
	Register(
		ctx context.Context,
		username string,
		phoneNumber string,
		password string,
) (core_domain.AuthUser, error)

	Login(
		ctx context.Context,
		phoneNumber string,
		password string,
	) (string, error)

	LoginWithRefresh(
		ctx context.Context,
		phoneNumber string,
		password string,
		refreshTokenTTL time.Duration,
	) (string, string, error)

	RefreshAccessToken(
		ctx context.Context,
		refreshToken string,
	) (string, error)
}


func (h *AuthHTTPHandler) InitRoutes(authService *auth_service.Service) *gin.Engine {
	router := gin.Default()

	

	routes := router.Group("/api/auth")
	{
		routes.GET("/health", h.Health)
		routes.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		routes.POST("/register", h.Register)
		routes.POST("/login", h.Login)
		routes.POST("/refresh", h.RefreshToken)
	}

	authorized := router.Group("/api/auth")
	authorized.Use(core_middleware.AuthMiddleware(authService))
	{
		authorized.GET("/authorized", h.Authorized)
	}
	return router
}
