SHELL := /bin/sh

APP_NAME := gops
CMD_PATH := ./cmd/server
DIST_DIR := dist
WEB_DIR := web
EMBED_DIR := internal/webui/dist
WEB_FALLBACK := internal/webui/fallback.html

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
BUILD_TIME := $(shell date -u '+%Y-%m-%d %H:%M:%S UTC')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
GO_VERSION := $(shell go version | awk '{print $$3}')

LDFLAGS := -s -w \
	-X 'main.APPVersion=$(VERSION)' \
	-X 'main.GoVersion=$(GO_VERSION)' \
	-X 'main.BuildTime=$(BUILD_TIME)' \
	-X 'main.GitCommit=$(GIT_COMMIT)'

PLATFORMS := linux/amd64 darwin/arm64

IMAGE ?= gin-vue:$(VERSION)
IMAGE_PLATFORMS ?= linux/amd64,linux/arm64
BUILDX_BUILDER ?=
BUILDX := docker buildx build
ifneq ($(strip $(BUILDX_BUILDER)),)
BUILDX += --builder $(BUILDX_BUILDER)
endif

.DEFAULT_GOAL := help

.PHONY: help all deps frontend frontend-check backend build release image image-multiarch run test check wire migrate-up migrate-down migrate-status migration-create generate clean

help:
	@echo "Gops build targets:"
	@echo "  make deps                   Install backend and frontend dependencies"
	@echo "  make frontend               Build frontend resources"
	@echo "  make backend                Build backend with the current embedded resources"
	@echo "  make build                  Build frontend and one self-contained backend binary"
	@echo "  make release                Build self-contained binaries for all platforms"
	@echo "  make image                  Build a container image for the current architecture"
	@echo "  make image-multiarch        Build and push a linux/amd64,linux/arm64 image"
	@echo "  make run                    Run the backend in development mode"
	@echo "  make test                   Run backend tests"
	@echo "  make check                  Run backend tests and frontend type checking"
	@echo "  make migrate-up             Apply all database migrations"
	@echo "  make migrate-down           Roll back one database migration"
	@echo "  make migrate-status         Show database migration status"
	@echo "  make migration-create NAME=name  Create a SQL migration"
	@echo "  make generate               Generate GORM model/query code"
	@echo "  make wire                   Generate Wire dependency injection code"
	@echo "  make clean                  Remove build output"

all: build

deps:
	go mod download
	pnpm --dir $(WEB_DIR) install 

frontend:
	pnpm --dir $(WEB_DIR) build

frontend-check:
	pnpm --dir $(WEB_DIR) typecheck

backend:
	@mkdir -p $(DIST_DIR)
	go build -trimpath -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(APP_NAME) $(CMD_PATH)

# Frontend files are copied into the Go package only for the duration of the
# compile. This keeps generated assets out of the working tree while still
# producing a single executable that contains the complete web application.
build: clean frontend wire
	@set -eu; \
		cleanup() { \
			rm -rf $(EMBED_DIR); \
			mkdir -p $(EMBED_DIR); \
			cp $(WEB_FALLBACK) $(EMBED_DIR)/index.html; \
		}; \
		trap cleanup EXIT INT TERM; \
		rm -rf $(EMBED_DIR); \
		mkdir -p $(EMBED_DIR); \
		cp -R $(WEB_DIR)/dist/. $(EMBED_DIR)/; \
		mkdir -p $(DIST_DIR); \
		go build -trimpath -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(APP_NAME) $(CMD_PATH)
	@echo "Built self-contained binary: $(DIST_DIR)/$(APP_NAME)"
	@ls -lh $(DIST_DIR)/$(APP_NAME)

release: clean frontend
	@set -eu; \
		package_dir=""; \
		cleanup() { \
			if [ -n "$$package_dir" ]; then rm -rf "$$package_dir"; fi; \
			rm -rf $(EMBED_DIR); \
			mkdir -p $(EMBED_DIR); \
			cp $(WEB_FALLBACK) $(EMBED_DIR)/index.html; \
		}; \
		trap cleanup EXIT INT TERM; \
		rm -rf $(EMBED_DIR); \
		mkdir -p $(EMBED_DIR); \
		cp -R $(WEB_DIR)/dist/. $(EMBED_DIR)/; \
		mkdir -p $(DIST_DIR); \
		for platform in $(PLATFORMS); do \
			os=$${platform%/*}; \
			arch=$${platform#*/}; \
			ext=""; \
			if [ "$$os" = "windows" ]; then ext=".exe"; fi; \
			package_name="$(APP_NAME)-$$os-$$arch"; \
			package_dir="$(DIST_DIR)/$$package_name"; \
			archive="$(DIST_DIR)/$$package_name.tar.gz"; \
			rm -rf "$$package_dir"; \
			mkdir -p "$$package_dir"; \
			output="$$package_dir/$(APP_NAME)$$ext"; \
			echo "Building $$archive"; \
			CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch go build -trimpath -ldflags "$(LDFLAGS)" -o "$$output" $(CMD_PATH); \
			cp config.yaml.example "$$package_dir/config.yaml.example"; \
			tar -czf "$$archive" -C "$$package_dir" .; \
			rm -rf "$$package_dir"; \
			package_dir=""; \
		done
	@ls -lh $(DIST_DIR)/$(APP_NAME)-*.tar.gz

image:
	docker build --file container/Dockerfile --tag $(IMAGE) .

# A multi-platform image cannot be loaded into Docker's classic local image
# store. Publish the manifest and architecture-specific images to the registry.
# Example: make image-multiarch IMAGE=registry.example.com/team/gops:v1.0.0
image-multiarch:
	@test -n "$(IMAGE)" || (echo "IMAGE is required"; exit 1)
	$(BUILDX) --file container/Dockerfile --platform $(IMAGE_PLATFORMS) --tag $(IMAGE) --push .

run:
	go run $(CMD_PATH)

test:
	go test ./...

check: test frontend-check

wire:
	cd cmd/server && wire

migrate-up:
	go run $(CMD_PATH) -t up

migrate-down:
	go run $(CMD_PATH) -t down

migrate-status:
	go run $(CMD_PATH) -t status

migration-create:
	@test -n "$(NAME)" || (echo "NAME is required, for example: make migration-create NAME=add_host_index"; exit 1)
	go run $(CMD_PATH) -t create $(NAME) sql

gmodel:
	go run $(CMD_PATH) -g

clean:
	@rm -rf $(DIST_DIR)
