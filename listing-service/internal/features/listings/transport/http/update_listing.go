package listings_transport_http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	core_domain "listing-service/internal/core/domain"
	core_errors "listing-service/internal/core/errors"
	"go.uber.org/zap"
)

//	@Summary		Обновить объявление
//	@Description	Обновляет объявление (требует авторизации, только владелец)
//	@Tags			listings
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string					true	"ID объявления"
//	@Param			request	body		UpdateListingRequest	true	"Поля для обновления"
//	@Success		200		{object}	ListingResponse
//	@Failure		400		{object}	map[string]string
//	@Failure		401		{object}	map[string]string
//	@Failure		403		{object}	map[string]string
//	@Failure		404		{object}	map[string]string
//	@Router			/{id} [patch]
func (h *ListingsHandler) UpdateListing(c *gin.Context) {
	userID, ok := c.MustGet("user_id").(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid listing id"})
		return
	}

	var req UpdateListingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	update := core_domain.ListingUpdate{}

	if req.Title != nil {
		update.Title = core_domain.Nullable[string]{Value: req.Title, Set: true}
	}
	if req.Description != nil {
		update.Description = core_domain.Nullable[string]{Value: req.Description, Set: true}
	}
	if req.Price != nil {
		update.Price = core_domain.Nullable[int64]{Value: req.Price, Set: true}
	}
	if req.Status != nil {
		update.Status = core_domain.Nullable[core_domain.ListingStatus]{Value: req.Status, Set: true}
	}
	if req.Mileage != nil {
		update.Mileage = core_domain.Nullable[int]{Value: req.Mileage, Set: true}
	}
	if req.Color != nil {
		update.Color = core_domain.Nullable[string]{Value: req.Color, Set: true}
	}
	if req.City != nil {
		update.City = core_domain.Nullable[string]{Value: req.City, Set: true}
	}
	if req.Region != nil {
		update.Region = core_domain.Nullable[string]{Value: req.Region, Set: true}
	}

	listing, err := h.service.UpdateListing(c.Request.Context(), id, userID, update)
	if err != nil {
		if errors.Is(err, core_errors.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "listing not found"})
			return
		}
		if errors.Is(err, core_errors.ErrForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		h.log.Error("update listing error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, toListingResponse(listing))
}
