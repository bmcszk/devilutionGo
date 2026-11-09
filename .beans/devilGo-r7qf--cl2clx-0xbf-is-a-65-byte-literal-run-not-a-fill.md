---
# devilGo-r7qf
title: CL2/CLX 0xBF is a 65-byte literal run, not a fill
status: scrapped
type: bug
priority: critical
created_at: 2026-10-28T16:35:09Z
updated_at: 2026-11-09T20:04:10Z
parent: devilGo-5wsb
---

## Goal

Fix the CLX/CL2 opcode `0xBF` misclassification. It is a literal run of 65 bytes in the C++,
not a fill run.

## Evidence (verified in the parent, both by reading the C++ and by executing a probe)

- Go `internal/assets/cl2.go:179` dispatches with `case control <= 0xBF:` to `decodeFillRun`,
  whose `runLen = 0xBF - control` is **0** for `0xBF`, and which then consumes a phantom value
  byte.
- C++ `Source/utils/clx_decode.hpp:25` `ClxFillMax = 0xBE`; `:31-34`
  `GetClxOpaqueFillWidth` asserts width `1..63`; `:16-20` `GetClxOpaquePixelsWidth(0xBF) == 65`.
- C++ *encoder* emits it: `Source/utils/clx_encode.hpp:44-49` `AppendClxPixelsRun` pushes
  `0xBF` followed by 65 literal bytes for every opaque non-repeating run `>= 0x41`.
- Executed probe (frame header + `0xBF` + 65 ascending bytes, width 65 height 1) fails on the
  current tree: `BUG REPRODUCED pixel 0 = 0, want 1`.
- The committed corpus does not catch it: `0xBF` occurs in 5 real frames across
  `data_hinticons`, `data_resistance`, `ui_art_dvl_but_sml`, but never at a command position, so
  the 20 goldens pass. The test helper `cl2L` cannot produce `0xBF` by construction.

## Acceptance criteria

- [ ] `decodeOneCommand` routes `0xBF` to the literal path, and fill is `0x80 <= control <= 0xBE`
      exactly as `clx_decode.hpp:23-35`.
- [ ] A test covering the `0xBF` boundary: a 65-byte literal run decodes to those 65 bytes, and a
      0xBE fill still decodes to its 1-byte fill (`0xBF - 0xBE == 1`).
- [ ] A test asserting the literal width range is `1..65` and the fill width range `1..63`.
- [ ] Existing behaviour preserved: all committed CLX digests still match.
- [ ] RED first: capture the failing probe output in the work log before the fix.

Refs: PR #3, devilGo-n89r

## Skills for coder
- ~/.agents/skills/ponytail/SKILL.md

- /home/blaze/.agents/skills/go-dev/SKILL.md — decoder fixes, byte-exact RLE semantics
- /home/blaze/.agents/skills/go-tests/SKILL.md — golden-fixture test style
- /home/blaze/.agents/skills/coding-rules/SKILL.md + ponytail — read fully before acting

## Acceptance criteria

- [ ] 0xBF dispatches to literal-run decode (65-byte literal per C++ cl2.cpp)
- [ ] Golden CLX fixtures re-verified: decoder output byte-identical to committed goldens
- [ ] New fixture covering a 0xBF run (from spawn.mpq asset) added to internal/testdata
- [ ] clx_bench shows no regression (>10% slowdown fails gate)

## Libraries

Go stdlib only — no new module dependencies to bind (repo policy; existing go.mod surface unchanged).
## Triage

- Brainer beans-audit of the ledger ran 2026-10-28 (devilGo-5wsb-beansaudit session; PR #3 verdict log). Per-bean brainer pass owed before implementation dispatch; wave placement, blockers and skills list set during the 2026-10-29 reorg (d083780).

## Reasons for Scrapping (2026-11-09)

Duplicate ledger entry of the same bug class. The surviving single source of truth is [CL2/CLX 0xBF is a 65-byte literal run, not a fill](.beans/devilGo-0jjm--cl2clx-0xbf-is-a-65-byte-literal-run-not-a-fill.md), status completed, with its own Proof of Work (2026-10-28, PR #3, commit 78d8772).

Parent verification against master ad1eea4 today (all green):
- internal/assets/cl2.go:279 `case control <= clxFillMax` with `clxFillMax = 0xBE` (:11) routes 0xBF to the literal path; decodeFillRun runLen 0xBF-control (1..63).
- cl2_test.go: boundary tests cl2ExpectLiteralRunAtBF (65-byte literal), cl2ExpectFillRunAtBE (1-byte fill). clx_test.go: TestCLXLiteralRunAt0xBF, TestCLXFillRunAt0xBE, TestCLXFillRunWidths + TestCLXLiteralRunWidths.
- Bench: BenchmarkCL2Decode ran clean (~5.1us/op synthetic frame; no regression in the 38-package green suite).
- This bean's own red-probe evidence (BUG REPRODUCED pixel 0 = 0, want 1) predates the fix and is preserved in the body above.
