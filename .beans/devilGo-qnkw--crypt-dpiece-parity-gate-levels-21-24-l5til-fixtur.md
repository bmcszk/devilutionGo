---
# devilGo-qnkw
title: Crypt dPiece parity gate (levels 21-24) + l5.til fixture
status: completed
type: task
priority: normal
created_at: 2026-11-06T14:48:01Z
updated_at: 2026-11-06T14:50:14Z
---

## Goal
Gate the Hellfire crypt Pass3 expansion: dPiece/dSpecial goldens for levels 21-24 x main/prev x the 8 pinned seeds, plus the l5.til mega-tile table the expansion reads.

## Research (disk, pinned upstream ca91d778f)
- internal/testdata/drlg/dpiece/ held levels 1-16 only (256 fixtures); crypt layout itself is already ported and scripted-tested by crypt_parity_test.go.
- internal/testdata/drlg/assets/ had l1-l4 .til/.min + town.til; l5.til absent.
- internal/diablo/levels/piece.go:71 already maps Crypt to mega-tile index 21 (Dirt-1, drlg_l1.cpp:1239-1241).
- nlevels\l5data\l5.til lives in hellfire.mpq (1736 bytes = 217 mega tiles x 4 uint16, headerless, GetTileCount(DTYPE_CRYPT)).

## Scope
internal/testdata/drlg/** (goldens, assets, README), internal/diablo/levels/dpiece_parity_test.go + dpiece_golden_test.go, tools/probe/** (additive, dump-only).

## Acceptance criteria
- [x] 64 crypt dPiece CSVs committed, naming/row shape identical to the existing corpus
- [x] assets/l5.til.csv committed + sha256 row in assets/manifest.txt
- [x] TestDPieceParity covers 21-24 GREEN with the corpus count pinned at 320
- [x] README records the crypt rows and l5.til
- [x] No game-logic edits (crypt.go / piece.go untouched)

## Proof of Work

### Fixtures (exactly 64 + 1)
- `ls internal/testdata/drlg/dpiece | wc -l` -> 320 (was 256; 64 new `l2[1-4]-dpiece-{main,prev}-<seed>.csv`)
- `ls internal/testdata/drlg/dpiece | grep -c 'l2[1-4]-dpiece'` -> 64
- `diff -rq` of the freshly converted 320-file corpus vs the committed dir -> 64 `Only in` lines, zero content diffs, so levels 1-16 were reproduced byte-for-byte
- `internal/testdata/drlg/assets/l5.til.csv`: header `217,4` + 868 micro tiles; manifest sha256 `3393ce3eba937a1b0e521d71d7438d97aafb28677560e905dd6b3aec257b170d` equals sha256 of the `nlevels\l5data\l5.til` member bytes it re-encodes

### Gate (GREEN, no game-logic edit)
- `go test ./internal/diablo/levels/ -run TestDPieceParity -v | grep -c 'PASS: TestDPieceParity/l2[1-4]'` -> 64
- spot check: `TestDPieceParity/l24-main-seed-2588 ... PASS`, `l24-prev-seed-2588 ... PASS`
- `TestDPieceCorpus` pins `len(dpieceVariants) * len(committedSeeds)` = 320

### Commands (real tails)
- `go build ./...` -> rc=0
- `gofmt -l .` -> empty
- `./scripts/lint.sh` -> `0 issues.` (simpurity: clean, 229 sources)
- `go vet ./internal/...` -> rc=0
- `go test ./... -count=1` -> 43 `ok` packages, 0 FAIL lines
- `python3 scripts/beans-lint.py` -> `beans-lint: OK (177 beans)`
- `just sqlc-check` -> `no raw SQL outside sqlc-generated code`, rc=0

### Probe recipe (reproducible)
1. Extract `nlevels\l5data\l5.{til,cel}` and `l5s.cel` from `hellfire.mpq` loose under `<build>/assets/nlevels/l5data/` (the probe build has no `hfmonk/hfmusic/hfvoice`, so `LoadHellfireArchives()` cannot be used).
2. `/tmp/dvlx-pinned/build-probe/dvg_probe dpiece <out> tools/probe/seeds.txt`
3. `go run ./tools/probe/dunconv dpiece <out>/dpiece internal/testdata/drlg/dpiece`
4. `dvg_probe asset <out> l5.til 'nlevels\l5data\l5.til'` then `go run ./tools/probe/dunconv til <raw-dir> internal/testdata/drlg/assets`
