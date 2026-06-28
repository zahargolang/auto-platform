package core_middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORS добавляет заголовки Access-Control-Allow-* только для разрешённых
// origins. "*" в списке означает "разрешить любой origin" — то же
// соглашение, что уже используется для ALLOWED_ORIGINS в остальных
// сервисах проекта (см. messenger-service: transport/ws.CheckOrigin).
//
// Preflight-запросы (OPTIONS) обрабатываются сразу, не доходя до хендлера.
func CORS(allowedOriginsList []string) gin.HandlerFunc {
	
	//Используем мапу в качестве хранения origins, так как получаем константное время время поиска.
	allowedOrigins := make(map[string]struct{}, len(allowedOriginsList))
	for _, origin := range allowedOriginsList {
		allowedOrigins[origin] = struct{}{}
	}

	//на моммент разработки поставим заглушку
	_, allowAll := allowedOrigins["*"]
	//все до return выполняется один раз, при старте сервера, а не на каждый запрос
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		_, ok := allowedOrigins[origin]

		if allowAll || ok {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}
