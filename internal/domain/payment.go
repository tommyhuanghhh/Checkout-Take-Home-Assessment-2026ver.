package domain

import (
	"time"
)

// 1. Strongly Typed Enums for Status
type PaymentStatus string

const (
	StatusPending    PaymentStatus = "Pending"
	StatusAuthorized PaymentStatus = "Authorized"
	StatusDeclined   PaymentStatus = "Declined"
)

// 2. The Encapsulated Entity
type Payment struct {
	id             string
	money          Money         // Uses your newly created Value Object!
	card           Card          // Uses your newly created Value Object!
	status         PaymentStatus
	idempotencyKey string        // Crucial for tracking who created this
	createdAt      time.Time     // Always track creation in financial systems
}

// 3. The Constructor (Notice it starts in a "Pending" state)
func NewPayment(id string, money Money, card Card, idempotencyKey string) Payment {
	return Payment{
		id:             id,
		money:          money,
		card:           card,
		status:         StatusPending,
		idempotencyKey: idempotencyKey,
		createdAt:      time.Now().UTC(),
	}
}


// 4. State Transition Methods (The ONLY way to change a payment's status)
// Authorize safely transitions the payment to Authorized.
// It returns an error if the payment is no longer Pending.
func (p *Payment) Authorize() error {
	if p.status != StatusPending {
		return ErrPaymentAlreadyProcessed
	}
	p.status = StatusAuthorized
	return nil
}

// Decline safely transitions the payment to Declined.
// It returns an error if the payment is no longer Pending.
func (p *Payment) Decline() error {
	if p.status != StatusPending {
		return ErrPaymentAlreadyProcessed
	}
	p.status = StatusDeclined
	return nil
}

// 5. Safe Getters
func (p Payment) ID() string             { return p.id }
func (p Payment) Money() Money           { return p.money }
func (p Payment) Card() Card             { return p.card }
func (p Payment) Status() PaymentStatus  { return p.status }
func (p Payment) IdempotencyKey() string { return p.idempotencyKey }
func (p Payment) CreatedAt() time.Time   { return p.createdAt }