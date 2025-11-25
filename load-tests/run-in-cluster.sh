#!/bin/bash

# Run K6 load tests inside Kubernetes cluster
# Wrapper script for managing in-cluster K6 tests
#
# Usage:
#   ./run-in-cluster.sh [command] [options]
#
# Commands:
#   run <test>     Run a specific test
#   list           List available tests
#   status         Show status of running tests
#   logs <test>    View logs for a test
#   stop <test>    Stop a running test
#   cleanup        Clean up completed jobs

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
NAMESPACE="inventory-system"

# Service URLs (internal Kubernetes DNS names)
PRODUCTS_SERVICE_URL="http://products-service:8001"
INVENTORY_SERVICE_URL="http://inventory-service:8002"
ORDERS_SERVICE_URL="http://orders-service:8003"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

show_help() {
    echo -e "${BLUE}K6 In-Cluster Test Runner${NC}"
    echo ""
    echo "Usage: $0 <command> [options]"
    echo ""
    echo "Commands:"
    echo "  run <test>     Run a specific test (smoke, products, inventory, orders, full)"
    echo "  run-high <test> Run high-load version of a test"
    echo "  list           List available tests"
    echo "  status         Show status of all K6 jobs"
    echo "  logs <test>    View logs for a test (or 'all' for all running tests)"
    echo "  stop <test>    Stop a running test"
    echo "  cleanup        Clean up all completed jobs"
    echo "  deploy-scripts Deploy/update test scripts ConfigMap"
    echo ""
    echo "Options:"
    echo "  -f, --follow   Follow logs in real-time (for 'logs' command)"
    echo "  -h, --help     Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 run smoke              # Run smoke test"
    echo "  $0 run products           # Run products service test"
    echo "  $0 run-high products      # Run high-load products test"
    echo "  $0 logs products -f       # Follow logs for products test"
    echo "  $0 status                 # Show all test statuses"
    echo "  $0 cleanup                # Clean up completed jobs"
}

list_tests() {
    echo -e "${BLUE}Available K6 Tests:${NC}"
    echo ""
    echo -e "${GREEN}Standard Tests (for validation):${NC}"
    echo "  smoke       - Quick smoke test (~1 min, 10 req/s)"
    echo "  products    - Products service test (~3 min, up to 50 req/s)"
    echo "  inventory   - Inventory service test (~3 min, up to 40 req/s)"
    echo "  orders      - Orders service test (~3 min, up to 20 req/s)"
    echo "  full        - Full scenario test (~7 min, multiple scenarios)"
    echo ""
    echo -e "${GREEN}High-Load Tests (for stress/HPA testing):${NC}"
    echo "  products-high    - Products service (~10 min, up to 100 req/s)"
    echo "  inventory-high   - Inventory service (~10 min, up to 80 req/s)"
    echo "  orders-high      - Orders service (~10 min, up to 50 req/s)"
    echo "  full-high        - Full scenario (~10 min, combined ~180 req/s)"
}

show_status() {
    echo -e "${BLUE}K6 Test Status:${NC}"
    echo ""
    
    # Show jobs
    echo -e "${YELLOW}Jobs:${NC}"
    kubectl get jobs -n "$NAMESPACE" -l app=k6-load-tests -o wide 2>/dev/null || echo "No K6 jobs found"
    echo ""
    
    # Show pods
    echo -e "${YELLOW}Pods:${NC}"
    kubectl get pods -n "$NAMESPACE" -l app=k6-load-tests -o wide 2>/dev/null || echo "No K6 pods found"
}

view_logs() {
    local test_name=$1
    local follow=$2
    local job_name=""
    
    case "$test_name" in
        smoke|smoke-test)
            job_name="k6-smoke-test"
            ;;
        products|products-service)
            job_name="k6-products-test"
            ;;
        inventory|inventory-service)
            job_name="k6-inventory-test"
            ;;
        orders|orders-service)
            job_name="k6-orders-test"
            ;;
        full|full-scenario)
            job_name="k6-full-scenario-test"
            ;;
        products-high)
            job_name="k6-products-high-load-test"
            ;;
        inventory-high)
            job_name="k6-inventory-high-load-test"
            ;;
        orders-high)
            job_name="k6-orders-high-load-test"
            ;;
        full-high)
            job_name="k6-full-high-load-test"
            ;;
        all)
            if [ "$follow" = "true" ]; then
                kubectl logs -f -l app=k6-load-tests -n "$NAMESPACE" --all-containers=true
            else
                kubectl logs -l app=k6-load-tests -n "$NAMESPACE" --all-containers=true
            fi
            return
            ;;
        *)
            echo -e "${RED}Unknown test: $test_name${NC}"
            exit 1
            ;;
    esac
    
    if [ "$follow" = "true" ]; then
        kubectl logs -f -l job-name="$job_name" -n "$NAMESPACE"
    else
        kubectl logs -l job-name="$job_name" -n "$NAMESPACE"
    fi
}

