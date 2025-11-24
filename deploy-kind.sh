set -e

CLUSTER_NAME="inventory-cluster"

# helper function to wait for pods to be ready
wait_for_pods() {
    local label=$1
    local namespace=$2
    local timeout=${3:-120}
    
    echo "Waiting for pods with label $label in namespace $namespace..."
    local count=0
    until kubectl get pods -l "$label" -n "$namespace" 2>/dev/null | grep -q Running; do
        if [ $count -ge $timeout ]; then
            echo "Timeout waiting for pods with label $label"
            kubectl get pods -l "$label" -n "$namespace"
            return 1
        fi
        echo "Waiting... ($count/$timeout)"
        sleep 2
        count=$((count + 2))
    done
    echo "Pods with label $label are running in namespace $namespace."

    # wait for ready condition
    kubectl wait --for=condition=ready pod -l "$label" -n "$namespace" --timeout=${timeout}s
}

# verify that the .env file exists
if [ ! -f .env ]; then
    echo "Error: .env file not found!"
    echo "Please create .env file with database credentials."
    exit 1
fi

echo "======================================"
echo "====  Deploying to Kind Cluster   ===="
echo "======================================"

# create a Kind cluster if it doesn't exist
if ! kind get clusters | grep -q "^${CLUSTER_NAME}$"; then
    echo "Creating Kind cluster..."
    kind create cluster --config kind-config.yaml
else
    echo "Kind cluster already exists"
fi

# create the namespace
echo "Creating namespace..."
kubectl apply -f k8s/namespace.yaml

# build Docker images
echo "Building Docker images..."
# products service
echo "Building products-service..."
docker build -t products-service:latest ./services/products
# inventory service
echo "Building inventory-service..."
docker build -t inventory-service:latest ./services/inventory
# orders service
echo "Building orders-service..."
docker build -t orders-service:latest ./services/orders
# frontend
echo "Building frontend..."
docker build -t frontend:latest ./frontend

# load images into cluster
echo "Loading images into Kind cluster..."
kind load docker-image products-service:latest --name ${CLUSTER_NAME}
kind load docker-image inventory-service:latest --name ${CLUSTER_NAME}
kind load docker-image orders-service:latest --name ${CLUSTER_NAME}
kind load docker-image frontend:latest --name ${CLUSTER_NAME}

# PostgreSQL deployment
# generate the secret configuration file
echo "Generating PostgreSQL configuration from .env file..."
bash k8s/postgres-create-config-from-env.sh
# apply the generated configuration yaml
echo "Applying PostgreSQL configuration..."
kubectl apply -f k8s/postgres-config-generated.yaml
# deploy PostgreSQL
echo "Deploying PostgreSQL..."
kubectl apply -f k8s/postgres-deployment.yaml
# wait for PostgreSQL deployment to be ready
echo "Waiting for PostgreSQL deployment to be ready..."
kubectl wait --for=condition=available --timeout=120s deployment/postgres -n inventory-system
# wait for the pod to be ready
echo "Waiting for PostgreSQL pod to be ready..."
kubectl wait --for=condition=ready pod -l app=postgres -n inventory-system --timeout=120s
# we'll give postgres a few seconds to fully initialize
echo "Allowing PostgreSQL to fully initialize: sleeping for 5 seconds..."
sleep 5
# setup the database
echo "Setting up PostgreSQL database..."
kubectl apply -f k8s/postgres-init-job.yaml
kubectl wait --for=condition=complete job/postgres-init -n inventory-system --timeout=120s
# wait for the db init job to be registered in Kubernetes
echo "Waiting for the postgres-init job to be created..."
sleep 3
until kubectl get job postgres-init -n inventory-system &> /dev/null; do
    echo "Waiting for job to be registered..."
    sleep 2
done
# wait for the db init job to complete
echo "Waiting for init job to complete..."
kubectl wait --for=condition=complete job/postgres-init -n inventory-system --timeout=120s

# deploy to Kubernetes
echo "Deploying to services Kubernetes..."
kubectl apply -f k8s/products-deployment.yaml
kubectl apply -f k8s/inventory-deployment.yaml
kubectl apply -f k8s/orders-deployment.yaml
kubectl apply -f k8s/frontend-deployment.yaml

# Wait for deployments to be ready
echo "Waiting for deployments to be ready..."
kubectl wait --for=condition=available --timeout=180s deployment/products-service -n inventory-system
kubectl wait --for=condition=available --timeout=180s deployment/inventory-service -n inventory-system
kubectl wait --for=condition=available --timeout=180s deployment/orders-service -n inventory-system
kubectl wait --for=condition=available --timeout=180s deployment/frontend -n inventory-system

echo "==========================================="
echo "==== Deployment completed successfully ===="
echo "==========================================="
echo ""
echo "Access the application at:"
echo "  Frontend: http://localhost:3000"
echo "  Products API: http://localhost:8001"
echo "  Inventory API: http://localhost:8002"
echo "  Orders API: http://localhost:8003"
echo ""
echo "To check status: kubectl get pods -n inventory-system"
echo "To view logs: kubectl logs -n inventory-system <pod-name>"
echo "          or  kubectl logs -l app=<service-name> -n inventory-system"
echo "To delete cluster: kind delete cluster --name ${CLUSTER_NAME}"
