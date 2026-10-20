# Parity fixture production

`tools/probe/dvg_probe` is the C++ oracle for parity fixtures. It is built
against devilutionX commit `ca91d778fba92a1f680afb1fb2cdf09337ff1b42` using
the patch in `tools/probe/patches/`.

| Artifact | Producer | Consumer |
|---|---|---|
| `rng/<seed>.bin` | `dvg_probe rng` | devilGo-m2x7 |
| `lighttables/*_tables.bin` | `dvg_probe lighttables` | lighting parity tests |
| `lighttables/falloffs.bin`, `cone.bin` | `dvg_probe lighttables` | devilGo-kx0l |
| `dun/*.dun`, `*.meta` | `dvg_probe dun` | devilGo-xcdz |
| `frames/*.idx.bin`, `*.pal.bin` | `dvg_probe frames` | devilGo-7aul |
| `save/spawn_0.sv` | `dvg_probe save` | devilGo-6xu2 / devilGo-oqho |
| `manifest.txt` | `dvg_probe manifest` | devilGo-1spo |

The producer writes only to its requested output directory; it never writes to
`test/fixtures/`. `tools/probe/out/` is the conventional ignored output
location. The pinned seed list and demo input are committed under
`tools/probe/`.

## Format notes

- RNG files contain 1000 little-endian `uint32_t` values per seed.
- Light tables are the native contiguous C++ arrays: 16×256 table bytes,
  16×128 falloff bytes, and 8×8×16×16 cone bytes.
- Dungeon files contain two little-endian `uint16_t` dimensions, the 40×40
  dungeon tile array, then the 80×80 transparency window. The `.meta` sidecar
  contains the little-endian view position followed by one level-type byte.
- Frame index files contain two little-endian dimensions followed by row-major
  8-bit pixels. Palette files contain 256 `SDL_Color` values (RGBA).
- Saves are vanilla spawn `.sv` archives. New-player `PlayerPack` timestamp
  fields `dwLowDateTime` and `dwHighDateTime` from `pack.h:34-35` are zeroed by
  the engine's `PackPlayer` initialization.

## Reproduction and determinism

Apply the patch to a clean checkout of the pinned devilutionX commit, configure
with `BUILD_TESTING=ON` and `-DDVG_PROBE_DIR=$PWD/tools/probe`, then build the
`dvg_probe` target. Run each subcommand twice from clean output/run directories
and compare `sha256sum` output or the generated manifest. MPQ assets are local
copyrighted inputs and are not committed.

## Committed fixture tree

Storage decision (devilGo-1spo, recorded): goldens are committed under
`test/fixtures/`; CI verifies hashes only (`just fixture-verify`) and never
generates. The produced copy lives in `test/fixtures/generated/` — same
relative layout as the probe output plus the committed inputs
(`demo_0.dmo`, `seeds.txt`) — with `manifest.txt` as its hash manifest.
`test/fixtures/static/` holds hand-authored fixtures (currently empty,
`.gitkeep`).

- `just fixture-refresh` — copies the probe run from `tools/probe/out/` into
  `test/fixtures/generated/` (re-runs the probe only when the output
directory is absent), prints the drift report (added/removed/changed paths
vs the committed manifest) and regenerates the inventory below.
- `just fixture-verify` — recomputes every manifest row's sha256 against
  the committed tree; exits nonzero on any mismatch or missing file.

## Committed artifact inventory

### StateHash — canonical state hashing

Cross-run/game comparisons hash canonical state with `StateHash` (FNV-1a 64,
`internal/testsupport/hash.go`): implement `HashableState` by returning the
canonical fields (identity, tile position, vitals, mode) in a fixed order —
field order is part of the hash. Never include caches, derived sprite
geometry, map-iteration-ordered collections or timers. The hash is
deterministic across runs and process restarts; the suite proves it with a
child-process re-exec (`demo_test.go` TestStateHashStableAcrossProcesses).

### Demo tapes

Recorded `.dmo` tapes are the scripted-run input for the frames/parity
fixtures (`demo_0.dmo` under `test/fixtures/generated/`). Loader:
`internal/testsupport/demo.go` (`LoadDemoTape`) — parses the fixed header
(version 3, recorded resolution, save slot) and exposes the ordered
`DemoStep` stream. The playback loop that drives the engine tick belongs to
its consumer bean; the SDL event mapping belongs to the input bean (upstream
reference: `demo::InitPlayBack` + `RunGameLoop`).

### End-to-end regeneration story

1. Build the C++ oracle against the pinned commit (tools/probe/README.md:
   apply patch, CMake with `BUILD_TESTING=ON` and `DVG_PROBE_DIR`, build
   `dvg_probe`).
2. Run each subcommand (`rng`, `lighttables`, `dun`, `frames`, `save`) from
   clean output/run directories, then `dvg_probe manifest` — two full runs
   must produce sha256-identical manifests (the determinism gate).
3. `just fixture-refresh` — copies the artifacts into
   `test/fixtures/generated/`, prints the drift report (added/removed/
   changed paths vs the committed manifest) and regenerates the inventory
   below.
4. `just fixture-verify` — hash-only check of the committed tree (CI path).
5. Golden tests assert with `CompareGrids` (per-cell diffs) and
   `StateHash` (whole-state fingerprint); drift means the port or the corpus
   changed — upstream algorithms are deterministic.

### Regenerated inventory

Pinned devilutionX commit: `ca91d778fba92a1f680afb1fb2cdf09337ff1b42` (probe run: `tools/probe/out/` per tools/probe/README.md).

| Class | Files | Contents |
|---|---|---|
| `dun/` | 256 | seeded DRLG dumps: `.dun` + `.meta` per (level, entry, seed) |
| `frames/` | 10 | timedemo frames: `.idx.bin` index buffer + `.pal.bin` palette |
| `lighttables/` | 6 | MakeLightTable() outputs: 4 leveltype tables, falloffs, cone |
| `rng/` | 8 | 1000 raw `GenerateRandomNumber()` outputs per seed |
| `save/` | 1 | vanilla spawn `.sv` hero archive |
| root files | 3 | `demo_0.dmo`, `manifest.txt`, `seeds.txt` |
