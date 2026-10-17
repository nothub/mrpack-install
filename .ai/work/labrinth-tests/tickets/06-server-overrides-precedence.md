---
id: 06
status: done
blocked-by: [05]
---
# server-overrides win over overrides regardless of zip order

**Delivers:** For a path present in both `overrides/` and `server-overrides/`, the server-overrides copy ends up on disk and in OverrideHashes, whatever order the zip stores them in.

- [x] Offline test with a fixture zip built in the test, storing `server-overrides/x` before `overrides/x`. After ExtractOverrides (under `t.Chdir`), `x` holds the server-overrides content, and OverrideHashes reports the server-overrides hash for `x`. It fails before the fix
- [x] The same assertions pass with the opposite entry order

## Answer
`iterServerOverrides` in `modrinth/mrpack/overrides.go` walks the zip twice: `overrides/` entries first, then `server-overrides/`. `ExtractOverrides` and `OverrideHashes` both go through it, so the later write or map entry wins on disk and in the hashes. `override.server()` lost its last caller and was removed. `TestServerOverridesWin` builds the fixture zip in both entry orders. Before the fix, the server-first order failed on content, SHA-1 and SHA-512.

Surprise: `OverrideHashes` had the same archive-order bug as `ExtractOverrides`. The filter they shared turned into the shared fix.

Review: both axes approved. Applied the nit to run subtests from a slice instead of a map, so they run in a stable order. Deferred, all pre-existing and out of scope:
- `ExtractOverrides` leaks `fileReader`/`outFile` on early error returns.
- `files.AssertSafe` calls `log.Fatalln` instead of returning an error.
- The `pathPrefixClient` branch in `prefix()` is unreachable.
- Duplicate names within one prefix still resolve last-in-zip-wins.
