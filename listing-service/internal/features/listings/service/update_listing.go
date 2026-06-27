package listings_service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	core_domain "listing-service/internal/core/domain"
	core_errors "listing-service/internal/core/errors"
	"go.uber.org/zap"
)

func (s *Service) UpdateListing(
	ctx context.Context,
	id uuid.UUID,
	requesterID uuid.UUID,
	update core_domain.ListingUpdate,
) (core_domain.Listing, error) {
	op := "ListingsService.UpdateListing"

	existing, err := s.repo.GetListingByID(ctx, id)
	if err != nil {
		return core_domain.Listing{}, fmt.Errorf("%s: %w", op, err)
	}

	if existing.UserID != requesterID {
		return core_domain.Listing{}, fmt.Errorf("%s: %w", op, core_errors.ErrForbidden)
	}

	updated, err := s.repo.UpdateListing(ctx, id, update)
	if err != nil {
		s.log.Error("failed to update listing", zap.String("op", op), zap.Error(err))
		return core_domain.Listing{}, fmt.Errorf("%s: %w", op, err)
	}

	return updated, nil
}
