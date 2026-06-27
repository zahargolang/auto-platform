package listings_repository

import (
	"context"
	"fmt"
	"strings"

	core_domain "listing-service/internal/core/domain"
)

func (r *Repository) GetListings(ctx context.Context, filter core_domain.ListingFilter) ([]core_domain.Listing, error) {
	opCtx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	args := []any{}
	argIdx := 1
	conditions := []string{}

	addArg := func(val any) int {
		args = append(args, val)
		idx := argIdx
		argIdx++
		return idx
	}

	if filter.Make != nil {
		conditions = append(conditions, fmt.Sprintf("make = $%d", addArg(*filter.Make)))
	}
	if filter.Model != nil {
		conditions = append(conditions, fmt.Sprintf("model = $%d", addArg(*filter.Model)))
	}
	if filter.YearFrom != nil {
		conditions = append(conditions, fmt.Sprintf("year >= $%d", addArg(*filter.YearFrom)))
	}
	if filter.YearTo != nil {
		conditions = append(conditions, fmt.Sprintf("year <= $%d", addArg(*filter.YearTo)))
	}
	if filter.PriceFrom != nil {
		conditions = append(conditions, fmt.Sprintf("price >= $%d", addArg(*filter.PriceFrom)))
	}
	if filter.PriceTo != nil {
		conditions = append(conditions, fmt.Sprintf("price <= $%d", addArg(*filter.PriceTo)))
	}
	if filter.MileageFrom != nil {
		conditions = append(conditions, fmt.Sprintf("mileage >= $%d", addArg(*filter.MileageFrom)))
	}
	if filter.MileageTo != nil {
		conditions = append(conditions, fmt.Sprintf("mileage <= $%d", addArg(*filter.MileageTo)))
	}
	if filter.FuelType != nil {
		conditions = append(conditions, fmt.Sprintf("fuel_type = $%d", addArg(string(*filter.FuelType))))
	}
	if filter.Transmission != nil {
		conditions = append(conditions, fmt.Sprintf("transmission = $%d", addArg(string(*filter.Transmission))))
	}
	if filter.BodyType != nil {
		conditions = append(conditions, fmt.Sprintf("body_type = $%d", addArg(string(*filter.BodyType))))
	}
	if filter.City != nil {
		conditions = append(conditions, fmt.Sprintf("city = $%d", addArg(*filter.City)))
	}
	if filter.Region != nil {
		conditions = append(conditions, fmt.Sprintf("region = $%d", addArg(*filter.Region)))
	}
	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", addArg(string(*filter.Status))))
	} else {
		conditions = append(conditions, "status = 'active'")
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := (filter.Page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`
		SELECT id, user_id, title, description, price, status,
			make, model, year, mileage, color, body_type, fuel_type, transmission, engine_volume,
			city, region, created_at, updated_at
		FROM listingservice.listings
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, addArg(limit), addArg(offset))

	rows, err := r.pool.Query(opCtx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get listings: %w", err)
	}
	defer rows.Close()

	var listings []core_domain.Listing
	for rows.Next() {
		var row listingRow
		if err := rows.Scan(
			&row.ID, &row.UserID, &row.Title, &row.Description, &row.Price, &row.Status,
			&row.Make, &row.Model, &row.Year, &row.Mileage, &row.Color, &row.BodyType,
			&row.FuelType, &row.Transmission, &row.EngineVolume,
			&row.City, &row.Region, &row.CreatedAt, &row.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan listing row: %w", err)
		}
		listings = append(listings, row.toDomain())
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return listings, nil
}
