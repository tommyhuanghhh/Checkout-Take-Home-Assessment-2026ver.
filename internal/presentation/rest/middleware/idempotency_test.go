package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"PaymentGateway/internal/presentation/rest"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRequireIdempotencyKey(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup a dummy router with the middleware
	r := gin.New()
	r.Use(RequireIdempotencyKey())
	r.POST("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "success")
	})

	t.Run("Header Present - Allows Request", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/test", nil)
		req.Header.Set(rest.HeaderIdempotencyKey, "idem_123")
		
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "success", w.Body.String())
	})

	t.Run("Header Missing - Aborts with 400", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/test", nil)
		// Missing header
		
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Idempotency-Key header is required")
	})
}