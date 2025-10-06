#!/bin/bash
set -e

NAMESPACE="${NAMESPACE:-ventros-crm}"
RELEASE_NAME="${RELEASE_NAME:-ventros-crm}"
VALUES_FILE="${VALUES_FILE:-values-dev.yaml}"

echo "🚀 Installing Ventros CRM with proper dependency handling..."
echo ""

# Step 1: Install without Temporal
echo "📦 Step 1/3: Installing infrastructure (PostgreSQL, Redis, RabbitMQ)..."
helm install "$RELEASE_NAME" . \
  -n "$NAMESPACE" \
  --create-namespace \
  -f "$VALUES_FILE" \
  --set temporal.enabled=false \
  --set replicaCount=0 \
  --wait \
  --timeout 10m

echo "✅ Infrastructure installed!"
echo ""

# Step 2: Wait for PostgreSQL secret
echo "⏳ Step 2/3: Waiting for PostgreSQL to be ready..."
kubectl wait --for=condition=ready pod \
  -l application=ventros-crm,cluster-name=ventros-crm-postgres \
  -n "$NAMESPACE" \
  --timeout=5m

echo "✅ PostgreSQL is ready!"
echo ""

# Step 3: Upgrade to enable Temporal and API
echo "📦 Step 3/3: Enabling Temporal and API..."
helm upgrade "$RELEASE_NAME" . \
  -n "$NAMESPACE" \
  -f "$VALUES_FILE" \
  --wait \
  --timeout 10m

echo ""
echo "🎉 Ventros CRM installed successfully!"
echo ""
echo "📊 Check status:"
echo "  kubectl get pods -n $NAMESPACE"
echo ""
echo "🔗 Access the API:"
echo "  kubectl port-forward -n $NAMESPACE svc/ventros-crm 8080:8080"
echo ""
