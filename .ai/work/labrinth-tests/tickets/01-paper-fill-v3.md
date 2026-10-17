---
id: 01
status: done
blocked-by: []
---
# Paper installer uses PaperMC fill v3

**Delivers:** `mrpack-install server paper` downloads a Paper server jar again. PaperMC's v2 API now answers `410 sunset`.

- [x] `mrpack-install server paper --minecraft-version 1.21.1 --server-dir <tmp>` exits 0 and writes the newest STABLE-channel Paper jar. This is a manual run. Automated coverage arrives with 14, which is blocked by this ticket
- [x] A Minecraft version with no stable Paper build returns an error that names the version, not a panic
- [x] No reference to `api.papermc.io/v2` remains

## Answer
`server/paper.go` now calls `fill.papermc.io/v3/projects/paper/versions/{mc}/builds?channel=STABLE`. It takes the first build, because the list is newest first, and downloads its `server:default` URL. Every error names the version.

Surprise: `builds/latest`, the endpoint the spec named, ignores the channel. For 26.1.1 it returns an ALPHA build. The spec bullet was amended to the STABLE filter.

Manual runs: 1.21.1 exits 0 and writes `paper-1.21.1-133.jar`, whose sha256 matches the API checksum. 26.1.1 (ALPHA builds only) gives `no stable paper build for 26.1.1`. 9.9.9 gives `fetching paper builds for 9.9.9: http status 404`.

Review: both axes approved. Fixed one comment wording nit. Deferred the jar sha256 check (the API sends `checksums.sha256`, and `vanilla.go` already uses `chksum.VerifyFile`), because no spec line asks for it and the v2 code never checked either.

`go test -race ./...` fails in `modrinth/api` on the rate-limit map race at `web/ratelimits.go:47`. That race predates this change and is ticket 03's job. `go test -vet=all ./...`, the command CI runs, passes.
