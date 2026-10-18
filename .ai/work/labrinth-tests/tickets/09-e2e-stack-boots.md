---
id: 09
status: done
blocked-by: []
---
# E2E stack boots with labrinth migrated

**Delivers:** `go test -tags e2e ./e2e/...` starts labrinth and its backing services from nothing, then tears them down.

- [x] New `e2e` package with `//go:build e2e`, with image refs (labrinth pinned by digest) in a dedicated file. TestMain starts postgres, redis, clickhouse, redpanda, the busybox cdn and labrinth as testcontainers-go GenericContainers on one network. It runs labrinth's migrations task, then starts labrinth with `--no-background-tasks --no-migrations`, `CDN_URL=http://localhost:<cdn port>` and `WHITELISTED_MODPACK_DOMAINS=["localhost"]`
- [x] Teardown removes every container, the network and the volume after m.Run, and also when setup fails partway. TestMain logs the time to readiness, which stays under 30 s with images cached
- [x] Tracer test: `GET /v2/tag/game_version` answers 200. A plain `go test ./...` is unaffected

## Answer
`e2e/images_test.go` holds the image refs, `e2e/main_test.go` holds TestMain and a `stack` type, and `e2e/stack_test.go` holds the tracer test. All three files have the `e2e` build tag. `stack.start` runs linearly: network, then postgres, redis, clickhouse, redpanda, cdn, migrate, labrinth. It sets the package var `labrinthURL`. Ticket 10's seed step goes between migrate and labrinth. `stack.runContainer` records every container it creates, even one that failed its wait, and puts the container's last 20 log lines into the error. Readiness measured 13.8 to 14.9 s with images cached. With `TESTCONTAINERS_RYUK_DISABLED=true`, no labelled container, network or volume remained after a pass, after an injected migration failure, or after an injected readiness timeout.

Deviations and choices:
- The code uses `testcontainers.Run`, the current core API, which wraps `GenericContainer`. No modules are used.
- Ryuk stays enabled. It is the only cleanup when a panic, a `-timeout` kill or Ctrl-C skips TestMain's defer. Volume mounts carry the session labels, so Ryuk reaps the volume as well.
- Teardown uses `StopTimeout(0)`. Graceful stops cost about 27 s per run, because busybox `httpd` as PID 1 ignores SIGTERM and waits out the stop timeout.
- The volume is removed with the last container terminated, which is postgres, the first one started. Docker refuses to remove a volume that a container still uses.
- labrinth's env sets only `DATABASE_URL`, `REDIS_URL`, `CLICKHOUSE_URL` and `KAFKA_BOOTSTRAP_SERVERS`. The `CLICKHOUSE_*` credentials match labrinth's defaults. `BIND_ADDR` is required, since it defaults to empty. `MOCK_FILE_PATH` is set explicitly, so the volume mount does not depend on a compiled-in default.
- The clickhouse wait uses HTTP `/ping` on the mapped port. The entrypoint creates the database on a server bound to localhost, so an exec probe could pass before the real server is up.

Surprises: adding testcontainers-go v0.44.0 bumps `golang.org/x/sys` (0.29 to 0.47) and `x/text` (0.28 to 0.40) through MVS, and both are linked into the shipped binary. The go directive stays at 1.25.4.

Review: both axes approved. Fixed: `wait.ForExit()` has no default timeout, so a hung migration would block until `go test`'s kill and skip teardown. It now has a 2 min exit timeout. Also fixed: the `run`/`stack.run` name clash, `tail` ignoring `sc.Err()`, and the implicit `MOCK_FILE_PATH`. Deferred:
- redpanda stays pinned to v26.1.10, the tag the spike ran, while clickhouse, redis and busybox float as in the spec table. Ticket 16's Renovate rule covers only labrinth.
- `CDN_URL` hard-codes `localhost`, as the spec requires, so a remote `DOCKER_HOST` would serve wrong file URLs.
- If the first container's Stop fails, Terminate returns before the volume removal, and the volume leaks when Ryuk is off.
