#!/usr/bin/env bash

set -uo pipefail

API_URL="${API_URL:-http://127.0.0.1:8090}"
BASE="${API_URL}/api/v1"
NUM_USERS="${NUM_USERS:-20}"
OUTPUT_FILE="$(dirname "$0")/users.json"
PASSWORD="BenchmarkPass123!"

echo "Waiting for API at ${BASE}/ping..."
for i in $(seq 1 30); do
	if curl -s "${BASE}/ping" >/dev/null 2>&1; then break; fi
	if [ "$i" -eq 30 ]; then
		echo "✗ App not responding at ${BASE}"
		exit 1
	fi
	sleep 2
done

users_json="["
separator=""

for i in $(seq 1 "$NUM_USERS"); do
	TIMESTAMP=$(date +%s%N | cut -c1-13)
	EMAIL="bench-${TIMESTAMP}-${i}@load-test.dev"
	IDEMPOTENCY_KEY=$(uuidgen | tr '[:upper:]' '[:lower:]')

	REG_RES=$(curl -s -w "\n%{http_code}" -X POST "${BASE}/auth/register" \
		-H "Content-Type: application/json" \
		-H "Idempotency-Key: ${IDEMPOTENCY_KEY}" \
		-H "x-skip-rate-limit: 1" \
		-d "{\"email\":\"${EMAIL}\",\"name\":\"Bench ${i}\",\"password\":\"${PASSWORD}\"}")

	REG_BODY=$(echo "$REG_RES" | sed -e '$ d')
	REG_STATUS=$(echo "$REG_RES" | tail -n1)

	if [ "$REG_STATUS" != "201" ]; then
		echo "✗ Register failed for ${EMAIL}. Status: ${REG_STATUS}, Body: ${REG_BODY}"
		continue
	fi

	LOGIN_RES=$(curl -s -w "\n%{http_code}" -X POST "${BASE}/auth/login" \
		-H "Content-Type: application/json" \
		-H "x-skip-rate-limit: 1" \
		-d "{\"email\":\"${EMAIL}\",\"password\":\"${PASSWORD}\"}")

	LOGIN_BODY=$(echo "$LOGIN_RES" | sed -e '$ d')
	LOGIN_STATUS=$(echo "$LOGIN_RES" | tail -n1)

	if [ "$LOGIN_STATUS" != "200" ]; then
		echo "✗ Login failed for ${EMAIL}. Status: ${LOGIN_STATUS}, Body: ${LOGIN_BODY}"
		continue
	fi

	TOKEN=$(echo "$LOGIN_BODY" | jq -r '.accessToken // empty')

	WS_KEY=$(uuidgen | tr '[:upper:]' '[:lower:]')
	WS_RES=$(curl -s -w "\n%{http_code}" -X POST "${BASE}/workspaces" \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer ${TOKEN}" \
		-H "Idempotency-Key: ${WS_KEY}" \
		-d "{\"name\":\"Bench WS ${i}\",\"description\":\"Load test workspace\"}")

	WS_BODY=$(echo "$WS_RES" | sed -e '$ d')
	WS_STATUS=$(echo "$WS_RES" | tail -n1)

	if [ "$WS_STATUS" != "201" ]; then
		echo "✗ Workspace creation failed for ${EMAIL}. Status: ${WS_STATUS}, Body: ${WS_BODY}"
		continue
	fi

	WS_ID=$(echo "$WS_BODY" | jq -r '.id // empty')

	users_json="${users_json}${separator}{\"email\":\"${EMAIL}\",\"token\":\"${TOKEN}\",\"workspaceId\":\"${WS_ID}\"}"
	separator=","
done

users_json="${users_json}]"
echo "$users_json" | jq '.' >"$OUTPUT_FILE"
echo "✅ Seeded $(jq 'length' "$OUTPUT_FILE") users to $OUTPUT_FILE"
