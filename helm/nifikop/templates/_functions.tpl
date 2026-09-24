{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "nifikop.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "nifikop.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Return the appropriate apiVersion value to use for the capi-operator managed k8s resources
*/}}
{{- define "nifikop.apiVersion" -}}
{{- printf "%s" "nificlusters.nifi.konpyutaika.com/v1" -}}
{{- end -}}

{{- define "userdefined.labels" }}
{{ if .Values.labels }}
{{- with .Values.labels }}
{{- toYaml . | nindent 4 }}
{{- end}}
{{- end}}
{{- end }}

{{- define "userdefined.annotations" }}
{{ if .Values.annotations }}
{{- with .Values.annotations }}
{{- toYaml . | nindent 4 }}
{{- end}}
{{- end}}
{{- end }}


{{- define "webhook.service.name" -}}
{{- $name := default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- printf "%s-webhook" $name -}}
{{- end -}}

{{- define "webhook.secret.name" -}}
{{- $name := default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- printf "%s-webhook-server-cert" $name -}}
{{- end -}}

{{- define "webhook.tls.mode" -}}
{{- $webhook := .Values.webhook | default dict -}}
{{- $tls := $webhook.tls | default dict -}}
{{- default "certManager" $tls.mode -}}
{{- end -}}

{{- define "webhook.tls.shared.secret.name" -}}
{{- $webhook := .Values.webhook | default dict -}}
{{- $tls := $webhook.tls | default dict -}}
{{- $tls.secretName -}}
{{- end -}}

{{- define "webhook.tls.secret.name" -}}
{{- $webhook := .Values.webhook | default dict -}}
{{- $tls := $webhook.tls | default dict -}}
{{- $mode := include "webhook.tls.mode" . -}}
{{- $sharedSecretName := include "webhook.tls.shared.secret.name" . -}}
{{- if eq $mode "existingSecret" -}}
{{- $existing := $tls.existingSecret | default dict -}}
{{- default $sharedSecretName $existing.name -}}
{{- else -}}
{{- $cm := $tls.certManager | default dict -}}
{{- default (default (include "webhook.secret.name" .) $sharedSecretName) $cm.secretName -}}
{{- end -}}
{{- end -}}

{{- define "webhook.certificate.name" -}}
{{- $name := default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- printf "%s-webhook-cert" $name -}}
{{- end -}}

{{- define "nifikop.openshift.scc.create" -}}
{{- if and .Values.openshift.scc.create (.Capabilities.APIVersions.Has "security.openshift.io/v1") -}}true{{- end -}}
{{- end -}}

{{- define "nifikop.openshift.scc.name" -}}
{{- if .Values.openshift.scc.existingName -}}
{{- .Values.openshift.scc.existingName -}}
{{- else -}}
{{- printf "%s-openshift-scc" (include "nifikop.name" .) -}}
{{- end -}}
{{- end -}}

{{- define "nifikop.webhook.tls.validate" -}}
{{- $webhook := .Values.webhook | default dict -}}
{{- if $webhook.enabled -}}
{{- $tls := $webhook.tls | default dict -}}
{{- $mode := include "webhook.tls.mode" . -}}
{{- if not (has $mode (list "certManager" "existingSecret")) -}}
{{- fail (printf "webhook.tls.mode must be one of [certManager existingSecret], got %q" $mode) -}}
{{- end -}}
{{- if eq $mode "existingSecret" -}}
{{- $existing := $tls.existingSecret | default dict -}}
{{- $secretName := default (include "webhook.tls.shared.secret.name" .) $existing.name -}}
{{- if not $secretName -}}
{{- fail "webhook.tls.existingSecret.name or webhook.tls.secretName is required when webhook.tls.mode=existingSecret" -}}
{{- end -}}
{{- end -}}
{{- if eq $mode "certManager" -}}
{{- if not .Values.certManager.enabled -}}
{{- fail "webhook.tls.mode=certManager requires certManager.enabled=true; to bring your own TLS secret set webhook.tls.mode=existingSecret and webhook.tls.existingSecret.name or webhook.tls.secretName" -}}
{{- end -}}
{{- $tlsCertManager := $tls.certManager | default dict -}}
{{- $issuerRef := $tlsCertManager.issuerRef | default dict -}}
{{- $issuerKind := default "Issuer" $issuerRef.kind -}}
{{- $issuerGroup := default "cert-manager.io" $issuerRef.group -}}
{{- if $tlsCertManager.createIssuer -}}
{{- if ne $issuerKind "Issuer" -}}
{{- fail "webhook.tls.certManager.createIssuer=true requires webhook.tls.certManager.issuerRef.kind=Issuer" -}}
{{- end -}}
{{- if ne $issuerGroup "cert-manager.io" -}}
{{- fail "webhook.tls.certManager.createIssuer=true requires webhook.tls.certManager.issuerRef.group=cert-manager.io" -}}
{{- end -}}
{{- end -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Name of the ServiceAccount the operator runs as.
*/}}
{{- define "nifikop.serviceAccountName" -}}
{{- if and .Values.serviceAccount .Values.serviceAccount.name -}}
{{- .Values.serviceAccount.name -}}
{{- else -}}
{{- template "nifikop.name" . -}}
{{- end -}}
{{- end -}}

