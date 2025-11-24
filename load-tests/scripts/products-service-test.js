import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');

export const options = {
  stages: [
    { duration: '30s', target: 10 },  // Ramp up to 10 users
    { duration: '1m', target: 10 },   // Stay at 10 users
    { duration: '30s', target: 50 },  // Ramp up to 50 users
    { duration: '1m', target: 50 },   // Stay at 50 users
    { duration: '30s', target: 0 },   // Ramp down to 0 users
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests should be below 500ms
    errors: ['rate<0.1'],              // Error rate should be less than 10%
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8001';

export default function () {
  // Test GET all products
  let res = http.get(`${BASE_URL}/products`);
  
  const getAllSuccess = check(res, {
    'GET /products status is 200': (r) => r.status === 200,
    'GET /products has body': (r) => r.body && r.body.length > 0,
  });
  
  errorRate.add(!getAllSuccess);
  
  if (!getAllSuccess) {
    console.log(`GET /products failed: Status=${res.status}, Body=${res.body}`);
  }

  sleep(1);

  // Get product IDs from the response
  if (res.status === 200 && res.body) {
    try {
      const products = JSON.parse(res.body);
      
      if (products && products.length > 0) {
        // Test GET single product with actual ID
        const productId = products[0].id;
        res = http.get(`${BASE_URL}/products/${productId}`);
        
        const getOneSuccess = check(res, {
          'GET /products/{id} status is 200': (r) => r.status === 200,
          'product has name': (r) => {
            try {
              const product = JSON.parse(r.body);
              return product.name !== undefined;
            } catch (e) {
              return false;
            }
          },
        });
        
        errorRate.add(!getOneSuccess);
      }
    } catch (e) {
      console.log(`Error parsing products: ${e}`);
      errorRate.add(true);
    }
  }

  sleep(1);

  // Test CREATE product
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
  
  const createSuccess = check(res, {
    'POST /products status is 201 or 200': (r) => r.status === 201 || r.status === 200,
  });
  
  errorRate.add(!createSuccess);
  
  if (!createSuccess) {
    console.log(`POST /products failed: Status=${res.status}, Body=${res.body}`);
  }

  sleep(2);
}
