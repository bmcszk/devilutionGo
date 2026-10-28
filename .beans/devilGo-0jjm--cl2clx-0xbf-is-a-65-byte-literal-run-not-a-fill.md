---
# devilGo-0jjm
title: CL2/CLX 0xBF is a 65-byte literal run, not a fill
status: completed
type: bug
priority: critical
created_at: 2026-10-28T16:35:33Z
updated_at: 2026-10-28T16:54:57Z
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

- [x] `decodeOneCommand` routes `0xBF` to the literal path, and fill is `0x80 <= control <= 0xBE`
      exactly as `clx_decode.hpp:23-35`.
- [x] A test covering the `0xBF` boundary: a 65-byte literal run decodes to those 65 bytes, and a
      0xBE fill still decodes to its 1-byte fill (`0xBF - 0xBE == 1`).
- [x] A test asserting the literal width range is `1..65` and the fill width range `1..63`.
- [x] Existing behaviour preserved: all committed CLX digests still match.
- [x] RED first: capture the failing probe output in the work log before the fix.

Refs: PR #3, devilGo-n89r

## Libraries
- none (stdlib only; test uses encoding/binary + testing)

## Skills for coder
- `~/.agents/skills/go-dev/SKILL.md` — read fully BEFORE acting
- `~/.agents/skills/coding-rules/SKILL.md` — read fully BEFORE acting
- `~/.agents/skills/coding-rules-notes/SKILL.md` — user amendments, read with coding-rules
- `~/.agents/skills/ponytail/SKILL.md` — read fully BEFORE acting

## Triage
Delegated read-only brainer audit (pi session `devilGo-5wsb-beansaudit`, report `/home/blaze/.hermes/cache/scratch/pr3-beans-audit.md`); every cited anchor re-verified in the parent against the pinned devilutionX checkout (`ca91d778f`) before filing.


## Proof of Work (2026-10-28, PR #3)

- RED captured (worker): `cl2_test.go:28 pixels = [0 0 ... 0], want [1 2 ... 65]`,
  `cl2_test.go:124` literal_width_65 (fill path gave runLen 0 + phantom value byte).
- Fix: `clxFillMax = 0xBE`; fill `0x80..0xBE`, literal `0xBF`/`>=0xC0` via `256-control`.
- Parent re-verification: own probe frame (`0xBF`+65 ascending bytes) now PASSES;
  20/20 CLX goldens unchanged; gofmt/build/vet clean; 26 pkgs ok; `scripts/lint.sh` `0 issues.`.
- Commit `559c2c1` (trailer `Pi-Session: devilGo-5wsb-coder`).
