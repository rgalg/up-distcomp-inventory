#!/bin/bash

set -e

echo "======================================"
echo "Verifying gRPC and Kubernetes Setup"
echo "======================================"
echo ""

# Check if required tools are installed
echo "Checking prerequisites..."

if ! command -v go &> /dev/null; then
    echo "ERROR: Go is not installed"
    exit 1
fi
echo "SUCCESS: Go is installed"

if ! command -v protoc &> /dev/null; then
    echo "ERROR: protoc is not installed"
    exit 1
fi
echo "SUCCESS: protoc is installed"

if ! command -v docker &> /dev/null; then
    echo "ERROR: Docker is not installed"
    exit 1
fi
echo "SUCCESS: Docker is installed"

if ! command -v kind &> /dev/null; then
    echo "WARNING: Kind is not installed (recommended for deployment)"
else
    echo "SUCCESS: Kind is installed"
fi

if ! command -v kubectl &> /dev/null; then
    echo "WARNING: kubectl is not installed (recommended for deployment)"
else
    echo "SUCCESS: kubectl is installed"
fi

echo ""
echo "Building services..."
echo ""

# Build Products service
echo "Building Products service..."
cd services/products
go build -o products ./cmd
if [ $? -eq 0 ]; then
    echo "SUCCESS: Products service builds successfully"
else
    echo "ERROR: Products service build failed"
    exit 1
fi

# Build Inventory service
echo "Building Inventory service..."
cd ../inventory
go build -o inventory ./cmd
if [ $? -eq 0 ]; then
    echo "SUCCESS: Inventory service builds successfully"
else
    echo "ERROR: Inventory service build failed"
    exit 1
fi

# Build Orders service
echo "Building Orders service..."
cd ../orders
go build -o orders ./cmd
if [ $? -eq 0 ]; then
    echo "SUCCESS: Orders service builds successfully"
else
    echo "ERROR: Orders service build failed"
    exit 1
fi

cd ../..

echo ""
echo "Verifying proto files..."
echo ""

# Check proto files exist
if [ -f "proto/products/products.proto" ]; then
    echo "SUCCESS: Products proto file exists"
else
    echo "ERROR: Products proto file missing"
    exit 1
fi

if [ -f "proto/inventory/inventory.proto" ]; then
    echo "SUCCESS: Inventory proto file exists"
else
    echo "ERROR: Inventory proto file missing"
    exit 1
fi

if [ -f "proto/orders/orders.proto" ]; then
    echo "SUCCESS: Orders proto file exists"
else
    echo "ERROR: Orders proto file missing"
    exit 1
fi

echo ""
echo "Verifying generated proto code..."
echo ""

# Check generated files
if [ -f "services/products/proto/products/products.pb.go" ]; then
    echo "SUCCESS: Products gRPC code generated"
else
    echo "ERROR: Products gRPC code missing"
    exit 1
fi

if [ -f "services/inventory/proto/inventory/inventory.pb.go" ]; then
    echo "SUCCESS: Inventory gRPC code generated"
else
    echo "ERROR: Inventory gRPC code missing"
    exit 1
fi

if [ -f "services/orders/proto/orders/orders.pb.go" ]; then
    echo "SUCCESS: Orders gRPC code generated"
else
    echo "ERROR: Orders gRPC code missing"
    exit 1
fi

echo ""
echo "Verifying Kubernetes manifests..."
echo ""

# Check K8s manifests
if [ -f "k8s/namespace.yaml" ]; then
    echo "SUCCESS: Namespace manifest exists"
else
    echo "ERROR: Namespace manifest missing"
    exit 1
fi

if [ -f "k8s/products-deployment.yaml" ]; then
    echo "SUCCESS: Products deployment manifest exists"
else
    echo "ERROR: Products deployment manifest missing"
    exit 1
fi

if [ -f "k8s/inventory-deployment.yaml" ]; then
    echo "SUCCESS: Inventory deployment manifest exists"
else
    echo "ERROR: Inventory deployment manifest missing"
    exit 1
fi

if [ -f "k8s/orders-deployment.yaml" ]; then
    echo "SUCCESS: Orders deployment manifest exists"
else
    echo "ERROR: Orders deployment manifest missing"
    exit 1
fi

if [ -f "k8s/frontend-deployment.yaml" ]; then
    echo "SUCCESS: Frontend deployment manifest exists"
else
    echo "ERROR: Frontend deployment manifest missing"
    exit 1
fi

if [ -f "kind-config.yaml" ]; then
    echo "SUCCESS: Kind configuration exists"
else
    echo "ERROR: Kind configuration missing"
    exit 1
fi

echo ""
echo "======================================"
echo "==== SUCCESS: All checks passed!  ===="
echo "======================================"
echo ""
echo "You can now deploy the application using:"
echo "  ./deploy-kind.sh"
