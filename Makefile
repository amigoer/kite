.PHONY: all build build-web build-go test test-race lint clean install release tidy

GO         ?= go
BIN_DIR    := bin
BIN_NAME   := kite
PKG        := ./cmd/kite
VERSION    := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"

all: build

build: build-web build-go

build-web:
	@if [ -d web ] && [ -f web/package.json ]; then \
		cd web && npm install --silent && npm run build; \
	else \
		mkdir -p web/dist && echo "<!-- placeholder -->" > web/dist/index.html; \
	fi

build-go:
	@mkdir -p $(BIN_DIR)
	$(GO) build $(LDFLAGS) -o $(BIN_DIR)/$(BIN_NAME) $(PKG)

test:
	$(GO) test ./... -cover

test-race:
	$(GO) test ./... -race -cover

lint:
	golangci-lint run

tidy:
	$(GO) mod tidy

clean:
	rm -rf $(BIN_DIR) web/dist web/node_modules

install: build
	$(GO) install $(LDFLAGS) $(PKG)

release:
	goreleaser release --clean
