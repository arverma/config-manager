{{- define "config-manager.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "config-manager.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := include "config-manager.name" . -}}
{{- printf "%s" $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{- define "config-manager.labels" -}}
app.kubernetes.io/name: {{ include "config-manager.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version | quote }}
{{- end -}}

{{- define "config-manager.selectorLabels" -}}
app.kubernetes.io/name: {{ include "config-manager.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{/*
Database secret name used by the API deployment.

Priority:
1) database.existingSecretName (user-provided Secret)
2) database.url (Secret created by this chart)
*/}}
{{- define "config-manager.dbSecretName" -}}
{{- if .Values.database.existingSecretName -}}
{{- .Values.database.existingSecretName -}}
{{- else if .Values.database.url -}}
{{- printf "%s-db" (include "config-manager.fullname" .) -}}
{{- else -}}
{{- fail "database.existingSecretName or database.url is required when using DATABASE_URL" -}}
{{- end -}}
{{- end -}}

{{/*
API secrets Secret name (created by ESO when enabled).
*/}}
{{- define "config-manager.apiSecretName" -}}
{{- if and .Values.api.externalSecret .Values.api.externalSecret.name -}}
{{- .Values.api.externalSecret.name -}}
{{- else -}}
{{- printf "%s-api-secrets" (include "config-manager.fullname" .) -}}
{{- end -}}
{{- end -}}

