# Orders Service

A Go-based microservice for managing orders in the Inventory Management System.

## Overview

The Orders Service is responsible for:
- Creating new orders with multiple items
- Managing order status (pending, fulfilled)
- Coordinating with Products and Inventory services via gRPC
- Calculating order totals based on product prices

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                       Orders Service                             │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────────┐    ┌──────────────────┐                   │
│  │   HTTP Handler   │    │   gRPC Handler   │                   │
│  │    (port 8003)   │    │    (port 9003)   │                   │
│  └────────┬─────────┘    └────────┬─────────┘                   │
│           │                       │                              │
│           └───────────┬───────────┘                              │
│                       ▼                                          │
│           ┌──────────────────────┐                               │
│           │     Controller       │                               │
│           │  (Business Logic)    │                               │
│           └──────────┬───────────┘                               │
│                      │                                           │
│         ┌────────────┼────────────┐                              │
│         ▼            ▼            ▼                              │
│   ┌───────────┐ ┌─────────┐ ┌───────────────┐                   │
│   │ Products  │ │ Repo    │ │  Inventory    │                   │
│   │ gRPC      │ │ (Local) │ │  gRPC Client  │                   │
│   │ Client    │ │         │ │               │                   │
│   └─────┬─────┘ └────┬────┘ └───────┬───────┘                   │
│         │            │              │                            │
│         ▼            ▼              ▼                            │
│   ┌───────────┐ ┌─────────┐  ┌───────────────┐                  │
│   │ Products  │ │PostgreSQL│  │  Inventory   │                  │
│   │ Service   │ │ (Orders) │  │   Service    │                  │
│   └───────────┘ └─────────┘  └───────────────┘                  │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## Ports

| Protocol | Port | Description |
|----------|------|-------------|
| HTTP | 8003 | REST API for frontend communication |
| gRPC | 9003 | Inter-service communication |

## Order Workflow

```
┌─────────────────────────────────────────────────────────────────┐
│                      Order Creation Flow                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  1. Client Request                                               │
│         │                                                        │
│         ▼                                                        │
│  2. Validate Products ──────────► Products Service (gRPC)        │
│         │                         - Get product details          │
│         │                         - Get prices                   │
│         ▼                                                        │
│  3. Reserve Stock ──────────────► Inventory Service (gRPC)       │
│         │                         - Reserve items                │
│         │                         - Check availability           │
│         ▼                                                        │
│  4. Create Order ───────────────► PostgreSQL                     │
│         │                         - Save order                   │
│         │                         - Save order items             │
│         ▼                                                        │
│  5. Return Order (status: pending)                               │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                     Order Fulfillment Flow                       │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  1. Fulfill Request                                              │
│         │                                                        │
│         ▼                                                        │
│  2. Get Order ──────────────────► PostgreSQL                     │
│         │                         - Verify status: pending       │
│         ▼                                                        │
│  3. Fulfill Reservation ────────► Inventory Service (gRPC)       │
│         │                         - Deduct reserved stock        │
│         ▼                                                        │
│  4. Update Order ───────────────► PostgreSQL                     │
│         │                         - Set status: fulfilled        │
│         ▼                                                        │
│  5. Return Order (status: fulfilled)                             │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## API Endpoints

### HTTP REST API

#### Health Check
```
GET /health
Response: "Orders service is healthy"
```

#### Get All Orders
```
GET /orders
Response: Array of order objects
```

**Example Response:**
```json
[
  {
    "id": 1,
    "customer_id": 123,
    "status": "pending",
    "total_amount": 1029.98,
    "created_at": "2024-01-15T10:30:00Z",
    "items": [
      {
        "product_id": 1,
        "quantity": 1,
        "price_at_order": 999.99
      },
      {
        "product_id": 2,
        "quantity": 1,
        "price_at_order": 29.99
      }
    ]
  }
]
```

#### Get Order by ID
```
GET /orders/{orderId}
Response: Order object
```

#### Create Order
```
POST /orders
Content-Type: application/json
Body: {
  "customer_id": 123,
  "items": [
    {"product_id": 1, "quantity": 2},
    {"product_id": 3, "quantity": 1}
  ]
}
Response: Created order object
```

#### Fulfill Order
```
POST /orders/{orderId}/fulfill
Response: Updated order object with status "fulfilled"
```

### gRPC API

The service implements the `OrderService` defined in `proto/orders/orders.proto`:

| Method | Request | Response | Description |
|--------|---------|----------|-------------|
| `GetOrder` | `GetOrderRequest` | `GetOrderResponse` | Get a single order by ID |
| `ListOrders` | `ListOrdersRequest` | `ListOrdersResponse` | Get all orders |
| `CreateOrder` | `CreateOrderRequest` | `CreateOrderResponse` | Create a new order |
| `FulfillOrder` | `FulfillOrderRequest` | `FulfillOrderResponse` | Fulfill an order |

## Project Structure

```
services/orders/
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
| `PORT` | 8003 | HTTP server port |
| `GRPC_PORT` | 9003 | gRPC server port |
| `DB_HOST` | (required) | PostgreSQL host |
| `DB_PORT` | (required) | PostgreSQL port |
| `DB_NAME` | inventory_db | Database name |
| `DB_USER` | (required) | Database username |
| `DB_PASSWORD` | (required) | Database password |
| `PRODUCTS_GRPC_ADDR` | products-service:9001 | Products service gRPC address |
| `INVENTORY_GRPC_ADDR` | inventory-service:9002 | Inventory service gRPC address |

