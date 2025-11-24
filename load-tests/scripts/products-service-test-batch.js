import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');

export const options = {
  stages: [
    { duration: '30s', target: 5 },   // Start with just 5 users
    { duration: '1m', target: 5 },
    { duration: '30s', target: 0 },
  ],
  
  // Critical: Limit concurrent requests
  batch: 5,  // Max 5 parallel requests
  batchPerHost: 3,  // Max 3 requests per host
  
  thresholds: {
    http_req_duration: ['p(95)<2000'],  // More lenient timeout
    errors: ['rate<0.2'],  // Allow more errors due to port-forward
  },
};

const BASE_URL = 'http://localhost:8001';

export default function () {
  const params = {
    headers: {
      'Connection': 'keep-alive',
    },
  };
  
  // Test GET all products
  let res = http.get(`${BASE_URL}/products`, params);
  
  const getAllSuccess = check(res, {
    'GET /products status is 200': (r) => r.status === 200,
    'response has products': (r) => {
      try {
        const products = JSON.parse(r.body);
        return Array.isArray(products) && products.length > 0;
      } catch (e) {
        return false;
      }
    },
  });
  
  errorRate.add(!getAllSuccess);

  sleep(2);  // Longer sleep to reduce request rate

  // Test GET single product
  res = http.get(`${BASE_URL}/products/1`, params);
  
  const getOneSuccess = check(res, {
    'GET /products/1 status is 200': (r) => r.status === 200,
    'product has name': (r) => {
      try {
        const product = JSON.parse(r.body);
        return product.name === 'Laptop';
      } catch (e) {
        return false;
      }
    },
  });
  
  errorRate.add(!getOneSuccess);

  sleep(2);

  // Test CREATE product (less frequently)
  if (__ITER % 3 === 0) {  // Only every 3rd iteration
    const newProduct = {
      name: `Test Product ${Date.now()}`,
      description: 'Load test product',
      price: 99.99,
      category: 'Test Category',
    };

    const postParams = {
      headers: {
        'Content-Type': 'application/json',
        'Connection': 'keep-alive',
      },
    };

    res = http.post(`${BASE_URL}/products`, JSON.stringify(newProduct), postParams);
    
    const createSuccess = check(res, {
      'POST /products status is 201 or 200': (r) => r.status === 201 || r.status === 200,
    });
    
    errorRate.add(!createSuccess);
  }

  sleep(3);  // Longer sleep at end
}
