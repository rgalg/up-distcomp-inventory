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

### System Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────────────────────────┐
│                                     Kind Cluster                                             │
│                                 (inventory-cluster)                                          │
├─────────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                              │
│                              Namespace: inventory-system                                     │
│                                                                                              │
│   External Access                                                                            │
│   ═══════════════                                                                           │
│                                                                                              │
│   ┌─────────────────────────────────────────────────────────────────────────────────────┐   │
│   │                             Frontend (LoadBalancer)                                  │   │
│   │                               localhost:3000                                         │   │
│   │                                                                                      │   │
│   │   ┌─────────────────┐                                                               │   │
│   │   │ NGINX Container │──────────► Reverse Proxy to Backend Services                  │   │
│   │   └─────────────────┘                                                               │   │
│   └─────────────────────────────────────────────────────────────────────────────────────┘   │
│                                           │                                                  │
│                                           │ HTTP                                             │
│                                           ▼                                                  │
│   ┌─────────────────────────────────────────────────────────────────────────────────────┐   │
│   │                           Backend Services (ClusterIP)                               │   │
│   │                                                                                      │   │
│   │   ┌─────────────────┐   ┌──────────────────┐   ┌─────────────────┐                  │   │
│   │   │ Products Service│   │ Inventory Service│   │  Orders Service │                  │   │
│   │   │  HTTP: 8001     │   │   HTTP: 8002     │   │   HTTP: 8003    │                  │   │
│   │   │  gRPC: 9001     │   │   gRPC: 9002     │   │   gRPC: 9003    │                  │   │
│   │   │  HPA: 1-5 pods  │   │   HPA: 1-5 pods  │   │   HPA: 1-5 pods │                  │   │
│   │   └────────┬────────┘   └────────┬─────────┘   └────────┬────────┘                  │   │
│   │            │                     │                       │                           │   │
│   │            │                     │ gRPC                  │                           │   │
│   │            │◄────────────────────┼───────────────────────┘                           │   │
│   │            │                     │                                                   │   │
│   └────────────┼─────────────────────┼───────────────────────────────────────────────────┘   │
│                │                     │                                                       │
│                └──────────┬──────────┘                                                       │
│                           │                                                                  │
│                           ▼                                                                  │
│   ┌─────────────────────────────────────────────────────────────────────────────────────┐   │
│   │                              PostgreSQL (ClusterIP)                                  │   │
│   │                                   Port: 5432                                         │   │
│   │                           Persistent Volume: 1Gi                                     │   │
│   └─────────────────────────────────────────────────────────────────────────────────────┘   │
│                                                                                              │
└─────────────────────────────────────────────────────────────────────────────────────────────┘
```

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

2. Configure your own environment variables
```bash
cp .env.template .env
# you should modify the default values in the new .env file
```

3. Create configurations and deploy everything to a Kind cluster (this will execute a script to create a .yaml based on the .sql DB schema, create the cluster, build images, deploy and configure a PostgreSQL service, and deploy both the frontend and backend services into the Kind cluster):
```bash
./deploy-kind.sh
```

5. Access the application:
- **Web Interface**: http://localhost:3000 (via LoadBalancer)

**Note**: Backend services (Products, Inventory, Orders) use ClusterIP and are only accessible within the cluster for internal communication. Access them through the frontend web interface or use kubectl port-forward for debugging:
```bash
# Example: Forward Products API for debugging
kubectl port-forward -n inventory-system svc/products-service 8001:8001
```

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

#### Release Reservation
```
POST /inventory/{productId}/release_reservation
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
- **gRPC Communication**: Orders service uses gRPC to communicate with Products and Inventory services (ports 9001-9003)
- **HTTP APIs**: Frontend communicates with services via RESTful HTTP APIs (ports 8001-8003)
- Orders service communicates with Products service to validate products and get pricing (via gRPC)
- Orders service communicates with Inventory service to reserve and fulfill stock (via gRPC)
- All services expose health check endpoints at `/health`
- Backend services and database use Kubernetes ClusterIP services for internal communication
- Frontend service uses LoadBalancer type for external access

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

