package inmemory

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisIdempotencyStore(t *testing.T) {
	// 1. Spin up an in-memory Miniredis server
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close() // Shut it down when the test suite finishes

	// 2. Setup the go-redis client to connect to the Miniredis server
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(), // Connects dynamically to the randomized local port
	})
	defer client.Close()

	// 3. Instantiate our adapter
	store := NewRedisIdempotencyStore(client)
	ctx := context.Background()

	t.Run("Happy Path - Acquire Lock, Get, and Set", func(t *testing.T) {
		key := "idem_happy"
		val := []byte("IN_PROGRESS")
		ttl := 1 * time.Minute

		// 1. Acquire lock
		locked, err := store.SetNX(ctx, key, val, ttl)
		assert.NoError(t, err)
		assert.True(t, locked)

		// 2. Get the value
		retrieved, err := store.Get(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, val, retrieved)

		// 3. Overwrite the value with the final JSON response
		finalVal := []byte(`{"status": "Authorized"}`)
		err = store.Set(ctx, key, finalVal, ttl)
		assert.NoError(t, err)

		// 4. Verify overwrite was successful
		retrievedFinal, err := store.Get(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, finalVal, retrievedFinal)
	})

	t.Run("Sad Path - SetNX Conflict (Lock already held)", func(t *testing.T) {
		key := "idem_conflict"
		val := []byte("IN_PROGRESS")
		ttl := 1 * time.Minute

		// First lock succeeds
		locked1, err := store.SetNX(ctx, key, val, ttl)
		assert.NoError(t, err)
		assert.True(t, locked1)

		// Second lock on the exact same key fails
		locked2, err := store.SetNX(ctx, key, val, ttl)
		assert.NoError(t, err)
		assert.False(t, locked2) // Conflict!
	})

	t.Run("Sad Path - Get Cache Miss handles redis.Nil safely", func(t *testing.T) {
		// Attempt to fetch a key that does not exist in Miniredis
		retrieved, err := store.Get(ctx, "idem_missing")

		// Prove that we caught `redis.Nil` and returned `nil, nil` safely
		assert.NoError(t, err)
		assert.Nil(t, retrieved)
	})
}