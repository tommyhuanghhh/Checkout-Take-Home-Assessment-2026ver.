# Stage 1: Build the binary (Updated to Go 1.25 to match your go.mod)
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Download Go modules (caching step)
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the optimized binary (Updated path to ./cmd/api)
RUN CGO_ENABLED=0 GOOS=linux go build -o /payment-gateway ./cmd/api

# Stage 2: Create the minimal runtime image
FROM alpine:latest

WORKDIR /app

# FIX: Force Alpine to download and apply the newest security patches for OS packages (like zlib)
RUN apk --no-cache upgrade

# Copy only the compiled binary from the builder stage
COPY --from=builder /payment-gateway .

# Expose the API port
EXPOSE 8090

# Run the binary
CMD ["./payment-gateway"]