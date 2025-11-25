# Kubernetes Configuration

This directory contains all Kubernetes manifests for deploying the Inventory Management System in a Kind cluster.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────────────────────────┐
│                                    Kind Cluster                                             │
│                              (inventory-cluster)                                            │
├─────────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                             │
│  ┌────────────────────────────────────────────────────────────────────────────────────────┐ │
│  │                             Namespace: inventory-system                                │ │
│  └────────────────────────────────────────────────────────────────────────────────────────┘ │
│                                                                                             │
│  ┌──────────────────────────────────────────────────────────────────────────────────────┐   │
│  │                              External Access Layer                                   │   │
│  │                                                                                      │   │
│  │   ┌─────────────────────────────┐    ┌──────────────────────────────────────────┐    │   │
│  │   │     Frontend Service        │    │         Grafana Service                  │    │   │
│  │   │     (LoadBalancer)          │    │         (LoadBalancer)                   │    │   │
│  │   │                             │    │                                          │    │   │
│  │   │  ┌───────────────────────┐  │    │  ┌─────────────────────────────────────┐ │    │   │
│  │   │  │  Port: 80 → :30000    │  │    │  │  Port: 3000 → :30300                │ │    │   │
│  │   │  │  Host: localhost:3000 │  │    │  │  Host: localhost:3001               │ │    │   │
│  │   │  └───────────────────────┘  │    │  │  (Load testing dashboards)          │ │    │   │
│  │   └─────────────────────────────┘    │  └─────────────────────────────────────┘ │    │   │
│  │                                      └──────────────────────────────────────────┘    │   │
│  └──────────────────────────────────────────────────────────────────────────────────────┘   │
│                                                                                             │
│  ┌──────────────────────────────────────────────────────────────────────────────────────┐   │
│  │                             Application Services Layer                               │   │
│  │                                                                                      │   │
│  │   ┌─────────────────────┐  ┌─────────────────────┐  ┌─────────────────────┐          │   │
│  │   │  Products Service   │  │  Inventory Service  │  │   Orders Service    │          │   │
│  │   │    (ClusterIP)      │  │    (ClusterIP)      │  │    (ClusterIP)      │          │   │
│  │   │                     │  │                     │  │                     │          │   │
│  │   │  HTTP: 8001         │  │  HTTP: 8002         │  │  HTTP: 8003         │          │   │
│  │   │  gRPC: 9001         │  │  gRPC: 9002         │  │  gRPC: 9003         │          │   │
│  │   │                     │  │                     │  │                     │          │   │
│  │   │  HPA: 1-5 replicas  │  │  HPA: 1-5 replicas  │  │  HPA: 1-5 replicas  │          │   │
│  │   │  CPU Target: 70%    │  │  CPU Target: 70%    │  │  CPU Target: 70%    │          │   │
│  │   └─────────────────────┘  └─────────────────────┘  └─────────────────────┘          │   │
│  │              │                        │                        │                     │   │
│  │              └────────────────────────┼────────────────────────┘                     │   │
│  │                                       │                                              │   │
│  │                                       ▼                                              │   │
│  │   ┌──────────────────────────────────────────────────────────────────────────────┐   │   │
│  │   │                              PostgreSQL                                      │   │   │
│  │   │                              (ClusterIP)                                     │   │   │
│  │   │                                                                              │   │   │
│  │   │   Service: postgres:5432                                                     │   │   │
│  │   │   Storage: PersistentVolumeClaim (1Gi)                                       │   │   │
│  │   │                                                                              │   │   │
│  │   └──────────────────────────────────────────────────────────────────────────────┘   │   │
│  └──────────────────────────────────────────────────────────────────────────────────────┘   │
│                                                                                             │
│  ┌──────────────────────────────────────────────────────────────────────────────────────┐   │
│  │                             Load Testing Infrastructure                              │   │
│  │                                  (Optional)                                          │   │
│  │                                                                                      │   │
│  │   ┌─────────────────────┐  ┌─────────────────────┐  ┌─────────────────────┐          │   │
│  │   │      InfluxDB       │  │       Grafana       │  │     K6 Jobs         │          │   │
│  │   │    (ClusterIP)      │  │    (LoadBalancer)   │  │    (Batch Jobs)     │          │   │
│  │   │    Port: 8086       │  │    Port: 3000       │  │                     │          │   │
│  │   └─────────────────────┘  └─────────────────────┘  └─────────────────────┘          │   │
│  │                                                                                      │   │
│  └──────────────────────────────────────────────────────────────────────────────────────┘   │
│                                                                                             │
└─────────────────────────────────────────────────────────────────────────────────────────────┘
```

## Service Communication

```
┌─────────────────────────────────────────────────────────────────────────────────────────────┐
│                              Service Communication Flow                                     │
├─────────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                             │
│                                                                                             │
│    External                    ┌─────────────────────────────────────────────────────┐      │
│    Request     ──────────────► │                    Frontend                         │      │
│    (:3000)                     │                 (NGINX Container)                   │      │
│                                └─────────────────────────┬───────────────────────────┘      │
│                                                          │                                  │
│                                        NGINX Reverse Proxy (HTTP)                           │
│                                ┌─────────────────────────┼───────────────────────────┐      │
│                                │                         │                           │      │
│                                ▼                         ▼                           ▼      │
│                   ┌──────────────────────┐  ┌──────────────────────┐  ┌──────────────────┐  │
│                   │  products-service    │  │  inventory-service   │  │  orders-service  │  │
│                   │       :8001          │  │       :8002          │  │       :8003      │  │
│                   └──────────────────────┘  └──────────────────────┘  └────────┬─────────┘  │
│                              │                         │                       │            │
│                              └─────────────────────────┼───────────────────────┘            │
│                                                        │                                    │
│                    ┌───────────────────────────────────┼────────────────────────────────┐   │
│                    │                                   │                                │   │
│                    │                   gRPC Inter-Service Communication                 │   │
│                    │                                   │                                │   │
│                    │      orders-service ──────────────┼──────────────────►             │   │
│                    │             │                     │                                │   │
│                    │             │ GetProduct(:9001)   │ ReserveStock(:9002)            │   │
│                    │             │                     │ FulfillReservation(:9002)      │   │
│                    │             ▼                     ▼                                │   │
│                    │   products-service:9001   inventory-service:9002                   │   │
│                    │                                                                    │   │
│                    └────────────────────────────────────────────────────────────────────┘   │
│                                                                                             │
│                                           │                                                 │
│                                           ▼                                                 │
│                              ┌──────────────────────────┐                                   │
│                              │        PostgreSQL        │                                   │
│                              │     postgres:5432        │                                   │
│                              │                          │                                   │
│                              │  • products table        │                                   │
│                              │  • inventory table       │                                   │
│                              │  • orders table          │                                   │
│                              │  • order_items table     │                                   │
│                              └──────────────────────────┘                                   │
│                                                                                             │
└─────────────────────────────────────────────────────────────────────────────────────────────┘
```

## Kubernetes DNS Service Discovery

All services use Kubernetes DNS for service discovery:

```
┌─────────────────────────────────────────────────────────────────────────────────────────────┐
│                               Kubernetes DNS Names                                          │
├─────────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                             │
│   Service Name              Full DNS Name                                  Type             │
│   ──────────────────────────────────────────────────────────────────────────────────────────│
│                                                                                             │
│   products-service          products-service.inventory-system.svc.cluster.local:8001        │
│   products-service          products-service.inventory-system.svc.cluster.local:9001 (gRPC) │
│                                                                                             │
│   inventory-service         inventory-service.inventory-system.svc.cluster.local:8002       │
│   inventory-service         inventory-service.inventory-system.svc.cluster.local:9002 (gRPC)│
│                                                                                             │
│   orders-service            orders-service.inventory-system.svc.cluster.local:8003          │
│   orders-service            orders-service.inventory-system.svc.cluster.local:9003 (gRPC)   │
│                                                                                             │
│   postgres                  postgres.inventory-system.svc.cluster.local:5432                │
│                                                                                             │
│   frontend                  frontend.inventory-system.svc.cluster.local:80                  │
│                                                                                             │
│   influxdb (optional)       influxdb.inventory-system.svc.cluster.local:8086                │
│   grafana (optional)        grafana.inventory-system.svc.cluster.local:3000                 │
│                                                                                             │
└─────────────────────────────────────────────────────────────────────────────────────────────┘
```

## Manifest Files

┌─────────────────────────────────────────────────────────────────────────────────────────────────┐
| File                                      | Description                                         |
|───────────────────────────────────────────|─────────────────────────────────────────────────────|
| `namespace.yaml`                          | Creates the `inventory-system` namespace            |
| `postgres-deployment.yaml`                | PostgreSQL deployment, service, and PVC             |
| `postgres-init-job.yaml`                  | Job to initialize database schema                   |
| `postgres-config-generated.yaml.template` | Template for database configuration                 |
| `products-deployment.yaml`                | Products service deployment and service             |
| `inventory-deployment.yaml`               | Inventory service deployment and service            |
| `orders-deployment.yaml`                  | Orders service deployment and service               |
| `frontend-deployment.yaml`                | Frontend deployment and LoadBalancer service        |
| `hpa.yaml`                                | Horizontal Pod Autoscalers for all backend services |
└─────────────────────────────────────────────────────────────────────────────────────────────────┘

### Configuration Scripts

┌──────────────────────────────────────────────────────────────────────────────────┐
| Script                               | Description                               |
|──────────────────────────────────────|───────────────────────────────────────────|
| `postgres-create-config-from-env.sh` | Generates ConfigMap/Secret from .env file |
| `postgres-create-init-configmap.sh`  | Creates ConfigMap from SQL schema         |
| `delete-generated-files.sh`          | Removes generated configuration files     |
└──────────────────────────────────────────────────────────────────────────────────┘

## Service Types

```
┌─────────────────────────────────────────────────────────────────────────────────────────────┐
│                                   Service Types                                             │
├─────────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                             │
│   LoadBalancer Services (External Access)                                                   │
│   ════════════════════════════════════════                                                  │
│                                                                                             │
│   ┌─────────────────────────────────────────────────────────────────────────────────────┐   │
│   │  Frontend          │  NodePort: 30000  │  Host: localhost:3000  │  External Access  │   │
│   │  Grafana           │  NodePort: 30300  │  Host: localhost:3001  │  Load Testing     │   │
│   └─────────────────────────────────────────────────────────────────────────────────────┘   │
│                                                                                             │
│   ClusterIP Services (Internal Only)                                                        │
│   ═══════════════════════════════════                                                       │
│                                                                                             │
│   ┌─────────────────────────────────────────────────────────────────────────────────────┐   │
│   │  products-service   │  HTTP: 8001, gRPC: 9001  │  Backend microservice              │   │
│   │  inventory-service  │  HTTP: 8002, gRPC: 9002  │  Backend microservice              │   │
│   │  orders-service     │  HTTP: 8003, gRPC: 9003  │  Backend microservice              │   │
│   │  postgres           │  Port: 5432              │  Database                          │   │
│   │  influxdb           │  Port: 8086              │  Metrics storage (load testing)    │   │
│   └─────────────────────────────────────────────────────────────────────────────────────┘   │
│                                                                                             │
└─────────────────────────────────────────────────────────────────────────────────────────────┘
```

## Horizontal Pod Autoscaling

```
┌─────────────────────────────────────────────────────────────────────────────────────────────┐
│                              Horizontal Pod Autoscaler Configuration                        │
├─────────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                             │
│   Service               Min Replicas    Max Replicas    CPU Target    Scale Up/Down         │
│   ──────────────────────────────────────────────────────────────────────────────────────    │
│   products-service            1              5             70%         30s / 60s            │
│   inventory-service           1              5             70%         30s / 60s            │
│   orders-service              1              5             70%         30s / 60s            │
│                                                                                             │
│   Scale Up Behavior:                                                                        │
│   ─────────────────────────────────────────────────────────────────────────                 │
│   • Can double pods every 30 seconds when CPU > 70%                                         │
│   • Stabilization window: 30 seconds                                                        │
│                                                                                             │
│   Scale Down Behavior:                                                                      │
│   ─────────────────────────────────────────────────────────────────────────                 │
│   • Reduces pods by 50% every 30 seconds when CPU < 70%                                     │
│   • Stabilization window: 60 seconds (prevents flapping)                                    │
│                                                                                             │
│                           HPA Scaling Visualization                                         │
│                                                                                             │
│        ┌────────────────────────────────────────────────────────────────┐                   │
│   Load │                         ╱╲    ╱╲                               │                   │
│        │                        ╱  ╲  ╱  ╲                              │                   │
│        │       ╱╲              ╱    ╲╱    ╲                             │                   │
│        │      ╱  ╲            ╱            ╲                            │                   │
│        │─────╱────╲──────────╱──────────────╲─────────────────────────  │                   │
│        └────────────────────────────────────────────────────────────────┘                   │
│                                         Time                                                │
│                                                                                             │
│        ┌────────────────────────────────────────────────────────────────┐                   │
│   Pods │               ┌──────────────────┐                             │                   │
│    5   │               │                  │                             │                   │
│    4   │            ┌──┘                  └──┐                          │                   │
│    3   │         ┌──┘                        └──┐                       │                   │
│    2   │      ┌──┘                              └──┐                    │                   │
│    1   │──────┘                                    └────────────────────│                   │
│        └────────────────────────────────────────────────────────────────┘                   │
│                                         Time                                                │
│                                                                                             │
└─────────────────────────────────────────────────────────────────────────────────────────────┘
```

## Health Checks

All backend services implement health checks for Kubernetes probes:

```yaml
livenessProbe:
  httpGet:
    path: /health
    port: <service-port>
  initialDelaySeconds: 10
  periodSeconds: 10

