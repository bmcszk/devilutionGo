# Agents.md — devilutionGo Architecture & Agent Guidelines

> **Purpose**: Single source of truth for any agent (human or AI) working on this codebase.  
> **Audience**: Contributors, AI coding agents, reviewers.  
> **Scope**: Project structure, coding conventions, domain boundaries, and the *why* behind every decision.

---

## What

**devilutionGo** is a **1:1, byte‑exact port** of **devilutionX** (the modern C++ Diablo 1 / Hellfire engine) to **Go 1.27 + Ebitengine v2 + SQLite**.

* **Fidelity mandate** – identical algorithms, constants, RNG, data values, save‑file format, network packets, pixel‑perfect rendering.  
* **TDD mandate** – every implementation task starts with a failing acceptance test (golden fixture, row‑count, pixel‑diff, or packet hex‑dump).  
* **Reusable engine** – the codebase is split into three layers so the *engine* can be reused for other isometric ARPGs.

| Layer | Package root | Responsibility | Example |
|-------|--------------|----------------|---------|
| **Engine (project‑agnostic)** | `internal/engine/` | Core game loop, ECS, rendering primitives, asset pipeline, audio, input, networking, save/load, configuration. No Diablo‑specific logic. | `internal/engine/assets`, `internal/engine/demo`, `internal/engine/txtdata` |
| **Common / shared game systems** | `internal/game/` | Systems that are generic but tuned for Diablo‑style games (isometric grid, line‑of‑sight, path‑finding, inventory UI, stats, combat formulas). | `internal/game/world`, `internal/game/testsupport` |
| **Diablo (project‑specific)** | `internal/diablo/` | All data definitions, level formats, quest logic, monster AI, item tables, spells, town services, MPQ asset loading, save‑file schema. | `internal/diablo/gamedata`, `internal/diablo/txtdata`, `internal/diablo/combat`, `internal/diablo/testsupport` |

> **Rule** – Code in `internal/engine/` must **never** import `internal/diablo/`. `internal/game/` may import `internal/engine/` only. `internal/diablo/` may import both.

---

## How

### 1. Development Workflow
| Step | Tool / Command | Notes |
|------|----------------|-------|
| **Bootstrap** | `just bootstrap` | `go mod tidy`, generate code, download test fixtures |
| **Run** | `just run` | Starts the game (windowed) |
| **Headless / CI** | `just test-headless` | Builds with `-tags=headless`, runs unit + parity tests |
| **Test** | `just test` | `go test ./... -count=1` (includes golden‑fixture tests) |
| **Lint** | `just lint` | `golangci-lint run ./...` |
| **Format** | `just fmt` | `gofmt -s -w .` |
| **Generate** | `just generate` | `go generate ./...` (shaders, asset manifests, etc.) |

All commands are defined in `justfile` at repo root.

### 1b. Binding rules digest (skills are canonical — this is the load-bearing subset)

Workers load AGENTS.md automatically (the `-nc` flag that disabled it is retired); the digest below is the load-bearing subset — full texts in the named skills. Full texts: `~/.agents/skills/coding-rules/SKILL.md`, `~/.agents/skills/go-dev/SKILL.md`, `~/.agents/skills/golang-testing/SKILL.md`. If a digest line conflicts with a skill, the skill wins — report the conflict.

| Rule | Requirement (binding) |
|-----|------------------------|
| Intent functions | Top-level functions show intent/business sequence; helpers stay private. |
| TDD | Failing test FIRST (golden/parity/unit); PoW includes the RED run output. Never weaken an assertion to pass. |
| Table-driven | `t.Run` sub-tests for multi-case logic; no copy-pasted test bodies. |
| No bool-flag params | Split named variants (`ProcessAnimation`/`ProcessAnimationReverse`), never `func f(x bool)`. |
| Minimal change | Smallest change that fully solves the bean; no drive-by refactors beyond the contract. |
| Comments | On all exported identifiers; cite upstream `file:line` for every ported constant/branch. |
| Duplication | Extract on 3rd occurrence; ported 1:1 C++ blocks get an exemption comment. |
| Gates before done | `gofmt -l` empty, `go vet` rc0, `golangci-lint` 0 issues, `go test ./internal/...` all ok — paste tails in the report. |
| Honest work only | No fabricated output, no skipped gates, no "should pass" claims; report failures plainly. |
| Scope jail | Edit only files the contract names; no `.beans/` writes; no git state changes. |

