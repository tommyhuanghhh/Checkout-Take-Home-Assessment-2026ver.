package domain

import (
	"context"
)

// PaymentRepository defines the storage contract for Payment entities.
type PaymentRepository interface {
	// FindByID retrieves a payment. Returns ErrPaymentNotFound if it doesn't exist.
	FindByID(ctx context.Context, id string) (*Payment, error)

	// Save persists a new payment or updates an existing one.
	Save(ctx context.Context, payment *Payment) error
}
