# 🚀 Helm Chart Release Guide

Este guia explica como fazer releases do Helm Chart do Ventros CRM usando GitHub Actions.

## 📋 Pré-requisitos

1. **GitHub Pages habilitado** no repositório:
   - Vá em: `Settings` → `Pages`
   - Source: `Deploy from a branch`
   - Branch: `gh-pages` / `root`
   - Clique em `Save`

2. **Permissões do GitHub Actions**:
   - Vá em: `Settings` → `Actions` → `General`
   - Em "Workflow permissions", selecione: `Read and write permissions`
   - Marque: `Allow GitHub Actions to create and approve pull requests`
   - Clique em `Save`

## 🎯 Como Fazer um Release

### Método 1: Atualizar Chart.yaml e Push (Recomendado) ⭐

A action oficial `helm/chart-releaser-action` detecta automaticamente mudanças de versão:

```bash
# 1. Atualizar a versão no Chart.yaml
cd .deploy/helm/ventros-crm
vim Chart.yaml  # Altere version: 0.1.0 para version: 0.2.0

# 2. Commit e push para main
git add Chart.yaml
git commit -m "chore: bump chart version to 0.2.0"
git push origin main

# 3. O GitHub Actions irá automaticamente:
#    - Detectar a nova versão
#    - Empacotar o Helm Chart
#    - Criar GitHub Release (v0.2.0)
#    - Publicar no GitHub Pages
#    - Atualizar index.yaml
```

### Método 2: Script Automatizado

Use o script `release.sh` para automatizar a atualização:

```bash
# Fazer release da versão 0.2.0
./.deploy/helm/release.sh 0.2.0

# Com mensagem customizada
./.deploy/helm/release.sh 0.2.0 "Add Redis clustering support"
```

### Método 3: Manual via Web UI

1. Edite `.deploy/helm/ventros-crm/Chart.yaml` no GitHub
2. Altere `version: 0.1.0` para `version: 0.2.0`
3. Commit direto na branch `main`
4. GitHub Actions detecta e publica automaticamente

## 📦 O Que Acontece Automaticamente

Quando você cria uma tag `v*.*.*`, o GitHub Actions:

1. ✅ Extrai a versão da tag (ex: `v0.1.0` → `0.1.0`)
2. ✅ Atualiza `Chart.yaml` com a versão correta
3. ✅ Empacota o Helm Chart (`.tgz`)
4. ✅ Cria/atualiza o branch `gh-pages`
5. ✅ Gera o `index.yaml` do Helm Repository
6. ✅ Cria uma página HTML de documentação
7. ✅ Publica no GitHub Pages
8. ✅ Cria um GitHub Release com o `.tgz` anexado

## 🌐 Acessando o Helm Repository

Após o primeiro release, seu Helm Repository estará disponível em:

```
https://<seu-usuario>.github.io/<nome-do-repo>/charts/
```

Exemplo:
```
https://caloi.github.io/ventros-crm/charts/
```

## 📥 Como Outras Pessoas Usam

### Adicionar o Repositório

```bash
helm repo add ventros https://caloi.github.io/ventros-crm/charts/
helm repo update
```

### Instalar o Chart

```bash
# Última versão
helm install ventros-crm ventros/ventros-crm \
  --namespace ventros-crm \
  --create-namespace

# Versão específica
helm install ventros-crm ventros/ventros-crm \
  --version 0.1.0 \
  --namespace ventros-crm \
  --create-namespace
```

### Listar Versões Disponíveis

```bash
helm search repo ventros/ventros-crm --versions
```

### Atualizar para Nova Versão

```bash
helm repo update
helm upgrade ventros-crm ventros/ventros-crm \
  --version 0.2.0 \
  --namespace ventros-crm
```

## 🔄 Versionamento

Siga o [Semantic Versioning](https://semver.org/):

- **MAJOR** (1.0.0): Mudanças incompatíveis na API
- **MINOR** (0.1.0): Novas funcionalidades compatíveis
- **PATCH** (0.0.1): Correções de bugs compatíveis

### Exemplos

```bash
# Primeira versão estável
git tag -a v1.0.0 -m "First stable release"

# Nova funcionalidade
git tag -a v1.1.0 -m "Add Redis clustering support"

# Correção de bug
git tag -a v1.1.1 -m "Fix PostgreSQL connection timeout"

# Breaking change
git tag -a v2.0.0 -m "Migrate to new PostgreSQL operator"
```

## 🐛 Troubleshooting

### GitHub Actions Falhou

1. Verifique os logs em: `Actions` → `Release Helm Chart`
2. Erros comuns:
   - **Permissões**: Verifique workflow permissions
   - **GitHub Pages**: Certifique-se de que está habilitado
   - **Tag inválida**: Use formato `v*.*.*` (ex: `v0.1.0`)

### GitHub Pages Não Atualiza

1. Vá em: `Settings` → `Pages`
2. Verifique se o branch `gh-pages` existe
3. Force um rebuild:
   ```bash
   git checkout gh-pages
   git commit --allow-empty -m "Trigger rebuild"
   git push origin gh-pages
   ```

### Helm Repo Não Encontra o Chart

```bash
# Limpar cache local
helm repo remove ventros
rm -rf ~/.cache/helm/repository/ventros-*

# Adicionar novamente
helm repo add ventros https://caloi.github.io/ventros-crm/charts/
helm repo update
```

## 📊 Verificar Status do Release

### Via GitHub CLI

```bash
# Listar releases
gh release list

# Ver detalhes de um release
gh release view v0.1.0

# Baixar assets de um release
gh release download v0.1.0
```

### Via Web

1. Acesse: `https://github.com/<usuario>/<repo>/releases`
2. Veja todos os releases publicados
3. Baixe os arquivos `.tgz` diretamente

## 🔗 Links Úteis

- [Helm Chart Best Practices](https://helm.sh/docs/chart_best_practices/)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [GitHub Pages Documentation](https://docs.github.com/en/pages)
- [Semantic Versioning](https://semver.org/)

## 📝 Checklist de Release

Antes de fazer um release, verifique:

- [ ] Código testado e funcionando
- [ ] `CHANGELOG.md` atualizado
- [ ] `Chart.yaml` com descrição correta
- [ ] `values.yaml` com valores padrão sensatos
- [ ] `README.md` do chart atualizado
- [ ] Testes do Helm passando (`helm lint`, `helm template`)
- [ ] Versão segue Semantic Versioning
- [ ] Tag criada no formato correto (`v*.*.*`)

## 🎉 Primeiro Release

Para fazer seu primeiro release:

```bash
# 1. Habilitar GitHub Pages (via web UI)
# 2. Configurar permissões do Actions (via web UI)

# 3. Criar primeira tag
git tag -a v0.1.0 -m "Initial release"
git push origin v0.1.0

# 4. Aguardar GitHub Actions terminar (~2-3 minutos)

# 5. Verificar
helm repo add ventros https://caloi.github.io/ventros-crm/charts/
helm search repo ventros/ventros-crm

# 6. Testar instalação
helm install test-crm ventros/ventros-crm --dry-run --debug
```

Pronto! Seu Helm Chart está publicado e pronto para uso! 🚀
