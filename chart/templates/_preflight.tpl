{{- define "dronerx.preflight" -}}
apiVersion: troubleshoot.sh/v1beta2
kind: Preflight
metadata:
  name: {{ include "dronerx.fullname" . }}-preflight
spec:
  {{- if or (not .Values.postgresql.enabled) .Values.ingress.tls.cloudflare.enabled }}
  collectors:
    {{- /* 3.1a: External DB connectivity collector (conditional) */}}
    {{- if not .Values.postgresql.enabled }}
    - run:
        collectorName: dronerx-db-check
        image: images.littleroom.co.nz/anonymous/index.docker.io/library/busybox:1.36
        command: ["sh", "-c"]
        args:
          - |
            nc -zv {{ .Values.externalDatabase.host }} {{ .Values.externalDatabase.port }} 2>&1 && echo "connected" || echo "connection_failed"
    {{- end }}
    {{- /* 3.1b: Cloudflare API connectivity collector (conditional) */}}
    {{- if .Values.ingress.tls.cloudflare.enabled }}
    - run:
        collectorName: dronerx-cloudflare-check
        image: images.littleroom.co.nz/anonymous/index.docker.io/library/busybox:1.36
        command: ["sh", "-c"]
        args:
          - |
            nc -zv api.cloudflare.com 443 2>&1 && echo "connected" || echo "connection_failed"
    {{- end }}
  {{- end }}
  analyzers:
    {{- /* 3.1a: External DB connectivity analyzer (conditional) */}}
    {{- if not .Values.postgresql.enabled }}
    - textAnalyze:
        checkName: External Database Connectivity
        fileName: dronerx-db-check.log
        regex: "connected"
        outcomes:
          - fail:
              when: "false"
              message: |
                Cannot connect to external database at {{ .Values.externalDatabase.host }}:{{ .Values.externalDatabase.port }}.
                Verify the host, port, and credentials are correct and the database is reachable from this cluster.
                Test manually: nc -zv {{ .Values.externalDatabase.host }} {{ .Values.externalDatabase.port }}
          - pass:
              when: "true"
              message: External database at {{ .Values.externalDatabase.host }}:{{ .Values.externalDatabase.port }} is reachable.
    {{- end }}
    {{- /* 3.1b: Cloudflare API connectivity analyzer (conditional) */}}
    {{- if .Values.ingress.tls.cloudflare.enabled }}
    - textAnalyze:
        checkName: Cloudflare API Connectivity
        fileName: dronerx-cloudflare-check.log
        regex: "connected"
        outcomes:
          - fail:
              when: "false"
              message: |
                Cannot reach Cloudflare API at api.cloudflare.com:443.
                cert-manager DNS-01 challenges require outbound HTTPS access to the Cloudflare API.
                Ensure firewall and proxy rules allow outbound TCP to api.cloudflare.com on port 443.
          - pass:
              when: "true"
              message: Cloudflare API is reachable.
    {{- end }}
    {{- /* 3.1c: Cluster resource checks (always) */}}
    - nodeResources:
        checkName: Cluster CPU Capacity
        outcomes:
          - fail:
              when: "sum(cpuAllocatable) < 2"
              message: |
                Insufficient CPU: cluster has less than 2 allocatable cores.
                DroneRx requires at least 2 CPU cores across all nodes.
          - warn:
              when: "sum(cpuAllocatable) < 4"
              message: |
                Cluster has fewer than 4 CPU cores. Performance may be degraded under load.
                Recommended: 4+ cores for production workloads.
          - pass:
              message: Cluster has sufficient CPU capacity.
    - nodeResources:
        checkName: Cluster Memory Capacity
        outcomes:
          - fail:
              when: "sum(memoryAllocatable) < 4Gi"
              message: |
                Insufficient memory: cluster has less than 4 GiB allocatable.
                DroneRx requires at least 4 GiB of memory across all nodes.
          - warn:
              when: "sum(memoryAllocatable) < 8Gi"
              message: |
                Cluster has less than 8 GiB memory. Production workloads may be constrained.
          - pass:
              message: Cluster has sufficient memory.
    {{- /* 3.1d: Kubernetes version check (always) */}}
    - clusterVersion:
        checkName: Kubernetes Version
        outcomes:
          - fail:
              when: "< 1.28.0"
              message: |
                Kubernetes version is not supported.
                DroneRx requires Kubernetes 1.28 or higher.
                Upgrade your cluster before installing.
                See: https://kubernetes.io/docs/tasks/administer-cluster/cluster-upgrade/
          - warn:
              when: "< 1.30.0"
              message: Kubernetes version is supported but upgrading to 1.30+ is recommended.
          - pass:
              message: Kubernetes version is supported.
    {{- /* 3.1e: Distribution check (always) */}}
    - distribution:
        checkName: Kubernetes Distribution
        outcomes:
          - fail:
              when: "== docker-desktop"
              message: |
                docker-desktop is not a supported Kubernetes distribution.
                It is a local development environment and lacks the storage and networking
                capabilities required for production use.
                Supported distributions: EKS, GKE, AKS, RKE2, k3s, OpenShift.
          - fail:
              when: "== microk8s"
              message: |
                microk8s is not a supported Kubernetes distribution.
                It lacks enterprise storage and networking support required by DroneRx.
                Supported distributions: EKS, GKE, AKS, RKE2, k3s, OpenShift.
          - pass:
              message: Kubernetes distribution is supported.
{{- end }}
