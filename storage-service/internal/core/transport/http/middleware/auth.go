package core_middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequireUserID читает X-User-Id — заголовок, который nginx Ingress
// проставляет из ответа auth-service после успешной проверки токена
// (auth-response-headers, см. helm/.../templates/ingress.yaml). Локальную
// JWT-проверку здесь больше не делаем — её один раз на весь кластер делает
// auth-service за gateway, остальные сервисы доверяют этому заголовку.
//
// Если запрос пришёл не через gateway (например, напрямую к Pod'у) или
// заголовка нет — отвечаем 401.
func RequireUserID() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := uuid.Parse(c.GetHeader("X-User-Id"))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}
