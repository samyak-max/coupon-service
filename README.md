# Coupon Service

A production-ready coupon management service for the medicine ordering platform. This service provides functionality for creating, validating, and managing coupons with various discount types and usage restrictions.

## Features

- Admin coupon creation with flexible configuration
- Coupon validation with multiple constraints
- Support for different discount types (percentage, fixed)
- Support for different usage types (one-time, multi-use, time-based)
- Concurrent usage handling
- Caching for improved performance
- OpenAPI documentation
- Containerized deployment

## Tech Stack

- Go 1.21
- Gin Web Framework
- GORM with PostgreSQL
- Swagger for API documentation
- Docker & Docker Compose

## Getting Started

### Prerequisites

- Docker and Docker Compose
- Go 1.21 (for local development)

### Running with Docker

1. Clone the repository:
```bash
git clone https://github.com/samyak-max/coupon-service.git
cd coupon-service
```

2. Start the services:
```bash
docker-compose up -d
```

The service will be available at http://localhost:8080

### API Documentation

Once the service is running, you can access the Swagger UI at:
http://localhost:8080/swagger/index.html

## API Endpoints

### 1. Create Coupon (Admin)

```http
POST /api/v1/coupons
```

Example request:
```json
{
  "code": "SAVE20",
  "expiry_date": "2025-12-31T23:59:59Z",
  "usage_type": "multi_use",
  "applicable_medicine_ids": ["med_123", "med_456"],
  "applicable_categories": ["painkiller"],
  "min_order_value": 500,
  "discount_type": "percentage",
  "discount_value": 20,
  "max_usage_per_user": 3
}
```

### 2. Get Applicable Coupons

```http
GET /api/v1/coupons/applicable
```

Example request:
```json
{
  "cart_items": [
    {
      "id": "med_123",
      "category": "painkiller",
      "price": 100,
      "quantity": 2
    }
  ],
  "order_total": 700,
  "timestamp": "2025-05-05T15:00:00Z"
}
```

### 3. Validate Coupon

```http
POST /api/v1/coupons/validate
```

Example request:
```json
{
  "coupon_code": "SAVE20",
  "cart_items": [
    {
      "id": "med_123",
      "category": "painkiller",
      "price": 100,
      "quantity": 2
    }
  ],
  "order_total": 700,
  "timestamp": "2025-05-05T15:00:00Z",
  "user_id": "user_123"
}
```

## Development

### Local Setup

1. Install dependencies:
```bash
go mod download
```

2. Set up the database:
```bash
docker-compose up db -d
```

3. Run the application:
```bash
go run cmd/server/main.go
```
