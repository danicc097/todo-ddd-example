#!/bin/bash
set -e

STACK=myapp

docker swarm init 2>/dev/null || true
docker network create --driver overlay myapp-net 2>/dev/null || true

docker stack deploy -c docker-compose.yml "$STACK"

MAX_RETRIES=30
COUNT=0
until docker exec "$(docker ps -q -f name="$STACK"_db)" pg_isready -U postgres || [ $COUNT -eq $MAX_RETRIES ]; do
	sleep 2
	((COUNT++))
done

if [ $COUNT -eq $MAX_RETRIES ]; then
	exit 1
fi

go mod tidy
make gen-sqlc
go generate ./...

docker build -t "$STACK"-go:latest .
docker service update --image "$STACK"-go:latest --force "$STACK"_go-app
