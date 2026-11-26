# AGENTS.md

## Build

To build an executable, run:

```sh
goreleaser build --clean --snapshot --single-target
```

## Release

To build a local snapshot release, run:

```sh
goreleaser release --clean --snapshot
```

To build and publish a full release, push a semver tag (with 'v' prefix) to any branch.
