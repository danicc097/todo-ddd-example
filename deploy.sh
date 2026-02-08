#!/bin/bash
set -e

go mod tidy

go generate ./...

docker build -t myapp-go:latest .

docker swarm init 2>/dev/null || true

docker stack deploy -c docker-compose.yml myapp

echo "Check status with: docker service ls"
