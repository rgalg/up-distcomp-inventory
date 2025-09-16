#!/bin/bash

echo "🚀 Starting Inventory Management System..."

# Build and start all services
docker compose up --build -d

echo "⏳ Waiting for services to start..."
sleep 10

# Check service health
echo "🔍 Checking service health..."

if curl -f http://localhost:8001/health > /dev/null 2>&1; then
    echo "✅ Products service is healthy"
else
    echo "❌ Products service is not responding"
fi

if curl -f http://localhost:8002/health > /dev/null 2>&1; then
    echo "✅ Inventory service is healthy"
else
    echo "❌ Inventory service is not responding"
fi

if curl -f http://localhost:8003/health > /dev/null 2>&1; then
    echo "✅ Orders service is healthy"
else
    echo "❌ Orders service is not responding"
fi

if curl -f http://localhost:3000 > /dev/null 2>&1; then
    echo "✅ Frontend is accessible"
else
    echo "❌ Frontend is not responding"
fi

echo ""
echo "🎉 Application is ready!"
echo "📱 Frontend: http://localhost:3000"
echo "🔧 Products API: http://localhost:8001"
echo "📦 Inventory API: http://localhost:8002"
echo "📋 Orders API: http://localhost:8003"
echo ""
echo "To stop the application, run: docker compose down"