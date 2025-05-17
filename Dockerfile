FROM golang:1.23 AS app-builder

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

# Corrected stage name here
COPY --from=app-builder /app/coupon-service .

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["./coupon-service"]
