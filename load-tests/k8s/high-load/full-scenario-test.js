import http from 'k6/http';
import { check, group, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// HIGH-LOAD FULL SCENARIO TEST
// Complete user journey at scale with realistic traffic patterns
// Runs multiple scenarios concurrently to simulate real-world usage

const errorRate = new Rate('errors');
const orderCreationTime = new Trend('order_creation_duration');
const ordersCreated = new Counter('orders_created');

export const options = {
  scenarios: {
    // Scenario 1: Browse-heavy users (most common)
    browsers: {
      executor: 'ramping-arrival-rate',
      startRate: 20,
      timeUnit: '1s',
      preAllocatedVUs: 30,
      maxVUs: 100,
      stages: [
        { duration: '1m', target: 40 },
        { duration: '3m', target: 80 },
        { duration: '3m', target: 100 },
        { duration: '2m', target: 60 },
        { duration: '1m', target: 0 },
      ],
      exec: 'browseProducts',
    },
    
    // Scenario 2: Inventory checkers
    inventory_checkers: {
      executor: 'ramping-arrival-rate',
      startRate: 10,
      timeUnit: '1s',
      preAllocatedVUs: 15,
      maxVUs: 50,
      stages: [
        { duration: '1m', target: 20 },
        { duration: '3m', target: 40 },
        { duration: '3m', target: 50 },
        { duration: '2m', target: 30 },
        { duration: '1m', target: 0 },
      ],
      exec: 'checkInventory',
    },
    
    // Scenario 3: Buyers (create orders)
    buyers: {
      executor: 'ramping-arrival-rate',
      startRate: 5,
      timeUnit: '1s',
      preAllocatedVUs: 20,
      maxVUs: 40,
      stages: [
        { duration: '1m', target: 10 },
        { duration: '3m', target: 20 },
        { duration: '3m', target: 30 },
        { duration: '2m', target: 15 },
        { duration: '1m', target: 0 },
      ],
      exec: 'createOrder',
    },
  },
  
  thresholds: {
    http_req_duration: ['p(95)<2000'],
    http_req_failed: ['rate<0.1'],
    errors: ['rate<0.1'],
    order_creation_duration: ['p(95)<3000'],
  },
};

// Service URLs - use internal Kubernetes DNS names
const PRODUCTS_URL = __ENV.PRODUCTS_URL || 'http://products-service:8001';
const INVENTORY_URL = __ENV.INVENTORY_URL || 'http://inventory-service:8002';
const ORDERS_URL = __ENV.ORDERS_URL || 'http://orders-service:8003';

// Scenario: Browse Products
export function browseProducts() {
  group('Browse Products', function () {
    // Get all products
    const res = http.get(`${PRODUCTS_URL}/products`);
    const success = check(res, {
      'products loaded': (r) => r.status === 200,
      'has products': (r) => {
        try {
          const products = JSON.parse(r.body);
          return Array.isArray(products) && products.length > 0;
        } catch (e) {
          return false;
        }
      },
    });
    errorRate.add(!success);

    // Browse individual product details
    if (res.status === 200) {
      try {
        const products = JSON.parse(res.body);
        if (products && products.length > 0) {
          const productId = products[Math.floor(Math.random() * products.length)].id;
          const detailRes = http.get(`${PRODUCTS_URL}/products/${productId}`);
          check(detailRes, {
            'product detail loaded': (r) => r.status === 200,
          });
        }
      } catch (e) {
        // Parsing failed
      }
    }
  });
  
  sleep(0.02);
}

// Scenario: Check Inventory
export function checkInventory() {
  group('Check Inventory', function () {
    // Get all inventory
    const res = http.get(`${INVENTORY_URL}/inventory`);
    const success = check(res, {
      'inventory list loaded': (r) => r.status === 200,
    });
    errorRate.add(!success);

    // Check specific product inventory
    const productId = Math.floor(Math.random() * 5) + 1;
    const detailRes = http.get(`${INVENTORY_URL}/inventory/${productId}`);
    check(detailRes, {
      'inventory detail loaded': (r) => r.status === 200,
      'has quantity info': (r) => {
        try {
          const item = JSON.parse(r.body);
          return item.quantity !== undefined;
        } catch (e) {
          return false;
        }
      },
    });
  });
  
  sleep(0.02);
}

// Scenario: Create Order (full user journey)
export function createOrder() {
  group('Full Purchase Journey', function () {
    // 1. Browse products first
    let res = http.get(`${PRODUCTS_URL}/products`);
    check(res, { 'products loaded': (r) => r.status === 200 });

    // 2. Check inventory
    const productId = Math.floor(Math.random() * 5) + 1;
    res = http.get(`${INVENTORY_URL}/inventory/${productId}`);
    check(res, { 'inventory checked': (r) => r.status === 200 });

    // 3. Create order
    const orderPayload = JSON.stringify({
      customer_id: Math.floor(Math.random() * 10000) + 1,
      items: [
        { product_id: productId, quantity: Math.floor(Math.random() * 3) + 1 },
      ],
    });
    const params = { headers: { 'Content-Type': 'application/json' } };

    const start = new Date();
    res = http.post(`${ORDERS_URL}/orders`, orderPayload, params);
    const duration = new Date() - start;
    orderCreationTime.add(duration);

    const orderCreated = check(res, {
      'order created': (r) => r.status === 200 || r.status === 201,
    });
    
    errorRate.add(!orderCreated);

    if (orderCreated) {
      ordersCreated.add(1);
      
      try {
        const body = JSON.parse(res.body);
        const orderId = body.id || body.order_id;
        
        // 4. Fulfill order
        sleep(0.2);
        const fulfillRes = http.post(`${ORDERS_URL}/orders/${orderId}/fulfill`, null, params);
        check(fulfillRes, {
          'order fulfilled': (r) => r.status === 200,
        });
      } catch (e) {
        // Fulfillment failed
      }
    }
  });
  
  sleep(0.05);
}

// Default function (not used with named scenarios, but required)
export default function () {
  browseProducts();
}
