package listings_transport_http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	core_domain "listing-service/internal/core/domain"
	"go.uber.org/zap"
)

//	@Summary		Список объявлений
//	@Description	Возвращает список объявлений с фильтрацией и пагинацией
//	@Tags			listings
//	@Produce		json
//	@Param			make			query		string	false	"Марка"
//	@Param			model			query		string	false	"Модель"
//	@Param			city			query		string	false	"Город"
//	@Param			region			query		string	false	"Регион"
//	@Param			fuel_type		query		string	false	"Тип топлива"
//	@Param			transmission	query		string	false	"Тип трансмиссии"
//	@Param			body_type		query		string	false	"Тип кузова"
//	@Param			year_from		query		int		false	"Год от"
//	@Param			year_to			query		int		false	"Год до"
//	@Param			price_from		query		int		false	"Цена от"
//	@Param			price_to		query		int		false	"Цена до"
//	@Param			mileage_from	query		int		false	"Пробег от"
//	@Param			mileage_to		query		int		false	"Пробег до"
//	@Param			page			query		int		false	"Страница"
//	@Param			limit			query		int		false	"Размер страницы"
//	@Success		200				{object}	ListingsResponse
//	@Failure		500				{object}	map[string]string
//	@Router			/ [get]
func (h *ListingsHandler) GetListings(c *gin.Context) {
	filter := core_domain.ListingFilter{
		Page:  1,
		Limit: 20,
	}

	if v := c.Query("make"); v != "" {
		filter.Make = &v
	}
	if v := c.Query("model"); v != "" {
		filter.Model = &v
	}
	if v := c.Query("city"); v != "" {
		filter.City = &v
	}
	if v := c.Query("region"); v != "" {
		filter.Region = &v
	}
	if v := c.Query("fuel_type"); v != "" {
		ft := core_domain.FuelType(v)
		filter.FuelType = &ft
	}
	if v := c.Query("transmission"); v != "" {
		tr := core_domain.TransmissionType(v)
		filter.Transmission = &tr
	}
	if v := c.Query("body_type"); v != "" {
		bt := core_domain.BodyType(v)
		filter.BodyType = &bt
	}
	if v, err := strconv.Atoi(c.Query("year_from")); err == nil {
		filter.YearFrom = &v
	}
	if v, err := strconv.Atoi(c.Query("year_to")); err == nil {
		filter.YearTo = &v
	}
	if v, err := strconv.ParseInt(c.Query("price_from"), 10, 64); err == nil {
		filter.PriceFrom = &v
	}
	if v, err := strconv.ParseInt(c.Query("price_to"), 10, 64); err == nil {
		filter.PriceTo = &v
	}
	if v, err := strconv.Atoi(c.Query("mileage_from")); err == nil {
		filter.MileageFrom = &v
	}
	if v, err := strconv.Atoi(c.Query("mileage_to")); err == nil {
		filter.MileageTo = &v
	}
	if v, err := strconv.Atoi(c.Query("page")); err == nil && v > 0 {
		filter.Page = v
	}
	if v, err := strconv.Atoi(c.Query("limit")); err == nil && v > 0 {
		filter.Limit = v
	}

	listings, err := h.service.GetListings(c.Request.Context(), filter)
	if err != nil {
		h.log.Error("get listings error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	items := make([]ListingResponse, 0, len(listings))
	for _, l := range listings {
		items = append(items, toListingResponse(l))
	}

	c.JSON(http.StatusOK, ListingsResponse{
		Items: items,
		Total: len(items),
	})
}
