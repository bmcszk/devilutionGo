# Proof of Work — asset verification wave

asset-verification-wave fingerprint: 5wsb verified against DIABDAT+hellfire+devilutionx.mpq on 2026-11-03

Branch: `dg-asset-pipeline-verification-wave`, base `41b257a` (origin/master).
Worktree: `/home/blaze/github/bmcszk/devilutionGo-assets-wave`.

Retail inputs used (read-only, never committed):
`/home/blaze/Games/Heroic/Diablo/DIABDAT.MPQ` (493 MiB),
`/home/blaze/Games/Heroic/Diablo/devilutionx.mpq`,
`/home/blaze/Games/Heroic/Diablo/hellfire/hellfire.mpq`.
`spawn.mpq` is absent locally — decision entry in `devilGo-5wsb` (see §Decisions).

## Gates (branch tip)

| Gate | Command | Result |
|---|---|---|
| build | `go build ./...` | rc0 |
| format | `gofmt -l internal/ cmd/` | empty |
| vet | `go vet ./...` | rc0 |
| lint | `scripts/lint.sh ./internal/assets/... --timeout=5m` | `0 issues.` |
| unit tests | `go test ./internal/assets/ -count=1` | `ok` |

## Stream 1 — asset pipeline (`devilGo-5wsb`)

### AC: extraction smoke over every retail record

New test: `internal/assets/mpq_smoke_test.go`
::`TestRealArchivesDecodeEveryRecord`. It iterates the `(listfile)` records of
`DIABDAT.MPQ`, `hellfire.mpq` and `devilutionx.mpq`, decodes each through
go-mpq into a scratch tree, and fails with the sorted per-record failure list.
The archives are environment inputs, so without `DG_ASSET_ARCHIVES` the test
skips instead of failing (same politeness as DevilutionX
`test/timedemo_test.cpp:47-49`).

Run (red — the AC is not met):

```
DG_ASSET_ARCHIVES=/home/blaze/Games/Heroic/Diablo:/home/blaze/Games/Heroic/Diablo/hellfire \
  go test ./internal/assets/ -run TestRealArchivesDecodeEveryRecord -v -timeout 40m
=== RUN   TestRealArchivesDecodeEveryRecord
    mpq_smoke_test.go:32: 3 record(s) failed to decode:
        DIABDAT.MPQ: listfile: extract listfile: file not found: (listfile)
        devilutionx.mpq: listfile: extract listfile: decrypt sectored file: decompress sector 0: multi zlib: create zlib reader: zlib: invalid header
        hellfire.mpq: listfile: extract listfile: file not found: (listfile)
--- FAIL: TestRealArchivesDecodeEveryRecord (0.00s)
```

Run without the archives (CI state):

```
go test ./internal/assets/ -run TestRealArchivesDecodeEveryRecord -v
--- SKIP: TestRealArchivesDecodeEveryRecord (0.00s)
    mpq_smoke_test.go:39: no retail MPQ found; set DG_ASSET_ARCHIVES to the Diablo install directories
```

Named-record spot checks (same codec path, name known):

```
devilutionx.mpq    fonts\12-05.clx        err=<nil> size=4304
DIABDAT.MPQ        ui_art\mainmenu.pcx   err=<nil> size=9161
hellfire.mpq       ui_art\mainmenu.pcx   err=<nil> size=5509
```

Findings (per-file failure list is the enumeration failure itself; nothing was
skipped silently):

* **F1 — retail archives hold no `(listfile)`.** `DIABDAT.MPQ` and
  `hellfire.mpq`: `HasFile("(listfile)") == false`. Record set cannot be
  enumerated by name at all, so "100% files decode" is not measurable through
  go-mpq alone; it needs the committed fallback name list the bean's
  open-question already anticipated. This is the bean's OPEN-QUESTION resolved
  in the negative.
* **F2 — go-mpq cannot decode `devilutionx.mpq`'s `(listfile)`.** The archive
  *does* carry one (`HasFile == true`), but extraction fails with
  `decompress sector 0: multi zlib: create zlib reader: zlib: invalid header`
  (`go-mpq@v0.3.1/compress.go:107-160` path). Ordinary files in the same
  archive decode (`fonts\12-05.clx`, 4304 bytes), so the defect is in the
  sectored-compressed path, not the archive. Enumeration of the port's own
  archive is blocked until that path is fixed or bypassed.
* **F3 — Huffman/ADPCM classes are unmeasurable with no enumeration.** The
  vendored reader upstream (`devilutionX/build-v2/_deps/mpqfs-src/src/mpq_stream.c:160-231`)
  accepts only PKWARE/zlib/bzip2 and explicitly rejects the rest;
  `go-mpq@v0.3.1/compress.go:85-95,117-135` rejects Huffman-only, ADPCM mono and
  ADPCM stereo. Those classes live on retail `.wav` records, which F1 makes
  unaddressable, so the AC's "proves Huffman not needed" claim is **not**
  provable in this wave.

Status: AC left **unticked** (evidence in `devilGo-5wsb` `## Findings`).

### AC: priority stack override

Existing tests prove it; no new test needed.

```
go test ./internal/assets/ -count=1 -run 'TestMPQStack' -v
--- PASS: TestMPQStackPriorityOverride/higher_priority_loaded_second_wins
--- PASS: TestMPQStackPriorityOverride/higher_priority_loaded_first_wins
--- PASS: TestMPQStackDuplicatePriority
--- PASS: TestMPQStackUnpackedDirectory/unpacked_tree_outranks_the_lower_priority_archive
```

`TestMPQStackPriorityOverride` registers the same inner name
(`txtdata\monsters.tsv`) at `MainMpqPriority` (1000) and
`DevilutionXMpqPriority` (9000) in both load orders and pins the high-priority
content; `TestMPQStackDuplicatePriority` pins upstream's keep-the-first rule for
an occupied priority (`assets.cpp:429-433`). AC ticked.

## Decisions

* `spawn.mpq` is not installed locally. Recorded in `devilGo-5wsb`: its content
  is the shareware slice of `DIABDAT.MPQ`, so `DIABDAT.MPQ` + `hellfire.mpq` +
  `devilutionx.mpq` cover the codec surface; the AC's `spawn.mpq` half is
  covered by the synthetic `spawn.mpq` archives the unit tests build
  (`TestMPQStackPriorityOverride` et al.). Amendment noted, not silent.
