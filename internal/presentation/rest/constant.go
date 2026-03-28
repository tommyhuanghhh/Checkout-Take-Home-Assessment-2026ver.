package rest

const (
	// Headers
	HeaderIdempotencyKey = "Idempotency-Key"

	// Gin Context Keys
	ContextKeyPaymentID = "payment_id"

	// Structured Logging Fields (Semantic Conventions)
	LogFieldMethod         = "method"
	LogFieldPath           = "path"
	LogFieldStatus         = "status"
	LogFieldLatency        = "latency"
	LogFieldClientIP       = "client_ip"
	LogFieldIdempotencyKey = "idempotency_key"
	LogFieldPaymentID      = "payment_id"
	LogFieldErrors         = "errors"

	//URL Param
	URLParamID = "id"

)