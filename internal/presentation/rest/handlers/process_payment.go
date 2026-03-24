package handlers

import (
	"encoding/json"
	"net/http"

	"PaymentGateway/internal/application/usecases"
	"PaymentGateway/internal/presentation/rest/dtos"

	"github.com/go-chi/chi/v5"
)

type PaymentsHandler struct {
	usecase usecases.PaymentsUsecase
}

func NewPaymentsHandler(usecase usecases.PaymentsUsecase) *PaymentsHandler {
	return &PaymentsHandler{
		usecase: usecase,
	}
}

// GetHandler returns an http.HandlerFunc that handles HTTP GET requests.
// It retrieves a payment record by its ID from the storage.
// The ID is expected to be part of the URL.
func (h *PaymentsHandler) GetHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		payment := h.usecase.GetPayment(id)

		if payment != nil {
			response := dtos.GetPaymentResponse{
				Id:                 payment.ID,
				PaymentStatus:      payment.Status,
				CardNumberLastFour: payment.CardNumberLastFour,
				ExpiryMonth:        payment.ExpiryMonth,
				ExpiryYear:         payment.ExpiryYear,
				Currency:           payment.Currency,
				Amount:             payment.Amount,
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(response); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

func (ph *PaymentsHandler) PostHandler() http.HandlerFunc {
	//TODO
	return nil
}
