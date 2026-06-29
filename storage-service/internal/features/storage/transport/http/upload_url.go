package transport_http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	core_logger "storage-service/internal/core/logger"
)

// NewUploadURL — единственная защищённая операция этого сервиса: выдать
// presigned PUT-ссылку, по которой фронтенд сам зальёт файл прямо в S3,
// минуя backend (тело файла никогда не проходит через этот сервис).
//
//	@Summary		Получить ссылку для загрузки фото
//	@Description	Возвращает presigned PUT URL (загрузить файл) и публичный URL (сохранить в listing)
//	@Tags			storage
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		UploadURLRequest	true	"Имя файла"
//	@Success		200		{object}	UploadURLResponse
//	@Failure		400		{object}	map[string]string
//	@Failure		401		{object}	map[string]string
//	@Router			/mine/upload-url [post]
func (h *Handler) NewUploadURL(c *gin.Context) {
	log := core_logger.FromContext(c.Request.Context())

	userID := c.MustGet("user_id").(uuid.UUID)

	var req UploadURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.service.NewUploadURL(c.Request.Context(), userID, req.Filename)
	if err != nil {
		log.Error("new upload url error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, UploadURLResponse{
		UploadURL: result.UploadURL,
		PublicURL: result.PublicURL,
	})
}
