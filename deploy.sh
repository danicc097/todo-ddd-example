#!/bin/bash

docker swarm init 2>/dev/null || true
# 'attachable' allows containers to talk to swarm
docker network create --driver overlay --attachable myapp-net 2>/dev/null || true

docker compose -f docker-compose.infra.yml up -d --remove-orphans

echo "Waiting for database to be ready..."
MAX_RETRIES=30
COUNT=0
until [ "$(docker inspect -f '{{.State.Running}}' myapp-db 2>/dev/null)" == "true" ] &&
	docker exec myapp-db pg_isready -U postgres || [ $COUNT -eq $MAX_RETRIES ]; do
	sleep 2
	((COUNT++))
	echo "Attempt $COUNT/$MAX_RETRIES..."
done

if [ $COUNT -eq $MAX_RETRIES ]; then
	echo "Error: Database did not become ready."
	docker logs myapp-db
	exit 1
fi

set -e

go mod download

make gen

make migrate-up

docker build --target prod -t myapp-go:latest .

docker stack deploy -c docker-compose.app.yml myapp
docker service update --force myapp_go-app
