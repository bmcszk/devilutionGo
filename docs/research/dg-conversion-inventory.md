# Conversion-surface inventory (dg-conv s1)

> **Snapshot status.** This page maps the tree as it stood before ADR-019
> (`docs/adrs/adr-019.md`) landed. Every `dgo-assets` row below is history: the
> command is gone, and its jobs live in `devilutionGo-migrate assets`.

Scope: every CLI under `cmd/` plus the probe tooling under `tools/probe/`, as
they exist on master at `da85abe`. This is a map of what runs today, not a plan.
Citations are `file:line` in this repo.

## Why this inventory exists

ADR-014 promised one conversion program (`docs/adrs/adr-014.md:12`). Five
`cmd/` binaries and four `tools/probe/` artifacts exist instead. Some belong to
the user pipeline, some are test instrumentation. This page separates them.

## Inventory

| Program | Class | Purpose | Interface | Reads | Writes | Gate | ADR |
|---|---|---|---|---|---|---|---|
| `cmd/devilutionGo` | runtime (one binary) | windowed client entry and headless CI/parity mode | `--debug` (`main.go:78`), `--headless`, `--ticks N` (`main.go:63-64`) | nothing on disk yet | log line only (`main.go:80`, headless at `headless.go:26`) | `main_test.go` | 001, 007, 017, 020 |
| `cmd/devilutionGo-migrate` | user tool | MPQ to assets dir, CSV tables, YAML list | `assets` \| `saves` \| `settings` (`main.go:47-57`); `--mpq` repeatable, `--out` (default `./assets-out`), `--dry-run` (`assets.go:100-108`); env `DG_ASSET_ARCHIVES` (`assets.go:113`) | MPQ archives; embedded name inventories (`internal/assets/names.go:13`) | `<out>/<archive-stem>/<inner path>`, `manifest.yaml` (`manifest.go:22,:138`), CSV tables, `.migrate-state.json` journal | 6 test files; `retail_test.go` skips without env | 005, 011, 013, 014 |
| `cmd/dgo-assets` | tool (graphics inventory) | manifest check/list, CEL/CL2/PCX to CLX | `check --manifest --root` (`cli.go:104-105`), `list --manifest --markdown` (`cli.go:137-141`), `convert --format/--in/--out/--widths/--frames/--transparent` (`cli.go:174-180`) | its own manifest schema (`internal/assets/manifest/manifest.go:55-93`) and the `--in` file | CLX bytes (`cli.go:246`), stdout tables | `cli_test.go` 6 tests | 013 + bean devilGo-l5fp |
| `tools/gamedatagoldens` | test infra | regenerate full-table golden JSON | no flags; fixed target list (`main.go:27-36`) | committed TSVs under `internal/testdata/gamedata/` | `internal/testdata/gamedata/goldens/*.json` (`main.go:23`) | `just refresh-gamedata-goldens` (`justfile:57-66`); consumers are the diablo gamedata tests | 010 |
| `tools/probe/dvg_probe.cpp` | test infra (C++, dump only) | produce parity fixtures from pinned devilutionX | `rng`, `lighttables`, `dun`, `drlg`, `stages`, `frames`, `save`, `manifest` (`README.md:52-61`) | pinned checkout `ca91d778f` + retail MPQs | `tools/probe/out/**` and `manifest.txt` (gitignored except the manifest) | `scripts/fixture_fixtures.py verify/refresh` (`justfile:45-53`) | 002, 010 |
| `tools/probe/patches/*.patch` | test infra | one build hook + two dump-only hooks | applied by hand (`README.md:23-35`) | devilutionX CMake and generators | nothing in this repo | same fixture gate | 010 |
| `tools/probe/dunconv` | test infra | `.dun`+`.meta` to committed CSV corpus | positional `<artifact-dir> <corpus-dir>` (`main.go:25`) | probe `drlg` artifacts | `internal/testdata/drlg/*.csv` via `testsupport.EncodeDunCSV` (`main.go:12`) | `internal/diablo/levels` golden tests | 010 |
| `tools/probe/affix_probe` | test infra | dump `SelectAffix` pool from pinned C++ | compiled harness (`affix_probe/README.md:1-5`) | devilutionX translation units, read-only | `internal/testdata/gamedata/items/affix_pool_probe.tsv` | `TestEligibleAffixesMatchesCppProbe` | 010 |

