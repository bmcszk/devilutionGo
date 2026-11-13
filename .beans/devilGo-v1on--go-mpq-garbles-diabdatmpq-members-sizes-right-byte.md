---
# devilGo-v1on
title: go-mpq garbles DIABDAT.MPQ members (sizes right, bytes wrong)
status: draft
type: bug
priority: high
created_at: 2026-11-05T23:56:47Z
updated_at: 2026-11-13T00:33:35Z
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
