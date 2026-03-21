{{/*
Common labels
*/}}
{{- define "todo-ddd-app.labels" -}}
app.kubernetes.io/part-of: myapp
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}
