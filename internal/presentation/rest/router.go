package rest

import (
	"log/slog"

	"PaymentGateway/internal/presentation/rest/handler"
	"PaymentGateway/internal/presentation/rest/middleware"

	"github.com/gin-gonic/gin"
)

// NewRouter constructs the Gin engine, attaches global middleware,
// and maps HTTP routes to their respective handlers.
func NewRouter(
	logger *slog.Logger,
	paymentHandler *handler.PaymentHandler,
	retrieveHandler *handler.RetrievePaymentHandler,
) *gin.Engine {

	// Set Gin to Release Mode in production to suppress debug output.
	// (You can make this dynamic via environment variables in main.go later)
	gin.SetMode(gin.ReleaseMode)

	// Initialize a completely blank Gin engine
	router := gin.New()

	// --- Global Middleware ---
	// 1. Recovery prevents the server from crashing on panics, returning a 500 instead.
	router.Use(gin.Recovery())
	// 2. Our custom structured, PCI-compliant JSON access logger.
	router.Use(middleware.AccessLogger(logger))

	// 1. Create the API Version group
	v1 := router.Group("/v1")
	{
		// 2. Create the Payments collection group under v1
		payments := v1.Group("/payments")
		{
			// POST /v1/payments
			payments.POST("", middleware.RequireIdempotencyKey(), paymentHandler.ProcessPayment)

			// GET /v1/payments/:id
			payments.GET("/:id", retrieveHandler.RetrievePayment)
		}
	}

	return router
}
