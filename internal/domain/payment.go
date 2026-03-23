package domain

type Payment struct {
	ID                 string
	Status             string
	CardNumberLastFour int
	ExpiryMonth        int
	ExpiryYear         int
	Currency           string
	Amount             int
}
