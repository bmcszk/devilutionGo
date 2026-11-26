#!/usr/bin/env bash
# Install the repo's git hooks. `$(git rev-parse --git-path hooks)` resolves to
# the shared hooks dir, so a worktree checkout installs into the MAIN .git/hooks
# (git's hooks are common to the clone, not per-worktree).
#
# Copies scripts/git-hooks/* verbatim and cmp-verifies each one: a hook that
# differs from its source is drift, and a missing gate must not be silent.
set -euo pipefail

root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
cd "$root"
hooks=$(git rev-parse --git-path hooks)
mkdir -p "$hooks"

for hook in commit-msg pre-commit pre-push; do
	src="scripts/git-hooks/$hook"
	dst="$hooks/$hook"
	cp "$src" "$dst"
	chmod +x "$dst"
	if ! cmp -s "$src" "$dst"; then
		echo "install-hooks: $dst differs from $src" >&2
		exit 1
	fi
	echo "installed $dst"
done
