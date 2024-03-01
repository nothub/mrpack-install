# Contributing

## Build

To build a development version, run:

```sh
goreleaser build --clean --snapshot
```

## Release

To build a snapshot release, run:

```sh
goreleaser release --clean --snapshot
```

To build and publish a full release, run:

```sh
git tag v0.42.0-indev
goreleaser release --clean --fail-fast
```
