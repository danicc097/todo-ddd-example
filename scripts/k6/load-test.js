// @ts-check
import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Counter, Trend, Rate } from 'k6/metrics';
import { SharedArray } from 'k6/data';
// @ts-ignore
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

/**
 * @typedef {import('k6/options').Scenario} Scenario
 * @typedef {import('k6/options').Options} Options
 * @typedef {{ token: string, workspaceId: string }} User
 * @typedef {{ id: string, title: string, status: string }} Todo
 */

const BASE_URL = (__ENV.API_URL || 'http://127.0.0.1:8090') + '/api/v1';
const SCENARIO = __ENV.SCENARIO || 'load';

/** @type {import('k6/data').SharedArray<User>} */
const users = new SharedArray('users', function () {
  const data = JSON.parse(open('./users.json'));
  if (!data || data.length === 0)
    throw new Error('users.json empty — run seed-users.sh first');
  return /** @type {User[]} */ (data);
});

const todoCreated = new Counter('todo_created_total');
const todoCompleted = new Counter('todo_completed_total');
const focusStarted = new Counter('focus_started_total');
const apiErrorRate = new Rate('api_error_rate');
const createTrend = new Trend('todo_create_ms', true);
const listTrend = new Trend('todo_list_ms', true);

/** @type {{ [name: string]: Scenario }} */
const SCENARIOS = {
  smoke: {
    executor: 'constant-vus',
    vus: 1,
    duration: '30s'
  },
  load: {
    executor: 'ramping-vus',
    startVUs: 0,
    stages: [
      { duration: '1m', target: 10 },
      { duration: '5m', target: 10 },
      { duration: '2m', target: 30 },
      { duration: '3m', target: 30 },
      { duration: '1m', target: 0 },
    ],
  },
  stress: {
    executor: 'ramping-vus',
    startVUs: 0,
    stages: [
      { duration: '2m', target: 20 },
      { duration: '2m', target: 50 },
      { duration: '2m', target: 100 },
      { duration: '5m', target: 100 },
      { duration: '2m', target: 0 },
    ],
  },
  soak: {
    executor: 'constant-vus',
    vus: 10,
    duration: '30m'
  },
  spike: {
    executor: 'ramping-vus',
    startVUs: 0,
    stages: [
      { duration: '10s', target: 5 },
      { duration: '30s', target: 100 },
      { duration: '10s', target: 5 },
      { duration: '1m', target: 5 },
      { duration: '10s', target: 0 },
    ],
  },
};

/** @type {Options} */
export const options = {
  scenarios: { [SCENARIO]: SCENARIOS[SCENARIO] || SCENARIOS.load },
  thresholds: {
    'http_req_duration{endpoint:create_todo}': ['p(95)<500', 'p(99)<1500'],
    'http_req_duration{endpoint:list_todos}': ['p(95)<300', 'p(99)<800'],
    'http_req_duration{endpoint:complete_todo}': ['p(95)<500', 'p(99)<1500'],
    'http_req_duration{endpoint:start_focus}': ['p(95)<300'],
    'http_req_duration{endpoint:ping}': ['p(99)<100'],
    'http_req_failed': ['rate<0.01'],
    'api_error_rate': ['rate<0.02'],
    'todo_create_ms': ['p(95)<500'],
    'todo_list_ms': ['p(95)<300'],
  },
};

/**
 * @param {string} token
 * @param {Record<string, string>} [extra]
 * @param {boolean} [omitContentType]
 */
function headers(token, extra = {}, omitContentType = false) {
  const traceId = uuidv4().replace(/-/g, '') + uuidv4().replace(/-/g, '').slice(0, 16);
  /** @type {Record<string, string>} */
  const h = {
    'Authorization': `Bearer ${token}`,
    'traceparent': `00-${traceId}-01`,
    'x-skip-rate-limit': '1',
  };

  if (!omitContentType) {
    h['Content-Type'] = 'application/json';
  }

  return Object.assign(h, extra);
}

/**
 * @param {import('k6/http').Response} res
 * @param {string} label
 */
function ok(res, label) {
  const passed = check(res, {
    [`${label} 2xx`]: (r) => r.status >= 200 && r.status < 300,
  });
  if (!passed) {
    console.error(`[${label}] Failed. Status: ${res.status}, Body: ${res.body}`);
  }
  apiErrorRate.add(!passed);
  return passed;
}

export default function () {
  const user = users[(__VU - 1) % users.length];
  const { token, workspaceId } = user;

  group('health', () => {
    const r = http.get(`${BASE_URL}/ping`, { tags: { endpoint: 'ping' } });
    check(r, { 'pong': (r) => r.status === 200 && r.body === 'pong' });
  });

  /** @type {string | undefined} */
  let todoId;

  group('create_todo', () => {
    const t0 = Date.now();
    const payload = JSON.stringify({
      title: `k6-todo-${uuidv4().slice(0, 8)}`
    });

    const r = http.post(
      `${BASE_URL}/workspaces/${workspaceId}/todos`,
      payload,
      {
        headers: headers(token, {}, false),
        tags: { endpoint: 'create_todo' }
      }
    );
    createTrend.add(Date.now() - t0);

    if (ok(r, 'create_todo')) {
      todoId = /** @type {string} */ (r.json('id'));
      todoCreated.add(1);
    }
  });

  if (!todoId) { sleep(1); return; }
  sleep(0.3);

  group('list_todos', () => {
    const t0 = Date.now();
    const r = http.get(
      `${BASE_URL}/workspaces/${workspaceId}/todos?limit=20&offset=0`,
      { headers: headers(token, {}, false), tags: { endpoint: 'list_todos' } }
    );
    listTrend.add(Date.now() - t0);
    ok(r, 'list_todos');
    check(r, {
      'list is array': (r) => Array.isArray(r.json()),
      'new todo in list': (r) => {
        const list = /** @type {Todo[]} */ (r.json());
        return Array.isArray(list) && list.some(t => t.id === todoId);
      },
    });
  });
  sleep(0.5);

  group('get_todo', () => {
    const r = http.get(`${BASE_URL}/todos/${todoId}`,
      { headers: headers(token, {}, false), tags: { endpoint: 'get_todo' } });
    ok(r, 'get_todo');
    check(r, {
      'status PENDING': (r) => r.json('status') === 'PENDING',
      'id matches': (r) => r.json('id') === todoId,
    });
  });
  sleep(0.2);

  group('focus_session', () => {
    const start = http.post(`${BASE_URL}/todos/${todoId}/focus/start`, '',
      { headers: headers(token, {}, true), tags: { endpoint: 'start_focus' } });
    if (ok(start, 'start_focus')) focusStarted.add(1);

    sleep(0.5);

    const stop = http.post(`${BASE_URL}/todos/${todoId}/focus/stop`, '',
      { headers: headers(token, {}, true), tags: { endpoint: 'stop_focus' } });
    ok(stop, 'stop_focus');
  });
  sleep(0.3);

  group('complete_todo', () => {
    const r = http.patch(`${BASE_URL}/todos/${todoId}/complete`, '',
      {
        headers: headers(token, {}, true),
        tags: { endpoint: 'complete_todo' }
      });
    if (ok(r, 'complete_todo')) todoCompleted.add(1);
  });

  sleep(1);
}
