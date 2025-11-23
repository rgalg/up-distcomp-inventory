#!/bin/bash

set -e

echo "======================================"
echo "Deploying to Kind Cluster"
echo "======================================"

# Create Kind cluster if it doesn't exist
if ! kind get clusters | grep -q inventory-cluster; then
    echo "Creating Kind cluster..."
    kind create cluster --config kind-config.yaml
else
    echo "Kind cluster already exists"
fi

# Build Docker images
echo "Building Docker images..."

echo "Building products-service..."
docker build -t products-service:latest ./services/products

echo "Building inventory-service..."
docker build -t inventory-service:latest ./services/inventory

echo "Building orders-service..."
docker build -t orders-service:latest ./services/orders

echo "Building frontend..."
docker build -t frontend:latest ./frontend

# Load images into Kind cluster
echo "Loading images into Kind cluster..."
kind load docker-image products-service:latest --name inventory-cluster
kind load docker-image inventory-service:latest --name inventory-cluster
kind load docker-image orders-service:latest --name inventory-cluster
kind load docker-image frontend:latest --name inventory-cluster

# Deploy to Kubernetes
echo "Deploying to Kubernetes..."
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/products-deployment.yaml
kubectl apply -f k8s/inventory-deployment.yaml
kubectl apply -f k8s/orders-deployment.yaml
kubectl apply -f k8s/frontend-deployment.yaml

# Wait for deployments to be ready
echo "Waiting for deployments to be ready..."
kubectl wait --for=condition=available --timeout=120s deployment/products-service -n inventory-system
kubectl wait --for=condition=available --timeout=120s deployment/inventory-service -n inventory-system
kubectl wait --for=condition=available --timeout=120s deployment/orders-service -n inventory-system
kubectl wait --for=condition=available --timeout=120s deployment/frontend -n inventory-system

echo "======================================"
echo "Deployment completed successfully!"
echo "======================================"
echo ""
echo "Access the application at:"
echo "  Frontend: http://localhost:3000"
echo "  Products API: http://localhost:8001"
echo "  Inventory API: http://localhost:8002"
echo "  Orders API: http://localhost:8003"
echo ""
echo "To check status: kubectl get pods -n inventory-system"
echo "To view logs: kubectl logs -n inventory-system <pod-name>"
echo "To delete cluster: kind delete cluster --name inventory-cluster"
