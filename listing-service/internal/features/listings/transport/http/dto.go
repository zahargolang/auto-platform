package listings_transport_http

import (
	"time"

	"github.com/google/uuid"
	core_domain "listing-service/internal/core/domain"
)

type CreateListingRequest struct {
	Title        string                   `json:"title"         binding:"required,min=3,max=200"`
	Description  string                   `json:"description"   binding:"required"`
	Price        int64                    `json:"price"         binding:"required,gt=0"`
	Make         string                   `json:"make"          binding:"required"`
	Model        string                   `json:"model"         binding:"required"`
	Year         int                      `json:"year"          binding:"required,gt=1900"`
	Mileage      int                      `json:"mileage"       binding:"min=0"`
	Color        string                   `json:"color"         binding:"required"`
	BodyType     core_domain.BodyType     `json:"body_type"     binding:"required"`
	FuelType     core_domain.FuelType     `json:"fuel_type"     binding:"required"`
	Transmission core_domain.TransmissionType `json:"transmission" binding:"required"`
	EngineVolume float64                  `json:"engine_volume" binding:"min=0"`
	City         string                   `json:"city"          binding:"required"`
	Region       string                   `json:"region"        binding:"required"`
}

type UpdateListingRequest struct {
	Title       *string                        `json:"title"`
	Description *string                        `json:"description"`
	Price       *int64                         `json:"price"`
	Status      *core_domain.ListingStatus     `json:"status"`
	Mileage     *int                           `json:"mileage"`
	Color       *string                        `json:"color"`
	City        *string                        `json:"city"`
	Region      *string                        `json:"region"`
}

type ListingResponse struct {
	ID           uuid.UUID                    `json:"id"`
	UserID       uuid.UUID                    `json:"user_id"`
	Title        string                       `json:"title"`
	Description  string                       `json:"description"`
	Price        int64                        `json:"price"`
	Status       core_domain.ListingStatus    `json:"status"`
	Make         string                       `json:"make"`
	Model        string                       `json:"model"`
	Year         int                          `json:"year"`
	Mileage      int                          `json:"mileage"`
	Color        string                       `json:"color"`
	BodyType     core_domain.BodyType         `json:"body_type"`
	FuelType     core_domain.FuelType         `json:"fuel_type"`
	Transmission core_domain.TransmissionType `json:"transmission"`
	EngineVolume float64                      `json:"engine_volume"`
	City         string                       `json:"city"`
	Region       string                       `json:"region"`
	CreatedAt    time.Time                    `json:"created_at"`
	UpdatedAt    time.Time                    `json:"updated_at"`
}

type ListingsResponse struct {
	Items []ListingResponse `json:"items"`
	Total int               `json:"total"`
}

func toListingResponse(l core_domain.Listing) ListingResponse {
	return ListingResponse{
		ID:           l.ID,
		UserID:       l.UserID,
		Title:        l.Title,
		Description:  l.Description,
		Price:        l.Price,
		Status:       l.Status,
		Make:         l.Make,
		Model:        l.Model,
		Year:         l.Year,
		Mileage:      l.Mileage,
		Color:        l.Color,
		BodyType:     l.BodyType,
		FuelType:     l.FuelType,
		Transmission: l.Transmission,
		EngineVolume: l.EngineVolume,
		City:         l.City,
		Region:       l.Region,
		CreatedAt:    l.CreatedAt,
		UpdatedAt:    l.UpdatedAt,
	}
}
