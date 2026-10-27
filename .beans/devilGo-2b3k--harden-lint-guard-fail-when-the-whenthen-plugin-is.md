---
# devilGo-2b3k
title: 'Harden lint guard: fail when the whenthen plugin is absent, add stock-binary oracle, single just recipe in CI'
status: completed
type: task
tags:
    - tooling
created_at: 2026-10-27T13:04:56Z
updated_at: 2026-10-27T13:04:56Z
---

## Goal

Make a silently-degraded custom golangci-lint build impossible, and give a second opinion that does not need the module plugin.

A custom build that omits the whenthen plugin reports 0 issues on everything — indistinguishable from a clean tree.

## Changes

1. `scripts/lint.sh` — after `"$bin" custom`, assert `whenthen:` is in `./custom-gcl linters`; exit 1 with a diagnostic when it is not.
2. `scripts/lint-oracle.sh` (new, 755) — run the repo config with the whenthen entries stripped, using the pinned stock v2.15.0 binary. Diagnostic, never a gate.
3. `justfile` — `lint` carries `./... --timeout=5m`; new `lint-oracle` recipe.
4. `.github/workflows/ci.yml` — lint step runs `just lint` via `extractions/setup-just@v4`, so CI and humans run the same recipe.

## Acceptance criteria

- [x] `bash -n` clean on both scripts
- [x] `just lint` exits 0 and actually loads whenthen
- [x] `just lint-oracle` exits 0 with the plugin stripped
- [x] guard discriminates: `grep -c '^whenthen:'` = 1, `grep -c '^nope:'` rc=1
- [x] no `.golangci.yml` change, no `//nolint`, no skips

## Deviation from the literal spec

The supplied snippet `if ! ./custom-gcl linters | grep -q '^whenthen:'; then` false-fires under `set -uo pipefail`: `grep -q` exits early, the producer takes SIGPIPE, `pipefail` surfaces 141, and `!` turns that into "missing". Measured:

```
./custom-gcl linters | grep -q '^whenthen:'; echo rc=$?   # rc=141
if ! ./custom-gcl linters | grep -q '^whenthen:'; then echo REPORTED-MISSING  # fires on a healthy build
```

Kept `grep -q` semantics via process substitution (`grep -q '^whenthen:' <(./custom-gcl linters)`), which is not a pipeline, so only grep's status is tested.

## Proof of Work

```
$ bash -n scripts/lint.sh && bash -n scripts/lint-oracle.sh && echo SYNTAX_OK
SYNTAX_OK

$ just lint 2>&1 | tail -2
scripts/lint.sh ./... --timeout=5m
0 issues.

$ just lint-oracle 2>&1 | tail -2
scripts/lint-oracle.sh ./... --timeout=5m
0 issues.

$ ./custom-gcl linters | grep -c '^whenthen:'
1

$ ./custom-gcl linters | grep -c '^nope:' ; echo "rc=$?"
0
rc=1

$ grep -n -A4 "name: lint" .github/workflows/ci.yml
18:      - name: lint
19-        run: just lint
20-      - name: test
21-        run: go test -count=1 ./...
22-      - name: beans-lint

$ git status --short
 M .github/workflows/ci.yml
 M justfile
 M scripts/lint.sh
?? scripts/lint-oracle.sh
```

Pre-fix guard run on the literal snippet (false positive, disproving it):

```
scripts/lint.sh ./... --timeout=5m
lint: whenthen missing from the custom build — the rule set did not load
error: recipe `lint` failed on line 18 with exit code 1
```

Residual: `scripts/lint-oracle.sh` strips the plugin with a small YAML-by-indent
Python pass, not a YAML parser; it aborts if the token survives, so a config
reshuffle fails loudly instead of silently linting the wrong rule set.

Refs: devilGo-yuqk (whenthen gate adoption)
