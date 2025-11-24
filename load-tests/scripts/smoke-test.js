import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  vus: 1,
  duration: '30s',
  
  noConnectionReuse: false,  // reuse connections (default but let's be explicit)
  userAgent: 'k6-load-test',
  
  // limit RPS to avoid overwhelming port-forward
  rps: 10,  // max 10 requests per second across all VUs
  
  // add longer timeouts
  httpDebug: 'full',
  
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
