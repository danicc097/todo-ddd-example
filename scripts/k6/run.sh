#!/usr/bin/env bash

set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
SCENARIO="${SCENARIO:-load}"
API_URL="${API_URL:-http://127.0.0.1:8090}"
NUM_USERS="${NUM_USERS:-20}"

for dep in k6 curl jq uuidgen; do
	command -v "$dep" >/dev/null || { echo "✗ Missing: $dep" && exit 1; }
done

make -C "${REPO_ROOT}" gen-k6

USERS_FILE="${SCRIPT_DIR}/users.json"
if [ ! -f "$USERS_FILE" ] || [ "${RESEED:-0}" = "1" ]; then
	k6 run --quiet \
		--env API_URL="${API_URL}" \
		--env NUM_USERS="${NUM_USERS}" \
		--env OUTPUT_FILE="${USERS_FILE}" \
		"${SCRIPT_DIR}/seed-users.ts"
else
	echo "── Using $(jq 'length' "$USERS_FILE") existing users (RESEED=1 to refresh)"
fi

mkdir -p "${SCRIPT_DIR}/results"
RESULT_FILE="${SCRIPT_DIR}/results/${SCENARIO}-$(date +%Y%m%dT%H%M%S).json"

K6_OUT="--out json=${RESULT_FILE}"
if [ "${K6_PUSH_PROMETHEUS:-0}" = "1" ]; then
	K6_OUT="${K6_OUT} --out experimental-prometheus-rw"
	export K6_PROMETHEUS_RW_SERVER_URL="${K6_PROMETHEUS_RW_URL:-http://127.0.0.1:9090/api/v1/write}"
	export K6_PROMETHEUS_RW_TREND_AS_NATIVE_HISTOGRAM=true
fi

echo "── Running Load Test..."
k6 run --env API_URL="${API_URL}" --env SCENARIO="${SCENARIO}" $K6_OUT "${SCRIPT_DIR}/load-test.ts"
