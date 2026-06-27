package listings_repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	core_errors "listing-service/internal/core/errors"
	core_postgres_pool "listing-service/internal/core/repository/postgres/pool"
)

func (r *Repository) DeleteListing(ctx context.Context, id uuid.UUID) error {
	opCtx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	const query = `DELETE FROM listingservice.listings WHERE id = $1`

	tag, err := r.pool.Exec(opCtx, query, id)
	if err != nil {
		return fmt.Errorf("delete listing: %w", mapRepoError(err))
	}

	if tag.RowsAffected() == 0 {
		return core_errors.ErrNotFound
	}

	_ = core_postgres_pool.ErrNoRows
	return nil
}
