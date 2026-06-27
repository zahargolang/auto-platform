package listings_repository

import (
	"context"
	"fmt"

	core_domain "listing-service/internal/core/domain"
	core_errors "listing-service/internal/core/errors"
	core_postgres_pool "listing-service/internal/core/repository/postgres/pool"
)

func (r *Repository) CreateListing(ctx context.Context, listing core_domain.Listing) (core_domain.Listing, error) {
	opCtx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	const query = `
		INSERT INTO listingservice.listings (
			id, user_id, title, description, price, status,
			make, model, year, mileage, color, body_type, fuel_type, transmission, engine_volume,
			city, region
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11, $12, $13, $14, $15,
			$16, $17
		)
		RETURNING id, user_id, title, description, price, status,
			make, model, year, mileage, color, body_type, fuel_type, transmission, engine_volume,
			city, region, created_at, updated_at
	`

	var row listingRow
	err := r.pool.QueryRow(opCtx, query,
		listing.ID,
		listing.UserID,
		listing.Title,
		listing.Description,
		listing.Price,
		listing.Status,
		listing.Make,
		listing.Model,
		listing.Year,
		listing.Mileage,
		listing.Color,
		listing.BodyType,
		listing.FuelType,
		listing.Transmission,
		listing.EngineVolume,
		listing.City,
		listing.Region,
	).Scan(
		&row.ID, &row.UserID, &row.Title, &row.Description, &row.Price, &row.Status,
		&row.Make, &row.Model, &row.Year, &row.Mileage, &row.Color, &row.BodyType,
		&row.FuelType, &row.Transmission, &row.EngineVolume,
		&row.City, &row.Region, &row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		return core_domain.Listing{}, fmt.Errorf("create listing: %w", mapRepoError(err))
	}

	return row.toDomain(), nil
}

func mapRepoError(err error) error {
	if err == core_postgres_pool.ErrNoRows {
		return core_errors.ErrNotFound
	}
	return err
}
