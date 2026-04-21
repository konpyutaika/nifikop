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

{{- define "webhook.tls.secret.name" -}}
{{- $webhook := .Values.webhook | default dict -}}
{{- $tls := $webhook.tls | default dict -}}
{{- default (include "webhook.secret.name" .) $tls.secretName -}}
{{- end -}}

{{- define "webhook.certificate.name" -}}
{{- $name := default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- printf "%s-webhook-cert" $name -}}
{{- end -}}

{{- define "nifikop.webhook.tls.validate" -}}
{{- $webhook := .Values.webhook | default dict -}}
{{- if $webhook.enabled -}}
{{- $tls := $webhook.tls | default dict -}}
{{- $mode := default "certManager" $tls.mode -}}
{{- if not (has $mode (list "certManager" "existingSecret")) -}}
{{- fail (printf "webhook.tls.mode must be one of [certManager existingSecret], got %q" $mode) -}}
{{- end -}}
{{- if eq $mode "existingSecret" -}}
{{- if not $tls.secretName -}}
{{- fail "webhook.tls.secretName is required when webhook.tls.mode=existingSecret" -}}
{{- end -}}
{{- end -}}
{{- if eq $mode "certManager" -}}
{{- if not .Values.certManager.enabled -}}
{{- fail "webhook.tls.mode=certManager requires certManager.enabled=true" -}}
{{- end -}}
{{- $tlsCertManager := $tls.certManager | default dict -}}
{{- $issuerRef := $tlsCertManager.issuerRef | default dict -}}
{{- $issuerName := default "selfsigned-issuer" $issuerRef.name -}}
{{- $issuerKind := default "Issuer" $issuerRef.kind -}}
{{- $issuerGroup := default "cert-manager.io" $issuerRef.group -}}
{{- if not $issuerName -}}
{{- fail "webhook.tls.certManager.issuerRef.name must not be empty when webhook.tls.mode=certManager" -}}
{{- end -}}
{{- if not $issuerKind -}}
{{- fail "webhook.tls.certManager.issuerRef.kind must not be empty when webhook.tls.mode=certManager" -}}
{{- end -}}
{{- if not $issuerGroup -}}
{{- fail "webhook.tls.certManager.issuerRef.group must not be empty when webhook.tls.mode=certManager" -}}
{{- end -}}
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
