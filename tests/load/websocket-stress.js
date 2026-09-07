import ws from 'k6/ws';
import { check, sleep } from 'k6';

export const options = {
  scenarios: {
    websocket_stress: {
      executor: 'ramping-vus',
      startVUs: 20,
      stages: [
        { duration: '3s', target: 200 },
        { duration: '6s', target: 500 },
        { duration: '4s', target: 500 },
        { duration: '2s', target: 0 },
      ],
      gracefulRampDown: '2s',
    },
  },
  thresholds: {
    'ws_connecting': ['p(95)<1000'],
    'checks': ['rate>0.99'],
  },
};

export default function () {
  const url = 'ws://127.0.0.1:8080/ws?token=dev-token-admin';

  const res = ws.connect(url, {}, function (socket) {
    socket.on('open', () => {
      socket.send(JSON.stringify({
        action: 'subscribe',
        topic: 'telemetry:cpu',
      }));
      socket.send(JSON.stringify({ action: 'ping' }));
    });

    socket.on('message', () => {
      socket.close();
    });

    socket.on('error', (e) => {
      console.error('WS error: ', e.error());
    });
  });

  check(res, {
    'status is 101': (r) => r && r.status === 101,
  });

  sleep(0.5);
}
