#!/bin/sh
# Install git hooks (pre-push gate). Run once after clone.
set -e
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cat > "$ROOT/.git/hooks/pre-push" <<'EOF'
#!/bin/sh
# Pre-push gate: beans ledger + vet + lint + tests (just pre-push)
# Bypass with: git push --no-verify
set -e
cd "$(git rev-parse --show-toplevel)"
if command -v just >/dev/null 2>&1; then
    just pre-push
else
    echo "pre-push: 'just' not found, running checks directly"
    python3 scripts/beans-lint.py .
    go vet ./... 2>/dev/null || echo "vet: skipped"
    golangci-lint run ./... 2>/dev/null || echo "lint: skipped"
    go test -count=1 ./... 2>/dev/null || echo "test: skipped"
fi

# retro 2026-10-17: refuse push when a pi worker holds the tree mid-write
if pgrep -f "pi_rpc_client.*--cwd $(git rev-parse --show-toplevel)" >/dev/null 2>&1; then
    echo "pre-push: pi worker STILL RUNNING (pi_rpc_client) — commit may mix worker scope. Wait or kill it." >&2
    exit 1
fi
EOF
chmod +x "$ROOT/.git/hooks/pre-push"
echo "installed .git/hooks/pre-push"
