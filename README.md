# Checkout Take-Home Assessment (2026) - Payment Gateway

## 1. Requirements Summary
Goal: Build an API-based Payment Gateway application that processes payments and retrieves payment details.

- Core Features:

    - Process a Payment: Accept payment requests and validate inputs (card number, expiry date, currency, amount, CVV). Forward the request to an acquiring bank (via the simulator).

    - Retrieve Payment Details: Allow clients to fetch details of a previously processed payment.

    - Bank Simulator Integration: Interaction with a provided Bank Simulator to process the payments.

- Constraints & Notes:

    - Work off the provided skeleton API.

    - DO NOT modify the imposters/ directory or .editorconfig.

    - Freedom to structure the project, apply architectural patterns, and choose testing libraries.

    - The solution must compile, run, and include automated tests.

## 2. Architecture: Conforming to Clean Architecture
This project strictly adheres to Clean Architecture (specifically the Ports and Adapters / Hexagonal pattern), ensuring business logic is decoupled from external frameworks.

**Project Structure**
- internal/domain (Enterprise Business Rules): Pure Go. Contains core entities (Payment, Card, Money) and custom domain errors. Defines "Output Ports" (Interfaces) for repositories.

- internal/application/usecase (Application Business Rules): Orchestrates data flow. Defines "Input Ports" (PaymentProcessor, PaymentRetriever) called by the Presentation layer.

- internal/infrastructure (Adapters): Implements Domain interfaces. Contains the HTTP client for the Acquiring Bank, in-memory database, and UUID generator.

- internal/presentation/rest (Delivery Mechanism): The HTTP layer (built with Gin). Translates web requests into Application Commands.

- cmd/api (Composition Root): Uses Google Wire to inject dependencies and wire layers together before booting the server.


## 3. Key Engineering Decisions
- Comprehensive Unit Testing: High coverage across all layers using table-driven tests. Handlers are tested via httptest to verify boundaries without side effects.

- Zero-Dependency Local Execution: Integrated alicebob/miniredis/v2 for a dynamic in-memory Redis instance. No external Docker network or Redis install required for native testing.

- Compile-Time Dependency Injection: Used google/wire to generate the dependency graph, ensuring missing dependencies fail at compile-time rather than runtime.

- Idempotency via Middleware: POST /v1/payments is protected by IdempotencyMiddleware. Replayed requests return cached responses instantly, bypassing the Bank Simulator.

- PCI-DSS Compliant Logging: Utilizes log/slog. Logs metadata and Correlation IDs while stripping sensitive JSON payload bodies (PAN/CVV) to prevent data leaks.

- Two-Phase Validation:
    - Structural: Gin’s Validator catches malformed fields (400 Bad Request).
    - Domain: Entities catch business violations (e.g., expired cards) ensuring domain invariants.

- Graceful Shutdown: Intercepts OS signals to allow a 10-second window for finishing active bank transactions.

## 4. Deliberate Omissions
- Persistent Database: For portability, an inmemory map was used for the PaymentRepository instead of PostgreSQL to reduce reviewer friction.

- External Redis Container: Replaced with miniredis to provide a seamless "clone and run" experience.

- Deep Context Logging: Piggybacking correlation IDs deep into Use Cases was scoped out to prevent over-engineering while maintaining core functionality.


## 5. Design Discussions
**ACID vs ACI in Test Doubles**
While the in-memory map lacks Durability (data is lost on crash), we preserve Atomicity, Consistency, and Isolation (ACI). We use sync.RWMutex to prevent race conditions and ensure thread-safe isolation.

**Why RDBMS over NoSQL?**
In production, a relational database is preferred. Strict ACID compliance is mission-critical for financial transactions to ensure money is never double-spent or lost. The "eventual consistency" of NoSQL is unacceptable for core payment ledgers.

**Why we DO NOT cache GET requests**
Caching GET /v1/payments/{id} is dangerous. If a refund or dispute modifies a payment status, a stale cache could lead a merchant to fulfill a canceled order. GET requests must always query the live Source of Truth.

**Future Scaling: Package by Feature**
The current "Package by Layer" structure is ideal for this single-domain microservice. For expansion into Refunds or Payouts, the architecture would transition to Bounded Contexts (e.g., internal/payments/..., internal/refunds/...) to avoid horizontal layers becoming cluttered.

## How to Run and Test
The project is containerized for a friction-free setup.

**Step 1: Spin up the environment**
`docker-compose up -d`

**Step 2: Query the API**
The API is exposed on port 8090.

1. Process a Payment (POST)

`
curl -X POST http://localhost:8090/v1/payments \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: req-abc-123" \
  -d '{
    "card_number": "1234567890123456",
    "expiry_month": 12,
    "expiry_year": 2026,
    "currency": "USD",
    "amount": 1500,
    "cvv": "123"
  }'
`

2. Retrieve a Payment (GET)

`
curl -X GET http://localhost:8090/v1/payments/<PAYMENT_ID_FROM_POST_RESPONSE>
`
3. swagger doc

`
curl -X GET http://localhost:8090/swagger/index.html
`
**Step 3: Test Idempotency**
Run the exact same POST command from Step 1 again. The response will be returned instantly from the cache, bypassing the Bank Simulator.