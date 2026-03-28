package acquiring_bank

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSimulatorClient_Process(t *testing.T) {
	ctx := context.Background()

	t.Run("Happy Path - Successfully formats request and parses 200 OK", func(t *testing.T) {
		// 1. Create a real local HTTP server to act as the Bank Simulator
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify the method and path
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/payments", r.URL.Path)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			// Read and verify the exact JSON body our client generated
			bodyBytes, _ := io.ReadAll(r.Body)
			var reqBody bankRequest
			err := json.Unmarshal(bodyBytes, &reqBody)
			assert.NoError(t, err)

			// PROVE THE DATE FORMATTING WORKED!
			assert.Equal(t, "04/2025", reqBody.ExpiryDate)
			assert.Equal(t, "4242424242424242", reqBody.CardNumber)
			assert.Equal(t, int64(1500), reqBody.Amount)

			// Return a successful bank response
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"authorized": true, "authorization_code": "auth_123"}`))
		}))
		defer server.Close() // Shut down the server when the test finishes

		// 2. Inject the dynamically generated mock server URL into our client
		client := NewSimulatorClient(server.URL, server.Client())

		// 3. Execute
		authorized, err := client.Process(ctx, 1500, "USD", "4242424242424242", 4, 2025, "123")

		// 4. Assert
		assert.NoError(t, err)
		assert.True(t, authorized)
	})

	t.Run("Sad Path - Handles 503 Service Unavailable", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable) // Bank goes down!
		}))
		defer server.Close()

		client := NewSimulatorClient(server.URL, server.Client())

		authorized, err := client.Process(ctx, 1500, "USD", "4242424242424242", 12, 2026, "123")

		assert.ErrorContains(t, err, "bank service unavailable (503)")
		assert.False(t, authorized)
	})

	t.Run("Sad Path - Request Cancelled via Context Timeout", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond) // Simulate slow bank
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewSimulatorClient(server.URL, server.Client())

		// Create a context that instantly cancels
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel() // Cancel before even sending!

		authorized, err := client.Process(cancelCtx, 1500, "USD", "4242", 12, 2026, "123")

		// Prove that the client aborts and bubbles up the context error
		assert.ErrorContains(t, err, "context canceled")
		assert.False(t, authorized)
	})

	t.Run("Sad Path - Handles 400 Bad Request (Malformed Payload)", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Simulate the bank rejecting our formatting or payload
			w.WriteHeader(http.StatusBadRequest) 
		}))
		defer server.Close()

		client := NewSimulatorClient(server.URL, server.Client())

		// Execute with some data
		authorized, err := client.Process(ctx, 1500, "USD", "4242424242424242", 12, 2026, "123")

		// Prove that we properly map the 400 status code to our specific error
		assert.ErrorContains(t, err, "bank rejected request as malformed (400)")
		assert.False(t, authorized)
	})
}