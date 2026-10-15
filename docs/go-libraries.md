# Go Libraries for devilutionGo 1:1 Port

Curated from awesome-ebitengine, pkg.go.dev, and direct source analysis. Each library verified for maintenance status and compatibility with Go 1.27+.

---

## Core Architecture

### ECS (Entity Component System)
| Library | Stars | Status | Use For |
|---------|-------|--------|---------|
| `github.com/yohamta/donburi` | 1.2K | Active | Primary ECS — entities: player, monsters, missiles, items, tiles |
| `github.com/mlange-42/ark` | 400+ | Active | Alternative archetype-based ECS |
| `github.com/sedyh/mizu` | 100+ | Active | ECS with scene management |

**Decision:** **donburi** — most mature, used in production games, good Ebiten integration.

### Scene/State Management
| Library | Stars | Use For |
|---------|-------|---------|
| `github.com/joelschutz/stagehand` | 300+ | Title→menu→game→pause→death screens, clean transitions |

### Grid/Pathfinding (Diablo's tile-based world)
| Library | Stars | Use For |
|---------|-------|---------|
| `github.com/s0rg/grid` | 200+ | A*, LOS, raycasting, shadow casting on grid |
| `github.com/SolarLune/paths` | 150+ | A* pathfinding for monster AI |
| `github.com/quasilyte/pathing` | 100+ | Grid-based pathfinding, efficient |
| `github.com/SolarLune/dngn` | 100+ | Random map generation helpers (supplement DRLG) |

**Decision:** **s0rg/grid** for LOS/raycasting + **SolarLune/paths** for monster A*.

### Camera
| Library | Stars | Use For |
|---------|-------|---------|
| `github.com/setanarut/kamera` | 200+ | Zoom, shake, lerp motion, rotation — Diablo viewport |

---

## Rendering & Graphics

### Sprite/Animation (CLX/CEL formats)
| Library | Stars | Use For |
|---------|-------|---------|
| `github.com/setanarut/aseprite` | 300+ | **Critical** — Aseprite parser (layers, tags, slices, flags, cels). Diablo CLX/CEL map to this. |
| `github.com/setanarut/aseplayer` | 100+ | Animation player for aseprite |
| `github.com/yohamta/ganim8` | 200+ | Animation library (alternative) |

**Note:** We must write **custom CLX/CEL/PCX decoders** for 1:1 byte parity. aseprite lib used for tooling/asset pipeline only.

### Font/Text
| Library | Stars | Use For |
|---------|-------|---------|
| `github.com/tinne26/etxt` | 200+ | Bitmap fonts, TRN palette coloring, wrapping, alignment |

### Tilemap
| Library | Stars | Use For |
|---------|-------|---------|
| `github.com/lafriks/go-tiled` | 300+ | TMX map parsing (dungeon levels → runtime tiles) |
| `github.com/SolarLune/ldtkgo` | 100+ | LDtk format alternative |

### Graphics Primitives & Effects
| Library | Stars | Use For |
|---------|-------|---------|
| `github.com/quasilyte/ebitengine-graphics` | 150+ | Sprites, lines, rects, batching, culling |
| `github.com/erparts/go-shapes` | 50+ | Lines, circles, rects (automap/debug overlay) |
| `github.com/mazznoer/colorgrad` | 400+ | Palette shifts, light tables, gamma, gradients |
| `github.com/lucasb-eyer/go-colorful` | 2K+ | Color space conversions, TRN manipulation |

### Shader (Kage)
| Resource | Use For |
|----------|---------|
| `ebiten/v2/shader` (stdlib) | Kage shader language — palette-lookup, light application |
| `github.com/tinne26/kage-desk` | Kage tutorials/references |

---

## Audio

| Library | Stars | Use For |
|---------|-------|---------|
| `github.com/SolarLune/resound` | 200+ | **Critical** — SFX panning, low-pass, delay, music control. Matches our `PanLogToLeftRight` + effects. |
| `github.com/hajimehoshi/go-mp3` | 500+ | MP3 decoder (devilutionx.mpq music overrides) |
| `github.com/hajimehoshi/ebiten/v2/audio/vorbis` | stdlib | OGG decoder |
| `github.com/hajimehoshi/ebiten/v2/audio/wav` | stdlib | WAV decoder (retail diabdat.mpq) |
| `github.com/quasilyte/xm` | 100+ | XM tracker music (if needed) |

**Note:** Retail = 22050 Hz 8-bit unsigned PCM WAV. We convert u8→F32 at load.

---

## Input

| Library | Stars | Use For |
|---------|-------|---------|
| `github.com/quasilyte/ebitengine-input` | 300+ | Godot-style action maps, keyboard/mouse/gamepad, fork/stick turn/walk |

---

## Asset/Resource Management

| Library | Stars | Use For |
|---------|-------|---------|
| `github.com/quasilyte/ebitengine-resource` | 200+ | Asset loading, caching, MPQ→memory→GPU pipeline |
| `github.com/quasilyte/gdata` | 100+ | Cross-platform save/prefs storage |

---

## Networking (TCP Lockstep)

| Library | Stars | Use For |
|---------|-------|---------|
| `github.com/xtaci/kcp-go` | 3K+ | Reliable UDP (alternative to TCP) |
| `github.com/golang/enet` | 500+ | ENet bindings |
| `github.com/quasilyte/necs` | 50+ | Networking layer for Donburi ECS |