stop_test() {
    local test_name=$1
    local job_name=""
    
    case "$test_name" in
        smoke|smoke-test)
            job_name="k6-smoke-test"
            ;;
        products|products-service)
            job_name="k6-products-test"
            ;;
        inventory|inventory-service)
            job_name="k6-inventory-test"
            ;;
        orders|orders-service)
            job_name="k6-orders-test"
            ;;
        full|full-scenario)
            job_name="k6-full-scenario-test"
            ;;
        products-high)
            job_name="k6-products-high-load-test"
            ;;
        inventory-high)
            job_name="k6-inventory-high-load-test"
            ;;
        orders-high)
            job_name="k6-orders-high-load-test"
            ;;
        full-high)
            job_name="k6-full-high-load-test"
            ;;
        all)
            echo -e "${YELLOW}Stopping all K6 tests...${NC}"
            kubectl delete jobs -n "$NAMESPACE" -l app=k6-load-tests 2>/dev/null || true
            echo -e "${GREEN}✓ All tests stopped${NC}"
            return
            ;;
        *)
            echo -e "${RED}Unknown test: $test_name${NC}"
            exit 1
            ;;
    esac
    
    echo -e "${YELLOW}Stopping test: $test_name${NC}"
    kubectl delete job "$job_name" -n "$NAMESPACE" 2>/dev/null || echo "Job not found"
    echo -e "${GREEN}✓ Test stopped${NC}"
}

cleanup() {
    echo -e "${YELLOW}Cleaning up completed K6 jobs...${NC}"
    
    # Delete completed jobs
    kubectl delete jobs -n "$NAMESPACE" -l app=k6-load-tests --field-selector status.successful=1 2>/dev/null || true
    
    # Delete failed jobs
    kubectl delete jobs -n "$NAMESPACE" -l app=k6-load-tests --field-selector status.failed=1 2>/dev/null || true
    
    echo -e "${GREEN}✓ Cleanup complete${NC}"
}

deploy_scripts() {
    echo -e "${YELLOW}Deploying K6 scripts ConfigMaps...${NC}"
    
    # Deploy standard scripts
    if [ -f "$SCRIPT_DIR/k8s/k6-configmap.yaml" ]; then
        kubectl apply -f "$SCRIPT_DIR/k8s/k6-configmap.yaml"
        echo -e "${GREEN}✓ Standard scripts ConfigMap deployed${NC}"
    fi
    
    # Deploy high-load scripts if they exist
    if [ -d "$SCRIPT_DIR/k8s/high-load" ]; then
        # Create ConfigMap from high-load scripts directory
        kubectl create configmap k6-high-load-scripts \
            --from-file="$SCRIPT_DIR/k8s/high-load/" \
            -n "$NAMESPACE" \
            --dry-run=client -o yaml | kubectl apply -f -
        echo -e "${GREEN}✓ High-load scripts ConfigMap deployed${NC}"
    fi
}

run_test() {
    local test_name=$1
    
    # Deploy scripts first
    deploy_scripts
    
    # Use deploy-k6-tests.sh for running
    "$SCRIPT_DIR/deploy-k6-tests.sh" "$test_name" -w
}

