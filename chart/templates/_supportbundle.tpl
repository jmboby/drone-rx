{{- define "dronerx.supportbundle" -}}
apiVersion: troubleshoot.sh/v1beta2
kind: SupportBundle
metadata:
  name: {{ include "dronerx.fullname" . }}-supportbundle
spec:
  collectors: []
  analyzers: []
{{- end }}
