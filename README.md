# Inventory Management System

A microservices-based inventory management application built with Go backend services, gRPC inter-service communication, and Kubernetes orchestration using Kind.

## Architecture

This application consists of four main components:

### Backend Microservices (Go)
1. **Products Service** (HTTP: 8001, gRPC: 9001) - Manages product catalog
2. **Inventory Service** (HTTP: 8002, gRPC: 9002) - Tracks stock levels and reservations
3. **Orders Service** (HTTP: 8003, gRPC: 9003) - Handles order processing

### Frontend
4. **Web Application** (Port 3000) - Simple HTML/CSS/JS interface

## Features

- **Product Management**: View and manage product catalog including product categories and descriptions
- **Inventory Tracking**: Real-time stock levels with reservation system to avoid overserving
- **Order Processing**: Create and fulfill orders with automatic inventory updates
- **gRPC Communication**: High-performance inter-service communication using gRPC
- **Service Discovery**: Kubernetes-native service discovery
- **Health Checking**: Automatic health monitoring of all services via Kubernetes probes
- **Microservices Architecture**: Independent, containerized services
- **Dual Protocol**: RESTful HTTP APIs for frontend and gRPC for inter-service communication
- **Web Interface**: User-friendly frontend to create and fulfill orders
- **Container Orchestration**: Kubernetes deployment using Kind for local development

## Quick Start

### Prerequisites
- Docker
- Kind (Kubernetes in Docker) - [Installation Guide](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
- kubectl - [Installation Guide](https://kubernetes.io/docs/tasks/tools/)
- Git

### Running the Application

1. Clone the repository:
```bash
git clone <repository-url>
cd up-distcomp-inventory
```

2. Deploy to Kind cluster (this will create the cluster, build images, and deploy):
```bash
./deploy-kind.sh
```

3. Access the application:
- **Web Interface**: http://localhost:3000
- **Products API**: http://localhost:8001
- **Inventory API**: http://localhost:8002
- **Orders API**: http://localhost:8003

### Managing the Deployment

Check pod status:
```bash
kubectl get pods -n inventory-system
```

View service logs:
```bash
kubectl logs -n inventory-system <pod-name>
```

Access a pod shell:
```bash
kubectl exec -it -n inventory-system <pod-name> -- sh
```

Delete the cluster:
```bash
kind delete cluster --name inventory-cluster
```

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

- **Service Discovery**: Kubernetes-native service discovery via DNS
- **Health Monitoring**: Kubernetes liveness and readiness probes
- **gRPC Communication**: Orders service uses gRPC to communicate with Products and Inventory services
- **HTTP APIs**: Frontend communicates with services via RESTful HTTP APIs
- Orders service communicates with Products service to validate products and get pricing (via gRPC)
- Orders service communicates with Inventory service to reserve and fulfill stock (via gRPC)
- All services expose health check endpoints at `/health`
- Services use Kubernetes ClusterIP services for internal communication

### gRPC Integration

Inter-service communication:
1. Each service exposes both HTTP and gRPC endpoints
2. HTTP endpoints (8001-8003) for frontend communication
3. gRPC endpoints (9001-9003) for inter-service communication
4. Orders service uses gRPC clients to call Products and Inventory services
5. Service discovery is handled by Kubernetes DNS (e.g., `products-service:9001`)

### Kubernetes Configuration

Services are configured via Kubernetes manifests with environment variables:

- `PORT`: HTTP port the service listens on (8001-8003)
- `GRPC_PORT`: gRPC port the service listens on (9001-9003)
- `PRODUCTS_GRPC_ADDR`: Products service gRPC address for Orders service
- `INVENTORY_GRPC_ADDR`: Inventory service gRPC address for Orders service
- `PRODUCTS_HOST`: Products service HTTP address for Orders service
- `INVENTORY_HOST`: Inventory service HTTP address for Orders service

## Sample Data

The application starts with sample data:

**Products:**
- Laptop ($999.99)
- Mouse ($29.99)
- Keyboard ($79.99)
- Monitor ($199.99)
- Desk Chair ($149.99)

**Inventory:**
- Each product has non-zero initial stock levels
- As there are no unfulfilled orders in the initial mock data, there is no reserved stock for any product either

## Technology Stack

- **Backend**: Go 1.21+ with Gorilla Mux router for HTTP and gRPC
- **Inter-service Communication**: gRPC with Protocol Buffers
- **Service Discovery**: Kubernetes DNS
- **Orchestration**: Kubernetes (Kind for local development)
- **Frontend**: HTML5, CSS3, Vanilla JavaScript
- **Containerization**: Docker
- **Networking**: Kubernetes ClusterIP and NodePort services
- **Data Storage**: In-memory (volatile)

## Architecture Benefits

- **Scalability**: Each service can be scaled independently via Kubernetes
- **Maintainability**: Clear separation of concerns
- **Reliability**: Service isolation prevents cascade failures
- **Performance**: gRPC provides high-performance inter-service communication
- **Service Discovery**: Kubernetes-native DNS-based service discovery
- **Cloud Native**: Designed for Kubernetes deployment
- **Flexibility**: Easy to modify or replace individual services
- **Production Ready**: Can be deployed to any Kubernetes cluster (GKE, EKS, AKS, etc.)
