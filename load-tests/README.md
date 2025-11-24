# Load Testing with K6

This directory contains load testing scripts for the inventory management system.

## Prerequisites

- K6 installed locally: `brew install k6` (macOS) or download from https://k6.io/
- Kubernetes cluster running with the application deployed
- Port forwarding set up for services

## Setup Port Forwarding

Before running tests, forward the service ports to localhost:

```bash
# Forward Products Service
kubectl port-forward -n inventory-system svc/products-service 8001:8001 &

# Forward Inventory Service
kubectl port-forward -n inventory-system svc/inventory-service 8002:8002 &

# Forward Orders Service
kubectl port-forward -n inventory-system svc/orders-service 8003:8003 &
```

## Running Tests

### Individual Service Tests

```bash
# Test Products Service
k6 run scripts/products-test.js

# Test Inventory Service
k6 run scripts/inventory-test.js

# Test Orders Service
k6 run scripts/orders-test.js
```

### Full Integration Test

```bash
k6 run scripts/full-scenario.js
```

### Test Scenarios

```bash
# Smoke test (minimal load)
k6 run --vus 1 --duration 1m scripts/full-scenario.js

# Load test (normal load)
k6 run scenarios/load-test.js

# Stress test (find breaking point, trigger HPA)
k6 run scenarios/stress-test.js

# Spike test (sudden traffic surge)
k6 run --stage 0s:10,1m:100,2m:10,3m:0 scripts/full-scenario.js
```

### Custom Configuration

```bash
# Override base URLs
k6 run -e BASE_URL=http://localhost:8001 scripts/products-test.js

# Set specific VUs and duration
k6 run --vus 50 --duration 5m scripts/full-scenario.js

# Output results to InfluxDB for Grafana visualization
k6 run --out influxdb=http://localhost:8086/k6 scripts/full-scenario.js
```

## Using Docker Compose (with Grafana Dashboard)

```bash
# Start K6 with InfluxDB and Grafana
docker-compose up -d influxdb grafana

# Run specific test
docker-compose run k6 run /scripts/full-scenario.js

# View results in Grafana
open http://localhost:3001
```

## Monitoring HPA During Tests

Watch HPA scaling in real-time:

```bash
# In a separate terminal
watch -n 2 'kubectl get hpa -n inventory-system'

# Or monitor pods
watch -n 2 'kubectl get pods -n inventory-system'
```

## Expected Results

### Smoke Test
- Validates basic functionality
- Should pass with <1% error rate

### Load Test
- Tests normal operational capacity
- Should maintain <500ms p95 latency
- Should not trigger HPA

### Stress Test
- Should trigger HPA around 50-70 VUs
- Pods should scale from 1 to 2-5 replicas
- System should handle load without failures

## Interpreting Results

K6 provides metrics like:
- `http_req_duration`: Response time (p50, p95, p99)
- `http_req_failed`: Failed request rate
- `iterations`: Number of completed scenarios
- `vus`: Virtual users
- `http_reqs`: Requests per second

Good performance indicators:
- p95 latency < 1s for orders, < 500ms for products/inventory
- Error rate < 10%
- Successful HPA scaling under load
