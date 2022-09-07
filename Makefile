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

README.md: out/$(BIN_NAME)-dev
	@echo "# $(BIN_NAME)" > README.md
	@echo "" >> README.md
	@echo "[![Go Reference](https://pkg.go.dev/badge/$(MOD_NAME).svg)](https://pkg.go.dev/$(MOD_NAME))" >> README.md
	@echo "" >> README.md
	@echo "A cli application for installing Minecraft servers and [Modrinth](https://modrinth.com/) [modpacks](https://docs.modrinth.com/docs/modpacks/format_definition/)." >> README.md
	@echo "" >> README.md
	@echo "---" >> README.md
	@echo "" >> README.md
	@echo "#### modpack deployment" >> README.md
	@echo "\`\`\`" >> README.md
	./out/mrpack-install-dev --help >> README.md
	@echo "\`\`\`" >> README.md
	@echo "" >> README.md
	@echo "---" >> README.md
	@echo "" >> README.md
	@echo "#### server deployment" >> README.md
	@echo "\`\`\`" >> README.md
	./out/mrpack-install-dev server --help >> README.md
	@echo "\`\`\`" >> README.md

clean:
	go clean
	go mod tidy
	rm -rf out

lint:
	go vet

test:
	go test -v -parallel $(shell grep -c -E "^processor.*[0-9]+" "/proc/cpuinfo") $(MOD_NAME)/...

.PHONY: clean lint test release
