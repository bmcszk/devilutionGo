---
# devilGo-v7ft
title: 'Migrate assets: MPQ to assets dir plus YAML asset lists'
status: completed
type: task
priority: high
created_at: 2026-11-03T12:49:56Z
updated_at: 2026-11-03T15:28:10Z
parent: devilGo-o8e7
---

# Migrate assets: MPQ to assets dir plus CSV tables and YAML asset lists

## Goal

ADR-014 assets subcommand on top of the ADR-013 law: `devilutionGo-migrate
assets` walks the configured MPQ set, extracts through the proven decoders at
full coverage, writes the assets dir layout, and emits YAML asset lists plus
CSV tables. The YAML is the manifest record: family, archive path, format,
frame layout, hashes; the graphics manifest work converges onto this same
YAML form (no TSV fork).

## Research

- internal/assets: exhaustive extraction proven on retail DIABDAT + hellfire
  + patch MPQs (100% of records, zero failures; PCX decodes).
- The graphics manifest feature names the manifest craft this tool converges
  with; ADR-013 fixed YAML as the list format.
- docs/recipes holds the extraction conventions workers follow today.

## Scope

- assets subcommand; YAML manifest writer; CSV table writer; hash summary.
- Tests on synthetic fixtures only (no MPQ in the repository); retail
  verification documented as an environment amendment, never a silent skip.

## Acceptance criteria

- [x] Given a retail MPQ in the configured dir, migration writes the assets dir, a valid YAML manifest and the CSV tables with per-file hashes.
- [x] Re-running is idempotent (same output hash summary).
- [x] go test ./... green on synthetic fixtures; no binaries committed.

## Libraries

- None new.

## Skills for coder

- /home/blaze/.agents/skills/coding-rules/SKILL.md
- /home/blaze/.agents/skills/ponytail/SKILL.md
- /home/blaze/.agents/skills/go-dev/SKILL.md
- /home/blaze/.agents/skills/go-tests/SKILL.md
## Dependencies & blockers

- Parent accepted; graphics manifest work converges onto the YAML target built here.

## Open questions

- Manifest filename + location: assets/manifest.yaml unless the graphics
  manifest feature already fixed another.

## Triage

TRIAGE brainer evidence 2026-11-03: brainer REQUEST-CHANGES on the architecture
PR produced findings; this text was rewritten against the recorded law and the
verified tree (real file seams, no invented files), lint-clean.

Claim evidence 2026-11-03: lane dispatched. Worker session task-assets-lane-coder,
branch dg-migrate-assets (worktree /home/blaze/github/bmcszk/dg-lane-migrate),
off master 1cb9ffd. Deliverable: the migration assets subcommand per this
bean; PR targets master.

## Progress log

2026-11-03 — assets lane landed end to end in `cmd/devilutionGo-migrate`:

- `main.go`: dispatch (`assets`, unknown → usage + exit 2, `saves`/`settings` →
  "not implemented yet" + exit 2).
- `assets.go`: `--mpq` (repeatable) / `--out` (default `./assets-out`) /
  `--dry-run`; discovery from `DG_ASSET_ARCHIVES` (path-list, `.mpq` files,
  sorted); base-name clash guard; per-run report + checksum summary.
- `extract.go`: one record per archive, output tree mirrors archive paths
  (`<out>/<archive-stem>/<inner path>`), decode gate per family, CEL/CL2 copied
  with no frame layout (dims live in caller tables upstream, documented in
  code), decoder panics contained per record.
- `manifest.go`: `manifest.yaml` (tool/version, archive fingerprints, per-file
  family/format/frames/rows/sha256, summary) + CSV writer for txtdata.
- `journal.go`: `.migrate-state.json` keyed by archive path with size+mtime
  fingerprint and per-file hashes; resume reuses only when the archive and every
  listed output are unchanged.
