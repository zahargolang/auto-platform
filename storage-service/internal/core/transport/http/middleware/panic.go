package core_middleware

import (
	core_logger "storage-service/internal/core/logger"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Panic перехватывает панику внутри обработчика и возвращает HTTP 500
// вместо того, чтобы уронить горутину запроса и оставить клиента без
// ответа. Использует defer + recover — стандартный паттерн Go.
//
// Регистрируется последним (ближе всего к Handler), чтобы внешние
// middleware (в первую очередь Trace) успели нормально отработать свою
// часть "после next" уже после восстановления — например, залогировать
// итоговый статус 500 в access-лог.
func Panic() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if p := recover(); p != nil {
				log := core_logger.FromContext(c.Request.Context())
				log.Error("unexpected panic while handling HTTP request", zap.Any("panic", p))

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "internal server error",
				})
			}
		}()

		c.Next()
	}
}
