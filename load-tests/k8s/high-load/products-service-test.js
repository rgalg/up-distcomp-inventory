import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

// HIGH-LOAD PRODUCTS SERVICE TEST
// Designed for in-cluster testing without port-forward limitations
// Can achieve 100+ req/s with proper resource allocation

const errorRate = new Rate('errors');

export const options = {
  scenarios: {
    high_load_products: {
      executor: 'ramping-arrival-rate',
      startRate: 10,          // Start at 10 requests per second
      timeUnit: '1s',
      preAllocatedVUs: 20,    // Pre-allocate 20 VUs
      maxVUs: 100,            // Allow up to 100 VUs for peak load
      stages: [
        { duration: '1m', target: 30 },    // Warm up to 30 req/s
        { duration: '2m', target: 60 },    // Ramp to 60 req/s
        { duration: '3m', target: 100 },   // Ramp to 100 req/s (should trigger HPA)
        { duration: '2m', target: 100 },   // Stay at 100 req/s
        { duration: '1m', target: 50 },    // Ramp down to 50 req/s
        { duration: '1m', target: 0 },     // Ramp down to 0
      ],
    },
  },
  
  thresholds: {
    http_req_duration: ['p(95)<500'],   // 95% of requests under 500ms
    http_req_failed: ['rate<0.05'],     // Less than 5% failed requests
    errors: ['rate<0.05'],
  },
};

// Use environment variable or default to internal Kubernetes service URL
const BASE_URL = __ENV.BASE_URL || 'http://products-service:8001';

export default function () {
  // Test GET all products
  let res = http.get(`${BASE_URL}/products`);
  
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

  // Get product IDs from the response
  if (res.status === 200 && res.body) {
    try {
      const products = JSON.parse(res.body);
      
      if (products && products.length > 0) {
        // Test GET single product with actual ID
        const productId = products[Math.floor(Math.random() * products.length)].id;
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
      errorRate.add(true);
    }
  }

  // Test CREATE product (less frequently to avoid database bloat)
  if (__ITER % 5 === 0) {
    const payload = JSON.stringify({
      name: `High Load Product ${Date.now()}-${__VU}`,
      description: 'High load test product',
      price: Math.random() * 1000,
      category: 'Load Test Category',
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
  }

  // Minimal sleep for high throughput
  sleep(0.01);
}
