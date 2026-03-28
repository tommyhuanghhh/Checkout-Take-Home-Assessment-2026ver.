package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"PaymentGateway/internal/application/usecase"
	"PaymentGateway/internal/domain"
	"PaymentGateway/internal/presentation/rest/dto"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockPaymentRetriever is a testify mock for the usecase.PaymentRetriever interface
type mockPaymentRetriever struct {
	mock.Mock
}

func (m *mockPaymentRetriever) Execute(ctx context.Context, cmd usecase.RetrievePaymentCommand) (*usecase.RetrievePaymentResult, error) {
	args := m.Called(ctx, cmd)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecase.RetrievePaymentResult), args.Error(1)
}

// Build-time check to ensure our mock matches the handler's required interface
var _ usecase.PaymentRetriever = (*mockPaymentRetriever)(nil)

func TestRetrievePaymentHandler_RetrievePayment(t *testing.T) {
	// Set Gin to Test Mode so it doesn't pollute the console output
	gin.SetMode(gin.TestMode)

	t.Run("Success - 200 OK", func(t *testing.T) {
		m := new(mockPaymentRetriever)
		handler := NewRetrievePaymentHandler(m)

		// Setup Gin Router
		r := gin.New()
		r.GET("/payments/:id", handler.RetrievePayment)

		expectedResult := &usecase.RetrievePaymentResult{
			ID:                 "pay_123",
			Status:             "Authorized",
			CardNumberLastFour: "4242",
			ExpiryMonth:        12,
			ExpiryYear:         2028,
			Currency:           "USD",
			Amount:             1000,
		}

		// Mock Expectation
		m.On("Execute", mock.Anything, usecase.RetrievePaymentCommand{ID: "pay_123"}).Return(expectedResult, nil)

		// Execute Request
		req, _ := http.NewRequest(http.MethodGet, "/payments/pay_123", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)

		var resp dto.GetPaymentResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "pay_123", resp.Id)
		assert.Equal(t, "Authorized", resp.PaymentStatus)
		assert.Equal(t, "4242", resp.CardNumberLastFour)
		
		m.AssertExpectations(t)
	})

	t.Run("Failure - 404 Not Found", func(t *testing.T) {
		m := new(mockPaymentRetriever)
		handler := NewRetrievePaymentHandler(m)

		r := gin.New()
		r.GET("/payments/:id", handler.RetrievePayment)

		// Mock the usecase returning the specific domain error
		m.On("Execute", mock.Anything, usecase.RetrievePaymentCommand{ID: "pay_invalid"}).
			Return(nil, domain.ErrPaymentNotFound)

		req, _ := http.NewRequest(http.MethodGet, "/payments/pay_invalid", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "payment not found")
	})

	t.Run("Failure - 500 Internal Server Error (Unexpected infrastructure issue)", func(t *testing.T) {
		m := new(mockPaymentRetriever)
		handler := NewRetrievePaymentHandler(m)

		r := gin.New()
		r.GET("/payments/:id", handler.RetrievePayment)

		// Mock a generic error (e.g., database connection dropped)
		m.On("Execute", mock.Anything, mock.Anything).
			Return(nil, errors.New("some unexpected database error"))

		req, _ := http.NewRequest(http.MethodGet, "/payments/pay_error", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "internal server error")
	})
}