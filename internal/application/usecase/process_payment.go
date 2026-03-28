package usecase

import (
	"context"
	"encoding/json"
	"time"
	"errors"
	"PaymentGateway/internal/domain"
)

// ErrIdempotencyConflict replaces domain.ErrPaymentAlreadyProcessed
var ErrIdempotencyConflict = errors.New("a request is already in progress for this idempotency key")

type PaymentProcessor interface {
    Execute(ctx context.Context, cmd ProcessPaymentCommand) (*ProcessPaymentResult, error)
}

// ProcessPaymentCommand represents the input data for the Use Case.
type ProcessPaymentCommand struct {
	IdempotencyKey string
	PAN            string
	ExpiryMonth    int
	ExpiryYear     int
	CVV            string
	Amount         int64
	Currency       string
}

// ProcessPaymentResult represents the output data from the Use Case.
type ProcessPaymentResult struct {
	ID                 string
	Status             string
	CardNumberLastFour string
	ExpiryMonth        int
	ExpiryYear         int
	Currency           string
	Amount             int64
}

// BankService defines the contract for communicating with the Acquiring Bank.
type BankService interface {
	// Process sends the payment details to the bank.
	// Returns a boolean indicating if the payment was authorized, and an error if the network failed.
	Process(ctx context.Context, amount int64, currency string, pan string, expMonth, expYear int, cvv string) (bool, error)
}

// ProcessPaymentUseCase is the concrete struct that handles the application logic.
type ProcessPaymentUseCase struct {
	repo             domain.PaymentRepository
	idempotencyStore domain.IdempotencyStore
	idGenerator      domain.IDGenerator
	bankService      BankService
}

// 3. Build-time check
var _ PaymentProcessor = (*ProcessPaymentUseCase)(nil)

func NewProcessPaymentUseCase(
	repo domain.PaymentRepository,
	idemStore domain.IdempotencyStore,
	idGen domain.IDGenerator,
	bank BankService,
) *ProcessPaymentUseCase {
	return &ProcessPaymentUseCase{
		repo:             repo,
		idempotencyStore: idemStore,
		idGenerator:      idGen,
		bankService:      bank,
	}
}

func (u *ProcessPaymentUseCase) Execute(ctx context.Context, cmd ProcessPaymentCommand) (*ProcessPaymentResult, error) {
	// --- 1. Idempotency Lock Phase ---
	locked, err := u.idempotencyStore.SetNX(ctx, cmd.IdempotencyKey, []byte(IdempotencyInProgressMarker), 24*time.Hour)
	if err != nil {
		return nil, err
	}
	if !locked {
		// The key already exists! Let's see if it's finished or currently in-flight.
		cachedBytes, err := u.idempotencyStore.Get(ctx, cmd.IdempotencyKey)
		if err != nil {
			return nil, err
		}
		
		if string(cachedBytes) == IdempotencyInProgressMarker {
			// Another request is currently processing this key. Return a Domain Error.
			return nil, ErrIdempotencyConflict
		}

		// It finished previously! Unmarshal the cached JSON result and return it immediately.
		var cachedResult ProcessPaymentResult
		if err := json.Unmarshal(cachedBytes, &cachedResult); err != nil {
			return nil, err
		}
		return &cachedResult, nil
	}

	// --- 2. Domain Validation Phase ---
	card, err := domain.NewCard(cmd.PAN, cmd.ExpiryMonth, cmd.ExpiryYear, cmd.CVV)
	if err != nil {
		return nil, err // Returns domain validation errors (e.g., ErrInvalidCardNumber)
	}

	money, err := domain.NewMoney(cmd.Amount, cmd.Currency)
	if err != nil {
		return nil, err
	}

	// --- 3. Entity Creation Phase ---
	paymentID := u.idGenerator.Generate()
	payment := domain.NewPayment(paymentID, money, card, cmd.IdempotencyKey)

	// --- 4. Bank Interaction Phase ---
	authorized, err := u.bankService.Process(ctx, cmd.Amount, cmd.Currency, cmd.PAN, cmd.ExpiryMonth, cmd.ExpiryYear, cmd.CVV)
	if err != nil {
		// If the bank is down (503), we return the error. 
		// The HTTP handler will translate this into a 502 Bad Gateway or 500 Internal Server Error.
		return nil, err
	}

	// --- 5. State Mutation Phase ---
	if authorized {
		if err := payment.Authorize(); err != nil {
			return nil, err
		}
	} else {
		if err := payment.Decline(); err != nil {
			return nil, err
		}
	}

	// --- 6. Persistence Phase ---
	if err := u.repo.Save(ctx, &payment); err != nil {
		return nil, err
	}

	// --- 7. Cache Update Phase (Final Save) ---
	result := &ProcessPaymentResult{
		ID:                 payment.ID(),
		Status:             string(payment.Status()),
		CardNumberLastFour: payment.Card().Last4(),
		ExpiryMonth:        payment.Card().ExpiryMonth(),
		ExpiryYear:         payment.Card().ExpiryYear(),
		Currency:           payment.Money().Currency(),
		Amount:             payment.Money().Amount(),
	}

	resultBytes, _ := json.Marshal(result)
	// Overwrite the "IN_PROGRESS" lock with the actual completed result
	_ = u.idempotencyStore.Set(ctx, cmd.IdempotencyKey, resultBytes, 24*time.Hour)

	return result, nil
}