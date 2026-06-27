package listings_service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	core_errors "listing-service/internal/core/errors"
	"go.uber.org/zap"
)

func (s *Service) DeleteListing(ctx context.Context, id uuid.UUID, requesterID uuid.UUID) error {
	op := "ListingsService.DeleteListing"

	existing, err := s.repo.GetListingByID(ctx, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if existing.UserID != requesterID {
		return fmt.Errorf("%s: %w", op, core_errors.ErrForbidden)
	}

	if err := s.repo.DeleteListing(ctx, id); err != nil {
		s.log.Error("failed to delete listing", zap.String("op", op), zap.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
