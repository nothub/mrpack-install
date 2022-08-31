MODNAME = $(shell go list -m)
BINNAME = $(shell basename $(MODNAME))
THREADS = $(shell grep -c -E "^processor.*[0-9]+" "/proc/cpuinfo")

$(BINNAME): clean lint test
	go build -race -o $@

clean:
	go clean
	go mod tidy

lint:
	go vet

test:
	go test -v -parallel $(THREADS) \
	$(MODNAME)/modrinth \
	$(MODNAME)/modrinth/mrpack


.PHONY: clean lint test
