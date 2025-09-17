# Inventory Management System

A microservices-based inventory management application built with Go backend services, Consul service discovery, and a simple web frontend.

## Architecture

This application consists of five main components:

### Backend Microservices (Go)
1. **Products Service** (Port 8001) - Manages product catalog
2. **Inventory Service** (Port 8002) - Tracks stock levels and reservations
3. **Orders Service** (Port 8003) - Handles order processing

### Service Discovery
4. **Consul** (Port 8500) - Service discovery and health checking

### Frontend
5. **Web Application** (Port 3000) - Simple HTML/CSS/JS interface

## Features

- **Product Management**: View and manage product catalog
- **Inventory Tracking**: Real-time stock levels with reservation system
- **Order Processing**: Create and fulfill orders with automatic inventory updates
- **Service Discovery**: Consul-based service discovery for dynamic service location
- **Health Checking**: Automatic health monitoring of all services
- **Microservices Architecture**: Independent, containerized services
- **RESTful APIs**: Clean API interfaces between services
- **Web Interface**: User-friendly frontend for all operations

## Quick Start

### Prerequisites
- Docker and Docker Compose
- Git

### Running the Application

1. Clone the repository:
```bash
git clone <repository-url>
cd up-compdist-inventory
```

2. Start all services with Docker Compose:
```bash
docker-compose up --build
```

3. Access the application:
- **Web Interface**: http://localhost:3000
- **Products API**: http://localhost:8001
- **Inventory API**: http://localhost:8002
- **Orders API**: http://localhost:8003
- **Consul UI**: http://localhost:8500 (Service discovery dashboard)

## API Documentation

### Products Service (Port 8001)

#### Get All Products
```
GET /products
Response: Array of product objects
```

#### Get Product by ID
```
GET /products/{id}
Response: Product object
```

#### Create Product
```
POST /products
Body: {
  "name": "Product Name",
  "description": "Product Description",
  "price": 99.99,
  "category": "Category"
}
```

### Inventory Service (Port 8002)

#### Get All Inventory
```
GET /inventory
Response: Array of inventory items
```

#### Get Inventory by Product ID
```
GET /inventory/{productId}
Response: Inventory item object
```

#### Update Inventory Quantity
```
PUT /inventory/{productId}
Body: {"quantity": 100}
```

#### Reserve Inventory
```
POST /inventory/{productId}/reserve
Body: {"quantity": 5}
```

#### Fulfill Reservation
```
POST /inventory/{productId}/fulfill
Body: {"quantity": 5}
```

### Orders Service (Port 8003)

#### Get All Orders
```
GET /orders
Response: Array of order objects
```

#### Get Order by ID
```
GET /orders/{id}
Response: Order object
```

#### Create Order
```
POST /orders
Body: {
  "customer_id": 123,
  "items": [
    {"product_id": 1, "quantity": 2},
    {"product_id": 3, "quantity": 1}
  ]
}
```

#### Fulfill Order
```
POST /orders/{id}/fulfill
```

## Development

### Running Services Individually

Each service can be run independently for development:

```bash
# Products Service
cd services/products
go run main.go

# Inventory Service
cd services/inventory
go run main.go

# Orders Service
cd services/orders
go run main.go
```

### Building Services

```bash
# Build all services
cd services/products && go build -o products .
cd ../inventory && go build -o inventory .
cd ../orders && go build -o orders .
```

## Service Communication & Discovery

- **Service Discovery**: Consul-based service discovery automatically locates service instances
- **Health Monitoring**: Consul health checks ensure only healthy services are discovered
- **Fallback Support**: Services gracefully fallback to environment variables if Consul is unavailable
- Orders service communicates with Products service to validate products and get pricing
- Orders service communicates with Inventory service to reserve and fulfill stock
- All services expose health check endpoints at `/health`
- Services use Docker networking for internal communication

### Consul Integration

Each service:
1. Registers itself with Consul on startup
2. Uses Consul to discover other services dynamically  
3. Implements health checks for monitoring
4. Gracefully deregisters on shutdown

Service discovery flow:
1. Orders service needs to call Products/Inventory services
2. Queries Consul for healthy service instances
3. Uses discovered endpoint for HTTP calls
4. Falls back to environment variables if Consul unavailable

## Service Discovery Configuration

Services can be configured with environment variables:

- `CONSUL_HOST`: Consul server hostname (default: localhost)
- `SERVICE_NAME`: Name to register in Consul
- `SERVICE_PORT`: Port number for service registration
- `PORT`: Port the service listens on

## Sample Data

The application starts with sample data:

**Products:**
- Laptop ($999.99)
- Mouse ($29.99)
- Keyboard ($79.99)
- Monitor ($199.99)
- Desk Chair ($149.99)

**Inventory:**
- Each product has initial stock levels with some reserved quantities

## Technology Stack

- **Backend**: Go 1.21+ with Gorilla Mux router
- **Service Discovery**: HashiCorp Consul
- **Frontend**: HTML5, CSS3, Vanilla JavaScript
- **Containerization**: Docker & Docker Compose
- **Networking**: Docker bridge networking
- **Data Storage**: In-memory (for simplicity)

## Architecture Benefits

- **Scalability**: Each service can be scaled independently
- **Maintainability**: Clear separation of concerns
- **Reliability**: Service isolation prevents cascade failures
- **Service Discovery**: Dynamic service location with health monitoring
- **Flexibility**: Easy to modify or replace individual services