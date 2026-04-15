{{- define "dronerx.supportbundle" -}}
apiVersion: troubleshoot.sh/v1beta2
kind: SupportBundle
metadata:
  name: {{ include "dronerx.fullname" . }}-supportbundle
spec:
  collectors:
    {{- /* 3.2: Per-component log collectors */}}
    - logs:
        name: dronerx/api-logs
        selector:
          - app.kubernetes.io/name={{ include "dronerx.name" . }}
          - app.kubernetes.io/component=api
        namespace: {{ .Release.Namespace }}
        limits:
          maxLines: 10000
          maxAge: 72h
    - logs:
        name: dronerx/frontend-logs
        selector:
          - app.kubernetes.io/name={{ include "dronerx.name" . }}
          - app.kubernetes.io/component=frontend
        namespace: {{ .Release.Namespace }}
        limits:
          maxLines: 5000
          maxAge: 48h
    {{- if eq .Values.postgresql.enabled true }}
    - logs:
        name: dronerx/postgres-logs
        selector:
          - cnpg.io/cluster={{ include "dronerx.fullname" . }}-db
        namespace: {{ .Release.Namespace }}
        limits:
          maxLines: 5000
          maxAge: 48h
    {{- end }}
    - logs:
        name: dronerx/nats-logs
        selector:
          - app.kubernetes.io/name={{ .Release.Name }}-nats
        namespace: {{ .Release.Namespace }}
        limits:
          maxLines: 5000
          maxAge: 48h
    {{- /* 3.3: Health endpoint HTTP collector */}}
    - http:
        collectorName: dronerx-health
        get:
          url: http://{{ include "dronerx.fullname" . }}-api.{{ .Release.Namespace }}.svc.cluster.local:{{ .Values.service.apiPort }}/healthz
  analyzers:
    {{- /* 3.3: Health endpoint textAnalyze */}}
    - textAnalyze:
        checkName: Application Health Endpoint
        fileName: dronerx-health.json
        regex: '"status":\s*"ok"'
        outcomes:
          - fail:
              when: "false"
              message: |
                Application health endpoint returned unhealthy status or is unreachable.
                The /healthz endpoint checks database and NATS connectivity.
                Check pod logs: kubectl logs -l app.kubernetes.io/component=api -n {{ .Release.Namespace }}
          - pass:
              when: "true"
              message: Application is healthy — database and NATS connections are working.
    {{- /* 3.4: Status analyzers for all app workloads */}}
    - deploymentStatus:
        name: {{ include "dronerx.fullname" . }}-api
        namespace: {{ .Release.Namespace }}
        outcomes:
          - fail:
              when: "< 1"
              message: |
                The DroneRx API deployment has no available replicas.
                The application is unavailable — users cannot place or track orders.
                Run: kubectl describe deployment {{ include "dronerx.fullname" . }}-api -n {{ .Release.Namespace }}
          - pass:
              message: DroneRx API is running.
    - deploymentStatus:
        name: {{ include "dronerx.fullname" . }}-frontend
        namespace: {{ .Release.Namespace }}
        outcomes:
          - fail:
              when: "< 1"
              message: |
                The DroneRx frontend deployment has no available replicas.
                Users cannot load the web interface.
                Run: kubectl describe deployment {{ include "dronerx.fullname" . }}-frontend -n {{ .Release.Namespace }}
          - pass:
              message: DroneRx frontend is running.
    - statefulsetStatus:
        name: {{ .Release.Name }}-nats
        namespace: {{ .Release.Namespace }}
        outcomes:
          - fail:
              when: "< 1"
              message: |
                The NATS StatefulSet has no ready replicas.
                Real-time order tracking and event streaming are unavailable.
                Run: kubectl describe statefulset {{ .Release.Name }}-nats -n {{ .Release.Namespace }}
          - pass:
              message: NATS is running.
    {{- /* 3.5: Known failure patterns — dependency connectivity errors */}}
    - textAnalyze:
        checkName: Application Dependency Failures
        fileName: dronerx/api-logs/*/api.log
        regex: 'metrics: send failed|metrics: marshal error|"level":"ERROR".*database|"level":"ERROR".*nats|NATS not connected'
        outcomes:
          - fail:
              when: "true"
              message: |
                Dependency communication failures detected in API logs (last 10,000 lines / 72h).
                This may indicate the Replicated SDK, database, or NATS is down or unreachable.
                Remediation:
                1. Check SDK pod: kubectl get pods -l app.kubernetes.io/name={{ include "dronerx.name" . }}-sdk -n {{ .Release.Namespace }}
                2. Check database pods: kubectl get pods -l cnpg.io/cluster={{ include "dronerx.fullname" . }}-db -n {{ .Release.Namespace }}
                3. Check NATS pods: kubectl get pods -l app.kubernetes.io/name={{ .Release.Name }}-nats -n {{ .Release.Namespace }}
                4. Check API configmap: kubectl get configmap {{ include "dronerx.fullname" . }}-api-config -n {{ .Release.Namespace }} -o yaml
          - pass:
              when: "false"
              message: No dependency communication failures detected in recent logs.
    {{- /* 3.6: Storage class and node readiness */}}
    - storageClass:
        checkName: Default Storage Class Present
        outcomes:
          - fail:
              when: "== false"
              message: |
                No default storage class is configured on this cluster.
                DroneRx requires a default storage class for PostgreSQL persistent volume claims.
                Configure a default storage class:
                  kubectl patch storageclass <name> -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}'
          - pass:
              message: A default storage class is available.
    - nodeResources:
        checkName: All Nodes Ready
        outcomes:
          - fail:
              when: "nodeCondition(Ready) == False"
              message: |
                One or more cluster nodes are not in the Ready state.
                Application pods may fail to schedule or run.
                Check node status: kubectl get nodes
                Check node events: kubectl describe node <node-name>
          - pass:
              message: All nodes are Ready.
{{- end }}
