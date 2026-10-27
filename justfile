default:
    @just --list

# Build the GUI binary
build-gui:
    go build -o bin/devilutionGo ./cmd/devilutionGo

# Build the headless binary (CI / parity testing)
build-headless:
    go build -tags=headless -o bin/devilutionGo-headless ./cmd/devilutionGo-headless

# Format code
fmt:
    gofmt -s -w .

# The only way to run lint (also what CI runs)
lint:
    scripts/lint.sh ./... --timeout=5m

# Vet
vet:
    go vet ./...

# Tests
test:
    go test -count=1 ./...

# Data audit: manifest/row-count parity for the committed gamedata tables
data-audit:
    go test -count=1 ./internal/diablo/gamedata/audit/...

# Recompute every sha256 row of the committed test/fixtures/manifest.txt against
# the local test/fixtures/generated/ tree (skips when no local artifacts;
# exit nonzero on mismatch or missing file).
fixture-verify:
    @python3 scripts/fixture_fixtures.py verify

# Copy the probe run (tools/probe/out/) into the local, untracked
# test/fixtures/generated/, refresh the committed text manifest at
# test/fixtures/manifest.txt, print the drift report and regenerate the
# docs/fixtures.md inventory.
fixture-refresh:
    @python3 scripts/fixture_fixtures.py refresh

# Regenerate full-table gamedata goldens from committed TSVs (zero-dependency):
# refuses on a dirty tree, runs the Go generator, prints a diffstat.
refresh-gamedata-goldens:
    #!/usr/bin/env bash
    set -euo pipefail
    if [ -n "$(git status --porcelain)" ]; then
        echo "refresh-gamedata-goldens: working tree is dirty; commit or stash first" >&2
        exit 1
    fi
    go run ./cmd/gamedatagoldens
    git --no-pager diff --stat -- internal/testdata/gamedata/goldens
    git status --porcelain internal/testdata/gamedata/goldens

# Validate .beans ledger (frontmatter, dep targets, skills paths, Libraries policy)
beans-lint:
    python3 scripts/beans-lint.py .

# Fast gate: ledger + vet + lint
check: beans-lint vet lint

# Full gate for git push (hook runs this)
pre-push: check test

# Full CI pipeline
ci: fmt vet lint test beans-lint

# Gate: ready for review (scripts/pr-ready.sh)
pr-ready:
    @scripts/pr-ready.sh

# Compact session status (scripts/status.sh)
status:
    @scripts/status.sh

# Install repo git hooks (pre-commit: pi-session trailer + lint)
install-hooks:
    @cp scripts/git-hooks/pre-commit scripts/git-hooks/commit-msg .git/hooks/
    @chmod +x .git/hooks/pre-commit .git/hooks/commit-msg
    @echo "installed .git/hooks/pre-commit .git/hooks/commit-msg"
