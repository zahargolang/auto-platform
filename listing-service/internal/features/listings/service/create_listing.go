package listings_service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	core_domain "listing-service/internal/core/domain"
	core_kafka "listing-service/internal/core/transport/kafka"
	"go.uber.org/zap"
)

func (s *Service) CreateListing(
	ctx context.Context,
	userID uuid.UUID,
	title string,
	description string,
	price int64,
	make_ string,
	model string,
	year int,
	mileage int,
	color string,
	bodyType core_domain.BodyType,
	fuelType core_domain.FuelType,
	transmission core_domain.TransmissionType,
	engineVolume float64,
	city string,
	region string,
) (core_domain.Listing, error) {
	op := "ListingsService.CreateListing"

	listing := core_domain.NewListing(
		userID, title, description, price,
		make_, model, year, mileage, color,
		bodyType, fuelType, transmission, engineVolume,
		city, region,
	)

	if err := listing.Validate(); err != nil {
		return core_domain.Listing{}, fmt.Errorf("%s: %w", op, err)
	}

	created, err := s.repo.CreateListing(ctx, listing)
	if err != nil {
		s.log.Error("failed to create listing", zap.String("op", op), zap.Error(err))
		return core_domain.Listing{}, fmt.Errorf("%s: %w", op, err)
	}

	event := core_kafka.ListingCreatedEvent{
		ID:     created.ID.String(),
		UserID: created.UserID.String(),
		Title:  created.Title,
		Price:  created.Price,
		Make:   created.Make,
		Model:  created.Model,
		Year:   created.Year,
		City:   created.City,
	}

	if err := s.publisher.Publish(ctx, core_kafka.NewMessage(
		core_kafka.TopicListingCreated,
		created.ID.String(),
		event,
	)); err != nil {
		s.log.Warn("failed to publish listing.created event", zap.String("op", op), zap.Error(err))
	}

	return created, nil
}
