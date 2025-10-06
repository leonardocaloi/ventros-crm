{{/*
Get PostgreSQL host
*/}}
{{- define "ventros-crm.postgresql.host" -}}
{{- if .Values.postgresOperator.enabled -}}
{{- printf "%s-postgres" (include "ventros-crm.fullname" .) -}}
{{- else -}}
{{- .Values.externalPostgresql.host -}}
{{- end -}}
{{- end -}}

{{/*
Get PostgreSQL port
*/}}
{{- define "ventros-crm.postgresql.port" -}}
{{- if .Values.postgresOperator.enabled -}}
5432
{{- else -}}
{{- .Values.externalPostgresql.port -}}
{{- end -}}
{{- end -}}

{{/*
Get PostgreSQL database
*/}}
{{- define "ventros-crm.postgresql.database" -}}
{{- if .Values.postgresOperator.enabled -}}
{{- .Values.postgresOperator.cluster.database -}}
{{- else -}}
{{- .Values.externalPostgresql.database -}}
{{- end -}}
{{- end -}}

{{/*
Get PostgreSQL username
*/}}
{{- define "ventros-crm.postgresql.username" -}}
{{- if .Values.postgresOperator.enabled -}}
{{- .Values.postgresOperator.cluster.username -}}
{{- else -}}
{{- .Values.externalPostgresql.username -}}
{{- end -}}
{{- end -}}

{{/*
Get PostgreSQL password secret name
*/}}
{{- define "ventros-crm.postgresql.secretName" -}}
{{- if .Values.postgresOperator.enabled -}}
{{- printf "%s.%s-postgres.credentials.postgresql.acid.zalan.do" .Values.postgresOperator.cluster.username (include "ventros-crm.fullname" .) -}}
{{- else -}}
{{- .Values.externalPostgresql.existingSecret -}}
{{- end -}}
{{- end -}}

{{/*
Get PostgreSQL password secret key
*/}}
{{- define "ventros-crm.postgresql.secretKey" -}}
{{- if .Values.postgresOperator.enabled -}}
password
{{- else -}}
{{- .Values.externalPostgresql.existingSecretPasswordKey -}}
{{- end -}}
{{- end -}}

{{/*
Get PostgreSQL SSL mode
*/}}
{{- define "ventros-crm.postgresql.sslMode" -}}
{{- if .Values.postgresOperator.enabled -}}
disable
{{- else -}}
{{- .Values.externalPostgresql.sslMode -}}
{{- end -}}
{{- end -}}

{{/*
Get Redis host
*/}}
{{- define "ventros-crm.redis.host" -}}
{{- if .Values.redis.enabled -}}
{{- printf "%s-redis-master" (include "ventros-crm.fullname" .) -}}
{{- else -}}
{{- .Values.externalRedis.host -}}
{{- end -}}
{{- end -}}

{{/*
Get Redis port
*/}}
{{- define "ventros-crm.redis.port" -}}
{{- if .Values.redis.enabled -}}
6379
{{- else -}}
{{- .Values.externalRedis.port -}}
{{- end -}}
{{- end -}}

{{/*
Get Redis password secret name (if auth enabled)
*/}}
{{- define "ventros-crm.redis.secretName" -}}
{{- if .Values.redis.enabled -}}
{{- if .Values.redis.auth.enabled -}}
{{- printf "%s-redis" (include "ventros-crm.fullname" .) -}}
{{- else -}}
{{- end -}}
{{- else -}}
{{- .Values.externalRedis.existingSecret -}}
{{- end -}}
{{- end -}}

{{/*
Get Redis password secret key
*/}}
{{- define "ventros-crm.redis.secretKey" -}}
{{- if .Values.redis.enabled -}}
redis-password
{{- else -}}
{{- .Values.externalRedis.existingSecretPasswordKey -}}
{{- end -}}
{{- end -}}

{{/*
Get Redis database
*/}}
{{- define "ventros-crm.redis.database" -}}
{{- if .Values.redis.enabled -}}
0
{{- else -}}
{{- .Values.externalRedis.database -}}
{{- end -}}
{{- end -}}

{{/*
Get RabbitMQ host
*/}}
{{- define "ventros-crm.rabbitmq.host" -}}
{{- if .Values.rabbitmqOperator.enabled -}}
{{- printf "%s-rabbitmq" (include "ventros-crm.fullname" .) -}}
{{- else if .Values.rabbitmq.enabled -}}
{{- printf "%s-rabbitmq" (include "ventros-crm.fullname" .) -}}
{{- else -}}
{{- .Values.externalRabbitmq.host -}}
{{- end -}}
{{- end -}}

{{/*
Get RabbitMQ port
*/}}
{{- define "ventros-crm.rabbitmq.port" -}}
{{- if .Values.rabbitmq.enabled -}}
5672
{{- else -}}
{{- .Values.externalRabbitmq.port -}}
{{- end -}}
{{- end -}}

{{/*
Get RabbitMQ username
*/}}
{{- define "ventros-crm.rabbitmq.username" -}}
{{- if .Values.rabbitmqOperator.enabled -}}
{{- .Values.rabbitmqOperator.cluster.auth.username -}}
{{- else if .Values.rabbitmq.enabled -}}
{{- .Values.rabbitmq.auth.username -}}
{{- else -}}
{{- .Values.externalRabbitmq.username -}}
{{- end -}}
{{- end -}}

