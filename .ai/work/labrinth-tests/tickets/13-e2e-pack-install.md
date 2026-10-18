---
id: 13
status: done
blocked-by: [06, 11]
---
# E2E pack install

**Delivers:** The built binary installs a seeded pack by slug, end to end, against labrinth.

- [x] TestMain builds the mrpack-install binary into a temp dir. Tests run `mrpack-install <slug> --host http://localhost:<port> --server-dir <t.TempDir>`
- [x] Fabric pack: minecraft and fabric-loader deps, one mod on the cdn, plus overrides, server-overrides and client-overrides, with one path present in both overrides and server-overrides. The server dir ends up with the fabric jar, the mod with a matching sha512, an overrides-only file, the server-overrides copy of the shared path, and no client override
- [x] Vanilla pack: minecraft dep only, one file on the cdn, overrides and server-overrides. The server dir ends up with the vanilla jar, the file with a matching sha512, and both override layers

## Answer
TestMain builds the binary with `go build` into a temp dir before the stack starts, so a compile error fails in seconds. `install` runs it with `t.Context()` against labrinth and a fresh `t.TempDir`. `TestInstallFabricPack` and `TestInstallVanillaPack` seed a mod, read its cdn URL from labrinth, and build the pack from a local `packIndex` type, so a JSON tag bug in `mrpack.Index` can't hide on both sides. The Fabric pack stores `server-overrides/` before `overrides/`, so the shared path also covers the ordering fix from ticket 06. Breaking `ExtractOverrides` on purpose failed both tests on the expected assertions.

Surprises:
- The Fabric jar name ends in the latest stable installer version, so the test pins loader 0.17.2 and matches `fabric-server-mc.1.21.8-loader.0.17.2-launcher.*.jar`.
- labrinth rejects `minecraft` as a modpack loader (`not a valid variant for mrpack_loaders`). The Vanilla pack's labrinth version is tagged `fabric`. mrpack-install picks the installer from the index dependencies, so the tag has no effect.
- A failed file download is only logged, and mrpack-install still exits 0. `install` therefore logs the subprocess output on success too, or the reason would be lost behind a missing-file assertion. Found by review.

Deferred review findings:
- `packArchive` and `mrpack()` in fixtures_test.go both build an mrpack zip. Folding `mrpack()` into `packArchive` touches ticket 12's code, so it is left for a cleanup.
- The binary is built without `-race`, so its concurrent downloads get no race detector in e2e. Enabling it needs a build-tag file to detect `-race` in the test process, and the spec doesn't ask for it.
- No timeout on the subprocess beyond `go test -timeout`. The Vanilla test downloads the full Mojang server jar on every run.