### 2. Project Layout (DDD‑aligned)

```
devilutionGo/
├─ cmd/
│  ├─ devilutionGo/          # main entry (GUI)
│  └─ devilutionGo-headless/ # headless binary for CI
├─ internal/
│  ├─ engine/                # <-- reusable engine
│  │  ├─ ecs/                # donburi wrapper, component/query helpers
│  │  ├─ render/             # index framebuffer, palette shader, batching
│  │  ├─ audio/              # resound wrapper, music streaming, SFX pool
│  │  ├─ input/              # action‑map, gamepad/keyboard
│  │  ├─ asset/              # MPQ reader, sprite/animation loaders, cache
│  │  ├─ net/                # TCP lockstep, delta compression
│  │  ├─ save/               # SQLite schema, .sv byte‑compat layer
│  │  ├─ config/             # options, i18n (gotext)
│  │  └─ util/               # RNG, math, logging
│  ├─ game/                  # <-- shared Diablo‑style systems
│  │  ├─ grid/               # isometric grid, LOS, A* (s0rg/grid + paths)
│  │  ├─ camera/             # kamera wrapper
│  │  ├─ ui/                 # ebitenui widgets (inventory, belts, panels)
│  │  ├─ combat/             # damage, resist, hit‑calc formulas
│  │  └─ inventory/          # grid‑based containers, belt, stash
│  └─ diablo/                # <-- Diablo‑specific
│     ├─ levels/             # DRLG, town, dungeon tile data
│     ├─ monsters/           # monstdat, AI, light sources
│     ├─ items/              # itemdat, affixes, generation
│     ├─ spells/             # misdat, missile logic
│     ├─ quests/             # quest FSM, towners, rewards
│     ├─ assets/             # MPQ → loose asset pipeline, CLX/CEL/PCX decoders
│     ├─ save/               # .sv schema, hero file, game state
│     └─ net/                # CMD_* packet (de)serialization
├─ assets/                   # build‑time only (source MPQs, TRNs, fonts)
├─ shaders/                  # .kage files (palette, light, TRN)
├─ test/
│  ├─ fixtures/              # golden dumps (dLight.dmp, falloffs.bin, etc.)
│  └─ parity/                # pixel‑diff, packet‑diff, save‑roundtrip tests
├─ docs/                     # requirements.md, lighting.md, sound.md, go‑libraries.md
├─ justfile
├─ go.mod / go.sum
> NOTE: the tree above is the PLANNED layout. Currently LANDED packages:
> engine/{crawl,path,random,tick,assets,demo,txtdata}, game/{game loop,world,testsupport},
> diablo/{combat,gamedata,txtdata,testsupport,missiles,enums,save}, cmd/{devilutionGo,devilutionGo-headless}.

└─ AGENTS.md                 # <-- you are here
```

### 3. Coding Conventions (enforced by `golangci-lint`)

| Rule | Reference |
|------|-----------|
| **Line length ≤ 120** | `golang-code-style` |
| **Error handling** – wrap with `fmt.Errorf("…: %w", err)` | `golang-error-handling` |
| **Table‑driven tests** – `testing.T.Run` sub‑tests | `golang-testing` |
| **No global mutable state** – use dependency injection via constructors | `golang-concurrency` |
| **Context‑aware cancellation** in all long‑running goroutines | `golang-concurrency` |
| **Benchmarks** for hot paths (render loop, path‑find, RNG) | `golang-performance` |
| **Documentation comments** on all exported types/functions | `coding-rules` |

### 4. Fidelity & Parity Gates (CI)

