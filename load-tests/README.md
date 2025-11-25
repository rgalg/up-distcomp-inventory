# Load Testing with K6

## Testing Approaches

This project supports two testing approaches for K6 load testing:

| Approach | Best For | Max Throughput | Setup Complexity |
|----------|----------|----------------|------------------|
| **Port-Forward** | Quick smoke tests, debugging | ~15-20 req/s | Low (local K6) |
| **In-Cluster** | High-load testing, HPA testing | 100+ req/s | Medium (K8s jobs) |

### Port-Forward Limitations

`kubectl port-forward` is single-threaded and bottlenecks at low request rates. When you try to run high-load tests through port-forward, you'll see connection errors even though the services are healthy. This is a limitation of port-forward, not your services.

**Use port-forward for:**
- Quick smoke tests (5-10 req/s)
- Debugging specific endpoints
- Development iteration

**Use in-cluster for:**
- Load testing (50+ req/s)
- Stress testing / HPA testing
- Realistic performance measurements

## Quick Start

### Option 1: Port-Forward Testing (Quick & Simple)

```bash
cd load-tests

# Setup port forwarding
./setup-port-forwarding.sh

# Run smoke test (5 req/s)
k6 run scripts/smoke-test.js

# Or use the helper script
./run-test.sh scripts/smoke-test.js
```

### Option 2: In-Cluster Testing (Recommended for Load Tests)

```bash
cd load-tests

# Run a test as a Kubernetes Job
./run-in-cluster.sh run smoke

# Run high-load test (triggers HPA)
./run-in-cluster.sh run-high products

# Interactive testing with debug pod
./create-debug-pod.sh
```

## In-Cluster Load Testing (Recommended for High Load)

Running tests inside Kubernetes eliminates port-forward bottlenecks and allows realistic load testing.

### Quick Start

```bash
# Deploy test scripts and run smoke test
./deploy-k6-tests.sh smoke

# View available tests
./deploy-k6-tests.sh --list

# Run specific tests
./deploy-k6-tests.sh products
./deploy-k6-tests.sh inventory
./deploy-k6-tests.sh orders
./deploy-k6-tests.sh full
```

### Using run-in-cluster.sh

More comprehensive script for managing in-cluster tests:

```bash
# List available tests
./run-in-cluster.sh list

# Run standard tests
./run-in-cluster.sh run smoke
./run-in-cluster.sh run products

# Run high-load tests (for HPA testing)
./run-in-cluster.sh run-high products
./run-in-cluster.sh run-high inventory

# View test status
./run-in-cluster.sh status

# View logs
./run-in-cluster.sh logs products
./run-in-cluster.sh logs products -f  # Follow logs

# Stop/cleanup
./run-in-cluster.sh stop products
./run-in-cluster.sh cleanup
```

### Using the Debug Pod

For interactive testing and experimentation:

```bash
# Create and connect to debug pod
./create-debug-pod.sh

# Inside the pod:
k6 run /scripts/smoke-test.js
k6 run /scripts/products-service-test.js
k6 run --vus 10 --duration 2m /scripts/smoke-test.js

# Exit with 'exit'

# Delete debug pod when done
./create-debug-pod.sh --delete
```

### Deploying Tests as Jobs

Tests run as Kubernetes Jobs for automated execution:

```bash
# Deploy ConfigMap with scripts
kubectl apply -f k8s/k6-configmap.yaml

# Run specific job
kubectl apply -f k8s/k6-job.yaml

# Watch job progress
kubectl get jobs -n inventory-system -l app=k6-load-tests -w

# View logs
kubectl logs -f -l app=k6-load-tests -n inventory-system
```

## High-Load Test Configurations

High-load tests are designed to stress the system and trigger HPA scaling.

| Test | Target Rate | Duration | Expected Behavior |
|------|-------------|----------|-------------------|
| products-high | 100 req/s | ~10 min | Should trigger HPA at ~60 req/s |
| inventory-high | 80 req/s | ~10 min | Should trigger HPA at ~50 req/s |
| orders-high | 50 req/s | ~10 min | Slower due to gRPC calls |
| full-high | 180 req/s combined | ~10 min | Multi-scenario realistic load |

### Running High-Load Tests

```bash
# Via helper script
./run-in-cluster.sh run-high products
./run-in-cluster.sh run-high full

# Watch HPA scaling
watch -n 2 'kubectl get hpa,pods -n inventory-system'
```

### Resource Requirements

High-load tests require more resources:

| Test Type | CPU Request | Memory Request |
|-----------|-------------|----------------|
| Standard | 200m | 256Mi |
| High-Load | 300m | 384Mi |

## Test Scripts Available

### Port-Forward Tests (Low Rate Limits)

Located in `scripts/`:

| Script | Rate | Duration | Purpose |
|--------|------|----------|---------|
| smoke-test.js | 5 req/s | 30s | Quick verification |
| products-service-test.js | 5-15 req/s | 3.5m | Products API test |
| inventory-service-test.js | 3-10 req/s | 3.5m | Inventory API test |
| orders-service-test.js | 2-8 req/s | 3.5m | Orders API test |
| full-scenario-test.js | 5-30 req/s | 8m | Full user journey |

### In-Cluster Tests (Higher Rates)

Located in `k8s/k6-configmap.yaml`:

| Script | Rate | Purpose |
|--------|------|---------|
| smoke-test.js | 10 req/s | Quick in-cluster verification |
| products-service-test.js | 10-50 req/s | Products load test |
| inventory-service-test.js | 10-40 req/s | Inventory load test |
| orders-service-test.js | 5-20 req/s | Orders load test |
| full-scenario-test.js | 5-30 req/s | Full journey test |

