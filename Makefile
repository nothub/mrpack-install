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
	go clean
	-rm -rf dist
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
	@echo "[![downloads](https://img.shields.io/github/downloads/nothub/mrpack-install/total.svg?style=flat-square&labelColor=5c5c5c&color=007D9C)](https://hub.lol/mrpack-install/releases/latest)" >> README.md
	@echo "[![discord](https://img.shields.io/discord/1149744662131777546?style=flat-square&labelColor=5c5c5c&color=007D9C)](https://discord.gg/QNbTeGHBRm)" >> README.md
	@echo "[![go pkg](https://pkg.go.dev/badge/hub.lol/mrpack-install.svg)](https://pkg.go.dev/hub.lol/mrpack-install)" >> README.md
	@echo "" >> README.md
	@echo "A cli application for installing Minecraft servers and [Modrinth](https://modrinth.com/) [modpacks](https://docs.modrinth.com/docs/modpacks/format_definition/)." >> README.md
	@echo "" >> README.md
	@echo "---" >> README.md
	@echo "## Usage" >> README.md
	@echo "#### modpack deployment" >> README.md
	@echo "\`\`\`" >> README.md
	./out/mrpack-install --help >> README.md
	@echo "\`\`\`" >> README.md
	@echo "#### modpack update" >> README.md
	@echo "\`\`\`" >> README.md
	./out/mrpack-install update --help >> README.md
	@echo "\`\`\`" >> README.md
	@echo "#### plain server deployment" >> README.md
	@echo "\`\`\`" >> README.md
	./out/mrpack-install server --help >> README.md
	@echo "\`\`\`" >> README.md
	@echo "## Install" >> README.md
	@echo "### Linux" >> README.md
	@echo "\`\`\`sh" >> README.md
	@echo "# download" >> README.md
	@echo "curl -sSL -o \"/tmp/mrpack-install\" \"https://github.com/nothub/mrpack-install/releases/download/$(VERSION)/mrpack-install-linux\"" >> README.md
	@echo "# install to a place in PATH" >> README.md
	@echo "sudo install -t \"/usr/local/bin\" \"/tmp/mrpack-install\"" >> README.md
	@echo "# run" >> README.md
	@echo "mrpack-install --help" >> README.md
	@echo "\`\`\`" >> README.md
