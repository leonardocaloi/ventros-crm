#!/bin/bash
set -e

echo "🔥 Limpando TUDO do Ventros CRM..."

# 1. Desinstala Helm
echo "📦 Desinstalando Helm release..."
helm uninstall ventros-crm -n ventros-crm 2>/dev/null || echo "Release já foi removido"

# 2. Aguarda pods terminarem
echo "⏳ Aguardando pods terminarem..."
sleep 5

# 3. Deleta PVCs (dados persistentes)
echo "💾 Deletando PVCs (dados antigos)..."
kubectl delete pvc --all -n ventros-crm 2>/dev/null || echo "Nenhum PVC encontrado"

# 4. Deleta namespace
echo "🗑️  Deletando namespace..."
kubectl delete namespace ventros-crm 2>/dev/null || echo "Namespace já foi removido"

# 5. Aguarda namespace ser deletado completamente
echo "⏳ Aguardando namespace ser deletado..."
while kubectl get namespace ventros-crm &>/dev/null; do
    echo "   Aguardando..."
    sleep 2
done

