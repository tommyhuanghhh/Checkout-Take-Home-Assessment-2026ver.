package inmemory

import (
	"context"
	"sync"

	"PaymentGateway/internal/domain"
)

// Build-time check to ensure InMemoryPaymentRepository implements the domain interface.
var _ domain.PaymentRepository = (*InMemoryPaymentRepository)(nil)

// InMemoryPaymentRepository is a thread-safe, in-memory implementation of the PaymentRepository.
type InMemoryPaymentRepository struct {
	mu   sync.RWMutex
	// We use a map instead of a slice for O(1) instant lookups.
	// We store the actual struct (domain.Payment) rather than a pointer to guarantee
	// that subsequent mutations to the pointer in the Use Case don't accidentally mutate our DB state.
	data map[string]domain.Payment
}

// NewInMemoryPaymentRepository creates a new instance of the repository.
func NewInMemoryPaymentRepository() *InMemoryPaymentRepository {
	return &InMemoryPaymentRepository{
		data: make(map[string]domain.Payment),
	}
}

// Save securely persists the payment to the in-memory map.
func (r *InMemoryPaymentRepository) Save(ctx context.Context, payment *domain.Payment) error {
	// Lock blocks any other goroutine from reading or writing until this operation finishes.
	// This provides strict Isolation (I) and Atomicity (A).
	r.mu.Lock()
	defer r.mu.Unlock()

	// Dereference the pointer to store a pristine copy of the value in the map.
	r.data[payment.ID()] = *payment

	return nil
}

// FindByID retrieves a payment by its UUID.
func (r *InMemoryPaymentRepository) FindByID(ctx context.Context, id string) (*domain.Payment, error) {
	// RLock allows multiple goroutines to read simultaneously, massively boosting throughput.
	// However, it will block if a goroutine is currently holding a Write Lock.
	r.mu.RLock()
	defer r.mu.RUnlock()

	payment, exists := r.data[id]
	if !exists {
		return nil, domain.ErrPaymentNotFound
	}

	// We return a pointer to a copy of the retrieved entity, ensuring the caller 
	// cannot directly mutate the data stored inside the repository map.
	return &payment, nil
}