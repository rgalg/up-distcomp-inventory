#!/bin/bash

set -e

echo "======================================"
echo "Verifying gRPC and Kubernetes Setup"
echo "======================================"
echo ""

# Check if required tools are installed
echo "Checking prerequisites..."

if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed"
    exit 1
fi
echo "✓ Go is installed"

if ! command -v protoc &> /dev/null; then
    echo "❌ protoc is not installed"
    exit 1
fi
echo "✓ protoc is installed"

if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed"
    exit 1
fi
echo "✓ Docker is installed"

if ! command -v kind &> /dev/null; then
    echo "⚠ Kind is not installed (optional for deployment)"
else
    echo "✓ Kind is installed"
fi

if ! command -v kubectl &> /dev/null; then
    echo "⚠ kubectl is not installed (optional for deployment)"
else
    echo "✓ kubectl is installed"
fi

echo ""
echo "Building services..."
echo ""

# Build Products service
echo "Building Products service..."
cd services/products
go build -o products ./cmd
if [ $? -eq 0 ]; then
    echo "✓ Products service builds successfully"
else
    echo "❌ Products service build failed"
    exit 1
fi

# Build Inventory service
echo "Building Inventory service..."
cd ../inventory
go build -o inventory ./cmd
if [ $? -eq 0 ]; then
    echo "✓ Inventory service builds successfully"
else
    echo "❌ Inventory service build failed"
    exit 1
fi

# Build Orders service
echo "Building Orders service..."
cd ../orders
go build -o orders ./cmd
if [ $? -eq 0 ]; then
    echo "✓ Orders service builds successfully"
else
    echo "❌ Orders service build failed"
    exit 1
fi

cd ../..

echo ""
echo "Verifying proto files..."
echo ""

# Check proto files exist
if [ -f "proto/products/products.proto" ]; then
    echo "✓ Products proto file exists"
else
    echo "❌ Products proto file missing"
    exit 1
fi

if [ -f "proto/inventory/inventory.proto" ]; then
    echo "✓ Inventory proto file exists"
else
    echo "❌ Inventory proto file missing"
    exit 1
fi

if [ -f "proto/orders/orders.proto" ]; then
    echo "✓ Orders proto file exists"
else
    echo "❌ Orders proto file missing"
    exit 1
fi

echo ""
echo "Verifying generated proto code..."
echo ""

# Check generated files
if [ -f "services/products/proto/products/products.pb.go" ]; then
    echo "✓ Products gRPC code generated"
else
    echo "❌ Products gRPC code missing"
    exit 1
fi

if [ -f "services/inventory/proto/inventory/inventory.pb.go" ]; then
    echo "✓ Inventory gRPC code generated"
else
    echo "❌ Inventory gRPC code missing"
    exit 1
fi

if [ -f "services/orders/proto/orders/orders.pb.go" ]; then
    echo "✓ Orders gRPC code generated"
else
    echo "❌ Orders gRPC code missing"
    exit 1
fi

echo ""
echo "Verifying Kubernetes manifests..."
echo ""

# Check K8s manifests
if [ -f "k8s/namespace.yaml" ]; then
    echo "✓ Namespace manifest exists"
else
    echo "❌ Namespace manifest missing"
    exit 1
fi

if [ -f "k8s/products-deployment.yaml" ]; then
    echo "✓ Products deployment manifest exists"
else
    echo "❌ Products deployment manifest missing"
    exit 1
fi

if [ -f "k8s/inventory-deployment.yaml" ]; then
    echo "✓ Inventory deployment manifest exists"
else
    echo "❌ Inventory deployment manifest missing"
    exit 1
fi

if [ -f "k8s/orders-deployment.yaml" ]; then
    echo "✓ Orders deployment manifest exists"
else
    echo "❌ Orders deployment manifest missing"
    exit 1
fi

if [ -f "k8s/frontend-deployment.yaml" ]; then
    echo "✓ Frontend deployment manifest exists"
else
    echo "❌ Frontend deployment manifest missing"
    exit 1
fi

if [ -f "kind-config.yaml" ]; then
    echo "✓ Kind configuration exists"
else
    echo "❌ Kind configuration missing"
    exit 1
fi

echo ""
echo "======================================"
echo "✓ All checks passed!"
echo "======================================"
echo ""
echo "You can now deploy the application using:"
echo "  ./deploy-kind.sh"