{{/*
Namespaces the operator watches when not cluster-wide, as a YAML list.
Defaults to the release namespace. Consume with `| fromYamlArray`.
*/}}
{{- define "nifikop.watchNamespaces" -}}
{{- if .Values.namespaces -}}
{{- toYaml .Values.namespaces -}}
{{- else -}}
{{- list .Release.Namespace | toYaml -}}
{{- end -}}
{{- end -}}

{{/*
Reject value combinations that cannot produce working RBAC.
*/}}
{{- define "nifikop.rbac.validate" -}}
{{- if and .Values.watchAnyNamespace (not .Values.createClusterScopedResources) -}}
{{- fail "watchAnyNamespace=true requires createClusterScopedResources=true (cluster-wide watch needs a ClusterRoleBinding)" -}}
{{- end -}}
{{- if and .Values.watchAnyNamespace .Values.namespaces -}}
{{- fail "namespaces must be empty when watchAnyNamespace=true" -}}
{{- end -}}
{{- if and .Values.certManager.clusterScoped (not .Values.createClusterScopedResources) -}}
{{- fail "certManager.clusterScoped=true requires createClusterScopedResources=true (ClusterIssuer access cannot be granted by a namespaced Role)" -}}
{{- end -}}
{{- end -}}

{{/*
Permissions on namespaced resources, shared by the ClusterRole and per-namespace Role variants.
*/}}
{{- define "nifikop.rbac.namespacedRules" -}}
- apiGroups:
  - ""
  resources:
  - pods
  - services
  - services/finalizers
  - endpoints
  - persistentvolumeclaims
  - events
  - configmaps
  - secrets
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - apps
  resources:
  - deployments
  - daemonsets
  - replicasets
  - statefulsets
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - monitoring.coreos.com
  resources:
  - servicemonitors
  verbs:
  - get
  - create
- apiGroups:
  - apps
  resourceNames:
  - nifikop
  resources:
  - deployments/finalizers
  verbs:
  - update
- apiGroups:
  - ""
  resources:
  - pods
  verbs:
  - get
- apiGroups:
  - apps
  resources:
  - replicasets
  - deployments
  verbs:
  - get
- apiGroups:
  - nifi.konpyutaika.com
  resources:
  - "nifiusers"
  - "nifiusergroups"
  - "nificlusters"
  - "nifidataflows"
  - "nifiregistryclients"
  - "nifiparametercontexts"
  - "nifinodegroupautoscalers"
  - "nificonnections"
  - "nifiresources"
  - "nifiusers/finalizers"
  - "nificlusters/finalizers"
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
  - deletecollection
- apiGroups:
  - autoscaling
  resources:
  - horizontalpodautoscalers
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - cert-manager.io
  resources:
  - issuers
  - certificates
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - nifi.konpyutaika.com
  resources:
  - nifiusers/status
  - nifiusergroups/status
  - nificlusters/status
  - nifidataflows/status
  - nifiregistryclients/status
  - nifiparametercontexts/status
  - nifinodegroupautoscalers/status
  - nificonnections/status
  - nifiresources/status
  verbs:
  - get
  - update
  - patch
- apiGroups:
  - policy
  resources:
  - poddisruptionbudgets
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - coordination.k8s.io
  resources:
  - leases
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - patch
  - delete
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - create
  - patch
{{- end -}}
