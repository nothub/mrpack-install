---
id: 15
status: claimed
blocked-by: [03, 09]
---
# CI e2e job

**Delivers:** Every push and PR runs the unit suite under the race detector and the e2e suite, and both gate the release.

- [ ] The `check` job runs `go test -v -race -vet=all ./...`. A new `e2e` job on ubuntu-24.04 runs `go test -v -race -tags e2e ./e2e/...`. The release job has `needs: [check, e2e]`. Both jobs are green on the PR that adds them
- [ ] The Answer section notes a manual step: branch protection on `trunk` must require `check` and `e2e`, which the maintainer sets on GitHub

## Answer
`check` now runs under `-race`. A new `e2e` job runs the build-tagged e2e suite, and `release` needs both jobs. The `e2e` checkout skips `fetch-depth: 0` because only goreleaser needs full history. actionlint passes. Both commands pass locally: the e2e suite took 24 s with images already cached.

The `e2e` build tag keeps the e2e package out of `check`'s `./...`, so nothing runs twice.

Unmeasured: a cold runner pulls about 2.3 GB of images inside the test binary, under `go test`'s default 10m `-timeout`, together with TestMain's `go build`. If the e2e job times out, look there first (raise `-timeout`, or pre-pull images).

**Manual step:** branch protection on `trunk` must require the `check` and `e2e` status checks. The maintainer sets this in the GitHub repo settings.
