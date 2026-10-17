---
id: 07
status: done
blocked-by: []
---
# Maven metadata test runs offline

**Delivers:** maven-metadata parsing is tested without network access.

- [x] A recorded quilt-installer `maven-metadata.xml` lives under `maven/testdata/`
- [x] The test serves it through httptest and asserts the parsed fields: artifactId, latest/release, and a non-empty versions list. The maven package tests make no network calls

## Answer
`maven/meta_test.go` serves `maven/testdata/quilt-installer/maven-metadata.xml` through `httptest.NewServer(http.FileServer(...))` and asserts artifactId, latest, release (all `0.15.1` in the recording) and a non-empty versions list. The fixture is a verbatim capture of the live quilt-installer metadata, fetched 2026-10-18 (`lastUpdated` 20260801020525). The test's comment names the source URL.

Verified offline: the compiled test passes under `unshare -rn` with only `lo` up. A bare `unshare -rn` also leaves loopback down, so httptest needs `ip link set lo up` there too.

Review fixes applied: `t.Parallel()` to match `modrinth/mrpack` prior art, and the source-URL comment.

Deferred review findings:
- Stronger versions assertions (exact count, first/last entry). The ticket asks only for a non-empty list, and the parsing path is one struct tag.
- Error-path cases (404, malformed XML). These were not asked for.
- `server/quilt.go`, `forge.go` and `neoforge.go` still call `maven.FetchMetadata` against live URLs at runtime. That is production behaviour and out of scope here.
