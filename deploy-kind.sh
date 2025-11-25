set -e

CLUSTER_NAME="inventory-cluster"


GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# helper function to wait for pods to be ready
wait_for_pods() {
    local label=$1
    local namespace=$2
    local timeout=${3:-120}
    
    printf "Waiting for pods with label $label in namespace $namespace...\n"
    local count=0
    until kubectl get pods -l "$label" -n "$namespace" 2>/dev/null | grep -q Running; do
        if [ $count -ge $timeout ]; then
            printf "Timeout waiting for pods with label $label in namespace $namespace\n"
            kubectl get pods -l "$label" -n "$namespace"
            return 1
        fi
        printf "Waiting... ($count/$timeout)\n"
        sleep 2
        count=$((count + 2))
    done
    printf "Pods with label $label are running in namespace $namespace.\n"

    # wait for ready condition
    kubectl wait --for=condition=ready pod -l "$label" -n "$namespace" --timeout=${timeout}s
}

# verify that the .env file exists
if [ ! -f .env ]; then
    printf "Error: .env file not found!\n"
    printf "Please create .env file with database credentials.\n"
    exit 1
fi

printf "======================================\n"
printf "====  Deploying to Kind Cluster   ====\n"
printf "======================================\n"

# create a Kind cluster if it doesn't exist
if ! kind get clusters | grep -q "^${CLUSTER_NAME}$"; then
    printf "Creating Kind cluster...\n"
    kind create cluster --config kind-config.yaml
else
    printf "Kind cluster already exists\n"
fi

# create the namespace
printf "Creating namespace...\n"
kubectl apply -f k8s/namespace.yaml

printf "======================================\n\n"

# -----------------------------------------
# build Docker images
# -----------------------------------------

printf "======================================\n"
printf "====    Building Docker images    ====\n"
printf "======================================\n"

# products service
printf "Building products-service...\n"
docker build -t products-service:latest ./services/products
# inventory service
printf "Building inventory-service...\n"
docker build -t inventory-service:latest ./services/inventory
# orders service
printf "Building orders-service...\n"
docker build -t orders-service:latest ./services/orders
# frontend
printf "Building frontend...\n"
docker build -t frontend:latest ./frontend

printf "======================================\n\n"

# -----------------------------------------
# load images into cluster
# -----------------------------------------
printf "======================================\n"
printf "====    Loading Docker images     ====\n"
printf "======================================\n"

kind load docker-image products-service:latest --name ${CLUSTER_NAME}
kind load docker-image inventory-service:latest --name ${CLUSTER_NAME}
kind load docker-image orders-service:latest --name ${CLUSTER_NAME}
kind load docker-image frontend:latest --name ${CLUSTER_NAME}

printf "======================================\n\n"

# -----------------------------------------
# PostgreSQL deployment
# -----------------------------------------

printf "======================================\n"
printf "====     Deploying PostgreSQL     ====\n"
printf "======================================\n"

# generate the secret configuration file
printf "Generating PostgreSQL configuration from .env file...\n"
bash k8s/postgres-create-config-from-env.sh
# apply the generated configuration
printf "Applying PostgreSQL configuration...\n"
kubectl apply -f k8s/postgres-config-generated.yaml

# generate the init ConfigMap from SQL schema
printf "Generating PostgreSQL init ConfigMap from SQL schema...\n"
bash k8s/postgres-create-init-configmap.sh
# verify that the DB ConfigMap was created
printf "Verifying postgres-config ConfigMap...\n"
if ! kubectl get configmap postgres-config -n inventory-system &>/dev/null; then
    printf "ERROR: postgres-config ConfigMap was not created!\n"
    exit 1
fi

# deploy PostgreSQL
printf "Deploying PostgreSQL...\n"
kubectl apply -f k8s/postgres-deployment.yaml

# wait for PostgreSQL deployment to be ready
printf "Waiting for PostgreSQL deployment to be ready...\n"
kubectl wait --for=condition=available --timeout=120s deployment/postgres -n inventory-system
# wait for the pod to be ready
printf "Waiting for PostgreSQL pod to be ready...\n"
kubectl wait --for=condition=ready pod -l app=postgres -n inventory-system --timeout=120s
# we'll give postgres a few seconds to fully initialize
printf "Allowing PostgreSQL to fully initialize: sleeping for 5 seconds...\n"
sleep 5

# setup the database
printf "Setting up PostgreSQL database...\n"
kubectl apply -f k8s/postgres-init-job.yaml
kubectl wait --for=condition=complete job/postgres-init -n inventory-system --timeout=120s

# wait for the db init job to be registered in Kubernetes
printf "Waiting for the postgres-init job to be created...\n"
sleep 3
until kubectl get job postgres-init -n inventory-system &> /dev/null; do
    printf "Waiting for job to be registered...\n"
    sleep 2
done
# show DB creation logs in real-time
printf "Monitoring database initialization...\n"
kubectl logs -f -n inventory-system -l job-name=postgres-init 2>/dev/null &
LOGS_PID=$!
# wait for the db init job to complete
printf "Waiting for init job to complete...\n"
if kubectl wait --for=condition=complete job/postgres-init -n inventory-system --timeout=120s; then
    printf "Database initialization completed successfully\n"
    kill $LOGS_PID 2>/dev/null || true
    wait $LOGS_PID 2>/dev/null || true