## Running Locally

### Prerequisites
- Go 1.25+
- PostgreSQL database
- Products Service running (for gRPC calls)
- Inventory Service running (for gRPC calls)
- Protocol Buffers compiler (protoc)

### Development

1. Set up environment variables:
```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=your_user
export DB_PASSWORD=your_password
export DB_NAME=inventory_db
export PRODUCTS_GRPC_ADDR=localhost:9001
export INVENTORY_GRPC_ADDR=localhost:9002
```

2. Run the service:
```bash
cd services/orders/cmd
go run main.go
```

### Building

```bash
cd services/orders
go build -o orders ./cmd
```

### Docker Build

```bash
docker build -t orders-service:latest ./services/orders
```

## Database Schema

The service uses two tables:

### Orders Table
```sql
CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    customer_id INTEGER NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    total_amount DECIMAL(10, 2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Order Items Table
```sql
CREATE TABLE order_items (
    id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL REFERENCES products(id),
    quantity INTEGER NOT NULL,
    price_at_order DECIMAL(10, 2) NOT NULL
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

The Orders Service acts as an orchestrator, coordinating between Products and Inventory services:

```
                                    gRPC
┌──────────────┐  GetProduct()   ┌──────────────────┐
│   Products   │◄────────────────│                  │
│   Service    │                 │                  │
└──────────────┘                 │                  │
                                 │  Orders Service  │
┌──────────────┐  ReserveStock() │                  │
│  Inventory   │◄────────────────│                  │
│   Service    │  FulfillRes()   │                  │
└──────────────┘                 └──────────────────┘
                                        ▲
                                        │ HTTP
                                        │
                                 ┌──────────────────┐
                                 │     Frontend     │
                                 │   (via nginx)    │
                                 └──────────────────┘
```

### gRPC Client Calls

**To Products Service:**
- `GetProduct()` - Validate product exists and get price for order total calculation

**To Inventory Service:**
- `ReserveStock()` - Reserve inventory when creating an order
- `FulfillReservation()` - Deduct inventory when fulfilling an order
- `ReleaseReservation()` - Release inventory if order is cancelled

## Order Statuses

| Status | Description |
|--------|-------------|
| `pending` | Order created, stock reserved, awaiting fulfillment |
| `fulfilled` | Order completed, reserved stock deducted |
