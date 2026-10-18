---
created: 2026-10-17
updated: 2026-10-17
---

## Problem Statement

The test suite gives little confidence that mrpack-install works.

- 15 of 19 tests call live third-party services (api.modrinth.com, maven.quiltmc.org) and assert live upstream data such as fixed version IDs. CI goes red when upstream data changes or a service is down, with no change on our side.
- The install and update path (`cmd`, `server`, `update`, `web/download`) has no tests at all. Nothing runs mrpack-install end to end.
- The mrpack tests are commented out and import a package that no longer exists. The mojang test file holds only a commented-out test. `testdata/example.mrpack` is committed but unused.
- Two tests are buggy. The project 404 test calls `Error()` on a possibly nil error, and the symlink test ignores the errors `Resolve` returns.
- Nothing caught that `mrpack-install server paper` is broken. PaperMC sunset its v2 API, which now answers `410 {"error":"sunset"}`.

## Solution

Labrinth-backed tests run mrpack-install against a real labrinth (Modrinth's backend) in containers, seeded by the test itself. One `e2e` package, behind a build tag, starts the stack once per run. It tests the Modrinth API client black-box and runs the built binary end to end. The remaining unit tests run offline in a plain `go test ./...`. CI runs both, as two jobs.

Two production changes come with it. `--host` accepts a full URL, so mrpack-install can reach a plain-HTTP labrinth. The Paper installer moves to PaperMC's fill v3 API.

## Success Criteria

1. `go test ./...` passes with no network and no Docker, and makes no outbound connections.
2. `go test -race -tags e2e ./...` passes on a machine with Docker and network access, starting from no running containers.
3. With images cached, the stack is ready for the first test within 30 s. The spike measured 9.4 s on a 12-core dev machine.
4. CI runs a unit job and an e2e job on every push and pull request. Both gate the release job, and branch protection on `trunk` requires both. Branch protection is a manual GitHub setting.
5. The e2e suite installs a seeded pack with a Vanilla server and with a Fabric server. Each install leaves the server jar, the pack's files and both override layers in the server dir.
6. The e2e suite runs `mrpack-install server` for vanilla, fabric and paper, and each produces a server jar.
7. Every Modrinth API client method that had a live test has an e2e test against labrinth.
8. `mrpack-install --host api.modrinth.com` still talks HTTPS, exactly as before.
9. Renovate opens a PR when the pinned labrinth image digest changes.

## User Stories

1. As a maintainer, I want `go test ./...` to run offline, so that unit tests pass on a plane and never flake on upstream outages.
2. As a maintainer, I want one command to run the labrinth-backed suite, so that I don't have to start containers by hand.
3. As a maintainer, I want the e2e containers removed when the run ends, including on failure, so that runs don't leak containers or volumes.
4. As a maintainer, I want the e2e suite to start one stack per run, so that RAM use stays around one stack's worth (2.6 GiB measured).
5. As a maintainer, I want the e2e tests to create their own projects, versions and files through labrinth's public API, so that the seed passes labrinth's own validation and survives internal schema changes.
6. As a maintainer, I want the base data (loaders, game versions, project types, users) loaded from labrinth's own seed dump at the pinned image's commit, so that the schema matches the image.
7. As a maintainer, I want labrinth started without background tasks, so that it doesn't sync game versions from Mojang at boot and collide with the seed.
8. As a maintainer, I want projects approved by an admin token after creation, so that unauthenticated reads (what mrpack-install does) can see them.
9. As a maintainer, I want uploaded files served over HTTP from labrinth's storage volume, so that mrpack-install can download them the way it downloads from cdn.modrinth.com.
10. As a maintainer, I want the pack's download URLs to use a hostname labrinth's modpack whitelist accepts, so that pack uploads aren't rejected.
11. As a maintainer, I want the API client tests to cover project lookup by slug and ID, project 404, batch project lookup with an unknown ID, project validity check, dependencies, version listing with and without loader and game-version filters, single and batch version lookup, version-from-hash, version-from-mrpack-file, and latest game version.
12. As a maintainer, I want an E2E install of a seeded Fabric pack, so that slug resolution, pack download, server jar install, file downloads and override extraction are covered in one run.
13. As a maintainer, I want an E2E install of a seeded pack with only a Minecraft dependency, so that the Vanilla path is covered.
14. As a maintainer, I want an E2E run of `mrpack-install server` for vanilla, fabric and paper, so that every installer that needs no Java has a test.
15. As a maintainer, I want the E2E tests to run the built binary as a subprocess, so that `log.Fatal` in `cmd` can't kill the test process and cobra's flag globals don't leak between runs.
16. As a maintainer, I want the E2E tests to assert on the resulting server dir (files present, hashes, override precedence), so that they test behavior, not log output.
17. As a user of a self-hosted labrinth, I want `--host` to accept a full URL such as `http://labrinth.lan:8000`, so that I can install from an instance without TLS.
18. As a user, I want `--host api.modrinth.com` to keep working unchanged, so that the change doesn't break existing scripts.
19. As a user, I want `mrpack-install server paper` to work again, so that I can set up Paper servers.
20. As a maintainer, I want maven metadata parsing tested against a recorded `maven-metadata.xml` served locally, so that it runs offline.
21. As a maintainer, I want the mrpack index, override extraction and override hashing tested against the committed example pack, so that pack parsing has offline coverage.
22. As a maintainer, I want the symlink test to check the errors it gets, so that a failing `Resolve` fails the test.
23. As a maintainer, I want the empty mojang test file gone, so that the tree has no dead tests.
24. As a maintainer, I want the labrinth image pinned by digest and bumped by Renovate, so that runs are reproducible and upstream changes arrive as reviewable PRs.
25. As a maintainer bumping the labrinth image, I want the seed dump and the PAT scope mask to come from the same labrinth commit as the image, and to know where that commit is recorded, so that a bump doesn't silently break seeding.
26. As a contributor without Docker, I want the e2e suite excluded from a plain `go test`, so that I can still run the unit tests.

## Implementation Decisions

**Production changes (five):**
- `modrinth/api` `NewClient(host)`: if the value contains `://`, use it as the base URL as given. Otherwise prepend `https://`, as today. Normalize to one trailing slash. The `--host` flag help says a full URL is accepted. The README's `--host` lines are edited by hand to match, because regenerating them with the readme tool also refetches the contributor list from the GitHub API.
- `server` Paper installer: move from `api.papermc.io/v2` to fill v3 (`fill.papermc.io/v3/projects/paper/versions/{mc}/builds?channel=STABLE`, newest first). `builds/latest` ignores the channel and can return an ALPHA build. The download URL comes from the first build's `server:default` download entry. The `Installer` interface doesn't change.
- `modrinth/mrpack` `OverrideHashes`: compute SHA-512 over the file content. Today the SHA-1 pass drains the reader, so SHA-512 is the hash of empty input. Found by the ticket review.
- `modrinth/mrpack` `ExtractOverrides`: apply `overrides/` first, then `server-overrides/`, whatever the zip entry order. Today it writes entries in archive order, so a pack that stores `server-overrides/` first ends up with the common copy winning. Found by the ticket review.
- `web` rate-limit tracking: guard the per-host map with a lock. Today concurrent responses write it unguarded, which parallel e2e tests and concurrent downloads can both trigger. Found by the ticket review.

**E2E package (`e2e`, build tag `e2e`):**
- `TestMain` starts the stack, seeds it, builds the mrpack-install binary into a temp dir, runs the tests, then tears everything down.
- Containers are testcontainers-go core `GenericContainer`s on one dedicated network, with no testcontainers modules and no compose module:

| Container | Image | Role |
|---|---|---|
| postgres | `postgres:15-alpine` | labrinth DB |
| redis | `redis:alpine` | labrinth cache |
| clickhouse | `clickhouse/clickhouse-server` | needed at boot (labrinth runs ClickHouse migrations) |
| redpanda | `redpandadata/redpanda`, dev-container mode, `--smp 1` | needed for uploads (labrinth publishes a Delphi scan event per file, with no switch to disable it) |
| labrinth | `ghcr.io/modrinth/labrinth:prod@sha256:…` | API under test |
| cdn | `busybox:stable` running `httpd`, storage volume mounted read-only | serves uploaded files, since labrinth's local storage backend writes files but never serves them |

- Elasticsearch, Mailpit and Gotenberg are left out. The spike showed labrinth boots and serves without them.
- Startup order:
  1. Backing services start, and each one waits until ready.
  2. labrinth runs once with `--run-background-task migrations`.
  3. The seed dump and the PAT inserts load through psql inside the postgres container.
  4. labrinth starts with `--no-background-tasks --no-migrations`.
  5. Readiness is `GET /v2/tag/loader` returning 200.
- labrinth env: DB, Redis, ClickHouse and Kafka addresses on the network. `CDN_URL=http://localhost:<cdn mapped port>`. `WHITELISTED_MODPACK_DOMAINS=["localhost"]`. The CDN container has to start first so that its mapped port is known.
- Seed data: labrinth's staging seed dump, copied into the e2e package's testdata from the commit in the image's `org.opencontainers.image.revision` label. Two PATs are inserted by SQL: a user PAT and an admin PAT. The seeding helpers upload and approve with the admin PAT, because a regular user may create only 10 projects a day (ticket 11). Both use the all-scopes mask for that labrinth version (`2^48-1` at revision `8b1d921`). labrinth treats a mask with unknown bits as no scopes.
- Seeding helpers in the e2e package (create project, upload version, approve) talk to labrinth's v2 API with multipart uploads. They live only in test code.
- Test fixtures: a minimal Fabric mod jar and pack contents are built at test time, because pack download URLs contain the CDN's mapped port.

**CI:** the unit job becomes `go test -v -race -vet=all ./...`, so the unit packages run under the race detector. A second job, `e2e`, runs `go test -v -race -tags e2e ./e2e/...` on `ubuntu-24.04`, which has Docker. Nothing runs twice. Both jobs gate the release job.

**Renovate:** a `customManagers` regex entry tracks the labrinth image reference in the e2e package (docker datasource, digest pinning).

**Layout:** `testdata/example.mrpack` moves to `modrinth/mrpack/testdata/`. The `.gitignore` exception changes from root `testdata/` to any `testdata/` dir.

## Testing Decisions

- A good test here asserts what the user sees: the files in the server dir and their hashes, the client's return values and errors. It doesn't assert log lines or internal calls.
- The e2e tests are black-box. API tests go through `ModrinthClient`'s exported methods. CLI tests go through the binary's flags and the resulting server dir.
- Each e2e test seeds the projects it needs under unique slugs, so that tests can run with `t.Parallel` and don't depend on order.
- Live network calls that remain: Mojang, Fabric meta and PaperMC fill, for server jars only. Every Modrinth call goes to the local labrinth.
- Offline unit tests: `maven` uses `httptest` serving a recorded `maven-metadata.xml`. `modrinth/mrpack` reads the committed example pack. `files` and `web` keep their current tests, with the symlink test's error checks fixed.
- Prior art: table-driven style in the `web` URL test, `t.Parallel` style in the current `modrinth/api` tests. The tests use stdlib `testing` only, with no assertion library.
- The current live `modrinth/api` tests are deleted. The e2e package takes over what they covered.

## Out of Scope

- Server flavors that need Java: Quilt, Forge, NeoForge. The spike measured 3–17 s and 65–185 MB each, plus a JDK in CI.
- The `update` subcommand and `update/*` packages.
- Faking the server jar hosts (Mojang, Fabric meta, PaperMC). The jar downloads stay live.
- Elasticsearch-backed endpoints (search).
- A scheduled job that runs against the real api.modrinth.com to catch drift.
- Running the e2e suite under Podman. It targets Docker, which is what CI has.
- `web/download` and `cmd` unit tests beyond what the E2E tests cover.

## Open Questions

- CI image pull time is unmeasured. The labrinth, ClickHouse and Redpanda images total about 2.3 GB. If the e2e job takes too long, consider caching images or capping Redpanda's memory (it used 1.8 GiB).
- labrinth's rate limiter never triggered during the spike. If parallel e2e tests hit it, `RATE_LIMIT_IGNORE_IPS` is the knob.
- The e2e job fails until the Paper fix lands. Either land the Paper fix first, or land them together.
