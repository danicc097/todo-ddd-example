import { check, group } from "k6";
import { Counter, Trend, Rate } from "k6/metrics";
import { SharedArray } from "k6/data";
import { Options, Scenario } from "k6/options";
import { uuidv4 } from "https://jslib.k6.io/k6-utils/1.4.0/index.js";

import { TodoDDDAPIClient, Todo, TodoStatus } from "./todoDDDAPI.ts";
import { Response } from "k6/http";

interface User {
  token: string;
  workspaceId: string;
}

const BASE_URL = (__ENV.API_URL || "http://127.0.0.1:8090") + "/api/v1";
const SCENARIO = __ENV.SCENARIO || "load";

const users = new SharedArray<User>("users", function () {
  const data = JSON.parse(open("./users.json"));
  if (!data || data.length === 0) {
    throw new Error("users.json empty — run seed-users.sh first");
  }
  return data as User[];
});

const todoCreated = new Counter("todo_created_total");
const todoCompleted = new Counter("todo_completed_total");
const focusStarted = new Counter("focus_started_total");
const apiErrorRate = new Rate("api_error_rate");
const createTrend = new Trend("todo_create_ms", true);
const listTrend = new Trend("todo_list_ms", true);

const SCENARIOS: Record<string, Scenario> = {
  /* TODO: check */
  smoke: {
    executor: "constant-arrival-rate",
    rate: 5,
    timeUnit: "1s",
    duration: "30s",
    preAllocatedVUs: 5,
    maxVUs: 50,
  },
  /* TODO: check */
  load: {
    executor: "ramping-arrival-rate",
    startRate: 0,
    timeUnit: "1s",
    preAllocatedVUs: 50,
    maxVUs: 300,
    stages: [
      { duration: "1m", target: 10 },
      { duration: "5m", target: 10 },
      { duration: "2m", target: 30 },
      { duration: "3m", target: 30 },
      { duration: "1m", target: 0 },
    ],
  },
  /* TODO: check */
  stress: {
    executor: "ramping-arrival-rate",
    startRate: 0,
    timeUnit: "1s",
    preAllocatedVUs: 100,
    maxVUs: 500,
    stages: [
      { duration: "20s", target: 20 },
      { duration: "2m", target: 25 },
      { duration: "5m", target: 50 },
      { duration: "1m", target: 0 },
    ],
  },
  /* TODO: check */
  soak: {
    executor: "constant-arrival-rate",
    rate: 15,
    timeUnit: "1s",
    duration: "30m",
    preAllocatedVUs: 50,
    maxVUs: 300,
  },
  spike: {
    executor: "ramping-arrival-rate",
    startRate: 0,
    timeUnit: "1s",
    preAllocatedVUs: 25,
    maxVUs: 300,
    stages: [
      { duration: "10s", target: 5 },
      { duration: "30s", target: 25 },
      { duration: "10s", target: 5 },
      { duration: "30s", target: 5 },
      { duration: "10s", target: 0 },
    ],
  },
};

// assumes we use limits from *dev compose files
export const options: Options = {
  scenarios: { [SCENARIO]: SCENARIOS[SCENARIO] || SCENARIOS.load },
  thresholds: {
    "http_req_duration{endpoint:createTodo}": ["p(95)<500", "p(99)<1500"],
    "http_req_duration{endpoint:getWorkspaceTodos}": ["p(95)<2000", "p(99)<4500"],
    "http_req_duration{endpoint:completeTodo}": ["p(95)<500", "p(99)<1500"],
    "http_req_duration{endpoint:startFocus}": ["p(95)<300"],
    "http_req_duration{endpoint:ping}": ["p(99)<100"],
    http_req_failed: ["rate<0.01"],
    api_error_rate: ["rate<0.02"],
    todo_create_ms: ["p(95)<500"],
    todo_list_ms: ["p(95)<2000"],
  },
};

/**
 * Builds base parameters required by the generated client for authentication and tracking.
 */
function getClientConfig(token: string, endpointTag: string) {
  const traceId = uuidv4().replace(/-/g, "") + uuidv4().replace(/-/g, "").slice(0, 16);

  return {
    headers: {
      Authorization: `Bearer ${token}`,
      traceparent: `00-${traceId}-01`,
      "x-skip-rate-limit": "1",
    },
    tags: { endpoint: endpointTag },
  };
}

function ok(res: Response, label: string): boolean {
  const passed = check(res, {
    [`${label} 2xx`]: (r) => r.status >= 200 && r.status < 300,
  });
  if (!passed) {
    console.error(`[${label}] Failed. Status: ${res.status}, Body: ${res.body}`);
  }
  apiErrorRate.add(!passed ? 1 : 0);
  return passed;
}

export default function () {
  const user = users[(__VU - 1) % users.length];
  const { token, workspaceId } = user;

  // Initialize the base client
  const api = new TodoDDDAPIClient({ baseUrl: BASE_URL });

  group("health", () => {
    const { response } = api.ping(getClientConfig(token, "ping"));
    check(response, { pong: (r) => r.status === 200 && r.body === "pong" });
  });

  let todoId = "";

  group("create_todo", () => {
    const t0 = Date.now();
    const { response, data } = api.createTodo(
      workspaceId,
      { title: `k6-todo-${uuidv4().slice(0, 8)}` },
      {}, // Headers specific to this request (Idempotency-Key if needed)
      getClientConfig(token, "createTodo"), // Request parameters mapping (headers/tags)
    );
    createTrend.add(Date.now() - t0);

    if (ok(response, "create_todo")) {
      todoId = data.id;
      todoCreated.add(1);
    }
  });

  if (!todoId) {
    return; // don't sleep, *-arrival-rate executors handle the pacing automatically
  }

  group("list_todos", () => {
    const t0 = Date.now();
    const { response, data } = api.getWorkspaceTodos(
      workspaceId,
      { limit: 100, offset: 0 },
      getClientConfig(token, "getWorkspaceTodos"),
    );
    listTrend.add(Date.now() - t0);

    ok(response, "list_todos");
    check(response, {
      "list is array": () => Array.isArray(data),
      "new todo in list": () => {
        return Array.isArray(data) && data.some((t: Todo) => t.id === todoId);
      },
    });
  });

  group("get_todo", () => {
    const { response, data } = api.getTodoByID(todoId, getClientConfig(token, "getTodoByID"));
    ok(response, "get_todo");
    check(response, {
      "status PENDING": () => data.status === TodoStatus.PENDING,
      "id matches": () => data.id === todoId,
    });
  });

  group("focus_session", () => {
    const { response: startRes } = api.startFocus(todoId, getClientConfig(token, "startFocus"));
    if (ok(startRes, "start_focus")) focusStarted.add(1);

    const { response: stopRes } = api.stopFocus(todoId, getClientConfig(token, "stopFocus"));
    ok(stopRes, "stop_focus");
  });

  group("complete_todo", () => {
    const { response } = api.completeTodo(todoId, {}, getClientConfig(token, "completeTodo"));
    if (ok(response, "complete_todo")) todoCompleted.add(1);
  });
}
