# Frontend

A simple HTML/CSS/JavaScript web application for the Inventory Management System.

## Overview

The Frontend provides a user interface for:
- Viewing the product catalog
- Checking inventory status and availability
- Creating new orders
- Viewing and fulfilling existing orders

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Frontend                                │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │                    NGINX Web Server                        │ │
│  │                       (Port 80)                            │ │
│  └───────────────────────────┬────────────────────────────────┘ │
│                              │                                  │
│     ┌────────────────────────┼────────────────────────┐         │
│     ▼                        ▼                        ▼         │
│  ┌─────────┐          ┌──────────┐            ┌──────────┐      │
│  │  HTML   │          │   CSS    │            │    JS    │      │
│  │ (Views) │          │ (Styles) │            │ (Logic)  │      │
│  └─────────┘          └──────────┘            └──────────┘      │
│                                                                 │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │              NGINX Reverse Proxy                           │ │
│  │                                                            │ │
│  │   /api/products/*  ────►  products-service:8001            │ │
│  │   /api/inventory/* ────►  inventory-service:8002           │ │
│  │   /api/orders/*    ────►  orders-service:8003              │ │
│  └────────────────────────────────────────────────────────────┘ │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## Features

### Products View
- Display all products from the catalog
- Show product details: name, description, category, price

### Inventory View
- Display current stock levels for all products
- Show available and reserved quantities
- Calculate available stock (total - reserved)

### Orders View
- List all orders with status
- Show order details: customer, items, total, status
- Fulfill pending orders with a single click

### Create Order
- Select multiple products
- Specify quantities
- Submit order with automatic stock reservation

## Project Structure

```
frontend/
├── index.html           # Main HTML page
├── script.js            # JavaScript application logic
├── styles.css           # CSS styles
├── nginx.conf           # NGINX configuration
├── Dockerfile           # Container configuration
└── README.md            # This file
```

## API Routes

The frontend communicates with backend services through NGINX reverse proxy:

| Frontend Route | Backend Service | Backend Endpoint |
|----------------|-----------------|------------------|
| `/api/products/*` | products-service:8001 | `/*` |
| `/api/inventory/*` | inventory-service:8002 | `/*` |
| `/api/orders/*` | orders-service:8003 | `/*` |

## Configuration

### NGINX Configuration

The `nginx.conf` file configures:
- Static file serving from `/usr/share/nginx/html`
- Reverse proxy to backend services
- CORS headers for development

### Backend Service URLs

The NGINX proxy uses Kubernetes DNS names to reach backend services:
```nginx
proxy_pass http://products-service.inventory-system.svc.cluster.local:8001/;
proxy_pass http://inventory-service.inventory-system.svc.cluster.local:8002/;
proxy_pass http://orders-service.inventory-system.svc.cluster.local:8003/;
```

## Running Locally

### Prerequisites
- Any web server (e.g., Python's `http.server`, nginx, Apache)
- Backend services running and accessible

### Development (without backend)

For static file serving only:
```bash
cd frontend
python3 -m http.server 3000
```

**Note:** API calls will fail without backend services.

### With Docker

```bash
docker build -t frontend:latest ./frontend
docker run -p 3000:80 frontend:latest
```

## Building

### Docker Build

```bash
docker build -t frontend:latest ./frontend
```

## Kubernetes Deployment

The frontend is deployed as a LoadBalancer service, making it externally accessible:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: frontend
  namespace: inventory-system
spec:
  selector:
    app: frontend
  ports:
  - name: http
    port: 80
    targetPort: 80
    nodePort: 30000
  type: LoadBalancer
```

### Accessing the Frontend

After deployment, access the frontend at:
- **Kind Cluster**: http://localhost:3000

## User Interface

```
┌─────────────────────────────────────────────────────────────────┐
│                  Inventory Management System                    │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────┐ ┌───────────┐ ┌──────────┐ ┌─────────────────┐    │
│  │ Products │ │ Inventory │ │  Orders  │ │  Create Order   │    │
│  └──────────┘ └───────────┘ └──────────┘ └─────────────────┘    │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │                                                             ││
│  │                    Content Area                             ││
│  │                                                             ││
│  │   ┌─────────┐  ┌─────────┐  ┌─────────┐                     ││
│  │   │  Card   │  │  Card   │  │  Card   │  ...                ││
│  │   │         │  │         │  │         │                     ││
│  │   └─────────┘  └─────────┘  └─────────┘                     ││
│  │                                                             ││
│  └─────────────────────────────────────────────────────────────┘│
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## JavaScript API Functions

The `script.js` file provides:

### Data Loading
- `loadProducts()` - Fetch and display products
- `loadInventory()` - Fetch and display inventory
- `loadOrders()` - Fetch and display orders

### Display Functions
- `displayProducts(products)` - Render product cards
- `displayInventory(inventory)` - Render inventory cards
- `displayOrders(orders)` - Render order cards

### Order Management
- `handleOrderSubmit(e)` - Create new order
- `fulfillOrder(orderId)` - Fulfill existing order

### Navigation
- `switchSection(sectionId)` - Switch between views
- `setupNavigation()` - Initialize navigation handlers

## Error Handling

The frontend displays user-friendly error messages:
- Network errors during API calls
- Invalid form inputs
- Backend service errors

Messages appear as dismissible notifications.

## Technologies Used

- **HTML5** - Page structure
- **CSS3** - Styling and layout
- **Vanilla JavaScript** - Application logic (no frameworks)
- **NGINX** - Web server and reverse proxy
- **Docker** - Containerization
