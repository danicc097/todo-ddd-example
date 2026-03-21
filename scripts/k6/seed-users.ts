import { check, sleep } from "k6";
import { uuidv4 } from "https://jslib.k6.io/k6-utils/1.4.0/index.js";
import { TodoDDDAPIClient } from "./todoDDDAPI.ts";

const API_URL = __ENV.API_URL || "http://127.0.0.1:8090";
const BASE = `${API_URL}/api/v1`;
const PASSWORD = "BenchmarkPass123!";
const NUM_USERS = __ENV.NUM_USERS ? parseInt(__ENV.NUM_USERS) : 20;

export const options = {
  vus: 1,
  iterations: 1, // setup() does all the work
};

const client = new TodoDDDAPIClient({
  baseUrl: BASE,
});

export function setup() {
  let ready = false;
  for (let i = 0; i < 30; i++) {
    const { response } = client.healthz();
    if (response.status === 200) {
      ready = true;
      break;
    }
    sleep(2);
  }
  if (!ready) throw new Error(`✗ App not responding at ${BASE}`);

  console.log(`── Seeding ${NUM_USERS} users...`);
  const users = [];
  const timestamp = new Date().getTime();

  for (let i = 1; i <= NUM_USERS; i++) {
    const email = `bench-${timestamp}-${i}@load-test.dev`;

    const reg = client.register(
      { email: email, name: `Bench ${i}`, password: PASSWORD },
      { "Idempotency-Key": uuidv4() },
      { headers: { "x-skip-rate-limit": "1" } },
    );

    if (!check(reg.response, { "registered (201)": (r) => r.status === 201 })) {
      console.error(`✗ Register failed for ${email}. Status: ${reg.response.status}`);
      continue;
    }

    const login = client.login(
      { email: email, password: PASSWORD },
      { headers: { "x-skip-rate-limit": "1" } },
    );

    if (!check(login.response, { "logged in (200)": (r) => r.status === 200 })) {
      console.error(`✗ Login failed for ${email}. Status: ${login.response.status}`);
      continue;
    }

    const token = login.data.accessToken;

    const ws = client.onboardWorkspace(
      { name: `Bench WS ${i}`, description: "Load test workspace" },
      { "Idempotency-Key": uuidv4() },
      { headers: { Authorization: `Bearer ${token}` } },
    );

    if (!check(ws.response, { "workspace created (201)": (r) => r.status === 201 })) {
      console.error(`✗ Workspace creation failed for ${email}. Status: ${ws.response.status}`);
      continue;
    }

    users.push({ email: email, token: token, workspaceId: ws.data.id });
  }

  return users;
}

export default function () {
  // all work done in setup()
}

// https://grafana.com/docs/k6/latest/results-output/end-of-test/custom-summary/
export function handleSummary(data: any) {
  return {
    [__ENV.OUTPUT_FILE || "users.json"]: JSON.stringify(data.setup_data, null, 2),
  };
}
