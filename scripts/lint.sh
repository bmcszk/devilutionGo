#!/usr/bin/env bash
# The only way to run lint: local, hooks and CI all call this script.
set -uo pipefail

dir=.cache/lint
bin="$dir/golangci-lint"

if [ ! -x "$bin" ] || [ .custom-gcl.yml -nt "$bin" ]; then
    mkdir -p "$dir"
    golangci-lint custom >/dev/null
fi

# A build that silently drops the plugin reports 0 issues on everything.
if ! grep -q '^whenthen:' <("$bin" linters); then
    echo "lint: whenthen missing from the build — refusing to report a result" >&2
    exit 1
fi

exec "$bin" run "$@"
