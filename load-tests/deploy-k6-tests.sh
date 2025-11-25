# Deploy K6 load tests to Kubernetes
# This script creates/updates the ConfigMap and runs K6 Jobs in-cluster
#
# Usage:
#   ./deploy-k6-tests.sh [test-name]
#
# Examples:
#   ./deploy-k6-tests.sh                    # Run smoke test (default)
#   ./deploy-k6-tests.sh smoke              # Run smoke test
#   ./deploy-k6-tests.sh products           # Run products service test
#   ./deploy-k6-tests.sh inventory          # Run inventory service test
#   ./deploy-k6-tests.sh orders             # Run orders service test
#   ./deploy-k6-tests.sh full               # Run full scenario test
#   ./deploy-k6-tests.sh --list             # List available tests
#   ./deploy-k6-tests.sh --cleanup          # Clean up completed jobs

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
NC='\033[0m' # No Color

show_help() {
    echo -e "${BLUE}K6 Load Test Deployer${NC}"
    echo ""
    echo "Usage: $0 [options] [test-name]"
    echo ""
    echo "Options:"
    echo "  -h, --help      Show this help message"
    echo "  -l, --list      List available tests"
    echo "  -c, --cleanup   Clean up completed jobs"
    echo "  -w, --watch     Watch test progress after starting"
    echo ""
    echo "Available tests:"
    echo "  smoke           Smoke test (default) - quick verification"
    echo "  products        Products service load test"
    echo "  inventory       Inventory service load test"
    echo "  orders          Orders service load test"
    echo "  full            Full scenario test"
    echo ""
    echo "Examples:"
    echo "  $0 smoke        # Run smoke test"
    echo "  $0 products -w  # Run products test and watch logs"
    echo "  $0 --cleanup    # Clean up completed jobs"
}

list_tests() {
    echo -e "${BLUE}Available K6 Tests:${NC}"
    echo ""
    echo -e "  ${GREEN}smoke${NC}      - Smoke test (30s, ~10 req/s)"
    echo -e "  ${GREEN}products${NC}   - Products service test (3m, up to 50 req/s)"
    echo -e "  ${GREEN}inventory${NC}  - Inventory service test (3m, up to 40 req/s)"
    echo -e "  ${GREEN}orders${NC}     - Orders service test (3m, up to 20 req/s)"
    echo -e "  ${GREEN}full${NC}       - Full scenario test (7m, multiple scenarios)"
    echo ""
    echo "Use: $0 <test-name> to run a specific test"
}

cleanup_jobs() {
    echo -e "${YELLOW}Cleaning up completed K6 jobs...${NC}"
    
    # Delete completed jobs
    kubectl delete jobs -n "$NAMESPACE" -l app=k6-load-tests --field-selector status.successful=1 2>/dev/null || true
    
    # Delete failed jobs
    kubectl delete jobs -n "$NAMESPACE" -l app=k6-load-tests --field-selector status.failed=1 2>/dev/null || true
    
    echo -e "${GREEN}SUCCESS: Cleanup complete${NC}"
}

deploy_configmap() {
    echo -e "${YELLOW}Deploying K6 scripts ConfigMap...${NC}"
    
    if [ -f "$SCRIPT_DIR/k8s/k6-configmap.yaml" ]; then
        kubectl apply -f "$SCRIPT_DIR/k8s/k6-configmap.yaml"
        echo -e "${GREEN}SUCCESS: ConfigMap deployed${NC}"
    else
        echo -e "${RED}ERROR:: k6-configmap.yaml not found${NC}"
        exit 1
    fi
}

run_test() {
    local test_name=$1
    local watch_logs=$2
    local job_name=""
    local script_name=""
    
    case "$test_name" in
        smoke|smoke-test)
            job_name="k6-smoke-test"
            script_name="smoke-test.js"
            ;;
        products|products-service)
            job_name="k6-products-test"
            script_name="products-service-test.js"
            ;;
        inventory|inventory-service)
            job_name="k6-inventory-test"
            script_name="inventory-service-test.js"
            ;;
        orders|orders-service)
            job_name="k6-orders-test"
            script_name="orders-service-test.js"
            ;;
        full|full-scenario)
            job_name="k6-full-scenario-test"
            script_name="full-scenario-test.js"
            ;;
        full-high|full-scenario-high-load)
            job_name="k6-grafana-full-scenario-test-high-load"
            script_name="full-scenario-test-high-load.js"
            ;;
        *)
            echo -e "${RED}Unknown test: $test_name${NC}"
            echo "Use --list to see available tests"
            exit 1
            ;;
    esac
    
    # Delete existing job if it exists
    kubectl delete job "$job_name" -n "$NAMESPACE" 2>/dev/null || true
    sleep 2
    
    echo -e "${YELLOW}Starting K6 test: $test_name${NC}"
    
    # Create the job
    cat <<EOF | kubectl apply -f -
apiVersion: batch/v1
kind: Job
metadata:
  name: $job_name
  namespace: $NAMESPACE
  labels:
    app: k6-load-tests
    test: $test_name
spec:
  ttlSecondsAfterFinished: 300
  backoffLimit: 0
  template:
    metadata:
      labels:
        app: k6-load-tests
        test: $test_name
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
    
    echo -e "${GREEN}SUCCESS: Job created: $job_name${NC}"
    
    # Wait for pod to start
    echo -e "${YELLOW}Waiting for pod to start...${NC}"
    kubectl wait --for=condition=Ready pod -l job-name="$job_name" -n "$NAMESPACE" --timeout=60s 2>/dev/null || true
    
    if [ "$watch_logs" = "true" ]; then
        echo -e "${YELLOW}Streaming logs (Ctrl+C to stop watching):${NC}"
        echo ""
        kubectl logs -f -l job-name="$job_name" -n "$NAMESPACE"
    else
        echo ""
        echo -e "${BLUE}To view logs:${NC}"
        echo "  kubectl logs -f -l job-name=$job_name -n $NAMESPACE"
        echo ""
        echo -e "${BLUE}To check job status:${NC}"
        echo "  kubectl get jobs $job_name -n $NAMESPACE"
    fi
}

# Parse arguments
WATCH_LOGS="false"
TEST_NAME="smoke"

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
        -c|--cleanup)
            cleanup_jobs
            exit 0
            ;;
        -w|--watch)
            WATCH_LOGS="true"
            shift
            ;;
        *)
            TEST_NAME="$1"
            shift
            ;;
    esac
done

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}ERROR:: kubectl is not installed or not in PATH${NC}"
    exit 1
fi

# Check if namespace exists
if ! kubectl get namespace "$NAMESPACE" &> /dev/null; then
    echo -e "${RED}ERROR:: Namespace '$NAMESPACE' does not exist${NC}"
    echo "Make sure your Kubernetes cluster is running and the namespace is created."
    exit 1
fi

# Deploy ConfigMap and run test
deploy_configmap
run_test "$TEST_NAME" "$WATCH_LOGS"