| Gate | Implementation | Artifact |
|------|----------------|----------|
| **Light tables** | `TestLightTablesGolden` – compare 16×256 bytes | `test/fixtures/light_render_benchmark/falloffs.bin` |
| **Vision rays** | `TestVisionRaysGolden` – 23×15 int8 array | `test/fixtures/vision_rays.bin` |
| **Dungeon render** | Pixel‑diff vs C++ headless dumper (seed dungeons) | `test/parity/render_*.png` |
| **Save/Load** | Round‑trip `.sv` → SQLite → `.sv` byte‑identical | `test/parity/save_roundtrip` |
| **Network** | Hex‑dump of CMD_* packets vs C++ capture | `test/fixtures/net/*.bin` |
| **Audio math** | Volume/pan log‑linear tables 100‑point sweep | `test/fixtures/audio_math.json` |

All gates are **blocking** in GitHub Actions (`just ci`).

### 5. Asset Pipeline (Build‑time only)

1. `assets/extract.sh` – uses `suprsokr/go-mpq` + custom PKWARE/Huffman/ADPCM to dump `diabdat.mpq` + `hellfire.mpq` + `devilutionx.mpq` → `assets/loose/`.
2. `assets/convert/` – Go tooling (`aseprite` lib) rewrites CLX/CEL/PCX → engine‑ready atlases (PNG + JSON meta).
3. `assets/manifest.json` – generated, checked‑in, consumed by `internal/engine/asset` at runtime.

> **No runtime MPQ reading** – keeps the engine pure Go, avoids CGO, guarantees deterministic loads.

### 6. Binding Program Decisions

| Decision | Rule |
|----------|------|
| ASSET vs CODE split | CODE = 100% Go rewrite — no C++ source, header, or compiled artifact is a dependency at build, test, CI, or runtime. The `~/github/bmcszk/devilutionX` checkout is a READ-ONLY research reference and offline fixture producer ONLY. ASSETS = identical to the original — same retail MPQ assets, extracted at build, loaded unmodified through Go decoders ported 1:1. |
| Pure Go (HARD) | Every dependency pure Go / no CGO: modernc.org/sqlite, gopher-lua, suprsokr/go-mpq. When a C library would be easier, pre-extract assets at build time or port the algorithm — never CGO. |
| Zero-dependency parity tests (HARD) | No Go test may read the C++ tree, `DG_DEVILUTIONX`, MPQs, or the network; no AC assumes an upstream checkout. Fixture pipeline: C++ probe runs ONCE offline against a pinned upstream commit → parsed-loader dump → COMMITTED TEXT goldens under `internal/testdata/` (TSV rows verbatim, grids CSV, binary dumps xxd — no raw binaries; manifest rows carry sha256s) → Go tests load committed bytes and FAIL (never skip) when a fixture is missing. |
| Saves (SQLite) | SQLite-only live format + one-time `.sv` importer CLI (no dual-format live writes). Per-hero DB files (one `.db` per hero, mirroring `single_N.sv`) plus one global `.db` for shared stash/metagame; `user_version` migration chain. SQLite data access via sqlc. |
| Settings | NOT in SQLite — 1:1 `diablo.ini` INI file (port `Source/utils/ini.cpp` codec; same SDL pref path + portable.ini; byte-interchangeable with C++ builds). |
| MPQ | extract-at-build v1 (`just assets`); runtime MPQ deferred, not cut. |
| Lua | deferred post-1.0, gopher-lua (pure Go); stub hook points in engine now. |
| Dependency authority | `docs/go-libraries.md` is the authoritative dependency list with must-port-from-C++ gaps (PKWARE/Huffman/ADPCM, CLX/CEL decoders, DRLG, RNG, MPQ encryption). Worker substitutions are defects; lib changes are program decisions. |
| Anti-overengineering guard | A binding mechanism that fires naturally (fail-on-missing fixture tests, ledger ACs) beats adding a meta-linter/CI grep for the same class; add machine checks only when no natural gate catches the violation. |

New program-level decisions are recorded in the milestone bean `## DECIDED` section AND here in the same session; later sessions read these, not chat history.

### 7. Porting & Data Discipline

