#!/usr/bin/env bash
# The only way to run lint: local, hooks and CI all call this script.
#
# Oracle law: the gate must be fail-closed and tool-faithful.
# 1. Vanilla golangci-lint at the pin, no plugins. The former custom build
#    (whenthen linked in) silently dropped revive `unhandled-error` findings
#    in multi-file packages that the vanilla binary at the SAME pin reports
#    (same revive v1.17.1 in both; the linked plugin perturbs the goanalysis
#    driver). A hook running the blind build said 0 while CI flagged - so the
#    custom build is no longer trusted as an oracle.
# 2. whenthen runs as its own `go vet -vettool` step (the repo ships
#    cmd/whenthen for exactly that) - no plugin linking needed.
# 3. A canary with a seeded revive violation must FAIL lint, else a 0-issues
#    verdict from the oracle is refused. This turns silent rule drops into
#    loud gate failures.
#
# Requirements: Go toolchain; the first run downloads the pinned modules.
set -uo pipefail

golangci_pin=v2.15.0
whenthen_pin=v0.1.2
dir=.cache/lint
mkdir -p "$dir"
golangci="$dir/golangci-lint"
vettool="$PWD/$dir/whenthen"
root=$PWD

printf 'golangci=%s\nwhenthen=%s\n' "$golangci_pin" "$whenthen_pin" > "$dir/pins.want"
if ! cmp -s "$dir/pins.want" "$dir/pins.txt" 2>/dev/null; then
    echo "lint: installing pinned oracles: golangci-lint $golangci_pin (vanilla), whenthen $whenthen_pin" >&2
    GOBIN="$root/$dir" go install "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$golangci_pin" || exit 1
    GOBIN="$root/$dir" go install "github.com/bmcszk/whenthen/cmd/whenthen@$whenthen_pin" || exit 1
    # a botched install would leave a stale or missing vettool: refuse
    [ -x "$vettool" ] || { echo "lint: whenthen vettool missing after install" >&2; exit 1; }
    mv "$dir/pins.want" "$dir/pins.txt"
fi

# Canary: the seeded violation must be reported. The canary carries its own
# minimal config so the repo rule set cannot mask the probe.
canary="$dir/canary"
mkdir -p "$canary"
cat > "$canary/go.mod" <<'EOF'
module lintcanary

go 1.27.3
EOF
cat > "$canary/.golangci.yml" <<'EOF'
version: "2"

linters:
  default: none
  enable:
    - revive

  settings:
    revive:
      rules:
        - name: unhandled-error
EOF
cat > "$canary/main.go" <<'EOF'
package main

import (
	"crypto/sha256"
	"fmt"
)

func main() {
	h := sha256.New()
	b := []byte("canary")
	h.Write(b)
	fmt.Println(h.Sum(nil))
}
EOF
if ( cd "$canary" && "$root/$golangci" run ./... >/dev/null 2>&1 ); then
    echo "lint: canary violation NOT reported - oracle dropped a seeded issue; refusing to verdict" >&2
    exit 1
fi

# The actual gate: vanilla golangci with the repo rules, then the GWT gate on
# the same package set. Package args only for go vet (golangci-style flags
# like --timeout do not apply there).
if ! "$root/$golangci" run "$@"; then
    exit 1
fi

pkgs=()
for arg in "$@"; do
    case "$arg" in -*) ;; *) pkgs+=("$arg") ;; esac
done
if [ "${#pkgs[@]}" -eq 0 ]; then
    pkgs=("./...")
fi

if ! go vet -vettool="$vettool" "${pkgs[@]}"; then
    exit 1
fi

exit 0
