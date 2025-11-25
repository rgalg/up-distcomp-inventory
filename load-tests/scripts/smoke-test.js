import http from 'k6/http';
import { check, sleep } from 'k6';

// NOTE: This test is designed to work with kubectl port-forward connections.
// Port-forward is single-threaded and not designed for high-throughput load testing,
// so we use a constant-arrival-rate executor to control the request rate precisely.
export const options = {
  // Use constant-arrival-rate executor to limit request rate
  // This prevents overwhelming port-forward connections
  scenarios: {
    smoke: {
      executor: 'constant-arrival-rate',
      rate: 5,                    // 5 requests per second (conservative for port-forward)
      timeUnit: '1s',
      duration: '30s',
      preAllocatedVUs: 2,         // pre-allocate 2 VUs
      maxVUs: 5,                  // allow up to 5 VUs if needed
    },
  },
  
  noConnectionReuse: false,  // reuse connections (default but let's be explicit)
  userAgent: 'k6-load-test',
  
  thresholds: {
    http_req_failed: ['rate<0.1'],
  },
};

const BASE_URL = 'http://localhost:8001';

export default function () {
  console.log('Testing GET /products...');
  
  const params = {
    headers: {
      'Connection': 'keep-alive',  // keep connections alive
    },
    timeout: '30s',  // timeout
  };
  
  const res = http.get(`${BASE_URL}/products`, params);
  
  console.log(`Response Status: ${res.status}`);
  console.log(`Response Body Length: ${res.body ? res.body.length : 0}`);
  
  if (res.status !== 200) {
    console.log(`Error: ${res.error}`);
    console.log(`Body: ${res.body}`);
  } else {
    console.log(`Success! Got products`);
  }
  
  check(res, {
    'status is 200': (r) => r.status === 200,
  });
  
  sleep(2);  // sleep between requests
}
