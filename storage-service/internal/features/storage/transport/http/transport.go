package transport_http

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	core_logger "storage-service/internal/core/logger"
	core_middleware "storage-service/internal/core/transport/http/middleware"
	storage_service "storage-service/internal/features/storage/service"
)

type Service interface {
	NewUploadURL(ctx context.Context, ownerID uuid.UUID, filename string) (storage_service.UploadURL, error)
}

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) InitRoutes(log *core_logger.Logger, allowedOrigins []string) *gin.Engine {
	router := gin.New()

	router.Use(
		core_middleware.CORS(allowedOrigins),
		core_middleware.RequestID(),
		core_middleware.Logger(log),
		core_middleware.Trace(),
		core_middleware.Panic(),
	)

	public := router.Group("/api/storage")
	{
		public.GET("/health", h.Health)
	}

	// отдельный путь — чтобы Ingress мог защитить именно эту операцию через
	// auth-url по path (тот же паттерн, что и /api/listings/mine)
	mine := router.Group("/api/storage/mine")
	mine.Use(core_middleware.RequireUserID())
	{
		mine.POST("/upload-url", h.NewUploadURL)
	}

	return router
}
