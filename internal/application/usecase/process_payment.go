package usecases

import (
	"PaymentGateway/internal/domain"
)

type PaymentsUsecase interface {
	GetPayment(id string) *domain.Payment
	// We'll add more methods like ProcessPayment later
}

type AcquiringBank interface {
	ProcessPayment(payment domain.Payment) error
}

type paymentsUsecase struct {
	repo domain.PaymentsRepository
}

func NewPaymentsUsecase(repo domain.PaymentsRepository) PaymentsUsecase {
	return &paymentsUsecase{repo: repo}
}

func (u *paymentsUsecase) GetPayment(id string) *domain.Payment {
	return u.repo.GetPayment(id)
}
