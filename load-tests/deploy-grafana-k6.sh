#!/bin/bash

# Deploy K6 load tests with Grafana visualization
# This script deploys InfluxDB and Grafana to Kubernetes for real-time 
# monitoring of K6 load test results through dashboards.
#
# Usage:
#   ./deploy-grafana-k6.sh                    # Deploy infrastructure only
#   ./deploy-grafana-k6.sh run [test-name]    # Deploy and run a test
#   ./deploy-grafana-k6.sh run smoke          # Run smoke test with Grafana output
#   ./deploy-grafana-k6.sh run products       # Run products service test
#   ./deploy-grafana-k6.sh run inventory      # Run inventory service test
#   ./deploy-grafana-k6.sh run orders         # Run orders service test
#   ./deploy-grafana-k6.sh run full           # Run full scenario test
#   ./deploy-grafana-k6.sh status             # Show status of all components
#   ./deploy-grafana-k6.sh logs [test]        # View test logs
#   ./deploy-grafana-k6.sh stop [test]        # Stop a running test
#   ./deploy-grafana-k6.sh cleanup            # Clean up test jobs
#   ./deploy-grafana-k6.sh delete             # Delete all Grafana infrastructure
#   ./deploy-grafana-k6.sh --list             # List available tests

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
NAMESPACE="inventory-system"
GRAFANA_PORT="3001"

# Service URLs (internal Kubernetes DNS names)
PRODUCTS_SERVICE_URL="http://products-service:8001"
INVENTORY_SERVICE_URL="http://inventory-service:8002"
ORDERS_SERVICE_URL="http://orders-service:8003"
INFLUXDB_URL="http://influxdb:8086/k6"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

show_help() {
    echo -e "${BLUE}K6 Load Testing with Grafana Visualization${NC}"
    echo ""
    echo "This tool deploys InfluxDB and Grafana for real-time monitoring"
    echo "of K6 load test results through interactive dashboards."
    echo ""
    echo "Usage: $0 [command] [options]"
    echo ""
    echo "Commands:"
    echo "  deploy          Deploy InfluxDB and Grafana infrastructure (default)"
    echo "  run <test>      Deploy infrastructure and run a specific test"
    echo "  status          Show status of all Grafana components"
    echo "  logs <test>     View logs for a specific test"
    echo "  stop <test>     Stop a running test (or 'all' for all tests)"
    echo "  cleanup         Clean up completed test jobs"
    echo "  delete          Delete all Grafana infrastructure"
    echo ""
    echo "Options:"
    echo "  -h, --help      Show this help message"
    echo "  -l, --list      List available tests"
    echo "  -w, --watch     Watch test progress after starting"
    echo "  -f, --follow    Follow logs in real-time"
    echo ""
    echo "Available tests:"
    echo "  smoke           Smoke test - quick verification (~1 min)"
    echo "  products        Products service load test (~3 min)"
    echo "  inventory       Inventory service load test (~3 min)"
    echo "  orders          Orders service load test (~3 min)"
    echo "  full            Full scenario test (~7 min)"
    echo ""
    echo "Examples:"
    echo "  $0                     # Deploy Grafana infrastructure"
    echo "  $0 run smoke           # Deploy and run smoke test"
    echo "  $0 run products -w     # Run products test and watch logs"
    echo "  $0 status              # Check status of all components"
    echo "  $0 logs smoke -f       # Follow logs for smoke test"
    echo "  $0 delete              # Remove all Grafana components"
    echo ""
    echo "After deployment, access Grafana at:"
    echo "  http://localhost:${GRAFANA_PORT}"
    echo ""
    echo "The K6 Load Testing Dashboard will be automatically provisioned."
}

list_tests() {
    echo -e "${BLUE}Available K6 Tests (with Grafana output):${NC}"
    echo ""
    echo -e "  ${GREEN}smoke${NC}      - Smoke test (~1 min, 10 req/s)"
    echo -e "  ${GREEN}products${NC}   - Products service test (~3 min, up to 50 req/s)"
    echo -e "  ${GREEN}inventory${NC}  - Inventory service test (~3 min, up to 40 req/s)"
    echo -e "  ${GREEN}orders${NC}     - Orders service test (~3 min, up to 20 req/s)"
    echo -e "  ${GREEN}full${NC}       - Full scenario test (~7 min, multiple scenarios)"
    echo ""
    echo "Run a test with: $0 run <test-name>"
    echo ""
    echo "View results in Grafana at: http://localhost:${GRAFANA_PORT}"
}

