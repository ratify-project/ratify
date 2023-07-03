{{/*
Expand the name of the chart.
*/}}
{{- define "ratify.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}


{{- define "ratify.podLabels" -}}
{{- if .Values.podLabels }}
{{- toYaml .Values.podLabels | nindent 8 }}
{{- end }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "ratify.fullname" -}}
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
{{- define "ratify.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "ratify.labels" -}}
helm.sh/chart: {{ include "ratify.chart" . }}
{{ include "ratify.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "ratify.selectorLabels" -}}
app.kubernetes.io/name: {{ include "ratify.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "ratify.serviceAccountName" -}}
{{- if or .Values.azureWorkloadIdentity.clientId .Values.serviceAccount.create }}
{{- default (include "ratify.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Choose the Gatekeeper api version for Assign
*/}}
{{- define "ratify.assignGKVersion" -}}
{{- if semverCompare ">= 3.11.0" .Values.gatekeeper.version }}
apiVersion: mutations.gatekeeper.sh/v1
{{- else }}
apiVersion: mutations.gatekeeper.sh/v1beta1
{{- end }}
{{- end }}

{{/*
Choose the Gatekeeper api version for External Data Provider
*/}}
{{- define "ratify.providerGKVersion" -}}
{{- if semverCompare ">= 3.11.0" .Values.gatekeeper.version }}
apiVersion: externaldata.gatekeeper.sh/v1beta1
{{- else }}
apiVersion: externaldata.gatekeeper.sh/v1alpha1
{{- end }}
{{- end }}

{{/*
Choose the caBundle field for External Data Provider
*/}}
{{- define "ratify.providerCabundle" -}}
{{- $top := index . 0 -}}
{{- $ca := index . 1 -}}
{{- if and $top.Values.provider.tls.crt $top.Values.provider.tls.key $top.Values.provider.tls.cabundle }}
caBundle: {{ $top.Values.provider.tls.cabundle | quote }}
{{- else }}
caBundle: {{ $ca.Cert | b64enc | replace "\n" "" }}
{{- end }}
{{- end }}

{{/*
Choose the certificate/key pair to enable TLS for HTTP server
*/}}
{{- define "ratify.tlsSecret" -}}
{{- $top := index . 0 -}}
{{- if and $top.Values.provider.tls.crt $top.Values.provider.tls.key $top.Values.provider.tls.cabundle $top.Values.provider.tls.caCert $top.Values.provider.tls.caKey }}
tls.crt: {{ $top.Values.provider.tls.crt | b64enc | quote }}  
tls.key: {{ $top.Values.provider.tls.key | b64enc | quote }}
ca.crt: {{ $top.Values.provider.tls.caCert | b64enc | quote }}
ca.key: {{ $top.Values.provider.tls.caKey | b64enc | quote }}
{{- else }}
{{- $cert := index . 1 -}}
{{- $ca := index . 2 -}}
tls.crt: {{ $cert.Cert | b64enc | quote }}
tls.key: {{ $cert.Key | b64enc | quote }}
ca.crt: {{ $ca.Cert | b64enc | quote }}
ca.key: {{ $ca.Key | b64enc | quote }}
{{- end }}
{{- end }}

{{/*
Set the namespace exclusions for Assign
*/}}
{{- define "ratify.assignExcludedNamespaces" -}}
{{- $gkNamespace := default "gatekeeper-system" .Values.gatekeeper.namespace -}}
- {{ $gkNamespace | quote}}
- "kube-system"
{{- if and (ne .Release.Namespace $gkNamespace) (ne .Release.Namespace "kube-system") }}
- {{ .Release.Namespace | quote}}
{{- end }}
{{- end }}