package dto

// GetPaymentResponse represents a previously processed payment.
// Note: While structurally identical to PostPaymentResponse right now, 
// keeping them separate is a Clean Architecture best practice so they can evolve independently.
type GetPaymentResponse struct {
	Id                 string `json:"id"`
	PaymentStatus      string `json:"payment_status"`
	CardNumberLastFour string `json:"card_number_last_four"`
	ExpiryMonth        int    `json:"expiry_month"`
	ExpiryYear         int    `json:"expiry_year"`
	Currency           string `json:"currency"`
	Amount             int    `json:"amount"`
}