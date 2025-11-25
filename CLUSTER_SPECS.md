# Especificaciones del Cluster - Sistema de Gestión de Inventario

## Índice

1. [Descripción General del Sistema](#1-descripción-general-del-sistema)
2. [Especificación de Microservicios](#2-especificación-de-microservicios)
3. [Base de Datos](#3-base-de-datos)
4. [Archivos YAML de Configuración](#4-archivos-yaml-de-configuración)
5. [Horizontal Pod Autoscaler (HPA)](#5-horizontal-pod-autoscaler-hpa)
6. [Capacidad de Cómputo del Cluster](#6-capacidad-de-cómputo-del-cluster)
7. [Scripts de Scale Testing con K6](#7-scripts-de-scale-testing-con-k6)
8. [EXTRAS](#8-extras)

---

## 1. Descripción General del Sistema

Este proyecto es un **Sistema de Gestión de Inventario** implementado con una arquitectura de microservicios. Está diseñado para demostrar la comunicación inter-servicios mediante gRPC, orquestación con Kubernetes, y capacidades de escalado automático.

### Diagrama de Arquitectura del Cluster

```
┌─────────────────────────────────────────────────────────────────────────────────────────────┐
│                                     Kind Cluster                                            │
│                                 (inventory-cluster)                                         │
├─────────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                             │
│                              Namespace: inventory-system                                    │
│                                                                                             │
│   Acceso Externo                                                                            │
│   ══════════════                                                                            │
│                                                                                             │
│   ┌─────────────────────────────────────────────────────────────────────────────────────┐   │
│   │                             Frontend (LoadBalancer)                                 │   │
│   │                               localhost:3000                                        │   │
│   │                                                                                     │   │
│   │   ┌─────────────────┐                                                               │   │
│   │   │ Contenedor NGINX│──────────► Proxy Reverso a Servicios Backend                  │   │
│   │   └─────────────────┘                                                               │   │
│   └─────────────────────────────────────────────────────────────────────────────────────┘   │
│                                           │                                                 │
│                                           │ HTTP                                            │
│                                           ▼                                                 │
│   ┌─────────────────────────────────────────────────────────────────────────────────────┐   │
│   │                           Servicios Backend (ClusterIP)                             │   │
│   │                                                                                     │   │
│   │   ┌─────────────────┐   ┌──────────────────┐   ┌─────────────────┐                  │   │
│   │   │Products Service │   │Inventory Service │   │ Orders Service  │                  │   │
│   │   │  HTTP: 8001     │   │   HTTP: 8002     │   │   HTTP: 8003    │                  │   │
│   │   │  gRPC: 9001     │   │   gRPC: 9002     │   │   gRPC: 9003    │                  │   │
│   │   │  HPA: 1-5 pods  │   │   HPA: 1-5 pods  │   │   HPA: 1-5 pods │                  │   │
│   │   └────────┬────────┘   └────────┬─────────┘   └────────┬────────┘                  │   │
│   │            │                     │                       │                          │   │
│   │            │                     │ gRPC                  │                          │   │
│   │            │◄────────────────────┼───────────────────────┘                          │   │
│   │            │                     │                                                  │   │
│   └────────────┼─────────────────────┼──────────────────────────────────────────────────┘   │
│                │                     │                                                      │
│                └──────────┬──────────┘                                                      │
│                           │                                                                 │
│                           ▼                                                                 │
│   ┌─────────────────────────────────────────────────────────────────────────────────────┐   │
│   │                              PostgreSQL (ClusterIP)                                 │   │
│   │                                   Puerto: 5432                                      │   │
│   │                           Volumen Persistente: 1Gi                                  │   │
│   └─────────────────────────────────────────────────────────────────────────────────────┘   │
│                                                                                             │
└─────────────────────────────────────────────────────────────────────────────────────────────┘
```

### Tipos de Services de Kubernetes

| Servicio | Tipo de Service | Exposición | Propósito |
|----------|----------------|------------|-----------|
| **Frontend** | LoadBalancer | Externa (localhost:3000) | Interfaz de usuario accesible desde fuera del cluster |
| **Products Service** | ClusterIP | Interna | Microservicio backend - comunicación interna |
| **Inventory Service** | ClusterIP | Interna | Microservicio backend - comunicación interna |
| **Orders Service** | ClusterIP | Interna | Microservicio backend - comunicación interna |
| **PostgreSQL** | ClusterIP | Interna | Base de datos - comunicación interna |

---

## 2. Especificación de Microservicios

El sistema está compuesto por **3 microservicios backend** escritos en **Go** que se comunican entre sí mediante **gRPC**.

### 2.1 Products Service (Servicio de Productos)

**Descripción:** Responsable de gestionar el catálogo de productos de la aplicación.

**Funcionalidades:**
- Consultar todos los productos del catálogo
- Consultar un producto específico por ID
- Crear nuevos productos
- Proveer información de productos a otros servicios vía gRPC

**Puertos:**
| Protocolo | Puerto | Uso |
|-----------|--------|-----|
| HTTP | 8001 | API REST para comunicación con el frontend |
| gRPC | 9001 | Comunicación inter-servicios |

**Endpoints HTTP:**
```
GET  /health           - Health check del servicio
GET  /products         - Obtener todos los productos
GET  /products/{id}    - Obtener producto por ID
POST /products         - Crear nuevo producto
```

**API gRPC (products.proto):**
```protobuf
service ProductService {
  rpc GetProduct(GetProductRequest) returns (GetProductResponse);
  rpc ListProducts(ListProductsRequest) returns (ListProductsResponse);
  rpc CreateProduct(CreateProductRequest) returns (CreateProductResponse);
}
```

**Recursos asignados:**
- CPU Request: 100m | CPU Limit: 500m
- Memory Request: 128Mi | Memory Limit: 512Mi

---

### 2.2 Inventory Service (Servicio de Inventario)

**Descripción:** Responsable de gestionar los niveles de stock y el sistema de reservaciones para evitar sobre-venta de productos.

**Funcionalidades:**
- Consultar niveles de stock de todos los productos
- Consultar stock de un producto específico
- Actualizar cantidades de stock
- Reservar stock para pedidos pendientes
- Completar reservaciones cuando se cumple un pedido
- Liberar reservaciones cuando se cancela un pedido

**Puertos:**
| Protocolo | Puerto | Uso |
|-----------|--------|-----|
| HTTP | 8002 | API REST para comunicación con el frontend |
| gRPC | 9002 | Comunicación inter-servicios |

**Endpoints HTTP:**
```
GET  /health                            - Health check del servicio
GET  /inventory                         - Obtener todo el inventario
GET  /inventory/{productId}             - Obtener inventario por producto
PUT  /inventory/{productId}             - Actualizar cantidad de stock
POST /inventory/{productId}/reserve     - Reservar stock
POST /inventory/{productId}/fulfill     - Completar reservación
POST /inventory/{productId}/release_reservation - Liberar reservación
```

**API gRPC (inventory.proto):**
```protobuf
service InventoryService {
  rpc GetInventory(GetInventoryRequest) returns (GetInventoryResponse);
  rpc ListInventory(ListInventoryRequest) returns (ListInventoryResponse);
  rpc UpdateStock(UpdateStockRequest) returns (UpdateStockResponse);
  rpc ReserveStock(ReserveStockRequest) returns (ReserveStockResponse);
  rpc FulfillReservation(FulfillReservationRequest) returns (FulfillReservationResponse);
  rpc ReleaseReservation(ReleaseReservationRequest) returns (ReleaseReservationResponse);
}
```

**Lógica de Stock:**
```
Stock Disponible = Stock Total - Stock Reservado
```

**Recursos asignados:**
- CPU Request: 100m | CPU Limit: 500m
- Memory Request: 128Mi | Memory Limit: 512Mi

---

### 2.3 Orders Service (Servicio de Pedidos)

**Descripción:** Responsable de gestionar el ciclo de vida de los pedidos, coordinando con los servicios de Products e Inventory mediante gRPC.

**Funcionalidades:**
- Consultar todos los pedidos
- Consultar un pedido específico por ID
- Crear nuevos pedidos (con validación de productos y reservación de stock)
- Completar pedidos (fulfill)
- Coordinar con Products Service para validar productos y obtener precios
- Coordinar con Inventory Service para reservar y completar stock

**Puertos:**
| Protocolo | Puerto | Uso |
|-----------|--------|-----|
| HTTP | 8003 | API REST para comunicación con el frontend |
| gRPC | 9003 | Comunicación inter-servicios |

**Endpoints HTTP:**
```
GET  /health              - Health check del servicio
GET  /orders              - Obtener todos los pedidos
GET  /orders/{id}         - Obtener pedido por ID
POST /orders              - Crear nuevo pedido
POST /orders/{id}/fulfill - Completar un pedido
```

**API gRPC (orders.proto):**
```protobuf
service OrderService {
  rpc GetOrder(GetOrderRequest) returns (GetOrderResponse);
  rpc ListOrders(ListOrdersRequest) returns (ListOrdersResponse);
  rpc CreateOrder(CreateOrderRequest) returns (CreateOrderResponse);
  rpc FulfillOrder(FulfillOrderRequest) returns (FulfillOrderResponse);
}
```

**Flujo de Creación de Pedido:**
```
1. Cliente envía solicitud de pedido
       │
       ▼
2. Orders Service valida productos ──────► Products Service (gRPC)
       │                                    - Obtiene detalles del producto
       │                                    - Obtiene precios
       ▼
3. Orders Service reserva stock ──────────► Inventory Service (gRPC)
       │                                    - Reserva items
       │                                    - Verifica disponibilidad
       ▼
4. Orders Service crea el pedido ─────────► PostgreSQL
       │                                    - Guarda pedido
       │                                    - Guarda items del pedido
       ▼
5. Retorna pedido (estado: pending)
```

**Recursos asignados:**
- CPU Request: 100m | CPU Limit: 500m
- Memory Request: 128Mi | Memory Limit: 512Mi

---

### Comunicación Inter-Servicios

```
┌──────────────────────────────────────────────────────────────────────────────────┐
│                        Flujo de Comunicación gRPC                                │
├──────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│                                  gRPC                                            │
│   ┌──────────────┐  GetProduct()   ┌──────────────────┐                          │
│   │   Products   │◄────────────────│                  │                          │
│   │   Service    │                 │                  │                          │
│   │  (:9001)     │                 │                  │                          │
│   └──────────────┘                 │  Orders Service  │                          │
│                                    │     (:9003)      │                          │
│   ┌──────────────┐  ReserveStock() │                  │                          │
│   │  Inventory   │◄────────────────│                  │                          │
│   │   Service    │  FulfillRes()   │                  │                          │
│   │  (:9002)     │  ReleaseRes()   │                  │                          │
│   └──────────────┘                 └──────────────────┘                          │
│                                                                                  │
└──────────────────────────────────────────────────────────────────────────────────┘
```

**Descubrimiento de Servicios:**
Los servicios utilizan el DNS de Kubernetes para el descubrimiento de servicios:
- `products-service:9001` - gRPC del servicio de productos
- `inventory-service:9002` - gRPC del servicio de inventario
- `orders-service:9003` - gRPC del servicio de pedidos

---

## 3. Base de Datos

### Tipo de Base de Datos

**PostgreSQL 16** - Base de datos relacional que proporciona:
- Soporte ACID para transacciones
- Integridad referencial con foreign keys
- Alto rendimiento para operaciones CRUD
- Almacenamiento persistente mediante PersistentVolumeClaim

### Service de Kubernetes

| Configuración | Valor |
|---------------|-------|
| Tipo de Service | ClusterIP |
| Puerto | 5432 |
| Nombre DNS | postgres.inventory-system.svc.cluster.local |
| Almacenamiento | PersistentVolumeClaim de 1Gi |

### Schema de la Base de Datos

```sql
-- Tabla de Productos
CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL,
    category VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabla de Inventario
CREATE TABLE IF NOT EXISTS inventory (
    product_id INTEGER PRIMARY KEY REFERENCES products(id) ON DELETE CASCADE,
    stock INTEGER NOT NULL DEFAULT 0,
    reserved INTEGER NOT NULL DEFAULT 0,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabla de Pedidos
CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    customer_id INTEGER NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    total_amount DECIMAL(10, 2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabla de Items de Pedido
CREATE TABLE IF NOT EXISTS order_items (
    id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL REFERENCES products(id),
    quantity INTEGER NOT NULL,
    price_at_order DECIMAL(10, 2) NOT NULL
);
```

### Diagrama Entidad-Relación

```
┌──────────────────┐       ┌──────────────────┐
│     products     │       │    inventory     │
├──────────────────┤       ├──────────────────┤
│ id (PK)          │◄──────│ product_id (PK)  │
│ name             │   1:1 │ stock            │
│ description      │       │ reserved         │
│ price            │       │ updated_at       │
│ category         │       └──────────────────┘
│ created_at       │
└────────┬─────────┘
         │
         │ 1:N
         ▼
┌──────────────────┐       ┌──────────────────┐
│   order_items    │       │     orders       │
├──────────────────┤       ├──────────────────┤
│ id (PK)          │       │ id (PK)          │
│ order_id (FK)    │◄──────│ customer_id      │
│ product_id (FK)  │   N:1 │ status           │
│ quantity         │       │ total_amount     │
│ price_at_order   │       │ created_at       │
└──────────────────┘       └──────────────────┘
```

### Datos de Muestra Iniciales

**Productos:**
| ID | Nombre | Precio | Categoría |
|----|--------|--------|-----------|
| 1 | Laptop | $999.99 | Electronics |
| 2 | Mouse | $29.99 | Electronics |
| 3 | Keyboard | $79.99 | Electronics |
| 4 | Monitor | $199.99 | Electronics |
| 5 | Desk Chair | $149.99 | Furniture |

**Inventario Inicial:**
| Product ID | Stock | Reservado |
|------------|-------|-----------|
| 1 | 50 | 0 |
| 2 | 100 | 0 |
| 3 | 25 | 0 |
| 4 | 30 | 0 |
| 5 | 15 | 0 |

---

## 4. Archivos YAML de Configuración

### 4.1 kind-config.yaml

**Propósito:** Configuración del cluster Kind (Kubernetes in Docker) que define los nodos y mapeos de puertos.

```yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: inventory-cluster
nodes:
- role: control-plane
  extraPortMappings:
  # Frontend expuesto externamente vía LoadBalancer
  - containerPort: 30000
    hostPort: 3000
    protocol: TCP
  # Grafana para dashboards de pruebas de carga
  - containerPort: 30300
    hostPort: 3001
    protocol: TCP
```

**Función:** 
- Define un cluster de un nodo (control-plane)
- Mapea el puerto 30000 del contenedor al puerto 3000 del host (Frontend)
- Mapea el puerto 30300 del contenedor al puerto 3001 del host (Grafana)

---

### 4.2 namespace.yaml

**Propósito:** Crear el namespace donde se despliegan todos los recursos.

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: inventory-system
```

**Función:** Aísla todos los recursos del sistema en un namespace dedicado para mejor organización y gestión.

---

### 4.3 postgres-deployment.yaml

**Propósito:** Desplegar PostgreSQL con almacenamiento persistente.

**Contenido:**
1. **PersistentVolumeClaim (postgres-pvc):**
   - Solicita 1Gi de almacenamiento
   - Modo de acceso: ReadWriteOnce

2. **Deployment (postgres):**
   - Imagen: postgres:16
   - Puerto: 5432
   - Variables de entorno desde ConfigMap y Secret
   - Health checks con `pg_isready`
   - Volumen persistente montado

3. **Service (postgres):**
   - Tipo: ClusterIP
   - Puerto: 5432
   - Acceso interno únicamente

**Función:** Proporciona la capa de persistencia de datos para todos los microservicios.

---

### 4.4 postgres-init-job.yaml

**Propósito:** Job de Kubernetes que inicializa el esquema de la base de datos.

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: postgres-init
  namespace: inventory-system
```

**Función:**
- Espera a que PostgreSQL esté listo
- Ejecuta el script SQL de inicialización
- Crea tablas y datos de muestra
- Se ejecuta una sola vez y termina

---

### 4.5 postgres-config-generated.yaml.template

**Propósito:** Template para generar ConfigMap y Secret con credenciales de la base de datos.

**Contenido:**
1. **ConfigMap (postgres-config):**
   - DB_HOST: postgres
   - DB_PORT: 5432
   - DB_NAME: nombre de la base de datos
   - DB_USER: usuario de la base de datos

2. **Secret (postgres-secret):**
   - DB_PASSWORD: contraseña (almacenada de forma segura)

**Función:** Centraliza la configuración de conexión a la base de datos que usan todos los microservicios.

---

### 4.6 products-deployment.yaml

**Propósito:** Desplegar el microservicio de productos.

**Contenido:**
1. **Deployment (products-service):**
   - Imagen: products-service:latest
   - Puertos: 8001 (HTTP), 9001 (gRPC)
   - Recursos: 100m-500m CPU, 128Mi-512Mi memoria
   - Variables de entorno para conexión a DB
   - Liveness y Readiness probes en /health

2. **Service (products-service):**
   - Tipo: ClusterIP
   - Puertos: 8001 (HTTP), 9001 (gRPC)

3. **HorizontalPodAutoscaler:**
   - Min: 1 réplica, Max: 5 réplicas
   - Target CPU: 70%

**Función:** Gestiona el catálogo de productos y expone APIs HTTP y gRPC.

---

### 4.7 inventory-deployment.yaml

**Propósito:** Desplegar el microservicio de inventario.

**Contenido:**
1. **Deployment (inventory-service):**
   - Imagen: inventory-service:latest
   - Puertos: 8002 (HTTP), 9002 (gRPC)
   - Recursos: 100m-500m CPU, 128Mi-512Mi memoria
   - Variables de entorno para conexión a DB
   - Liveness y Readiness probes en /health

2. **Service (inventory-service):**
   - Tipo: ClusterIP
   - Puertos: 8002 (HTTP), 9002 (gRPC)

**Función:** Gestiona niveles de stock y reservaciones de inventario.

---

### 4.8 orders-deployment.yaml

**Propósito:** Desplegar el microservicio de pedidos.

**Contenido:**
1. **Deployment (orders-service):**
   - Imagen: orders-service:latest
   - Puertos: 8003 (HTTP), 9003 (gRPC)
   - Recursos: 100m-500m CPU, 128Mi-512Mi memoria
   - Variables de entorno para conexión a DB
   - Variables de entorno para conexión gRPC a otros servicios:
     - PRODUCTS_GRPC_ADDR: products-service:9001
     - INVENTORY_GRPC_ADDR: inventory-service:9002
   - Liveness y Readiness probes en /health

2. **Service (orders-service):**
   - Tipo: ClusterIP
   - Puertos: 8003 (HTTP), 9003 (gRPC)

**Función:** Orquesta la creación y gestión de pedidos coordinando con Products e Inventory.

---

### 4.9 frontend-deployment.yaml

**Propósito:** Desplegar la aplicación web frontend.

**Contenido:**
1. **Deployment (frontend):**
   - Imagen: frontend:latest (NGINX)
   - Puerto: 80
   - NGINX actúa como servidor web y proxy reverso

2. **Service (frontend):**
   - **Tipo: LoadBalancer** (exposición externa)
   - Puerto: 80
   - NodePort: 30000
   - Accesible en localhost:3000

**Función:** Proporciona la interfaz de usuario y enruta peticiones a los microservicios backend.

---

### 4.10 hpa.yaml

**Propósito:** Configurar Horizontal Pod Autoscalers para los tres microservicios backend.

**Contenido (por cada servicio):**
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: <service-name>
  minReplicas: 1
  maxReplicas: 5
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  behavior:
    scaleUp:
      stabilizationWindowSeconds: 30
      policies:
      - type: Percent
        value: 100
        periodSeconds: 30
    scaleDown:
      stabilizationWindowSeconds: 60
      policies:
      - type: Percent
        value: 50
        periodSeconds: 30
```

**Función:** Escala automáticamente los pods basándose en la utilización de CPU.

---

## 5. Horizontal Pod Autoscaler (HPA)

### Configuración del HPA

| Parámetro | Valor | Descripción |
|-----------|-------|-------------|
| **minReplicas** | 1 | Mínimo número de pods |
| **maxReplicas** | 5 | Máximo número de pods |
| **Target CPU** | 70% | Umbral de CPU para escalar |
| **Ventana de estabilización (scale up)** | 30 segundos | Tiempo antes de escalar hacia arriba |
| **Ventana de estabilización (scale down)** | 60 segundos | Tiempo antes de escalar hacia abajo |

### Comportamiento de Escalado

**Scale Up (Escalar hacia arriba):**
- Se activa cuando CPU promedio > 70%
- Puede duplicar pods cada 30 segundos
- Ventana de estabilización: 30 segundos

**Scale Down (Escalar hacia abajo):**
- Se activa cuando CPU promedio < 70%
- Reduce pods en 50% cada 30 segundos
- Ventana de estabilización: 60 segundos (previene fluctuaciones)

### Visualización del Escalado

```
        ┌────────────────────────────────────────────────────────────────┐
   Load │                         ╱╲    ╱╲                               │
        │                        ╱  ╲  ╱  ╲                              │
        │       ╱╲              ╱    ╲╱    ╲                             │
        │      ╱  ╲            ╱            ╲                            │
        │─────╱────╲──────────╱──────────────╲─────────────────────────  │
        └────────────────────────────────────────────────────────────────┘
                                     Tiempo

        ┌────────────────────────────────────────────────────────────────┐
   Pods │               ┌──────────────────┐                             │
    5   │               │                  │                             │
    4   │            ┌──┘                  └──┐                          │
    3   │         ┌──┘                        └──┐                       │
    2   │      ┌──┘                              └──┐                    │
    1   │──────┘                                    └────────────────────│
        └────────────────────────────────────────────────────────────────┘
                                     Tiempo
```

### Recursos del Metrics Server

Para que el HPA funcione, se despliega un **metrics-server** que recopila métricas de CPU y memoria de los pods.

---

## 6. Capacidad de Cómputo del Cluster

### Recursos por Servicio

| Servicio | CPU Request | CPU Limit | Memory Request | Memory Limit |
|----------|-------------|-----------|----------------|--------------|
| products-service | 100m | 500m | 128Mi | 512Mi |
| inventory-service | 100m | 500m | 128Mi | 512Mi |
| orders-service | 100m | 500m | 128Mi | 512Mi |
| PostgreSQL | - | - | - | - |
| Frontend | - | - | - | - |

### Capacidad Máxima del Sistema

Con la configuración de HPA (1-5 réplicas por servicio):

| Escenario | Products | Inventory | Orders | Total Pods | CPU Total |
|-----------|----------|-----------|--------|------------|-----------|
| **Mínimo** | 1 | 1 | 1 | 3 | 300m - 1500m |
| **Máximo** | 5 | 5 | 5 | 15 | 1500m - 7500m |

### Carga de Trabajo Soportada

Basado en las pruebas de carga y la configuración del HPA:

| Enfoque de Prueba | Throughput | Usuarios Virtuales | Descripción |
|-------------------|------------|-------------------|-------------|
| **Port-Forward** | ~15-20 req/s | 2-5 VUs | Pruebas rápidas, debugging |
| **In-Cluster** | 100+ req/s | 10-50 VUs | Pruebas de carga alta |
| **Con Grafana** | 100+ req/s | 10-50 VUs | Pruebas con visualización |

### Estimación de Usuarios Concurrentes

Considerando los resultados de las pruebas de carga:

| Configuración | Usuarios Paralelos | Requests/segundo | Notas |
|---------------|-------------------|------------------|-------|
| **1 réplica por servicio** | 5-10 usuarios | 15-30 req/s | Carga mínima |
| **3 réplicas por servicio** | 20-30 usuarios | 60-90 req/s | Carga moderada |
| **5 réplicas por servicio** | 40-50 usuarios | 100-150 req/s | Carga alta |

**Nota:** El HPA escalará automáticamente cuando la CPU supere el 70%, permitiendo manejar picos de carga sin intervención manual.

---

## 7. Scripts de Scale Testing con K6

### Ubicación de Scripts

Los scripts de pruebas de carga se encuentran en el directorio `load-tests/`.

### Enfoques de Pruebas

| Enfoque | Mejor Para | Throughput Máximo | Complejidad |
|---------|------------|-------------------|-------------|
| **Port-Forward** | Pruebas rápidas, debugging | ~15-20 req/s | Baja |
| **In-Cluster** | Pruebas de alta carga, HPA | 100+ req/s | Media |
| **Grafana** | Monitoreo en tiempo real | 100+ req/s | Media |

### Scripts Disponibles

#### Scripts para Port-Forward (load-tests/scripts/)

| Script | Tasa | Duración | Propósito |
|--------|------|----------|-----------|
| smoke-test.js | 5 req/s | 30s | Verificación rápida |
| products-service-test.js | 5-15 req/s | 3.5m | Prueba de Products API |
| inventory-service-test.js | 3-10 req/s | 3.5m | Prueba de Inventory API |
| orders-service-test.js | 2-8 req/s | 3.5m | Prueba de Orders API |
| full-scenario-test.js | 5-30 req/s | 8m | Prueba de flujo completo |

#### Scripts para In-Cluster (load-tests/k8s/)

| Script | Tasa | Propósito |
|--------|------|-----------|
| smoke-test.js | 10 req/s | Verificación in-cluster |
| products-service-test.js | 10-50 req/s | Carga alta en Products |
| inventory-service-test.js | 10-40 req/s | Carga alta en Inventory |
| orders-service-test.js | 5-20 req/s | Carga alta en Orders |
| full-scenario-test.js | 5-30 req/s | Flujo completo |

### Ejemplo de Script de Prueba (smoke-test.js)

```javascript
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  scenarios: {
    smoke_test: {
      executor: 'constant-arrival-rate',
      rate: 5,              // 5 requests por segundo
      timeUnit: '1s',
      duration: '30s',
      preAllocatedVUs: 2,   // 2 VUs pre-asignados
      maxVUs: 5,            // máximo 5 VUs
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.1'],       // <10% de errores
    http_req_duration: ['p(95)<2000'],   // 95% bajo 2s
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8001';

export default function () {
  const res = http.get(`${BASE_URL}/products`);
  
  check(res, {
    'status is 200': (r) => r.status === 200,
    'response has body': (r) => r.body && r.body.length > 0,
  });
  
  sleep(0.1);
}
```

### Ejecución de Pruebas

**Port-Forward:**
```bash
cd load-tests
./setup-port-forwarding.sh
./run-test.sh scripts/smoke-test.js
```

**In-Cluster:**
```bash
cd load-tests
./run-in-cluster.sh run smoke
./run-in-cluster.sh run-high products
```

**Con Grafana:**
```bash
cd load-tests
./deploy-grafana-k6.sh deploy
./deploy-grafana-k6.sh run smoke
# Acceder a http://localhost:3001 para dashboard
```

### Configuraciones de Alta Carga

| Prueba | Tasa Objetivo | Duración | Comportamiento Esperado |
|--------|---------------|----------|------------------------|
| products-high | 100 req/s | 10 min | Activa HPA a ~60 req/s |
| inventory-high | 80 req/s | 10 min | Activa HPA a ~50 req/s |
| orders-high | 50 req/s | 10 min | Menor throughput debido a latencia adicional de llamadas gRPC a Products e Inventory |
| full-high | 180 req/s combinado | 10 min | Carga realista multi-escenario |

**Nota sobre latencia de Orders Service:** El servicio de Orders tiene menor throughput porque cada operación de creación de pedido requiere:
1. Llamada gRPC a Products Service para validar productos y obtener precios
2. Llamada gRPC a Inventory Service para reservar stock
Esta latencia adicional es esperada en arquitecturas de microservicios con orquestación.

### Monitoreo de HPA Durante Pruebas

```bash
# Ver estado del HPA
kubectl get hpa -n inventory-system

# Monitorear en tiempo real
watch -n 2 'kubectl get hpa,pods -n inventory-system'

# Detalles del HPA
kubectl describe hpa products-service-hpa -n inventory-system
```

---

## 8. EXTRAS

### 8.1 Frontend como Entrada de Datos

El sistema incluye una **aplicación web frontend** construida con HTML5, CSS3 y JavaScript vanilla.

#### Arquitectura del Frontend

```
┌─────────────────────────────────────────────────────────────────┐
│                         Frontend                                │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │                    NGINX Web Server                        │ │
│  │                       (Puerto 80)                          │ │
│  └───────────────────────────┬────────────────────────────────┘ │
│                              │                                  │
│     ┌────────────────────────┼────────────────────────┐         │
│     ▼                        ▼                        ▼         │
│  ┌─────────┐          ┌──────────┐            ┌──────────┐      │
│  │  HTML   │          │   CSS    │            │    JS    │      │
│  │ (Vistas)│          │(Estilos) │            │ (Lógica) │      │
│  └─────────┘          └──────────┘            └──────────┘      │
│                                                                 │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │              NGINX Proxy Reverso                           │ │
│  │                                                            │ │
│  │   /api/products/*  ────►  products-service:8001            │ │
│  │   /api/inventory/* ────►  inventory-service:8002           │ │
│  │   /api/orders/*    ────►  orders-service:8003              │ │
│  └────────────────────────────────────────────────────────────┘ │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

#### Funcionalidades del Frontend

1. **Vista de Productos:**
   - Muestra todos los productos del catálogo
   - Información: nombre, descripción, categoría, precio

2. **Vista de Inventario:**
   - Muestra niveles de stock actuales
   - Cantidades disponibles y reservadas
   - Cálculo de stock disponible (total - reservado)

3. **Vista de Pedidos:**
   - Lista todos los pedidos con su estado
   - Detalles: cliente, items, total, estado
   - Botón para completar pedidos pendientes

4. **Crear Pedido:**
   - Selección de múltiples productos
   - Especificación de cantidades
   - Envío con reservación automática de stock

#### Acceso

- **URL:** http://localhost:3000
- **Tipo de Service:** LoadBalancer
- **NodePort:** 30000

---

### 8.2 Grafana y Prometheus para Observabilidad

El sistema incluye infraestructura de **observabilidad** para monitoreo de pruebas de carga.

#### Arquitectura de Observabilidad

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Kubernetes Cluster                          │
│  ┌─────────────┐     ┌─────────────┐     ┌─────────────────────────┐│
│  │   K6 Job    │───▶│  InfluxDB   │◀───│        Grafana          ││
│  │ (Load Test) │     │  (Métricas) │     │ (Dashboard @ :3001)     ││
│  └─────────────┘     └─────────────┘     └─────────────────────────┘│
│         │                                          ▲                │
│         ▼                                          │                │
│  ┌─────────────────────────────────────────┐       │                │
│  │        Servicios de Aplicación          │       │                │
│  │  (Products, Inventory, Orders)          │       │                │
│  └─────────────────────────────────────────┘       │                │
└────────────────────────────────────────────────────│────────────────┘
                                                     │
                                              Acceso Browser
                                           http://localhost:3001
```

#### Componentes

| Componente | Tipo de Service | Puerto | Propósito |
|------------|----------------|--------|-----------|
| **InfluxDB** | ClusterIP | 8086 | Almacenamiento de métricas de series temporales |
| **Grafana** | LoadBalancer | 3001 | Visualización de dashboards |

#### Dashboard de K6

El dashboard auto-provisionado incluye:

1. **HTTP Requests per Second:** Tasa de peticiones en tiempo real
2. **HTTP Request Duration:** Percentiles de tiempo de respuesta (p50, p90, p95, p99)
3. **Virtual Users:** Conteo de VUs activos durante la prueba
4. **Error Rate:** Porcentaje de fallos de peticiones HTTP
5. **Data Transfer:** Bytes enviados y recibidos por segundo
6. **Iterations:** Iteraciones de prueba completadas por segundo
7. **Checks Pass Rate:** Tasa de éxito de validaciones K6

#### Configuración de Grafana

**Acceso:**
- URL: http://localhost:3001
- Credenciales: admin / admin
- Acceso anónimo habilitado con rol Admin

**Características:**
- Auto-refresh cada 5 segundos durante pruebas
- Selección de rango de tiempo
- Análisis de percentiles
- Tracking de errores
- Consultas personalizadas con InfluxDB

#### Archivos de Configuración

| Archivo | Propósito |
|---------|-----------|
| `grafana-deployment.yaml` | Deployment y Service LoadBalancer |
| `influxdb-deployment.yaml` | Deployment y Service ClusterIP |
| `grafana-configmaps.yaml` | Datasources, dashboard provider, y dashboard K6 |
| `k6-grafana-job.yaml` | Templates de Jobs K6 con output a InfluxDB |

#### Uso

```bash
# Desplegar infraestructura
./deploy-grafana-k6.sh deploy

# Ejecutar prueba con visualización
./deploy-grafana-k6.sh run smoke

# Verificar estado
./deploy-grafana-k6.sh status

# Limpiar
./deploy-grafana-k6.sh delete
```

---

## Resumen de Arquitectura

```
┌─────────────────────────────────────────────────────────────────────────────────────────────┐
│                                   RESUMEN DEL SISTEMA                                       │
├─────────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                             │
│  MICROSERVICIOS (3):                                                                        │
│  ┌─────────────────────┐  ┌─────────────────────┐  ┌─────────────────────┐                  │
│  │  Products Service   │  │ Inventory Service   │  │   Orders Service    │                  │
│  │  (Go + gRPC + HTTP) │  │ (Go + gRPC + HTTP)  │  │ (Go + gRPC + HTTP)  │                  │
│  │  Service: ClusterIP │  │  Service: ClusterIP │  │  Service: ClusterIP │                  │
│  │  HPA: 1-5 pods      │  │  HPA: 1-5 pods      │  │  HPA: 1-5 pods      │                  │
│  └─────────────────────┘  └─────────────────────┘  └─────────────────────┘                  │
│                                                                                             │
│  BASE DE DATOS:                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────────────────────┐    │
│  │  PostgreSQL 16                                                                      │    │
│  │  Service: ClusterIP                                                                 │    │
│  │  Tablas: products, inventory, orders, order_items                                   │    │
│  │  Almacenamiento: PersistentVolumeClaim 1Gi                                          │    │
│  └─────────────────────────────────────────────────────────────────────────────────────┘    │
│                                                                                             │
│  COMUNICACIÓN:                                                                              │
│  • Inter-servicios: gRPC (puertos 9001-9003)                                                │
│  • Frontend → Backend: HTTP (puertos 8001-8003)                                             │
│  • Descubrimiento: Kubernetes DNS                                                           │
│                                                                                             │
│  ESCALADO:                                                                                  │
│  • HPA en los 3 microservicios                                                              │
│  • Min: 1 pod, Max: 5 pods                                                                  │
│  • Target CPU: 70%                                                                          │
│  • Capacidad: 40-50 usuarios concurrentes @ 100-150 req/s (máximo)                          │
│                                                                                             │
│  EXTRAS:                                                                                    │
│  • Frontend (LoadBalancer @ localhost:3000)                                                 │
│  • Grafana + InfluxDB para observabilidad (LoadBalancer @ localhost:3001)                   │
│  • Scripts de pruebas de carga K6                                                           │
│                                                                                             │
└─────────────────────────────────────────────────────────────────────────────────────────────┘
```

---

*Documento generado para el proyecto de Sistema de Gestión de Inventario con microservicios en Kubernetes.*
