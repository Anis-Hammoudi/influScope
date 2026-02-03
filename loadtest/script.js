import http from 'k6/http';
import { check, sleep } from 'k6';

// Config: Simulates 50 users hitting your API at once
export const options = {
    stages: [
        { duration: '10s', target: 50 }, // Ramp up to 50 users
        { duration: '30s', target: 50 }, // Stay there
        { duration: '10s', target: 0 },  // Ramp down
    ],
};

export default function () {
    // 1. Search for "tech"
    const res = http.get('http://localhost:8080/search?q=tech');

    // 2. Validate response is fast (< 200ms) and successful
    check(res, {
        'is status 200': (r) => r.status === 200,
        'is fast': (r) => r.timings.duration < 200,
    });

    sleep(1);
}