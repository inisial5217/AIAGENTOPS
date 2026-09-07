import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  scenarios: {
    api_throughput: {
      executor: 'constant-arrival-rate',
      rate: 1000,
      timeUnit: '1s',
      duration: '10s',
      preAllocatedVUs: 40,
      maxVUs: 150,
    },
  },
  thresholds: {
    http_req_duration: ['p(99)<200'], // p99 latency < 200ms
    http_req_failed: ['rate<0.01'],    // error rate < 1%
  },
};

export default function () {
  const res = http.get('http://127.0.0.1:8080/healthz');
  check(res, {
    'status 200': (r) => r.status === 200,
  });

  sleep(0.01);
}
