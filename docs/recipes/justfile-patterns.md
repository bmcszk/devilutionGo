# justfile patterns for devilutionGo

Canonical command surface — all dev, CI, asset, release steps. Keep in sync with AGENTS.md.

## Pitfall: stale copied justfiles

The repo's justfile was bootstrapped from a template of an UNRELATED project ("BLACKSWAN Investment Microservice Monorepo") — header comment, binary names, and variable block all wrong, with empty recipe bodies (`build:`, `lint:`, `test:` with no commands). When scaffolding any repo, verify the justfile's header/vars reference THIS project and that every recipe has a body; `just --list` after writing is the cheap check (it renders every recipe name and catches parse errors).

## Core recipes

```just
# Variables
GO := "go"
GO_TEST := "go test -count=1"
GO_BUILD := "go build"
GO_LINT := "golangci-lint run ./..."
GO_FMT := "gofmt -s -w ."

TAGS_HEADLESS := "-tags=headless"
TAGS_DEMO := "-tags=demo"

BIN_NAME := "devilutionGo"
BIN_HEADLESS := "devilutionGo-headless"
CMD_DIR := "./cmd"
BUILD_DIR := "./bin"
```

## Dev recipes

```just
run:
	{{GO_BUILD}} -o {{BUILD_DIR}}/{{BIN_NAME}} {{CMD_DIR}}/{{BIN_NAME}}
	{{BUILD_DIR}}/{{BIN_NAME}}

run-headless:
	{{GO_BUILD}} {{TAGS_HEADLESS}} -o {{BUILD_DIR}}/{{BIN_HEADLESS}} {{CMD_DIR}}/{{BIN_HEADLESS}}
	{{BUILD_DIR}}/{{BIN_HEADLESS}}

build: build-gui build-headless
```

## Quality recipes

```just
fmt:
	{{GO_FMT}}

lint:
	{{GO_LINT}}

vet:
	{{GO_VET}} ./...

generate:
	{{GO_GEN}}

mod-tidy:
	{{GO_MOD}}
```

## Test recipes

```just
test:
	{{GO_TEST}} ./...

test-headless:
	{{GO_TEST}} {{TAGS_HEADLESS}} ./...

test-parity:
	{{GO_TEST}} -run "Parity|Golden" ./...

bench:
	{{GO_TEST}} -bench=. -benchmem ./internal/...
```

## Asset recipes (build-time only)

```just
assets-extract:
	{{GO}} run {{CMD_DIR}}/asset-extract/main.go

assets-convert:
	{{GO}} run {{CMD_DIR}}/asset-convert/main.go

assets: assets-extract assets-convert
```

## CI recipe (blocking in GitHub Actions)

```just
ci: mod-tidy fmt vet lint test-headless test-parity
```

## Release

```just
release:
	goreleaser release --clean
```

## Usage in contracts

Reference recipes by name: "Run `just ci` — it must pass." Never inline the commands; the justfile is the single source of truth.