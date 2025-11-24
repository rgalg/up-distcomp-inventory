import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

// custom metrics
const errorRate = new Rate('errors');

export const options = {
  stages: [
    { duration: '30s', target: 10 },  // ramp up to 10 users
    { duration: '1m', target: 10 },   // stay at 10 users
    { duration: '30s', target: 50 },  // ramp up to 50 users
    { duration: '1m', target: 50 },   // stay at 50 users
    { duration: '30s', target: 0 },   // ramp down to 0 users
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests should be below 500ms
    errors: ['rate<0.1'],             // error rate should be less than 10%
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8001';

export default function () {
  // test GET all products
  let res = http.get(`${BASE_URL}/products`);
  const success = check(res, {
    'status is 200': (r) => r.status === 200,
    'response has products': (r) => JSON.parse(r.body).length > 0,
  });
  errorRate.add(!success);

  sleep(1);

  // test GET for a single product (assuming product ID 1 exists)
  res = http.get(`${BASE_URL}/products/1`);
  check(res, {
    'product detail status is 200': (r) => r.status === 200,
    'product has name': (r) => JSON.parse(r.body).name !== undefined,
  });

  sleep(1);

  // test CREATE product (stress test)
  const payload = JSON.stringify({
    name: `Test Product ${Date.now()}`,
    description: 'Load test product',
    price: 99.99,
    category: 'Test Category',
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  res = http.post(`${BASE_URL}/products`, payload, params);
  check(res, {
    'create product status is 201': (r) => r.status === 201 || r.status === 200,
  });

  sleep(2);
}
