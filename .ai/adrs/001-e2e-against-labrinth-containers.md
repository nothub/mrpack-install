# ADR-001: E2E tests run against a real labrinth in containers

## Status
Accepted

## Date
2026-10-18

## Context
15 of 19 tests called api.modrinth.com or maven.quiltmc.org and asserted live upstream data, so CI went red on upstream changes and outages. The install path (`cmd`, `server`, `update`, `web/download`) had no tests. The Paper installer had been broken since PaperMC sunset its v2 API, and nothing noticed.

## Decision
Modrinth behaviour is tested against labrinth, Modrinth's own backend, booted from `ghcr.io/modrinth/labrinth:prod` pinned by digest. One `e2e` package behind the `e2e` build tag starts the stack once per run with testcontainers-go core containers: postgres, redis, clickhouse, redpanda, labrinth, and a busybox `httpd` that serves labrinth's storage volume as the CDN. The tests call `ModrinthClient` black-box and run the built binary as a subprocess. Everything else is a unit test that runs offline under a plain `go test ./...`.

Seeding is split by stability. Base data (loaders, game versions, users) comes from labrinth's own seed dump, vendored from the commit in the image's `org.opencontainers.image.revision` label. Projects and versions go through the public v2 API, so labrinth's validators check every fixture. Uploads use an admin PAT, because a regular user may create only 10 projects per day and admins have no limit.

## Alternatives Considered

### Keep testing against api.modrinth.com
- Rejected: the flakiness this effort set out to remove.

### httptest fakes of the Modrinth API
- Pros: no Docker, fast.
- Rejected: a fake encodes our reading of the API and drifts silently. It would also miss labrinth's upload validation (modpack domain whitelist, duplicate file hashes), which shapes every pack the tests build.

### Seeding by SQL only
- Rejected: it couples the fixtures to labrinth's internal schema, which changes with every image bump. The public API only changes with the API.

## Consequences
- The e2e suite needs Docker, about 2.3 GB of images and about 2.6 GiB of RAM. Contributors without Docker still run the unit tests.
- Server jars still download live from Mojang, Fabric meta and PaperMC. Quilt, Forge and NeoForge need Java and stay untested.
- An image bump must re-check `seedDump`, `allScopes` and the seed user IDs in `e2e/images_test.go` against the new revision. labrinth reads a scope mask with unknown bits as no scopes, so a stale mask fails every authenticated call.
- Renovate bumps the digest with no `minimumReleaseAge`. `prod` is a mutable tag whose digests carry no release timestamp, so an age gate would hold every update.
- testcontainers-go raised `golang.org/x/sys` and `x/text` in the shipped binary through MVS.
