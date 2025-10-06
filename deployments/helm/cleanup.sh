#!/bin/bash
# Script de limpeza completa do Helm + CRDs
# Uso: ./cleanup.sh [namespace]

set -e

NAMESPACE="${1:-ventros-crm}"

echo "🧹 Limpando namespace: $NAMESPACE"
echo "=================================="

# 1. Uninstall Helm release
echo ""
echo "📦 1. Removendo Helm release..."
if helm list -n $NAMESPACE | grep -q ventros-crm; then
    helm uninstall ventros-crm -n $NAMESPACE --wait --timeout 5m || true
    echo "✅ Helm release removido"
else
    echo "⚠️  Nenhum release encontrado"
fi

# 2. Deletar RabbitMQ Clusters (CRD)
echo ""
echo "🐰 2. Removendo RabbitMQ Clusters..."
if kubectl get rabbitmqcluster -n $NAMESPACE 2>/dev/null | grep -q ventros-crm; then
    # Remover finalizers primeiro para evitar travamento
    echo "   Removendo finalizers..."
    kubectl patch rabbitmqcluster --all -n $NAMESPACE \
        --type json \
        -p='[{"op": "remove", "path": "/metadata/finalizers"}]' 2>/dev/null || true
    
    # Agora deletar
    kubectl delete rabbitmqcluster --all -n $NAMESPACE --wait=false || true
    echo "✅ RabbitMQ Clusters removidos"
else
    echo "⚠️  Nenhum RabbitMQ Cluster encontrado"
fi

# 3. Deletar PostgreSQL Clusters (CRD)
echo ""
echo "🐘 3. Removendo PostgreSQL Clusters..."
if kubectl get postgresql -n $NAMESPACE 2>/dev/null | grep -q ventros-crm; then
    # Remover finalizers primeiro para evitar travamento
    echo "   Removendo finalizers..."
    kubectl patch postgresql --all -n $NAMESPACE \
        --type json \
        -p='[{"op": "remove", "path": "/metadata/finalizers"}]' 2>/dev/null || true
    
    # Agora deletar
    kubectl delete postgresql --all -n $NAMESPACE --wait=false || true
    echo "✅ PostgreSQL Clusters removidos"
else
    echo "⚠️  Nenhum PostgreSQL Cluster encontrado"
fi

# 4. Deletar Jobs órfãos
echo ""
echo "⚙️  4. Removendo Jobs..."
kubectl delete jobs --all -n $NAMESPACE --wait --timeout=1m || true
echo "✅ Jobs removidos"

# 5. Deletar Pods órfãos
echo ""
echo "🔄 5. Removendo Pods órfãos..."
kubectl delete pods --all -n $NAMESPACE --grace-period=0 --force || true
echo "✅ Pods removidos"

# 6. Deletar PVCs (opcional - cuidado com dados!)
echo ""
read -p "❓ Deletar PVCs (dados serão perdidos)? [y/N] " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    kubectl delete pvc --all -n $NAMESPACE --wait --timeout=2m || true
    echo "✅ PVCs removidos"
else
    echo "⏭️  PVCs mantidos"
fi

# 7. Deletar Secrets do Helm
echo ""
echo "🔐 7. Removendo Secrets do Helm..."
kubectl delete secrets -n $NAMESPACE -l owner=helm || true
echo "✅ Secrets do Helm removidos"

# 8. Verificar recursos restantes
echo ""
echo "🔍 8. Verificando recursos restantes..."
echo ""
kubectl get all,pvc,secrets,configmaps -n $NAMESPACE 2>/dev/null || echo "Namespace vazio ou não existe"

echo ""
echo "=================================="
echo "✅ Limpeza concluída!"
echo ""
echo "Para reinstalar:"
echo "  helm install ventros-crm ./deployments/helm/ventros-crm \\"
echo "    --namespace $NAMESPACE \\"
echo "    --create-namespace \\"
echo "    --values deployments/helm/ventros-crm/values-dev.yaml \\"
echo "    --timeout 15m \\"
echo "    --wait"
