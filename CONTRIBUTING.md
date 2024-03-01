# Contributing

## Build

To build a development version, run:

```
go build -race -o out/mrpack-install
```

## Release

To build a snapshot release, run:

```sh
goreleaser healthcheck
goreleaser release --clean --snapshot
```

To build and publish a full release, run:

```sh
goreleaser healthcheck
goreleaser release --clean --fail-fast
```
