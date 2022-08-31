MOD_NAME = $(shell go list -m)
BIN_NAME = $(shell basename $(MOD_NAME))

$(BIN_NAME): clean lint test
	go build -race -o $@

clean:
	go clean
	go mod tidy

lint:
	go vet

test:
	go test -v -parallel $(shell grep -c -E "^processor.*[0-9]+" "/proc/cpuinfo") \
	$(MOD_NAME)/modrinth \
	$(MOD_NAME)/modrinth/mrpack \
	$(MOD_NAME)/server


.PHONY: clean lint test
