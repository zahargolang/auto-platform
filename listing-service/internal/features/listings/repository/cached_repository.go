package listings_repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	core_domain "listing-service/internal/core/domain"
	core_logger "listing-service/internal/core/logger"
	core_redis_cache "listing-service/internal/core/repository/redis"
)

const (
	listingCacheTTL  = 10 * time.Minute
	listingsCacheTTL = 30 * time.Second
)

// listingsStore — то, что нужно CachedRepository от нижележащего репозитория.
// Определён здесь, на стороне потребителя, чтобы декоратор можно было
// тестировать с fake-реализацией без поднятия Postgres.
type listingsStore interface {
	CreateListing(ctx context.Context, listing core_domain.Listing) (core_domain.Listing, error)
	GetListingByID(ctx context.Context, id uuid.UUID) (core_domain.Listing, error)
	GetListings(ctx context.Context, filter core_domain.ListingFilter) ([]core_domain.Listing, error)
	UpdateListing(ctx context.Context, id uuid.UUID, update core_domain.ListingUpdate) (core_domain.Listing, error)
	DeleteListing(ctx context.Context, id uuid.UUID) error
}

// CachedRepository — read-through кэш поверх listingsStore: одиночные
// объявления кэшируются по ID и инвалидируются при изменении/удалении,
// а списки кэшируются по хэшу фильтра на короткий TTL без активной
// инвалидации (точечно инвалидировать все комбинации фильтров непрактично,
// а короткий TTL делает устаревание приемлемым для ленты объявлений).
type CachedRepository struct {
	repo  listingsStore
	cache core_redis_cache.Cache
	log   *core_logger.Logger
}

func NewCachedRepository(repo listingsStore, cache core_redis_cache.Cache, log *core_logger.Logger) *CachedRepository {
	return &CachedRepository{repo: repo, cache: cache, log: log}
}

func listingCacheKey(id uuid.UUID) string {
	return fmt.Sprintf("listing:%s", id)
}

func listingsCacheKey(filter core_domain.ListingFilter) (string, error) {
	body, err := json.Marshal(filter)
	if err != nil {
		return "", fmt.Errorf("marshal filter: %w", err)
	}

	h := fnv.New64a()
	h.Write(body)

	return fmt.Sprintf("listings:%x", h.Sum64()), nil
}

func (r *CachedRepository) CreateListing(ctx context.Context, listing core_domain.Listing) (core_domain.Listing, error) {
	return r.repo.CreateListing(ctx, listing)
}

func (r *CachedRepository) GetListingByID(ctx context.Context, id uuid.UUID) (core_domain.Listing, error) {
	key := listingCacheKey(id)

	if cached, err := r.cache.Get(ctx, key); err == nil {
		var listing core_domain.Listing
		if jsonErr := json.Unmarshal([]byte(cached), &listing); jsonErr == nil {
			return listing, nil
		}
	} else if !errors.Is(err, core_redis_cache.ErrCacheMiss) {
		r.log.Warn("cache get failed", zap.String("key", key), zap.Error(err))
	}

	listing, err := r.repo.GetListingByID(ctx, id)
	if err != nil {
		return core_domain.Listing{}, err
	}

	r.setCache(ctx, key, listing, listingCacheTTL)

	return listing, nil
}

func (r *CachedRepository) GetListings(ctx context.Context, filter core_domain.ListingFilter) ([]core_domain.Listing, error) {
	key, keyErr := listingsCacheKey(filter)
	if keyErr == nil {
		if cached, err := r.cache.Get(ctx, key); err == nil {
			var listings []core_domain.Listing
			if jsonErr := json.Unmarshal([]byte(cached), &listings); jsonErr == nil {
				return listings, nil
			}
		} else if !errors.Is(err, core_redis_cache.ErrCacheMiss) {
			r.log.Warn("cache get failed", zap.String("key", key), zap.Error(err))
		}
	}

	//если в кэше не найдены нужные данные, то в таком случае идем в кэш
	listings, err := r.repo.GetListings(ctx, filter)
	if err != nil {
		return nil, err
	}

	//записываем полученные данные из БД в кэш
	if keyErr == nil {
		r.setCache(ctx, key, listings, listingsCacheTTL)
	}

	return listings, nil
}

func (r *CachedRepository) UpdateListing(ctx context.Context, id uuid.UUID, update core_domain.ListingUpdate) (core_domain.Listing, error) {
	updated, err := r.repo.UpdateListing(ctx, id, update)
	if err != nil {
		return core_domain.Listing{}, err
	}

	//удаляем данные из кэша, так как они уже не актуальные
	r.invalidate(ctx, id)

	return updated, nil
}

func (r *CachedRepository) DeleteListing(ctx context.Context, id uuid.UUID) error {
	if err := r.repo.DeleteListing(ctx, id); err != nil {
		return err
	}

	r.invalidate(ctx, id)

	return nil
}

func (r *CachedRepository) setCache(ctx context.Context, key string, value any, ttl time.Duration) {
	body, err := json.Marshal(value)
	if err != nil {
		r.log.Warn("cache marshal failed", zap.String("key", key), zap.Error(err))
		return
	}

	if err := r.cache.Set(ctx, key, string(body), ttl); err != nil {
		r.log.Warn("cache set failed", zap.String("key", key), zap.Error(err))
	}
}

func (r *CachedRepository) invalidate(ctx context.Context, id uuid.UUID) {
	key := listingCacheKey(id)
	if err := r.cache.Del(ctx, key); err != nil {
		r.log.Warn("cache invalidate failed", zap.String("key", key), zap.Error(err))
	}
}
