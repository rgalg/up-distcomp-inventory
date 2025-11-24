// minimal load (only to verify system functionality)
import { full_scenario } from '../scripts/full-scenario.js';

export const options = {
  vus: 1,
  duration: '1m',
  thresholds: {
    http_req_duration: ['p(95)<1000'],
    errors: ['rate<0.01'],
  },
};

export default full_scenario;
