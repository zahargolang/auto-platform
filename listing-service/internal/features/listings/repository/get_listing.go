package listings_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	core_domain "listing-service/internal/core/domain"
	core_errors "listing-service/internal/core/errors"
	core_postgres_pool "listing-service/internal/core/repository/postgres/pool"
)

func (r *Repository) GetListingByID(ctx context.Context, id uuid.UUID) (core_domain.Listing, error) {
	opCtx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	const query = `
		SELECT id, user_id, title, description, price, status,
			make, model, year, mileage, color, body_type, fuel_type, transmission, engine_volume,
			city, region, created_at, updated_at
		FROM listingservice.listings
		WHERE id = $1
	`

	var row listingRow
	err := r.pool.QueryRow(opCtx, query, id).Scan(
		&row.ID, &row.UserID, &row.Title, &row.Description, &row.Price, &row.Status,
		&row.Make, &row.Model, &row.Year, &row.Mileage, &row.Color, &row.BodyType,
		&row.FuelType, &row.Transmission, &row.EngineVolume,
		&row.City, &row.Region, &row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return core_domain.Listing{}, core_errors.ErrNotFound
		}
		return core_domain.Listing{}, fmt.Errorf("get listing by id: %w", err)
	}

	return row.toDomain(), nil
}
