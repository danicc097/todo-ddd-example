#!/bin/bash
set -euo pipefail

RED='\033[0;31m'
GRN='\033[0;32m'
NC='\033[0m'

DEBUG=0

while [[ $# -gt 0 ]]; do
	case "$1" in
	-d | --debug)
		DEBUG=1
		shift
		;;
	*)
		shift
		;;
	esac
done

NAMESPACE="myapp"
ERRORS=0

log() { echo "==> $*"; }
pass() { echo -e "${GRN}  ✓ $*${NC}"; }
fail() {
	echo -e "${RED}  ✗ $*${NC}"
	((ERRORS++)) || true
}

if [ "$DEBUG" = "1" ]; then
	REDIR="/dev/stderr"
else
	REDIR="/dev/null"
fi

log "Validating Helm chart..."
if helm template myapp ./chart >"$REDIR" 2>&1; then
	pass "Helm chart renders successfully"
else
	fail "Helm chart failed to render"
	[ "$DEBUG" = "1" ] && helm template myapp ./chart 2>&1 || true
fi

log "Validating all values files..."
for f in chart/values*.yaml; do
	name=$(basename "$f")
	if helm template myapp ./chart -f "$f" >"$REDIR" 2>&1; then
		pass "Renders with $name"
	else
		fail "Failed to render with $name"
	fi
done

log "Linting Helm chart..."
if helm lint ./chart >"$REDIR" 2>&1; then
	pass "Helm lint passed"
else
	fail "Helm lint failed"
	[ "$DEBUG" = "1" ] && helm lint ./chart 2>&1 || true
fi

