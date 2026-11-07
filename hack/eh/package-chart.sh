#!/usr/bin/env bash
# Packages charts/temporal-operator for a fork release tag (v<upstream>-eh.<n>).
#
#   usage: hack/eh/package-chart.sh <tag> <out-dir>
#
# The chart version is the tag without its "v", and appVersion is the tag, so the
# chart's default image is ghcr.io/eh-homelab/temporal-operator:<tag>, the image
# the same release pushes. One number moves the image and the CRDs together.
#
# The CRDs come from `make artifacts` (out/release/artifacts), the same file the
# release attaches as temporal-operator.crds.yaml, not from the copy in git.
# The script packages a temporary copy, so the tree is not changed.
set -euo pipefail

tag="${1:?usage: package-chart.sh <tag> <out-dir>}"
out="${2:?usage: package-chart.sh <tag> <out-dir>}"
[[ "$tag" =~ ^v[0-9]+\.[0-9]+\.[0-9]+-eh\.[0-9]+$ ]] || {
  echo "tag must look like v0.22.0-eh.4, got: $tag" >&2
  exit 1
}
version="${tag#v}"
crds="${RELEASE_PATH:-out/release/artifacts}/temporal-operator.crds.yaml"
[ -s "$crds" ] || { echo "$crds is missing: run make artifacts first" >&2; exit 1; }

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
cp -r charts/temporal-operator "$work/"
chart="$work/temporal-operator"
cp "$crds" "$chart/crds/temporal-operator.crds.yaml"
sed -i -e "s/^version: .*/version: ${version}/" \
  -e "s/^appVersion: .*/appVersion: \"${tag}\"/" "$chart/Chart.yaml"

helm lint "$chart"
mkdir -p "$out"
helm package "$chart" -d "$out"
