package middleware

import (
	"net/http"

	"PaymentGateway/internal/presentation/rest"

	"github.com/gin-gonic/gin"
)

// RequireIdempotencyKey ensures that the request contains an Idempotency-Key header.
// If it is missing, it immediately aborts the request with a 400 Bad Request.
func RequireIdempotencyKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		idempotencyKey := c.GetHeader(rest.HeaderIdempotencyKey)
		
		if idempotencyKey == "" {
			// AbortWithStatusJSON prevents any downstream handlers or Use Cases from executing
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Idempotency-Key header is required"})
			return
		}

		// Header exists, allow the request to proceed to the handler
		c.Next()
	}
}