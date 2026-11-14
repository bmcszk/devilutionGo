---
# devilGo-od1d
title: wrap-layout oracle goldens from the offline C++ probe (text F4)
status: completed
type: task
priority: normal
created_at: 2026-11-11T17:30:00Z
updated_at: 2026-11-13T19:06:54Z
---

---
# devilGo-od1d
title: wrap-layout oracle goldens from the offline C++ probe (text F4)
status: draft
type: task
priority: normal
created_at: 2026-11-11T17:30:00Z
updated_at: 2026-11-11T17:30:00Z
---

# Wrap-layout oracle goldens (C++ probe run once)

## Symptom

text_test.go's wrap expectations (feat/text, devilGo-v1hy) are hand-derived
from committed glyph widths + the ported algorithm — reviewer F4: the bean's
wrap AC ("wrap widths identical for 20 sample strings") is not proven against
upstream OUTPUT, and the worker self-corrected the table twice before GREEN
(text.s2.log), the known failure mode of hand-derived evidence.

## Fix

Run the offline C++ producer chain ONCE (pattern: internal/testdata/render/
producer/oracle.cpp + the ADR-019 tools carve-out): a small gtest/probe that
feeds the 20 golden-source strings to upstream text_render.cpp DrawString/
WordWrap at the pinned commit and dumps per-line (startX, lineWidth) pairs
as TSV; commit the TSV under internal/testdata/assets/fonts/; a Go test
strict-loads it and asserts internal/render wrapLines matches line-for-line.

## Acceptance criteria

- [x] C++ probe committed (source, like oracle.cpp) + 20-string TSV golden
- [x] Go strict-loader test asserts wrapLines output == TSV per line
- [x] beans-lint clean; no binary artifacts (TSV text only)

## Architecture guardrails (tech-debt triage 2026-11-12, binding for this bean's lane)

Source: the ok51 tech-debt triage (task-techdeptTriage-brainer) — these are the
lessons from the mistakes already made, applied BEFORE this bean starts. The
contract for this bean restates whichever apply; violating one is a review
REQUEST-CHANGES, not a debate.

1. Program-surface law (ADR-020/021 + one-tool law): no new binaries. Fixture
   instruments are subcommands of tools/goldens; user-facing surface stays
   exactly cmd/devilutionGo + cmd/devilutionGo-migrate. No `func main` outside
   the allowlist.
2. Tooling problems: fix immediately, no hacks.

## Progress log — oracle goldens slice

- Corpus: 31 strings (20 LineWidth-parity + 11 wrap cases) extracted verbatim
  from internal/render/text_test.go TestWrapLayout + TestFontLineWidthParity;
  golden TSV holds 41 rows (one per wrapped line).
- Probe route: verbatim copies (route B) of WordWrapString
  (text_render.cpp:741-821), GetLineWidth (cpp:631-659),
  IsBreakableWhitespace (cpp:1012-1015), IsBreakAllowed (cpp:286-289), with
  DecodeFirstUtf8CodePoint standing in for utils/utf8.cpp — #include of the
  whole TU rejected (unbounded font-loading/Surface dependency surface).
  Glyph widths decoded at runtime from the committed 12-05.clx.hex golden,
  so no widths are hand-written. Every copied block carries a verbatim
  ca91d778f path:line-range comment; all line numbers re-verified at the pin.
- Ran ONCE offline with g++ -std=c++20 -O2; upstream output matched every
  existing hand-derived expectation (no corrections needed).
- Strict loader test: internal/render/text_oracle_test.go TestWrapLayoutOracle
  (t.Fatalf on missing/short/malformed TSV); asserts line count, per-line
  start-X consistency, last cursor X and last line width against the golden.
  Existing hand-derived pins kept untouched.
- Non-vacuity: flipped golden cell (stringIdx 23, lineWidth 9 -> 99) ->
  exact red `text_oracle_test.go:190: line 0 startX 99 inconsistent with
  width 8`; restored -> green.
- Gates: go build OK; gofmt -l empty; go vet OK; ./scripts/lint.sh ./... ->
  0 issues; beans-lint OK (193 beans, after repointing one dead
  go-dev-notes skill ref in devilGo-ocl1 to ~/.agents/skills/go-dev/SKILL.md);
  go test ./... -count=1 -> 50 pkgs ok, 0 FAIL.
- Ticked the two oracle ACs (probe+TSV golden, Go strict-loader test,
  beans-lint clean/no binaries); integration-hash AC stays open.

## Skills for coder

- `~/.agents/skills/coding-rules/SKILL.md` — read fully before editing
- `~/.agents/skills/go-dev/SKILL.md` — Go rules + gates
- `~/.agents/skills/go-tests/SKILL.md` — strict-loader test design
- /home/blaze/.agents/skills/ebiten/rendering/SKILL.md — glyph/sheet fixtures

- `~/.agents/skills/ponytail/SKILL.md` — smallest change that fully solves the task

## Proof of Work — closed from merged PR #70

- PR #70 (feat/text-oracle @ 8a9f71f) squash-merged by user as 9a1597e161 (2026-11-13T18:57:47Z, mergedBy bmcszk).
- Delivered: internal/testdata/render/producer/wraporacle.cpp (verbatim-copy probe at pin ca91d778f, route B), wrap_golden.tsv (41 rows / 31 strings), internal/render/text_oracle_test.go TestWrapLayoutOracle strict loader; existing pin tests untouched.
- Gates at head (parent re-run + postmerge70 drill at 9a1597e161): lint 0 issues, gofmt/vet clean, beans-lint OK 194 beans, go test 50 pkgs 0 FAIL; CI ci+bench SUCCESS.
- Review: brainer APPROVE (static) — provenance, corpus fidelity (all 41 rows recomputed), strict loader, mutation proof, laws, PR body all HOLD. F1/F2 record-only (parent operator re-runs green, od1d.review.md).
- Non-vacuity: TSV cell flip -> exact cell-level red -> restore -> green (worker + parent OC4).
- Residual: F1 provenance-comment gap in wraporacle.cpp accepted record-only (reviewer: goldens stand); integration-hash AC of v1hy stays open elsewhere.
