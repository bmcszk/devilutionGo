---
# devilGo-1v0o
title: 'devilGo-76tu C: DRLG quest hook, crypt variant, TestCryptParity gate'
status: todo
type: task
created_at: 2026-11-04T04:08:38Z
updated_at: 2026-11-04T04:08:38Z
parent: devilGo-97mn
blocked_by:
    - devilGo-gico
---

Subtask of devilGo-76tu, depends on the fix passes (devilGo-gico). Scope: the quest hook, the crypt variant, and the crypt parity gate.

## Goal
1. Port `DRLG_CheckQuests` so the `11-quests` stage dump equals the port state and the committed cathedral goldens match exactly (this closes `TestL1CathedralParity`).
2. Port the crypt differences of the same generator and wire a `TestCryptParity` gate over the committed l21-l24 fixtures (already in internal/testdata/drlg, currently unreferenced by any test).

## Upstream reference (READ-ONLY)
/tmp/dvlx-pinned = devilutionX @ ca91d778f.
- DRLG_CheckQuests: `grep -rn 'DRLG_CheckQuests' Source/` (quests.cpp) — the cathedral/crypt branches it takes for levels 1-4 and 21-24.
- Crypt-only functions in Source/levels/drlg_l1.cpp: FixCryptDirtTiles, CryptSubstitution, PlaceCryptStairs, PlaceCryptLights, SetCryptSetPieceRoom, plus `DRLG_LPass3` in Source/levels/gendung.cpp (Pass3 is part of CreateL5Dungeon at drlg_l1.cpp:1358-1372, so crypt parity needs it).
- Crypt set-piece data: nlevels/l5data/*.dun, already committed as internal/testdata/drlg/setpieces/*.csv.

## Requirements
- 1:1 port, `file:line` citations; unexported methods on *Dungeon in the landed style.
- `TestCryptParity` mirrors `TestL1CathedralParity`: levels 21-24 x main/prev x 8 seeds, byte equality of tiles/trans/view; it must FAIL (not skip) on a missing fixture.
- Every ported function gets at least one test; table-driven where multi-case.
- No new deps, no go.mod edits, no committed-fixture edits, no .beans writes, no git state changes.

## Acceptance criteria
- [ ] AC 1: `go test ./internal/diablo/levels/ -count=1` green: TestL1CathedralParity (l1-l4) AND TestCryptParity (l21-l24) plus the dev-stage oracles 01..11 where /tmp/drlg-stages exists.
- [ ] AC 2: `gofmt -l internal/diablo/levels` empty; `go vet ./internal/...` rc0; `./scripts/lint.sh` 0 issues; `go build ./...`.
- [ ] AC 3: Report: files, citations, gate tails, and any remaining red cell with got/want.

## Stop rules
Never edit a golden. Never weaken an assertion. If crypt cannot close, deliver the cathedral green and report the crypt cell precisely.

## Triage (brainer, 2026-11-04)
Delegated slice of a parent already triaged TRIAGED-OK (devilGo-76tu): the
upstream symbols, fixture path and gate are all cited in the body above and were
verified on disk (`Source/levels/drlg_l1.cpp` @ ca91d778f, committed corpus under
`internal/testdata/drlg/`, oracle dumps under `/tmp/drlg-stages`). No open
question blocks work; no new dependency is introduced.

## Libraries
- stdlib only (no third-party deps; no go.mod edits)

## Skills for coder
- `/home/blaze/.agents/skills/coding-rules/SKILL.md` — read fully BEFORE acting
- `/home/blaze/.agents/skills/ponytail/SKILL.md` — read fully BEFORE acting
- `/home/blaze/.agents/skills/go-dev/SKILL.md` — RNG-call-order discipline, 1:1 port rules
- `/home/blaze/.agents/skills/go-tests/SKILL.md` — table-driven tests, fixture matrix
