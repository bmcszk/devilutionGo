---
# devilGo-lkj5
title: 'Gate parity: lint on commit, CI-equivalent suite on push, same checks in GHA (Support)'
status: completed
type: task
priority: normal
created_at: 2026-11-26T15:09:10Z
updated_at: 2026-11-26T16:20:02Z
parent: devilGo-p1t8
---

# Gate parity: lint on commit, CI-equivalent suite on push, same checks in GHA (Support)

## Goal
Hooks and GHA enforce the SAME checks, from one source: lint on commit (exists, kept), pre-push runs the full CI-equivalent gate suite, GHA mirrors it as named steps, the installer stops shipping paper gates, and the commit format gets its house-law check. All coding-rules caps already live in .golangci.yml (funlen 60/40, cyclop 10, revive function-length 100, file-length 1000, lll 120, nolintlint strict) and run via the single fail-closed oracle scripts/lint.sh — this lane moves no lint rules, it wires the same checks into every enforcement point.

## Research (triage task-gp-triage, 2026-11-26, parent spot-checked)
- THREE pre-push variants exist (W1): the repo copy scripts/git-hooks/pre-push (dirty-tree + foreign-worker guard + build/lint/test), the live .git/hooks/pre-push on this clone (gofmt/vet/lint/test/beans-lint, no build/windowed), and the installer's weak heredoc (|| echo skipped). The unified hook keeps the repo copy's foreign-worker guard (D7) and gains the missing gates.
- TestAC4InteractiveDemoSmoke does NOT skip asset-less (W2): no DG_ASSET_ARCHIVES read; boot.capture -> syncScene -> composeSceneAt returns "DG_ASSET_ARCHIVES not set" (window_scene.go:154-156). Asset-less windowed cmd tests are RED on master today, so the windowed-test gate needs the S0 enabling skip FIRST (in-bean, not a drive-by).
- House <=72 commit law is NOT what master practice follows (W3: last ten ledger commits 80-150 chars); the hook enforces <=100 RUNES as the hard stop, the 72 law stays the written rule — recorded as a deliberate gap.
- beans-lint.py:71 mandates ponytail + coding-rules in every todo body (verified).
- lint.sh does NOT run the .golangci.yml formatters (golangci run vs fmt) — gofmt -l is additive, not redundant.

## Decisions (D1-D9, triage-resolved)
- D1 single-source the LIST: CI keeps separate NAMED steps, each running the identical command string the hook runs; no CI-side gates.sh monolith.
- D2 commit format hook-side ONLY (GitHub squash titles unhooked; no GHA commit-format job): subject regex `^(feat|fix|hotfix|chore|docs|refactor|test|ci|build|style|perf)(\([a-z0-9._-]+\))?!?: [^A-Z].*[^.]$`, <=100 runes, exemptions `Merge `/`Revert `/`revert: `/`squash!`/`fixup!` + MERGE_HEAD; runs after the Pi-Session check so HERMES_ALLOW_DIRECT_COMMIT=1 bypasses both.
- D3 windowed in CI asset-less (scene skips on env; no windowed test calls ebiten.RunGame; no display needed) — packages exactly `./internal/present/scene ./cmd/devilutionGo`.
- D4 sqlc via `go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1` (ledger-claimed version; coder confirms with `sqlc version` per AC9), no third-party action; pin single-sourced in scripts/sqlc-pin.txt.
- D5 Pi-Session trailer stays hook-side.
- D6 pre-push: windowed build+vet always; windowed TESTS asset-less LAST, after S0; if measured >30s, fall back to build+vet and record the number.
- D7 the repo copy's foreign-worker guard (self-ancestry exclusion) MUST survive into the unified hook (the installer's pgrep variant was the old bug).
- D8 pre-push stops discarding test output (`>/dev/null` goes) — failures print the tool's own summary.
- D9 gofmt -l whole-tree in hook AND CI (lint.sh covers formatters: none — gofmt is additive).

