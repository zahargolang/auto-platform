package transport_http

import (
	"context"
	core_domain "user-service/internal/core/domain"
	core_logger "user-service/internal/core/logger"
	core_middleware "user-service/internal/core/transport/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "user-service/docs" // swagger docs
)

type HTTPHandler struct {
	service Service
}

func NewHTTPHandler(
	userRepo Service,
) *HTTPHandler {
	return &HTTPHandler{
		service: userRepo,
	}
}

type Service interface {
	SaveUser(
		ctx context.Context,
		ID uuid.UUID,
		username string,
		phoneNumber string,
	) (error)

	GetUser(
		ctx context.Context,
		id uuid.UUID,
	) (core_domain.User, error)
}


func (h *HTTPHandler) InitRoutes(
	log *core_logger.Logger,
	allowedOrigins []string,
) *gin.Engine {
	router := gin.Default()
	
	router.Use(
		core_middleware.CORS(allowedOrigins),
		core_middleware.RequestID(),
		core_middleware.Logger(log),
		core_middleware.Trace(),
		core_middleware.Panic(),
	)

	routes := router.Group("/api/user")
	{
		routes.GET("/health", h.HealthCheck)
		routes.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		routes.GET("/:id", h.GetProfile)
		//routes.POST("/save", h.SaveUser)
	}

	// отдельный путь — чтобы Ingress мог защитить его через auth-url по path,
	// не пересекаясь с публичным GET /api/user/:id
	me := router.Group("/api/user/me")
	me.Use(core_middleware.RequireUserID())
	{
		me.GET("", h.GetMyProfile)
	}

	return router
}