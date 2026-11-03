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
| unit tests | `go test ./... -count=1` | 37 packages `ok`, 0 failures; `-v` PASS lines 993, FAIL 0 |

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

## Stream 2 — stale-bean gate verification (read-only + bean writes)

Every gate below ran in this worktree at the branch tip. No source file was
edited in this stream.

| bean | gates | AC verdict | outcome |
|---|---|---|---|
| devilGo-4xin corpses | `go test ./internal/diablo/entities/...` ok | 5/5 ticked with test names | **completed** |
| devilGo-oqho packs/streams | `go test ./internal/persist/...` ok | 3/3 ticked (`TestItemPack*`, `TestSaveGameStatesGolden`, `TestLoadLevelStateHashRoundTrip`) | **completed** |
| devilGo-8wdk appdiag | `go test ./cmd/... ./internal/diablo/appdiag/... ./internal/engine/appfat/...` ok | AC2 + AC4 ticked; AC1 caller-blocked (111 `app_fatal(` + 80 `ErrDlg/ErrSdl` sites upstream, callers unported); AC3 wave-gated | in-progress |
| devilGo-baxb affix/drop | `go test ./internal/diablo/gamedata/itemgen/...` ok (143 tests) | AC1/AC3/AC4 ticked; AC2 half open (no C++ probe fixture, see v16/expansion findings) | in-progress |
| devilGo-1n0a testing/parity | `go test ./...` ok | 0/5 met: no Appendix K matrix, no timedemo replay, no `.dun` fixtures (`find . -name '*.dun'` empty), no C++ state-hash diff, benchmarks exist but no CI budget | in-progress |
| devilGo-klu5 coding-rules | build/vet/gofmt/lint green | per-file table below: 483 files, 479 clean, 4 flagged (5 marker lines) | in-progress |
| devilGo-ocl1 phase markers | same + marker scan | 996/996 test funcs marked, 0 when-before-given, 0 then-before-when, 4 flagged files; docstring target fails | in-progress |

Marker findings, verified by reading (not just by the scanner):

| file:line | test | verdict |
|---|---|---|
| `internal/diablo/missiles/add_slice2_test.go:159` | `TestAddChargedBolt` | two `// when` lines (159, 161); the first has no statement — violation |
| `internal/diablo/missiles/add_slice2_test.go:191` | `TestAddHolyBolt` | same shape (191, 193) — violation |
| `internal/diablo/stores/hold_test.go:206` | `TestBodySlotMapping` | `// when` then `// then` with no statement under the when — violation |
| `internal/diablo/inventory/grid_test.go:327` | `TestReorganizeInventory` | one `// then` split across 327/328, second line starts with "then" — ambiguity |
| `internal/diablo/missiles/processmissiles_test.go:81` | `TestProcessMissiles_ClearsMissileTileFlagsBeforeThePass` | `// then: ...` whose prose continues on 82 starting with "then" — ambiguity |
| `internal/assets/trn_test.go:289` | `TestTRNFixtureCensus` | descriptive comment line directly above the `// when` marker — scanner may read it as prose continuation |

Also open on `devilGo-ocl1`: `scripts/fixture_fixtures.py:2-26` module docstring is
25 lines (22 non-blank) against the AC's "~8 contract lines", and the AC's "52
test files" inventory is stale (247 today).

`devilGo-klu5` cannot close with the five marker findings above; its LOC-delta AC
is answered by Appendix C (no source file changed in this wave).

## Stream 3 — devilGo-mwlm (missiles) closure

Substance verified from the merged slices and today's run; the three open gates
were PoW boilerplate and are ticked with the evidence below.

```
go test ./internal/diablo/missiles/... ./internal/diablo/gamedata/misdat/... -count=1
ok  github.com/bmcszk/devilutionGo/internal/diablo/missiles
ok  github.com/bmcszk/devilutionGo/internal/diablo/gamedata/misdat
```

