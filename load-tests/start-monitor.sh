GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${YELLOW}=== HPA Monitoring Dashboard ===${NC}\n"
echo "Press Ctrl+C to stop"
echo ""

while true; do
    clear
    echo -e "${YELLOW}=== Horizontal Pod Autoscaler Status ===${NC}"
    echo "$(date '+%Y-%m-%d %H:%M:%S')"
    echo ""
    
    # HPA Status
    echo -e "${BLUE}HPA Metrics:${NC}"
    kubectl get hpa -n inventory-system -o custom-columns=\
NAME:.metadata.name,\
REFERENCE:.spec.scaleTargetRef.name,\
TARGETS:.status.currentMetrics[0].resource.current.averageUtilization,\
TARGET:.spec.metrics[0].resource.target.averageUtilization,\
MIN:.spec.minReplicas,\
MAX:.spec.maxReplicas,\
REPLICAS:.status.currentReplicas,\
DESIRED:.status.desiredReplicas 2>/dev/null || echo "No HPA found"
    
    echo ""
    echo -e "${BLUE}Pod Status:${NC}"
    kubectl get pods -n inventory-system \
        -l "app in (products-service,inventory-service,orders-service)" \
        -o custom-columns=\
NAME:.metadata.name,\
STATUS:.status.phase,\
READY:.status.conditions[?\(@.type==\'Ready\'\)].status,\
RESTARTS:.status.containerStatuses[0].restartCount,\
AGE:.metadata.creationTimestamp 2>/dev/null | \
        awk 'NR==1 || /products-service|inventory-service|orders-service/'
    
    echo ""
    echo -e "${BLUE}Resource Usage (from metrics-server):${NC}"
    kubectl top pods -n inventory-system --containers=false 2>/dev/null | \
        grep -E "NAME|products-service|inventory-service|orders-service" || \
        echo "Metrics not available (metrics-server may not be running)"
    
    echo ""
    echo -e "${BLUE}Recent HPA Events:${NC}"
    kubectl get events -n inventory-system \
        --field-selector involvedObject.kind=HorizontalPodAutoscaler \
        --sort-by='.lastTimestamp' \
        -o custom-columns=TIME:.lastTimestamp,MESSAGE:.message \
        2>/dev/null | tail -5 || echo "No recent events"
    
    sleep 2
done
