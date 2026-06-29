package auth_transport_http

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	core_domain "github.com/zosinkin/social_network/internal/core/domain"
	core_logger "github.com/zosinkin/social_network/internal/core/logger"
	core_middleware "github.com/zosinkin/social_network/internal/core/transport/http/middleware"

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

	ValidateToken(tokenString string) (jwt.MapClaims, error)
}


func (h *AuthHTTPHandler) InitRoutes(
	log *core_logger.Logger,
) *gin.Engine {
	// gin.New() (не gin.Default()) — иначе встроенные Logger()/Recovery()
	// дублировали бы наши же Trace()/Panic() ниже.
	router := gin.Default()

	//регистрируем middleware в указанном порядке, это и есть порядок выполнения на входе
	//каждый следующий вызывается через с.Next()
	router.Use(
		core_middleware.RequestID(),
		core_middleware.Logger(log),
		core_middleware.Trace(),
		core_middleware.Panic(),
	)

	routes := router.Group("/api/auth")
	{
		routes.GET("/health", h.Health)
		routes.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		routes.POST("/register", h.Register)
		routes.POST("/login", h.Login)
		routes.POST("/refresh", h.RefreshToken)
		// цель nginx auth-url на "защищённом" Ingress (см. helm); какие пути
		// туда попадают решает сам Ingress, а не этот хендлер
		routes.GET("/validate", h.ValidateAuth)
	}

	authorized := router.Group("/api/auth")
	{
		authorized.GET("/authorized", h.Authorized)
	}
	return router
}
