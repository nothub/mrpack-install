---
id: 02
status: done
blocked-by: []
---
# `--host` accepts a full URL

**Delivers:** `--host http://labrinth.lan:8000` talks plain HTTP to that address. `--host api.modrinth.com` keeps using HTTPS.

- [x] An offline table test in a new `modrinth/api/client_test.go` covers a bare host, a bare host with a trailing slash, an http URL with a port, and an https URL. The base URL always ends in exactly one slash. `go test ./modrinth/api -run TestNewClient` passes. Trap: `api_test.go` already declares a package-level `client` and an `init()`, so pick other names
- [x] The `--host` flag help in `cmd` says a full URL is accepted
- [x] The README's `--host` lines are hand-edited to match. Don't regenerate with `tools/readme.go`, which refetches contributors from the GitHub API and churns that list

## Answer
`NewClient` keeps a value containing `://` as given and prepends `https://` otherwise, then trims trailing slashes and appends one. `modrinth/api/client_test.go` covers bare host, trailing slash, http with port, https, multiple trailing slashes and a path prefix. Flag help and the five README `--host` lines say "or full URL, e.g. http://labrinth.lan:8000".

Surprise: `go test -race ./modrinth/api` fails on `trunk` already, from the unguarded rate-limit map at `web/ratelimits.go:47`. Ticket 03 owns that. Without `-race` the full suite passes.

Deferred review findings (Consider, both not regressions):
- `--host ""`, `--host /` or `--host http://` yield a base URL with no host (`https:/`), so requests fail late with an unclear error. The old code gave `https:///`, equally broken. A `u.Host == ""` check would fail fast; left out as nobody asked for it.
- A query or fragment in `--host` gets the slash appended after it, malforming every endpoint URL. Unlikely input, not rejected.
