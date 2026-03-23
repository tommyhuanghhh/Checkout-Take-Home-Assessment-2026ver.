package domain


type PaymentsRepository interface {
	GetPayment(id string) *Payment
	AddPayment(payment Payment) error
}
