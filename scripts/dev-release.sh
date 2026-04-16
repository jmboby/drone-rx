#!/usr/bin/env bash
# Create a dev Replicated release for manual testing.
# Usage: ./scripts/dev-release.sh [channel]
# Default channel: Unstable
set -euo pipefail

CHANNEL="${1:-Unstable}"
VERSION="0.0.0-dev.$(date +%s)"

# Ensure we run from repo root
cd "$(git rev-parse --show-toplevel)"

echo "Creating dev release ${VERSION} on channel ${CHANNEL}..."

# Substitute chart version but use 'latest' for image tags (dev images don't exist)
sed -i.bak "s|tag: \"[^\"]*\" # x-release-please-version|tag: \"latest\" # x-release-please-version|g" chart/values.yaml
sed -i.bak "s|^version:.*|version: ${VERSION}|" chart/Chart.yaml
sed -i.bak "s|^appVersion:.*|appVersion: \"${VERSION}\"|" chart/Chart.yaml
sed -i.bak "s/\$VERSION/${VERSION}/g" replicated/dronerx-chart.yaml
# Also set image tags to 'latest' in the HelmChart CR
sed -i.bak "s|tag: ${VERSION}|tag: latest|g" replicated/dronerx-chart.yaml

# Build chart dependencies
helm repo add cnpg https://cloudnative-pg.github.io/charts 2>/dev/null || true
helm repo add nats https://nats-io.github.io/k8s/helm/charts 2>/dev/null || true
helm repo add traefik https://traefik.github.io/charts 2>/dev/null || true
helm repo update >/dev/null 2>&1
helm dependency build chart/

# Pull dependency charts for release bundling
mkdir -p charts/cnpg-operator charts/traefik
helm pull cnpg/cloudnative-pg --version 0.28.0 --untar --untardir charts/cnpg-operator
helm pull traefik/traefik --version 39.0.7 --untar --untardir charts/traefik

# Create the release
replicated release create \
  --version "${VERSION}" \
  --promote "${CHANNEL}" \
  --ensure-channel

echo ""
echo "Release ${VERSION} created on ${CHANNEL}."
echo "Reverting local file changes..."

# Revert substitutions
git checkout chart/values.yaml chart/Chart.yaml replicated/dronerx-chart.yaml
rm -f chart/values.yaml.bak chart/Chart.yaml.bak replicated/dronerx-chart.yaml.bak
rm -rf chart/charts/ charts/

echo "Done. Local files restored."
