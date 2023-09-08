MOD_NAME = $(shell go list -m)
BIN_NAME = $(shell basename $(MOD_NAME))
VERSION  ?= $(shell git describe --tags --abbrev=0 --match v[0-9]* 2> /dev/null || echo "v0.0.0")
LDFLAGS  := -X '$(MOD_NAME)/buildinfo.Tag=$(VERSION)'
LDFLAGS  += -extldflags=-static
GOFLAGS  := -tags netgo,timetzdata
GOFLAGS  += -ldflags="$(LDFLAGS)"

out/$(BIN_NAME): $(shell ls go.mod go.sum *.go **/*.go)
	go build $(GOFLAGS) -race -o out/$(BIN_NAME)

.PHONY: release
release: clean
	GOOS=linux   GOARCH=amd64 go build $(GOFLAGS) -o out/$(BIN_NAME)-linux
	GOOS=linux   GOARCH=arm64 go build $(GOFLAGS) -o out/$(BIN_NAME)-linux-arm64
	GOOS=darwin  GOARCH=amd64 go build $(GOFLAGS) -o out/$(BIN_NAME)-darwin
	GOOS=darwin  GOARCH=arm64 go build $(GOFLAGS) -o out/$(BIN_NAME)-darwin-arm64
	GOOS=windows GOARCH=amd64 go build $(GOFLAGS) -o out/$(BIN_NAME)-windows.exe

.PHONY: clean
clean:
	go clean
	-rm -rf out
	-rm -rf mc

.PHONY: check
check:
	go vet
	go test -v -parallel $(shell grep -c -E "^processor.*[0-9]+" "/proc/cpuinfo") $(MOD_NAME)/...

.PHONY: dl-stats
dl-stats:
	./tools/dl-stats.go | tee dl-stats.yaml

README.md: out/$(BIN_NAME)
	@echo "# $(BIN_NAME)" > README.md
	@echo "" >> README.md
	@echo "[![downloads](https://img.shields.io/github/downloads/nothub/mrpack-install/total.svg?style=flat-square&labelColor=5c5c5c&color=007D9C)](https://github.com/nothub/mrpack-install/releases/latest)" >> README.md
	@echo "[![discord](https://img.shields.io/discord/1149744662131777546?style=flat-square&labelColor=5c5c5c&color=007D9C)](https://discord.gg/QNbTeGHBRm)" >> README.md
	@echo "[![go pkg](https://pkg.go.dev/badge/github.com/nothub/mrpack-install.svg)](https://pkg.go.dev/github.com/nothub/mrpack-install)" >> README.md
	@echo "" >> README.md
	@echo "A cli application for installing Minecraft servers and [Modrinth](https://modrinth.com/) [modpacks](https://docs.modrinth.com/docs/modpacks/format_definition/)." >> README.md
	@echo "" >> README.md
	@echo "---" >> README.md
	@echo "" >> README.md
	@echo "#### modpack deployment" >> README.md
	@echo "\`\`\`" >> README.md
	./out/mrpack-install --help >> README.md
	@echo "\`\`\`" >> README.md
	@echo "" >> README.md
	@echo "---" >> README.md
	@echo "" >> README.md
	@echo "#### modpack update" >> README.md
	@echo "\`\`\`" >> README.md
	./out/mrpack-install update --help >> README.md
	@echo "\`\`\`" >> README.md
	@echo "" >> README.md
	@echo "---" >> README.md
	@echo "" >> README.md
	@echo "#### plain server deployment" >> README.md
	@echo "\`\`\`" >> README.md
	./out/mrpack-install server --help >> README.md
	@echo "\`\`\`" >> README.md
