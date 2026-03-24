package domain

import "errors"

// --- Validation & Value Object Errors ---
var (
	ErrInvalidAmount       = errors.New("amount must be greater than zero")
	ErrUnsupportedCurrency = errors.New("unsupported currency code")
	ErrInvalidCardNumber   = errors.New("invalid card number")
	ErrCardExpired         = errors.New("card is expired")
	ErrInvalidCVV          = errors.New("invalid CVV")
)

// --- Entity & State Errors ---
var (
	ErrPaymentAlreadyProcessed = errors.New("payment has already been processed and cannot change state")
)

// --- Repository Errors ---
var (
	ErrPaymentNotFound = errors.New("payment not found")
)