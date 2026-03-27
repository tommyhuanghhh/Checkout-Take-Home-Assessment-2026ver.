package usecase

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"PaymentGateway/internal/domain"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockIdempotencyStore struct{ mock.Mock }

func (m *MockIdempotencyStore) SetNX(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error) {
	args := m.Called(ctx, key, value, ttl)
	return args.Bool(0), args.Error(1)
}
func (m *MockIdempotencyStore) Get(ctx context.Context, key string) ([]byte, error) {
	args := m.Called(ctx, key)
	if args.Get(0) != nil {
		return args.Get(0).([]byte), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockIdempotencyStore) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

type MockPaymentRepo struct{ mock.Mock }

func (m *MockPaymentRepo) Save(ctx context.Context, payment *domain.Payment) error {
	args := m.Called(ctx, payment)
	return args.Error(0)
}

func (m *MockPaymentRepo) FindByID(ctx context.Context, id string) (*domain.Payment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.Payment), args.Error(1)
	}
	return nil, args.Error(1)
}

type MockIDGenerator struct{ mock.Mock }

func (m *MockIDGenerator) Generate() string {
	args := m.Called()
	return args.String(0)
}

type MockBankService struct{ mock.Mock }

func (m *MockBankService) Process(ctx context.Context, amount int64, currency string, pan string, expMonth, expYear int, cvv string) (bool, error) {
	args := m.Called(ctx, amount, currency, pan, expMonth, expYear, cvv)
	return args.Bool(0), args.Error(1)
}

