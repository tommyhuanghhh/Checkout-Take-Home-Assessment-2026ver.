// internal/domain/money.go
package domain

// The Bounded Whitelist required by the assessment!
var supportedCurrencies = map[string]bool{
	"USD": true,
	"EUR": true,
	"GBP": true,
}

// Money is an immutable Value Object representing a monetary value in its minor unit.
type Money struct {
	amount   int64  // Represents cents/pence (no decimals!)
	currency string // 3-letter ISO code
}

// NewMoney is the constructor that enforces our domain rules.
func NewMoney(amount int64, currency string) (Money, error) {
	if amount <= 0 {
		return Money{}, ErrInvalidAmount
	}

	if !supportedCurrencies[currency] {
		return Money{}, ErrUnsupportedCurrency
	}

	return Money{
		amount:   amount,
		currency: currency,
	}, nil
}

// Getters
func (m Money) Amount() int64      { return m.amount }
func (m Money) Currency() string   { return m.currency }