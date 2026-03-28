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

type RetrievePaymentHandler struct {
	useCase usecase.PaymentRetriever
}

func NewRetrievePaymentHandler(u usecase.PaymentRetriever) *RetrievePaymentHandler {
	return &RetrievePaymentHandler{
		useCase: u,
	}
}

func (h *RetrievePaymentHandler) RetrievePayment(c *gin.Context) {
	// 1. Set a strict timeout for the database read operation (e.g., 5 seconds).
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// 2. Extract the Payment ID from the Gin URL parameter
	id := c.Param(rest.URLParamID)
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payment ID parameter is required"})
		return
	}

	// 3. Map to the Use Case Command
	cmd := usecase.RetrievePaymentCommand{
		ID: id,
	}

	// 4. Execute the Use Case
	result, err := h.useCase.Execute(ctx, cmd)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// 5. Map the Result back to the GET Response DTO
	resp := dto.GetPaymentResponse{
		Id:                 result.ID,
		PaymentStatus:      result.Status,
		CardNumberLastFour: result.CardNumberLastFour,
		ExpiryMonth:        result.ExpiryMonth,
		ExpiryYear:         result.ExpiryYear,
		Currency:           result.Currency,
		Amount:             int(result.Amount),
	}

	// 6. Return 200 OK
	c.JSON(http.StatusOK, resp)
}

// handleError safely translates domain and system errors into RESTful HTTP status codes.
func (h *RetrievePaymentHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrPaymentNotFound):
		// 404 is the only correct response when an ID doesn't exist in a REST API.
		c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})

	case errors.Is(err, context.DeadlineExceeded):
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": "database read timeout"})

	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}