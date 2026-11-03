# Proof of Work — asset verification wave

asset-verification-wave fingerprint: 5wsb verified against DIABDAT+hellfire+devilutionx.mpq on 2026-11-03

Branch: `dg-asset-pipeline-verification-wave`, base `41b257a` (origin/master).
Worktree: `/home/blaze/github/bmcszk/devilutionGo-assets-wave`.

Retail inputs used (read-only, never committed):
`/home/blaze/Games/Heroic/Diablo/DIABDAT.MPQ` (493 MiB),
`/home/blaze/Games/Heroic/Diablo/devilutionx.mpq`,
`/home/blaze/Games/Heroic/Diablo/hellfire/hellfire.mpq`,
`Patch_rt.mpq`, `hfmusic.mpq`.
`spawn.mpq` is absent locally — decision entry in `devilGo-5wsb` (§Decisions).

## Gates (branch tip)

| Gate | Command | Result |
|---|---|---|
| build | `go build ./...` | rc0 |
| format | `gofmt -l internal/ cmd/` | empty |
| vet | `go vet ./...` | rc0 |
| lint | `scripts/lint.sh ./internal/... ./cmd/... --timeout=6m` | `0 issues.` |
| unit tests | `go test ./... -count=1` | all `ok` |

## Stream 1 — asset pipeline (`devilGo-5wsb`)

### AC results

| # | AC | Verdict | Evidence |
|---|---|---|---|
| 1 | extraction smoke, 100% decode | **not met** | new test red; F1/F2/F4 |
| 2 | priority stack override | met | `TestMPQStackPriorityOverride`, `TestMPQStackDuplicatePriority` |
| 3 | CLX frames vs independent decode | met (deviation) | `TestCLXDecodeMatchesCommittedHashes` (21 files), `TestCLXGoldenFrame0Bytes` |
| 4 | PCX ui_art pixel-exact | **not met** | F4; only the synthetic 4x3 fixture covers the decoder |
| 5 | TRN 256-entry table | met | `TestDecodeTRNGolden` (strengthened), `TestApplyTRNRemapsIndicesInPlace` |
| 6 | txtdata exact row counts | met (AC number corrected) | `TestUpstreamTablesParseFully`, `TestUpstreamManifestCoversCorpus` |
| 7 | txtdata edge cases | met (deviation recorded) | `TestLoadNormalizesLineEndings`, `TestLoadEmptyIsEmptyTable`, `TestLoadRejectsCorruptTable` |
| 8 | SFX 22050 / music switch / no per-frame alloc | **partial** | catalog + `TestGetLevelMusic` only; no audio device layer |

### AC1 — exhaustive extraction smoke (not met)

New test `internal/assets/mpq_smoke_test.go::TestRealArchivesDecodeEveryRecord`
iterates the `(listfile)` records of `DIABDAT.MPQ`, `hellfire.mpq` and
`devilutionx.mpq`, decodes each through go-mpq into a scratch tree, and fails
with the sorted per-record failure list. Without `DG_ASSET_ARCHIVES` it skips
(MPQs are environment inputs; same politeness as DevilutionX
`test/timedemo_test.cpp:47-49`).

Red run (the AC is not met):

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

CI state (no MPQs installed):

```
go test ./internal/assets/ -run TestRealArchivesDecodeEveryRecord -v
--- SKIP: TestRealArchivesDecodeEveryRecord (0.00s)
    mpq_smoke_test.go:39: no retail MPQ found; set DG_ASSET_ARCHIVES to the Diablo install directories
```

Because enumeration by name is impossible for the retail archives, the record
census was measured instead by decrypting the MPQ block and hash tables with the
StormLib table keys (`"(block table)"`, `"(hash table)"`, PKWARE
`HashString(name, MPQ_HASH_FILE_KEY)`):

| archive | file bytes | records | IMPLODE 0x100 | COMPRESS 0x200 | stored | encrypted |
|---|---|---|---|---|---|---|
| DIABDAT.MPQ | 517501282 | 2910 | 1752 | 0 | 1158 | 2900 |
| hellfire.mpq | 65502336 | 584 | 582 | 0 | 2 | 582 |
| devilutionx.mpq | 5664284 | 210 | 0 | 210 | 0 | 1 |
| Patch_rt.mpq | 593853 | 16 | 15 | 1 | 0 | 16 |
| hfmusic.mpq | 34379360 | 2 | 2 | 0 | 0 | 2 |

