package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"PaymentGateway/internal/application/usecase"
	"PaymentGateway/internal/presentation/rest"
	"PaymentGateway/internal/presentation/rest/dto"
	"PaymentGateway/internal/presentation/rest/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockPaymentProcessor is a testify mock for the usecase.PaymentProcessor interface
type mockPaymentProcessor struct {
	mock.Mock
}

func (m *mockPaymentProcessor) Execute(ctx context.Context, cmd usecase.ProcessPaymentCommand) (*usecase.ProcessPaymentResult, error) {
	args := m.Called(ctx, cmd)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecase.ProcessPaymentResult), args.Error(1)
}

var _ usecase.PaymentProcessor = (*mockPaymentProcessor)(nil)

// setupRouter wires the middleware and handler together for testing
func setupRouter(h *PaymentHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// Inject the middleware just like we will in main.go!
	r.Use(middleware.RequireIdempotencyKey())
	r.POST("/payments", h.ProcessPayment)
	return r
}

func TestPaymentHandler_ProcessPayment(t *testing.T) {
	validDTO := dto.PostPaymentRequest{
		CardNumber:  "4242424242424242",
		ExpiryMonth: 12,
		ExpiryYear:  2028,
		Currency:    "USD",
		Amount:      1000,
		Cvv:         "123",
	}

	t.Run("Success - 201 Created", func(t *testing.T) {
		m := new(mockPaymentProcessor)
		handler := NewPaymentHandler(m)
		router := setupRouter(handler)

		expectedResult := &usecase.ProcessPaymentResult{
			ID:                 "pay_123",
			Status:             "Authorized",
			CardNumberLastFour: "4242",
		}

		m.On("Execute", mock.Anything, mock.Anything).Return(expectedResult, nil)

		body, _ := json.Marshal(validDTO)
		req, _ := http.NewRequest(http.MethodPost, "/payments", bytes.NewBuffer(body))
		req.Header.Set(rest.HeaderIdempotencyKey, "test-key")
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var resp dto.PostPaymentResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "pay_123", resp.Id)
		m.AssertExpectations(t)
	})

	t.Run("Failure - Missing Idempotency Key (Intercepted by Middleware)", func(t *testing.T) {
		m := new(mockPaymentProcessor)
		handler := NewPaymentHandler(m)
		router := setupRouter(handler)

		body, _ := json.Marshal(validDTO)
		req, _ := http.NewRequest(http.MethodPost, "/payments", bytes.NewBuffer(body))
		// Explicitly NOT setting the Idempotency-Key header

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Idempotency-Key header is required")
		
		// Prove the handler/usecase was NEVER called!
		m.AssertNotCalled(t, "Execute") 
	})

	t.Run("Failure - Invalid Card (Phase 1 Validation)", func(t *testing.T) {
		m := new(mockPaymentProcessor)
		handler := NewPaymentHandler(m)
		router := setupRouter(handler)

		invalidDTO := validDTO
		invalidDTO.CardNumber = "123" // Too short

		body, _ := json.Marshal(invalidDTO)
		req, _ := http.NewRequest(http.MethodPost, "/payments", bytes.NewBuffer(body))
		req.Header.Set(rest.HeaderIdempotencyKey, "test-key")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "CardNumber")
		m.AssertNotCalled(t, "Execute")
	})

	t.Run("Failure - Idempotency Conflict (409)", func(t *testing.T) {
		m := new(mockPaymentProcessor)
		handler := NewPaymentHandler(m)
		router := setupRouter(handler)

		m.On("Execute", mock.Anything, mock.Anything).Return(nil, usecase.ErrIdempotencyConflict)

		body, _ := json.Marshal(validDTO)
		req, _ := http.NewRequest(http.MethodPost, "/payments", bytes.NewBuffer(body))
		req.Header.Set(rest.HeaderIdempotencyKey, "test-key")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
	})
}