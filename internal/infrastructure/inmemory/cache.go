package inmemory

import (
	"context"
	"errors"
	"time"

	"PaymentGateway/internal/domain"

	"github.com/redis/go-redis/v9"
)

// Build-time check to ensure RedisIdempotencyStore implements the domain interface
var _ domain.IdempotencyStore = (*RedisIdempotencyStore)(nil)

// RedisIdempotencyStore is a Redis-backed implementation of the IdempotencyStore.
type RedisIdempotencyStore struct {
	client *redis.Client
}

// NewRedisIdempotencyStore creates a new Redis cache adapter.
func NewRedisIdempotencyStore(client *redis.Client) *RedisIdempotencyStore {
	return &RedisIdempotencyStore{
		client: client,
	}
}

// SetNX attempts to acquire a distributed lock for the given key.
// Returns true if the lock was acquired, false if it already exists.
func (r *RedisIdempotencyStore) SetNX(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error) {
	err := r.client.SetArgs(ctx, key, value, redis.SetArgs{
		Mode: "NX", // Only set if the key does NOT exist
		TTL:  ttl,
	}).Err()

	if err != nil {
		// If the lock fails because the key already exists, go-redis returns redis.Nil.
		// We safely catch this and return false (lock not acquired) instead of a system error.
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		// A true infrastructure failure (e.g., Redis is down)
		return false, err
	}

	// The command succeeded, meaning we exclusively acquired the lock
	return true, nil
}

// Get retrieves the raw bytes stored at the given key.
func (r *RedisIdempotencyStore) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		// Cache miss
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}
	return val, nil
}

// Set forcefully overwrites the key with the final cached value (the JSON response).
func (r *RedisIdempotencyStore) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}