else
    printf "ERROR: Database initialization failed or timed out\n"
    kill $LOGS_PID 2>/dev/null || true
    wait $LOGS_PID 2>/dev/null || true
    printf "\n"
    printf "Job status:\n"
    kubectl describe job postgres-init -n inventory-system
    printf "\n"
    printf "Pod logs:\n"
    kubectl logs -n inventory-system -l job-name=postgres-init --tail=50
    exit 1
fi

printf "======================================\n\n"

# -----------------------------------------
# deploy application services
# -----------------------------------------

printf "======================================\n"
printf "====    Deploying Application     ====\n"
printf "======================================\n"

# deploy services to Kubernetes
printf "Deploying services to Kubernetes...\n"
kubectl apply -f k8s/products-deployment.yaml
kubectl apply -f k8s/inventory-deployment.yaml
kubectl apply -f k8s/orders-deployment.yaml
kubectl apply -f k8s/frontend-deployment.yaml

# Wait for deployments to be ready
printf "Waiting for deployments to be ready...\n"
kubectl wait --for=condition=available --timeout=180s deployment/products-service -n inventory-system
kubectl wait --for=condition=available --timeout=180s deployment/inventory-service -n inventory-system
kubectl wait --for=condition=available --timeout=180s deployment/orders-service -n inventory-system
kubectl wait --for=condition=available --timeout=180s deployment/frontend -n inventory-system

# deploy Horizontal Pod Autoscalers
printf "Deploying Horizontal Pod Autoscalers...\n"
kubectl apply -f k8s/hpa.yaml

printf "======================================\n\n"

# -----------------------------------------
# metrics server for HPA
# -----------------------------------------

printf "======================================\n"
printf "====   Deploying Metrics Server   ====\n"
printf "======================================\n"

# Install metrics-server
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml

echo "Waiting for metrics-server deployment to be created (5 seconds)..."
sleep 5

# Patch metrics-server for Kind
kubectl patch deployment metrics-server -n kube-system --type='json' -p='[
  {
    "op": "add",
    "path": "/spec/template/spec/containers/0/args/-",
    "value": "--kubelet-insecure-tls"
  },
  {
    "op": "add",
    "path": "/spec/template/spec/containers/0/args/-",
    "value": "--kubelet-preferred-address-types=InternalIP"
  }
]' 2>/dev/null || echo "metrics-server patch already applied or failed"

echo "Waiting for metrics-server to be ready..."
kubectl wait --for=condition=ready pod -l k8s-app=metrics-server -n kube-system --timeout=90s || {
  echo -e "${YELLOW}⚠${NC} metrics-server not ready yet, it may take a few more minutes"
}

echo -e "\n${GREEN}SUCCESS:${NC} metrics-server installed"

# Wait a bit for initial metrics to populate
echo -e "\n${YELLOW}=== Verifying HPA Configuration ===${NC}"
# Wait for the HPA
echo "Waiting for metrics to populate (30 seconds)..."
sleep 3
until kubectl get hpa -n inventory-system &>/dev/null; do
    printf "Waiting HPA metrics server to be ready...\n"
    sleep 2
done
if kubectl get hpa -n inventory-system &>/dev/null; then
  echo -e "${GREEN}SUCESS:${NC} HPA resources created:"
  kubectl get hpa -n inventory-system
else
  echo -e "${RED}ERROR:${NC} No HPA resources found"
fi

echo -e "\n${YELLOW}Note: It may take 1-2 minutes for HPA metrics to show actual values instead of <unknown>${NC}"
echo -e "Check status with: ${GREEN}kubectl get hpa -n inventory-system${NC}"
echo -e "Monitor during load tests with: ${GREEN}watch -n 2 'kubectl get hpa -n inventory-system'${NC}"

printf "======================================\n\n"

# ------------------------------------------
# Success message
# ------------------------------------------

printf "===========================================\n"
printf "==== Deployment completed successfully ====\n"
printf "===========================================\n"

printf "\n"
printf "Access the application at:\n"
printf "  Frontend: http://localhost:3000 (via LoadBalancer)\n"
printf "\n"
printf "Backend services use ClusterIP and are only accessible within the cluster.\n"
printf "Use 'kubectl port-forward' for debugging backend services:\n"
printf "  kubectl port-forward -n inventory-system svc/products-service 8001:8001\n"
printf "  kubectl port-forward -n inventory-system svc/inventory-service 8002:8002\n"
printf "  kubectl port-forward -n inventory-system svc/orders-service 8003:8003\n"
printf "\n"
printf "To check status: kubectl get pods -n inventory-system\n"
printf "To view logs: kubectl logs -n inventory-system <pod-name>\n"
printf "          or  kubectl logs -l app=<service-name> -n inventory-system\n"
printf "To delete cluster: kind delete cluster --name ${CLUSTER_NAME}\n"

printf "===========================================\n\n"
