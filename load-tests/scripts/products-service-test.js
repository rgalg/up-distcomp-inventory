import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

// PORT-FORWARD TESTING ONLY
// This test is designed for kubectl port-forward with limited throughput (5-15 req/s).
// For high-load testing, use the in-cluster tests in k8s/high-load/ directory.

// Custom metrics
const errorRate = new Rate('errors');

export const options = {
  // Use ramping-arrival-rate executor for controlled request rate
  // This is port-forward friendly with rates starting at 5 req/s up to 15 req/s max
  scenarios: {
    products_load_test: {
      executor: 'ramping-arrival-rate',
      startRate: 5,           // Start at 5 requests per second
      timeUnit: '1s',
      preAllocatedVUs: 5,     // Pre-allocate 5 VUs
      maxVUs: 20,             // Allow up to 20 VUs if needed
      stages: [
        { duration: '30s', target: 5 },   // Warm up at 5 req/s
        { duration: '1m', target: 10 },   // Ramp to 10 req/s
        { duration: '1m', target: 15 },   // Ramp to 15 req/s (port-forward max)
        { duration: '30s', target: 5 },   // Ramp down to 5 req/s
        { duration: '30s', target: 0 },   // Ramp down to 0
      ],
    },
  },
  
  noConnectionReuse: false,  // Enable connection pooling for better performance
  
  thresholds: {
    http_req_duration: ['p(95)<2000'], // 95% of requests should be below 2s (lenient for port-forward)
    errors: ['rate<0.1'],              // Error rate should be less than 10%
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8001';

export default function () {
  const params = {
    headers: {
      'Connection': 'keep-alive',  // Enable connection pooling
    },
    timeout: '30s',  // Generous timeout for port-forward
  };
  
  // Test GET all products
  let res = http.get(`${BASE_URL}/products`, params);
  
  const getAllSuccess = check(res, {
    'GET /products status is 200': (r) => r.status === 200,
    'GET /products has body': (r) => r.body && r.body.length > 0,
  });
  
  errorRate.add(!getAllSuccess);
  
  if (!getAllSuccess) {
    console.log(`GET /products failed: Status=${res.status}, Body=${res.body}`);
  }

  sleep(0.1);  // Small sleep for port-forward stability

  // Get product IDs from the response
  if (res.status === 200 && res.body) {
    try {
      const products = JSON.parse(res.body);
      
      if (products && products.length > 0) {
        // Test GET single product with actual ID
        const productId = products[0].id;
        res = http.get(`${BASE_URL}/products/${productId}`, params);
        
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

  sleep(0.1);

  // Test CREATE product
  const payload = JSON.stringify({
    name: `Test Product ${Date.now()}`,
    description: 'Load test product',
    price: 99.99,
    category: 'Test Category',
  });

  const postParams = {
    headers: {
      'Content-Type': 'application/json',
      'Connection': 'keep-alive',
    },
    timeout: '30s',
  };

  res = http.post(`${BASE_URL}/products`, payload, postParams);
  
  const createSuccess = check(res, {
    'POST /products status is 201 or 200': (r) => r.status === 201 || r.status === 200,
  });
  
  errorRate.add(!createSuccess);
  
  if (!createSuccess) {
    console.log(`POST /products failed: Status=${res.status}, Body=${res.body}`);
  }

  sleep(0.1);
}