- Unaddressable archives (no `(listfile)`, no committed inventory) are reported
  and skipped instead of failing the run: retail `DRTL_IX86_108_109.mpq` is one.
- `internal/assets/names.go` + `internal/assets/names/` (moved from
  `internal/testdata/assets/mpq/`): the committed retail name inventories are now
  embedded and exported as `assets.NameInventory`, because the tool needs them at
  runtime and MPQ decrypts by name.

## Proof of Work

### Commits (branch `dg-migrate-assets`, base 92a5b73)

| commit | unit |
|---|---|
| 794ec56 | migrate assets subcommand (dispatch, discovery, extraction, CSV, YAML manifest, journal, tests) |
| 3aa3ff4 | report and skip archives whose records cannot be named |
| 85c5275 | name a format for extensionless records; contain decoder panics per record |

### Files

New — `cmd/devilutionGo-migrate/`: `main.go`, `assets.go`, `extract.go`,
`manifest.go`, `journal.go`; tests `main_test.go`, `assets_test.go`,
`extract_test.go`, `manifest_test.go`, `journal_test.go`, `retail_test.go`,
`fixtures_test.go`.

New — `internal/assets/names.go` (embedded `NameInventory`), `internal/assets/names/`
(`diabdat.txt` 2910 names, `hellfire.txt` 584 names, `README.md`; moved from
`internal/testdata/assets/mpq/`, path references updated).

Changed — `internal/assets/mpq_smoke_test.go` (reads `assets.NameInventory`
instead of the testdata path), `go.mod`/`go.sum` (`gopkg.in/yaml.v3 v3.0.1`,
catalog row 145, no other new dependency).

### Gates (run in this worktree)

```
gofmt -l .                  -> (empty)
go build ./...              -> ok
go vet ./...                -> ok
scripts/lint.sh ./...       -> 0 issues.
go test ./... -count=1      -> all packages ok
```

`go test ./cmd/devilutionGo-migrate/ -count=1 -v -run TestRetailAssetsMigration`
without the env: `SKIP: DG_ASSET_ARCHIVES is unset: retail archives are not
installed in this environment`. With it: `ok ... 5.621s`.

### Retail verification (environment amendment, not a CI dependency)

`DG_ASSET_ARCHIVES=/home/blaze/Games/Heroic/Diablo devilutionGo-migrate assets
--out /tmp/dg-assets-out` (DIABDAT.MPQ, Patch_rt.mpq, devilutionx.mpq;
hellfire.mpq lives in the game's `hellfire/` subdir and was not in the scanned dir):

```
skipped archive DRTL_IX86_108_109.mpq: ships no (listfile) and no committed name inventory
wrote 3134 file(s), skipped 0 already-journaled file(s), 0 failure(s)
summary sha256 8c37b1f2e32bec9f637dfd836d5466dc8ac5648af836f00a05823e6caf8d8afe
```

- 4.2 s cold run, 0 decode failures; families sprite 1530, raw 1291, trn 166,
  font 58, table 47, palette 42; 47 CSV tables written (e.g.
  `devilutionx/txtdata/Experience.csv`, 50 rows).
- Rerun on the same dir: `wrote 0 file(s), skipped 3134 ...` in 0.44 s with the
  same summary (journal resume).
- Two clean runs into different dirs: `manifest.yaml` byte-identical apart from
  the out path (verified with `diff`), same summary.
- Fixes the retail run forced: extensionless records get `format: raw` (was an
  empty string), and a decoder panic on retail bytes is contained as a record
  failure instead of killing the batch.

### Known limits (documented, not silent)

- CEL and CL2 are extracted byte-identically without a decode gate: their frame
  dimensions live in the caller's tables upstream (`cel_to_clx.cpp:84-86`,
  `cl2_to_clx.cpp:66`), not in the container. The manifest carries no `frames`
  for them.
- Discovery is per configured directory, not recursive.
- An archive with no `(listfile)` and no committed inventory is reported and
  skipped; it contributes no files.
