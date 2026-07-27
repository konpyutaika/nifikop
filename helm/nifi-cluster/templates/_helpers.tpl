{{/*
Expand the name of the chart.
*/}}
{{- define "nifi-cluster.name" -}}
{{- default .Chart.Name .Values.cluster.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "nifi-cluster.fullname" -}}
{{- if .Values.cluster.fullnameOverride }}
{{- .Values.cluster.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.cluster.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Resolve the service account name used by NiFi pods in kubernetes manager mode.
*/}}
{{- define "nifi-cluster.managerServiceAccountName" -}}
{{- tpl (default (include "nifi-cluster.fullname" .) .Values.cluster.managerServiceAccount.name) . -}}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "nifi-cluster.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "nifi-cluster.labels" -}}
helm.sh/chart: {{ include "nifi-cluster.chart" . }}
{{ include "nifi-cluster.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "nifi-cluster.selectorLabels" -}}
app.kubernetes.io/name: {{ include "nifi-cluster.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
OpenShift SCC helpers.
*/}}
{{- define "nifi-cluster.openshift.scc.create" -}}
{{- if and .Values.cluster.openshift.scc.create (.Capabilities.APIVersions.Has "security.openshift.io/v1") -}}true{{- end -}}
{{- end -}}

{{- define "nifi-cluster.openshift.scc.name" -}}
{{- if .Values.cluster.openshift.scc.existingName -}}
{{- .Values.cluster.openshift.scc.existingName -}}
{{- else -}}
{{- printf "%s-openshift-scc" (include "nifi-cluster.fullname" .) -}}
{{- end -}}
{{- end -}}

{{/*
True when any SCC mode is active (create or existing).
*/}}
{{- define "nifi-cluster.openshift.scc.enabled" -}}
{{- if or (eq (include "nifi-cluster.openshift.scc.create" .) "true") (not (empty .Values.cluster.openshift.scc.existingName)) -}}true{{- end -}}
{{- end -}}

{{/*
Resolve the workload service account name for NiFi pods.
When manager=kubernetes, use the manager SA. Otherwise use the SCC SA name.
*/}}
{{- define "nifi-cluster.workloadServiceAccountName" -}}
{{- if eq .Values.cluster.manager "kubernetes" -}}
{{- include "nifi-cluster.managerServiceAccountName" . -}}
{{- else -}}
{{- tpl (default (include "nifi-cluster.fullname" .) .Values.cluster.openshift.scc.serviceAccount.name) . -}}
{{- end -}}
{{- end -}}

{{/*
Resolve the default node service account name to inject into nodeConfigGroups.
Returns the workload SA when SCC mode is active or manager=kubernetes.
*/}}
{{- define "nifi-cluster.defaultNodeServiceAccountName" -}}
{{- if eq (include "nifi-cluster.openshift.scc.enabled" .) "true" -}}
{{- include "nifi-cluster.workloadServiceAccountName" . -}}
{{- else if eq .Values.cluster.manager "kubernetes" -}}
{{- include "nifi-cluster.managerServiceAccountName" . -}}
{{- end -}}
{{- end -}}
