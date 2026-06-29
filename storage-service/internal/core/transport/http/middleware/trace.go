package core_middleware

import (
	core_logger "storage-service/internal/core/logger"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Trace логирует начало и конец обработки каждого запроса — единая точка
// access-лога вместо разрозненных вызовов внутри отдельных хендлеров.
// Код статуса читается через c.Writer.Status() — Gin сам его отслеживает,
// в отличие от чистого net/http, где для этого нужна обёртка над
// http.ResponseWriter.
//
// Должен идти ПОСЛЕ Logger — нужен логгер из контекста запроса.
func Trace() gin.HandlerFunc {
	return func(c *gin.Context) {
		log := core_logger.FromContext(c.Request.Context())
		before := time.Now()

		log.Debug(
			">>> incoming HTTP request",
			zap.String("http_method", c.Request.Method),
			zap.Time("time", before.UTC()),
		)

		c.Next()

		log.Debug(
			"<<< done HTTP request",
			zap.Int("status_code", c.Writer.Status()),
			zap.Duration("latency", time.Since(before)),
		)
	}
}
