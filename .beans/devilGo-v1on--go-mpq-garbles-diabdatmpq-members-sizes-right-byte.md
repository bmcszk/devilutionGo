---
# devilGo-v1on
title: go-mpq garbles DIABDAT.MPQ members (sizes right, bytes wrong)
status: scrapped
type: bug
priority: high
created_at: 2026-11-05T23:56:47Z
updated_at: 2026-11-26T17:56:12Z
---

---
# devilGo-v1on
title: go-mpq garbles DIABDAT.MPQ members (sizes right, bytes wrong)
status: draft
type: bug
priority: high
created_at: 2026-11-05T23:56:47Z
updated_at: 2026-11-05T23:56:47Z
---

Found while extracting town assets for devilGo-t1lp.

## Symptom

`internal/engine/assets` (go-mpq v0.3.1) reads `devilutionx.mpq` and
`hellfire.mpq` correctly but returns **garbled bytes** for `DIABDAT.MPQ`
members. Sizes are right; content is not.

Evidence, `levels\l1data\l1.til` (must start `0,1,2,3` as uint16 LE, see
`internal/testdata/drlg/assets/l1.til.csv`):

```
$ go run /tmp/mpqlist DIABDAT.MPQ 'levels\l1data\l1.til'
levels\l1data\l1.til: 1174 bytes
1a da da ae 3c bd 37 c9 76 34 65 29 ...
```

Same reader, same call:

```
nlevels\towndata\town.til from hellfire.mpq == /tmp/mpq-hf/.../Town.TIL  (equal=true)
levels\towndata\automap.dun from devilutionx.mpq == plain 40x40 header       (ok)
```

Reference extractions at `/tmp/mpq-diablo/DIABDAT` and `/tmp/dg-assets-out/DIABDAT`
disagree with the reader: `sector1s.dun` starts `19 00 19 00` there, and the
reader returns `ac 00 47 d2`.

## Why it matters

Any Go path that extracts from retail `DIABDAT.MPQ` (asset pipeline, importer,
goldens) produces wrong bytes silently. The currently committed `drlg/assets`
CSVs came from `cmd/devilutionGo-migrate`, i.e. the broken reader — that is why
`assets/l1.til.csv` re-encoded to `4,0,7,0,...` (see `internal/testdata/drlg/README.md`,
"the five CSVs first committed with this corpus were wrong").

## Next step (triage)

Compare go-mpq's sector-offset vs hash-table handling against
`/tmp/mpq-diablo` (which decrypts correctly): is DIABDAT.MPQ using a protected
block table / different sector offset table format, or does the reader mis-handle
multi-sector files?

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



## Reasons for Scrapping

**Superseded — fixed before filing, and not on any live path.**

The garble this bean reports was real. It was the go-mpq v0.3.1 read path returning `MPQ_FILE_IMPLODE` payloads verbatim and skipping decryption of non-4-aligned sectors. It was diagnosed to root cause and fixed by the bug "MPQ_FILE_IMPLODE (PKWARE DCL) records are undecodable" (completed 2026-11-03, commits `c238493` + `ccfd719`): `internal/assets/mpqarchive.go` owns the read path and go-mpq stays only as the test-archive writer. This bean was filed 2026-11-05, two days after that fix landed.

**The bean's own acceptance shape holds on master.** Committed `internal/testdata/render/tiles/til/l1.til.hex` is a byte-level dump of `DIABDAT.MPQ` `levels\l1data\l1.til` and begins `0000 0100 0200 0300` — uint16 LE 0 1 2 3 — sha256-pinned by `render/tiles/manifest.txt` and `TestTilesFixtureManifest`. It agrees value-for-value with `internal/testdata/drlg/assets/l1.til.csv`. A fresh parent run of the production path reproduces the same head and size (1648 bytes).

**The named target is dead code.** `internal/engine/assets` has zero production importers; only its own test imports it. Every live reader — `present/scene` (town.go:90-92, warrior.go:142-143), `tools/goldens`, and `cmd/devilutionGo-migrate` — uses `internal/assets`, which does not read through go-mpq at all. So "any Go path that extracts from retail DIABDAT produces wrong bytes silently" is false today.

**The bean's "Why it matters" attribution is wrong.** It blames the committed drlg CSVs' `4,0,7,0,...` values on this reader. The drlg README (internal/testdata/drlg/README.md:204-206) records the actual cause: those CSVs held the raw **byte** stream instead of uint16 values — a CSV encoding bug in the producer, fixed 2026-11-04, unrelated to MPQ decoding.

**Its proposed next step is already answered.** "Compare go-mpq's sector offset vs hash-table handling" is the root cause the completed bug confirmed.

No re-scope is warranted. The package's lowercase `".mpq"` probe and the parent's NOT FOUND observation are latent only: nothing needs `internal/engine/assets` to open `DIABDAT.MPQ`, on Linux or elsewhere, and the production `Load` already takes a variadic ext for the retail case. The go-mpq v0.3.1 pin stays — it is the writer for synthetic test archives.

Triage: task-v1on-triage-brainer (2026-11-26, parent repro evidence + brainer tree audit with 4 corrections to the parent brief). Full verdict: .hermes/lane/gate-parity/v1on-triage.md
