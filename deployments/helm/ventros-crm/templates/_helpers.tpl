{{/*
Expand the name of the chart.
*/}}
{{- define "ventros-crm.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "ventros-crm.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "ventros-crm.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "ventros-crm.labels" -}}
helm.sh/chart: {{ include "ventros-crm.chart" . }}
{{ include "ventros-crm.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "ventros-crm.selectorLabels" -}}
app.kubernetes.io/name: {{ include "ventros-crm.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Selector labels for API pods only
*/}}
{{- define "ventros-crm.apiSelectorLabels" -}}
app.kubernetes.io/name: {{ include "ventros-crm.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: api
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "ventros-crm.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "ventros-crm.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the name of the config map
*/}}
{{- define "ventros-crm.configMapName" -}}
{{- printf "%s-config" (include "ventros-crm.fullname" .) }}
{{- end }}

{{/*
Return the proper image name
*/}}
{{- define "ventros-crm.image" -}}
{{- $tag := .Values.image.tag | default .Chart.AppVersion }}
{{- printf "%s:%s" .Values.image.repository $tag }}
{{- end }}

{{/*
Return the proper migration image name
*/}}
{{- define "ventros-crm.migrationImage" -}}
{{- $tag := .Values.migrations.image.tag | default .Chart.AppVersion }}
{{- printf "%s:%s" .Values.migrations.image.repository $tag }}
{{- end }}

{{/*
Return RabbitMQ host
*/}}
{{- define "ventros-crm.rabbitmq.host" -}}
{{- if .Values.rabbitmqOperator.enabled }}
{{- printf "%s-rabbitmq" (include "ventros-crm.fullname" .) }}
{{- else if .Values.rabbitmq.enabled }}
{{- printf "%s-rabbitmq" .Release.Name }}
{{- else }}
{{- .Values.externalRabbitmq.host }}
{{- end }}
{{- end }}

{{/*
Return RabbitMQ username
*/}}
{{- define "ventros-crm.rabbitmq.username" -}}
{{- if or .Values.rabbitmqOperator.enabled .Values.rabbitmq.enabled }}
{{- .Values.rabbitmq.auth.username | default "user" }}
{{- else }}
{{- .Values.externalRabbitmq.username }}
{{- end }}
{{- end }}

{{/*
Return RabbitMQ password
*/}}
{{- define "ventros-crm.rabbitmq.password" -}}
{{- if or .Values.rabbitmqOperator.enabled .Values.rabbitmq.enabled }}
{{- .Values.rabbitmq.auth.password | default "password" }}
{{- else }}
{{- .Values.externalRabbitmq.password }}
{{- end }}
{{- end }}
