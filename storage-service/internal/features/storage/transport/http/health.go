package transport_http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

//	@Summary	Health check
//	@Tags		storage
//	@Produce	json
//	@Success	200	{object}	map[string]string
//	@Router		/health [get]
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