check_prerequisites() {
    # Check if kubectl is available
    if ! command -v kubectl &> /dev/null; then
        echo -e "${RED}Error: kubectl is not installed or not in PATH${NC}"
        exit 1
    fi

    # Check if namespace exists
    if ! kubectl get namespace "$NAMESPACE" &> /dev/null; then
        echo -e "${RED}Error: Namespace '$NAMESPACE' does not exist${NC}"
        echo "Make sure your Kubernetes cluster is running and the namespace is created."
        exit 1
    fi
}

deploy_k6_scripts() {
    echo -e "${YELLOW}Deploying K6 test scripts...${NC}"
    
    if [ -f "$SCRIPT_DIR/k8s/k6-configmap.yaml" ]; then
        kubectl apply -f "$SCRIPT_DIR/k8s/k6-configmap.yaml"
        echo -e "${GREEN}✓ K6 scripts ConfigMap deployed${NC}"
    else
        echo -e "${RED}Error: k6-configmap.yaml not found${NC}"
        exit 1
    fi
}

deploy_influxdb() {
    echo -e "${YELLOW}Deploying InfluxDB...${NC}"
    
    kubectl apply -f "$SCRIPT_DIR/k8s/grafana/influxdb-deployment.yaml"
    
    # Wait for InfluxDB to be ready
    echo -e "${YELLOW}Waiting for InfluxDB to be ready...${NC}"
    kubectl wait --for=condition=available deployment/influxdb -n "$NAMESPACE" --timeout=120s 2>/dev/null || true
    
    # Additional wait for the database to initialize
    sleep 5
    
    echo -e "${GREEN}✓ InfluxDB deployed${NC}"
}

deploy_grafana_configmaps() {
    echo -e "${YELLOW}Deploying Grafana ConfigMaps...${NC}"
    
    kubectl apply -f "$SCRIPT_DIR/k8s/grafana/grafana-configmaps.yaml"
    
    echo -e "${GREEN}✓ Grafana ConfigMaps deployed${NC}"
}

deploy_grafana() {
    echo -e "${YELLOW}Deploying Grafana...${NC}"
    
    kubectl apply -f "$SCRIPT_DIR/k8s/grafana/grafana-deployment.yaml"
    
    # Wait for Grafana to be ready
    echo -e "${YELLOW}Waiting for Grafana to be ready...${NC}"
    kubectl wait --for=condition=available deployment/grafana -n "$NAMESPACE" --timeout=120s 2>/dev/null || true
    
    echo -e "${GREEN}✓ Grafana deployed${NC}"
}

deploy_infrastructure() {
    echo -e "${BLUE}=== Deploying K6 Load Testing with Grafana ===${NC}"
    echo ""
    
    check_prerequisites
    
    deploy_k6_scripts
    deploy_influxdb
    deploy_grafana_configmaps
    deploy_grafana
    
    echo ""
    echo -e "${GREEN}✓ Infrastructure deployed successfully!${NC}"
    echo ""
    echo -e "${BLUE}Access Grafana at:${NC}"
    echo -e "  http://localhost:${GRAFANA_PORT}"
    echo ""
    echo -e "${BLUE}Dashboard:${NC}"
    echo "  K6 Load Testing Dashboard (auto-provisioned)"
    echo ""
    echo -e "${BLUE}Credentials (for admin access):${NC}"
    echo "  Username: admin"
    echo "  Password: admin"
    echo ""
    echo -e "${YELLOW}Note: Anonymous access is enabled with Admin role.${NC}"
    echo ""
    echo -e "Run a test with: ${GREEN}$0 run smoke${NC}"
}

