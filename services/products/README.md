# Products Service

A Go-based microservice for managing the product catalog in the Inventory Management System.

## Overview

The Products Service is responsible for:
- Managing product information (name, description, price, category)
- Providing product data to other services via gRPC
- Exposing REST API endpoints for frontend access

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                      Products Service                            │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────────┐    ┌──────────────────┐                   │
│  │   HTTP Handler   │    │   gRPC Handler   │                   │
│  │    (port 8001)   │    │    (port 9001)   │                   │
│  └────────┬─────────┘    └────────┬─────────┘                   │
│           │                       │                              │
│           └───────────┬───────────┘                              │
│                       ▼                                          │
│           ┌──────────────────────┐                               │
│           │     Controller       │                               │
│           │  (Business Logic)    │                               │
│           └──────────┬───────────┘                               │
│                      ▼                                           │
│           ┌──────────────────────┐                               │
│           │     Repository       │                               │
│           │   (Data Access)      │                               │
│           └──────────┬───────────┘                               │
│                      ▼                                           │
│           ┌──────────────────────┐                               │
│           │     PostgreSQL       │                               │
│           │     (Database)       │                               │
│           └──────────────────────┘                               │
└─────────────────────────────────────────────────────────────────┘
```

## Ports

| Protocol | Port | Description |
|----------|------|-------------|
| HTTP | 8001 | REST API for frontend communication |
| gRPC | 9001 | Inter-service communication |

## API Endpoints

### HTTP REST API

#### Health Check
```
GET /health
Response: "Products service is healthy"
```

#### Get All Products
```
GET /products
Response: Array of product objects
```

**Example Response:**
```json
[
  {
    "id": 1,
    "name": "Laptop",
    "description": "High-performance laptop",
    "price": 999.99,
    "category": "Electronics",
    "created_at": "2024-01-15T10:30:00Z"
  }
]
```

#### Get Product by ID
```
GET /products/{productId}
Response: Product object
```

#### Create Product
```
POST /products
Content-Type: application/json
Body: {
  "name": "Product Name",
  "description": "Product Description",
  "price": 99.99,
  "category": "Category"
}
Response: Created product object
```

### gRPC API

The service implements the `ProductService` defined in `proto/products/products.proto`:

| Method | Request | Response | Description |
|--------|---------|----------|-------------|
| `GetProduct` | `GetProductRequest` | `GetProductResponse` | Get a single product by ID |
| `ListProducts` | `ListProductsRequest` | `ListProductsResponse` | Get all products |
| `CreateProduct` | `CreateProductRequest` | `CreateProductResponse` | Create a new product |

## Project Structure

```
services/products/
├── cmd/
│   └── main.go              # Application entry point
├── internal/
│   ├── controller/          # Business logic layer
│   ├── handler/             # HTTP and gRPC handlers
│   ├── repository/          # Data access layer
│   └── error.go             # Custom error definitions
├── pkg/                     # Shared packages
├── proto/                   # Generated protobuf files
├── Dockerfile               # Container configuration
├── go.mod                   # Go module definition
├── go.sum                   # Go module checksums
└── README.md                # This file
```

## Configuration

The service is configured via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | 8001 | HTTP server port |
| `GRPC_PORT` | 9001 | gRPC server port |
| `DB_HOST` | (required) | PostgreSQL host |
| `DB_PORT` | (required) | PostgreSQL port |
| `DB_NAME` | inventory_db | Database name |
| `DB_USER` | (required) | Database username |
| `DB_PASSWORD` | (required) | Database password |

## Running Locally

### Prerequisites
- Go 1.25+
- PostgreSQL database
- Protocol Buffers compiler (protoc)

### Development

1. Set up environment variables:
```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=your_user
export DB_PASSWORD=your_password
export DB_NAME=inventory_db
```

2. Run the service:
```bash
cd services/products/cmd
go run main.go
```

### Building

```bash
cd services/products
go build -o products ./cmd
```

### Docker Build

```bash
docker build -t products-service:latest ./services/products
```

## Database Schema

The service uses the `products` table:

```sql
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL,
    category VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## Health Checks

The service exposes a health check endpoint at `/health` used by Kubernetes probes:
- **Liveness Probe**: Checks if the service is running
- **Readiness Probe**: Checks if the service is ready to accept traffic

## Dependencies

- `github.com/gorilla/mux` - HTTP router
- `github.com/lib/pq` - PostgreSQL driver
- `github.com/joho/godotenv` - Environment variable loading
- `google.golang.org/grpc` - gRPC framework

## Integration with Other Services

```
┌─────────────┐       gRPC        ┌──────────────────┐
│   Orders    │◄──────────────────│ Products Service │
│   Service   │   GetProduct()    │                  │
└─────────────┘                   └──────────────────┘
       │
       │                          ┌──────────────────┐
       │        HTTP              │     Frontend     │
       │◄─────────────────────────│  (via nginx)     │
       │                          └──────────────────┘
```

The Orders Service calls the Products Service via gRPC to:
- Validate that products exist when creating orders
- Retrieve product prices for order total calculation