run_high_load_test() {
    local test_name=$1
    local job_name=""
    local script_name=""
    
    case "$test_name" in
        products|products-service)
            job_name="k6-products-high-load-test"
            script_name="products-service-test.js"
            ;;
        inventory|inventory-service)
            job_name="k6-inventory-high-load-test"
            script_name="inventory-service-test.js"
            ;;
        orders|orders-service)
            job_name="k6-orders-high-load-test"
            script_name="orders-service-test.js"
            ;;
        full|full-scenario)
            job_name="k6-full-high-load-test"
            script_name="full-scenario-test.js"
            ;;
        *)
            echo -e "${RED}Unknown high-load test: $test_name${NC}"
            echo "Available: products, inventory, orders, full"
            exit 1
            ;;
    esac
    
    # Deploy scripts first
    deploy_scripts
    
    # Delete existing job if it exists
    kubectl delete job "$job_name" -n "$NAMESPACE" 2>/dev/null || true
    sleep 2
    
    echo -e "${YELLOW}Starting high-load K6 test: $test_name${NC}"
    
    # Create the job
    cat <<EOF | kubectl apply -f -
apiVersion: batch/v1
kind: Job
metadata:
  name: $job_name
  namespace: $NAMESPACE
  labels:
    app: k6-load-tests
    test: $test_name-high-load
spec:
  ttlSecondsAfterFinished: 600
  backoffLimit: 0
  template:
    metadata:
      labels:
        app: k6-load-tests
        test: $test_name-high-load
    spec:
      restartPolicy: Never
      containers:
        - name: k6
          image: grafana/k6:latest
          command:
            - k6
            - run
            - /scripts/$script_name
          env:
            - name: BASE_URL
              value: "$PRODUCTS_SERVICE_URL"
            - name: PRODUCTS_URL
              value: "$PRODUCTS_SERVICE_URL"
            - name: INVENTORY_URL
              value: "$INVENTORY_SERVICE_URL"
            - name: ORDERS_URL
              value: "$ORDERS_SERVICE_URL"
          resources:
            requests:
              cpu: 300m
              memory: 384Mi
            limits:
              cpu: 1000m
              memory: 1Gi
          volumeMounts:
            - name: scripts
              mountPath: /scripts
              readOnly: true
      volumes:
        - name: scripts
          configMap:
            name: k6-high-load-scripts
EOF
    
    echo -e "${GREEN}✓ High-load job created: $job_name${NC}"
    
    # Wait for pod to start
    echo -e "${YELLOW}Waiting for pod to start...${NC}"
    kubectl wait --for=condition=Ready pod -l job-name="$job_name" -n "$NAMESPACE" --timeout=60s 2>/dev/null || true
    
    echo -e "${YELLOW}Streaming logs (Ctrl+C to stop watching):${NC}"
    echo ""
    kubectl logs -f -l job-name="$job_name" -n "$NAMESPACE"
}

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}Error: kubectl is not installed or not in PATH${NC}"
    exit 1
fi

# Parse arguments
FOLLOW_LOGS="false"
COMMAND=""
ARG=""

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -f|--follow)
            FOLLOW_LOGS="true"
            shift
            ;;
        run|run-high|list|status|logs|stop|cleanup|deploy-scripts)
            COMMAND="$1"
            shift
            if [[ $# -gt 0 && ! "$1" =~ ^- ]]; then
                ARG="$1"
                shift
            fi
            ;;
        *)
            ARG="$1"
            shift
            ;;
    esac
done

# Execute command
case "$COMMAND" in
    run)
        if [ -z "$ARG" ]; then
            echo -e "${RED}Error: Please specify a test to run${NC}"
            echo "Use: $0 list to see available tests"
            exit 1
        fi
        run_test "$ARG"
        ;;
    run-high)
        if [ -z "$ARG" ]; then
            echo -e "${RED}Error: Please specify a test to run${NC}"
            echo "Available: products, inventory, orders, full"
            exit 1
        fi
        run_high_load_test "$ARG"
        ;;
    list)
        list_tests
        ;;
    status)
        show_status
        ;;
    logs)
        if [ -z "$ARG" ]; then
            echo -e "${RED}Error: Please specify a test name or 'all'${NC}"
            exit 1
        fi
        view_logs "$ARG" "$FOLLOW_LOGS"
        ;;
    stop)
        if [ -z "$ARG" ]; then
            echo -e "${RED}Error: Please specify a test to stop or 'all'${NC}"
            exit 1
        fi
        stop_test "$ARG"
        ;;
    cleanup)
        cleanup
        ;;
    deploy-scripts)
        deploy_scripts
        ;;
    *)
        show_help
        exit 1
        ;;
esac
