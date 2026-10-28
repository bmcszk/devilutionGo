---
# devilGo-260w
title: 'Brainer-check nits: stale CL2 spec comment, CL2 sheet test gap, orphaned comment, limitation record'
status: completed
type: task
priority: normal
created_at: 2026-10-28T17:22:52Z
updated_at: 2026-10-28T17:26:20Z
parent: devilGo-5wsb
---

## Goal

Close the four verified nits from the brainer check of the audit-fix round
(`/home/blaze/.hermes/cache/scratch/dg-auditfix-check.md`, verdict SHIP-WITH-NITS).
All are comment/test-hygiene items; no behaviour change except adding a missing test.

## Findings (parent-verified in this tree)

1. **Stale command spec, `internal/assets/cl2.go:177-180`** — the overview comment block still
   documents `0x80..0xBF: fill run: (0xBF - control) copies…`, contradicting the fixed dispatch
   (`clxFillMax = 0xBE` at `:11-12`, `decodeFillRun` at `:204`). Update the block to the C++
   truth (`clx_decode.hpp:23-35`): `0x80..0xBE` fill (1..63), `0xBF` literal 65,
   `0xC0..0xFF` literal (1..64).
2. **`DecodeCL2` sheet branch untested** — `decodeCL2Sheet` (cl2.go:31-51) is wired into
   `DecodeCL2` but `cl2_test.go` has no group-offset test; only the CLX test exercises the
   shared `groupRanges`. Add `cl2ExpectSheetDecodes` mirroring `clxExpectSheetDecodes`
   (clx_test.go:141-160): a 2-group CL2 buffer with a group directory, asserting group sizes,
   per-group frame pixels, and that the pre-fix code REJECTED it (cite the old error path in a
   comment). Also assert the single-group path is untouched (existing digests/tests).
3. **Misplaced comment, `internal/assets/trn_test.go:307`** — `// trnFixtureNames lists the
   committed TRN fixtures by file name.` sits directly above
   `TestMonsterTypeFromMonsterIDMapsCouncilorFamily`. Move it to its actual declaration (or
   delete if the name is self-documenting — state which).
4. **Record the CL2 sheet limitation in code** — `DecodeCL2(data, width, height)` applies ONE
   width/height to every frame of every group, while C++ `Cl2ToClx` takes per-frame widths
   (`cl2_to_clx.cpp:66` `PointerOrValue`) and derives heights by line counting. Add a short
   doc comment on `DecodeCL2` (1-2 lines, what/why): multi-group CL2 with per-frame dims
   cannot be decoded faithfully through this API yet; per-frame dims tracked for the
   renderer port. No API change in this round.

## Acceptance criteria

- [x] cl2.go command-spec comment block matches `clx_decode.hpp:23-35` exactly.
- [x] A CL2 multi-group test exists and FAILS when the sheet branch is reverted; existing CL2
      tests and all 20 CLX golden digests unchanged.
- [x] No orphaned comment above `TestMonsterTypeFromMonsterIDMapsCouncilorFamily`.
- [x] `DecodeCL2` doc comment records the single-dims limitation.
- [x] Gates: `gofmt -l internal/` empty; `go build ./...`; `go vet ./...`; `go test -count=1 ./...`
      0 FAIL; `scripts/lint.sh ./...` → `0 issues.`

## Session law
Commit trailer `Pi-Session: devilGo-5wsb-coder`. One commit. NEVER push, NEVER touch `.beans/`,
stdlib only. FIRST tool call is the cl2.go comment edit.

## Libraries
- none (comment + test only)

## Skills for coder
- `~/.agents/skills/go-dev/SKILL.md` — read fully BEFORE acting
- `~/.agents/skills/coding-rules/SKILL.md` — read fully BEFORE acting
- `~/.hermes/skills/coding-rules-notes/SKILL.md` — user amendments, read with coding-rules
- `~/.agents/skills/ponytail/SKILL.md` — read fully BEFORE acting

## Triage
Delegated brainer check (pi session `devilGo-5wsb-brainer`, report
`/home/blaze/.hermes/cache/scratch/dg-auditfix-check.md`); all four findings re-verified in the
parent by grep/read before filing (brainer Finding 5 about ledger re-homing was checked and
REJECTED — the ledger changes are in the tree).


## Proof of Work (2026-10-28, PR #3, commit `2da1e67`)

- F1: `cl2.go:182-188` command block now `0x80..0xBE` fill 1..63 / `0xBF` 65 literal /
  `0xC0..0xFF` literal 1..64 (cites `clx_decode.hpp:23-35`) — stale `0x80..0xBF` text gone.
- F2: `TestDecodeCL2/sheet_of_lists` added. **Revert-proven in the parent by mutation**: disabling
  the sheet branch makes it fail with `CL2: last frame offset 6656 != data size 51` (the exact
  pre-fix rejection); restoring passes 14/14 in-package CL2 tests.
- F3: `DecodeCL2` doc comment records the single-dims limitation vs `cl2_to_clx.cpp:66`.
- F4: orphaned `trnFixtureNames` comment removed above `TestMonsterTypeFromMonsterIDMapsCouncilorFamily`.
- Gates: `gofmt -l internal/` empty; build/vet ok; 26 pkgs ok / 0 FAIL; `scripts/lint.sh` `0 issues.`
