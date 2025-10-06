# 🚀 Helm Chart Release - Quick Start

## ✅ Melhor Prática (Oficial do Helm)

Usamos a **action oficial do Helm**: [`helm/chart-releaser-action`](https://github.com/helm/chart-releaser-action)

### Como Funciona

1. **Você atualiza** a versão no `Chart.yaml`
2. **Faz commit** e push para `main`
3. **GitHub Actions detecta** automaticamente a mudança
4. **Publica automaticamente**:
   - Cria GitHub Release (com tag `v0.1.0`)
   - Empacota o chart (`.tgz`)
   - Publica no GitHub Pages
   - Atualiza `index.yaml`

---

## 📝 Passo a Passo

### 1. Atualizar Versão

```bash
# Editar Chart.yaml
vim .deploy/helm/ventros-crm/Chart.yaml

# Alterar:
version: 0.1.0  →  version: 0.2.0
appVersion: "0.1.0"  →  appVersion: "0.2.0"
```

### 2. Commit e Push

```bash
git add .deploy/helm/ventros-crm/Chart.yaml
git commit -m "chore: bump chart version to 0.2.0"
git push origin main
```

### 3. Aguardar GitHub Actions

- Acesse: `https://github.com/seu-usuario/ventros-crm/actions`
- Workflow: **Release Charts**
- Tempo: ~2-3 minutos

### 4. Verificar Release

```bash
# Ver releases criados
gh release list

# Ou acesse:
# https://github.com/seu-usuario/ventros-crm/releases
```

---

## 🤖 Usando o Script Automatizado

```bash
# Release da versão 0.2.0
./.deploy/helm/release.sh 0.2.0

# Com mensagem customizada
./.deploy/helm/release.sh 0.2.0 "Add Redis clustering support"
```

O script faz tudo automaticamente:
- ✅ Valida a versão
- ✅ Verifica se já existe
- ✅ Faz lint do chart
- ✅ Atualiza `Chart.yaml`
- ✅ Commit e push para `main`

---

## 📦 Como Usuários Instalam

Após o release, qualquer pessoa pode usar:

```bash
# Adicionar repositório
helm repo add ventros https://seu-usuario.github.io/ventros-crm/

# Atualizar repos
helm repo update

# Instalar versão específica
helm install ventros-crm ventros/ventros-crm \
  --version 0.2.0 \
  --namespace ventros-crm \
  --create-namespace

# Listar versões disponíveis
helm search repo ventros/ventros-crm --versions
```

---

## 🔧 Configuração Inicial (Uma Vez)

### 1. Habilitar GitHub Pages

1. Vá em: `Settings` → `Pages`
2. Source: `Deploy from a branch`
3. Branch: `gh-pages` / `root`
4. Clique em `Save`

### 2. Configurar Permissões

1. Vá em: `Settings` → `Actions` → `General`
2. Workflow permissions: `Read and write permissions`
3. Marque: `Allow GitHub Actions to create and approve pull requests`
4. Clique em `Save`

### 3. Fazer Primeiro Release

```bash
# Certifique-se de que Chart.yaml está na versão correta
cat .deploy/helm/ventros-crm/Chart.yaml | grep version

# Se já está em 0.1.0, faça um commit vazio para trigger
git commit --allow-empty -m "chore: trigger initial helm release"
git push origin main

# Ou use o script
./.deploy/helm/release.sh 0.1.0 "Initial release"
```

---

## 🎯 Vantagens desta Abordagem

### ✅ Oficial do Helm
- Mantido pela equipe do Helm
- Usado por milhares de projetos
- Seguindo best practices

### ✅ Automático
- Detecta mudanças automaticamente
- Não precisa criar tags manualmente
- Publica tudo automaticamente

### ✅ Simples
- Apenas atualizar `Chart.yaml`
- Um commit = um release
- Sem scripts complexos

### ✅ Confiável
- Gerencia `index.yaml` automaticamente
- Cria releases consistentes
- Suporta múltiplas versões

---

## 📚 Documentação Completa

Para mais detalhes, veja:
- [RELEASE.md](./RELEASE.md) - Guia completo de release
- [README.md](./ventros-crm/README.md) - Documentação do chart
- [Helm Chart Releaser Action](https://github.com/helm/chart-releaser-action)

---

## 🐛 Troubleshooting

### Workflow não executou

```bash
# Verificar se há mudanças na versão
git log --oneline -n 5

# Verificar workflows
gh workflow list
gh workflow view "Release Charts"
```

### Release não apareceu

```bash
# Verificar logs do workflow
gh run list --workflow="Release Charts"
gh run view <run-id> --log

# Verificar branch gh-pages
git fetch origin gh-pages
git log origin/gh-pages
```

### Chart não aparece no repo

```bash
# Limpar cache local
helm repo remove ventros
rm -rf ~/.cache/helm/repository/ventros-*

# Adicionar novamente
helm repo add ventros https://seu-usuario.github.io/ventros-crm/
helm repo update
helm search repo ventros/ventros-crm
```

---

## 🎉 Pronto!

Agora você tem um **Helm Repository profissional** com releases automatizados! 🚀
