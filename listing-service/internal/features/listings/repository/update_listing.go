package listings_repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	core_domain "listing-service/internal/core/domain"
	core_errors "listing-service/internal/core/errors"
	core_postgres_pool "listing-service/internal/core/repository/postgres/pool"
)

func (r *Repository) UpdateListing(ctx context.Context, id uuid.UUID, update core_domain.ListingUpdate) (core_domain.Listing, error) {
	opCtx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	setClauses := []string{}
	args := []any{}
	argIdx := 1

	addSet := func(col string, val any) {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", col, argIdx))
		args = append(args, val)
		argIdx++
	}

	if update.Title.Set {
		addSet("title", update.Title.Value)
	}
	if update.Description.Set {
		addSet("description", update.Description.Value)
	}
	if update.Price.Set {
		addSet("price", update.Price.Value)
	}
	if update.Status.Set {
		addSet("status", update.Status.Value)
	}
	if update.Mileage.Set {
		addSet("mileage", update.Mileage.Value)
	}
	if update.Color.Set {
		addSet("color", update.Color.Value)
	}
	if update.City.Set {
		addSet("city", update.City.Value)
	}
	if update.Region.Set {
		addSet("region", update.Region.Value)
	}

	if len(setClauses) == 0 {
		return r.GetListingByID(ctx, id)
	}

	setClauses = append(setClauses, "updated_at = NOW()")
	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE listingservice.listings
		SET %s
		WHERE id = $%d
		RETURNING id, user_id, title, description, price, status,
			make, model, year, mileage, color, body_type, fuel_type, transmission, engine_volume,
			city, region, created_at, updated_at
	`, strings.Join(setClauses, ", "), argIdx)

	var row listingRow
	err := r.pool.QueryRow(opCtx, query, args...).Scan(
		&row.ID, &row.UserID, &row.Title, &row.Description, &row.Price, &row.Status,
		&row.Make, &row.Model, &row.Year, &row.Mileage, &row.Color, &row.BodyType,
		&row.FuelType, &row.Transmission, &row.EngineVolume,
		&row.City, &row.Region, &row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return core_domain.Listing{}, core_errors.ErrNotFound
		}
		return core_domain.Listing{}, fmt.Errorf("update listing: %w", err)
	}

	return row.toDomain(), nil
}
