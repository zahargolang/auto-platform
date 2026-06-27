package listings_service

import (
	"context"
	"fmt"

	core_domain "listing-service/internal/core/domain"
	"go.uber.org/zap"
)

func (s *Service) GetListings(ctx context.Context, filter core_domain.ListingFilter) ([]core_domain.Listing, error) {
	op := "ListingsService.GetListings"

	listings, err := s.repo.GetListings(ctx, filter)
	if err != nil {
		s.log.Error("failed to get listings", zap.String("op", op), zap.Error(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return listings, nil
}