#### Service Types

The application follows Kubernetes best practices for service exposure:

- **Frontend Service**: LoadBalancer type - Exposed externally for user access
- **Backend Services** (Products, Inventory, Orders): ClusterIP type - Internal communication only
- **Database** (PostgreSQL): ClusterIP type - Internal communication only

#### Horizontal Pod Autoscaling

The backend services (Products, Inventory, Orders) are configured with Horizontal Pod Autoscalers (HPA):

- **Minimum Replicas**: 1
- **Maximum Replicas**: 5
- **Target CPU Utilization**: 70%
- **Scale Up Behavior**: Can double pods every 30 seconds when needed
- **Scale Down Behavior**: Gradually reduces pods by 50% every 30 seconds with 60s stabilization window

To view HPA status:
```bash
kubectl get hpa -n inventory-system
```

To test autoscaling, generate load on the services using tools like Locust or K6.

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

- **Backend**: Go 1.25+ with Gorilla Mux router for HTTP and gRPC
- **Inter-service Communication**: gRPC with Protocol Buffers
- **Service Discovery**: Kubernetes DNS
- **Orchestration**: Kubernetes (Kind for local development)
- **Frontend**: HTML5, CSS3, Vanilla JavaScript
- **Containerization**: Docker
- **Networking**: Kubernetes ClusterIP services (internal) and LoadBalancer (frontend)
- **Auto-scaling**: Horizontal Pod Autoscaler (HPA) for backend services
- **Data Storage**: PostgreSQL database with persistent storage

## Architecture Benefits

- **Scalability**: Each service can be scaled independently via Kubernetes
- **Maintainability**: Clear separation of concerns
- **Reliability**: Service isolation prevents cascade failures
- **Performance**: gRPC provides high-performance inter-service communication
- **Service Discovery**: Kubernetes-native DNS-based service discovery
- **Cloud Native**: Designed for Kubernetes deployment
- **Flexibility**: Easy to modify or replace individual services
- **Production Ready**: Can be deployed to any Kubernetes cluster (GKE, EKS, AKS, etc.)

## Load Testing

The project includes comprehensive load testing infrastructure using K6. Three testing approaches are supported:

| Approach | Best For | Max Throughput | Setup Complexity |
|----------|----------|----------------|------------------|
| **Port-Forward** | Quick smoke tests, debugging | ~15-20 req/s | Low |
| **In-Cluster** | High-load testing, HPA testing | 100+ req/s | Medium |
| **Grafana** | Load testing with real-time dashboards | 100+ req/s | Medium |

### Quick Load Test

```bash
cd load-tests

# Port-forward testing (quick smoke tests)
./run-test.sh scripts/smoke-test.js

# In-cluster testing (higher load)
./run-in-cluster.sh run smoke

# Grafana-based testing (with dashboards at localhost:3001)
./deploy-grafana-k6.sh run smoke
```

For detailed load testing documentation, see [load-tests/README.md](load-tests/README.md).

## Project Structure

```
up-distcomp-inventory/
├── services/                    # Backend microservices
│   ├── products/               # Products service (Go)
│   ├── inventory/              # Inventory service (Go)
│   └── orders/                 # Orders service (Go)
├── frontend/                    # Web application (HTML/CSS/JS)
├── k8s/                        # Kubernetes manifests
├── load-tests/                 # K6 load testing scripts
├── proto/                      # Protocol Buffer definitions
├── postgres-config/            # Database schema
├── deploy-kind.sh              # Deployment script
├── kind-config.yaml            # Kind cluster configuration
├── verify-setup.sh             # Setup verification script
└── README.md                   # This file
```

## Additional Documentation

- [Products Service](services/products/README.md) - Product catalog management
- [Inventory Service](services/inventory/README.md) - Stock and reservation management
- [Orders Service](services/orders/README.md) - Order processing
- [Frontend](frontend/README.md) - Web interface
- [Kubernetes Configuration](k8s/README.md) - Infrastructure and deployment
- [Load Testing](load-tests/README.md) - Performance testing
