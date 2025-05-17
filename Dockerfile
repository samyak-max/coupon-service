FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o coupon-service ./cmd/server

# Create a minimal production image
FROM alpine:latest

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/coupon-service .

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["./coupon-service"] 