## Classification

**User pipeline (conversion of retail data):** `cmd/devilutionGo-migrate`
only. It is the program a user runs, it is the one ADR-014 names, and it is the
one that owns MPQ decoding, CSV emission and the YAML list.

**Runtime programs (not conversion):** `cmd/devilutionGo`, one binary with a
windowed and a headless mode (ADR-020). It reads game state and assets, never
archives.

**Test instrumentation (never shipped, never user-facing):** `dvg_probe.cpp` +
its three patches, `dunconv`, `affix_probe`, `gamedatagoldens`. Every one of
them exists to produce or refresh a committed fixture. The zero-dependency law
forbids any of them running in `go test` or CI (`AGENTS.md` section 6); they run
offline, once, against the pinned checkout.

**The one misfit:** `cmd/dgo-assets`. It is neither user pipeline nor pure test
infra. `check` and `list` are build-time gates a user or CI would run, and
`convert` is the CEL/CL2/PCX to CLX step the retail bundle needs. It ships as a
second top-level command with its own copy of the "asset list" concept.

## Current user pipeline for MPQ to assets (measured)

1. Install the retail archives and point the tool at them, either with
   `DG_ASSET_ARCHIVES` or repeated `--mpq` (`cmd/devilutionGo-migrate/assets.go:20`,
   `:100-108`). `resolveArchives` reads the env path list when no flag is given
   (`assets.go:129-140`).
2. Run `devilutionGo-migrate assets --out <dir>`. The default output is
   `./assets-out` (`assets.go:22`), which is gitignored (`.gitignore:14`).
3. The run writes the extracted tree, `manifest.yaml`, the CSV tables and the
   journal (`assets.go:260-263`). CEL and CL2 are copied byte for byte, not
   converted (`extract.go:248-252`). PCX is decoded only to report geometry
   (`extract.go:312`).
4. Nothing else runs. There is no `just` recipe, no `go:generate` and no test
   that invokes `dgo-assets convert` or `dgo-assets check`: the only mentions
   outside `cmd/dgo-assets` are the bean text and the manifest package comment
   (`internal/assets/manifest/manifest.go:7`).

### The ADR-005 embed path is unlanded

`assets.go:21` calls `./assets-out` "the assets dir the build embeds
(ADR-005)", but no embed of it exists. The repo's only `go:embed` directives are
`internal/assets/names/*.txt` (`internal/assets/names.go:13`) and the SQLite
schemas (`internal/persist/store_sqlite_map.go:20,:25,:30`). `assets/` holds a
`.gitkeep` and nothing else. So ADR-005 step "extract, then embed" currently has
its extract leg only.

## The actual split: two YAML files named `manifest.yaml`

Both writers call their output `manifest.yaml`, and the shapes do not match.

* Migration list (`cmd/devilutionGo-migrate/manifest.go:85-92`): `tool`,
  `version`, `archives[]`, `files[]` (path, archive, source, family, format,
  frames, rows, bytes, sha256), `summary`. Writer name
  `devilutionGo-migrate` (`manifest.go:25`).
* Graphics inventory (`internal/assets/manifest/manifest.go:91-96`): `tool`,
  `version`, `entries[]` (family, path, format, bundled, frames, width,
  width_source, widths, loader), `summary`. Writer name `dgo-assets`
  (`manifest.go:28`), loader references required to match
  `Source/...cpp:NNN` (`manifest.go:31`).

Consequences, all verified by reading the readers:

