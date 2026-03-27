package inmemory

import (
	"context"
	"sync"
	"testing"
	"time"

	"PaymentGateway/internal/domain"

	"github.com/stretchr/testify/assert"
)

func TestInMemoryPaymentRepository(t *testing.T) {
	ctx := context.Background()

	// Helper to create a valid dummy payment
	createDummyPayment := func(id string) domain.Payment {
		money, _ := domain.NewMoney(1000, "USD")
		card, _ := domain.NewCard("4242424242424242", 12, time.Now().Year()+1, "123")
		return domain.NewPayment(id, money, card, "idem_123")
	}

	t.Run("Happy Path - Saves and Retrieves Payment", func(t *testing.T) {
		repo := NewInMemoryPaymentRepository()
		payment := createDummyPayment("pay_1")

		err := repo.Save(ctx, &payment)
		assert.NoError(t, err)

		retrieved, err := repo.FindByID(ctx, "pay_1")
		assert.NoError(t, err)
		assert.NotNil(t, retrieved)
		assert.Equal(t, "pay_1", retrieved.ID())
		assert.Equal(t, domain.StatusPending, retrieved.Status())
	})

	t.Run("Sad Path - Returns NotFound Error", func(t *testing.T) {
		repo := NewInMemoryPaymentRepository()

		retrieved, err := repo.FindByID(ctx, "pay_unknown")
		assert.ErrorIs(t, err, domain.ErrPaymentNotFound)
		assert.Nil(t, retrieved)
	})

	t.Run("Advanced - Proves Pointer Memory Isolation", func(t *testing.T) {
		repo := NewInMemoryPaymentRepository()
		payment := createDummyPayment("pay_isolated")
		_ = repo.Save(ctx, &payment)

		// Fetch the payment
		retrieved1, _ := repo.FindByID(ctx, "pay_isolated")
		
		// Mutate the retrieved payment (simulate Use Case doing work)
		_ = retrieved1.Authorize()
		assert.Equal(t, domain.StatusAuthorized, retrieved1.Status())

		// Fetch it AGAIN from the DB
		retrieved2, _ := repo.FindByID(ctx, "pay_isolated")

		// The DB should STILL be Pending! 
		// This proves the Use Case cannot bypass the .Save() method!
		assert.Equal(t, domain.StatusPending, retrieved2.Status())
	})

	t.Run("Advanced - Thread Safety under Concurrent Load", func(t *testing.T) {
		repo := NewInMemoryPaymentRepository()
		var wg sync.WaitGroup

		// Spawn 500 Goroutines trying to SAVE simultaneously
		for i := 0; i < 500; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				payment := createDummyPayment("pay_concurrent")
				_ = repo.Save(ctx, &payment) // All hammering the exact same map concurrently
			}(i)
		}

		// Spawn 500 Goroutines trying to READ simultaneously
		for i := 0; i < 500; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, _ = repo.FindByID(ctx, "pay_concurrent")
			}()
		}

		wg.Wait() // Wait for all 1,000 goroutines to finish
		
		// If the test reaches here without panicking or deadlocking, we pass!
		retrieved, err := repo.FindByID(ctx, "pay_concurrent")
		assert.NoError(t, err)
		assert.NotNil(t, retrieved)
	})
}