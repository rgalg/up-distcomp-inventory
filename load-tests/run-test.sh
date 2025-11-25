#!/bin/bash

# K6 Load Test Runner (via Port-Forward)
#
# This script runs K6 load tests via kubectl port-forward.
# Suitable for quick smoke tests and low-rate testing only.
#
# IMPORTANT: Port-forward has throughput limitations (~15-20 req/s max).
# For high-load testing, use the in-cluster approach instead:
#   ./run-in-cluster.sh run <test>
#   ./create-debug-pod.sh
#
# See README.md for more information on testing approaches.

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=== K6 Load Test Runner (Port-Forward) ===${NC}\n"

echo -e "${BLUE}NOTE: This runs tests via port-forward which has rate limitations.${NC}"
echo -e "${BLUE}For high-load testing, use: ./run-in-cluster.sh run <test>${NC}\n"

# Check if k6 is installed
if ! command -v k6 &> /dev/null; then
    echo -e "${RED}ERROR: k6 is not installed${NC}"
    echo "Install with: brew install k6 (macOS) or visit https://k6.io/docs/get-started/installation/"
    exit 1
fi

# Setup port forwarding
echo -e "${YELLOW}Setting up port forwarding...${NC}"
pkill -f "port-forward.*inventory-system" 2>/dev/null || true

kubectl port-forward -n inventory-system svc/products-service 8001:8001 > /tmp/pf-products.log 2>&1 &
kubectl port-forward -n inventory-system svc/inventory-service 8002:8002 > /tmp/pf-inventory.log 2>&1 &
kubectl port-forward -n inventory-system svc/orders-service 8003:8003 > /tmp/pf-orders.log 2>&1 &

echo "Waiting for port forwards to establish..."
sleep 5

# Verify services
echo -e "\n${YELLOW}Verifying service health...${NC}"
ALL_HEALTHY=true

for port in 8001 8002 8003; do
    if curl -s -f http://localhost:$port/health > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} Port $port is healthy"
    else
        echo -e "${RED}✗${NC} Port $port is not responding"
        ALL_HEALTHY=false
    fi
done

if [ "$ALL_HEALTHY" = false ]; then
    echo -e "\n${RED}ERROR: Not all services are healthy. Check your Kubernetes deployment.${NC}"
    echo "Debug with: kubectl get pods -n inventory-system"
    exit 1
fi

# Run test
echo -e "\n${YELLOW}Running load test...${NC}\n"

if [ -z "$1" ]; then
    TEST_SCRIPT="scripts/products-service-test.js"
else
    TEST_SCRIPT="$1"
fi

echo "Test script: $TEST_SCRIPT"
echo ""

k6 run "$TEST_SCRIPT"

# Cleanup
echo -e "\n${YELLOW}Cleaning up port forwards...${NC}"
pkill -f "port-forward.*inventory-system" 2>/dev/null || true

echo -e "${GREEN}Done!${NC}"
