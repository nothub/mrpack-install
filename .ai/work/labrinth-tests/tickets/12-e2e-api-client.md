---
id: 12
status: done
blocked-by: [03, 07, 11]
---
# API client tests run against labrinth

**Delivers:** Every ModrinthClient method that had a live test is now covered against labrinth, and the unit suite is fully offline.

- [x] e2e tests (t.Parallel, unique slugs) cover GetProject (slug, id, 404), GetProjects (unknown id dropped), CheckProjectValidity, GetDependencies, GetProjectVersions (unfiltered, loader filter, game-version filter, empty), GetVersion, GetVersions, VersionFromHash and VersionFromMrpackFile. `go test -race -tags e2e ./e2e/...` passes
- [x] `modrinth/api/api_test.go` is deleted
- [x] `unshare -rn sh -c 'ip link set lo up && go test ./...'` passes. A fresh net namespace has `lo` down, which would break the httptest tests

## Answer
`e2e/api_test.go` holds 10 parallel tests, one per ticket item, each seeding its own `e2e-api-*` slugs. Expected values are the ids labrinth returned on create, so no test depends on seed data other than game versions and loaders. GetProjectVersions is one table over a project with a fabric/1.20.1 and a quilt/1.21.8 version: unfiltered returns both, the loader and game-version filters return one each, and `forge` returns none. The quilt validator accepts a jar that only has `fabric.mod.json`, so `fabricJar` covers both loaders. The 404 test matches the `http status 404` prefix, because `web.GetModel` appends ` - <body>` when the error body parses. GetLatestGameVersion stays covered by `seed_test.go`.

`fixtures_test.go` gained `seedMod` (create, check one version, approve) and `mrpack(t, name)`, a pack with an empty `files` list plus one override. labrinth's modpack validator rejects a pack that has neither files nor an `overrides/` entry. Ticket 13 will need a richer pack builder and can grow this one. For a `.mrpack` upload, labrinth's v2 route turns the given loaders into `mrpack_loaders` and sets the loader to `mrpack`.

`modrinth/api/api_test.go` is deleted. `unshare -rn sh -c 'ip link set lo up && go test ./...'` passes.

Each test was shown to fail on its own failure mode: two mutation batches on `endpoints.go`/`format.go` (dropped or swapped query params, wrong paths, wrong hash algorithm, first id dropped from batch lookups) failed all tests except the 404 one, which failed when GetProject swallowed its error.

Surprise: `arrayAsParam` quotes the caller's slice in place, so `GetProjects(ids)` and the `GetProjectVersions` filters mutate their arguments. The tests pass inline literals, so it does not show here. Not fixed: it is production code outside this ticket.

Review: both axes approved. Fixed: ProjectType assertion restored from the old test, stable subtest names (`slug`/`id` in place of random ids), version-count guard moved into `seedMod`, `mrpack` comment trimmed. Deferred: the old batch-lookup slug check (ids already prove the unknown-id drop); the shared `fabric1218` value, safe while tests only reassign `Dependencies`; overlap with `seeding_test.go`, outside this ticket. `modrinth/api/adrena.mrpack` is a gitignored download left behind by the deleted test and can be removed locally.
