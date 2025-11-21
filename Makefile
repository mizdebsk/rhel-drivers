MODULE      := github.com/mizdebsk/rhel-drivers
CMD_DIR     := ./cmd/rhel-drivers
BIN_DIR     := dist/bin
BIN_NAME    := rhel-drivers

VERSION     := $(shell git describe --tags --dirty --always 2>/dev/null || echo "dev")

GOFLAGS     :=
LDFLAGS     := -X main.version=$(VERSION)

all: build

dirs:
	mkdir -p $(BIN_DIR)

vendor:
	go mod tidy
	go mod vendor

build: dirs
	GOFLAGS="$(GOFLAGS)" go build \
	    -ldflags="$(LDFLAGS)" \
	    -o $(BIN_DIR)/$(BIN_NAME) \
	    $(CMD_DIR)

test:
	GOFLAGS="$(GOFLAGS)" go test ./...

clean:
	rm -rf dist

.PHONY: all build clean vendor test dirs
