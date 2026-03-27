package acquiring_bank

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"PaymentGateway/internal/application/usecase"
)

// Build-time check to ensure SimulatorClient implements the usecase.BankService interface.
var _ usecase.BankService = (*SimulatorClient)(nil)

// SimulatorClient is the concrete HTTP adapter for communicating with the external bank simulator.
type SimulatorClient struct {
	client  *http.Client
	baseURL string
}

// NewSimulatorClient creates a new bank simulator adapter. 
// We inject the http.Client so we can configure custom connection pooling in main.go.
func NewSimulatorClient(baseURL string, client *http.Client) *SimulatorClient {
	return &SimulatorClient{
		baseURL: baseURL,
		client:  client,
	}
}

// bankRequest represents the exact JSON structure expected by the simulator.
type bankRequest struct {
	CardNumber string `json:"card_number"`
	ExpiryDate string `json:"expiry_date"` // Must be formatted as "MM/YYYY"
	Currency   string `json:"currency"`
	Amount     int64  `json:"amount"`
	CVV        string `json:"cvv"`
}

// bankResponse represents the exact JSON structure returned by the simulator.
type bankResponse struct {
	Authorized        bool   `json:"authorized"`
	AuthorizationCode string `json:"authorization_code,omitempty"`
}

// Process sends the HTTP POST request to the bank simulator.
func (s *SimulatorClient) Process(ctx context.Context, amount int64, currency string, pan string, expMonth, expYear int, cvv string) (bool, error) {
	// 1. Map Use Case primitives to the Bank's specific HTTP Request format
	reqBody := bankRequest{
		CardNumber: pan,
		ExpiryDate: fmt.Sprintf("%02d/%d", expMonth, expYear), // e.g., "04/2025"
		Currency:   currency,
		Amount:     amount,
		CVV:        cvv,
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return false, fmt.Errorf("failed to marshal bank request: %w", err)
	}

	// 2. Construct the HTTP Request using the provided Context (for timeouts/cancellations)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/payments", bytes.NewReader(jsonBytes))
	if err != nil {
		return false, fmt.Errorf("failed to create bank request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// 3. Execute the Request using our pooled client
	resp, err := s.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("bank network error: %w", err)
	}
	defer resp.Body.Close()

	// 4. Handle HTTP Status Codes based on simulator rules
	if resp.StatusCode == http.StatusServiceUnavailable {
		return false, fmt.Errorf("bank service unavailable (503)")
	}
	if resp.StatusCode == http.StatusBadRequest {
		return false, fmt.Errorf("bank rejected request as malformed (400)")
	}
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected bank status code: %d", resp.StatusCode)
	}

	// 5. Decode the 200 OK Response
	var result bankResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("failed to decode bank response: %w", err)
	}

	return result.Authorized, nil
}