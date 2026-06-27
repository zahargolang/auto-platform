package listings_service

import (
	"context"

	"github.com/google/uuid"
	core_domain "listing-service/internal/core/domain"
	core_logger "listing-service/internal/core/logger"
	core_kafka "listing-service/internal/core/transport/kafka"
)

type Service struct {
	repo      ListingRepo
	publisher EventPublisher
	log       *core_logger.Logger
}

type ListingRepo interface {
	CreateListing(ctx context.Context, listing core_domain.Listing) (core_domain.Listing, error)
	GetListingByID(ctx context.Context, id uuid.UUID) (core_domain.Listing, error)
	GetListings(ctx context.Context, filter core_domain.ListingFilter) ([]core_domain.Listing, error)
	UpdateListing(ctx context.Context, id uuid.UUID, update core_domain.ListingUpdate) (core_domain.Listing, error)
	DeleteListing(ctx context.Context, id uuid.UUID) error
}

type EventPublisher interface {
	Publish(ctx context.Context, message core_kafka.Message) error
}

func NewService(repo ListingRepo, publisher EventPublisher, log *core_logger.Logger) *Service {
	return &Service{
		repo:      repo,
		publisher: publisher,
		log:       log,
	}
}
