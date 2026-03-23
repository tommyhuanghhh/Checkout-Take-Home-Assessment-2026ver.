package repository

import (
	"PaymentGateway/internal/domain"
)

type memoryPaymentsRepository struct {
	payments []domain.Payment
}

// Ensure memoryPaymentsRepository implements domainRepo.PaymentsRepository
var _ domain.PaymentsRepository = (*memoryPaymentsRepository)(nil)

func NewPaymentsRepository() domain.PaymentsRepository {
	return &memoryPaymentsRepository{
		payments: []domain.Payment{},
	}
}

func (ps *memoryPaymentsRepository) GetPayment(id string) *domain.Payment {
	for _, element := range ps.payments {
		if element.ID == id {
			return &element
		}
	}
	return nil
}

func (ps *memoryPaymentsRepository) AddPayment(payment domain.Payment) error {
	ps.payments = append(ps.payments, payment)
	return nil
}
