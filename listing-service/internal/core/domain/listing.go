package core_domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	core_errors "listing-service/internal/core/errors"
)

type ListingStatus string
type BodyType string
type FuelType string
type TransmissionType string

const (
	ListingStatusActive   ListingStatus = "active"
	ListingStatusInactive ListingStatus = "inactive"
	ListingStatusSold     ListingStatus = "sold"

	BodyTypeSedan     BodyType = "sedan"
	BodyTypeSUV       BodyType = "suv"
	BodyTypeHatchback BodyType = "hatchback"
	BodyTypeCoupe     BodyType = "coupe"
	BodyTypeWagon     BodyType = "wagon"
	BodyTypeMinivan   BodyType = "minivan"
	BodyTypePickup    BodyType = "pickup"

	FuelTypeGasoline FuelType = "gasoline"
	FuelTypeDiesel   FuelType = "diesel"
	FuelTypeElectric FuelType = "electric"
	FuelTypeHybrid   FuelType = "hybrid"
	FuelTypeLPG      FuelType = "lpg"

	TransmissionAutomatic TransmissionType = "automatic"
	TransmissionManual    TransmissionType = "manual"
	TransmissionRobot     TransmissionType = "robot"
	TransmissionVariator  TransmissionType = "variator"
)

type Listing struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Title       string
	Description string
	Price       int64
	Status      ListingStatus

	Make         string
	Model        string
	Year         int
	Mileage      int
	Color        string
	BodyType     BodyType
	FuelType     FuelType
	Transmission TransmissionType
	EngineVolume float64

	City   string
	Region string

	CreatedAt time.Time
	UpdatedAt time.Time
}

type ListingUpdate struct {
	Title        Nullable[string]
	Description  Nullable[string]
	Price        Nullable[int64]
	Status       Nullable[ListingStatus]
	Mileage      Nullable[int]
	Color        Nullable[string]
	City         Nullable[string]
	Region       Nullable[string]
}

type ListingFilter struct {
	Make         *string
	Model        *string
	YearFrom     *int
	YearTo       *int
	PriceFrom    *int64
	PriceTo      *int64
	MileageFrom  *int
	MileageTo    *int
	FuelType     *FuelType
	Transmission *TransmissionType
	BodyType     *BodyType
	City         *string
	Region       *string
	Status       *ListingStatus
	Page         int
	Limit        int
}

func NewListing(
	userID uuid.UUID,
	title string,
	description string,
	price int64,
	make_ string,
	model string,
	year int,
	mileage int,
	color string,
	bodyType BodyType,
	fuelType FuelType,
	transmission TransmissionType,
	engineVolume float64,
	city string,
	region string,
) Listing {
	return Listing{
		ID:           uuid.New(),
		UserID:       userID,
		Title:        title,
		Description:  description,
		Price:        price,
		Status:       ListingStatusActive,
		Make:         make_,
		Model:        model,
		Year:         year,
		Mileage:      mileage,
		Color:        color,
		BodyType:     bodyType,
		FuelType:     fuelType,
		Transmission: transmission,
		EngineVolume: engineVolume,
		City:         city,
		Region:       region,
	}
}

func (l Listing) Validate() error {
	titleLen := len([]rune(l.Title))
	if titleLen < 3 || titleLen > 200 {
		return fmt.Errorf("invalid `title` len: %d: %w", titleLen, core_errors.ErrInvalidArgument)
	}

	if l.Price <= 0 {
		return fmt.Errorf("price must be greater than 0: %w", core_errors.ErrInvalidArgument)
	}

	if l.Year < 1900 || l.Year > time.Now().Year()+1 {
		return fmt.Errorf("invalid `year`: %d: %w", l.Year, core_errors.ErrInvalidArgument)
	}

	if l.Mileage < 0 {
		return fmt.Errorf("mileage cannot be negative: %w", core_errors.ErrInvalidArgument)
	}

	if l.EngineVolume < 0 {
		return fmt.Errorf("engine volume cannot be negative: %w", core_errors.ErrInvalidArgument)
	}

	cityLen := len([]rune(l.City))
	if cityLen < 1 || cityLen > 100 {
		return fmt.Errorf("invalid `city` len: %d: %w", cityLen, core_errors.ErrInvalidArgument)
	}

	return nil
}
