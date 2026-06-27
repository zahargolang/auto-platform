package listings_transport_http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

//	@Summary		Создать объявление
//	@Description	Создаёт новое объявление о продаже автомобиля (требует авторизации)
//	@Tags			listings
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		CreateListingRequest	true	"Данные объявления"
//	@Success		201		{object}	ListingResponse
//	@Failure		400		{object}	map[string]string
//	@Failure		401		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/ [post]
func (h *ListingsHandler) CreateListing(c *gin.Context) {
	userID, ok := c.MustGet("user_id").(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req CreateListingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	listing, err := h.service.CreateListing(
		c.Request.Context(),
		userID,
		req.Title,
		req.Description,
		req.Price,
		req.Make,
		req.Model,
		req.Year,
		req.Mileage,
		req.Color,
		req.BodyType,
		req.FuelType,
		req.Transmission,
		req.EngineVolume,
		req.City,
		req.Region,
	)
	if err != nil {
		h.log.Error("create listing error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, toListingResponse(listing))
}
