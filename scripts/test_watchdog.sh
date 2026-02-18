#!/usr/bin/env bash

LOCKDIR="/tmp/todo-ddd-test-watchdog.lock"
if ! mkdir "$LOCKDIR" 2>/dev/null; then
	exit 0
fi

trap 'rmdir "$LOCKDIR"' EXIT

TIMEOUT_SEC=300

YELLOW='\033[1;33m'
NC='\033[0m'

while true; do
	if pgrep -f "dlv|\.test|go test" >/dev/null; then
		date +%s >.test_last_run
	fi

	LAST_RUN=$(cat .test_last_run 2>/dev/null || echo 0)
	NOW=$(date +%s)

	if [ $((NOW - LAST_RUN)) -ge $TIMEOUT_SEC ]; then
		CONTAINERS=$(docker ps -a -q --filter "label=todo-ddd-test=true")

		if [ -n "$CONTAINERS" ]; then
			echo -e "\n${YELLOW}[Watchdog] ${TIMEOUT_SEC}s of test inactivity. Destroying test containers...${NC}"
			echo "$CONTAINERS" | xargs docker rm -f -v >/dev/null 2>&1
		fi

		rm -f .test_last_run
		exit 0
	fi
	sleep 60
done
