#!/bin/bash

# Create and connect to K6 debug pod for interactive testing
# This pod stays running and allows you to manually run K6 tests
#
# Usage:
#   ./create-debug-pod.sh           # Create pod and exec into it
#   ./create-debug-pod.sh --delete  # Delete the debug pod

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
NAMESPACE="inventory-system"
POD_NAME="k6-debug"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

show_help() {
    echo -e "${BLUE}K6 Debug Pod Manager${NC}"
    echo ""
    echo "Usage: $0 [options]"
    echo ""
    echo "Options:"
    echo "  -h, --help     Show this help message"
    echo "  -d, --delete   Delete the debug pod"
    echo "  -s, --status   Show debug pod status"
    echo "  -c, --connect  Connect to existing debug pod"
    echo ""
    echo "Without options, creates the debug pod and connects to it."
    echo ""
    echo "Once inside the pod, you can run:"
    echo "  k6 run /scripts/smoke-test.js"
    echo "  k6 run /scripts/products-service-test.js"
    echo "  k6 run /scripts/high-load/full-scenario-test.js"
}

delete_pod() {
    echo -e "${YELLOW}Deleting debug pod...${NC}"
    kubectl delete pod "$POD_NAME" -n "$NAMESPACE" 2>/dev/null || echo "Pod not found"
    echo -e "${GREEN}✓ Debug pod deleted${NC}"
}

show_status() {
    echo -e "${BLUE}Debug Pod Status:${NC}"
    kubectl get pod "$POD_NAME" -n "$NAMESPACE" -o wide 2>/dev/null || echo "Pod not found"
}

deploy_scripts() {
    echo -e "${YELLOW}Deploying K6 scripts...${NC}"
    
    # Deploy standard scripts ConfigMap
    if [ -f "$SCRIPT_DIR/k8s/k6-configmap.yaml" ]; then
        kubectl apply -f "$SCRIPT_DIR/k8s/k6-configmap.yaml"
    fi
    
    # Deploy high-load scripts ConfigMap
    if [ -d "$SCRIPT_DIR/k8s/high-load" ]; then
        kubectl create configmap k6-high-load-scripts \
            --from-file="$SCRIPT_DIR/k8s/high-load/" \
            -n "$NAMESPACE" \
            --dry-run=client -o yaml | kubectl apply -f -
    fi
    
    echo -e "${GREEN}✓ Scripts deployed${NC}"
}

create_pod() {
    # Check if pod already exists
    if kubectl get pod "$POD_NAME" -n "$NAMESPACE" &>/dev/null; then
        echo -e "${YELLOW}Debug pod already exists${NC}"
        return 0
    fi
    
    echo -e "${YELLOW}Creating debug pod...${NC}"
    
    # Apply the debug pod manifest
    if [ -f "$SCRIPT_DIR/k8s/k6-debug-pod.yaml" ]; then
        kubectl apply -f "$SCRIPT_DIR/k8s/k6-debug-pod.yaml"
    else
        # Create inline if manifest doesn't exist
        cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: $POD_NAME
  namespace: $NAMESPACE
  labels:
    app: k6-debug
spec:
  containers:
    - name: k6
      image: grafana/k6:latest
      command:
        - sleep
        - infinity
      env:
        - name: BASE_URL
          value: "http://products-service:8001"
        - name: PRODUCTS_URL
          value: "http://products-service:8001"
        - name: INVENTORY_URL
          value: "http://inventory-service:8002"
        - name: ORDERS_URL
          value: "http://orders-service:8003"
      resources:
        requests:
          cpu: 100m
          memory: 128Mi
        limits:
          cpu: 500m
          memory: 512Mi
      volumeMounts:
        - name: scripts
          mountPath: /scripts
          readOnly: true
        - name: high-load-scripts
          mountPath: /scripts/high-load
          readOnly: true
  volumes:
    - name: scripts
      configMap:
        name: k6-scripts
    - name: high-load-scripts
      configMap:
        name: k6-high-load-scripts
        optional: true
  restartPolicy: Never
EOF
    fi
    
    echo -e "${GREEN}✓ Debug pod created${NC}"
}

wait_for_pod() {
    echo -e "${YELLOW}Waiting for pod to be ready...${NC}"
    kubectl wait --for=condition=Ready pod/"$POD_NAME" -n "$NAMESPACE" --timeout=120s
}

connect_to_pod() {
    echo ""
    echo -e "${BLUE}╔══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║                    K6 Debug Pod                               ║${NC}"
    echo -e "${BLUE}╠══════════════════════════════════════════════════════════════╣${NC}"
    echo -e "${BLUE}║  Available scripts:                                          ║${NC}"
    echo -e "${BLUE}║    /scripts/smoke-test.js                                    ║${NC}"
    echo -e "${BLUE}║    /scripts/products-service-test.js                         ║${NC}"
    echo -e "${BLUE}║    /scripts/inventory-service-test.js                        ║${NC}"
    echo -e "${BLUE}║    /scripts/orders-service-test.js                           ║${NC}"
    echo -e "${BLUE}║    /scripts/full-scenario-test.js                            ║${NC}"
    echo -e "${BLUE}║    /scripts/high-load/*.js (if deployed)                     ║${NC}"
    echo -e "${BLUE}║                                                              ║${NC}"
    echo -e "${BLUE}║  Example commands:                                           ║${NC}"
    echo -e "${BLUE}║    k6 run /scripts/smoke-test.js                             ║${NC}"
    echo -e "${BLUE}║    k6 run --vus 5 --duration 1m /scripts/smoke-test.js       ║${NC}"
    echo -e "${BLUE}║    k6 run /scripts/high-load/products-service-test.js        ║${NC}"
    echo -e "${BLUE}║                                                              ║${NC}"
    echo -e "${BLUE}║  Type 'exit' to leave the pod                                ║${NC}"
    echo -e "${BLUE}╚══════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    
    kubectl exec -it "$POD_NAME" -n "$NAMESPACE" -- /bin/sh
}

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}Error: kubectl is not installed or not in PATH${NC}"
    exit 1
fi

# Check if namespace exists
if ! kubectl get namespace "$NAMESPACE" &> /dev/null; then
    echo -e "${RED}Error: Namespace '$NAMESPACE' does not exist${NC}"
    exit 1
fi

# Parse arguments
case "$1" in
    -h|--help)
        show_help
        exit 0
        ;;
    -d|--delete)
        delete_pod
        exit 0
        ;;
    -s|--status)
        show_status
        exit 0
        ;;
    -c|--connect)
        connect_to_pod
        exit 0
        ;;
    "")
        # Default: deploy scripts, create pod, and connect
        deploy_scripts
        create_pod
        wait_for_pod
        connect_to_pod
        ;;
    *)
        echo -e "${RED}Unknown option: $1${NC}"
        show_help
        exit 1
        ;;
esac
