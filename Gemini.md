# Checkout Take Home Assessment 2026ver

This project is initialized and assisted by Gemini.

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
- connection pools
- using context to implement timeout
- singleflight for cache stampede prevention
- Cache whenever you can, but dont forget the security issue(PCI DSS)
- Limit goroutine with worker pool

**Security--PCI DSS Compliant**
- We assume client must provide idempotency key with each request
- Validation: Hash the incoming request body. If the client sends an existing Idempotency-Key but with a different request body (e.g., changed the amount from $10 to $100), reject it immediately (HTTP 409 Conflict or 400 Bad Request). This prevents malicious reuse of keys.
- we use json tag to validate http request, and value objects creation to double check(in case requests coming in via gRPC or other non-http protocols)

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


## Database Design
**Why RDBMS over NoSQL?**
- ACID is mission-critical for db transaction, and RDBMS could more easily implement ACID. 


