# devilutionGo - Justfile for build automation and development tasks

# Variables
GO := "go"
GO_TEST := "go test -count=1"
GO_BUILD := "go build"
GO_VET := "go vet"
GO_CLEAN := "go clean"
GO_GET := "go get"
GO_LINT := "golangci-lint run ./..."
GO_FMT := "gofmt -s -w ."
GO_GEN := "go generate ./..."
GO_MOD := "go mod tidy"

# Build tags
TAGS_HEADLESS := "-tags=headless"
TAGS_DEMO := "-tags=demo"

# Directories
BUILD_DIR := "./bin"
CMD_DIR := "./cmd"
INTERNAL_DIR := "./internal"
TEST_DIR := "./test"
ASSETS_DIR := "./assets"
SHADERS_DIR := "./shaders"
DOCS_DIR := "./docs"

# Main binary names
BIN_NAME := "devilutionGo"
BIN_HEADLESS := "devilutionGo-headless"

# True when the module has at least one Go package (scaffold may not exist yet)
HAS_GO_SRC := `find . -name '*.go' -not -path './bin/*' | grep -q . && echo yes || echo no`

# Default
default: help

# Help
help:
    @just --list

# =============================================================================
# Development
# =============================================================================

# Run the game (GUI)
run: build-gui
    {{BUILD_DIR}}/{{BIN_NAME}}

# Run headless (for CI/testing)
run-headless: build-headless
    {{BUILD_DIR}}/{{BIN_HEADLESS}}

# Build all binaries
build: build-gui build-headless

build-gui:
    {{GO_BUILD}} -o {{BUILD_DIR}}/{{BIN_NAME}} {{CMD_DIR}}/{{BIN_NAME}}

build-headless:
    {{GO_BUILD}} {{TAGS_HEADLESS}} -o {{BUILD_DIR}}/{{BIN_HEADLESS}} {{CMD_DIR}}/{{BIN_HEADLESS}}

# Clean build artifacts
clean:
    {{GO_CLEAN}}
    rm -rf {{BUILD_DIR}}

# =============================================================================
# Code Quality
# =============================================================================

# Format code
fmt:
    {{GO_FMT}}

# Lint (golangci-lint; skips gracefully before scaffold exists)
lint:
    @if [ "{{HAS_GO_SRC}}" = "yes" ]; then {{GO_LINT}}; else echo "lint: no Go sources yet, skipping"; fi

# Vet (skips gracefully before scaffold exists)
vet:
    @if [ "{{HAS_GO_SRC}}" = "yes" ]; then {{GO_VET}} ./...; else echo "vet: no Go sources yet, skipping"; fi

# Generate code (shaders, asset manifests, etc.)
generate:
    {{GO_GEN}}

# Tidy modules
mod-tidy:
    {{GO_MOD}}

# Validate the .beans ledger (frontmatter, dep targets, skills paths, Libraries policy)
beans-lint:
    python3 scripts/beans-lint.py .

# Run every cheap check: beans ledger + vet + lint (fast gate)
check: beans-lint vet lint

# =============================================================================
# Testing
# =============================================================================

# Run all tests (skips gracefully before scaffold exists)
test:
    @if [ "{{HAS_GO_SRC}}" = "yes" ]; then {{GO_TEST}} ./...; else echo "test: no Go sources yet, skipping"; fi

# Run tests with race detector
test-race:
    @if [ "{{HAS_GO_SRC}}" = "yes" ]; then {{GO_TEST}} -race ./...; else echo "test-race: no Go sources yet, skipping"; fi

# Run headless tests (CI mode)
test-headless:
    @if [ "{{HAS_GO_SRC}}" = "yes" ]; then {{GO_TEST}} {{TAGS_HEADLESS}} ./...; else echo "test-headless: no Go sources yet, skipping"; fi

# Run tests with coverage
test-cover:
    @if [ "{{HAS_GO_SRC}}" = "yes" ]; then {{GO_TEST}} -coverprofile=coverage.out ./...; {{GO}} tool cover -html=coverage.out -o coverage.html; else echo "test-cover: no Go sources yet, skipping"; fi

# Run specific test pattern
test-pattern pattern:
    {{GO_TEST}} -run "{{pattern}}" ./...

# Run parity tests only
test-parity:
    @if [ "{{HAS_GO_SRC}}" = "yes" ]; then {{GO_TEST}} -run "Parity|Golden" ./...; else echo "test-parity: no Go sources yet, skipping"; fi

# Run benchmarks
bench:
    {{GO_TEST}} -bench=. -benchmem ./internal/...

# =============================================================================
# Assets Pipeline (build-time only)
# =============================================================================

# Extract MPQ assets to loose files
assets-extract:
    @echo "Extracting MPQ assets..."
    {{GO}} run {{CMD_DIR}}/asset-extract/main.go

# Convert CLX/CEL/PCX to engine-ready atlases
assets-convert:
    @echo "Converting sprites to atlases..."
    {{GO}} run {{CMD_DIR}}/asset-convert/main.go

# Full asset pipeline
assets: assets-extract assets-convert

# =============================================================================
# CI / Release
# =============================================================================

# Full CI pipeline (runs in GitHub Actions)
ci: mod-tidy fmt vet lint test-headless test-parity beans-lint

# Pre-push gate: what git push runs (ledger + vet + lint + tests)
pre-push: beans-lint vet lint test

# Pre-commit checks
pre-commit: fmt vet lint test

# Release build (cross-compile via goreleaser)
release:
    goreleaser release --clean

# Release snapshot (local test)
release-snapshot:
    goreleaser release --snapshot --clean

# =============================================================================
# Documentation
# =============================================================================

# Show docs location
docs:
    @echo "Docs are in {{DOCS_DIR}}"

# =============================================================================
# Utilities
# =============================================================================

# Update all dependencies
deps-upgrade:
    {{GO_GET}} -u ./...
    {{GO_MOD}}

# Show outdated dependencies
deps-outdated:
    {{GO}} list -u -m all

# Download test fixtures (if any)
fixtures-download:
    @echo "Downloading test fixtures..."
    # Add fixture download logic here

# List all just targets
list:
    @just --list
