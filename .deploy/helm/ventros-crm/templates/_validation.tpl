{{/*
Validation helpers to catch configuration errors early
*/}}

{{/*
Validate PostgreSQL configuration
*/}}
{{- define "ventros-crm.validate.postgresql" -}}
{{- $hasExternal := and .Values.externalPostgresql.host (ne .Values.externalPostgresql.host "external-postgres.example.com") -}}
{{- if not (or .Values.postgresOperator.enabled $hasExternal) -}}
{{- fail "ERROR: PostgreSQL must be configured. Either enable postgresOperator.enabled=true or provide externalPostgresql.host" -}}
{{- end -}}
{{- if and .Values.postgresOperator.enabled $hasExternal -}}
{{- fail "ERROR: Cannot use both internal PostgreSQL (postgresOperator.enabled=true) and external PostgreSQL. Choose one." -}}
{{- end -}}
{{- end -}}

{{/*
Validate Redis configuration
*/}}
{{- define "ventros-crm.validate.redis" -}}
{{- $hasExternal := and .Values.externalRedis.host (ne .Values.externalRedis.host "external-redis.example.com") -}}
{{- if not (or .Values.redis.enabled $hasExternal) -}}
{{- fail "ERROR: Redis must be configured. Either enable redis.enabled=true or provide externalRedis.host" -}}
{{- end -}}
{{- if and .Values.redis.enabled $hasExternal -}}
{{- fail "ERROR: Cannot use both internal Redis (redis.enabled=true) and external Redis. Choose one." -}}
{{- end -}}
{{- end -}}

{{/*
Validate RabbitMQ configuration
*/}}
{{- define "ventros-crm.validate.rabbitmq" -}}
{{- $hasExternal := and .Values.externalRabbitmq.host (ne .Values.externalRabbitmq.host "external-rabbitmq.example.com") -}}
{{- if not (or .Values.rabbitmqOperator.enabled .Values.rabbitmq.enabled $hasExternal) -}}
{{- fail "ERROR: RabbitMQ must be configured. Enable rabbitmqOperator.enabled=true, rabbitmq.enabled=true, or provide externalRabbitmq.host" -}}
{{- end -}}
{{- $count := 0 -}}
{{- if .Values.rabbitmqOperator.enabled -}}{{ $count = add $count 1 }}{{- end -}}
{{- if .Values.rabbitmq.enabled -}}{{ $count = add $count 1 }}{{- end -}}
{{- if $hasExternal -}}{{ $count = add $count 1 }}{{- end -}}
{{- if gt $count 1 -}}
{{- fail "ERROR: Only one RabbitMQ option can be enabled at a time (rabbitmqOperator, rabbitmq, or externalRabbitmq)" -}}
{{- end -}}
{{- end -}}

{{/*
Validate Temporal configuration
*/}}
{{- define "ventros-crm.validate.temporal" -}}
{{- $hasExternal := and .Values.temporal.external.host (ne .Values.temporal.external.host "external-temporal.example.com") -}}
{{- if and .Values.temporal.enabled $hasExternal -}}
{{- fail "ERROR: Cannot use both internal Temporal (temporal.enabled=true) and external Temporal. Choose one." -}}
{{- end -}}
{{- end -}}

{{/*
Validate required secrets
*/}}
{{- define "ventros-crm.validate.secrets" -}}
{{- if and (not .Values.secrets.existingSecret) (not .Values.secrets.wahaApiKey) -}}
{{- fail "ERROR: WAHA API Key must be provided via secrets.wahaApiKey or secrets.existingSecret" -}}
{{- end -}}
{{- if and (not .Values.secrets.existingSecret) (not .Values.secrets.adminPassword) -}}
{{- fail "ERROR: Admin password must be provided via secrets.adminPassword or secrets.existingSecret" -}}
{{- end -}}
{{- end -}}

{{/*
Run all validations
*/}}
{{- define "ventros-crm.validate.all" -}}
{{- include "ventros-crm.validate.postgresql" . -}}
{{- include "ventros-crm.validate.redis" . -}}
{{- include "ventros-crm.validate.rabbitmq" . -}}
{{- include "ventros-crm.validate.temporal" . -}}
{{- include "ventros-crm.validate.secrets" . -}}
{{- end -}}
