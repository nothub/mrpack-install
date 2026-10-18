---
id: 16
status: done
blocked-by: [09]
---
# Renovate tracks the labrinth image digest

**Delivers:** A new labrinth prod digest arrives as a Renovate PR.

- [x] A `customManagers` regex entry in renovate.json matches the image ref in the e2e image-ref file (docker datasource, digest pinning). The config validates
- [x] A Renovate dry run (local platform, in a container) shows the dependency extracted and a digest update proposed. If `minimumReleaseAge` blocks digest updates that have no release timestamp, a packageRule exempts this dependency
- [x] A comment at the image ref says the vendored seed dump and the PAT scope mask must match the image's revision label

## Answer
A `customManagers` regex entry in `renovate.json` matches `imageLabrinth` in `e2e/images_test.go` (docker datasource, `currentDigest` capture). `renovate-config-validator --strict` passes. A packageRule sets `minimumReleaseAge: null` for `ghcr.io/modrinth/labrinth` and adds `prBodyNotes` telling the bump PR's reviewer to re-check `seedDump` and `allScopes`. The two comments about the revision label merged into one at the image ref.

Verified with Renovate 44.167.1 in a container (`platform=local`, `dry-run=lookup`). Before the change only github-actions and gomod were extracted. After it, the regex manager extracted the labrinth dep. A scratch copy pinned to an older real digest (tag `cal-dev-588`) proposed the update to the current prod digest on branch `renovate/ghcr.io-modrinth-labrinth-prod`.

Surprises:
- The anticipated block was real. Renovate 44 defaults `minimumReleaseAgeBehaviour` to `timestamp-required`, and the prod digest has no release timestamp, so the update sat at `pendingChecks: true`.
- `timestamp-optional` was tried first and worked, but the review pointed out that `prod` is a mutable tag: if a timestamp ever appeared, the 7-day clock would restart on every deploy and could hold updates forever without a trace. It also logged a WARN on every run. `minimumReleaseAge: null` states the intent and runs clean.
- gofmt drops `imageLabrinth` out of the `=` alignment group, because a comment line starts a new section. The regex uses `\s*` around `=`, so it survives either form.

Deferred: whether a real PR opens can only be confirmed by the first Renovate run on the hosted platform after a prod deploy.
