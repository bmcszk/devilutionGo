#!/usr/bin/env bash
# The only way to run lint: local, hooks and CI all call this script.
# Needs no preinstalled golangci-lint: the pin lives in .custom-gcl.yml.
set -uo pipefail

dir=.cache/lint
builder="$dir/builder"
bin="$dir/golangci-lint"
version=$(sed -n 's/^version:[[:space:]]*//p' .custom-gcl.yml | head -1)

if [ -z "$version" ]; then
    echo "lint: no version: line in .custom-gcl.yml" >&2
    exit 1
fi

if [ ! -x "$builder" ] || [ "$(cat "$dir/builder-version" 2>/dev/null)" != "$version" ]; then
    mkdir -p "$dir"
    GOBIN="$PWD/$dir" go install "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$version"
    mv "$dir/golangci-lint" "$builder"
    printf '%s\n' "$version" > "$dir/builder-version"
fi

if [ ! -x "$bin" ] || [ .custom-gcl.yml -nt "$bin" ] || [ "$builder" -nt "$bin" ]; then
    "$builder" custom >/dev/null
fi

# A build that silently drops the plugin reports 0 issues on everything.
if ! grep -q '^whenthen:' <("$bin" linters); then
    echo "lint: whenthen missing from the build — refusing to report a result" >&2
    exit 1
fi

exec "$bin" run "$@"
