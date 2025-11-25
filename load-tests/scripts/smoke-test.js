import http from 'k6/http';
import { check, sleep } from 'k6';

// PORT-FORWARD TESTING ONLY
// This test is designed to run via kubectl port-forward which has limited throughput.
// For high-load testing, use the in-cluster tests in k8s/high-load/ directory.
// See README.md for more information on testing approaches.

export const options = {
  // Use constant-arrival-rate executor for predictable request rate
  // This is ideal for port-forward testing where we need to limit throughput
  scenarios: {
    smoke_test: {
      executor: 'constant-arrival-rate',
      rate: 5,              // 5 requests per second (port-forward friendly)
      timeUnit: '1s',
      duration: '30s',
      preAllocatedVUs: 2,   // pre-allocate 2 VUs
      maxVUs: 5,            // allow up to 5 VUs if needed
    },
  },
  
  noConnectionReuse: false,  // reuse connections for better performance
  userAgent: 'k6-load-test',
  
  thresholds: {
    http_req_failed: ['rate<0.1'],  // less than 10% failed requests
    http_req_duration: ['p(95)<2000'], // 95% of requests under 2s
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8001';

export default function () {
  const params = {
    headers: {
      'Connection': 'keep-alive',  // keep connections alive
    },
    timeout: '30s',  // generous timeout for port-forward
  };
  
  const res = http.get(`${BASE_URL}/products`, params);
  
  const success = check(res, {
    'status is 200': (r) => r.status === 200,
    'response has body': (r) => r.body && r.body.length > 0,
  });
  
  if (!success) {
    console.log(`GET /products failed: Status=${res.status}, Error=${res.error}`);
  }
  
  // Small sleep to allow port-forward to stabilize between requests
  sleep(0.1);
}
