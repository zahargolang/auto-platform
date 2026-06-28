package core_middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestIDHeader — заголовок для передачи уникального идентификатора запроса.
// Позволяет отслеживать один запрос через все логи и системы.
const RequestIDHeader = "X-Request-ID"

// RequestID обеспечивает каждый запрос уникальным идентификатором.
// Если клиент передаёт X-Request-ID — используем его (полезно для
// распределённой трассировки между сервисами). Иначе генерируем новый.
// Идентификатор добавляется и в заголовок ответа, чтобы клиент мог его
// использовать при обращении в поддержку/дебаге.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.NewString()
		}

		c.Request.Header.Set(RequestIDHeader, requestID)
		c.Writer.Header().Set(RequestIDHeader, requestID)

		c.Next()
	}
}
