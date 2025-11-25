# Inventory Service

A Go-based microservice for managing stock levels and reservations in the Inventory Management System.

## Overview

The Inventory Service is responsible for:
- Tracking stock levels for each product
- Managing stock reservations for pending orders
- Fulfilling reservations when orders are completed
- Releasing reservations when orders are cancelled

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     Inventory Service                           │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────────────┐    ┌──────────────────┐                   │
│  │   HTTP Handler   │    │   gRPC Handler   │                   │
│  │    (port 8002)   │    │    (port 9002)   │                   │
│  └────────┬─────────┘    └────────┬─────────┘                   │
│           │                       │                             │
│           └───────────┬───────────┘                             │
│                       ▼                                         │
│           ┌──────────────────────┐                              │
│           │     Controller       │                              │
│           │  (Business Logic)    │                              │
│           └──────────┬───────────┘                              │
│                      ▼                                          │
│           ┌──────────────────────┐                              │
│           │     Repository       │                              │
│           │   (Data Access)      │                              │
│           └──────────┬───────────┘                              │
│                      ▼                                          │
│           ┌──────────────────────┐                              │
│           │     PostgreSQL       │                              │
│           │     (Database)       │                              │
│           └──────────────────────┘                              │
└─────────────────────────────────────────────────────────────────┘
```

## Ports

┌───────────────────────────────────────────────────────┐
| Protocol | Port | Description                         |
|──────────|──────|─────────────────────────────────────|
| HTTP     | 8002 | REST API for frontend communication |
| gRPC     | 9002 | Inter-service communication         |
└───────────────────────────────────────────────────────┘

## Stock Management

The Inventory Service implements a reservation system to prevent overselling:

```
┌─────────────────────────────────────────────────────────────────┐
│                      Stock Lifecycle                            │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│   Total Stock = Available + Reserved                            │
│                                                                 │
│   ┌─────────────┐    Reserve     ┌─────────────┐                │
│   │  Available  │ ─────────────► │  Reserved   │                │
│   │    Stock    │                │    Stock    │                │
│   └─────────────┘                └─────────────┘                │
│         ▲                              │                        │
│         │                              │                        │
│         │ Release                      │ Fulfill                │
│         │ Reservation                  │ Reservation            │
│         │                              ▼                        │
│         │                        ┌─────────────┐                │
│         └─────────────────────── │   Shipped   │                │
│                                  │   (Stock    │                │
│                                  │  Deducted)  │                │
│                                  └─────────────┘                │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## API Endpoints

### HTTP REST API

#### Health Check
```
GET /health
Response: "Inventory service is healthy"
```

#### Get All Inventory
```
GET /inventory
Response: Array of inventory items
```

**Example Response:**
```json
[
  {
    "product_id": 1,
    "stock": 50,
    "reserved": 5,
    "updated_at": "2024-01-15T10:30:00Z"
  }
]
```

#### Get Inventory by Product ID
```
GET /inventory/{productId}
Response: Inventory item object
```

#### Update Stock
```
PUT /inventory/{productId}
Content-Type: application/json
Body: {"quantity": 100}
Response: Updated inventory item
```

#### Reserve Stock
```
POST /inventory/{productId}/reserve
Content-Type: application/json
Body: {"quantity": 5}
Response: Updated inventory item
```

#### Fulfill Reservation
```
POST /inventory/{productId}/fulfill
Content-Type: application/json
Body: {"quantity": 5}
Response: Updated inventory item
```

#### Release Reservation
```
POST /inventory/{productId}/release_reservation
Content-Type: application/json
Body: {"quantity": 5}
Response: Updated inventory item
```

### gRPC API

The service implements the `InventoryService` defined in `proto/inventory/inventory.proto`:
┌─────────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
| Method               | Request                     | Response                     | Description                 |
|──────────────────────|─────────────────────────────|──────────────────────────────|─────────────────────────────|
| `GetInventory`       | `GetInventoryRequest`       | `GetInventoryResponse`       | Get inventory for a product |
| `ListInventory`      | `ListInventoryRequest`      | `ListInventoryResponse`      | Get all inventory items     |
| `UpdateStock`        | `UpdateStockRequest`        | `UpdateStockResponse`        | Update stock quantity       |
| `ReserveStock`       | `ReserveStockRequest`       | `ReserveStockResponse`       | Reserve stock for an order  |
| `FulfillReservation` | `FulfillReservationRequest` | `FulfillReservationResponse` | Fulfill a reservation       |
| `ReleaseReservation` | `ReleaseReservationRequest` | `ReleaseReservationResponse` | Release a reservation       |
└─────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘

## Project Structure

```
services/inventory/
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

┌──────────────────────────────────────────────────┐
| Variable      | Default      | Description       |
|───────────────|──────────────|───────────────────|
| `PORT`        | 8002         | HTTP server port  |
| `GRPC_PORT`   | 9002         | gRPC server port  |
| `DB_HOST`     | (required)   | PostgreSQL host   |
| `DB_PORT`     | (required)   | PostgreSQL port   |
| `DB_NAME`     | inventory_db | Database name     |
| `DB_USER`     | (required)   | Database username |
| `DB_PASSWORD` | (required)   | Database password |
└──────────────────────────────────────────────────┘

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
cd services/inventory/cmd
go run main.go
```

### Building

```bash
cd services/inventory
go build -o inventory ./cmd
```

### Docker Build

```bash
docker build -t inventory-service:latest ./services/inventory
```

## Database Schema

The service uses the `inventory` table:

```sql
CREATE TABLE inventory (
    product_id INTEGER PRIMARY KEY REFERENCES products(id) ON DELETE CASCADE,
    stock INTEGER NOT NULL DEFAULT 0,
    reserved INTEGER NOT NULL DEFAULT 0,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
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
┌─────────────┐       gRPC          ┌───────────────────┐
│   Orders    │◄──────────────────  │ Inventory Service │
│   Service   │ ReserveStock()      │                   │
│             │ FulfillReservation()│                   │
│             │ ReleaseReservation()│                   │
└─────────────┘                     └───────────────────┘
       │
       │                           ┌───────────────────┐
       │        HTTP               │     Frontend      │
       │◄──────────────────────────│  (via nginx)      │
       │                           └───────────────────┘
```

The Orders Service calls the Inventory Service via gRPC to:
- Reserve stock when creating a new order
- Fulfill reservations when an order is completed
- Release reservations if an order is cancelled

## Available Stock Calculation

The available stock for a product is calculated as:

```
Available Stock = Total Stock - Reserved Stock
```

This ensures that stock reserved for pending orders cannot be allocated to new orders, preventing overselling.
