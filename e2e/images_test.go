//go:build e2e

package e2e

const (
	imagePostgres   = "postgres:15-alpine"
	imageRedis      = "redis:alpine"
	imageClickhouse = "clickhouse/clickhouse-server"
	imageRedpanda   = "redpandadata/redpanda:v26.1.10"
	imageBusybox    = "busybox:stable"
	imageLabrinth   = "ghcr.io/modrinth/labrinth:prod@sha256:01b599426ac8b1665f6718da43ef355c4645fb1e41effcfeefde33954bb2b607"
)
