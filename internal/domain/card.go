package domain

import (
	"fmt"
	"strings"
	"time"
)

// Card is an immutable Value Object representing a credit card.
type Card struct {
	pan         string // Primary Account Number (unexported for security)
	expiryMonth int
	expiryYear  int
	cvv         string // Unexported
}

// NewCard acts as the constructor and the ultimate validation gate.
func NewCard(pan string, expMonth, expYear int, cvv string) (Card, error) {
	// 1. Remove any spaces or dashes the client might have sent
	cleanPAN := strings.ReplaceAll(pan, " ", "")
	cleanPAN = strings.ReplaceAll(cleanPAN, "-", "")

	// 2. Validate PAN (Luhn Check)
	if !isValidLuhn(cleanPAN) {
		return Card{}, ErrInvalidCardNumber
	}

	// 3. Validate Expiry Date
	if isExpired(expMonth, expYear) {
		return Card{}, ErrCardExpired
	}

	// 4. Validate CVV (Basic length check)
	if len(cvv) < 3 || len(cvv) > 4 {
		return Card{}, ErrInvalidCVV
	}

	return Card{
		pan:         cleanPAN,
		expiryMonth: expMonth,
		expiryYear:  expYear,
		cvv:         cvv,
	}, nil
}

// Getters (Notice there are NO setters!)

func (c Card) ExpiryMonth() int { return c.expiryMonth }
func (c Card) ExpiryYear() int  { return c.expiryYear }
func (c Card) CVV() string      { return c.cvv }

// Last4 returns the last 4 digits for safe API responses
func (c Card) Last4() string {
	if len(c.pan) < 4 {
		return ""
	}
	return c.pan[len(c.pan)-4:]
}

// String implements the fmt.Stringer interface.
// THIS IS CRITICAL FOR PCI COMPLIANCE! If a dev accidentally logs the Card struct,
// it will automatically call this method instead of leaking the raw PAN.
func (c Card) String() string {
	return fmt.Sprintf("Card{PAN: ************%s, Exp: %02d/%d}", c.Last4(), c.expiryMonth, c.expiryYear)
}

// --- Private Validation Helpers ---

func isExpired(month, year int) bool {
	now := time.Now()
	// Ensure we are checking against the end of the expiry month
	// Example: Expiry 10/2026 means valid until Oct 31, 2026.
	expiryDate := time.Date(year, time.Month(month+1), 1, 0, 0, 0, 0, time.UTC).Add(-time.Nanosecond)
	return now.After(expiryDate)
}

// isValidLuhn implements the Luhn algorithm to validate credit card numbers.
func isValidLuhn(number string) bool {
	if len(number) < 13 || len(number) > 19 {
		return false
	}

	var sum int
	alternate := false

	// Iterate backwards through the digits
	for i := len(number) - 1; i >= 0; i-- {
		// Convert byte character to integer
		digit := int(number[i] - '0')
		if digit < 0 || digit > 9 {
			return false // Contains non-numeric characters
		}

		if alternate {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		alternate = !alternate
	}

	return sum%10 == 0
}