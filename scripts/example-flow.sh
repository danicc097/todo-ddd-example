#!/bin/bash

set -euo pipefail

if ! command -v oathtool &>/dev/null; then
	echo "package oath-toolkit is required."
	exit 1
fi

export CLICOLOR_FORCE=1 # lipgloss inside subshells

SEED="mfa-test-$(date +%s)" # we delete ws at the end, cannot reuse idemp key

echo "Using seed: $SEED"

seeded_uuid() {
	uuidgen --md5 --namespace @dns --name "${SEED}-$1"
}

echo -e "\n=== REGISTER ALICE ==="
ALICE_RES=$(./todo-cli register -d -p '{"email": "alice-'"$SEED"'@example.com", "name": "Alice", "password": "Password123!"}' --idempotency-key "$(seeded_uuid "alice-reg")")
ALICE_ID=$(echo "$ALICE_RES" | jq -r .id)
echo "Alice ID: $ALICE_ID"

echo -e "\n=== LOGIN ALICE ==="
LOGIN_RES=$(./todo-cli login -d -p '{"email": "alice-'"$SEED"'@example.com", "password": "Password123!"}')
export API_TOKEN=$(echo "$LOGIN_RES" | jq -r .accessToken)

echo -e "\n=== CREATE WORKSPACE ==="
WS_RES=$(./todo-cli onboard-workspace -d -p '{"name": "Alice HQ", "description": "Top Secret"}' --idempotency-key "$(seeded_uuid "alice-ws")")
WS_ID=$(echo "$WS_RES" | jq -r .id)
echo "Workspace ID: $WS_ID"

echo -e "\n=== CREATE TODO ==="
TODO_RES=$(./todo-cli create-todo "$WS_ID" -d -p '{"title": "Deploy to production"}' --idempotency-key "$(seeded_uuid "todo-1")")
TODO_ID=$(echo "$TODO_RES" | jq -r .id)
echo "Todo ID: $TODO_ID"

echo -e "\n=== CREATE TAG ==="
TAG_RES=$(./todo-cli create-tag "$WS_ID" -d -p '{"name": "critical-'"${SEED:0:10}"'"}' --idempotency-key "$(seeded_uuid "tag-crit")")
TAG_ID=$(echo "$TAG_RES" | jq -r .id)
echo "Tag ID: $TAG_ID"

echo -e "\n=== ASSIGN TAG TO TODO ==="
./todo-cli assign-tag-to-todo "$TODO_ID" -d -p "{\"tagId\": \"$TAG_ID\"}" --idempotency-key "$(seeded_uuid "assign-crit")" >/dev/null
echo "Tag assigned"

echo -e "\n=== REGISTER BOB ==="
BOB_RES=$(./todo-cli register -d -p '{"email": "bob-'"$SEED"'@example.com", "name": "Bob", "password": "Password123!"}' --idempotency-key "$(seeded_uuid "bob-reg")")
BOB_ID=$(echo "$BOB_RES" | jq -r .id)
echo "Bob ID: $BOB_ID"

echo -e "\n=== ADD BOB TO WORKSPACE ==="
role="MEMBER"
./todo-cli add-workspace-member "$WS_ID" -d -p "{\"userId\": \"$BOB_ID\", \"role\": \"$role\"}" --idempotency-key "$(seeded_uuid "bob-ws")" >/dev/null
echo "Bob added as $role"

echo -e "\n=== ATTEMPT TO DELETE WORKSPACE (WITHOUT MFA) ==="
if ./todo-cli delete-workspace "$WS_ID" -d; then
	echo "Error: Deleting workspace without MFA should have failed!"
	exit 1
fi
echo "Deletion failed as expected"

echo -e "\n===  INITIATE TOTP ==="
TOTP_RES=$(./todo-cli initiate-totp -d)
URI=$(echo "$TOTP_RES" | jq -r .provisioningUri)
SECRET=$(echo "$URI" | sed -n 's/.*secret=\([^&]*\).*/\1/p')
echo "TOTP secret: $SECRET"

echo -e "\n===  GENERATE TOTP CODE ==="
CODE=$(oathtool --totp -b "$SECRET")
echo "Code: $CODE"

echo -e "\n===  VERIFY TOTP CODE ==="
VERIFY_RES=$(./todo-cli verify-totp -d -p "{\"code\": \"$CODE\"}")
export API_TOKEN=$(echo "$VERIFY_RES" | jq -r .accessToken)
echo "MFA Token acquired"

echo -e "\n===  DELETE WORKSPACE (WITH MFA) ==="
./todo-cli delete-workspace "$WS_ID" -d
echo "Workspace deleted"