* Parity: `TestRegistryCompleteness` counts the committed corpus at runtime (50
  distinct `addFn`, 36 distinct `processFn`, every name registered),
  `TestRowCountsParity` (68 rows), `TestHFMisdatStrictSuperset`,
  `TestHFRegistrySuperset`, `TestMovementDistributionDomain`, plus the
  `CheckMissileCol` rotation/collision tapes.
* TDD: RED evidence per slice lives in the Progress log and the slice commits
  (a2b723b, 8533ff6, 66f7264) — slice 2's inverted-expectation probe and slice
  3b's mutation-RED-checked tests. Raw transcripts are in those session logs, not
  the repo; the repo carries the resulting tests.
* Deviations (now recorded in this bean **and** `docs/requirements.md` §V3
  "Recorded deviations (Go port, devilGo-mwlm)"): `ProcessMissiles` does not
  advance animation/ManaShield timers; unported Hellfire `processFn`s dispatch to
  `unportedProcess` (`internal/diablo/missiles/registry.go:563,571,578`).

Bean marked **completed** with a Summary + Proof of Work section.

## Appendix A — per-file verdict table (483 files)

Columns: file, `go vet`, `golangci-lint`, `gofmt`, phase markers. `CLEAN` means no
finding from that tool for that file; `-` means the check does not apply
(non-test file for markers). Running totals: 483 files, 996 test functions, 479
clean, 4 flagged.

