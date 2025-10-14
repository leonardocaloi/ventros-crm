# Ventros CRM Role

## Description

This role deploys the Ventros CRM application to a Kubernetes cluster using Helm. It leverages existing operators (PostgreSQL Operator and RabbitMQ Operator) that should already be installed in the cluster.

## Requirements

- Kubernetes cluster (RKE2)
- `kubernetes.core` Ansible collection
- Helm 3.x
- Existing PostgreSQL Operator (Zalando) in `postgres-operator` namespace
- Existing RabbitMQ Operator in `rabbitmq-system` namespace

## Role Variables

All variables are prefixed with `ventros_crm_` to avoid conflicts:

### Core Settings

```yaml
ventros_crm_namespace: "ventros-crm"   # Kubernetes namespace
ventros_crm_enabled: true               # Enable/disable deployment
```

### Helm Configuration

```yaml
ventros_crm_helm_repo_name: "ventros"
ventros_crm_helm_repo_url: "https://leonardocaloi.github.io/ventros-crm/charts/"
ventros_crm_helm_chart_ref: "ventros/ventros-crm"
ventros_crm_helm_chart_version: "0.1.0"
```

### Application Configuration

```yaml
ventros_crm_image_repository: "leonardocaloi/ventros-crm"
ventros_crm_image_tag: "0.1.0"
ventros_crm_image_pull_policy: "IfNotPresent"
ventros_crm_replicas: 1
```

### Ingress Configuration

```yaml
ventros_crm_ingress_enabled: true
ventros_crm_ingress_class: "nginx"
ventros_crm_ingress_host: "crm.ventros.cloud"
ventros_crm_ingress_tls: true
ventros_crm_ingress_cert_issuer: "letsencrypt-clusterissuer"
```

### PostgreSQL Configuration

```yaml
ventros_crm_postgres_operator_install: false  # Use existing operator
ventros_crm_postgres_enabled: true
ventros_crm_postgres_team_id: "ventros"
ventros_crm_postgres_instances: 1
ventros_crm_postgres_storage_size: "10Gi"
ventros_crm_postgres_storage_class: "longhorn"
```

### RabbitMQ Configuration

```yaml
ventros_crm_rabbitmq_operator_install: false  # Use existing operator
ventros_crm_rabbitmq_enabled: true
ventros_crm_rabbitmq_replicas: 1
ventros_crm_rabbitmq_storage_size: "5Gi"
```

### Redis Configuration

```yaml
ventros_crm_redis_enabled: true
ventros_crm_redis_storage_size: "1Gi"
```

### Temporal Configuration

```yaml
ventros_crm_temporal_enabled: true
ventros_crm_temporal_frontend_replicas: 1
ventros_crm_temporal_history_replicas: 1
ventros_crm_temporal_matching_replicas: 1
ventros_crm_temporal_worker_replicas: 1
```

## Dependencies

This role depends on existing operators in the cluster:
- `postgres_operator` role (if `ventros_crm_postgres_operator_install: true`)
- `rabbitmq_operator` role (if `ventros_crm_rabbitmq_operator_install: true`)

By default, it assumes operators are already installed and will verify their presence.

## Example Playbook

```yaml
- hosts: k8s_control
  roles:
    - role: ventros_crm
      vars:
        ventros_crm_namespace: "production"
        ventros_crm_ingress_host: "crm.production.ventros.cloud"
        ventros_crm_replicas: 3
        ventros_crm_autoscaling_enabled: true
        ventros_crm_postgres_instances: 2
        ventros_crm_rabbitmq_replicas: 3
```

## Tags

- `infrastructure` - Namespace and operator verification
- `helm` - Helm repository and chart deployment
- `deploy` - Application deployment
- `verify` - Post-deployment verification

## Security Notes

- Sensitive values should be stored in Ansible Vault
- Default JWT and API key secrets must be changed in production
- Use appropriate resource limits for your environment
- Ensure TLS is enabled for ingress in production

## Endpoints

After deployment, the following endpoints will be available:

- **Application**: `https://<ingress_host>/`
- **Health Check**: `https://<ingress_host>/health`
- **Readiness Check**: `https://<ingress_host>/ready`
- **Temporal UI**: `https://<ingress_host>:8088` (if port-forwarded)

## Troubleshooting

### Verify operators are installed

```bash
kubectl get deployment -n postgres-operator postgres-operator
kubectl get deployment -n rabbitmq-system rabbitmq-operator
```

### Check pod status

```bash
kubectl get pods -n ventros-crm
kubectl logs -n ventros-crm -l app.kubernetes.io/name=ventros-crm
```

### Verify dependencies

```bash
kubectl exec -n ventros-crm <pod-name> -- wget -qO- http://localhost:8080/ready
```

## License

MIT

## Author Information

Created and maintained by Leonardo Caloi for Ventros.
