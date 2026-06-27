package listings_repository

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	core_domain "listing-service/internal/core/domain"
	core_logger "listing-service/internal/core/logger"
	core_redis_cache "listing-service/internal/core/repository/redis"
)

func testLogger() *core_logger.Logger {
	return &core_logger.Logger{Logger: zap.NewNop()}
}

// fakeStore — ручная fake-реализация listingsStore. Каждый тест сам
// решает, что вернуть, через поля-функции; счётчики вызовов позволяют
// проверять, что кэш реально избавляет от похода в "БД".
type fakeStore struct {
	getListingByIDFunc func(ctx context.Context, id uuid.UUID) (core_domain.Listing, error)
	getListingsFunc    func(ctx context.Context, filter core_domain.ListingFilter) ([]core_domain.Listing, error)
	updateListingFunc  func(ctx context.Context, id uuid.UUID, update core_domain.ListingUpdate) (core_domain.Listing, error)
	deleteListingFunc  func(ctx context.Context, id uuid.UUID) error

	getListingByIDCalls int
	getListingsCalls    int
}

func (f *fakeStore) CreateListing(ctx context.Context, listing core_domain.Listing) (core_domain.Listing, error) {
	return listing, nil
}

func (f *fakeStore) GetListingByID(ctx context.Context, id uuid.UUID) (core_domain.Listing, error) {
	f.getListingByIDCalls++
	return f.getListingByIDFunc(ctx, id)
}

func (f *fakeStore) GetListings(ctx context.Context, filter core_domain.ListingFilter) ([]core_domain.Listing, error) {
	f.getListingsCalls++
	return f.getListingsFunc(ctx, filter)
}

func (f *fakeStore) UpdateListing(ctx context.Context, id uuid.UUID, update core_domain.ListingUpdate) (core_domain.Listing, error) {
	return f.updateListingFunc(ctx, id, update)
}

func (f *fakeStore) DeleteListing(ctx context.Context, id uuid.UUID) error {
	return f.deleteListingFunc(ctx, id)
}

// fakeCache — потокобезопасная in-memory реализация core_redis_cache.Cache.
type fakeCache struct {
	mu   sync.Mutex
	data map[string]string
}

func newFakeCache() *fakeCache {
	return &fakeCache{data: make(map[string]string)}
}

func (c *fakeCache) Get(ctx context.Context, key string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	val, ok := c.data[key]
	if !ok {
		return "", core_redis_cache.ErrCacheMiss
	}
	return val, nil
}

func (c *fakeCache) Set(ctx context.Context, key string, value string, _ time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = value
	return nil
}

func (c *fakeCache) Del(ctx context.Context, keys ...string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, key := range keys {
		delete(c.data, key)
	}
	return nil
}

func (c *fakeCache) Close() error { return nil }

func TestCachedRepository_GetListingByID_CachesAfterMiss(t *testing.T) {
	id := uuid.New()
	want := core_domain.Listing{ID: id, Title: "Camry"}

	store := &fakeStore{
		getListingByIDFunc: func(ctx context.Context, gotID uuid.UUID) (core_domain.Listing, error) {
			return want, nil
		},
	}
	cache := newFakeCache()
	repo := NewCachedRepository(store, cache, testLogger())

	for i := 0; i < 3; i++ {
		got, err := repo.GetListingByID(context.Background(), id)
		if err != nil {
			t.Fatalf("GetListingByID() error = %v", err)
		}
		if got.Title != want.Title {
			t.Fatalf("GetListingByID() = %+v, want %+v", got, want)
		}
	}

	if store.getListingByIDCalls != 1 {
		t.Fatalf("expected underlying store to be hit once, got %d calls", store.getListingByIDCalls)
	}
}

func TestCachedRepository_UpdateListing_InvalidatesCache(t *testing.T) {
	id := uuid.New()
	store := &fakeStore{
		getListingByIDFunc: func(ctx context.Context, gotID uuid.UUID) (core_domain.Listing, error) {
			return core_domain.Listing{ID: id, Title: "stale"}, nil
		},
		updateListingFunc: func(ctx context.Context, gotID uuid.UUID, update core_domain.ListingUpdate) (core_domain.Listing, error) {
			return core_domain.Listing{ID: id, Title: "fresh"}, nil
		},
	}
	cache := newFakeCache()
	repo := NewCachedRepository(store, cache, testLogger())

	if _, err := repo.GetListingByID(context.Background(), id); err != nil {
		t.Fatalf("GetListingByID() error = %v", err)
	}

	if _, err := cache.Get(context.Background(), listingCacheKey(id)); err != nil {
		t.Fatalf("expected listing to be cached before update, got error: %v", err)
	}

	if _, err := repo.UpdateListing(context.Background(), id, core_domain.ListingUpdate{}); err != nil {
		t.Fatalf("UpdateListing() error = %v", err)
	}

	if _, err := cache.Get(context.Background(), listingCacheKey(id)); err == nil {
		t.Fatalf("expected cache to be invalidated after update")
	}
}

func TestCachedRepository_DeleteListing_InvalidatesCache(t *testing.T) {
	id := uuid.New()
	store := &fakeStore{
		getListingByIDFunc: func(ctx context.Context, gotID uuid.UUID) (core_domain.Listing, error) {
			return core_domain.Listing{ID: id, Title: "to-delete"}, nil
		},
		deleteListingFunc: func(ctx context.Context, gotID uuid.UUID) error {
			return nil
		},
	}
	cache := newFakeCache()
	repo := NewCachedRepository(store, cache, testLogger())

	if _, err := repo.GetListingByID(context.Background(), id); err != nil {
		t.Fatalf("GetListingByID() error = %v", err)
	}

	if err := repo.DeleteListing(context.Background(), id); err != nil {
		t.Fatalf("DeleteListing() error = %v", err)
	}

	if _, err := cache.Get(context.Background(), listingCacheKey(id)); err == nil {
		t.Fatalf("expected cache to be invalidated after delete")
	}
}

func TestCachedRepository_GetListings_CachesByFilter(t *testing.T) {
	city := "Almaty"
	filter := core_domain.ListingFilter{City: &city, Page: 1, Limit: 20}
	want := []core_domain.Listing{{ID: uuid.New(), City: "Almaty"}}

	store := &fakeStore{
		getListingsFunc: func(ctx context.Context, gotFilter core_domain.ListingFilter) ([]core_domain.Listing, error) {
			return want, nil
		},
	}
	cache := newFakeCache()
	repo := NewCachedRepository(store, cache, testLogger())

	for i := 0; i < 2; i++ {
		got, err := repo.GetListings(context.Background(), filter)
		if err != nil {
			t.Fatalf("GetListings() error = %v", err)
		}
		if len(got) != 1 || got[0].City != "Almaty" {
			t.Fatalf("GetListings() = %+v, want %+v", got, want)
		}
	}

	if store.getListingsCalls != 1 {
		t.Fatalf("expected underlying store to be hit once, got %d calls", store.getListingsCalls)
	}
}
