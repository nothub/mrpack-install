---
id: 03
status: done
blocked-by: []
---
# Rate-limit tracking is safe for concurrent requests

**Delivers:** Concurrent HTTP responses no longer race on the global rate-limit map in `web`.

- [x] Offline test: an httptest server returns `x-ratelimit-*` headers to many concurrent requests through one `web.Client`. `go test -race ./web` passes. It fails before the fix
- [x] Map access is guarded by a lock, and the lock is not held while `awaitRateLimits` sleeps. Await and update behaviour are otherwise unchanged

## Answer
`web/ratelimits.go` guards the per-host map with a package-level `sync.Mutex`. `awaitRateLimits` snapshots the entry pointer under the lock and sleeps after unlocking. This is sound because `updateRateLimits` always stores a fresh entry and never mutates a published one; a comment records that invariant. `web/ratelimits_test.go` sends 50 concurrent `GetJson` calls through one `Client` to an httptest server.

Surprises:
- Before the fix the test died with the runtime's `fatal error: concurrent map writes`, not a race-detector report. That check is best effort: without `-race`, the pre-fix code failed only 2 of 10 runs (standards review).

Deferred:
- CI runs `go test` without `-race` (`.github/workflows/ci.yaml:18`), so a reintroduced race usually slips through. Ticket 15 adds `-race` to the unit job.
- No test covers "lock not held during sleep": the server sends `remaining=300`, so the sleep branch never runs. The ticket did not ask for one.
- The test leaves its `127.0.0.1:<port>` entry in the global map. Harmless at `remaining=300`. A future test sending `remaining<=1` should reset the map in `t.Cleanup`.
- Pre-existing behaviour, out of scope: the sleep lasts the full `ResetDelay` instead of the time left, the log prints "60s seconds", and after a redirect the entry is keyed on the redirect target host.
