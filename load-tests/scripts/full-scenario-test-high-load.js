import http from 'k6/http';
import { check, group, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

const errorRate = new Rate('errors');
const orderCreationTime = new Trend('order_creation_duration');

export const options = {
  stages: [
    { duration: '1m', target: 10 },   // warm up
    { duration: '2m', target: 50 },   // normal load
    { duration: '2m', target: 100 },  // high load (should trigger HPA)
    { duration: '3m', target: 150 },  // peak load (should trigger HPA)
    { duration: '1m', target: 30 },   // scale down
    { duration: '1m', target: 0 },    // cool down
  ],
  thresholds: {
    http_req_duration: ['p(95)<2000'],
    errors: ['rate<0.15'],
    order_creation_duration: ['p(95)<3000'],
  },
};

const PRODUCTS_URL = __ENV.PRODUCTS_URL || 'http://localhost:8001';
const INVENTORY_URL = __ENV.INVENTORY_URL || 'http://localhost:8002';
const ORDERS_URL = __ENV.ORDERS_URL || 'http://localhost:8003';

export default function () {
  group('Browse Products', function () {
    const res = http.get(`${PRODUCTS_URL}/products`);
    check(res, { 'products loaded': (r) => r.status === 200 });
    sleep(1);
  });

  group('Check Inventory', function () {
    const res = http.get(`${INVENTORY_URL}/inventory/1`);
    check(res, { 'inventory checked': (r) => r.status === 200 });
    sleep(1);
  });

  group('Create and Fulfill Order', function () {
    const orderPayload = JSON.stringify({
      customer_id: Math.floor(Math.random() * 1000) + 1,
      items: [
        { product_id: 1, quantity: 1 },
        { product_id: 3, quantity: 1 },
      ],
    });

    const params = { headers: { 'Content-Type': 'application/json' } };

    // measure order creation time
    const start = new Date();
    const res = http.post(`${ORDERS_URL}/orders`, orderPayload, params);
    const duration = new Date() - start;
    orderCreationTime.add(duration);

    const orderCreated = check(res, {
      'order created': (r) => r.status === 200 || r.status === 201,
    });

    errorRate.add(!orderCreated);

    if (orderCreated) {
      const body = JSON.parse(res.body);
      const orderId = body.id || body.order_id;
      
      sleep(2);

      // fulfill the order
      const fulfillRes = http.post(`${ORDERS_URL}/orders/${orderId}/fulfill`);
      check(fulfillRes, { 'order fulfilled': (r) => r.status === 200 });
    }

    sleep(3);
  });
}
