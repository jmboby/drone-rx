{{/*
Expand the name of the chart.
*/}}
{{- define "dronerx.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "dronerx.fullname" -}}
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
{{- define "dronerx.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "dronerx.labels" -}}
helm.sh/chart: {{ include "dronerx.chart" . }}
app.kubernetes.io/name: {{ include "dronerx.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
API labels
*/}}
{{- define "dronerx.api.labels" -}}
{{ include "dronerx.labels" . }}
app.kubernetes.io/component: api
{{- end }}

{{/*
API selector labels
*/}}
{{- define "dronerx.api.selectorLabels" -}}
app.kubernetes.io/name: {{ include "dronerx.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: api
{{- end }}

{{/*
Frontend labels
*/}}
{{- define "dronerx.frontend.labels" -}}
{{ include "dronerx.labels" . }}
app.kubernetes.io/component: frontend
{{- end }}

{{/*
Frontend selector labels
*/}}
{{- define "dronerx.frontend.selectorLabels" -}}
app.kubernetes.io/name: {{ include "dronerx.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: frontend
{{- end }}

{{/*
Database configuration guard.
Fails if postgresql is disabled but no external host is provided.
Called automatically by the dronerx.databaseURL helper.
*/}}
{{- define "dronerx.validateDatabase" -}}
{{- if and (not .Values.postgresql.enabled) (not .Values.externalDatabase.host) -}}
  {{- fail "Invalid configuration: postgresql.enabled=false but externalDatabase.host is not set. Either enable embedded postgres or provide an external database host." -}}
{{- end -}}
{{- end -}}

{{/*
Database URL helper.
Uses CNPG cluster when postgresql.enabled=true, otherwise external DB.
*/}}
{{- define "dronerx.databaseURL" -}}
{{- include "dronerx.validateDatabase" . -}}
{{- if .Values.postgresql.enabled -}}
{{/* sslmode=disable is correct for cluster-local CNPG traffic */}}
{{- printf "postgres://dronerx:$(DB_PASSWORD)@%s-db-rw:5432/dronerx?sslmode=disable" (include "dronerx.fullname" .) -}}
{{- else -}}
{{- printf "postgres://%s:$(DB_PASSWORD)@%s:%d/%s?sslmode=%s" .Values.externalDatabase.user .Values.externalDatabase.host (int .Values.externalDatabase.port) .Values.externalDatabase.name (.Values.externalDatabase.sslmode | default "require") -}}
{{- end -}}
{{- end -}}

{{/*
NATS URL helper.
*/}}
{{- define "dronerx.natsURL" -}}
{{- if .Values.nats.enabled }}
{{- printf "nats://%s-nats:4222" .Release.Name }}
{{- else }}
{{- "nats://nats:4222" }}
{{- end }}
{{- end }}

{{/*
Image pull secrets helper.
Includes the Replicated enterprise-pull-secret when available, plus any user-provided secrets.
*/}}
{{- define "dronerx.imagePullSecrets" -}}
{{- if .Values.global }}
{{- if .Values.global.replicated }}
{{- if .Values.global.replicated.dockerconfigjson }}
- name: enterprise-pull-secret
{{- end }}
{{- end }}
{{- end }}
{{- range .Values.imagePullSecrets }}
- name: {{ .name }}
{{- end }}
{{- end }}

{{/*
TLS secret name helper.
*/}}
{{- define "dronerx.tlsSecretName" -}}
{{- if .Values.ingress.tls.secretName }}
{{- .Values.ingress.tls.secretName }}
{{- else }}
{{- printf "%s-tls" (include "dronerx.fullname" .) }}
{{- end }}
{{- end }}