if kubectl cluster-info --context "kind-myapp" >"$REDIR" 2>&1; then
	log "Checking cluster health..."

	DEPLOYMENTS=$(kubectl get deployments -n "$NAMESPACE" -o jsonpath='{.items[*].metadata.name}' 2>"$REDIR" || echo "")
	if [ -n "$DEPLOYMENTS" ]; then
		for dep in $DEPLOYMENTS; do
			AVAILABLE=$(kubectl get deployment "$dep" -n "$NAMESPACE" -o jsonpath='{.status.availableReplicas}' 2>"$REDIR" || echo "0")
			DESIRED=$(kubectl get deployment "$dep" -n "$NAMESPACE" -o jsonpath='{.spec.replicas}' 2>"$REDIR" || echo "0")
			if [ "${AVAILABLE:-0}" = "$DESIRED" ] && [ "$DESIRED" != "0" ]; then
				pass "Deployment '$dep': ${AVAILABLE}/${DESIRED} replicas available"
			else
				fail "Deployment '$dep': ${AVAILABLE:-0}/${DESIRED} replicas available"
			fi
		done
	fi

	STATEFULSETS=$(kubectl get statefulsets -n "$NAMESPACE" -o jsonpath='{.items[*].metadata.name}' 2>"$REDIR" || echo "")
	if [ -n "$STATEFULSETS" ]; then
		for sts in $STATEFULSETS; do
			READY=$(kubectl get statefulset "$sts" -n "$NAMESPACE" -o jsonpath='{.status.readyReplicas}' 2>"$REDIR" || echo "0")
			DESIRED=$(kubectl get statefulset "$sts" -n "$NAMESPACE" -o jsonpath='{.spec.replicas}' 2>"$REDIR" || echo "0")
			if [ "${READY:-0}" = "$DESIRED" ] && [ "$DESIRED" != "0" ]; then
				pass "StatefulSet '$sts': ${READY}/${DESIRED} replicas ready"
			else
				fail "StatefulSet '$sts': ${READY:-0}/${DESIRED} replicas ready"
			fi
		done
	fi

	NOT_RUNNING=$(kubectl get pods -n "$NAMESPACE" --field-selector=status.phase!=Running,status.phase!=Succeeded -o name 2>"$REDIR" || true)
	if [ -z "$NOT_RUNNING" ]; then
		pass "All pods are running"
	else
		fail "Some pods are not running: $NOT_RUNNING"
	fi

	POSTGRES_POD=$(kubectl get pod -n "$NAMESPACE" -l app.kubernetes.io/name=postgresql -o jsonpath='{.items[0].metadata.name}' 2>"$REDIR" || echo "")
	if [ -z "$POSTGRES_POD" ]; then

		POSTGRES_POD=$(kubectl get pod -n "$NAMESPACE" -l app.kubernetes.io/name=postgres -o jsonpath='{.items[0].metadata.name}' 2>"$REDIR" || echo "")
	fi
	if [ -n "$POSTGRES_POD" ]; then
		if kubectl exec -n "$NAMESPACE" "$POSTGRES_POD" -- pg_isready -U postgres >"$REDIR" 2>&1; then
			pass "Postgres is accepting connections"
		else
			fail "Postgres is not accepting connections"
		fi
	fi

	REDIS_POD=$(kubectl get pod -n "$NAMESPACE" -l app.kubernetes.io/name=redis -o jsonpath='{.items[0].metadata.name}' 2>"$REDIR" || echo "")
	if [ -n "$REDIS_POD" ]; then
		if kubectl exec -n "$NAMESPACE" "$REDIS_POD" -- redis-cli ping 2>"$REDIR" | grep -q PONG; then
			pass "Redis is responding"
		else
			fail "Redis is not responding"
		fi
	fi

	RABBITMQ_POD=$(kubectl get pod -n "$NAMESPACE" -l app.kubernetes.io/name=rabbitmq -o jsonpath='{.items[0].metadata.name}' 2>"$REDIR" || echo "")
	if [ -n "$RABBITMQ_POD" ]; then
		if kubectl exec -n "$NAMESPACE" "$RABBITMQ_POD" -- sh -c "nc -z localhost 5672" >"$REDIR" 2>&1; then
			pass "RabbitMQ is responding"
		else
			fail "RabbitMQ is not responding"
		fi
	fi

	if kubectl get ingress -n "$NAMESPACE" myapp-ingress >"$REDIR" 2>&1; then
		pass "Ingress resource exists"
	else
		fail "Ingress resource not found"
	fi

	CRASH_PODS=$(kubectl get pods -n "$NAMESPACE" -o jsonpath='{range .items[*]}{range .status.containerStatuses[*]}{.state.waiting.reason}{"\n"}{end}{end}' 2>"$REDIR" | grep -c "CrashLoopBackOff" || true)
	if [ "$CRASH_PODS" = "0" ]; then
		pass "No pods in CrashLoopBackOff"
	else
		fail "$CRASH_PODS pod(s) in CrashLoopBackOff"
	fi

	log "Checking service accessibility from host..."

	API_URL="${API_URL:-http://127.0.0.1:8090}"

	if curl -sf --max-time 5 "${API_URL}/healthz" >"$REDIR" 2>&1; then
		pass "API healthz endpoint accessible at ${API_URL}/healthz"
	else
		fail "API healthz endpoint not accessible at ${API_URL}/healthz"
	fi

	if curl -sf --max-time 5 "${API_URL}/api/v1/docs" >"$REDIR" 2>&1; then
		pass "API docs accessible at ${API_URL}/api/v1/docs"
	else
		fail "API docs not accessible at ${API_URL}/api/v1/docs"
	fi

	JAEGER_URL="${API_URL/127.0.0.1/jaeger.localhost}"
	if curl -sf --max-time 5 "$JAEGER_URL" >"$REDIR" 2>&1; then
		pass "Jaeger UI accessible at $JAEGER_URL"
	else
		fail "Jaeger UI not accessible at $JAEGER_URL"
	fi

	PROM_URL="${API_URL/127.0.0.1/prometheus.localhost}"
	if curl -sf --max-time 5 "${PROM_URL}/-/ready" >"$REDIR" 2>&1; then
		pass "Prometheus accessible at $PROM_URL"
	else
		fail "Prometheus not accessible at $PROM_URL"
	fi

	RABBIT_URL="${API_URL/127.0.0.1/rabbitmq.localhost}"
	if curl -sf --max-time 5 "$RABBIT_URL" >"$REDIR" 2>&1; then
		pass "RabbitMQ Management UI accessible at $RABBIT_URL"
	else
		fail "RabbitMQ Management UI not accessible at $RABBIT_URL"
	fi
else
	log "Skipping cluster health checks (no cluster running)"
fi

echo ""
if [ $ERRORS -eq 0 ]; then
	echo -e "${GRN}All validations passed.${NC}"
else
	echo -e "${RED}$ERRORS validation(s) failed.${NC}"
	exit 1
fi
