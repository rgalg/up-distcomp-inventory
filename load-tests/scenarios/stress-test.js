// high load to test Horizontal Pod Autoscaler (HPA) behavior
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '2m', target: 50 },   // ramp up to 50 users
    { duration: '5m', target: 50 },   // stay at 50
    { duration: '2m', target: 100 },  // ramp to 100 (should trigger HPA)
    { duration: '5m', target: 100 },  // stay at 100
    { duration: '2m', target: 150 },  // ramp to 150
    { duration: '5m', target: 150 },  // stay at 150
    { duration: '5m', target: 0 },    // ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<3000'],
    http_req_failed: ['rate<0.3'],
  },
};

const ORDERS_URL = __ENV.ORDERS_URL || 'http://localhost:8003';

export default function () {
  const payload = JSON.stringify({
    customer_id: Math.floor(Math.random() * 10000),
    items: [
      { product_id: Math.floor(Math.random() * 5) + 1, quantity: Math.floor(Math.random() * 5) + 1 },
    ],
  });

  const params = { headers: { 'Content-Type': 'application/json' } };
  const res = http.post(`${ORDERS_URL}/orders`, payload, params);
  
  check(res, { 'status is 2xx': (r) => r.status >= 200 && r.status < 300 });
  
  sleep(1);
}