readinessProbe:
  httpGet:
    path: /health
    port: <service-port>
  initialDelaySeconds: 5
  periodSeconds: 5
```

PostgreSQL uses `pg_isready` command for health checks.

## Resource Requests and Limits

```
┌─────────────────────────────────────────────────────────────────────────────────────────────┐
│                                Resource Configuration                                       │
├─────────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                             │
│   Service               CPU Request    CPU Limit    Memory Request    Memory Limit          │
│   ──────────────────────────────────────────────────────────────────────────────────────    │
│   products-service        100m          500m          128Mi            512Mi                │
│   inventory-service       100m          500m          128Mi            512Mi                │
│   orders-service          100m          500m          128Mi            512Mi                │
│   grafana (optional)      100m          500m          256Mi            512Mi                │
│   influxdb (optional)     100m          500m          256Mi            1Gi                  │
│                                                                                             │
└─────────────────────────────────────────────────────────────────────────────────────────────┘
```

## Configuration Management

```
┌─────────────────────────────────────────────────────────────────────────────────────────────┐
│                              Configuration Management                                       │
├─────────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                             │
│   ConfigMaps                                                                                │
│   ──────────────────────────────────────────────────────────────────────────────────────    │
│                                                                                             │
│   ┌─────────────────────┬─────────────────────────────────────────────────────────────┐     │
│   │  postgres-config    │  Database connection settings (host, port, name, user)      │     │
│   │  postgres-init-sql  │  SQL schema for database initialization                     │     │
│   │  k6-scripts         │  Load testing scripts (optional)                            │     │
│   │  grafana-*          │  Grafana datasources and dashboards (optional)              │     │
│   └─────────────────────┴─────────────────────────────────────────────────────────────┘     │
│                                                                                             │
│   Secrets                                                                                   │
│   ──────────────────────────────────────────────────────────────────────────────────────    │
│                                                                                             │
│   ┌─────────────────────┬─────────────────────────────────────────────────────────────┐     │
│   │  postgres-secret    │  Database password (base64 encoded)                         │     │
│   └─────────────────────┴─────────────────────────────────────────────────────────────┘     │
│                                                                                             │
└─────────────────────────────────────────────────────────────────────────────────────────────┘
```

## Deployment Order

```
┌─────────────────────────────────────────────────────────────────────────────────────────────┐
│                              Deployment Sequence                                            │
├─────────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                             │
│   1. Namespace                                                                              │
│      └──► kubectl apply -f namespace.yaml                                                   │
│                                                                                             │
│   2. Configuration                                                                          │
│      └──► postgres-create-config-from-env.sh                                                │
│      └──► kubectl apply -f postgres-config-generated.yaml                                   │
│      └──► postgres-create-init-configmap.sh                                                 │
│                                                                                             │
│   3. Database                                                                               │
│      └──► kubectl apply -f postgres-deployment.yaml                                         │
│      └──► wait for postgres to be ready                                                     │
│      └──► kubectl apply -f postgres-init-job.yaml                                           │
│      └──► wait for init job to complete                                                     │
│                                                                                             │
│   4. Backend Services                                                                       │
│      └──► kubectl apply -f products-deployment.yaml                                         │
│      └──► kubectl apply -f inventory-deployment.yaml                                        │
│      └──► kubectl apply -f orders-deployment.yaml                                           │
│      └──► wait for deployments to be ready                                                  │
│                                                                                             │
│   5. Frontend                                                                               │
│      └──► kubectl apply -f frontend-deployment.yaml                                         │
│                                                                                             │
│   6. Autoscaling                                                                            │
│      └──► kubectl apply -f hpa.yaml                                                         │
│                                                                                             │
└─────────────────────────────────────────────────────────────────────────────────────────────┘
```

## Useful Commands

### View Resources
```bash
# Get all pods in namespace
kubectl get pods -n inventory-system

