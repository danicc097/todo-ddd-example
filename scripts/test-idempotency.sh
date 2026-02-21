#!/bin/bash

set -euo pipefail

export CLICOLOR_FORCE=1

SEED="idemp-test-$(date +%s)"
ALICE_EMAIL="alice-${SEED}@example.com"
ALICE_NAME="Alice"
PASSWORD="Password123!"
WORKSPACE_NAME="Workspace ${SEED}"

./todo-cli -d register -p "{\"email\": \"${ALICE_EMAIL}\", \"name\": \"${ALICE_NAME}\", \"password\": \"${PASSWORD}\"}"

export API_TOKEN=$(./todo-cli -d login -p "{\"email\": \"${ALICE_EMAIL}\", \"password\": \"${PASSWORD}\"}" | jq -r .accessToken)

WS_ID=$(./todo-cli -d onboard-workspace -p "{\"name\": \"${WORKSPACE_NAME}\"}" | jq -r .id)

IDEMP_KEY=$(uuidgen)

TODO_TITLE="Idempotent Task"
RES1=$(./todo-cli -d create-todo "$WS_ID" -p "{\"title\": \"${TODO_TITLE}\"}" --idempotency-key "$IDEMP_KEY")
TODO_ID1=$(echo "$RES1" | jq -r .id)

RES2=$(./todo-cli -d create-todo "$WS_ID" -p "{\"title\": \"${TODO_TITLE}\"}" --idempotency-key "$IDEMP_KEY")
TODO_ID2=$(echo "$RES2" | jq -r .id)

if [ "$TODO_ID1" != "$TODO_ID2" ]; then
	echo "Error: todo IDs do not match for same idempotency key"
	exit 1
fi

COUNT=$(./todo-cli -d get-workspace-todos "$WS_ID" | jq '. | length')

if [ "$COUNT" != "1" ]; then
	echo "Error: expected 1 todo, found $COUNT"
	exit 1
fi
