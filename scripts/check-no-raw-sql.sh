#!/usr/bin/env bash
# Gate: the persistence store's data access must be sqlc-generated. Fails when a
# non-generated file issues a query with a SQL string literal of its own.
#
# The only exemption is sqlc's own output directory. Everything else, including
# the PRAGMA user_version read and the embedded schema DDL, goes through a
# named constant or the generated access so this gate stays strict.
set -euo pipefail

cd "$(dirname "$0")/.."

pattern='(Query|QueryRow|Exec|Prepare)Context?\([^)]*("|`)[[:space:]]*(SELECT|INSERT|UPDATE|DELETE|CREATE|ALTER|DROP|PRAGMA|REPLACE)'

hits="$(grep -rnE "$pattern" --include='*.go' internal/persist \
    | grep -v '^internal/persist/sqlc/' || true)"

if [[ -n "$hits" ]]; then
    echo "raw SQL found outside sqlc-generated code:" >&2
    echo "$hits" >&2
    exit 1
fi

echo "no raw SQL outside sqlc-generated code"