# Get all services
kubectl get svc -n inventory-system

# Get HPA status
kubectl get hpa -n inventory-system

# Watch pods in real-time
watch -n 2 'kubectl get pods -n inventory-system'
```

### Debugging
```bash
# View pod logs
kubectl logs -n inventory-system <pod-name>

# Follow logs in real-time
kubectl logs -f -n inventory-system -l app=<service-name>

# Access pod shell
kubectl exec -it -n inventory-system <pod-name> -- sh

# Describe pod for events
kubectl describe pod -n inventory-system <pod-name>
```

### Port Forwarding
```bash
# Forward products service
kubectl port-forward -n inventory-system svc/products-service 8001:8001

# Forward inventory service
kubectl port-forward -n inventory-system svc/inventory-service 8002:8002

# Forward orders service
kubectl port-forward -n inventory-system svc/orders-service 8003:8003
```

### Scaling
```bash
# Manual scaling
kubectl scale deployment products-service -n inventory-system --replicas=3

# View HPA events
kubectl describe hpa products-service-hpa -n inventory-system
```

## Kind Cluster Configuration

The cluster is configured via `kind-config.yaml`:

```yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: inventory-cluster
nodes:
- role: control-plane
  extraPortMappings:
  - containerPort: 30000    # Frontend
    hostPort: 3000
    protocol: TCP
  - containerPort: 30300    # Grafana (load testing)
    hostPort: 3001
    protocol: TCP
```

## Cleanup

```bash
# Delete all resources in namespace
kubectl delete namespace inventory-system

# Delete the Kind cluster
kind delete cluster --name inventory-cluster
```