{{/*
Get RabbitMQ password secret name
*/}}
{{- define "ventros-crm.rabbitmq.secretName" -}}
{{- if .Values.rabbitmqOperator.enabled -}}
{{- printf "%s-rabbitmq-credentials" (include "ventros-crm.fullname" .) -}}
{{- else if .Values.rabbitmq.enabled -}}
{{- printf "%s-rabbitmq" (include "ventros-crm.fullname" .) -}}
{{- else -}}
{{- .Values.externalRabbitmq.existingSecret -}}
{{- end -}}
{{- end -}}

{{/*
Get RabbitMQ password secret key
*/}}
{{- define "ventros-crm.rabbitmq.secretKey" -}}
{{- if .Values.rabbitmqOperator.enabled -}}
password
{{- else if .Values.rabbitmq.enabled -}}
rabbitmq-password
{{- else -}}
{{- .Values.externalRabbitmq.existingSecretPasswordKey -}}
{{- end -}}
{{- end -}}

{{/*
Get Temporal host
*/}}
{{- define "ventros-crm.temporal.host" -}}
{{- if .Values.temporal.enabled -}}
{{- printf "%s-temporal-frontend.%s.svc.cluster.local" (include "ventros-crm.fullname" .) .Release.Namespace -}}
{{- else -}}
{{- .Values.temporal.external.host -}}
{{- end -}}
{{- end -}}

{{/*
Get Temporal port
*/}}
{{- define "ventros-crm.temporal.port" -}}
{{- if .Values.temporal.enabled -}}
7233
{{- else -}}
{{- .Values.temporal.external.port -}}
{{- end -}}
{{- end -}}

{{/*
Get Temporal namespace
*/}}
{{- define "ventros-crm.temporal.namespace" -}}
{{- if .Values.temporal.enabled -}}
default
{{- else -}}
{{- .Values.temporal.external.namespace -}}
{{- end -}}
{{- end -}}

{{/*
Get Temporal host:port
*/}}
{{- define "ventros-crm.temporal.hostPort" -}}
{{- printf "%s:%s" (include "ventros-crm.temporal.host" .) (include "ventros-crm.temporal.port" . | toString) -}}
{{- end -}}

{{/*
Get RabbitMQ URL (AMQP format)
NOTA: A senha não é incluída aqui pois será injetada via variável de ambiente RABBITMQ_PASSWORD
A aplicação deve construir a URL completa usando: amqp://user:password@host:port/
*/}}
{{- define "ventros-crm.rabbitmq.url" -}}
{{- $host := include "ventros-crm.rabbitmq.host" . -}}
{{- $port := include "ventros-crm.rabbitmq.port" . -}}
{{- printf "amqp://%s:%s/" $host $port -}}
{{- end -}}

{{/*
Get Temporal PostgreSQL database name (default store)
NOTA: Seguindo padrão do docker-compose, o Temporal usa o MESMO database da aplicação
O Temporal cria schemas separados automaticamente dentro do database (temporal, temporal_visibility)
*/}}
{{- define "ventros-crm.temporal.database.default" -}}
{{- include "ventros-crm.postgresql.database" . -}}
{{- end -}}

{{/*
Get Temporal PostgreSQL database name (visibility store)
NOTA: Mesmo database da aplicação (Temporal cria schema separado automaticamente)
*/}}
{{- define "ventros-crm.temporal.database.visibility" -}}
{{- include "ventros-crm.postgresql.database" . -}}
{{- end -}}

{{/*
Build Temporal server config for PostgreSQL (usado quando useInternalPostgres=true)
Este helper injeta automaticamente os valores de conexão no subchart do Temporal
*/}}
{{- define "ventros-crm.temporal.postgresConfig" -}}
{{- if and .Values.temporal.enabled .Values.temporal.useInternalPostgres -}}
{{- if .Values.postgresOperator.enabled -}}
server:
  config:
    persistence:
      default:
        driver: "sql"
        sql:
          driver: "postgres12"
          host: {{ include "ventros-crm.postgresql.host" . }}
          port: {{ include "ventros-crm.postgresql.port" . }}
          database: {{ include "ventros-crm.temporal.database.default" . }}
          user: {{ include "ventros-crm.postgresql.username" . }}
          existingSecret: {{ include "ventros-crm.postgresql.secretName" . }}
          secretName: {{ include "ventros-crm.postgresql.secretName" . }}
          maxConns: 20
          maxIdleConns: 20
          maxConnLifetime: "1h"
      visibility:
        driver: "sql"
        sql:
          driver: "postgres12"
          host: {{ include "ventros-crm.postgresql.host" . }}
          port: {{ include "ventros-crm.postgresql.port" . }}
          database: {{ include "ventros-crm.temporal.database.visibility" . }}
          user: {{ include "ventros-crm.postgresql.username" . }}
          existingSecret: {{ include "ventros-crm.postgresql.secretName" . }}
          secretName: {{ include "ventros-crm.postgresql.secretName" . }}
          maxConns: 10
          maxIdleConns: 10
          maxConnLifetime: "1h"
{{- end -}}
{{- end -}}
{{- end -}}
