import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

// HIGH-LOAD ORDERS SERVICE TEST
// Designed for in-cluster testing without port-forward limitations
// Moderate load (30-50 req/s) because orders service makes gRPC calls

const errorRate = new Rate('errors');
const orderCreationTime = new Trend('order_creation_duration');

export const options = {
  scenarios: {
    high_load_orders: {
      executor: 'ramping-arrival-rate',
      startRate: 5,           // Start at 5 requests per second
      timeUnit: '1s',
      preAllocatedVUs: 15,    // Pre-allocate 15 VUs
      maxVUs: 60,             // Allow up to 60 VUs for peak load
      stages: [
        { duration: '1m', target: 15 },    // Warm up to 15 req/s
        { duration: '2m', target: 30 },    // Ramp to 30 req/s
        { duration: '3m', target: 50 },    // Ramp to 50 req/s (should trigger HPA)
        { duration: '2m', target: 50 },    // Stay at 50 req/s
        { duration: '1m', target: 25 },    // Ramp down to 25 req/s
        { duration: '1m', target: 0 },     // Ramp down to 0
      ],
    },
  },
  
  thresholds: {
    http_req_duration: ['p(95)<1000'],  // 95% of requests under 1s (orders are slower due to gRPC)
    http_req_failed: ['rate<0.05'],     // Less than 5% failed requests
    errors: ['rate<0.05'],
    order_creation_duration: ['p(95)<2000'], // 95% of order creations under 2s
  },
};

// Use environment variable or default to internal Kubernetes service URL
const BASE_URL = __ENV.BASE_URL || 'http://orders-service:8003';

export default function () {
  // Get all orders
  let res = http.get(`${BASE_URL}/orders`);
  const listSuccess = check(res, {
    'orders list status is 200': (r) => r.status === 200,
  });
  errorRate.add(!listSuccess);

  // Create a new order (tests gRPC communication with Products and Inventory services)
  const orderPayload = JSON.stringify({
    customer_id: Math.floor(Math.random() * 10000) + 1,
    items: [
      { product_id: Math.floor(Math.random() * 5) + 1, quantity: Math.floor(Math.random() * 3) + 1 },
      { product_id: Math.floor(Math.random() * 5) + 1, quantity: Math.floor(Math.random() * 2) + 1 },
    ],
  });

  const params = { headers: { 'Content-Type': 'application/json' } };

  const start = new Date();
  res = http.post(`${BASE_URL}/orders`, orderPayload, params);
  const duration = new Date() - start;
  orderCreationTime.add(duration);
  
  const orderCreated = check(res, {
    'create order status is 200 or 201': (r) => r.status === 200 || r.status === 201,
    'order has id': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.id !== undefined || body.order_id !== undefined;
      } catch (e) {
        return false;
      }
    },
  });

  errorRate.add(!orderCreated);

  if (orderCreated && res.body) {
    try {
      const body = JSON.parse(res.body);
      const orderId = body.id || body.order_id;
      
      // Small delay before fulfillment
      sleep(0.1);

      // Fulfill the order
      res = http.post(`${BASE_URL}/orders/${orderId}/fulfill`, null, params);
      const fulfillSuccess = check(res, {
        'fulfill order successful': (r) => r.status === 200,
      });
      errorRate.add(!fulfillSuccess);
    } catch (e) {
      errorRate.add(true);
    }
  }

  // Minimal sleep for high throughput
  sleep(0.02);
}
