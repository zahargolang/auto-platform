import http from 'k6/http'
import { check, sleep } from 'k6'
import { Rate, Trend } from 'k6/metrics'

const errorRate = new Rate('errors')
const listingDuration = new Trend('listing_duration', true)

export const options = {
  stages: [
    { duration: '30s', target: 50  },
    { duration: '30s', target: 200 },
    { duration: '30s', target: 500 },
    { duration: '60s', target: 500 },
    { duration: '30s', target: 1000},
    { duration: '60s', target: 1000},
    { duration: '30s', target: 0   },
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],
    errors: ['rate<0.01'],
  },
}

// minikube: HOST_HEADER=auto-platform MINIKUBE_IP=192.168.49.2 k6 run load_test.js
// production: BASE_URL=https://yourdomain.com k6 run load_test.js
const BASE_URL = __ENV.BASE_URL || 'https://auto-platfrom.ru'
const HOST_HEADER = __ENV.HOST_HEADER || 'auto-platform'

export default function () {
  const res = http.get(`${BASE_URL}/api/listings`, {
    headers: { 'Host': HOST_HEADER },
  })

  const ok = check(res, {
    'status 200': (r) => r.status === 200,
    'body not empty': (r) => r.body && r.body.length > 0,
  })

  errorRate.add(!ok)
  listingDuration.add(res.timings.duration)

  sleep(0.1)
}
