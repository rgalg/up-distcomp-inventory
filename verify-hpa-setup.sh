GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}=== Verifying HPA Setup ===${NC}\n"

# Check metrics-server
echo -e "${YELLOW}1. Checking metrics-server...${NC}"
if kubectl get deployment metrics-server -n kube-system &>/dev/null; then
  READY=$(kubectl get deployment metrics-server -n kube-system -o jsonpath='{.status.readyReplicas}')
  if [ "$READY" -gt 0 ]; then
    echo -e "${GREEN}✓${NC} metrics-server is running"
  else
    echo -e "${RED}✗${NC} metrics-server is not ready"
  fi
else
  echo -e "${RED}✗${NC} metrics-server is not installed"
fi

# Check node metrics
echo -e "\n${YELLOW}2. Checking node metrics...${NC}"
if kubectl top nodes &>/dev/null; then
  echo -e "${GREEN}✓${NC} Node metrics available"
  kubectl top nodes
else
  echo -e "${RED}✗${NC} Node metrics not available"
fi

# Check pod metrics
echo -e "\n${YELLOW}3. Checking pod metrics...${NC}"
if kubectl top pods -n inventory-system &>/dev/null; then
  echo -e "${GREEN}✓${NC} Pod metrics available"
  kubectl top pods -n inventory-system
else
  echo -e "${YELLOW}⚠${NC} Pod metrics not available yet (may need more time)"
fi

# Check HPA
echo -e "\n${YELLOW}4. Checking HPA configuration...${NC}"
HPA_COUNT=$(kubectl get hpa -n inventory-system --no-headers 2>/dev/null | wc -l)
if [ "$HPA_COUNT" -eq 3 ]; then
  echo -e "${GREEN}✓${NC} All 3 HPAs configured"
  kubectl get hpa -n inventory-system
else
  echo -e "${RED}✗${NC} Expected 3 HPAs, found $HPA_COUNT"
fi

# Check resource requests on deployments
echo -e "\n${YELLOW}5. Checking resource requests...${NC}"
for svc in products-service inventory-service orders-service; do
  CPU_REQ=$(kubectl get deployment $svc -n inventory-system -o jsonpath='{.spec.template.spec.containers[0].resources.requests.cpu}' 2>/dev/null)
  if [ -n "$CPU_REQ" ]; then
    echo -e "${GREEN}✓${NC} $svc has CPU request: $CPU_REQ"
  else
    echo -e "${RED}✗${NC} $svc missing CPU request"
  fi
done

echo -e "\n${YELLOW}=== Summary ===${NC}"
echo "HPA will show actual CPU percentages once metrics are available (1-2 minutes after deployment)"
echo "Monitor with: watch -n 2 'kubectl get hpa -n inventory-system'"
