package handler

import (
	"context"
	"errors"
	"net/http"
	"time"

	"PaymentGateway/internal/application/usecase"
	"PaymentGateway/internal/domain"
	"PaymentGateway/internal/presentation/rest"
	"PaymentGateway/internal/presentation/rest/dto"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	useCase usecase.ProcessPaymentUseCase
}

func NewPaymentHandler(u usecase.ProcessPaymentUseCase) *PaymentHandler {
	return &PaymentHandler{
		useCase: u,
	}
}

func (h *PaymentHandler) ProcessPayment(c *gin.Context) {
	// 1. Set a hard timeout for the entire request lifecycle (e.g., 10 seconds).
	// This ensures we don't hang the client if downstream services are unresponsive.
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// 2. Extract Idempotency Key (Required for Phase 1 Validation)
	idempotencyKey := c.GetHeader(rest.HeaderIdempotencyKey)
	if idempotencyKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Idempotency-Key header is required"})
		return
	}

	// 3. Bind and Validate JSON DTO
	var req dto.PostPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Gin uses validator v10 under the hood; this catches bad cards, 
		// invalid currencies, and malformed JSON before hitting our Use Case.
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 4. Map DTO to Use Case Command
	cmd := usecase.ProcessPaymentCommand{
		IdempotencyKey: idempotencyKey,
		PAN:            req.CardNumber,
		ExpiryMonth:    req.ExpiryMonth,
		ExpiryYear:     req.ExpiryYear,
		CVV:            req.Cvv,
		Amount:         int64(req.Amount), // Cast to int64 as required by usecase
		Currency:       req.Currency,
	}

	// 5. Execute Use Case with the Timeout Context
	result, err := h.useCase.Execute(ctx, cmd)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// 6. Map Result back to Response DTO
	resp := dto.PostPaymentResponse{
		Id:                 result.ID,
		PaymentStatus:      result.Status,
		CardNumberLastFour: result.CardNumberLastFour,
		ExpiryMonth:        result.ExpiryMonth,
		ExpiryYear:         result.ExpiryYear,
		Currency:           result.Currency,
		Amount:             int(result.Amount),
	}

	c.JSON(http.StatusCreated, resp)
}

// handleError maps internal errors to the correct HTTP status codes.
func (h *PaymentHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, usecase.ErrIdempotencyConflict):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	
	case errors.Is(err, domain.ErrInvalidCardNumber), 
	     errors.Is(err, domain.ErrInvalidCVV), 
	     errors.Is(err, domain.ErrCardExpired):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

	case errors.Is(err, context.DeadlineExceeded):
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": "request timed out"})

	default:
		// Log the actual error here in a real app
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}