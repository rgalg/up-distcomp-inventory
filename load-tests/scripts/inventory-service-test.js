import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');

export const options = {
  stages: [
    { duration: '30s', target: 20 },
    { duration: '2m', target: 20 },
    { duration: '30s', target: 0 },
  ],
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
