package core_goredis_cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	core_redis_cache "listing-service/internal/core/repository/redis"
)

type Client struct {
	rdb       *redis.Client
	opTimeout time.Duration
}

func NewClient(ctx context.Context, config Config) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
	})

	pingCtx, cancel := context.WithTimeout(ctx, config.Timeout)
	defer cancel()

	if err := rdb.Ping(pingCtx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return &Client{
		rdb:       rdb,
		opTimeout: config.Timeout,
	}, nil
}

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	opCtx, cancel := context.WithTimeout(ctx, c.opTimeout)
	defer cancel()

	val, err := c.rdb.Get(opCtx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", core_redis_cache.ErrCacheMiss
		}
		return "", fmt.Errorf("redis get: %w", err)
	}

	return val, nil
}

func (c *Client) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	opCtx, cancel := context.WithTimeout(ctx, c.opTimeout)
	defer cancel()

	if err := c.rdb.Set(opCtx, key, value, ttl).Err(); err != nil {
		return fmt.Errorf("redis set: %w", err)
	}

	return nil
}

func (c *Client) Del(ctx context.Context, keys ...string) error {
	opCtx, cancel := context.WithTimeout(ctx, c.opTimeout)
	defer cancel()

	if err := c.rdb.Del(opCtx, keys...).Err(); err != nil {
		return fmt.Errorf("redis del: %w", err)
	}

	return nil
}

func (c *Client) Close() error {
	return c.rdb.Close()
}