func TestProcessPaymentUseCase_Execute(t *testing.T) {
	// Mathematically valid Luhn PANs mapped to the Bank Simulator rules
	authorizedPAN := "4242424242424341" // Ends in Odd (1) -> Authorized
	declinedPAN   := "4242424242424242" // Ends in Even (2) -> Declined
	errorPAN      := "4242424242424440" // Ends in Zero (0) -> 503 Error
	futureYear := time.Now().Year() + 1
	ctx := context.Background()

	t.Run("Happy Path - Successfully Authorizes Payment", func(t *testing.T) {
		// 1. Setup Mocks
		mockStore := new(MockIdempotencyStore)
		mockRepo := new(MockPaymentRepo)
		mockIDGen := new(MockIDGenerator)
		mockBank := new(MockBankService)

		cmd := ProcessPaymentCommand{
			IdempotencyKey: "idem_123",
			PAN:            authorizedPAN,
			ExpiryMonth:    12,
			ExpiryYear:     futureYear,
			CVV:            "123",
			Amount:         1000,
			Currency:       "USD",
		}

		// 2. Define Expectations (Strict Order of Operations)
		// Lock succeeds
		mockStore.On("SetNX", ctx, "idem_123", []byte(IdempotencyInProgressMarker), IdempotencyTTL).Return(true, nil)
		
		// ID generated
		mockIDGen.On("Generate").Return("pay_999")
		
		// Bank authorizes
		mockBank.On("Process", ctx, int64(1000), "USD", authorizedPAN, 12, futureYear, "123").Return(true, nil)
		
		// Repo saves (using mock.Anything to avoid matching the exact struct pointer memory address)
		mockRepo.On("Save", ctx, mock.AnythingOfType("*domain.Payment")).Return(nil)
		
		// Cache overwritten with final JSON
		mockStore.On("Set", ctx, "idem_123", mock.AnythingOfType("[]uint8"), IdempotencyTTL).Return(nil)

		// 3. Execute
		uc := NewProcessPaymentUseCase(mockRepo, mockStore, mockIDGen, mockBank)
		result, err := uc.Execute(ctx, cmd)

		// 4. Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "pay_999", result.ID)
		assert.Equal(t, string(domain.StatusAuthorized), result.Status)
		assert.Equal(t, "4341", result.CardNumberLastFour)

		// Ensure all mocked methods were called exactly as expected
		mockStore.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
		mockBank.AssertExpectations(t)
	})

	t.Run("Idempotency Hit - Returns Cached Result Directly", func(t *testing.T) {
		mockStore := new(MockIdempotencyStore)
		// We don't even need to mock the Repo, Bank, or IDGen because they should NEVER be called!
		
		cmd := ProcessPaymentCommand{IdempotencyKey: "idem_recovered"}

		// Simulate a previously completed payment JSON sitting in the cache
		cachedResult := ProcessPaymentResult{
			ID:                 "pay_recovered_1",
			Status:             string(domain.StatusAuthorized),
			CardNumberLastFour: "4242",
		}
		cachedBytes, _ := json.Marshal(cachedResult)

		// SetNX returns false (lock exists)
		mockStore.On("SetNX", ctx, "idem_recovered", []byte(IdempotencyInProgressMarker), IdempotencyTTL).Return(false, nil)
		// Get returns the JSON
		mockStore.On("Get", ctx, "idem_recovered").Return(cachedBytes, nil)

		uc := NewProcessPaymentUseCase(nil, mockStore, nil, nil) // Passing nils to prove they aren't touched!
		result, err := uc.Execute(ctx, cmd)

		assert.NoError(t, err)
		assert.Equal(t, "pay_recovered_1", result.ID)
		assert.Equal(t, "4242", result.CardNumberLastFour)
		
		mockStore.AssertExpectations(t)
	})

	t.Run("Domain Rejection - Fails Fast on Invalid Card", func(t *testing.T) {
		mockStore := new(MockIdempotencyStore)
		
		cmd := ProcessPaymentCommand{
			IdempotencyKey: "idem_bad_card",
			PAN:            "4111", // Too short, invalid Luhn
			ExpiryMonth:    12,
			ExpiryYear:     futureYear,
		}

		// Lock succeeds, but then validation fails immediately
		mockStore.On("SetNX", ctx, "idem_bad_card", []byte(IdempotencyInProgressMarker), IdempotencyTTL).Return(true, nil)

		// Notice we pass `nil` for Repo, Bank, and IDGen to prove they are never invoked
		uc := NewProcessPaymentUseCase(nil, mockStore, nil, nil)
		result, err := uc.Execute(ctx, cmd)

		// Assert it returns the specific domain error and nil result
		assert.ErrorIs(t, err, domain.ErrInvalidCardNumber)
		assert.Nil(t, result)
		
		mockStore.AssertExpectations(t)
	})

	t.Run("Bank Rejection - Saves and Returns Declined Status", func(t *testing.T) {
		mockStore := new(MockIdempotencyStore)
		mockRepo := new(MockPaymentRepo)
		mockIDGen := new(MockIDGenerator)
		mockBank := new(MockBankService)

		cmd := ProcessPaymentCommand{
			IdempotencyKey: "idem_declined",
			PAN:            declinedPAN,
			ExpiryMonth:    12,
			ExpiryYear:     futureYear,
			CVV:            "123",
			Amount:         1000,
			Currency:       "USD",
		}

		// Lock succeeds, ID generates
		mockStore.On("SetNX", ctx, "idem_declined", []byte(IdempotencyInProgressMarker), IdempotencyTTL).Return(true, nil)
		mockIDGen.On("Generate").Return("pay_declined_1")
		
		// BANK RETURNS FALSE (Declined)
		mockBank.On("Process", ctx, int64(1000), "USD", declinedPAN, 12, futureYear, "123").Return(false, nil)
		
		// We still expect to Save to DB and Set the cache!
		mockRepo.On("Save", ctx, mock.AnythingOfType("*domain.Payment")).Return(nil)
		mockStore.On("Set", ctx, "idem_declined", mock.AnythingOfType("[]uint8"), IdempotencyTTL).Return(nil)

		uc := NewProcessPaymentUseCase(mockRepo, mockStore, mockIDGen, mockBank)
		result, err := uc.Execute(ctx, cmd)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, string(domain.StatusDeclined), result.Status) // Proves the status changed!

		mockStore.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
		mockBank.AssertExpectations(t)
	})

	t.Run("Bank Failure - Aborts and Returns Error", func(t *testing.T) {
		mockStore := new(MockIdempotencyStore)
		mockRepo := new(MockPaymentRepo)
		mockIDGen := new(MockIDGenerator)
		mockBank := new(MockBankService)

		cmd := ProcessPaymentCommand{
			IdempotencyKey: "idem_bank_down",
			PAN:            errorPAN,
			ExpiryMonth:    12,
			ExpiryYear:     futureYear,
			CVV:            "123",
			Amount:         1000,
			Currency:       "USD",
		}

		mockStore.On("SetNX", ctx, "idem_bank_down", []byte(IdempotencyInProgressMarker), IdempotencyTTL).Return(true, nil)
		mockIDGen.On("Generate").Return("pay_err_1")
		
		// BANK RETURNS ERROR (e.g., 503 Service Unavailable)
		expectedErr := errors.New("bank unavailable: 503")
		mockBank.On("Process", ctx, int64(1000), "USD", errorPAN, 12, futureYear, "123").Return(false, expectedErr)
		
		// CRITICAL: We DO NOT mock Repo.Save or Store.Set. 
		// If the Use Case tries to call them, the test will instantly fail!

		uc := NewProcessPaymentUseCase(mockRepo, mockStore, mockIDGen, mockBank)
		result, err := uc.Execute(ctx, cmd)

		assert.ErrorIs(t, err, expectedErr) // Proves the error bubbled up
		assert.Nil(t, result)               // Proves no result was returned

		mockStore.AssertExpectations(t)
		mockRepo.AssertExpectations(t) // Guarantees the DB was never touched
		mockBank.AssertExpectations(t)
	})
}