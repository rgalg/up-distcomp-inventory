import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  vus: 1,
  duration: '30s',
};

const BASE_URL = 'http://127.0.0.1:8001';  // Use 127.0.0.1 instead of localhost

export default function () {
  console.log('Testing GET /products...');
  
  const res = http.get(`${BASE_URL}/products`);
  
  console.log(`Response Status: ${res.status}`);
  console.log(`Response Body Length: ${res.body ? res.body.length : 0}`);
  
  if (res.status !== 200) {
    console.log(`Error: ${res.error}`);
    console.log(`Body: ${res.body}`);
  } else {
    console.log(`Success! First 100 chars: ${res.body.substring(0, 100)}`);
  }
  
  check(res, {
    'status is 200': (r) => r.status === 200,
  });
  
  sleep(1);
}