```
file	vet	lint	gofmt	markers
cache/lint/canary/main.go	CLEAN	CLEAN	CLEAN	-
cmd/devilutionGo/main.go	CLEAN	CLEAN	CLEAN	-
cmd/devilutionGo-headless/main.go	CLEAN	CLEAN	CLEAN	-
cmd/devilutionGo-headless/main_test.go	CLEAN	CLEAN	CLEAN	CLEAN
cmd/gamedatagoldens/main.go	CLEAN	CLEAN	CLEAN	-
internal/assets/cel.go	CLEAN	CLEAN	CLEAN	-
internal/assets/cel_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/assets/cl2.go	CLEAN	CLEAN	CLEAN	-
internal/assets/cl2_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/assets/clx.go	CLEAN	CLEAN	CLEAN	-
internal/assets/clx_bench_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/assets/clx_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/assets/color_translations.go	CLEAN	CLEAN	CLEAN	-
internal/assets/color_translations_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/assets/font.go	CLEAN	CLEAN	CLEAN	-
internal/assets/font_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/assets/mpq.go	CLEAN	CLEAN	CLEAN	-
internal/assets/mpq_smoke_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/assets/mpq_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/assets/palette.go	CLEAN	CLEAN	CLEAN	-
internal/assets/palette_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/assets/pcx.go	CLEAN	CLEAN	CLEAN	-
internal/assets/pcx_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/assets/trn.go	CLEAN	CLEAN	CLEAN	-
internal/assets/trn_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/assets/txtdata.go	CLEAN	CLEAN	CLEAN	-
internal/assets/txtdata_golden_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/assets/txtdata_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/animation/animationinfo.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/animation/animationinfo_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/appdiag/appdiag.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/appdiag/appdiag_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/combat/combat.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/combat/combat_fuzz_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/combat/combat_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/doc.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/entities/corpse.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/entities/corpse_decode_fixture_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/entities/corpse_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/entities/entities.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/entities/imports_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/entities/item.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/entities/monster.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/entities/object.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/entities/player.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/entities/size_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/enums/cursors.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/enums/durability.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/enums/enums_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/enums/gold.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/enums/simflags.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/enums/simflags_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/enums/spawn.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/gamedata/audio/audio.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/gamedata/audio/audio_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/audio/sfx.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/gamedata/audio/sfx_ids.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/gamedata/audio/sfx_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/audit/audit.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/gamedata/audit/audit_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/experience/experience.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/gamedata/experience/experience_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/experience/fixture_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/itemdat/itemdat.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/gamedata/itemdat/itemdat_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/itemgen/affix.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/gamedata/itemgen/affix_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/itemgen/attrs.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/gamedata/itemgen/attrs_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/itemgen/book.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/gamedata/itemgen/book_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/itemgen/buff_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/itemgen/coin_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/itemgen/distribution_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/itemgen/drop.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/gamedata/itemgen/drop_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/itemgen/durability.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/gamedata/itemgen/durability_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/itemgen/expansion_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/itemgen/histogram_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/itemgen/ivalue_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/itemgen/level.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/gamedata/itemgen/level_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/itemgen/mincap_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/itemgen/mods.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/gamedata/itemgen/mods_table_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/itemgen/mods_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/itemgen/name.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/gamedata/itemgen/name_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/itemgen/oil.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/gamedata/itemgen/oil_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/itemgen/outline.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/gamedata/itemgen/outline_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/itemgen/pools.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/gamedata/itemgen/pools_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/itemgen/power_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/itemgen/setup.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/gamedata/itemgen/setup_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/itemgen/spawn_histogram_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/itemgen/spawn_item.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/gamedata/itemgen/spawn_item_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/itemgen/spawn_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/itemgen/staff.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/gamedata/itemgen/staff_spell_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/itemgen/staff_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/itemgen/unique.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/gamedata/itemgen/unique_cursor_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/itemgen/unique_level_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/itemgen/unique_spawn.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/gamedata/itemgen/unique_spawn_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/itemgen/unique_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/itemgen/useful.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/gamedata/itemgen/useful_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/itemgen/v16_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/itemgen/value.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/gamedata/leveldat/leveldat.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/gamedata/leveldat/leveldat_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/misdat/hf_registry_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/misdat/misdat.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/gamedata/misdat/misdat_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/misdat/movement_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/misdat/registry.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/gamedata/misdat/stub_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/monstdat/monstdat.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/gamedata/monstdat/monstdat_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/objdat/objdat.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/gamedata/objdat/objdat_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/objdat/operators.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/gamedata/quests/availability.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/gamedata/quests/availability_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/quests/forcequests.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/gamedata/quests/forcequests_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/quests/helpers_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/quests/init.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/gamedata/quests/init_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/quests/kill.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/gamedata/quests/kill_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/quests/log.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/gamedata/quests/log_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/quests/multiquest.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/gamedata/quests/multiquest_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/quests/pools_seed_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/quests/quests.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/gamedata/quests/quests_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/quests/resync.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/gamedata/quests/resync_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/quests/returnlevel.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/gamedata/quests/returnlevel_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/quests/towners.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/gamedata/quests/towners_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/quests/victim.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/gamedata/quests/victim_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/textdat/textdat.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/gamedata/textdat/textdat_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/gamedata/towners/towners.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/gamedata/towners/towners_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/inventory/belt.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/inventory/belt_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/inventory/consume.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/inventory/consume_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/inventory/equip.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/inventory/equip_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/inventory/gold.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/inventory/gold_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/inventory/grid.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/inventory/grid_test.go	CLEAN	CLEAN	CLEAN	markers:TestReorganizeInventory:empty-then@19
internal/diablo/inventory/hold.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/inventory/hold_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/inventory/inventory.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/acidpuddle_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/acidsplate_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/add.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/add_acid_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/add_arrow_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/add_berserk.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/add_berserk_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/add_bigexplosion.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/add_bigexplosion_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/add_bows.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/add_bows_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/add_diabloapocalypse.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/add_elementalarrow_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/add_explosion_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/add_fireball_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/add_firebolt_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/add_firewall_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/add_fixture_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/add_golem.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/add_guardian.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/add_helpers.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/add_lightning_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/add_magicmissile_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/add_magmaball_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/add_mana.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/add_mana_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/add_novaball_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/add_reflect.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/add_ringoffire.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/add_runes.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/add_runes_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/add_slice2_test.go	CLEAN	CLEAN	CLEAN	markers:TestAddChargedBolt:empty-when@5,TestAddHolyBolt:empty-when@5
internal/diablo/missiles/add_spawns.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/add_spawns_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/add_stealpotions.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/add_stealpotions_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/add_warp.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/add_warp_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/addmissile.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/addmissile_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/apocalypse_hellfire_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/apocalypse_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/apocalypseboom_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/appfat_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/bonespirit_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/callbacks_belt.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/callbacks_berserk.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/callbacks_monster.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/cast_world_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/chainlightning_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/collection.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/collection_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/collision.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/collision_gates_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/collision_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/cursoradds_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/diabloapocalypse_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/direction.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/direction16.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/direction16_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/direction_helpers.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/direction_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/displacement.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/displacement_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/elemental_arrow.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/elemental_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/flamewave_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/flashbottom_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/flashtop_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/golem_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/guardian_fade_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/guardian_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/healing_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/hellfire_registry_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/inferno_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/infernocontrol_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/infravision_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/lightningcontrol_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/line.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/line_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/manashield_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/misdat_wiring_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/missile.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/missilelist.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/missilelist_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/motion.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/noop_callbacks.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/nova_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/phasing_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/process.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/process_acidpuddle.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/process_acidsplate.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/process_apocalypse.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/process_apocalypsebox.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/process_bonespirit.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/process_callbacks_mana.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/process_callbacks_spawn.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/process_chainlightning.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/process_elemental.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/process_firewall.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/process_firewall_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/process_flash.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/process_flashbottom.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/process_guardian.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/process_ii_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/process_iii_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/process_inferno.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/process_infernocontrol.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/process_infravision.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/process_lightning.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/process_lightning_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/process_lightningcontrol.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/process_nova.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/process_novaball_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/process_phasing.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/process_rage.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/process_redportal.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/process_resurrect.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/process_rhino.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/process_stonecurse.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/process_teleport.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/process_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/process_townportal.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/process_wall.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/process_weaponexplosion.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/processmissiles.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/processmissiles_test.go	CLEAN	CLEAN	CLEAN	markers:TestProcessMissiles_ClearsMissileTileFlagsBeforeThePass:empty-then@11
internal/diablo/missiles/rage_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/redportal_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/registry.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/registry_add_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/registry_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/resurrectbeam_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/rhino_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/rotation.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/rotation_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/slice4_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/stonecurse_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/teleport_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/teleport_view_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/townportal_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/trapdisarm_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/wall.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/missiles/wall_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/missiles/weaponexplosion_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/save/codec.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/save/codec_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/save/diablosha.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/save/diablosha_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/stores/enter.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/stores/esc.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/stores/fixtures_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/stores/gossip.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/stores/hold.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/stores/hold_test.go	CLEAN	CLEAN	CLEAN	markers:TestBodySlotMapping:empty-when@27
internal/diablo/stores/line.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/stores/machine.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/stores/machine_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/stores/menu.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/stores/premium.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/stores/premium_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/stores/price.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/stores/price_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/stores/scroll_repeat_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/stores/talkid.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/stores/talkid_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/stores/trade_enter.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/testsupport/corpus.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/testsupport/demo.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/testsupport/demo_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/testsupport/dun.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/testsupport/dun_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/txtdata/classes.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/txtdata/classes_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/txtdata/experience.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/txtdata/experience_fixture_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/txtdata/experience_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/txtdata/mana.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/txtdata/mana_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/txtdata/spelldata.go	CLEAN	CLEAN	CLEAN	-
internal/diablo/txtdata/spelldata_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/diablo/txtdata/testutil_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/engine/appfat/appfat.go	CLEAN	CLEAN	CLEAN	-
internal/engine/appfat/appfat_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/engine/assets/mpq.go	CLEAN	CLEAN	CLEAN	-
internal/engine/assets/mpq_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/engine/budgets_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/engine/config.go	CLEAN	CLEAN	CLEAN	-
internal/engine/config_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/engine/crawl.go	CLEAN	CLEAN	CLEAN	-
internal/engine/crawl_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/engine/demo/doc.go	CLEAN	CLEAN	CLEAN	-
internal/engine/demo/format.go	CLEAN	CLEAN	CLEAN	-
internal/engine/demo/format_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/engine/demo/playback.go	CLEAN	CLEAN	CLEAN	-
internal/engine/demo/playback_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/engine/doc.go	CLEAN	CLEAN	CLEAN	-
internal/engine/ini_fixes_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/engine/ini_import.go	CLEAN	CLEAN	CLEAN	-
internal/engine/ini_import_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/engine/path.go	CLEAN	CLEAN	CLEAN	-
internal/engine/path_bench_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/engine/path_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/engine/prof_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/engine/random.go	CLEAN	CLEAN	CLEAN	-
internal/engine/random_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/engine/roundtrip_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/engine/tick.go	CLEAN	CLEAN	CLEAN	-
internal/engine/txtdata/override.go	CLEAN	CLEAN	CLEAN	-
internal/engine/txtdata/override_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/engine/txtdata/parser.go	CLEAN	CLEAN	CLEAN	-
internal/engine/txtdata/parser_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/game/debug.go	CLEAN	CLEAN	CLEAN	-
internal/game/debug_cheats.go	CLEAN	CLEAN	CLEAN	-
internal/game/debug_flag.go	CLEAN	CLEAN	CLEAN	-
internal/game/debug_flag_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/game/debug_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/game/doc.go	CLEAN	CLEAN	CLEAN	-
internal/game/doom.go	CLEAN	CLEAN	CLEAN	-
internal/game/doom_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/game/game.go	CLEAN	CLEAN	CLEAN	-
internal/game/game_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/game/testsupport/grids.go	CLEAN	CLEAN	CLEAN	-
internal/game/testsupport/hash.go	CLEAN	CLEAN	CLEAN	-
internal/game/world/grid.go	CLEAN	CLEAN	CLEAN	-
internal/game/world/tileprops_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/gameplay/inventory.go	CLEAN	CLEAN	CLEAN	-
internal/gameplay/inventory_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/gameplay/itemgen.go	CLEAN	CLEAN	CLEAN	-
internal/gameplay/itemgen_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/gameplay/quests.go	CLEAN	CLEAN	CLEAN	-
internal/gameplay/quests_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/gameplay/stores/stores.go	CLEAN	CLEAN	CLEAN	-
internal/gameplay/stores/talkid_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/light/tables.go	CLEAN	CLEAN	CLEAN	-
internal/light/tables_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/persist/codec.go	CLEAN	CLEAN	CLEAN	-
internal/persist/codec_password_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/persist/codec_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/persist/cursor.go	CLEAN	CLEAN	CLEAN	-
internal/persist/demo_compare.go	CLEAN	CLEAN	CLEAN	-
internal/persist/demo_compare_detail_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/persist/demo_compare_saves.go	CLEAN	CLEAN	CLEAN	-
internal/persist/demo_compare_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/persist/doc.go	CLEAN	CLEAN	CLEAN	-
internal/persist/droppeditems.go	CLEAN	CLEAN	CLEAN	-
internal/persist/droppeditems_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/persist/exportsv_layout_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/persist/heroitems.go	CLEAN	CLEAN	CLEAN	-
internal/persist/heroitems_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/persist/hotkeys.go	CLEAN	CLEAN	CLEAN	-
internal/persist/hotkeys_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/persist/importdir.go	CLEAN	CLEAN	CLEAN	-
internal/persist/importdir_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/persist/importstash.go	CLEAN	CLEAN	CLEAN	-
internal/persist/importsv.go	CLEAN	CLEAN	CLEAN	-
internal/persist/itempack.go	CLEAN	CLEAN	CLEAN	-
internal/persist/itempack_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/persist/itemremap.go	CLEAN	CLEAN	CLEAN	-
internal/persist/itemremap_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/persist/itemsave.go	CLEAN	CLEAN	CLEAN	-
internal/persist/itemsave_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/persist/levels_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/persist/levelsave.go	CLEAN	CLEAN	CLEAN	-
internal/persist/levelsave_statehash_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/persist/levelsave_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/persist/levelseeds.go	CLEAN	CLEAN	CLEAN	-
internal/persist/levelseeds_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/persist/memmap_diff.go	CLEAN	CLEAN	CLEAN	-
internal/persist/memmap_diff_commands.go	CLEAN	CLEAN	CLEAN	-
internal/persist/memmap_diff_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/persist/memmap_diff_walk.go	CLEAN	CLEAN	CLEAN	-
internal/persist/migrate.go	CLEAN	CLEAN	CLEAN	-
internal/persist/migrate_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/persist/missilesave.go	CLEAN	CLEAN	CLEAN	-
internal/persist/missilesave_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/persist/monster_save.go	CLEAN	CLEAN	CLEAN	-
internal/persist/monster_save_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/persist/netpack.go	CLEAN	CLEAN	CLEAN	-
internal/persist/netpack_codec.go	CLEAN	CLEAN	CLEAN	-
internal/persist/netpack_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/persist/packitem.go	CLEAN	CLEAN	CLEAN	-
internal/persist/packitem_helpers.go	CLEAN	CLEAN	CLEAN	-
internal/persist/packitem_modes_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/persist/packitem_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/persist/packplayer.go	CLEAN	CLEAN	CLEAN	-
internal/persist/packplayer_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/persist/pfile.go	CLEAN	CLEAN	CLEAN	-
internal/persist/pfile_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/persist/player_netpack.go	CLEAN	CLEAN	CLEAN	-
internal/persist/playerpack.go	CLEAN	CLEAN	CLEAN	-
internal/persist/playerpack_decode.go	CLEAN	CLEAN	CLEAN	-
internal/persist/playerpack_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/persist/questsave.go	CLEAN	CLEAN	CLEAN	-
internal/persist/questsave_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/persist/savegame.go	CLEAN	CLEAN	CLEAN	-
internal/persist/savegame_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/persist/schema_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/persist/sqlc/db.go	CLEAN	CLEAN	CLEAN	-
internal/persist/sqlc/models.go	CLEAN	CLEAN	CLEAN	-
internal/persist/sqlc/store.sql.go	CLEAN	CLEAN	CLEAN	-
internal/persist/stash.go	CLEAN	CLEAN	CLEAN	-
internal/persist/stash_read.go	CLEAN	CLEAN	CLEAN	-
internal/persist/stash_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/persist/store_legacy.go	CLEAN	CLEAN	CLEAN	-
internal/persist/store_legacy_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/persist/store_sqlite.go	CLEAN	CLEAN	CLEAN	-
internal/persist/store_sqlite_atomic_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/persist/store_sqlite_map.go	CLEAN	CLEAN	CLEAN	-
internal/persist/store_sqlite_test.go	CLEAN	CLEAN	CLEAN	CLEAN
internal/testsupport/parity/assert.go	CLEAN	CLEAN	CLEAN	-
internal/testsupport/parity/compare.go	CLEAN	CLEAN	CLEAN	-
internal/testsupport/parity/fixture.go	CLEAN	CLEAN	CLEAN	-
internal/testsupport/parity/parity_test.go	CLEAN	CLEAN	CLEAN	CLEAN
```

