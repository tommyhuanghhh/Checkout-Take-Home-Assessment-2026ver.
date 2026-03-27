# Checkout-Take-Home-Assessment-2026ver.

## Requirements Summary

**Goal:** Build an API-based Payment Gateway application that processes payments and retrieves payment details.

**Core Features:**
1. **Process a Payment:** Accept payment requests and validate inputs (card number, expiry date, currency, amount, CVV). Forward the request to an acquiring bank (via the simulator).
2. **Retrieve Payment Details:** Allow clients to fetch details of a previously processed payment.
3. **Bank Simulator Integration:** The payment gateway must interact with a provided Bank Simulator to process the payments.

**Go-Specific Constraints & Notes:**
- Work off the skeleton Payment Gateway API provided.
- **DO NOT** modify the `imposters/` directory (which contains the bank simulator configuration).
- **DO NOT** modify the `.editorconfig` file to ensure consistent code formatting.
- You are free to change the project structure (which fits perfectly with our plan to use Clean Architecture) and use your preferred test libraries.
- The project is set up with Swaggo for API auto-documentation. Swagger UI will be available at `http://localhost:8090/swagger/index.html`.

**Assessment Criteria & Expectations:**
- The code must compile and run successfully.
- The solution must include automated tests.
- Code should demonstrate simplicity, maintainability, and good API design principles.

## Architectural Requirement and Design
**Project Structure**
- This project would follow Clean Architecture principles.
- There will be presentation layer, application layer, infrastructure layer, and domain layer
- We should 
- Presentation layer will contain the handlers, models for request and response, middleware, and 
- Application layer handles application logic. It will contain usecases, 
- Domain layer
- Infrastructure layer

**Dependency and Interfaces**
- We follow the Dependency Inversion principles, which means we will have interfaces for each layer, and we only depend on interfaces.
- We use Wire to automatically complete the Dependency Injection.

**Unit Test**
- We will have unit tests for each layer.
 
**Infrastructure Details**
- connection pools -->we will have 2 separate connection pools: one for redis, one for http client
- using context to implement timeout
- singleflight for cache stampede prevention
- Cache whenever you can, but dont forget the security issue(PCI DSS)
- Limit goroutine with worker pool

**Security--PCI DSS Compliant**
- We assume client must provide idempotency key with each request
- Validation: Hash the incoming request body. If the client sends an existing Idempotency-Key but with a different request body (e.g., changed the amount from $10 to $100), reject it immediately (HTTP 409 Conflict or 400 Bad Request). This prevents malicious reuse of keys.
- we use json tag to validate http request, and value objects creation to double check(in case requests coming in via gRPC or other non-http protocols)

**Security--Exactly Once**

**Caching and Storage**
- Client-generated UUID will be the key for Redis storage
- 

**Interface and Build-Time Check**
- Placement:
- Ensure Build-Time check

**Bank Simulator**
- run "docker-compose up" to spin up the simulator
- call http://localhost:8080/payments with POST
- Example:
    {
        "card_number": "2222405343248877",
        "expiry_date": "04/2025",
        "currency": "GBP",
        "amount": 100,
        "cvv": "123"
    }
## API Design
**Post Payment**
- /v1/payments
- Requested Fields in Request Body:
    - card_number
    - expiry_month
    - expiry_year
    - currency
    - amount(int) -->why dont use string?
    - cvv
- We assume client will provide Idempotency Key, mandatorily
- Requested Fields in Response Body:

**Get Payment Details**
- /v1/payments/{paymentId}
- Requested Fields in Request Body:
    - paymentId(not idempotency key)
- Requested Fields in Response Body:
    - ID(payment id)
    - Status(Authorized/ Declined)
    - last four card digits
    - expiry_month
    - expiry_yesr
    - currency
    - amount
- Assume GET payment when payment status is pending is impossible

**Development Order--Outward**
1. Domain Layer: Start in internal/domain/. This is pure Go. No libraries, no JSON tags, no HTTP.
2. Application Layer(Usecase): The "How", the entire application flow lies here
3. Infrastructure Layer: fulfilling the contracts (interfaces) that the Domain and Use Cases already defined.
4. Presentation Layer: The "delivery" (Dont forget to set timeout in context)
4.5 logging and config file
5. main.go& wire.go: Initialize the Logger and Config + Setup the Dependency Injection
6. CI/CD
## Storage Design
**Why RDBMS over NoSQL?**
- ACID is mission-critical for db transaction, and RDBMS could more easily implement ACID. 

**You do not need to integrate with a real storage engine or database. It is fine to use the test double repository provided in the sample code to represent this.**
- ACID, What's left?:
    - Atomicity (A) and Isolation (I) could be achieved with Mutex
    - Consistency (C) will move to Domain Layer
    - Durability (D) is gone

**Why GET payment should ignore the cache?**
- the GET /v1/payments/{id} API must read directly from the Database (Repository), never the Cache. Here is why I think so:
1. The Key Mismatch: When the client calls POST /v1/payments, they pass an Idempotency-Key in the HTTP header (e.g., req-abc). This is what the cache uses as its lookup key. But when they call GET /v1/payments/pay_123, they are passing the generated Payment ID. The cache literally has no idea what pay_123 is.
2. The TTL (Time-to-Live): Idempotency is a temporary shield to prevent double-charging during network hiccups. The keys usually expire after 24 hours. If a merchant calls GET to look up a payment from 3 days ago, the cache will be empty. The database is permanent.
3. The Source of Truth: If a background process refunds a payment, it updates the database. If I serve GET requests from a stale cache, I will be returning inaccurate financial data.

## CI/CD Design
**Partial CD plan and config injection**
- App Code: Expects primitives injected via constructors (like baseURL string).
- Main.go: Reads Environment Variables and builds a Config struct.
- Dockerfile: Only builds the Go binary. No config mapping.
- Docker-Compose: Injects the actual URL strings as environment variables when spinning up the containers

## Future Scaling & Architecture
- For the scope of this assessment, the application is packaged by architectural layer (Clean Architecture) with a single, flat Domain package. If this service were to grow to encompass other domains (e.g., Refunds, Disputes), I would transition the structure to 'Package by Feature / Bounded Context' to prevent the Domain and Use Case packages from becoming bloated.