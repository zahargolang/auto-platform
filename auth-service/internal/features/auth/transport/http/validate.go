package auth_transport_http

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

// extractToken достаёт access-токен из запроса. Обычные запросы несут его в
// Authorization: Bearer — но браузерный WebSocket API не умеет ставить
// кастомные заголовки на хендшейк, поэтому для /api/messenger/ws токен едет
// через ?token= в URL. nginx не позволяет достать query string оригинального
// запроса внутри auth_request-сабреквеста через $arg_* (он видит только URI
// самого сабреквеста, без query) — но X-Original-URL он шлёт всегда
// автоматически (не через snippet), и в нём лежит $request_uri оригинального
// запроса целиком, с query string. Поэтому фолбэк — отсюда, а не из nginx.
func extractToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	parts := strings.Split(authHeader, " ")
	if len(parts) == 2 && parts[0] == "Bearer" && parts[1] != "" {
		return parts[1]
	}

	if originalURL := c.GetHeader("X-Original-URL"); originalURL != "" {
		if u, err := url.Parse(originalURL); err == nil {
			return u.Query().Get("token")
		}
	}

	return ""
}

// ValidateAuth — цель nginx-аннотации auth-url на "защищённом" Ingress
// (см. helm/.../templates/ingress.yaml). Какие пути считаются защищёнными
// решает сам Ingress через раздельные path-правила: сюда долетают только
// запросы к уже отобранным protected-путям, поэтому здесь только проверка
// токена — без какой-либо публичный/защищённый логики.
//
//	@Summary		Проверка авторизации для API Gateway (nginx auth_request)
//	@Tags			auth
//	@Produce		json
//	@Success		200
//	@Failure		401	{object}	map[string]string
//	@Router			/validate [get]
func (h *AuthHTTPHandler) ValidateAuth(c *gin.Context) {
	tokenString := extractToken(c)
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
		return
	}

	claims, err := h.authService.ValidateToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
		return
	}

	userID, ok := claims["sub"].(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
		return
	}

	c.Header("X-User-Id", userID)
	c.Status(http.StatusOK)
}
