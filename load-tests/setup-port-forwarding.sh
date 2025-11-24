# Kill any existing port forwards
pkill -f "port-forward.*inventory-system" || true

# Setup port forwards
kubectl port-forward -n inventory-system svc/products-service 8001:8001 > /tmp/pf-products.log 2>&1 &
kubectl port-forward -n inventory-system svc/inventory-service 8002:8002 > /tmp/pf-inventory.log 2>&1 &
kubectl port-forward -n inventory-system svc/orders-service 8003:8003 > /tmp/pf-orders.log 2>&1 &

# Wait for connections
sleep 5

# Verify health
for port in 8001 8002 8003; do
    echo -n "Port $port: "
    curl -s http://localhost:$port/health && echo " Seems OK" || echo " Failed"
done
