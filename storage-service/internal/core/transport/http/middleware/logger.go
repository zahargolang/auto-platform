package core_middleware

import (
	core_logger "storage-service/internal/core/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logger кладёт логгер, обогащённый полями request_id и url, в контекст
// запроса — все последующие middleware и обработчики достают его через
// core_logger.FromContext и автоматически логируют эти поля вместе с
// остальными.
//
// Должен быть зарегистрирован ПОСЛЕ RequestID — иначе поле request_id
// будет пустым.
func Logger(log *core_logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(RequestIDHeader)

		l := log.With(
			zap.String("request_id", requestID),
			zap.String("url", c.Request.URL.String()),
		)

		ctx := core_logger.ToContext(c.Request.Context(), l)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
