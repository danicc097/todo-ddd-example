#!/bin/bash
set -e

# 'attachable' allows containers to talk to swarm
docker network create --driver overlay --attachable myapp-net 2>/dev/null || true
docker swarm init 2>/dev/null || true

docker compose -f docker-compose.infra.yml up -d --remove-orphans

echo "Waiting for database to be ready..."
MAX_RETRIES=30
COUNT=0
until docker exec myapp-db pg_isready -U postgres >/dev/null 2>&1 || [ $COUNT -eq $MAX_RETRIES ]; do
	sleep 2
	((COUNT++))
	echo -n "."
done

if [ $COUNT -eq $MAX_RETRIES ]; then
	echo "Error: Database did not become ready."
	exit 1
fi

go mod tidy
make gen

make migrate-up

docker build --target dev -t myapp-go:latest .

docker stack deploy -c docker-compose.app.yml myapp
docker service update --force myapp_go-app
