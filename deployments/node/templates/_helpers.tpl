{{/*
Expand the name of the chart.
*/}}
{{- define "fluidos.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "fluidos.fullname" -}}
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
Create version used to select the Fluidos node version to be installed .
*/}}
{{- define "fluidos.version" -}}
{{- if .Values.tag }}
{{- .Values.tag }}
{{- else if .Chart.AppVersion }}
{{- .Chart.AppVersion }}
{{- else }}
{{- fail "At least one between .Values.tag and .Chart.AppVersion should be set" }}
{{- end }}
{{- end }}

{{/*
The suffix added to the Fluidos images, to identify CI builds.
*/}}
{{- define "fluidos.suffix" -}}
{{/* https://semver.org/#is-there-a-suggested-regular-expression-regex-to-check-a-semver-string */}}
{{- $semverregex := "^v(?P<major>0|[1-9]\\d*)\\.(?P<minor>0|[1-9]\\d*)\\.(?P<patch>0|[1-9]\\d*)(?:-(?P<prerelease>(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\\.(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\\.[0-9a-zA-Z-]+)*))?$" }}
{{- if or (eq .Values.tag "") (mustRegexMatch $semverregex .Values.tag) }}
{{- print "" }}
{{- else }}
{{- print "-ci" }}
{{- end }}
{{- end }}

{{/*
Create a name prefixed with the chart name, it accepts a dict which contains the field "name".
*/}}
{{- define "fluidos.prefixedName" -}}
{{- printf "%s-%s" (include "fluidos.name" .) .name }}
{{- end }}

{{/*
Create the file name of a role starting from a prefix, it accepts a dict which contains the field "prefix".
*/}}
{{- define "fluidos.role-filename" -}}
{{- printf "files/%s-%s" .prefix "Role.yaml" }}
{{- end }}

{{/*
Create the file name of a cluster role starting from a prefix, it accepts a dict which contains the field "prefix".
*/}}
{{- define "fluidos.cluster-role-filename" -}}
{{- printf "files/%s-%s" .prefix "ClusterRole.yaml" }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "fluidos.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "fluidos.labels" -}}
{{ include "fluidos.selectorLabels" . }}
helm.sh/chart: {{ include "fluidos.chart" . }}
app.kubernetes.io/version: {{ quote (include "fluidos.version" .) }}
app.kubernetes.io/managed-by: {{ quote .Release.Service }}
{{- end }}

{{/*
Selector labels, it accepts a dict which contains fields "name" and "module"
*/}}
{{- define "fluidos.selectorLabels" -}}
app.kubernetes.io/name: {{ quote .name }}
app.kubernetes.io/instance: {{ quote (printf "%s-%s" .Release.Name .name) }}
app.kubernetes.io/component: {{ quote .module }}
app.kubernetes.io/part-of: {{ quote (include "fluidos.name" .) }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "fluidos.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "fluidos.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Get the Pod security context
*/}}
{{- define "fluidos.podSecurityContext" -}}
runAsNonRoot: true
runAsUser: 1000
runAsGroup: 1000
fsGroup: 1000
{{- end -}}

{{/*
Get the Container security context
*/}}
{{- define "fluidos.containerSecurityContext" -}}
allowPrivilegeEscalation: false
{{- end -}}