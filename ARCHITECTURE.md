# Architecture

## Testing

Modrinth behaviour is tested against labrinth, Modrinth's own backend, so no test depends on live upstream data. A fake of the Modrinth API would encode our reading of it and miss labrinth's upload validation.

- `go test ./...` runs offline with no Docker. Fixtures live in each package's `testdata/`. `maven` serves a recorded `maven-metadata.xml` through `httptest`.
- `go test -tags e2e ./e2e/...` boots one labrinth stack per run in `TestMain`: postgres, redis, clickhouse, redpanda, labrinth and a busybox CDN. It seeds the stack, builds the binary, and tears everything down, with Ryuk reaping what a killed run leaves behind.
- Base data comes from labrinth's seed dump at the image revision. Each test creates its own projects under unique slugs through the v2 API with the admin PAT, then approves them, so tests run in parallel. Fixtures go through the public API because labrinth's internal schema changes with every image bump, and its API does not.
- Every e2e test runs the binary through `mrpackInstall` in `e2e/main_test.go`. Its stdin is a pipe that stays open and never delivers data, so a run that reads stdin hangs until the per-run timeout in `mrpackInstall` fails the test. No run may read stdin.
- `e2e/images_test.go` holds the labrinth digest and every value tied to its revision. Renovate bumps the digest, and the bump PR body names the values to re-check.
- CI runs two jobs, `check` (unit, `-race`) and `e2e` (`-timeout 15m`). `release` needs both.
