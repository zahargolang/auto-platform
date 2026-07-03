package transport_http

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	core_logger "messenger-service/internal/core/logger"
	core_middleware "messenger-service/internal/core/transport/http/middleware"
	core_domain "messenger-service/internal/features/messenger/domain"

	_ "messenger-service/docs" // swagger docs
)

type Service interface {
	CreateOrGetConversation(ctx context.Context, listingID, requesterID, recipientID uuid.UUID) (core_domain.Conversation, error)
	ListConversations(ctx context.Context, userID uuid.UUID) ([]core_domain.Conversation, error)
	ListMessages(ctx context.Context, conversationID, requesterID uuid.UUID, page, limit int) ([]core_domain.Message, error)
}

// WSHandler — то немногое, что транспортному слою нужно от WS-хендлера:
// сама регистрация маршрута. Описано здесь, на стороне потребителя, чтобы
// этот пакет не зависел от пакета ws напрямую — *ws.Handler удовлетворяет
// этому интерфейсу просто по совпадению формы метода.
type WSHandler interface {
	ServeWS(c *gin.Context)
}

type Handler struct {
	service Service
	ws      WSHandler
}

func NewHandler(service Service, ws WSHandler) *Handler {
	return &Handler{service: service, ws: ws}
}

func (h *Handler) InitRoutes(
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
	

	public := router.Group("/api/messenger")
	{
		public.GET("/health", h.Health)
		public.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// отдельная защищённая группа для WS — Ingress матчит по path, поэтому
	// /ws не может остаться в публичной группе. Браузерный WebSocket API не
	// умеет ставить заголовок Authorization на хендшейк, поэтому токен на
	// этот путь долетает через ?token= в query (см. helm/.../ingress.yaml,
	// auth-snippet с фолбэком на $arg_token).
	ws := router.Group("/api/messenger/ws")
	ws.Use(core_middleware.RequireUserID())
	{
		ws.GET("", h.ws.ServeWS)
	}

	mine := router.Group("/api/messenger/mine")
	mine.Use(core_middleware.RequireUserID())
	{
		mine.POST("/conversation", h.CreateConversation)
		mine.GET("/conversations", h.ListConversations)
		mine.GET("/conversations/:id/messages", h.ListMessages)
	}

	return router
}
