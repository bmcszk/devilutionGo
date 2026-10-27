#!/usr/bin/env bash
# One lint entry point for `just lint` and CI.
# Always uses the golangci-lint v2.15.0 built with the repo's Go (go.mod),
# not whatever `golangci-lint` is on PATH — that one may be built with an older
# toolchain and reports fewer findings than CI.
set -uo pipefail

version=v2.15.0
cache="${XDG_CACHE_HOME:-$HOME/.cache}/devilutionGo/golangci-lint-$version"
bin="$cache/golangci-lint"

if [ ! -x "$bin" ]; then
    mkdir -p "$cache"
    GOBIN="$cache" go install "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$version"
fi

if [ -f .custom-gcl.yml ]; then
    "$bin" custom >/dev/null
    exec ./custom-gcl run "$@"
fi

exec "$bin" run "$@"
