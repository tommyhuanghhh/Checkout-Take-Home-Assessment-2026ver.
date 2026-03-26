package domain

import (
	"context"
	"time"
)

// IdempotencyStore defines the contract for preventing duplicate payment requests.
type IdempotencyStore interface {
	// SetNX attempts to lock the idempotency key.
	// Returns true if the lock was acquired, false if it already exists.
	SetNX(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error)

	// Get retrieves the cached response for a given idempotency key.
	// Returns ErrIdempotencyKeyNotFound if the key does not exist.
	Get(ctx context.Context, key string) ([]byte, error)

	// Set unconditionally overwrites the key. Used to save the final Payment result 
	// after processing finishes, replacing the initial lock.
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
}