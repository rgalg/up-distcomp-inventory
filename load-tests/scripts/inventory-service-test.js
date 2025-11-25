import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');

// NOTE: This test uses ramping-arrival-rate executor to control request rate precisely.
// This is important when testing via kubectl port-forward, which is single-threaded
// and cannot handle high request volumes. Adjust 'startRate' and 'stages' for your needs.
export const options = {
  // Use ramping-arrival-rate executor for better control over request rate
  // This prevents overwhelming port-forward connections
  scenarios: {
    inventory_test: {
      executor: 'ramping-arrival-rate',
      startRate: 5,                 // Start with 5 requests per second
      timeUnit: '1s',
      preAllocatedVUs: 10,          // Pre-allocate 10 VUs
      maxVUs: 30,                   // Allow up to 30 VUs if needed
      stages: [
        { duration: '30s', target: 5 },   // Stay at 5 req/s for warmup
        { duration: '2m', target: 10 },   // Ramp up to 10 req/s
        { duration: '30s', target: 0 },   // Ramp down to 0
      ],
    },
  },
  
  // Connection settings for port-forward compatibility
  noConnectionReuse: false,         // Reuse connections to reduce overhead
  
  thresholds: {
    http_req_duration: ['p(95)<500'],
    errors: ['rate<0.1'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8002';

export default function () {
  // fet all inventory items
  let res = http.get(`${BASE_URL}/inventory`);
  const success = check(res, {
    'inventory list status is 200': (r) => r.status === 200,
  });
  errorRate.add(!success);

  sleep(1);

  // get inventory data for a specific product
  res = http.get(`${BASE_URL}/inventory/1`);
  check(res, {
    'inventory detail status is 200': (r) => r.status === 200,
  });

  sleep(1);

  // reserve inventory for a specific product
  const reservePayload = JSON.stringify({ quantity: 2 });
  const params = { headers: { 'Content-Type': 'application/json' } };

  res = http.post(`${BASE_URL}/inventory/1/reserve`, reservePayload, params);
  check(res, {
    'reserve inventory successful': (r) => r.status === 200,
  });

  sleep(1);

  // fulfill a reservation
  res = http.post(`${BASE_URL}/inventory/1/fulfill`, reservePayload, params);
  check(res, {
    'fulfill inventory successful': (r) => r.status === 200,
  });

  sleep(2);
}
