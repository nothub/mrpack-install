MOD_NAME = $(shell go list -m)
BIN_NAME = $(shell basename $(MOD_NAME))

out/$(BIN_NAME)-dev: clean lint test
	mkdir -p out
	go build -race -o $@

release: clean
	mkdir -p out
	GOOS=linux   GOARCH=amd64 go build -o out/$(BIN_NAME)-linux
	GOOS=linux   GOARCH=arm64 go build -o out/$(BIN_NAME)-linux-arm64
	GOOS=darwin  GOARCH=amd64 go build -o out/$(BIN_NAME)-darwin
	GOOS=darwin  GOARCH=arm64 go build -o out/$(BIN_NAME)-darwin-arm64
	GOOS=windows GOARCH=amd64 go build -o out/$(BIN_NAME)-windows

clean:
	go clean
	go mod tidy
	rm -rf out

lint:
	go vet

test:
	go test -v -parallel $(shell grep -c -E "^processor.*[0-9]+" "/proc/cpuinfo") $(MOD_NAME)/...

.PHONY: clean lint test release