## Scope (S0-S7)
- S0 cmd/devilutionGo/ac4_demo_gate_test.go: requireArchives-style skip at the top of TestAC4InteractiveDemoSmoke (reuse the live_view_test.go:92-101 helper shape). PRECONDITION for the windowed gates.
- S1 scripts/git-hooks/commit-msg: + conventional format gate per D2 (comments stripped, subject line, exemptions, rune-length, expected-shape message on fail).
- S2 scripts/git-hooks/pre-push: unified gate list per D1/D6-D9: gofmt -l .; go build ./...; go build -tags windowed ./...; go vet ./...; go vet -tags windowed ./...; scripts/lint.sh ./... --timeout=5m; scripts/check-no-raw-sql.sh; just sqlc-check; go test -count=1 ./... (visible summary on fail); go test -count=1 -tags windowed ./internal/present/scene ./cmd/devilutionGo (last); python3 scripts/beans-lint.py . Each prints `pre-push: <name>` first, `pre-push: <name> FAILED` + exit 1 on failure. Keeps dirty-tree, HERMES_SKIP_PREPUSH, foreign-worker guard verbatim.
- S3 scripts/install-hooks.sh: rewrite — set -euo pipefail, copy ALL THREE hooks from scripts/git-hooks/ into $(git rev-parse --git-path hooks), chmod +x, verify byte-identical with cmp, no inline heredoc.
- S4 justfile: install-hooks recipe -> @scripts/install-hooks.sh; pre-push recipe comment names the hook as source.
- S5 .github/workflows/ci.yml: go-version-file: go.mod in BOTH jobs; ci job = named steps mirroring S2 in order (gofmt, build, build (windowed), vet, vet (windowed), lint, no-raw-sql, sqlc-check with the go-install pin step, test, test (windowed), beans-lint); zero DG_ASSET_ARCHIVES references.
- S6 scripts/sqlc-pin.txt + justfile sqlc-check: .cache/sqlc install guard lint.sh-style (pins file + cmp), read the pin from the file in hook and CI alike.
- S7 PoW: seeded-failure proof per new gate (both tails verbatim).

