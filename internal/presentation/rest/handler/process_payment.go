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
	useCase usecase.PaymentProcessor
}

func NewPaymentHandler(u usecase.PaymentProcessor) *PaymentHandler {
	return &PaymentHandler{
		useCase: u,
	}
}

func (h *PaymentHandler) ProcessPayment(c *gin.Context) {
	// 1. Set a hard timeout for the entire request lifecycle (e.g., 10 seconds).
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// 2. Extract Idempotency Key
	// Note: We no longer need to check if this is empty. The RequireIdempotencyKey 
	// middleware mathematically guarantees this header exists before this handler is ever called.
	idempotencyKey := c.GetHeader(rest.HeaderIdempotencyKey)

	// 3. Bind and Validate JSON DTO
	var req dto.PostPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
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
		Amount:         int64(req.Amount), 
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

	c.Set(rest.ContextKeyPaymentID, result.ID)
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}