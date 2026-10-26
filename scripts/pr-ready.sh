#!/usr/bin/env bash
# Gate: only claim a PR ready for review when every check passes.
set -uo pipefail

fail=0
ok() { printf 'OK   %s\n' "$1"; }
bad() { printf 'FAIL %s\n' "$1"; fail=1; }

# 1. Working tree clean
dirty=$(git status --porcelain | wc -l | tr -d ' ')
[ "$dirty" = "0" ] && ok "working tree clean" || bad "working tree dirty ($dirty paths)"

# 2. Nothing unpushed
upstream=$(git rev-parse --abbrev-ref --symbolic-full-name '@{u}' 2>/dev/null || true)
if [ -n "$upstream" ]; then
	n=$(git rev-list --count "$upstream..HEAD" 2>/dev/null || echo 0)
	[ "$n" = "0" ] && ok "no unpushed commits" || bad "$n unpushed commit(s) — PR does not contain them"
else
	bad "no upstream tracking branch"
fi

# 3. Lint clean
if command -v golangci-lint >/dev/null 2>&1; then
	out=$(golangci-lint run ./... 2>&1)
	# golangci-lint prints summary to stderr, issues to stdout
	# exit code 1 = issues found
	if [ $? -eq 0 ]; then
		ok "lint clean"
	else
		summary=$(printf '%s\n' "$out" | grep -E '^[0-9]+ issues in [0-9]+ files$' | head -1)
		[ -n "$summary" ] && bad "lint: $summary" || bad "lint issues present"
	fi
else
	bad "golangci-lint not installed"
fi

# 4. Build
go build ./... >/dev/null 2>&1 && ok "build passes" || bad "build fails"

# 5. Tests
go test -count=1 ./... >/dev/null 2>&1 && ok "tests pass" || bad "tests fail"

# 6. PR file list covers every changed path vs the base branch
if command -v gh >/dev/null 2>&1; then
	base=$(gh pr view --json baseRefName --jq .baseRefName 2>/dev/null || true)
	if [ -n "$base" ] && git rev-parse --verify "origin/$base" >/dev/null 2>&1; then
		changed=$(git diff --name-only "origin/$base...HEAD" | sort)
		inpr=$(gh pr diff --name-only 2>/dev/null | sort)
		if [ -z "$inpr" ]; then
			# gh pr diff refuses >300-file PRs (HTTP 406), so page the files API instead.
			num=$(gh pr view --json number --jq .number 2>/dev/null || true)
			[ -n "$num" ] && inpr=$(gh api --paginate "repos/{owner}/{repo}/pulls/$num/files" \
				--jq '.[].filename' 2>/dev/null | sort)
		fi
		if [ -z "$inpr" ]; then
			bad "cannot read PR file list"
		else
			missing=$(comm -23 <(printf '%s\n' "$changed") <(printf '%s\n' "$inpr") | head -5)
			if [ -z "$missing" ]; then
				ok "PR contains every changed file"
			else
				bad "PR missing changed path(s): $(printf '%s' "$missing" | tr '\n' ' ')"
			fi
		fi
	else
		bad "cannot resolve PR base branch"
	fi
else
	bad "gh not installed"
fi

echo "---"
[ "$fail" = "0" ] && echo "READY FOR REVIEW" || echo "NOT READY"
exit "$fail"
