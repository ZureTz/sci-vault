package cache

import (
	"context"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheConnector wraps a Redis client
type CacheConnector struct {
	client *redis.Client
}

// NewCacheConnector creates a new CacheConnector instance
func NewCacheConnector(addr, password string, db int) *CacheConnector {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	slog.Info("connected to redis cache", "addr", addr, "db", db)

	return &CacheConnector{
		client: client,
	}
}

// Health checks the Redis connection status
func (cc *CacheConnector) Health(ctx context.Context) error {
	_, err := cc.client.Ping(ctx).Result()
	return err
}

// Set sets a key-value pair
func (cc *CacheConnector) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return cc.client.Set(ctx, key, value, expiration).Err()
}

// Get gets the value of a key
func (cc *CacheConnector) Get(ctx context.Context, key string) (string, error) {
	return cc.client.Get(ctx, key).Result()
}

// Del deletes specified keys
func (cc *CacheConnector) Del(ctx context.Context, keys ...string) (int64, error) {
	return cc.client.Del(ctx, keys...).Result()
}

// Exists checks if a key exists
func (cc *CacheConnector) Exists(ctx context.Context, keys ...string) (int64, error) {
	return cc.client.Exists(ctx, keys...).Result()
}

// Expire sets the expiration time for a key
func (cc *CacheConnector) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	return cc.client.Expire(ctx, key, expiration).Result()
}

// TTL gets the remaining time to live for a key
func (cc *CacheConnector) TTL(ctx context.Context, key string) (time.Duration, error) {
	return cc.client.TTL(ctx, key).Result()
}

// Incr increments the value of a key by 1
func (cc *CacheConnector) Incr(ctx context.Context, key string) (int64, error) {
	return cc.client.Incr(ctx, key).Result()
}

// IncrBy increments the value of a key by a specified amount
func (cc *CacheConnector) IncrBy(ctx context.Context, key string, incr int64) (int64, error) {
	return cc.client.IncrBy(ctx, key, incr).Result()
}

// Close closes the Redis connection
func (cc *CacheConnector) Close() error {
	return cc.client.Close()
}

// Client returns the underlying Redis client reference (advanced usage)
func (cc *CacheConnector) Client() *redis.Client {
	return cc.client
}
