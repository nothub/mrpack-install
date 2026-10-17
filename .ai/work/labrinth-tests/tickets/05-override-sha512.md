---
id: 05
status: done
blocked-by: [04]
---
# OverrideHashes computes the correct SHA-512

**Delivers:** Override SHA-512 hashes, which update's packstate uses, match the file content. Today they're the hash of empty input, because the SHA-1 pass drains the reader.

- [x] A test asserts that OverrideHashes' SHA-512 for the example pack's overrides equals the sha512 of their content. It fails before the fix
- [x] Each hash is computed over the full content

## Answer
`OverrideHashes` reads each override entry once with `io.ReadAll`, then hashes the bytes through a fresh `bytes.NewReader` per algorithm. `TestOverrideHashes` now asserts SHA-512 against a `sha512sum` literal for `"\n"`. Before the fix it got `cf83e1…`, the SHA-512 of empty input, which confirms the drained-reader diagnosis.

Review: both axes approved. Fixed: the new read-error return skipped `r.Close()`, so the reader is now closed before the error check. Deferred: dropping `chksum` for `crypto/sha1`/`crypto/sha512` + `hex` (one pass, no double buffering) is a larger change than the bug needs, and override files are small. Pre-existing and out of scope: `ExtractOverrides` leaks readers on error paths, and `OverrideHashes` calls `log.Fatalln` from library code.
