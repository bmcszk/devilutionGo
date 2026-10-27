#!/usr/bin/env bash
# Run the repo rule set with the whenthen plugin removed, using the pinned stock
# binary. The stock binary cannot load a module plugin, so this is the only way
# to get a second opinion on the remaining rules. Diagnostic, never a gate.
set -uo pipefail

version=v2.15.0
cache="${XDG_CACHE_HOME:-$HOME/.cache}/devilutionGo/golangci-lint-$version"
bin="$cache/golangci-lint"

if [ ! -x "$bin" ]; then
    mkdir -p "$cache"
    GOBIN="$cache" go install "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$version"
fi

cfg=$(mktemp "${TMPDIR:-/tmp}/golangci-oracle.XXXXXX.yml")
trap 'rm -f "$cfg"' EXIT

python3 - "$cfg" <<'PY'
import sys

lines = open(".golangci.yml", encoding="utf-8").read().splitlines(keepends=True)
out = []
skip_child_indent = None

for line in lines:
    stripped = line.strip()

    if stripped == "- whenthen":
        continue

    if skip_child_indent is not None:
        if not stripped or line.startswith(skip_child_indent):
            continue
        skip_child_indent = None

    if stripped == "whenthen:":
        skip_child_indent = line[: len(line) - len(line.lstrip())]
        continue

    out.append(line)

text = "".join(out)
if "whenthen" in text:
    sys.exit("oracle: could not strip whenthen from the config")

open(sys.argv[1], "w", encoding="utf-8").write(text)
PY

exec "$bin" run --config "$cfg" "$@"
