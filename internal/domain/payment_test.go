package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupDummyValueObjects creates guaranteed-valid domain objects to inject into the Payment.
// This keeps our entity tests clean and focused on entity behavior, not validation.
func setupDummyValueObjects(t *testing.T) (Money, Card) {
	money, err := NewMoney(1050, "USD")
	require.NoError(t, err)

	card, err := NewCard("4242424242424242", 12, time.Now().Year()+1, "123")
	require.NoError(t, err)

	return money, card
}

func TestNewPayment(t *testing.T) {
	money, card := setupDummyValueObjects(t)
	id := "pay_test_123"
	idempotencyKey := "idem_abc_987"

	payment := NewPayment(id, money, card, idempotencyKey)

	// Verify all fields are assigned correctly using the safe Getters (Black-Box style)
	assert.Equal(t, id, payment.ID())
	assert.Equal(t, money, payment.Money())
	assert.Equal(t, card, payment.Card())
	assert.Equal(t, idempotencyKey, payment.IdempotencyKey())
	
	// CRITICAL: Ensure the entity is born in the Pending state
	assert.Equal(t, StatusPending, payment.Status())
	
	// Ensure the creation timestamp was actually generated
	assert.False(t, payment.CreatedAt().IsZero())
	// Ensure it was created in UTC
	assert.Equal(t, time.UTC, payment.CreatedAt().Location())
}

func TestPayment_StateMachine(t *testing.T) {
	money, card := setupDummyValueObjects(t)

	// We use table-driven tests to exhaustively prove the invariant rules 
	// of our Authorization/Decline state machine.
	tests := []struct {
		name           string
		initialState   PaymentStatus
		action         func(p *Payment) error
		expectedStatus PaymentStatus
		expectedErr    error
	}{
		// --- Happy Paths (Valid Transitions) ---
		{
			name:           "Pending to Authorized",
			initialState:   StatusPending,
			action:         func(p *Payment) error { return p.Authorize() },
			expectedStatus: StatusAuthorized,
			expectedErr:    nil,
		},
		{
			name:           "Pending to Declined",
			initialState:   StatusPending,
			action:         func(p *Payment) error { return p.Decline() },
			expectedStatus: StatusDeclined,
			expectedErr:    nil,
		},
		
		// --- Sad Paths (Illegal Transitions from Authorized) ---
		{
			name:           "Authorized to Authorized (Illegal)",
			initialState:   StatusAuthorized,
			action:         func(p *Payment) error { return p.Authorize() },
			expectedStatus: StatusAuthorized, // State should not change
			expectedErr:    ErrPaymentAlreadyProcessed,
		},
		{
			name:           "Authorized to Declined (Illegal)",
			initialState:   StatusAuthorized,
			action:         func(p *Payment) error { return p.Decline() },
			expectedStatus: StatusAuthorized, // State should not change
			expectedErr:    ErrPaymentAlreadyProcessed,
		},

		// --- Sad Paths (Illegal Transitions from Declined) ---
		{
			name:           "Declined to Authorized (Illegal)",
			initialState:   StatusDeclined,
			action:         func(p *Payment) error { return p.Authorize() },
			expectedStatus: StatusDeclined, // State should not change
			expectedErr:    ErrPaymentAlreadyProcessed,
		},
		{
			name:           "Declined to Declined (Illegal)",
			initialState:   StatusDeclined,
			action:         func(p *Payment) error { return p.Decline() },
			expectedStatus: StatusDeclined, // State should not change
			expectedErr:    ErrPaymentAlreadyProcessed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 1. Arrange: Create the payment and force it into the test's initial state
			payment := NewPayment("pay_123", money, card, "idem_123")
			payment.status = tt.initialState 

			// 2. Act: Attempt the state transition
			err := tt.action(&payment)

			// 3. Assert: Verify the error and ensure the state is exactly what we expect
			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedStatus, payment.Status())
		})
	}
}