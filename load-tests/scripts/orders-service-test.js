import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

// PORT-FORWARD TESTING ONLY
// This test is designed for kubectl port-forward with limited throughput (max 8 req/s).
// Orders service calls other services via gRPC, so we use lower rates.
// For high-load testing, use the in-cluster tests in k8s/high-load/ directory.

const errorRate = new Rate('errors');

export const options = {
  // Use ramping-arrival-rate executor for controlled request rate
  // Lower rate (max 8 req/s) because orders service makes gRPC calls to other services
  scenarios: {
    orders_load_test: {
      executor: 'ramping-arrival-rate',
      startRate: 2,           // Start at 2 requests per second
      timeUnit: '1s',
      preAllocatedVUs: 5,     // Pre-allocate 5 VUs
      maxVUs: 15,             // Allow up to 15 VUs if needed
      stages: [
        { duration: '30s', target: 4 },   // Warm up to 4 req/s
        { duration: '1m', target: 8 },    // Ramp to 8 req/s (port-forward max for orders)
        { duration: '1m', target: 8 },    // Stay at 8 req/s
        { duration: '30s', target: 4 },   // Ramp down to 4 req/s
        { duration: '30s', target: 0 },   // Ramp down to 0
      ],
    },
  },
  
  noConnectionReuse: false,  // Enable connection pooling
  
  thresholds: {
    http_req_duration: ['p(95)<3000'], // 95% under 3s (orders might take longer due to gRPC calls)
    errors: ['rate<0.1'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8003';

export default function () {
  const params = {
    headers: {
      'Connection': 'keep-alive',
    },
    timeout: '30s',
  };
  
  // Get all orders
  let res = http.get(`${BASE_URL}/orders`, params);
  const success = check(res, {
    'orders list status is 200': (r) => r.status === 200,
  });
  errorRate.add(!success);

  sleep(0.1);

  // Create order (tests gRPC communication with Products and Inventory services)
  const orderPayload = JSON.stringify({
    customer_id: Math.floor(Math.random() * 1000) + 1,
    items: [
      { product_id: 1, quantity: 1 },
      { product_id: 2, quantity: 2 },
    ],
  });

  const postParams = {
    headers: {
      'Content-Type': 'application/json',
      'Connection': 'keep-alive',
    },
    timeout: '30s',
  };

  res = http.post(`${BASE_URL}/orders`, orderPayload, postParams);
  const orderCreated = check(res, {
    'create order status is 200 or 201': (r) => r.status === 200 || r.status === 201,
  });

  if (orderCreated) {
    try {
      const body = JSON.parse(res.body);
      const orderId = body.id || body.order_id;
      
      sleep(0.2);  // Small delay before fulfillment

      // Fulfill order
      res = http.post(`${BASE_URL}/orders/${orderId}/fulfill`, null, postParams);
      check(res, {
        'fulfill order successful': (r) => r.status === 200,
      });
    } catch (e) {
      console.log(`Error parsing order response: ${e}`);
    }
  }

  sleep(0.1);
}
