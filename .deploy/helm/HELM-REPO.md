# 📦 Ventros CRM Helm Repository

O Helm chart do Ventros CRM está disponível via GitHub Pages.

## 🚀 Quick Start

### Adicionar o repositório Helm

```bash
helm repo add ventros https://leonardocaloi.github.io/ventros-crm/charts/
helm repo update
```

### Instalar o chart

```bash
# Instalação básica
helm install ventros-crm ventros/ventros-crm \
  --namespace ventros-crm \
  --create-namespace

# Com valores customizados
helm install ventros-crm ventros/ventros-crm \
  --namespace ventros-crm \
  --create-namespace \
  --values values-dev.yaml
```

### Verificar versões disponíveis

```bash
helm search repo ventros
```

## 🔗 Links Úteis

- **Helm Repository**: https://leonardocaloi.github.io/ventros-crm/charts/
- **Repository Index**: https://leonardocaloi.github.io/ventros-crm/charts/index.yaml
- **GitHub Releases**: https://github.com/leonardocaloi/ventros-crm/releases
- **GitHub Repository**: https://github.com/leonardocaloi/ventros-crm

## 📥 Download Manual

Se preferir baixar o chart diretamente:

```bash
# Download da última versão
wget https://github.com/leonardocaloi/ventros-crm/releases/download/v0.1.0/ventros-crm-0.1.0.tgz

# Instalar do arquivo local
helm install ventros-crm ./ventros-crm-0.1.0.tgz \
  --namespace ventros-crm \
  --create-namespace
```

## 🔄 Atualização do Repositório

O repositório Helm é atualizado automaticamente via GitHub Actions quando:
1. Um push é feito na branch `main`
2. A versão no `Chart.yaml` é diferente da última release

O workflow:
1. Empacota o chart com todas as dependências
2. Cria uma GitHub Release com o arquivo `.tgz`
3. Atualiza o índice do Helm repository na branch `gh-pages`

## 🛠️ Uso Avançado

### Listar todas as versões

```bash
helm search repo ventros --versions
```

### Instalar versão específica

```bash
helm install ventros-crm ventros/ventros-crm \
  --version 0.1.0 \
  --namespace ventros-crm \
  --create-namespace
```

### Upgrade

```bash
helm upgrade ventros-crm ventros/ventros-crm \
  --namespace ventros-crm \
  --values values-dev.yaml
```

### Desinstalar

```bash
helm uninstall ventros-crm --namespace ventros-crm
```

## 🐛 Troubleshooting

### Repositório não encontrado

Se o comando `helm repo add` falhar, verifique:

1. **GitHub Pages está ativo**: https://github.com/leonardocaloi/ventros-crm/settings/pages
2. **Index existe**: https://leonardocaloi.github.io/ventros-crm/charts/index.yaml
3. **Releases existem**: https://github.com/leonardocaloi/ventros-crm/releases

### Chart não aparece no search

```bash
# Limpar cache e atualizar
helm repo remove ventros
helm repo add ventros https://leonardocaloi.github.io/ventros-crm/charts/
helm repo update
helm search repo ventros
```

### Forçar atualização do repositório

Se você fez uma nova release mas o chart não aparece:

```bash
# Trigger manual do workflow
gh workflow run helm-release.yaml

# Ou faça um commit vazio para trigger
git commit --allow-empty -m "chore: trigger helm release"
git push
```