`sum(compressedSize)` lands within 0.03% of each archive's byte length
(DIABDAT 517346918 vs 517501282), so the census is coherent.

Findings:

* **F1 — the retail archives hold no `(listfile)`** (`HasFile("(listfile)") ==
  false` for `DIABDAT.MPQ` and `hellfire.mpq`). Records cannot be addressed by
  name, so "100% of files decode" is not measurable through go-mpq alone: it
  needs the committed fallback name list the bean's open question anticipated.
  That open question resolves in the negative.
* **F2 — go-mpq cannot decode `devilutionx.mpq`'s `(listfile)`.** The archive
  does carry one, but extraction fails with `decompress sector 0: multi zlib:
  create zlib reader: zlib: invalid header`. It is the single encrypted record in
  that archive (table above), so the defect is in the encrypted-sectored path,
  not in the archive.
* **F4 — retail payloads use the legacy `MPQ_FILE_IMPLODE` flag, which go-mpq
  ignores.** 1752/2910 DIABDAT and 582/584 hellfire records are imploded
  (PKWARE DCL). go-mpq decodes only `MPQ_FILE_COMPRESS` payloads
  (`go-mpq@v0.3.1/mpq.go:518-563`) and returns an imploded payload verbatim with
  a nil error — silent corruption. Extract of `ui_art\mainmenu.pcx` from
  `DIABDAT.MPQ` "succeeds" with 9161 bytes beginning `39 69 EF FA`;
  `assets.DecodePCX` rejects them (`invalid manufacturer byte 0x39 (expected
  0x0A)`). Upstream handles the flag explicitly
  (`devilutionX/build-v2/_deps/mpqfs-src/src/mpq_stream.c:153-160` routes
  `MPQ_FILE_IMPLODE` through `pk_explode_sector`). Filed as draft bug
  `devilGo-76ei`.
* **F5 — the Huffman claim is measurable and holds, but it is not what breaks
  the AC.** No record in `DIABDAT.MPQ` or `hellfire.mpq` carries the
  `MPQ_FILE_COMPRESS` mask, so Huffman/ADPCM never appear there. 100% decode
  fails on IMPLODE, not on Huffman.

### AC2 — priority stack (met, no new test needed)

```
go test ./internal/assets/ -count=1 -run 'TestMPQStack' -v
--- PASS: TestMPQStackPriorityOverride/higher_priority_loaded_second_wins
--- PASS: TestMPQStackPriorityOverride/higher_priority_loaded_first_wins
--- PASS: TestMPQStackDuplicatePriority
--- PASS: TestMPQStackUnpackedDirectory/unpacked_tree_outranks_the_lower_priority_archive
```

`TestMPQStackPriorityOverride` registers the same inner name
(`txtdata\monsters.tsv`) at `MainMpqPriority` (1000) and
`DevilutionXMpqPriority` (9000) in both load orders and pins the
high-priority content; `TestMPQStackDuplicatePriority` pins upstream's
keep-the-first rule for an occupied priority (`assets.cpp:429-433`).

### AC3 — CLX goldens (met, deviation recorded)

```
go test ./internal/assets/ -count=1 -run 'TestCLX' -v
--- PASS: TestCLXDecodeMatchesCommittedHashes        (21 committed files)
--- PASS: TestCLXGoldenFrame0Bytes                  (3 full frame-0 dumps)
--- PASS: TestCLXLiteralRunAt0xBF / TestCLXFillRunAt0xBE / ...
```

Deviation (recorded in `devilGo-5wsb`): the oracle is an independent **Python**
port of `clx_decode.hpp:23-35` writing `internal/testdata/gfx/clx/*.sha256`
(digest of concatenated decoded indices + dims + frame count), not a compiled
C++ dump; the source bytes are the devilutionX checkout's `assets/` at
`ca91d778fba92a1f680afb1fb2cdf09337ff1b42`, not retail MPQ bytes
(`internal/testdata/gfx/README.md`).

