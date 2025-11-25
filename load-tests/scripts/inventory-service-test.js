import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

// PORT-FORWARD TESTING ONLY
// This test is designed for kubectl port-forward with limited throughput (max 10 req/s).
// For high-load testing, use the in-cluster tests in k8s/high-load/ directory.

const errorRate = new Rate('errors');

export const options = {
  // Use ramping-arrival-rate executor for controlled request rate
  // Port-forward friendly with max 10 req/s
  scenarios: {
    inventory_load_test: {
      executor: 'ramping-arrival-rate',
      startRate: 3,           // Start at 3 requests per second
      timeUnit: '1s',
      preAllocatedVUs: 5,     // Pre-allocate 5 VUs
      maxVUs: 15,             // Allow up to 15 VUs if needed
      stages: [
        { duration: '30s', target: 5 },   // Warm up to 5 req/s
        { duration: '1m', target: 10 },   // Ramp to 10 req/s (port-forward max)
        { duration: '1m', target: 10 },   // Stay at 10 req/s
        { duration: '30s', target: 5 },   // Ramp down to 5 req/s
        { duration: '30s', target: 0 },   // Ramp down to 0
      ],
    },
  },
  
  noConnectionReuse: false,  // Enable connection pooling
  
  thresholds: {
    http_req_duration: ['p(95)<2000'],  // 95% under 2s (lenient for port-forward)
    errors: ['rate<0.1'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8002';

export default function () {
  const params = {
    headers: {
      'Connection': 'keep-alive',
    },
    timeout: '30s',
  };
  
  // Get all inventory items
  let res = http.get(`${BASE_URL}/inventory`, params);
  const success = check(res, {
    'inventory list status is 200': (r) => r.status === 200,
  });
  errorRate.add(!success);

  sleep(0.1);

  // Get inventory data for a specific product
  res = http.get(`${BASE_URL}/inventory/1`, params);
  check(res, {
    'inventory detail status is 200': (r) => r.status === 200,
  });

  sleep(0.1);

  // Reserve inventory for a specific product
  const reservePayload = JSON.stringify({ quantity: 2 });
  const postParams = {
    headers: {
      'Content-Type': 'application/json',
      'Connection': 'keep-alive',
    },
    timeout: '30s',
  };

  res = http.post(`${BASE_URL}/inventory/1/reserve`, reservePayload, postParams);
  check(res, {
    'reserve inventory successful': (r) => r.status === 200,
  });

  sleep(0.1);

  // Fulfill a reservation
  res = http.post(`${BASE_URL}/inventory/1/fulfill`, reservePayload, postParams);
  check(res, {
    'fulfill inventory successful': (r) => r.status === 200,
  });

  sleep(0.1);
}
