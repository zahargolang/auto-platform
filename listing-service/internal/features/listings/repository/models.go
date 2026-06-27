package listings_repository

import (
	"time"

	"github.com/google/uuid"
	core_domain "listing-service/internal/core/domain"
)

type listingRow struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	Title        string
	Description  string
	Price        int64
	Status       string
	Make         string
	Model        string
	Year         int
	Mileage      int
	Color        string
	BodyType     string
	FuelType     string
	Transmission string
	EngineVolume float64
	City         string
	Region       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (r listingRow) toDomain() core_domain.Listing {
	return core_domain.Listing{
		ID:           r.ID,
		UserID:       r.UserID,
		Title:        r.Title,
		Description:  r.Description,
		Price:        r.Price,
		Status:       core_domain.ListingStatus(r.Status),
		Make:         r.Make,
		Model:        r.Model,
		Year:         r.Year,
		Mileage:      r.Mileage,
		Color:        r.Color,
		BodyType:     core_domain.BodyType(r.BodyType),
		FuelType:     core_domain.FuelType(r.FuelType),
		Transmission: core_domain.TransmissionType(r.Transmission),
		EngineVolume: r.EngineVolume,
		City:         r.City,
		Region:       r.Region,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
	}
}
