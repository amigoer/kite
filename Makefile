APP_NAME   := kite
BUILD_DIR  := build
WEB_DIR    := web/admin
HOOKS_DIR  := .githooks
GOFLAGS    := -trimpath

# Build identity stamped into the binary at link time and surfaced via
# /api/v1/health. Override VERSION when cutting a tagged release
# (`make build VERSION=v1.0.0`); git describe is the best fallback for ad-hoc
# dev builds.
VERSION    ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT     ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
VERSION_PKG := github.com/kite-plus/kite/internal/version
LDFLAGS    := -s -w \
              -X $(VERSION_PKG).Version=$(VERSION) \
              -X $(VERSION_PKG).Commit=$(COMMIT) \
              -X $(VERSION_PKG).Date=$(BUILD_DATE)

.PHONY: all build dev clean web-install web-build web-dev run hooks-install

## ── 生产构建 ──────────────────────────────────────────────

all: build

# make build  构建前端 + 内嵌到 Go 二进制
build: web-build
	@echo "==> Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	GOPATH="$$HOME/go" GOMODCACHE="$$HOME/go/pkg/mod" \
	  go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME) ./cmd/kite
	@echo "==> Done: $(BUILD_DIR)/$(APP_NAME)"

run: build
	./$(BUILD_DIR)/$(APP_NAME)

## ── 开发模式 ──────────────────────────────────────────────

# make dev  同时启动前端 dev server 和后端
dev: clean web-build
	@echo "==> Starting development servers..."
	@$(MAKE) dev-backend &
	@$(MAKE) dev-frontend
	@wait

dev-backend:
	GOPATH="$$HOME/go" GOMODCACHE="$$HOME/go/pkg/mod" \
	  go run -tags dev ./cmd/kite

dev-frontend:
	cd $(WEB_DIR) && npm run dev

## ── 前端 ─────────────────────────────────────────────────

web-install:
	cd $(WEB_DIR) && npm install

web-build: web-install
	@echo "==> Building frontend..."
	cd $(WEB_DIR) && npm run build

web-dev:
	cd $(WEB_DIR) && npm run dev

## ── 清理 ─────────────────────────────────────────────────

clean:
	rm -rf $(BUILD_DIR)
	rm -rf $(WEB_DIR)/dist

## ── Git Hooks ───────────────────────────────────────────

hooks-install:
	@git config core.hooksPath $(HOOKS_DIR)
	@chmod +x $(HOOKS_DIR)/pre-commit
	@chmod +x $(HOOKS_DIR)/pre-push
	@echo "==> Git hooks enabled via $(HOOKS_DIR)"
