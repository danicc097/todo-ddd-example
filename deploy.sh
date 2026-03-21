#!/bin/bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

CLUSTER_NAME="myapp"
NAMESPACE="myapp"
HELM_VALUES_FILE="${HELM_VALUES:-chart/values.yaml}"

log() { echo "==> $*"; }

for cmd in kind kubectl docker helm; do
	if ! command -v "$cmd" &>/dev/null; then
		echo "Error: '$cmd' is required but not installed." && exit 1
	fi
done

if ! kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
	log "Creating kind cluster '${CLUSTER_NAME}'..."
	kind create cluster --config k8s/kind-config.yaml --wait 5m
else
	log "Kind cluster '${CLUSTER_NAME}' already exists."
fi

kubectl cluster-info --context "kind-${CLUSTER_NAME}" >/dev/null

log "Waiting for node to be Ready..."
kubectl wait --for=condition=Ready node --all --timeout=300s

log "Updating Helm repos..."
helm repo list | grep -q "^bitnami\s" || helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo list | grep -q "^traefik\s" || helm repo add traefik https://traefik.github.io/charts
helm repo update bitnami traefik

log "Installing/updating Traefik Ingress Controller..."
helm upgrade --install traefik traefik/traefik \
	--namespace traefik --create-namespace \
	-f k8s/traefik-values.yaml \
	--wait --timeout 5m

log "Updating Helm chart dependencies..."
helm dependency update ./chart

# Extract chart tarballs (required for Helm 3.20+ with OCI registries)
cd chart/charts
for f in *.tgz; do [ -f "$f" ] && tar xzf "$f" && rm -f "$f"; done
cd "$SCRIPT_DIR"

log "Deploying infrastructure via Helm..."
helm upgrade --install myapp ./chart \
	-f "$HELM_VALUES_FILE" \
	--namespace "$NAMESPACE" --create-namespace \
	--set goApp.replicaCount=0 \
	--wait --timeout 5m

go mod download
make gen

log "Running database migrations..."
make migrate-up

log "Building application image..."
docker build --target prod -t myapp-go:latest .

log "Loading image into kind cluster..."
kind load docker-image myapp-go:latest --name "$CLUSTER_NAME"

if ! kubectl get secret jwt-keys -n "$NAMESPACE" &>/dev/null; then
	if [ ! -f private.pem ] || [ ! -f public.pem ]; then
		log "Generating JWT key pair..."
		openssl genpkey -algorithm RSA -out private.pem -pkeyopt rsa_keygen_bits:2048
		openssl rsa -pubout -in private.pem -out public.pem
	fi
	log "Creating JWT keys secret..."
	kubectl create secret generic jwt-keys \
		--from-file=private.pem=./private.pem \
		--from-file=public.pem=./public.pem \
		-n "$NAMESPACE"
fi

log "Creating OpenAPI spec configmap..."
kubectl create configmap openapi-spec \
	--from-file=openapi.yaml=./openapi.yaml \
	-n "$NAMESPACE" \
	--dry-run=client -o yaml | kubectl apply -f -

log "Deploying application via Helm..."
helm upgrade --install myapp ./chart \
	-f "$HELM_VALUES_FILE" \
	--namespace "$NAMESPACE" \
	--wait --timeout 5m

log "Verifying all pods are ready..."
kubectl wait --for=condition=ready pod --all -n "$NAMESPACE" --timeout=120s

echo ""
echo "============================================"
echo "  Deployment complete!"
echo "============================================"
echo "  API docs:    http://127.0.0.1:8090/api/v1/docs"
echo "  Jaeger UI:   http://jaeger.localhost:8090"
echo "  Prometheus:  http://prometheus.localhost:8090"
echo "  RabbitMQ:    http://rabbitmq.localhost:8090"
echo "  Postgres:    127.0.0.1:5732 (via NodePort)"
echo "============================================"
