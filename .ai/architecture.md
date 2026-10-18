# Architecture

## Testing

- `go test ./...` runs offline with no Docker. Fixtures live in each package's `testdata/`. `maven` serves a recorded `maven-metadata.xml` through `httptest` (ADR-001).
- `go test -tags e2e ./e2e/...` boots one labrinth stack per run in `TestMain`: postgres, redis, clickhouse, redpanda, labrinth and a busybox CDN. It seeds the stack, builds the binary, and tears everything down, with Ryuk reaping what a killed run leaves behind (ADR-001).
- Base data comes from labrinth's seed dump at the image revision. Each test creates its own projects under unique slugs through the v2 API with the admin PAT, then approves them, so tests run in parallel (ADR-001).
- `e2e/images_test.go` holds the labrinth digest and every value tied to its revision. Renovate bumps the digest, and the bump PR body names the values to re-check (ADR-001).
- CI runs two jobs, `check` (unit, `-race`) and `e2e`. `release` needs both (ADR-001).
