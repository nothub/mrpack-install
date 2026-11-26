package e2e

const (
	imagePostgres   = "postgres:15-alpine"
	imageRedis      = "redis:alpine"
	imageClickhouse = "clickhouse/clickhouse-server"
	imageRedpanda   = "redpandadata/redpanda:v26.1.10"
	imageBusybox    = "busybox:stable"
	imageTemporal   = "temporalio/temporal:1.10.0"
	// Renovate bumps this digest (renovate.json). Each bump means re-checking
	// seedDump, metaSchema, metaSeed and allScopes below against the image's
	// org.opencontainers.image.revision label, which TestMain logs.
	imageLabrinth = "ghcr.io/modrinth/labrinth:prod@sha256:d173735a9a861a02de100382fd649e9988e0694648b6e47d935b7e8de5937f9c"
)

const (
	// seedDump is from apps/labrinth/fixtures/ at that revision, vendored in
	// testdata.
	seedDump = "labrinth-seed-data-202508052143.sql"
	// metaSeed is from apps/labrinth/fixtures/ at that revision too. It fills
	// the Minecraft metadata database labrinth reads game versions from.
	metaSeed = "minecraft-meta-seed-data-202508052143.sql"
	// allScopes sets every bit of Scopes in apps/labrinth/src/models/v3/pats.rs.
	// labrinth reads a mask with unknown bits as no scopes at all.
	allScopes = 1<<52 - 1
	// User IDs from the seed dump.
	seedUserID  = 905016946785301
	seedAdminID = 103587649610509
)

// metaSchema is the Minecraft metadata database schema, from
// apps/meta-generator/toasty/migrations/ at that revision, vendored in
// testdata with a meta-generator- prefix. labrinth reads this database but
// never migrates it, the meta-generator service does.
var metaSchema = []string{
	"meta-generator-20261014_180018_initial.sql",
	"meta-generator-20261018_170446_maven_exports.sql",
}
