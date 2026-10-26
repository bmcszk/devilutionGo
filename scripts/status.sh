#!/usr/bin/env bash
# Compact session status: six lines instead of six full command dumps.
set -uo pipefail

printf 'tree:     %s\n' "$([ -z "$(git status --porcelain)" ] && echo clean || echo "$(git status --porcelain | wc -l | tr -d ' ') dirty")"
printf 'head:     %s\n' "$(git log --oneline -1)"
up=$(git rev-parse --abbrev-ref --symbolic-full-name '@{u}' 2>/dev/null || true)
printf 'unpushed: %s\n' "$([ -n "$up" ] && git rev-list --count "$up..HEAD" || echo '?')"
lint=$(golangci-lint run ./... 2>&1 | sed -n 's/^golangci-lint: //p' | head -1)
printf 'lint:     %s\n' "${lint:-clean}"
printf 'build:    %s\n' "$(go build ./... >/dev/null 2>&1 && echo pass || echo FAIL)"
printf 'tests:    %s\n' "$(go test -count=1 ./... >/dev/null 2>&1 && echo pass || echo FAIL)"
if command -v gh >/dev/null 2>&1; then
	printf 'pr:       %s\n' "$(gh pr view --json number,state,mergeable,files --jq '"#\(.number) \(.state) \(.mergeable) files=\(.files|length)"' 2>/dev/null || echo none)"
fi
