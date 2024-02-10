MOD_NAME = $(shell go list -m)
BIN_NAME = $(shell basename $(MOD_NAME))

out/$(BIN_NAME): $(shell ls go.mod go.sum *.go **/*.go)
	go build -race -o out/$(BIN_NAME)
	upx --brute out/$(BIN_NAME)

.PHONY: release
release: clean
	./tools/build.sh

.PHONY: clean
clean:
	./tools/clean.sh

.PHONY: check
check:
	./tools/check.sh

.PHONY: dl-stats
dl-stats:
	./tools/dl-stats.go | tee dl-stats.yaml

README.md: out/$(BIN_NAME)
	./tools/readme.sh
