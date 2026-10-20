# devilutionGo — build automation
# vet/lint/test skip gracefully while the module has no Go sources (HAS_GO_SRC guard).
HAS_GO_SRC := `find . -name '*.go' -not -path './bin/*' | grep -q . && echo yes || echo no`

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

# Lint (golangci-lint)
lint:
    @if [ "{{HAS_GO_SRC}}" = "yes" ]; then golangci-lint run ./...; else echo "lint: no Go sources yet"; fi

# Vet
vet:
    @if [ "{{HAS_GO_SRC}}" = "yes" ]; then go vet ./...; else echo "vet: no Go sources yet"; fi

# Tests
test:
    @if [ "{{HAS_GO_SRC}}" = "yes" ]; then go test -count=1 ./...; else echo "test: no Go sources yet"; fi

# Recompute every manifest row's sha256 against test/fixtures/generated/
# (hash-only: no regeneration; exits nonzero on mismatch or missing file).
fixture-verify:
    @python3 scripts/fixture_fixtures.py verify

# Copy the probe run into test/fixtures/generated/, print the manifest drift report, regenerate the docs/fixtures.md inventory.
fixture-refresh:
    @python3 scripts/fixture_fixtures.py refresh

# Validate .beans ledger (frontmatter, dep targets, skills paths, Libraries policy)
beans-lint:
    python3 scripts/beans-lint.py .

# Fast gate: ledger + vet + lint
check: beans-lint vet lint

# Full gate for git push (hook runs this)
pre-push: check test

# Full CI pipeline
ci: fmt vet lint test beans-lint
