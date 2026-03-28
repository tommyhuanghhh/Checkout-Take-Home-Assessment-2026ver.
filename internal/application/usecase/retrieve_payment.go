package usecase

import (
	"context"

	"PaymentGateway/internal/domain"
)

// Inside application/usecase/retrieve_payment.go
type PaymentRetriever interface {
    Execute(ctx context.Context, query RetrievePaymentCommand) (*RetrievePaymentResult, error)
}

type RetrievePaymentCommand struct {
	ID string
}

type RetrievePaymentResult struct {
	ID                 string
	Status             string
	CardNumberLastFour string
	ExpiryMonth        int
	ExpiryYear         int
	Currency           string
	Amount             int64
}

type RetrievePaymentUseCase struct {
	repo domain.PaymentRepository
}

var _ PaymentRetriever = (*RetrievePaymentUseCase)(nil)

func NewRetrievePaymentUseCase(repo domain.PaymentRepository) *RetrievePaymentUseCase {
	return &RetrievePaymentUseCase{
		repo: repo,
	}
}

func (u *RetrievePaymentUseCase) Execute(ctx context.Context, cmd RetrievePaymentCommand) (*RetrievePaymentResult, error) {
	// 1. Fetch the Entity from the Repository
	payment, err := u.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		// If the payment isn't found, this will return domain.ErrPaymentNotFound
		return nil, err
	}

	// 2. Map the Domain Entity to the Use Case Result DTO
	result := &RetrievePaymentResult{
		ID:                 payment.ID(),
		Status:             string(payment.Status()),
		CardNumberLastFour: payment.Card().Last4(),
		ExpiryMonth:        payment.Card().ExpiryMonth(),
		ExpiryYear:         payment.Card().ExpiryYear(),
		Currency:           payment.Money().Currency(),
		Amount:             payment.Money().Amount(),
	}

	// 3. Return the decoupled result
	return result, nil
}