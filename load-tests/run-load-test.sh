#!/bin/bash

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo -e "${YELLOW}=== K6 Load Test with Stable Port Forwarding ===${NC}\n"

# Function to setup port forwarding with health check
setup_port_forward() {
    local service=$1
    local port=$2
    
    echo -e "${YELLOW}Setting up port-forward for $service on port $port...${NC}"
    
    # Kill any existing port-forward on this port
    lsof -ti:$port | xargs kill -9 2>/dev/null || true
    sleep 1
    
    # Start port-forward
    kubectl port-forward -n inventory-system svc/$service $port:$port > /dev/null 2>&1 &
    local PF_PID=$!
    
    # Wait and verify
    sleep 3
    
    if curl -sf http://localhost:$port/health > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} Port $port ready (PID: $PF_PID)"
        return 0
    else
        echo -e "${RED}✗${NC} Port $port failed to start"
        return 1
    fi
}

# Function to monitor and restart port-forward if it dies
monitor_port_forward() {
    local service=$1
    local port=$2
    
    while true; do
        if ! lsof -ti:$port > /dev/null 2>&1; then
            echo -e "${YELLOW}⚠ Port $port died, restarting...${NC}"
            setup_port_forward $service $port
        fi
        sleep 5
    done
}

# Cleanup function
cleanup() {
    echo -e "\n${YELLOW}Cleaning up port forwards...${NC}"
    pkill -P $$ 2>/dev/null || true
    lsof -ti:8001,8002,8003 | xargs kill -9 2>/dev/null || true
    exit 0
}

trap cleanup INT TERM EXIT

# Setup initial port forwards
echo -e "${YELLOW}Starting port forwards...${NC}\n"

setup_port_forward products-service 8001 || exit 1
setup_port_forward inventory-service 8002 || exit 1
setup_port_forward orders-service 8003 || exit 1

# Start monitors in background
monitor_port_forward products-service 8001 &
monitor_port_forward inventory-service 8002 &
monitor_port_forward orders-service 8003 &

echo -e "\n${GREEN}✓ All port forwards ready${NC}\n"

# Determine which test to run
TEST_SCRIPT=${1:-"scripts/products-service-test-batch.js"}

if [ ! -f "$TEST_SCRIPT" ]; then
    echo -e "${RED}Error: Test script '$TEST_SCRIPT' not found${NC}"
    exit 1
fi

echo -e "${YELLOW}Running load test: $TEST_SCRIPT${NC}\n"

# Run k6 with retries
MAX_RETRIES=3
RETRY_COUNT=0

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if k6 run "$TEST_SCRIPT"; then
        echo -e "\n${GREEN}✓ Load test completed successfully${NC}"
        break
    else
        RETRY_COUNT=$((RETRY_COUNT + 1))
        if [ $RETRY_COUNT -lt $MAX_RETRIES ]; then
            echo -e "\n${YELLOW}⚠ Test failed, restarting port forwards and retrying ($RETRY_COUNT/$MAX_RETRIES)...${NC}\n"
            
            # Restart all port forwards
            setup_port_forward products-service 8001
            setup_port_forward inventory-service 8002
            setup_port_forward orders-service 8003
            
            sleep 5
        else
            echo -e "\n${RED}✗ Load test failed after $MAX_RETRIES attempts${NC}"
            exit 1
        fi
    fi
done

# Keep port forwards alive for manual testing if desired
echo -e "\n${YELLOW}Port forwards are still active. Press Ctrl+C to stop.${NC}"
wait
