#!/usr/bin/env bash
# Gate: the generated sqlc access is current and no data query is hand-written.
#
# The sqlc version is single-sourced in scripts/sqlc-pin.txt (D4): the hook and
# CI both reach it through this script, so the pin cannot drift. Fail-closed
# like scripts/lint.sh — a missing binary installs from the pin, and an install
# that cannot happen FAILS the gate, it never prints "skipped".
set -euo pipefail

root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
cd "$root"

pin=$(tr -d '[:space:]' < scripts/sqlc-pin.txt)
[ -n "$pin" ] || { echo "sqlc-check: scripts/sqlc-pin.txt is empty" >&2; exit 1; }

dir=.cache/sqlc
mkdir -p "$dir"
printf 'sqlc=%s\n' "$pin" > "$dir/pin.want"
if [ ! -x "$dir/sqlc" ] || ! cmp -s "$dir/pin.want" "$dir/pin.txt" 2>/dev/null; then
	echo "sqlc-check: installing pinned sqlc $pin" >&2
	if ! GOBIN="$root/$dir" go install "github.com/sqlc-dev/sqlc/cmd/sqlc@$pin"; then
		echo "sqlc-check: pinned sqlc $pin install failed (no cache, no module proxy?)" >&2
		exit 1
	fi
	[ -x "$dir/sqlc" ] || { echo "sqlc-check: sqlc binary missing after install" >&2; exit 1; }
	mv "$dir/pin.want" "$dir/pin.txt"
else
	rm -f "$dir/pin.want"
fi

"$root/$dir/sqlc" generate
git diff --exit-code -- internal/persist/sqlc
scripts/check-no-raw-sql.sh
