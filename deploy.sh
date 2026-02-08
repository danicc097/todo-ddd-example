#!/bin/bash
set -e

STACK=myapp

go mod tidy

go generate ./...

docker build -t "$STACK"-go:latest .

docker swarm init 2>/dev/null || true

docker stack deploy -c docker-compose.yml "$STACK"

echo "Check status with: docker service ls"
