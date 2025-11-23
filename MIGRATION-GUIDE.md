# Migration Guide: Consul to Kubernetes

This document describes the migration from Consul-based service discovery to Kubernetes-native service discovery with gRPC inter-service communication.

## What Changed

### Before (Consul + HTTP)
- Services used HTTP for all communication
- Consul provided service discovery
- Docker Compose for orchestration
- Services registered with Consul on startup
- Dynamic service discovery via Consul API

### After (Kubernetes + gRPC)
- Services use gRPC for inter-service communication
- HTTP endpoints still available for frontend
- Kubernetes provides service discovery via DNS
- Kind (Kubernetes in Docker) for local development
- Static service addresses via Kubernetes services

## Architecture Changes

### Service Communication

**Before:**
```
Orders Service → HTTP → Consul (discovery) → HTTP → Products Service
                                          → HTTP → Inventory Service
```

**After:**
```
Orders Service → gRPC → products-service:9001 (K8s DNS)
               → gRPC → inventory-service:9002 (K8s DNS)

Frontend → HTTP → products-service:8001
         → HTTP → inventory-service:8002
         → HTTP → orders-service:8003
```

### Port Allocation

| Service | HTTP Port | gRPC Port | Purpose |
|---------|-----------|-----------|---------|
| Products | 8001 | 9001 | Product catalog management |
| Inventory | 8002 | 9002 | Stock tracking and reservations |
| Orders | 8003 | 9003 | Order processing |
| Frontend | 3000 | - | Web interface |

### Service Discovery

**Before (Consul):**
```go
// Services queried Consul for endpoints
host, err := consulClient.DiscoverService("products")
// Returns: "products-container:8001"
```

**After (Kubernetes DNS):**
```go
// Services use static K8s service names
productsAddr := "products-service:9001"
// K8s DNS automatically resolves to correct pod
```

## File Structure Changes

### New Files

```
proto/
├── products/products.proto       # Products service gRPC definitions
├── inventory/inventory.proto     # Inventory service gRPC definitions
└── orders/orders.proto           # Orders service gRPC definitions

k8s/
├── namespace.yaml                # K8s namespace
├── products-deployment.yaml      # Products deployment and service
├── inventory-deployment.yaml     # Inventory deployment and service
├── orders-deployment.yaml        # Orders deployment and service
└── frontend-deployment.yaml      # Frontend deployment and service

kind-config.yaml                  # Kind cluster configuration
deploy-kind.sh                    # Deployment script
verify-setup.sh                   # Verification script
```

### Removed Files

```
services/*/pkg/consul/client.go   # Consul client code
docker-compose.yml                # Docker Compose configuration
```

### Modified Files

```
services/*/cmd/main.go            # Added gRPC server setup
services/*/Dockerfile             # Added gRPC port exposure
services/*/go.mod                 # Added gRPC dependencies, removed Consul
services/*/internal/handler/      # Added gRPC handlers
README.md                         # Updated documentation
```

## Code Changes

### Products Service

**Before:**
```go
// Only HTTP server
http.ListenAndServe(":8001", r)
```

**After:**
```go
// Both HTTP and gRPC servers
go func() {
    http.ListenAndServe(":8001", r)
}()

grpcServer := grpc.NewServer()
pb.RegisterProductServiceServer(grpcServer, grpcHandler)
grpcServer.Serve(lis)
```

### Orders Service

**Before:**
```go
// HTTP call to Products service
host, _ := consulClient.DiscoverService("products")
resp, err := http.Get(fmt.Sprintf("http://%s/products/%d", host, productID))
```

**After:**
```go
// gRPC call to Products service
productsConn := grpc.NewClient("products-service:9001")
productsClient := products_pb.NewProductServiceClient(productsConn)
productResp, err := productsClient.GetProduct(ctx, &products_pb.GetProductRequest{
    Id: productID,
})
```

## Deployment Changes

### Before (Docker Compose)

```bash
docker-compose up --build
```

### After (Kubernetes with Kind)

```bash
./deploy-kind.sh
```

## Configuration Changes

### Environment Variables

**Before:**
```yaml
CONSUL_HOST=consul
SERVICE_NAME=products
SERVICE_PORT=8001
```

**After:**
```yaml
PORT=8001
GRPC_PORT=9001
PRODUCTS_GRPC_ADDR=products-service:9001
INVENTORY_GRPC_ADDR=inventory-service:9002
```

## Benefits of Migration

1. **Performance**: gRPC is faster than HTTP/REST for inter-service communication
2. **Type Safety**: Protocol Buffers provide strong typing
3. **Cloud Native**: Kubernetes-ready for production deployment
4. **Scalability**: Easier to scale services independently with K8s
5. **Simplified**: No need to manage separate service discovery infrastructure
6. **Production Ready**: Can deploy to any Kubernetes cluster (GKE, EKS, AKS)

## Deployment Instructions

### Prerequisites

```bash
# Install Kind
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-amd64
chmod +x ./kind
sudo mv ./kind /usr/local/bin/kind

# Install kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x kubectl
sudo mv kubectl /usr/local/bin/kubectl
```

### Deploy

```bash
# Verify setup
./verify-setup.sh

# Deploy to Kind
./deploy-kind.sh
```

### Access

- **Frontend**: http://localhost:3000
- **Products API**: http://localhost:8001
- **Inventory API**: http://localhost:8002
- **Orders API**: http://localhost:8003

### Management

```bash
# Check pods
kubectl get pods -n inventory-system

# View logs
kubectl logs -n inventory-system <pod-name>

# Delete cluster
kind delete cluster --name inventory-cluster
```

## Troubleshooting

### Services not starting

```bash
# Check pod status
kubectl get pods -n inventory-system

# View logs
kubectl logs -n inventory-system <pod-name>

# Describe pod for events
kubectl describe pod -n inventory-system <pod-name>
```

### gRPC connection errors

```bash
# Check if services are running
kubectl get svc -n inventory-system

# Test service resolution
kubectl run -it --rm debug --image=busybox --restart=Never -- nslookup products-service.inventory-system
```

### Port conflicts

If ports are already in use, modify the port mappings in `kind-config.yaml`:

```yaml
extraPortMappings:
- containerPort: 30000
  hostPort: 3000  # Change this if port 3000 is in use
```

## Rollback

If you need to rollback to the Consul-based version:

```bash
git checkout <previous-commit>
docker-compose up --build
```

## Future Enhancements

Potential improvements for the future:

1. **TLS**: Add mTLS for secure gRPC communication
2. **Load Balancing**: Implement client-side load balancing for gRPC
3. **Observability**: Add Prometheus metrics and distributed tracing
4. **Service Mesh**: Consider Istio or Linkerd for advanced traffic management
5. **Horizontal Scaling**: Configure HorizontalPodAutoscaler for auto-scaling
6. **Persistent Storage**: Add databases for data persistence
7. **CI/CD**: Implement automated deployment pipelines
