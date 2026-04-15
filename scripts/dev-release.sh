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

# Substitute $VERSION placeholders (same as CI does)
sed -i.bak "s|tag: \"[^\"]*\" # x-release-please-version|tag: \"${VERSION}\" # x-release-please-version|g" chart/values.yaml
sed -i.bak "s|^version:.*|version: ${VERSION}|" chart/Chart.yaml
sed -i.bak "s|^appVersion:.*|appVersion: \"${VERSION}\"|" chart/Chart.yaml
sed -i.bak "s/\$VERSION/${VERSION}/g" replicated/dronerx-chart.yaml

# Build chart dependencies
helm repo add cnpg https://cloudnative-pg.github.io/charts 2>/dev/null || true
helm repo add nats https://nats-io.github.io/k8s/helm/charts 2>/dev/null || true
helm repo update >/dev/null 2>&1

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

echo "Done. Local files restored."
