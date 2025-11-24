import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');

export const options = {
  stages: [
    { duration: '30s', target: 15 },
    { duration: '2m', target: 15 },
    { duration: '30s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<1000'], // orders might take longer due to gRPC calls (this service calls Products and Inventory services)
    errors: ['rate<0.1'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8003';

export default function () {
  // get all orders
  let res = http.get(`${BASE_URL}/orders`);
  const success = check(res, {
    'orders list status is 200': (r) => r.status === 200,
  });
  errorRate.add(!success);

  sleep(1);

  // create order (tests gRPC communication with Products and Inventory services)
  const orderPayload = JSON.stringify({
    customer_id: Math.floor(Math.random() * 1000) + 1,
    items: [
      { product_id: 1, quantity: 1 },
      { product_id: 2, quantity: 2 },
    ],
  });

  const params = { headers: { 'Content-Type': 'application/json' } };

  res = http.post(`${BASE_URL}/orders`, orderPayload, params);
  const orderCreated = check(res, {
    'create order status is 200 or 201': (r) => r.status === 200 || r.status === 201,
  });

  if (orderCreated) {
    const orderId = JSON.parse(res.body).id || JSON.parse(res.body).order_id;
    
    sleep(2);

    // fulfill order
    res = http.post(`${BASE_URL}/orders/${orderId}/fulfill`);
    check(res, {
      'fulfill order successful': (r) => r.status === 200,
    });
  }

  sleep(2);
}
