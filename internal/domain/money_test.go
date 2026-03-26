package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMoney(t *testing.T) {
	tests := []struct {
		name        string
		amount      int64
		currency    string
		expectedErr error
	}{
		{
			name:        "Valid USD",
			amount:      1000,
			currency:    "USD",
			expectedErr: nil,
		},
		{
			name:        "Valid EUR",
			amount:      500,
			currency:    "EUR",
			expectedErr: nil,
		},
		{
			name:        "Valid GBP",
			amount:      100,
			currency:    "GBP",
			expectedErr: nil,
		},
		{
			name:        "Invalid Amount - Zero",
			amount:      0,
			currency:    "USD",
			expectedErr: ErrInvalidAmount,
		},
		{
			name:        "Invalid Amount - Negative",
			amount:      -50,
			currency:    "EUR",
			expectedErr: ErrInvalidAmount,
		},
		{
			name:        "Unsupported Currency - CAD",
			amount:      1000,
			currency:    "CAD",
			expectedErr: ErrUnsupportedCurrency,
		},
		{
			name:        "Unsupported Currency - Lowercase",
			amount:      1000,
			currency:    "usd",
			expectedErr: ErrUnsupportedCurrency,
		},
		{
			name:        "Unsupported Currency - Empty",
			amount:      1000,
			currency:    "",
			expectedErr: ErrUnsupportedCurrency,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			money, err := NewMoney(tt.amount, tt.currency)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.amount, money.Amount())
				assert.Equal(t, tt.currency, money.Currency())
			}
		})
	}
}

func TestMoney_Getters(t *testing.T) {
	money := Money{amount: 5000, currency: "GBP"}
	assert.Equal(t, int64(5000), money.Amount())
	assert.Equal(t, "GBP", money.Currency())
}