show_status() {
    echo -e "${BLUE}=== Grafana Load Testing Status ===${NC}"
    echo ""
    
    echo -e "${YELLOW}Deployments:${NC}"
    kubectl get deployments -n "$NAMESPACE" -l component=k6-grafana -o wide 2>/dev/null || echo "No deployments found"
    echo ""
    
    echo -e "${YELLOW}Services:${NC}"
    kubectl get services -n "$NAMESPACE" -l component=k6-grafana -o wide 2>/dev/null || echo "No services found"
    echo ""
    
    echo -e "${YELLOW}Pods:${NC}"
    kubectl get pods -n "$NAMESPACE" -l component=k6-grafana -o wide 2>/dev/null || echo "No pods found"
    echo ""
    
    echo -e "${YELLOW}K6 Load Test Jobs:${NC}"
    kubectl get jobs -n "$NAMESPACE" -l app=k6-grafana-load-tests -o wide 2>/dev/null || echo "No K6 jobs found"
    echo ""
    
    echo -e "${YELLOW}K6 Test Pods:${NC}"
    kubectl get pods -n "$NAMESPACE" -l app=k6-grafana-load-tests -o wide 2>/dev/null || echo "No K6 test pods found"
    echo ""
    
    # Check if Grafana is accessible
    if kubectl get svc grafana -n "$NAMESPACE" &>/dev/null; then
        echo -e "${GREEN}Grafana URL: http://localhost:${GRAFANA_PORT}${NC}"
    fi
}

run_test() {
    local test_name=$1
    local watch_logs=$2
    local job_name=""
    local script_name=""
    
    case "$test_name" in
        smoke|smoke-test)
            job_name="k6-grafana-smoke-test"
            script_name="smoke-test.js"
            ;;
        products|products-service)
            job_name="k6-grafana-products-test"
            script_name="products-service-test.js"
            ;;
        inventory|inventory-service)
            job_name="k6-grafana-inventory-test"
            script_name="inventory-service-test.js"
            ;;
        orders|orders-service)
            job_name="k6-grafana-orders-test"
            script_name="orders-service-test.js"
            ;;
        full|full-scenario)
            job_name="k6-grafana-full-scenario-test"
            script_name="full-scenario-test.js"
            ;;
        *)
            echo -e "${RED}Unknown test: $test_name${NC}"
            echo "Use --list to see available tests"
            exit 1
            ;;
    esac
    
    # Check if infrastructure is deployed
    if ! kubectl get deployment influxdb -n "$NAMESPACE" &>/dev/null; then
        echo -e "${YELLOW}Infrastructure not deployed. Deploying now...${NC}"
        deploy_infrastructure
        echo ""
    fi
    
    # Delete existing job if it exists
    kubectl delete job "$job_name" -n "$NAMESPACE" 2>/dev/null || true
    sleep 2
    
    echo -e "${YELLOW}Starting K6 test with Grafana output: $test_name${NC}"
    
    # Create the job
    cat <<EOF | kubectl apply -f -
apiVersion: batch/v1
kind: Job
metadata:
  name: $job_name
  namespace: $NAMESPACE
  labels:
    app: k6-grafana-load-tests
    test: $test_name
spec:
  ttlSecondsAfterFinished: 600
  backoffLimit: 0
  template:
    metadata:
      labels:
        app: k6-grafana-load-tests
        test: $test_name
    spec:
      restartPolicy: Never
      containers:
        - name: k6
          image: grafana/k6:latest
          command:
            - k6
            - run
            - --out
            - influxdb=$INFLUXDB_URL
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
              cpu: 200m
              memory: 256Mi
            limits:
              cpu: 500m
              memory: 512Mi
          volumeMounts:
            - name: scripts
              mountPath: /scripts
              readOnly: true
      volumes:
        - name: scripts
          configMap:
            name: k6-scripts
EOF
    
    echo -e "${GREEN}✓ Job created: $job_name${NC}"
    
    # Wait for pod to start
    echo -e "${YELLOW}Waiting for test pod to start...${NC}"
    kubectl wait --for=condition=Ready pod -l job-name="$job_name" -n "$NAMESPACE" --timeout=60s 2>/dev/null || true
    
    echo ""
    echo -e "${GREEN}Test started! View results in Grafana:${NC}"
    echo -e "  http://localhost:${GRAFANA_PORT}"
    echo ""
    
    if [ "$watch_logs" = "true" ]; then
        echo -e "${YELLOW}Streaming logs (Ctrl+C to stop watching):${NC}"
        echo ""
        kubectl logs -f -l job-name="$job_name" -n "$NAMESPACE"
    else
        echo -e "${BLUE}To view logs:${NC}"
        echo "  kubectl logs -f -l job-name=$job_name -n $NAMESPACE"
        echo ""
        echo -e "${BLUE}Or use:${NC}"
        echo "  $0 logs $test_name -f"
    fi
}

