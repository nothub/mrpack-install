//go:build e2e

package e2e

const (
	imagePostgres   = "postgres:15-alpine"
	imageRedis      = "redis:alpine"
	imageClickhouse = "clickhouse/clickhouse-server"
	imageRedpanda   = "redpandadata/redpanda:v26.1.10"
	imageBusybox    = "busybox:stable"
	// Renovate bumps this digest (renovate.json). Each bump means re-checking
	// seedDump and allScopes below against the image's
	// org.opencontainers.image.revision label, 8b1d921 for this digest.
	imageLabrinth = "ghcr.io/modrinth/labrinth:prod@sha256:954c2e5864fd1dea37ff96671036ef54c64e7ccee0d72ed6b75d4456302c1a22"
)

const (
	// seedDump is apps/labrinth/fixtures/ at that revision, vendored in testdata.
	seedDump = "labrinth-seed-data-202508052143.sql"
	// allScopes sets every bit of Scopes in apps/labrinth/src/models/v3/pats.rs.
	// labrinth reads a mask with unknown bits as no scopes at all.
	allScopes = 1<<48 - 1
	// User IDs from the seed dump.
	seedUserID  = 905016946785301
	seedAdminID = 103587649610509
)
