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
| **Engine (project‑agnostic)** | `internal/engine/` | Core game loop, ECS, rendering primitives, asset pipeline, audio, input, networking, save/load, configuration. No Diablo‑specific logic. | `internal/engine/ecs`, `internal/engine/render`, `internal/engine/audio` |
| **Common / shared game systems** | `internal/game/` | Systems that are generic but tuned for Diablo‑style games (isometric grid, line‑of‑sight, path‑finding, inventory UI, stats, combat formulas). | `internal/game/grid`, `internal/game/combat`, `internal/game/inventory` |
| **Diablo (project‑specific)** | `internal/diablo/` | All data definitions, level formats, quest logic, monster AI, item tables, spells, town services, MPQ asset loading, save‑file schema. | `internal/diablo/levels`, `internal/diablo/monsters`, `internal/diablo/items` |

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

Pi sessions in print mode historically ran with `-nc` (no AGENTS.md) and without these skills installed; the digest below closes that gap. Full texts: `~/.agents/skills/coding-rules/SKILL.md`, `~/.agents/skills/go-dev/SKILL.md`, `~/.agents/skills/golang-testing/SKILL.md`. If a digest line conflicts with a skill, the skill wins — report the conflict.

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