**Decision:** Start with **stdlib `net` TCP** (matches devilutionX lockstep). KCP as optional upgrade.

---

## Data Formats / Archive

### MPQ (Mo'PaQ) — Critical
| Library | Stars | Status | Notes |
|---------|-------|--------|-------|
| `github.com/suprsokr/go-mpq` | 150+ | Active | **Primary** — Pure Go, MPQ v1/v2, zlib, PKWARE DCL implode **read**, patch chains, signatures. **Limitation:** No Huffman (0x01) or ADPCM (0x40/0x80) audio decompression. |
| `github.com/OpenDiablo2/mpq` | 100+ | GPL-3.0 | Diablo II focused, transcoder |

**Critical Gap:** Diablo 1 MPQ uses **PKWARE DCL implode** for data files (go-mpq supports read). Audio files may use Huffman+ADPCM — **go-mpq CANNOT decode**. Options:
1. Port C++ `encrypt.cpp` (mpqfs PKWARE explode + Huffman/ADPCM)
2. Pre-extract assets to loose files (lose MPQ compatibility)
3. Use `mpqfs` via CGO (breaks pure Go)

**Recommendation:** Pure Go port of `mpqfs` PKWARE explode + Huffman/ADPCM decode (in `internal/mpq`).

### SMK Video (Cutscenes)
| Library | Stars | Status |
|---------|-------|--------|
| `github.com/mewmew/smk` | 20+ | Header parse only — **no frame decode** |
| `github.com/greg-kennedy/libsmacker` | C lib | Full decode — would need CGO |

**Decision:** Pre-extract cutscene frames to PNG sequences (asset build step). SMK decode in pure Go is not viable.

### Translation (.po/.mo)
| Library | Stars | Use For |
|---------|-------|---------|
| `github.com/leonelquinteros/gotext` | 1.5K | GNU gettext PO/MO loader, pluralization, contexts |

---

## Testing & Parity

| Library | Stars | Use For |
|---------|-------|---------|
| `github.com/ebitengine/ebitest` | 100+ | UI testing via inputs + assertions |
| `github.com/stretchr/testify` | 20K+ | Assertions, mocks, suites |
| `github.com/google/go-cmp` | 5K+ | Deep equality for golden fixtures |

---

## Dev Tools

| Library | Stars | Use For |
|---------|-------|---------|
| `github.com/ebiten/ebiten/v2/exp/shaderprecomp` | stdlib | Shader precompilation |
| `golang.org/x/tools/gopls` | stdlib | LSP |
| `github.com/golangci/golangci-lint` | 10K+ | Linting |

---

## Summary: Must-Have Direct Dependencies

```go
// go.mod core deps
github.com/hajimehoshi/ebiten/v2           v2.9.x  // Engine
github.com/yohamta/donburi                 v1.x    // ECS
github.com/joelschutz/stagehand            v1.x    // Scenes
github.com/s0rg/grid                       v1.x    // Grid/LOS
github.com/SolarLune/paths                 v1.x    // A* pathfinding
github.com/setanarut/kamera                v1.x    // Camera
github.com/setanarut/aseprite              v1.x    // CLX/CEL tooling
github.com/tinne26/etxt                    v1.x    // Fonts
github.com/lafriks/go-tiled                v1.x    // TMX maps
github.com/quasilyte/ebitengine-graphics   v1.x    // Primitives/batching
github.com/mazznoer/colorgrad              v1.x    // Palettes
github.com/lucasb-eyer/go-colorful         v1.x    // Color spaces
github.com/SolarLune/resound               v1.x    // Audio FX
github.com/hajimehoshi/go-mp3              v1.x    // MP3 music
github.com/quasilyte/ebitengine-input      v1.x    // Input
github.com/quasilyte/ebitengine-resource   v1.x    // Assets
github.com/quasilyte/gdata                 v1.x    // Saves
github.com/leonelquinteros/gotext          v1.x    // i18n
github.com/suprsokr/go-mpq                 v0.3.x  // MPQ reader (data files)
```

---

## Must-Port from C++ (No Go Equivalent)

| C++ Component | File | Effort |
|---------------|------|--------|
| PKWARE DCL explode (implode decompress) | `3rdParty/mpqfs/src/mpq_pkware.c` | M |
| Huffman decode (MPQ audio 0x01) | `3rdParty/mpqfs/src/mpq_huffman.c` | M |
| ADPCM mono/stereo decode (0x40/0x80) | `3rdParty/mpqfs/src/mpq_adpcm.c` | M |
| CLX/CEL/CEL2/PCX decoders | `Source/engine/clx*.cpp` | L |
| DRLG dungeon generation | `Source/levels/gendung.cpp` | L |
| Light tables + falloffs + cone interp | `Source/engine/light_tables.cpp` | M |
| Vision rays (Bresenham) | `Source/vision.cpp` | S |
| RNG (xoshiro + LCG exact) | `Source/engine/random.cpp` | S |
| MPQ encryption (hash/block table) | `Source/encrypt.cpp` | S |

---

## Build Tags Strategy

```go
//go:build !headless
// +build !headless

//go:build headless
// +build headless
```

- `headless` tag: disables Ebiten window, enables CI automated runs
- `demo` tag: enables timedemo record/replay

---

## Version Pinning Policy

- Ebiten: pin minor (v2.9.x) — major upgrades break shader API
- donburi: pin major (v1.x)
- Go stdlib deps: use `golang.org/x/...` at matching Go version
- All deps: `go mod vendor` for reproducible CI