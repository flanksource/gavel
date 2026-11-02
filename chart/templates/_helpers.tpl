{{/*
Names, labels and checked settings shared by the gavel chart's templates.
*/}}

{{- define "gavel.name" -}}
{{- .Chart.Name | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "gavel.fullname" -}}
{{- if contains .Chart.Name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name .Chart.Name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}

{{- define "gavel.labels" -}}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" }}
{{ include "gavel.selectorLabels" . }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/* gavel serve's default --serve-pod-selector matches app.kubernetes.io/name=gavel. */}}
{{- define "gavel.selectorLabels" -}}
app.kubernetes.io/name: {{ include "gavel.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "gavel.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "gavel.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{- define "gavel.postgresName" -}}
{{- printf "%s-postgresql" (include "gavel.fullname" .) | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/* The Secret holding the rendered ~/.gavel.yaml. */}}
{{- define "gavel.configSecretName" -}}
{{- default (printf "%s-config" (include "gavel.fullname" .)) .Values.storage.existingSecret }}
{{- end }}

{{/* Fails the render when a hosted install lacks a setting it cannot start without. */}}
{{- define "gavel.validate" -}}
{{- $_ := required "publicURL is required: the origin users reach gavel at, such as https://gavel.example.com" .Values.publicURL }}
{{- $_ := required "clerk.issuer is required: the Clerk Frontend API URL" .Values.clerk.issuer }}
{{- $_ := required "clerk.publishableKey is required: the Clerk publishable key (pk_...)" .Values.clerk.publishableKey }}
{{- $_ := required "clerk.clientId is required: the public OAuth client ID gavel login uses" .Values.clerk.clientId }}
{{- if .Values.secret.create }}
{{- $_ := required "clerk.secretKey is required: the Clerk secret key (sk_...), or set secret.create=false and bring the Secret" .Values.clerk.secretKey }}
{{- end }}
{{- if .Values.githubApp.id }}
{{- $_ := required "githubApp.slug is required with githubApp.id: the App's URL name, as in github.com/apps/<slug>" .Values.githubApp.slug }}
{{- $_ := required "githubApp.privateKeySecretRef.name is required with githubApp.id: the Secret holding the App's private key PEM" .Values.githubApp.privateKeySecretRef.name }}
{{- end }}
{{- if .Values.sandbox.enabled }}
{{- $_ := required "sandbox.substrate.gavelName is required with sandbox.enabled: the name substrate's credential provider knows this gavel by, e.g. beta" .Values.sandbox.substrate.gavelName }}
{{- $_ := required "sandbox.substrate.credentialProviderToken.secretName is required with sandbox.enabled: the Secret holding the token substrate's credential provider fetches model keys with" .Values.sandbox.substrate.credentialProviderToken.secretName }}
{{- end }}
{{- if and .Values.sandbox.enabled (not (regexMatch "@sha256:[0-9a-f]{64}$" .Values.sandbox.image)) }}
{{- fail (printf "sandbox.image must be pinned by digest, as repo@sha256:<digest>; substrate runs images only by digest, got %q" .Values.sandbox.image) }}
{{- end }}
{{- if and (eq .Values.storage.type "gcs") (not .Values.storage.existingSecret) }}
{{- $_ := required "storage.gcs.credentials is required for gcs storage: the service-account key JSON" .Values.storage.gcs.credentials }}
{{- end }}
{{- if and (has .Values.storage.type (list "s3" "gcs")) (not .Values.storage.existingSecret) }}
{{- $_ := required (printf "storage.bucket is required for %s storage" .Values.storage.type) .Values.storage.bucket }}
{{- end }}
{{- if not (has .Values.storage.type (list "local" "s3" "gcs")) }}
{{- fail (printf "storage.type must be local, s3 or gcs, got %q" .Values.storage.type) }}
{{- end }}
{{- end }}

{{/* The Secret holding CLERK_SECRET_KEY and GAVEL_SECRET_KEY. */}}
{{- define "gavel.secretName" -}}
{{- default (include "gavel.fullname" .) .Values.secret.name }}
{{- end }}