### AC5 — TRN 256-entry table (met, test strengthened)

`TestDecodeTRNGolden` previously spot-checked six entries; it now compares **all
256** entries against the committed 256-byte `gendata/pause.trn` fixture
(`internal/testdata/gfx/trn/gendata_pause.trn.hex`). `TestApplyTRNRemapsIndicesInPlace`
pins `dst = trn[src]` (`clx_render.cpp:423`).

```
go test ./internal/assets/ -count=1 -run 'TestDecodeTRNGolden|TestApplyTRN' -v
--- PASS: TestDecodeTRNGolden
--- PASS: TestApplyTRNRemapsIndicesInPlace
```

### AC6 — txtdata row counts (met, AC number corrected)

49 committed corpus tables are drained and counted against
`internal/testdata/assets/txtdata/manifest.tsv`, whose counts were measured once
offline from the pinned checkout and whose sha256s pin the bytes; a missing
fixture fails rather than skips.

```
go test ./internal/assets/ -count=1 -run 'TestUpstream' -v | tail -3
--- PASS: TestUpstreamTablesParseFully
--- PASS: TestUpstreamManifestCoversCorpus
```

Measured counts (`awk 'END{print NR}'` over the committed files): monstdat **113
lines = header + 112 data rows** (the AC's "113 data lines" counted the header),
Experience 51 lines = 50 data rows, spelldat 36, misdat 68, itemdat 168,
towners 12, quest_dialog 10. The synthetic schema fixture
`golden/synthetic-monstdat-schema.tsv` carries 113 *data* rows, one more than
the corpus — it holds invented values and claims no corpus parity, but it
encodes the AC's miscount and is worth one row of alignment later.

### AC7 — txtdata edge cases (met, deviation recorded)

`TestLoadNormalizesLineEndings` covers `bom.tsv`, `crlf.tsv`, `cr.tsv`,
`lf_no_trail.tsv`; `TestLoadEmptyIsEmptyTable` covers the zero-byte export;
`TestLoadRejectsCorruptTable` covers a ragged row and a duplicate column.
Fixtures live in `internal/testdata/assets/txtdata`, per the zero-dependency
rule, not in `test/fixtures/txtdata`.

Deviation (recorded in `devilGo-5wsb`, filed as draft `devilGo-mihf`): the
loader normalizes a lone CR and accepts `cr.tsv`, while upstream returns
`BadRecordTerminator` when a CR is not followed by LF
(`Source/data/parser.cpp:4-27`). Unreachable with the real corpus (CRLF only),
but it is a behavior difference and is recorded rather than silent.

### AC8 — audio (partial)

Ported and pinned: the effects catalog and 22050 default
(`internal/diablo/gamedata/audio`, `DefaultSampleRate` ← `options.cpp:55`,
`options.cpp:520`), the level→music decision (`TestGetLevelMusic`), the effect
path/rate contract (`TestSfxCatalogSamplesAtDefaultRate`), movies manifest
(`TestMovieManifest`).

Absent, so the AC is not met: no `ebiten/v2/audio.Context` at 22050 that
actually plays a sample, no music track switch at runtime, and no allocation
evidence for the audio path.

## Decisions

* `spawn.mpq` is not installed locally. Recorded in `devilGo-5wsb`: its content
  is the shareware slice of `DIABDAT.MPQ`, so `DIABDAT.MPQ` + `hellfire.mpq` +
  `devilutionx.mpq` cover the codec surface, and the AC's `spawn.mpq` half is
  covered by the synthetic `spawn.mpq` archives the unit tests build. Amendment
  noted, not silent.
* CLX golden oracle is an independent Python port, not a compiled C++ dump (AC3
  above). Byte source is the pinned devilutionX checkout, not retail MPQ.
* txtdata lone-CR leniency is a recorded deviation from upstream; the retail
  corpus is CRLF-only.

## Follow-ups filed

* `devilGo-76ei` (draft, high) — `MPQ_FILE_IMPLODE` records are undecodable:
  retail DIABDAT/hellfire payloads return verbatim.
* `devilGo-mihf` (draft) — txtdata loader accepts a lone CR where upstream
  returns `BadRecordTerminator`.