## Acceptance criteria
- [x] AC1 commit-msg format gate: `git commit -m "Updated the thing"` REJECTED naming the accepted shape; `git commit -m "fix(persist): clamp slot to 0..98"` ACCEPTED; a Merge message EXEMPT; HERMES_ALLOW_DIRECT_COMMIT=1 accepts anything; Pi-Session check still fires on a conventional subject lacking the trailer
- [x] AC2 length rule: 101-rune subject rejected, 100-rune accepted; an em-dash subject measured in RUNES (seed one at the boundary, show the count the hook prints)
- [x] AC3 gate parity: every gate command in scripts/git-hooks/pre-push also appears verbatim as a run: in .github/workflows/ci.yml (two lists side by side in PoW)
- [x] AC4 seeded failures, hook side, each observed as `pre-push: <name> FAILED` + tool output: (a) gofmt bad-indent, (b) windowed-build syntax error in a tag-only file, (c) windowed vet Printf mismatch, (d) raw SQL in non-sqlc persist file, (e) sqlc-check after editing queries without regenerating, (f) beans-lint broken frontmatter, (g) a broken selhero assertion under the tag (after S0)
- [x] AC5 seeded failures, CI side: named steps red for (b), (d), (e) — step names visible in gh run view: gofmt / build / build (windowed) / vet / vet (windowed) / lint / no-raw-sql / sqlc-check / test / test (windowed) / beans-lint
- [x] AC6 asset-less windowed: `env -u DG_ASSET_ARCHIVES go test -count=1 -tags windowed ./internal/present/scene ./cmd/devilutionGo` rc0 after S0 with TestAC4InteractiveDemoSmoke SKIP (paste --- SKIP); wall time recorded (>30s => D6 fallback + record)
- [x] AC7 CI needs no assets (PARENT DECIDED after review B1, option b): both the hook and the test (windowed) CI step run the IDENTICAL env -u DG_ASSET_ARCHIVES go test -count=1 -tags windowed ./internal/present/scene ./cmd/devilutionGo string, and grep -c DG_ASSET_ARCHIVES= .github/workflows/ci.yml = 0 (the var is unset, never assigned - provably asset-less regardless of runner env)
- [x] AC8 drift killed: grep -n go-version ci.yml shows ONLY go-version-file: go.mod lines (both jobs)
- [x] AC9 sqlc pin single-sourced: scripts/sqlc-pin.txt holds the pin; local `sqlc version` output pasted proving the committed sqlc/* was generated by it (or regenerated + empty diff); .cache/sqlc removed -> re-installs, passes; proxy unreachable + no cache => FAILS, never "skipped"
- [x] AC10 installer fresh-clone simulation: clone -> just install-hooks prints three installed lines; cmp proves all three identical to sources; grep -c skipped on the installed pre-push = 0; repeat from a worktree checkout -> hooks land in the MAIN .git/hooks
- [x] AC11 foreign-worker guard survives: own-ancestry push SUCCEEDS; a synthetic foreign pi_rpc_client process blocks the push with `another pi worker is still running` (both pasted)
- [x] AC12 test output visible: one seeded failing untagged test; pre-push output contains that test's --- FAIL line
- [x] AC13 full gates green at the branch head (all tails pasted): gofmt empty; builds both tags rc0; vet both tags rc0; lint.sh 0 issues; untagged suite ok; AC6 rc0; beans-lint OK

## Libraries
stdlib bash + pinned sqlc build tool (go install @pin); no new Go module.

## Skills for coder
- ~/.agents/skills/coding-rules/SKILL.md (+ references/enforcement-gates.md — this wave IS a delegated rule pass: seeded canaries + fixed-verdict citation laws)
- ~/.agents/skills/go-dev/SKILL.md
- ~/.agents/skills/ponytail/SKILL.md — smallest landable diff (also beans-lint-mandated)
- ~/.agents/skills/shell-traps/SKILL.md — the hooks are bash; the installer's `|| echo skipped` is a textbook silent-failure trap
- ~/.agents/skills/git/SKILL.md — the commit-naming law (D2) and the CI-equivalent pre-push recipe being encoded
- ~/.hermes/skills/dev/devilutiongo-lane-ops/SKILL.md — repo hook contract + push-hook discipline + sqlc gate law

## Triage
Filed 2026-11-26 (parent audit .hermes/lane/gate-parity/audit.md after the user ordered: git hooks - lint on commit, test on push, same checks in GHA, all coding-rules rules and more). Deep triage DONE (task-gp-triage-brainer; parent spot-checked W1 three-variants, W2 AC4 no-skip, beans-lint ponytail law). Verdict PROMOTE-TODO with S0 as the in-bean enabling fix; corrections W1/W2 and D1-D9 recorded above. Full verdict: .hermes/lane/gate-parity/triage-result.md



## DECIDED (parent, 2026-11-26, after review round 1)

- B1: option (b) — both hook and CI share the env -u prefixed windowed-test string; AC7 reworded above (the reviewer's stronger-gate option). The coder's original tie-break (bare string + grep=0) is superseded; the weaker guarantee is rejected because a future runner/secret exporting the var would silently make the windowed gate asset-dependent.
- B2: the installer's cmp compares the installed hook against git show HEAD:scripts/git-hooks/<hook> (the committed source, not the worktree's possibly-dirty copy) — that catches cp failure AND dirty-source drift; the comment states exactly what is guarded. The mixed-checkout proof (main on master, worktree on the branch, install from worktree, installed == branch source and != master source) lands in the round-2 PoW. Symlinked hooks-dir note added as a comment (cp intentionally writes through).



## Proof of Work

- TDD-equivalent: every new gate carries a SEEDED-FAILURE proof with BOTH tails verbatim (7 hook-side seeds a-g + 3 CI-side named-step reds via the closed canary PR #88, runs 45850717199 / 45850912059 / 45851294462) - .hermes/lane/gate-parity/wave-result.md. Clean rehearsal: all 11 gates green in 42.1s.
- Review round 1 (task-gp-review-brainer): REQUEST-CHANGES, B1 (AC7 tie-break was a coder-side AC ruling; parent DECIDED option b - env -u prefixed shared string, grep 'DG_ASSET_ARCHIVES=' = 0) + B2 (installer cmp was tautological; parent DECIDED cmp against git show HEAD committed source + mixed-checkout proof + symlink note). Parent decisions recorded in the bean (## DECIDED, commit 94c9e2d).
- Repair round (task-gp-wave-coder resumed, cd20c42): both fixes landed; repair-result.md carries byte-identical string proof, parity re-check (11 gates), seeded windowed failure still failing + clean pass green, fresh-clone + mixed-checkout + dirty-source installer proofs.
- Review round 2 (task-gp-review-brainer, fix-diff only): **APPROVE** - both blockers resolved per the parent's DECIDED, both tails present, nothing outside the three gate files moved (.hermes/lane/gate-parity/review2-result.md).
- Parent gate re-run at cd20c42 (CI-shaped): gofmt empty; builds both tags rc0; vet both tags rc0; lint.sh 0 issues; check-no-raw-sql clean; sqlc-check clean (pinned v1.31.1); untagged suite 54 pkgs ok; env -u windowed pair ok (cmd 15.2s); beans-lint OK (237). Clean-clone oracle at 09b8e02 rc0 (50 internal pkgs + windowed scene) - pre-repair head; post-repair delta is bash/yaml only.
- Residual: the house <=72 commit-subject law remains written-but-not-enforced (hook enforces 100 runes; master practice 80-150) - deliberate, recorded in the bean; actions/@v4 tags not SHA-pinned and gofumpt/gci/golines not enforced, status.sh raw golangci call - recorded as out-of-scope follow-ups in the triage.
