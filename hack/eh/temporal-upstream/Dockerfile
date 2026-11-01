# Not built. This pins the upstream Temporal release that
# eh-temporal-images.yaml builds, so that Dependabot opens a PR when
# temporalio releases a new server version. Merging that PR runs
# eh-temporal-upstream.yaml, which tags temporal-v<version>-eh.1 and builds
# ghcr.io/eh-homelab/temporal-{server,admin-tools}:<version>.
FROM temporalio/server:1.32.1