### High-Load Tests

Located in `k8s/high-load/`:

| Script | Rate | Duration |
|--------|------|----------|
| products-service-test.js | 10-100 req/s | 10 min |
| inventory-service-test.js | 10-80 req/s | 10 min |
| orders-service-test.js | 5-50 req/s | 10 min |
| full-scenario-test.js | Combined 180 req/s | 10 min |

## Monitoring Tests

### Watching Test Progress

```bash
# Watch K6 job status
kubectl get jobs -n inventory-system -l app=k6-load-tests -w

# Watch pods
watch -n 2 'kubectl get pods -n inventory-system -l app=k6-load-tests'

# Watch HPA scaling
watch -n 2 'kubectl get hpa -n inventory-system'

# Watch all resources
watch -n 2 'kubectl get jobs,pods,hpa -n inventory-system'
```

### Viewing K6 Job Logs

```bash
# Stream logs from running test
kubectl logs -f -l job-name=k6-products-test -n inventory-system

# Get logs from completed test
kubectl logs -l job-name=k6-smoke-test -n inventory-system

# Get all K6 logs
kubectl logs -l app=k6-load-tests -n inventory-system --all-containers=true
```

### Checking HPA Scaling Behavior

```bash
# Watch HPA metrics
watch -n 5 'kubectl get hpa -n inventory-system -o wide'

# Describe HPA for details
kubectl describe hpa products-service-hpa -n inventory-system

# Watch pod count during test
watch -n 2 'kubectl get pods -n inventory-system | grep -E "(products|inventory|orders)"'
```

## Port-Forward Testing (Light Testing)

For quick smoke tests and debugging, use port-forward:

### Setup

```bash
# Setup port forwarding
./setup-port-forwarding.sh

# Or manually:
kubectl port-forward -n inventory-system svc/products-service 8001:8001 &
kubectl port-forward -n inventory-system svc/inventory-service 8002:8002 &
kubectl port-forward -n inventory-system svc/orders-service 8003:8003 &
```

### Run Tests

```bash
# Using helper script
./run-test.sh scripts/smoke-test.js

# Direct K6 commands
k6 run scripts/smoke-test.js
k6 run scripts/products-service-test.js
k6 run scripts/inventory-service-test.js
```

### With Docker + Grafana Visualization

```bash
# Start InfluxDB and Grafana
docker compose up -d influxdb grafana

# Wait for services
sleep 10

# Run test with InfluxDB output
docker compose run --rm k6 run --out influxdb=http://localhost:8086/k6 /scripts/products-service-test.js

# View results in Grafana at http://localhost:3001
```

## Troubleshooting

### Port Forwarding Issues

Port-forward has inherent limitations:

1. **Single-threaded**: Can only handle one connection at a time
2. **Bottlenecks at ~15-20 req/s**: Even healthy services fail at higher rates
3. **Connection timeouts**: Long-running connections may drop

**Solutions:**
- Use lower rate limits (5-15 req/s)
- Use in-cluster testing for higher loads
- Check port-forward processes: `ps aux | grep port-forward`
- View logs: `tail -f /tmp/pf-*.log`
- Restart: `pkill -f "port-forward.*inventory-system" && ./setup-port-forwarding.sh`

### In-Cluster Testing Issues

**Pod not starting:**
```bash
# Check pod status
kubectl get pods -n inventory-system -l app=k6-load-tests

# Check events
kubectl describe pod -n inventory-system -l app=k6-load-tests

# Check ConfigMap exists
kubectl get configmap k6-scripts -n inventory-system
```

**Test failing:**
```bash
# Check service connectivity from inside cluster
kubectl run curl-test --rm -it --image=curlimages/curl -- curl http://products-service:8001/health

# Check service endpoints
kubectl get endpoints -n inventory-system
```

### Debug Pod Usage

```bash
# Create debug pod
./create-debug-pod.sh

# Inside pod, test connectivity
wget -qO- http://products-service:8001/health
wget -qO- http://inventory-service:8002/health
wget -qO- http://orders-service:8003/health

# Run custom test
k6 run --vus 5 --duration 30s /scripts/smoke-test.js
```

### Test Failures

```bash
# Test service connectivity (port-forward)
curl -v http://localhost:8001/products
curl -v http://localhost:8002/inventory
curl -v http://localhost:8003/orders

# Check Kubernetes pods
kubectl get pods -n inventory-system
kubectl logs -n inventory-system -l app=products-service

# Run minimal test
k6 run --vus 1 --duration 10s scripts/smoke-test.js
```

## Cleanup

```bash
# Stop port forwards
pkill -f "port-forward.*inventory-system"

# Clean up K6 jobs
./run-in-cluster.sh cleanup

# Delete debug pod
./create-debug-pod.sh --delete

# Stop Docker services
docker compose down

# Remove volumes (deletes test data)
docker compose down -v
```

## Example Output

Successful test output:

```
     ✓ status is 200
     ✓ response has products
     ✓ product detail status is 200
     
     checks.........................: 100.00% ✓ 450  ✗ 0
     data_received..................: 123 kB  2.1 kB/s
     data_sent......................: 45 kB   750 B/s
     http_req_duration..............: avg=45.3ms  min=12ms  med=38ms  max=150ms  p(95)=89ms
     http_reqs......................: 150     2.5/s
     iterations.....................: 50      0.833333/s
```