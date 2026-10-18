---
id: 08
status: done
blocked-by: []
---
# Symlink test checks Resolve errors

**Delivers:** A failing `Resolve` fails `TestResolveSymlink`.

- [x] Both `Resolve` calls check err with t.Fatal
- [x] Both the plain-file case and the symlink case compare against `filepath.EvalSymlinks(target)`, so they hold under a symlinked TMPDIR

## Answer
`TestResolveSymlink` now fails on a `Resolve` error in both cases. Both cases also compare against `filepath.EvalSymlinks(file.Name())`, computed before either `Resolve` call. The old symlink check (`resolved == link`) was a negative check and would pass any wrong result. It is now exact equality.

Red evidence: under `TMPDIR` set to a symlink, the old plain-file case failed. For the error check, a mutant `Resolve` that returns `(path, err)` for symlinks passed the old test and fails the new one.

Deferred from review:
- `IsSubpath` (`files/util.go:58`) uses `strings.HasPrefix` with no separator check. As a result, `/srv/mc-evil/x` counts as under `/srv/mc`, and `AssertSafe` inherits the hole. Its callers are `cmd/root.go:141`, `update/command.go:33`, and `modrinth/mrpack/overrides.go:63`. It needs its own ticket because it is outside this ticket's scope.
- `Resolve` skips `EvalSymlinks` when the path does not exist, so a missing leaf under a symlinked parent stays unresolved. No test covers that branch.
- Nits not taken: the `CreateTemp` handle is never closed, and `t.TempDir()` would replace it. The two assert blocks are duplicated. Both are pre-existing or borderline.
