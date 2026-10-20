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

## Fixture storage (untracked goldens)

Storage decision (devilGo-1spo, superseded by the review-prep binary purge):
probe-produced goldens are **not committed**. The authoritative probe run
lives in `tools/probe/out/` and a refreshed local copy lives in
`test/fixtures/generated/` (same relative layout plus the probe inputs
`demo_0.dmo`, `seeds.txt`) — both gitignored and regenerable. Only the TEXT
manifest is committed, at `test/fixtures/manifest.txt`: the sha256 of every
generated artifact at refresh time. Parity tests that consume the generated
tree skip (with a clear message) when the artifacts are absent, so CI
without probe artifacts runs the suite short of the generated golden tests.
`test/fixtures/static/` holds hand-authored fixtures (currently empty,
`.gitkeep`).

- `just fixture-refresh` — copies the probe run from `tools/probe/out/`
  (never generates itself; fails with reproduction steps when the output
directory is absent), refreshes the committed manifest at
`test/fixtures/manifest.txt`, prints the drift report (added/removed/
changed paths) and regenerates the inventory below. Binaries stay out of
`git add` — they are gitignored.
- `just fixture-verify` — recomputes every manifest row's sha256 of the
  committed manifest against the local `test/fixtures/generated/` tree;
  skips with a clear message when no local artifacts exist, exits nonzero
  on any mismatch or missing file.

## Committed artifact inventory

### StateHash — canonical state hashing

Cross-run/game comparisons hash canonical state with `StateHash` (FNV-1a 64,
`internal/game/testsupport/hash.go`): implement `HashableState` by returning the
canonical fields (identity, tile position, vitals, mode) in a fixed order —
field order is part of the hash. Never include caches, derived sprite
geometry, map-iteration-ordered collections or timers. The hash is
deterministic across runs and process restarts; the suite proves it with a
child-process re-exec (`demo_test.go` TestStateHashStableAcrossProcesses).

### Demo tapes

Recorded `.dmo` tapes are the scripted-run input for the frames/parity
fixtures (`demo_0.dmo` — probe input under `tools/probe/`, refreshed into
the local `test/fixtures/generated/`). Loader:
`internal/diablo/testsupport/demo.go` (`LoadDemoTape`) — parses the fixed header
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
3. `just fixture-refresh` — copies the artifacts into the local, untracked
   `test/fixtures/generated/` and refreshes the committed manifest at
   `test/fixtures/manifest.txt`; prints the drift report (added/removed/
   changed paths) and regenerates the inventory below.
4. `just fixture-verify` — hash-only check of the local generated tree
   against the committed manifest (skip-aware; see above).
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
| root files | 2 | `demo_0.dmo`, `seeds.txt` |
