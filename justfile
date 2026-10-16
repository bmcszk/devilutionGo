# devilutionGo — build automation
# vet/lint/test no-op until the scaffold bean lands (repo has no Go sources yet).
HAS_GO_SRC := `find . -name '*.go' -not -path './bin/*' | grep -q . && echo yes || echo no`

default:
    @just --list

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

# Validate .beans ledger (frontmatter, dep targets, skills paths, Libraries policy)
beans-lint:
    python3 scripts/beans-lint.py .

# Fast gate: ledger + vet + lint
check: beans-lint vet lint

# Full gate for git push (hook runs this)
pre-push: check test

# Full CI pipeline
ci: fmt vet lint test beans-lint
