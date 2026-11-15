#!/usr/bin/env bash
# Checks that the packaged chart selects the same pods as the release manifest.
#
#   usage: hack/eh/check-chart-selectors.sh <chart.tgz> [<temporal-operator.yaml>]
#
# A Deployment's selector cannot change after creation. If the chart's selector
# differs from the manifest's, the chart cannot adopt an install made from the
# manifest: the sync fails on the immutable field. The webhook Service must
# select the same pods, or the webhooks lose their endpoints.
set -euo pipefail

chart="${1:?usage: check-chart-selectors.sh <chart.tgz> [<temporal-operator.yaml>]}"
manifest="${2:-${RELEASE_PATH:-out/release/artifacts}/temporal-operator.yaml}"

rendered=$(mktemp)
trap 'rm -f "$rendered"' EXIT
helm template temporal-operator "$chart" -n temporal-system >"$rendered"

rc=0
for q in \
  'select(.kind == "Deployment") | .spec.selector' \
  'select(.kind == "Service") | .spec.selector'; do
  want=$(yq -o=json -I=0 "$q" "$manifest")
  have=$(yq -o=json -I=0 "$q" "$rendered")
  if [ -z "$want" ] || [ "$want" != "$have" ]; then
    echo "::error::chart and manifest differ for ${q}: manifest ${want:-<none>}, chart ${have:-<none>}"
    rc=1
  fi
done
exit "$rc"
