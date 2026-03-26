package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewCard(t *testing.T) {
	// The universally accepted valid test card for payment gateways
	validPAN := "4242424242424242"
	futureYear := time.Now().Year() + 1
	currentMonth := int(time.Now().Month())
	currentYear := time.Now().Year()

	tests := []struct {
		name        string
		pan         string
		expMonth    int
		expYear     int
		cvv         string
		expectedErr error
	}{
		{
			name:        "Valid Card",
			pan:         validPAN,
			expMonth:    12,
			expYear:     futureYear,
			cvv:         "123",
			expectedErr: nil,
		},
		{
			name:        "Valid Card with Spaces",
			pan:         "4242 4242 4242 4242",
			expMonth:    12,
			expYear:     futureYear,
			cvv:         "123",
			expectedErr: nil,
		},
		{
			name:        "Valid Card with Dashes",
			pan:         "4242-4242-4242-4242",
			expMonth:    12,
			expYear:     futureYear,
			cvv:         "123",
			expectedErr: nil,
		},
		{
			name:        "Valid Expiry Current Month",
			pan:         validPAN,
			expMonth:    currentMonth,
			expYear:     currentYear,
			cvv:         "123",
			expectedErr: nil,
		},
		{
			name:        "Invalid Luhn Check",
			pan:         "4242424242424243", // Changed last digit to force Luhn failure
			expMonth:    12,
			expYear:     futureYear,
			cvv:         "123",
			expectedErr: ErrInvalidCardNumber,
		},
		{
			name:        "Invalid Length Too Short",
			pan:         "424242424242", // 12 digits
			expMonth:    12,
			expYear:     futureYear,
			cvv:         "123",
			expectedErr: ErrInvalidCardNumber,
		},
		{
			name:        "Expired Card (Past Year)",
			pan:         validPAN,
			expMonth:    12,
			expYear:     2000,
			cvv:         "123",
			expectedErr: ErrCardExpired,
		},
		{
			name:        "Invalid CVV Too Short",
			pan:         validPAN,
			expMonth:    12,
			expYear:     futureYear,
			cvv:         "12",
			expectedErr: ErrInvalidCVV,
		},
		{
			name:        "Invalid CVV Too Long",
			pan:         validPAN,
			expMonth:    12,
			expYear:     futureYear,
			cvv:         "12345",
			expectedErr: ErrInvalidCVV,
		},
		{
			name:        "Letters in PAN",
			pan:         "4242ABCD42424242",
			expMonth:    12,
			expYear:     futureYear,
			cvv:         "123",
			expectedErr: ErrInvalidCardNumber,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card, err := NewCard(tt.pan, tt.expMonth, tt.expYear, tt.cvv)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.cvv, card.CVV())
				assert.Equal(t, tt.expMonth, card.ExpiryMonth())
				assert.Equal(t, tt.expYear, card.ExpiryYear())
				// Test the public Last4() method instead of the unexported pan field!
				assert.Equal(t, "4242", card.Last4())
			}
		})
	}
}

func TestCard_GettersAndFormatting(t *testing.T) {
	// We can use a direct struct initialization here since we are just testing getters,
	// bypassing the NewCard constructor logic.
	card := Card{
		pan:         "4242424242421234",
		expiryMonth: 5,
		expiryYear:  2028,
		cvv:         "456",
	}

	assert.Equal(t, 5, card.ExpiryMonth())
	assert.Equal(t, 2028, card.ExpiryYear())
	assert.Equal(t, "456", card.CVV())
	assert.Equal(t, "1234", card.Last4())

	expectedString := "Card{PAN: ************1234, Exp: 05/2028}"
	assert.Equal(t, expectedString, card.String())
}
