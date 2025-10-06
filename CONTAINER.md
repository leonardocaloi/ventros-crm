# 🐳 Container Runtime Support

Ventros CRM é **agnóstico ao runtime de containers**. Funciona com qualquer ferramenta compatível com OCI (Open Container Initiative).

---

## 📦 Runtimes Suportados

### ✅ Docker
```bash
# Build
docker build -f Containerfile -t ventros-crm:latest .

# Run full stack
docker compose up -d

# Run apenas infra (sem API)
docker compose up -d postgres rabbitmq redis temporal
```

### ✅ Podman
```bash
# Build
podman build -f Containerfile -t ventros-crm:latest .

# Run full stack
podman-compose up -d
# ou
podman play kube deployments/kubernetes/

# Run apenas infra
podman-compose up -d postgres rabbitmq redis temporal
```

### ✅ Buildah
```bash
# Build (mais controle)
buildah bud -f Containerfile -t ventros-crm:latest .

# Push para registry
buildah push ventros-crm:latest docker://registry.example.com/ventros-crm:latest
```

### ✅ nerdctl (containerd)
```bash
# Build
nerdctl build -f Containerfile -t ventros-crm:latest .

# Run
nerdctl compose up -d
```

---

## 📋 Arquivos Container

### Raiz do Projeto (OCI Standard)
```
Containerfile       # Build file (padrão OCI)
entrypoint.sh      # Entrypoint script
compose.yaml       # Compose file (agnóstico)
```

### Deployments
```
deployments/
├── docker/        # Docker-specific configs
│   ├── init.sql
│   └── seeds/
├── kubernetes/    # K8s manifests
└── helm/          # Helm charts
```

---

## 🎯 Por que Containerfile?

**Containerfile** é o padrão **OCI (Open Container Initiative)**:
- ✅ Funciona com Docker, Podman, Buildah, etc
- ✅ Não depende de vendor específico
- ✅ Compatível com Kubernetes/OpenShift
- ✅ Futuro-proof

**Dockerfile** é específico do Docker, mas todos os runtimes o aceitam por compatibilidade.

---

## 🚀 Quick Start

### Opção 1: Docker
```bash
# Full stack
make docker-up
# ou
docker compose up -d
```

### Opção 2: Podman (rootless)
```bash
# Full stack
podman-compose up -d

# Apenas infra
podman-compose up -d postgres rabbitmq redis
```

### Opção 3: Build local + Run infra
```bash
# Sobe apenas infraestrutura
docker compose up -d postgres rabbitmq redis temporal

# Roda app localmente (Go)
make run
```

---

## 🔧 Makefile Targets

```bash
make container-build    # Build com runtime padrão ($CONTAINER_RUNTIME)
make container-run      # Run com compose
make container-stop     # Para containers
make container-clean    # Remove containers e volumes
```

**Configurar runtime**:
```bash
# Docker (default)
export CONTAINER_RUNTIME=docker

# Podman
export CONTAINER_RUNTIME=podman

# Buildah (apenas build)
export CONTAINER_RUNTIME=buildah
```

---

## 📖 Documentação Adicional

- **Docker**: https://docs.docker.com/
- **Podman**: https://docs.podman.io/
- **Buildah**: https://buildah.io/
- **OCI Spec**: https://opencontainers.org/

---

## 🆘 Troubleshooting

### Podman não encontra Containerfile
```bash
# Usar -f explicitamente
podman build -f Containerfile -t ventros-crm .
```

### Permissões no Podman (rootless)
```bash
# Ajustar subuid/subgid
sudo usermod --add-subuids 100000-165535 $USER
sudo usermod --add-subgids 100000-165535 $USER
```

### Compose file não reconhecido
```bash
# Docker Compose v2
docker compose -f compose.yaml up

# Docker Compose v1 (legado)
docker-compose -f compose.yaml up
```

---

**Escolha o runtime que preferir - Ventros CRM funciona com todos!** 🚀
