import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');

// NOTE: This test uses ramping-arrival-rate executor to control request rate precisely.
// This is important when testing via kubectl port-forward, which is single-threaded
// and cannot handle high request volumes. The Orders service also makes gRPC calls
// to Products and Inventory services, so we use lower rates than other tests.
export const options = {
  // Use ramping-arrival-rate executor for better control over request rate
  // This prevents overwhelming port-forward connections
  // Using lower rates because Orders service makes gRPC calls to other services
  scenarios: {
    orders_test: {
      executor: 'ramping-arrival-rate',
      startRate: 3,                 // Start with 3 requests per second (lower due to gRPC calls)
      timeUnit: '1s',
      preAllocatedVUs: 10,          // Pre-allocate 10 VUs
      maxVUs: 25,                   // Allow up to 25 VUs if needed
      stages: [
        { duration: '30s', target: 3 },   // Stay at 3 req/s for warmup
        { duration: '2m', target: 8 },    // Ramp up to 8 req/s
        { duration: '30s', target: 0 },   // Ramp down to 0
      ],
    },
  },
  
  // Connection settings for port-forward compatibility
  noConnectionReuse: false,         // Reuse connections to reduce overhead
  
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