view_logs() {
    local test_name=$1
    local follow=$2
    local job_name=""
    
    case "$test_name" in
        smoke|smoke-test)
            job_name="k6-grafana-smoke-test"
            ;;
        products|products-service)
            job_name="k6-grafana-products-test"
            ;;
        inventory|inventory-service)
            job_name="k6-grafana-inventory-test"
            ;;
        orders|orders-service)
            job_name="k6-grafana-orders-test"
            ;;
        full|full-scenario)
            job_name="k6-grafana-full-scenario-test"
            ;;
        all)
            if [ "$follow" = "true" ]; then
                kubectl logs -f -l app=k6-grafana-load-tests -n "$NAMESPACE" --all-containers=true
            else
                kubectl logs -l app=k6-grafana-load-tests -n "$NAMESPACE" --all-containers=true
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
            job_name="k6-grafana-smoke-test"
            ;;
        products|products-service)
            job_name="k6-grafana-products-test"
            ;;
        inventory|inventory-service)
            job_name="k6-grafana-inventory-test"
            ;;
        orders|orders-service)
            job_name="k6-grafana-orders-test"
            ;;
        full|full-scenario)
            job_name="k6-grafana-full-scenario-test"
            ;;
        all)
            echo -e "${YELLOW}Stopping all K6 Grafana tests...${NC}"
            kubectl delete jobs -n "$NAMESPACE" -l app=k6-grafana-load-tests 2>/dev/null || true
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

cleanup_jobs() {
    echo -e "${YELLOW}Cleaning up completed K6 Grafana jobs...${NC}"
    
    # Delete completed jobs
    kubectl delete jobs -n "$NAMESPACE" -l app=k6-grafana-load-tests --field-selector status.successful=1 2>/dev/null || true
    
    # Delete failed jobs
    kubectl delete jobs -n "$NAMESPACE" -l app=k6-grafana-load-tests --field-selector status.failed=1 2>/dev/null || true
    
    echo -e "${GREEN}✓ Cleanup complete${NC}"
}

delete_infrastructure() {
    echo -e "${YELLOW}Deleting Grafana load testing infrastructure...${NC}"
    
    # Delete test jobs
    kubectl delete jobs -n "$NAMESPACE" -l app=k6-grafana-load-tests 2>/dev/null || true
    
    # Delete deployments
    kubectl delete deployment grafana -n "$NAMESPACE" 2>/dev/null || true
    kubectl delete deployment influxdb -n "$NAMESPACE" 2>/dev/null || true
    
    # Delete services
    kubectl delete service grafana -n "$NAMESPACE" 2>/dev/null || true
    kubectl delete service influxdb -n "$NAMESPACE" 2>/dev/null || true
    
    # Delete configmaps
    kubectl delete configmap grafana-datasources -n "$NAMESPACE" 2>/dev/null || true
    kubectl delete configmap grafana-dashboards-provider -n "$NAMESPACE" 2>/dev/null || true
    kubectl delete configmap grafana-k6-dashboard -n "$NAMESPACE" 2>/dev/null || true
    
    echo -e "${GREEN}✓ Infrastructure deleted${NC}"
}

# Parse arguments
WATCH_LOGS="false"
FOLLOW_LOGS="false"
COMMAND="deploy"
ARG=""

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -l|--list)
            list_tests
            exit 0
            ;;
        -w|--watch)
            WATCH_LOGS="true"
            shift
            ;;
        -f|--follow)
            FOLLOW_LOGS="true"
            shift
            ;;
        deploy|run|status|logs|stop|cleanup|delete)
            COMMAND="$1"
            shift
            if [[ $# -gt 0 && ! "$1" =~ ^- ]]; then
                ARG="$1"
                shift
            fi
            ;;
        *)
            if [ -z "$ARG" ]; then
                ARG="$1"
            fi
            shift
            ;;
    esac
done

# Check prerequisites for most commands
if [[ "$COMMAND" != "help" ]]; then
    check_prerequisites
fi

# Execute command
case "$COMMAND" in
    deploy)
        deploy_infrastructure
        ;;
    run)
        if [ -z "$ARG" ]; then
            ARG="smoke"
        fi
        run_test "$ARG" "$WATCH_LOGS"
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
        cleanup_jobs
        ;;
    delete)
        delete_infrastructure
        ;;
    *)
        show_help
        exit 1
        ;;
esac
