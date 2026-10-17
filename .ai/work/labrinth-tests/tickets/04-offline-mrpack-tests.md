---
id: 04
status: done
blocked-by: []
---
# Offline mrpack tests replace dead test files

**Delivers:** Pack parsing has offline coverage, and no commented-out test files remain.

- [x] `testdata/example.mrpack` moves to `modrinth/mrpack/testdata/`. The `.gitignore` exception covers any `testdata/` dir
- [x] Tests cover ReadIndex, ExtractOverrides and the OverrideHashes SHA-1 values against that pack. The ExtractOverrides test uses `t.Chdir(serverDir)` and doesn't run in parallel, because ExtractOverrides checks paths relative to the CWD (cmd chdirs before calling it). It asserts that overrides and server-overrides land and client-overrides don't
- [x] `mojang/api_test.go` is deleted, and the commented-out mrpack test is replaced

## Answer
`modrinth/mrpack/mrpack_test.go` now holds three offline tests: `TestReadIndex`, `TestExtractOverrides` (serial, `t.Chdir` into an absolute temp server dir, mirroring `cmd/root.go`), and `TestOverrideHashes`. The pack moved to `modrinth/mrpack/testdata/`. The `.gitignore` exception is now `!/**/testdata/*.mrpack`. `mojang/api_test.go` is gone.

Surprises:
- Every override file in the example pack holds the same single `\n` byte, so content checks can't tell layers apart. The tests assert which files exist and don't assert which layer won. Ticket 06 builds its own fixture for precedence anyway.
- Mutation check: letting client overrides through and swapping SHA-1 for MD5 both turned the tests red. The same run printed SHA-512 `cf83e135…` for every override, the empty-input hash, which confirms the ticket 05 bug. `TestOverrideHashes` deliberately asserts SHA-1 only. Ticket 05 should add the SHA-512 assertion there.

Review: both axes approved. The standards reviewer asked for the mojang deletion in a separate commit. Declined, because the deletion is this ticket's criterion 3 and the workflow is one commit per ticket. Nits deferred: `TestReadIndex` could compare against a want struct, and it pins only `Files[0]`. Both are fine for coverage at this size.
