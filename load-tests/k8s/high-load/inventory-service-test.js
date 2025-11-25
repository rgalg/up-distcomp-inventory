import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

// HIGH-LOAD INVENTORY SERVICE TEST
// Designed for in-cluster testing without port-forward limitations
// Targets 50-80 req/s with extended duration

const errorRate = new Rate('errors');

export const options = {
  scenarios: {
    high_load_inventory: {
      executor: 'ramping-arrival-rate',
      startRate: 10,          // Start at 10 requests per second
      timeUnit: '1s',
      preAllocatedVUs: 15,    // Pre-allocate 15 VUs
      maxVUs: 80,             // Allow up to 80 VUs for peak load
      stages: [
        { duration: '1m', target: 25 },    // Warm up to 25 req/s
        { duration: '2m', target: 50 },    // Ramp to 50 req/s
        { duration: '3m', target: 80 },    // Ramp to 80 req/s (should trigger HPA)
        { duration: '2m', target: 80 },    // Stay at 80 req/s
        { duration: '1m', target: 40 },    // Ramp down to 40 req/s
        { duration: '1m', target: 0 },     // Ramp down to 0
      ],
    },
  },
  
  thresholds: {
    http_req_duration: ['p(95)<500'],   // 95% of requests under 500ms
    http_req_failed: ['rate<0.05'],     // Less than 5% failed requests
    errors: ['rate<0.05'],
  },
};

// Use environment variable or default to internal Kubernetes service URL
const BASE_URL = __ENV.BASE_URL || 'http://inventory-service:8002';

export default function () {
  // Get all inventory items
  let res = http.get(`${BASE_URL}/inventory`);
  const listSuccess = check(res, {
    'inventory list status is 200': (r) => r.status === 200,
    'inventory list has items': (r) => {
      try {
        const items = JSON.parse(r.body);
        return Array.isArray(items) && items.length > 0;
      } catch (e) {
        return false;
      }
    },
  });
  errorRate.add(!listSuccess);

  // Get inventory for a random product (1-5)
  const productId = Math.floor(Math.random() * 5) + 1;
  res = http.get(`${BASE_URL}/inventory/${productId}`);
  const detailSuccess = check(res, {
    'inventory detail status is 200': (r) => r.status === 200,
    'inventory has quantity': (r) => {
      try {
        const item = JSON.parse(r.body);
        return item.quantity !== undefined;
      } catch (e) {
        return false;
      }
    },
  });
  errorRate.add(!detailSuccess);

  // Reserve and fulfill inventory (less frequently to avoid depleting stock)
  if (__ITER % 3 === 0) {
    const quantity = Math.floor(Math.random() * 3) + 1;
    const reservePayload = JSON.stringify({ quantity: quantity });
    const params = { headers: { 'Content-Type': 'application/json' } };

    // Reserve inventory
    res = http.post(`${BASE_URL}/inventory/${productId}/reserve`, reservePayload, params);
    const reserveSuccess = check(res, {
      'reserve inventory successful': (r) => r.status === 200,
    });
    errorRate.add(!reserveSuccess);

    // Fulfill reservation
    res = http.post(`${BASE_URL}/inventory/${productId}/fulfill`, reservePayload, params);
    check(res, {
      'fulfill inventory successful': (r) => r.status === 200,
    });
  }

  // Minimal sleep for high throughput
  sleep(0.01);
}
