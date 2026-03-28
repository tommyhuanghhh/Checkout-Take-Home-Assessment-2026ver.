package dto


type PostPaymentRequest struct {
	// Validation: Required, must be numbers only, between 14 and 19 digits.
	CardNumber  string `json:"card_number" binding:"required,credit_card"`
	
	// Validation: Required, standard 1-12 months.
	ExpiryMonth int    `json:"expiry_month" binding:"required,min=1,max=12"`
	
	// Validation: Required, must be at least the current year
	ExpiryYear  int    `json:"expiry_year" binding:"required,min=2026"`
	
	// Validation: Required, exactly 3 alphabetical characters (e.g., "USD", "GBP").
	Currency string `json:"currency" binding:"required,iso4217"`
	
	// Validation: Required, must be greater than 0 (minor units/cents).
	Amount      int    `json:"amount" binding:"required,min=1"`
	
	// Validation: Required, 3 to 4 numeric digits.
	Cvv         string `json:"cvv" binding:"required,numeric,min=3,max=4"`
}

// PostPaymentResponse is what we return to the merchant after processing.
type PostPaymentResponse struct {
	Id                 string `json:"id"`
	PaymentStatus      string `json:"payment_status"` // "Authorized", "Declined", "Rejected"
	CardNumberLastFour string `json:"card_number_last_four"`
	ExpiryMonth        int    `json:"expiry_month"`
	ExpiryYear         int    `json:"expiry_year"`
	Currency           string `json:"currency"`
	Amount             int    `json:"amount"`
}