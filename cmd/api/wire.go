//go:build wireinject
// +build wireinject

package main

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"

	"PaymentGateway/internal/application/usecase"
	"PaymentGateway/internal/domain"
	"PaymentGateway/internal/infrastructure/acquiring_bank"
	"PaymentGateway/internal/infrastructure/inmemory"
	"PaymentGateway/internal/infrastructure/uuid"
	"PaymentGateway/internal/presentation/rest"
	"PaymentGateway/internal/presentation/rest/handler"
)

// InitializeRouter wires up the dependencies and returns a configured Gin engine.
func InitializeRouter(
	logger *slog.Logger,
	redisClient *redis.Client,
	httpClient *http.Client,
	bankSimulatorURL string, // This maps directly to the 'baseURL' param in NewSimulatorClient
) *gin.Engine {
	wire.Build(
		// 1. Presentation Layer
		rest.NewRouter,
		handler.NewPaymentHandler,
		handler.NewRetrievePaymentHandler,

		// 2. Application Layer (Use Cases & Interface Bindings)
		usecase.NewProcessPaymentUseCase,
		wire.Bind(new(usecase.PaymentProcessor), new(*usecase.ProcessPaymentUseCase)),

		usecase.NewRetrievePaymentUseCase,
		wire.Bind(new(usecase.PaymentRetriever), new(*usecase.RetrievePaymentUseCase)),

		// 3. Infrastructure Layer (Adapters & Interface Bindings)
		inmemory.NewInMemoryPaymentRepository,
		wire.Bind(new(domain.PaymentRepository), new(*inmemory.InMemoryPaymentRepository)),

		inmemory.NewRedisIdempotencyStore,
		wire.Bind(new(domain.IdempotencyStore), new(*inmemory.RedisIdempotencyStore)),

		uuid.NewUUIDGenerator,
		wire.Bind(new(domain.IDGenerator), new(*uuid.UUIDGenerator)),

		acquiring_bank.NewSimulatorClient,
		wire.Bind(new(usecase.BankService), new(*acquiring_bank.SimulatorClient)),
	)
	return nil
}
