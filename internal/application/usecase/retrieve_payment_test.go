package usecase

import (
	"context"
	"testing"
	"time"
	"errors"
	"PaymentGateway/internal/domain"

	"github.com/stretchr/testify/assert"
)

func TestRetrievePaymentUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("Happy Path - Successfully Retrieves and Maps Payment", func(t *testing.T) {
		// 1. Setup Mock (Reusing MockPaymentRepo from process_payment_test.go!)
		mockRepo := new(MockPaymentRepo)

		// 2. Create a mathematically valid Domain Entity to act as the DB record
		money, _ := domain.NewMoney(1500, "USD")
		card, _ := domain.NewCard("4242424242424242", 12, time.Now().Year()+1, "123")
		
		payment := domain.NewPayment("pay_123", money, card, "idem_abc")
		_ = payment.Authorize() // Transition to a realistic state

		// 3. Define Expectations
		mockRepo.On("FindByID", ctx, "pay_123").Return(&payment, nil)

		// 4. Execute
		cmd := RetrievePaymentCommand{ID: "pay_123"}
		uc := NewRetrievePaymentUseCase(mockRepo)
		result, err := uc.Execute(ctx, cmd)

		// 5. Assert the mapping was mathematically perfect
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "pay_123", result.ID)
		assert.Equal(t, string(domain.StatusAuthorized), result.Status)
		assert.Equal(t, "4242", result.CardNumberLastFour)
		assert.Equal(t, 12, result.ExpiryMonth)
		assert.Equal(t, card.ExpiryYear(), result.ExpiryYear)
		assert.Equal(t, "USD", result.Currency)
		assert.Equal(t, int64(1500), result.Amount)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Sad Path - Propagates Not Found Error", func(t *testing.T) {
		mockRepo := new(MockPaymentRepo)

		// Setup mock to simulate a database miss
		mockRepo.On("FindByID", ctx, "pay_not_found").Return(nil, domain.ErrPaymentNotFound)

		cmd := RetrievePaymentCommand{ID: "pay_not_found"}
		uc := NewRetrievePaymentUseCase(mockRepo)

		result, err := uc.Execute(ctx, cmd)

		// Assert it fails fast and returns the domain error exactly
		assert.ErrorIs(t, err, domain.ErrPaymentNotFound)
		assert.Nil(t, result)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Sad Path - Propagates Infrastructure Error", func(t *testing.T) {
		mockRepo := new(MockPaymentRepo)
		expectedErr := errors.New("database connection timeout")

		// Setup mock to simulate a database failure (e.g., 500 Internal Server Error)
		mockRepo.On("FindByID", ctx, "pay_db_down").Return(nil, expectedErr)

		cmd := RetrievePaymentCommand{ID: "pay_db_down"}
		uc := NewRetrievePaymentUseCase(mockRepo)

		result, err := uc.Execute(ctx, cmd)

		// Assert the infrastructure error bubbles up unchanged
		assert.ErrorIs(t, err, expectedErr)
		assert.Nil(t, result)

		mockRepo.AssertExpectations(t)
	})
}