package e2e

// Renovate bumps every image here (renovate.json). Its regex expects each
// value as name:tag@digest.
const (
	imagePostgres   = "postgres:15-alpine@sha256:e18480a22d6e48f4d816a6696769b1ec09ac73a824c5649e49c7d8b684913f3d"
	imageRedis      = "redis:8-alpine@sha256:5af80107c1eeecb484a2269b80a26d12bfa6521787da4663bd9f19f60a163483"
	imageClickhouse = "clickhouse/clickhouse-server:26.11@sha256:88ddb6e4ee3c73760ed8ea30777d6f993eaa3f9e8062e3605a69f84ed0717258"
	imageRedpanda   = "redpandadata/redpanda:v26.1.10@sha256:394cee6e2076a23f9b7b01f5d68228d52429fe683fd087de11c8a68ff5287037"
	imageBusybox    = "busybox:1.36@sha256:73aaf090f3d85aa34ee199857f03fa3a95c8ede2ffd4cc2cdb5b94e566b11662"
	imageTemporal   = "temporalio/temporal:1.10.1@sha256:b228dadf9cdb6c7c02a1394b89dd76f827aac9760e458fce42f0899a10141225"
	// Each labrinth bump means re-checking seedDump, metaSchema, metaSeed and
	// allScopes below against the image's org.opencontainers.image.revision
	// label, which TestMain logs.
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
