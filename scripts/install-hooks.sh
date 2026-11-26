#!/usr/bin/env bash
# Install the repo's git hooks. `$(git rev-parse --git-path hooks)` resolves to
# the shared hooks dir, so a cross-checkout install (main on master, worktree on
# a branch) intentionally installs THAT checkout's COMMITTED hooks into the
# shared .git/hooks: the last installer run wins, by design.
#
# Guarded: each installed hook is byte-identical to its COMMITTED source
# (`git show HEAD:...`). A failed write, and a dirty worktree copy that would
# otherwise lie about the committed hook, both cannot reach the shared hooks.
#
# If the hooks dir is a symlink (dotfiles/chezmoi setups), the write follows it
# to the real file.
set -euo pipefail

root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
cd "$root"
hooks=$(git rev-parse --git-path hooks)
mkdir -p "$hooks"

for hook in commit-msg pre-commit pre-push; do
	src="scripts/git-hooks/$hook"
	dst="$hooks/$hook"
	if ! git diff --quiet HEAD -- "$src"; then
		echo "install-hooks: $src is dirty; installing the committed hook instead" >&2
	fi
	git show "HEAD:$src" > "$dst"
	chmod +x "$dst"
	if ! cmp -s <(git show "HEAD:$src") "$dst"; then
		echo "install-hooks: $dst differs from HEAD:$src" >&2
		exit 1
	fi
	echo "installed $dst"
done
