---
id: 14
status: done
blocked-by: [01, 13]
---
# E2E `server` subcommand for vanilla, fabric and paper

**Delivers:** Every installer that needs no Java is exercised through the binary.

- [x] Parallel subtests run `mrpack-install server <flavor> --minecraft-version 1.21.1 --server-dir <tmp>` with the binary from TestMain, and assert exit 0 and a non-empty jar

## Answer
`e2e/server_test.go` runs `server vanilla|fabric|paper --minecraft-version 1.21.1 --server-dir <tmp>` through the TestMain binary as parallel subtests. Each one asserts exit 0 and exactly one non-empty `*.jar`. Without `--server-file` each flavor names the jar after its download (`server.jar`, `paper-1.21.1-<build>.jar`, `fabric-server-mc…jar`), so the test globs.

Red was shown by swapping the pre-ticket-01 Paper v2 installer back in. The paper subtest failed with `http status 410`, while vanilla and fabric passed.

The test never touches labrinth, but `TestMain` is package-wide, so `-run TestServer` still boots the stack (about 22 s). The spec's one-package layout accepts that cost.

Review approved. Deferred: the exec/CombinedOutput/Fatalf block duplicates `install()` in `install_test.go`. With two call sites a shared helper doesn't pay yet. Only vanilla verifies jar content (Mojang SHA-1). Fabric and paper verify nothing, so the size check is all the test has for them, and fixing that is a production change outside this ticket.