* `dgo-assets check` defaults to `assets/manifest.yaml` (`cli.go:104`) and calls
  `manifest.Read` + `Validate` (`cli.go:88-96`). Run against a file the
  migration tool wrote, `Tool` and `Version` pass, `Entries` decodes as empty,
  so `Missing` returns nothing and `check` exits 0 printing "0 entries in 0
  families" (`cli.go:118-131`). The coverage gate silently passes on the wrong
  file.
* `Validate` refuses a missing `tool` and a wrong `version` only
  (`internal/assets/manifest/manifest.go:126-131`). It does not refuse an empty
  entry list.
* The migration list is per extracted file (3134 files in the retail PoW). The
  graphics inventory is per loader call site, so its entries are a subset with
  required `loader` and `width_source` data that extraction cannot produce.

So the split is not just two commands: it is two incompatible readings of one
ADR-013 YAML slot, with a fail-open gate between them.

## Gates per program (what would break if a program moved)

| Test / gate | Pins |
|---|---|
| `cmd/devilutionGo-migrate/assets_test.go` (9 tests) | tree layout, `manifest.yaml`, CSV tables, journal resume, dry-run, decode-failure exit 1, unaddressable-archive skip, traversal rejection |
| `cmd/devilutionGo-migrate/cli_test.go` | process exit codes 0/1/2 through the real dispatch in a child process |
| `cmd/devilutionGo-migrate/{extract,journal,manifest}_test.go` | per-record decode, journal fingerprint, summary hash |
| `cmd/devilutionGo-migrate/retail_test.go` | real retail run; skips when `DG_ASSET_ARCHIVES` is unset |
| `cmd/dgo-assets/internal/cli/cli_test.go:17,46,72,91,122,151` | check missing-entry loader line, check pass, invalid manifest, list output, convert output equals the converter goldens, bad invocations |
| `internal/assets/{convert,convert_pcx,clx_encode}_test.go` | CEL/CL2/PCX to CLX byte parity vs committed C++ oracle dumps and committed `.clx` re-encodes |
| `internal/assets/manifest/manifest_test.go` (6 tests) | entry invariants, `Missing`, YAML round trip, `Finalize` |
| `internal/diablo/levels/*_parity_test.go` | the committed `internal/testdata/drlg` corpus `dunconv` produces |
| `internal/diablo/gamedata` goldens | the JSON `tools/gamedatagoldens` produces |

Fixtures at stake: `internal/testdata/gfx/{cel,cl2,pcx,clx,producer}` (16 derived
inputs, per-frame widths, oracle goldens, producer recipe) and
`internal/testdata/drlg/*.csv`.

## Facts the ADR-019 draft must respect

1. Conversion code already lives in `internal/assets` (`cel_to_clx.go`,
   `cl2_to_clx.go`, `pcx_to_clx.go`, `clx_encode.go`), not in `cmd/`. Moving the
   `convert` command changes no conversion logic.
2. Manifest model already lives in `internal/assets/manifest`, engine layer, no
   Diablo imports (`manifest.go:5-7`). Family enumeration is the only part that
   needs `internal/diablo`, and it was never landed.
3. `dgo-assets` is a developer tool, not a release pipeline step (bean
   devilGo-l5fp, resolved open question 2026-10-16).
4. ADR-003 already shows the house pattern for retiring a command name: the
   status line records the move and the body stays unchanged
   (`docs/adrs/adr-003.md:4`).
5. No caller invokes `dgo-assets` today (grep outside `cmd/dgo-assets` returns
   only the bean file and `manifest.go:7`), so a rename has no downstream
   breakage inside the repo.

## Sources

Repo files cited above, plus `docs/adrs/adr-003.md`, `adr-005.md`,
`adr-013.md`, `adr-014.md`, `adr-017.md`, `AGENTS.md` sections 1b and 6, beans
`devilGo-o8e7`, `devilGo-v7ft`, `devilGo-l5fp`, `justfile`, and
`tools/probe/README.md`.
