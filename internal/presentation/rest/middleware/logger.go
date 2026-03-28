package middleware

import (
	"log/slog"
	"time"

	"PaymentGateway/internal/presentation/rest"

	"github.com/gin-gonic/gin"
)

func AccessLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 1. Process the request
		c.Next()

		// 2. Capture response metadata
		latency := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path
		clientIP := c.ClientIP()

		// 3. Extract Correlation IDs
		idempotencyKey := c.GetHeader(rest.HeaderIdempotencyKey)
		
		paymentID := c.Param(rest.URLParamID)
		if paymentID == "" {
			paymentID = c.GetString(rest.ContextKeyPaymentID) 
		}

		errs := c.Errors.ByType(gin.ErrorTypePrivate).String()

		logLevel := slog.LevelInfo
		if status >= 400 && status < 500 {
			logLevel = slog.LevelWarn
		} else if status >= 500 {
			logLevel = slog.LevelError
		}

		// 4. Build the log attributes dynamically using unified semantic keys
		attrs := []any{
			slog.String(rest.LogFieldMethod, method),
			slog.String(rest.LogFieldPath, path),
			slog.Int(rest.LogFieldStatus, status),
			slog.Duration(rest.LogFieldLatency, latency),
			slog.String(rest.LogFieldClientIP, clientIP),
		}

		if idempotencyKey != "" {
			attrs = append(attrs, slog.String(rest.LogFieldIdempotencyKey, idempotencyKey))
		}
		if paymentID != "" {
			attrs = append(attrs, slog.String(rest.LogFieldPaymentID, paymentID))
		}
		if errs != "" {
			attrs = append(attrs, slog.String(rest.LogFieldErrors, errs))
		}

		// 5. Write the structured log
		logger.Log(c.Request.Context(), logLevel, "HTTP Access Log", attrs...)
	}
}