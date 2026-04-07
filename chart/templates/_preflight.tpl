{{- define "dronerx.preflight" -}}
apiVersion: troubleshoot.sh/v1beta2
kind: Preflight
metadata:
  name: {{ include "dronerx.fullname" . }}-preflight
spec:
  analyzers: []
{{- end }}