## Appendix B — net LOC delta (base 41b257a .. tip)

Only two Go files changed, both tests; **no source file was edited in this
wave**, so there is no per-package source movement to report.

| path | added | deleted |
|---|---|---|
| internal/assets/mpq_smoke_test.go | 142 | 0 |
| internal/assets/trn_test.go | 4 | 5 |
| docs/requirements.md | 14 | 0 |
| docs/pow/asset-verification-wave.md | 767 | 0 |
| .beans/*.md (12 files) | 442 | 30 |
| **total** | **1369** | **35** |

## Sign-off

Wave closed at this commit; gates re-run at the tip (build rc0, `gofmt -l`
empty, vet rc0, lint 0 issues, `go test ./...` 37 packages ok / 0 failures).

asset-verification-wave fingerprint: 5wsb verified against DIABDAT+hellfire+devilutionx.mpq on 2026-11-03

## Slice 2 — closers

### Marker repair (devilGo-klu5, devilGo-ocl1)

The five findings from Appendix A were repaired by giving each phase a real
statement or by stopping prose from looking like a marker:

| file:line (before) | defect | repair |
|---|---|---|
| `internal/diablo/missiles/add_slice2_test.go:159` | duplicated empty `// when` | deleted; the real `// when` keeps the subtest build |
| `internal/diablo/missiles/add_slice2_test.go:191` | same in `TestAddHolyBolt` | deleted |
| `internal/diablo/stores/hold_test.go:206` | `// when` with no statement | `_, ok := stores.BodySlot(0)` now sits under `when`, the check under `then` |
| `internal/diablo/inventory/grid_test.go:327` | second `then` line started with "then" | one `then` phase, prose reflowed |
| `internal/diablo/missiles/processmissiles_test.go:81` | `// then:` prose continued with "then" | prose reworded, no marker-looking continuation |
| `internal/assets/trn_test.go:289` | descriptive comment ran into the `// when` marker | blank line separates prose from marker |

Re-scan with the same strict scanner as Appendix A (every marker-shaped line
counts, a marker with no non-comment statement below it is a finding):

```
test_funcs=996 flagged=0
```

Before this slice: 996 test functions, 4 files / 5 findings.

### Docstring trim (devilGo-ocl1)

`scripts/fixture_fixtures.py` module docstring: 23 non-blank lines -> 7 (what is
stored, that only the sha256 manifest is committed, one line per mode). The
claims match observed runs: `python3 scripts/fixture_fixtures.py verify` with no
`test/fixtures/generated/` tree prints the skip line and exits 0.

Recorded while verifying: `refresh` crashes with `FileNotFoundError` when
`tools/probe/out/` exists but `tools/probe/demo_0.dmo` does not, instead of
refusing like its own `PROBE_OUT` guard — filed as `devilGo-z3io` (draft).

The bean's "52 test files" inventory is stale: `go test ./... -v` reports 247
test files / 996 test functions today. The per-file table remains Appendix A
(483 Go files, now 483 clean).

### Gate table (branch tip)

| gate | command | result |
|---|---|---|
| build | `go build ./...` | rc0 |
| format | `gofmt -l internal/ cmd/` | empty |
| vet | `go vet ./...` | rc0 |
| lint | `scripts/lint.sh ./...` | `0 issues.` |
| tests | `go test ./... -count=1` | 37 packages ok, 0 FAIL |
| markers | strict re-scan | 996 functions, 0 flagged |

### Touched-file verdict (vs base f289c67)

| file | vet | lint | gofmt | markers | LOC |
|---|---|---|---|---|---|
| internal/diablo/missiles/add_slice2_test.go | CLEAN | CLEAN | CLEAN | CLEAN | +0/−4 |
| internal/diablo/stores/hold_test.go | CLEAN | CLEAN | CLEAN | CLEAN | +3/−1 |
| internal/diablo/inventory/grid_test.go | CLEAN | CLEAN | CLEAN | CLEAN | +2/−2 |
| internal/diablo/missiles/processmissiles_test.go | CLEAN | CLEAN | CLEAN | CLEAN | +2/−2 |
| internal/assets/trn_test.go | CLEAN | CLEAN | CLEAN | CLEAN | +1/−0 |
| scripts/fixture_fixtures.py | n/a | n/a | n/a | n/a | +5/−22 |
| **Go total** | | | | | **+8/−9** |
