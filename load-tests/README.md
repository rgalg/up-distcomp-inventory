# Load Testing with K6

## Quick Start

### 1. Setup Port Forwarding

```bash
cd <wherever_this_is>/up-distcomp-inventory

./load-tests/setup-port-forwarding.sh
```

### 2. Run Tests

#### Option A: Run K6 Locally (Recommended for Testing)

```bash
cd <wherever_this_is>/up-distcomp-inventory/load-tests

# Simple smoke test
k6 run --vus 1 --duration 30s scripts/products-service-test.js

# Full scenario
k6 run scripts/full-scenario.js

# Stress test
k6 run scenarios/stress-test.js
```

#### Option B: Run with Docker + Grafana Visualization

```bash
cd <wherever_this_is>/up-distcomp-inventory/load-tests

# Start InfluxDB and Grafana
docker compose up -d influxdb grafana

# Wait for services to start
sleep 10

# Run test with InfluxDB output
docker compose run --rm k6 run --out influxdb=http://localhost:8086/k6 /scripts/products-service-test.js

# Or run full scenario
docker compose run --rm k6 run --out influxdb=http://localhost:8086/k6 /scripts/full-scenario.js

# View results in Grafana
echo "Open Grafana at http://localhost:3001"
```

### 3. Monitor Kubernetes HPA

In a separate terminal:

```bash
# Watch HPA scaling
watch -n 2 'kubectl get hpa -n inventory-system'

# Watch pods
watch -n 2 'kubectl get pods -n inventory-system'
```

## Test Scripts Available

### Individual Service Tests
- `scripts/products-service-test.js` - Tests Products API
- `scripts/inventory-service-test.js` - Tests Inventory API  
- `scripts/orders-service-test.js` - Tests Orders API

### Integration Tests
- `scripts/full-scenario.js` - Complete user journey (browse → order → fulfill)

### Load Profiles
- Smoke test: `--vus 1 --duration 1m`
- Load test: `--vus 10 --duration 3m`
- Stress test: `scenarios/stress-test.js`
- Spike test: `--stage 0s:10,30s:100,1m:0`

## Troubleshooting

### Port Forwarding Issues

```bash
# Check if port forwards are running
ps aux | grep port-forward

# Check logs
tail -f /tmp/pf-*.log

# Restart port forwards
pkill -f "port-forward.*inventory-system"
# Then run setup command again
```

### Docker Issues

```bash
# Check container status
docker compose ps

# View k6 logs
docker compose logs k6

# View InfluxDB logs
docker compose logs influxdb

# Restart services
docker compose down
docker compose up -d influxdb grafana
```

### Test Failures

```bash
# Test service connectivity first
curl -v http://localhost:8001/products
curl -v http://localhost:8002/inventory
curl -v http://localhost:8003/orders

# Run minimal test
k6 run --vus 1 --duration 10s scripts/products-service-test.js

# Check Kubernetes pods
kubectl get pods -n inventory-system
kubectl logs -n inventory-system -l app=products-service
```

## Cleanup

```bash
# Stop port forwards
pkill -f "port-forward.*inventory-system"

# Stop Docker services
cd ~/dev/up-distcomp-inventory/load-tests
docker compose down

# Remove volumes (optional - deletes test data)
docker compose down -v
```

## Example Output

Successful test output should look like:

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