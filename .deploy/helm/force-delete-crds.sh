#!/bin/bash
# Script rápido para forçar deleção de CRDs travados
# Uso: ./force-delete-crds.sh [namespace]

NAMESPACE="${1:-ventros-crm}"

echo "🔥 Forçando deleção de CRDs no namespace: $NAMESPACE"

# RabbitMQ
if kubectl get rabbitmqcluster -n $NAMESPACE 2>/dev/null | grep -q .; then
    echo "🐰 Removendo finalizers do RabbitMQ..."
    kubectl patch rabbitmqcluster --all -n $NAMESPACE \
        --type json \
        -p='[{"op": "remove", "path": "/metadata/finalizers"}]' 2>/dev/null || true
    kubectl delete rabbitmqcluster --all -n $NAMESPACE --wait=false 2>/dev/null || true
fi

# PostgreSQL
if kubectl get postgresql -n $NAMESPACE 2>/dev/null | grep -q .; then
    echo "🐘 Removendo finalizers do PostgreSQL..."
    kubectl patch postgresql --all -n $NAMESPACE \
        --type json \
        -p='[{"op": "remove", "path": "/metadata/finalizers"}]' 2>/dev/null || true
    kubectl delete postgresql --all -n $NAMESPACE --wait=false 2>/dev/null || true
fi

echo "✅ Finalizers removidos! Aguarde alguns segundos..."
sleep 3

echo ""
echo "Recursos restantes:"
kubectl get all,postgresql,rabbitmqcluster -n $NAMESPACE 2>/dev/null || echo "Namespace limpo!"
