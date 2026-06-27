package listings_service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	core_domain "listing-service/internal/core/domain"
	"go.uber.org/zap"
)

func (s *Service) GetListingByID(ctx context.Context, id uuid.UUID) (core_domain.Listing, error) {
	op := "ListingsService.GetListingByID"

	listing, err := s.repo.GetListingByID(ctx, id)
	if err != nil {
		s.log.Error("failed to get listing", zap.String("op", op), zap.Error(err))
		return core_domain.Listing{}, fmt.Errorf("%s: %w", op, err)
	}

	return listing, nil
}