- **Classify upstream symbols by call graph before porting a module:** grep each symbol's callers; port the game-called surface only, cut the rest as scaffolding — parity needs identical OUTPUTS, not the C++ API surface. Check both directions (a symbol needed only by a dumper/jump still counts as game-called).
- **Grafting onto C++ globals:** a pipeline reading a mutable global gets an injected sink struct the tests script — never a Go global. Entity-parameterized pipelines get an explicit-inputs Ref + injectable RNG so roll ORDER is testable before entity structs exist.
- **Upstream UB: document, don't port silently** — out-of-bounds reads unreachable with real inputs get a documented deviation note in the bean PoW; an unrecorded behavior deviation is a defect.
- **Corpus measurement:** row counts with `awk 'END{print NR}'`, never `wc -l` (trailing-newline undercount); assert `len(rows)` from the committed fixture at runtime, never a checked-in literal. Compare deltas BY ID SET, never by position; regenerate id lists from `diff`/`sort -u` output, never from memory. With a header row, file line N = data row N−2 (0-based) — recompute per lookup.
- **Inspect actual column data before typing the struct** — numeric-sounding columns are often enum-name strings (`sfxnr`=`Cain1`, `ai`=`Zombie`, `treasure`=`Uniq(CLEAVER)`); tolerate both `""` and the literal sentinel.
- **Pin tests to functionality, not scaffolding:** golden streams and bound tests earn their place; statistical proofs of C++ internals are test bloat. A test file 2x the implementation it pins is a trim candidate.
- **Acceptance check for a ported slice:** every ported function maps to at least one test.
- **Renderer bindings (G10/G11):** 640×480 is the WORLD LOGICAL SURFACE — `Layout()` returns the window size and the world surface is integer-scaled to fit; parity goldens at 1x AND 2x window. FilterNearest, index framebuffer + 16 palette-row textures + Kage palette-lookup shader (not ColorScale), atlas ≤4096², batching = same atlas+blend+filter runs (count with `-tags=ebitenginedebug`), CPU palette fades via ReplacePixels, Update/Draw split with tick interpolation. Sound: 22050 Hz default, resound SFX panning, go-mp3 music, wav/vorbis via Ebiten built-ins.
- **Lighting (L1–L8):** dLight 0=bright/15=dark, MAXLIGHTS=32, LightFalloffs float32 exact, LightConeInterpolations double sqrt, ProcessLightList two-pass with Solid-skip, VisionRays 23×15 verbatim, TRN hero states replace palette row.

---

## Goal

| Milestone | Target | Success Criteria |
|-----------|--------|------------------|
| **M0 – Scaffold** | `devilGo-i56j` (already created) | Repo compiles, `just run` opens a black window at 640×480, headless binary passes `go test ./...` |
| **M1 – Playable Spawn** (town + Cathedral L1) | `devilGo-ftbu` | Player can walk, attack, pick up items, open inventory, save/load `.sv` round‑trip, music plays |
| **M2 – Full Single‑Player** | `devilGo-vinx` (M2) | All 16 dungeon levels, all quests, Hell difficulty, Hellfire content, multiplayer‑ready net code |
| **M3 – Multiplayer** | `devilGo-vinx` (M3) | TCP lockstep 4‑player games, byte‑identical packet trace vs devilutionX |
| **M4 – Polish & Release** | 1.0 tag | Zero parity failures in CI, macOS/Windows/Linux binaries via `goreleaser`, documentation site |

**Long‑term**: The `internal/engine/` package becomes a **standalone Go module** (`github.com/bmcszk/d2engine`) that other isometric ARPG projects can import without Diablo assets.

---

## Quick Reference for Agents

| Need | Where to look |
|------|---------------|
| *How to add a new component?* | `internal/engine/ecs/component.go` + `donburi` docs |
| *Where is the palette shader?* | `shaders/palette.kage` + `internal/engine/render/palette.go` |
| *How to write a parity test?* | `test/parity/` – follow `TestFalloffsGolden` pattern |
| *How to extend the save schema?* | `internal/engine/save/schema.sql` + `internal/diablo/save/hero.go` |
| *Where are Diablo constants?* | `internal/diablo/defs/` (generated from `assets/txtdata/` TSVs) |
| *How to run only lighting tests?* | `go test -run Light ./internal/...` |
| *How to update MPQ assets?* | `just assets-extract && just assets-convert` |

---

**Remember**: Every change must keep the **1:1 contract**. If you cannot prove parity with a golden fixture, the change is incomplete.

--- 

*Generated 2026‑10‑16 — keep this file in sync with the living architecture.*