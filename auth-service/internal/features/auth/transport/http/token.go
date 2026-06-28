package auth_transport_http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	core_errors "github.com/zosinkin/social_network/internal/core/errors"
	core_logger "github.com/zosinkin/social_network/internal/core/logger"
)



// RefreshToken выдаёт новый access-токен по валидному refresh-токену.
//
//	@Summary		Обновление access-токена
//	@Description	Выдаёт новый access-токен по валидному refresh-токену
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		RefreshRequest	true	"Refresh-токен"
//	@Success		200		{object}	RefreshResponse
//	@Failure		400		{object}	map[string]string
//	@Failure		401		{object}	map[string]string
//	@Router			/refresh [post]
func (h *AuthHTTPHandler) RefreshToken(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Invalid request payload": err.Error(),
		})
		return
	}

	token, err := h.authService.RefreshAccessToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		log := core_logger.FromContext(c.Request.Context())
		if errors.Is(err, core_errors.ErrInvalidToken) || errors.Is(err, core_errors.ErrExpiredToken) {
			log.Debug("refresh token: invalid or expired", zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{
				"Invalid or expired refresh token": err.Error(),
			})
			return
		} else {
			log.Error("refresh token: unexpected error", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"Internal server error": err.Error(),
			})
		}
		return
	}

	response := RefreshResponse{Token: token}

	c.JSON(http.StatusOK, response)
}
