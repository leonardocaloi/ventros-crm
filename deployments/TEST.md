# 🧪 Testes Docker vs Podman

## 🎯 Cenários de Teste

### 1️⃣ Teste: Só Infraestrutura (DEV)

#### Docker
```bash
# Build (não necessário para compose.dev.yaml)
# Run infra
docker compose -f deployments/compose.dev.yaml up -d

# Verificar
docker compose -f deployments/compose.dev.yaml ps
docker ps

# Logs
docker compose -f deployments/compose.dev.yaml logs -f

# Parar
docker compose -f deployments/compose.dev.yaml down

# Limpar volumes
docker compose -f deployments/compose.dev.yaml down -v
```

#### Podman
```bash
# Run infra
podman-compose -f deployments/compose.dev.yaml up -d

# Verificar
podman-compose -f deployments/compose.dev.yaml ps
podman ps

# Logs
podman-compose -f deployments/compose.dev.yaml logs -f

# Parar
podman-compose -f deployments/compose.dev.yaml down

# Limpar volumes
podman-compose -f deployments/compose.dev.yaml down -v
```

---

### 2️⃣ Teste: Full Stack (Infra + API)

#### Docker
```bash
# Build da imagem
docker build -f deployments/Containerfile -t ventros-crm:latest .

# Run full stack
docker compose -f deployments/compose.yaml up -d

# Verificar API
curl http://localhost:8080/health

# Verificar todos os serviços
docker compose -f deployments/compose.yaml ps

# Parar
docker compose -f deployments/compose.yaml down
```

#### Podman
```bash
# Build da imagem
podman build -f deployments/Containerfile -t ventros-crm:latest .

# Run full stack
podman-compose -f deployments/compose.yaml up -d

# Verificar API
curl http://localhost:8080/health

# Verificar todos os serviços
podman-compose -f deployments/compose.yaml ps

# Parar
podman-compose -f deployments/compose.yaml down
```

---

## 📊 Checklist de Testes

### ✅ Teste 1: compose.dev.yaml (só infra)

- [ ] **Docker**
  - [ ] Build passa sem erros
  - [ ] Todos os containers sobem
  - [ ] Healthchecks passam
  - [ ] PostgreSQL conecta (porta 5432)
  - [ ] RabbitMQ UI acessível (http://localhost:15672)
  - [ ] Redis responde (redis-cli ping)
  - [ ] Temporal UI acessível (http://localhost:8088)

- [ ] **Podman**
  - [ ] Build passa sem erros
  - [ ] Todos os containers sobem
  - [ ] Healthchecks passam
  - [ ] PostgreSQL conecta (porta 5432)
  - [ ] RabbitMQ UI acessível (http://localhost:15672)
  - [ ] Redis responde (redis-cli ping)
  - [ ] Temporal UI acessível (http://localhost:8088)

### ✅ Teste 2: compose.yaml (full stack)

- [ ] **Docker**
  - [ ] Imagem builda com sucesso
  - [ ] Infra sobe corretamente
  - [ ] API sobe e conecta na infra
  - [ ] /health retorna 200 OK
  - [ ] Swagger acessível (http://localhost:8080/swagger/index.html)
  - [ ] Migrations executam
  - [ ] Logs sem erros críticos

- [ ] **Podman**
  - [ ] Imagem builda com sucesso
  - [ ] Infra sobe corretamente
  - [ ] API sobe e conecta na infra
  - [ ] /health retorna 200 OK
  - [ ] Swagger acessível (http://localhost:8080/swagger/index.html)
  - [ ] Migrations executam
  - [ ] Logs sem erros críticos

---

## 🐛 Troubleshooting

### Podman: Permissão negada em volumes
```bash
# Ajustar SELinux (Fedora/RHEL)
sudo semanage fcontext -a -t container_file_t "/path/to/volumes(/.*)?"
sudo restorecon -Rv /path/to/volumes
```

### Podman: Porta já em uso
```bash
# Verificar portas
ss -tuln | grep -E '5432|5672|6379|7233|8080|8088'

# Parar processos
podman stop --all
```

### Docker/Podman: Network não encontrada
```bash
# Recriar network
docker network rm ventros-network
docker network create ventros-network

# Ou com Podman
podman network rm ventros-network
podman network create ventros-network
```

---

## 📈 Resultado Esperado

### ✅ Sucesso

Ambos Docker e Podman devem:
1. ✅ Buildar a imagem sem erros
2. ✅ Subir todos os containers
3. ✅ Passar todos os healthchecks
4. ✅ API responder em /health
5. ✅ Temporal funcionar corretamente

### ❌ Diferenças Conhecidas

- **Podman rootless**: Pode ter problemas com portas < 1024
- **Podman no Mac**: Requer Podman Desktop ou VM
- **Docker no Linux**: Requer usuário no grupo `docker`

---

## 🎯 Recomendação

Para **desenvolvimento local**:
- Use `compose.dev.yaml` + `make run` (Go nativo)
- Mais rápido que rebuild de containers

Para **teste de produção**:
- Use `compose.yaml` completo
- Testa o comportamento containerizado

---

**Data**: 2025-10-06  
**Testado**: Docker 24.x, Podman 4.x
