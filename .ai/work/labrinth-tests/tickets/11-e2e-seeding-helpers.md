---
id: 11
status: done
blocked-by: [10]
---
# E2E seeding helpers

**Delivers:** Tests can create approved projects with versions and files in labrinth.

- [x] Helpers: create a project with an initial version from any file plus caller-supplied version data (loaders, game versions, dependencies), upload an extra version, and approve with the admin PAT
- [x] Test: a seeded mod project fetched by slug through ModrinthClient returns the right id, slug and type, and its file downloads from the cdn with a matching sha512

## Answer
`e2e/fixtures_test.go` holds the helpers. `createProject(t, projectType, slug, file, versionData)` posts `/v2/project` with one initial version and returns the v2 response decoded into `api.Project`. `uploadVersion(t, projectID, file, versionData)` posts `/v2/version` and returns an `api.Version`. `approve(t, projectID)` sends `PATCH /v2/project/{id}` `{"status":"approved"}` with the admin PAT. `versionData` carries version number, loaders, game versions and `dependency` entries. `dependency` is a local struct with `omitempty` ids, because `api.Dependency` would send `""` ids that labrinth rejects. `upload` builds the multipart form with `data` as the first part and the file part named after the file, which is what `file_parts` refers to. `fabricJar(t, modID)` builds a minimal jar holding only `fabric.mod.json`. `e2e/seeding_test.go` covers the ticket's test (id, slug and type by slug, cdn bytes equal to the upload, sha512 as reported). A second test covers `uploadVersion`: both version ids are listed. Both tests fail with 404 when `approve` is removed.

Rules checked in labrinth's source at `8b1d921`:
- A file whose SHA-1 exists in another project is rejected as a duplicate, so every fixture file needs distinct bytes. `fabricJar` embeds the mod id for that reason.
- v2 `dependencies` has no serde default, so `null` fails to parse. `newVersionCreate` sends `[]`.
- `is_draft: false` puts the project in `processing`, which needs at least one initial version. Unauthenticated reads 404 until approval.
- A missing `fabric.mod.json` is only a warning.

Surprise: regular users may create 10 projects per UTC day (`user_limits.projects_per_day`, default from migration `20260916120000_version_limits.sql`). On a fresh DB the 11th `POST /v2/project` answered `400 limit_reached`. The user chose to upload with the admin PAT, since `UserLimits::adjust_for_user` gives admins no limits. That departs from the spec's "user PAT for uploads", so spec.md now says so. `userPAT` is still used by `TestSeedPATsAuthenticate`. Version limits (150 per day, 10000 per project) do not apply to admins either.

Review: both axes approved. Fixed: a length guard before `created.Versions[0]`, and `createProject` takes `api.ProjectType` in place of a string. Deferred: renaming the generic package-wide names `file`, `upload` and `do`, which the reviewer left to the author. Rename them in ticket 13 if they clash with locals there. Also deferred: switching `seed_test.go` to `do`, which is outside this ticket.
