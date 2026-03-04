#!/usr/bin/env bash

set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCENARIO="${SCENARIO:-load}"
API_URL="${API_URL:-http://127.0.0.1:8090}"

for dep in k6 curl jq uuidgen; do
	command -v "$dep" >/dev/null || { echo "✗ Missing: $dep" && exit 1; }
done

USERS_FILE="${SCRIPT_DIR}/users.json"
if [ ! -f "$USERS_FILE" ] || [ "${RESEED:-0}" = "1" ]; then
	API_URL="$API_URL" bash "${SCRIPT_DIR}/seed-users.sh"
else
	echo "── Using $(jq 'length' "$USERS_FILE") existing users (RESEED=1 to refresh)"
fi

mkdir -p "${SCRIPT_DIR}/results"
RESULT_FILE="${SCRIPT_DIR}/results/${SCENARIO}-$(date +%Y%m%dT%H%M%S).json"

# https://grafana.com/docs/k6/latest/results-output/real-time/prometheus-remote-write/
K6_OUT="--out json=${RESULT_FILE}"
if [ "${K6_PUSH_PROMETHEUS:-0}" = "1" ]; then
	K6_OUT="${K6_OUT} --out experimental-prometheus-rw"
	export K6_PROMETHEUS_RW_SERVER_URL="${K6_PROMETHEUS_RW_URL:-http://127.0.0.1:9090/api/v1/write}"
	export K6_PROMETHEUS_RW_TREND_AS_NATIVE_HISTOGRAM=true
fi

k6 run --env API_URL="${API_URL}" --env SCENARIO="${SCENARIO}" $K6_OUT "${SCRIPT_DIR}/load-test.js"
