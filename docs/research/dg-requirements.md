# devilutionX → Go/Ebiten/SQLite Conversion — Requirements Study

Reference: `/home/blaze/github/bmcszk/devilutionX` (fork bmcszk/diasurgical, HEAD `ca91d778f` "Stick turn/walk + GetLeftStickDirection unification").
Target: `/home/blaze/github/bmcszk/devilutionGo` (module `github.com/bmcszk/devilutionGo`, go 1.27.1, `.beans.yml` prefix `devilGo-`, `.golangci.yml` + `justfile` present, no Go sources yet).
Measured corpus: 604 C/C++ files under `Source/` (~147.6k LOC incl. headers), 70 test files under `test/`, ~66 fixtures families under `test/fixtures/`.

All `file:line` citations below were opened and read during this study. Where a claim comes from the web, a URL is given.

---

## Global invariants

These bind the entire rewrite. Every subsystem section assumes them.

### GI-1 — Determinism is the prime directive
- The single global RNG is a Borland-compatible LCG `x ≡ 0x015A4E35·x + 1 (mod 2^32)` — `Source/engine/random.cpp:16` (`std::linear_congruential_engine<uint32_t, 0x015A4E35, 1, 0> diabloGenerator`), seeded per-use by `SetRndSeed` (`Source/engine/random.cpp:64`). Level generation begins with `SetRndSeed(rseed)` (`Source/levels/drlg_l1.cpp:1301` inside `CreateL5Dungeon(uint32_t rseed, lvl_entry entry)` at `:1298`).
- `GenerateRnd(v)`: for `v <= 0x7FFF` uses high bits `(advanceRndSeed() >> 16) % v`, else `advanceRndSeed() % v`; `advanceRndSeed` is `abs((int32)lcg())` with the `INT_MIN` special case — `Source/engine/random.hpp:55-90`. These quirks (negative draws, the 32768..65534 bias) are load-bearing for item/level parity.
- Per-entity RNG seeds: monsters carry `rndItemSeed` (item drops on death) and `aiSeed` (`Source/monster.h:229-232`); objects carry `_oRndSeed` for container loot (`Source/objects.h:73-76`); items carry `_iSeed` (`Source/items.h:191`). The Go engine must replicate the "reserve seed then advance" pattern (`ReserveSeedSequence`, `Source/engine/random.cpp:52`).
- Consequence: no goroutine may touch game-logic RNG; logic runs single-threaded in one deterministic tick function (see GI-2).

### GI-2 — Fixed 20 TPS logic tick, decoupled rendering
- `gnTickDelay = 50` ms (`Source/diablo.cpp:146`), recomputed as `1000 / sgGameInitInfo.nTickRate` when the in-game speed slider or multiplayer negotiation sets a rate 20–50 (`Source/gamemenu.cpp:277`, `Source/multi.cpp:879`). 20 TPS is the default.
- The main loop (`RunGameLoop`, `Source/diablo.cpp:867`) polls SDL events, then `nthread_has_500ms_passed(&drawGame)` (`Source/nthread.cpp:214`) decides whether a logic tick is due; if not due it only re-renders. One tick = `game_loop()` (`Source/diablo.cpp:3490`) → up to `multi_handle_delta()` iterations → `GameLogic()` (`Source/diablo.cpp:1523`).
- `GameLogic()` fixed order: ProcessInput → ProcessPlayers → (dungeon: ProcessMonsters → ProcessObjects → ProcessMissiles → ProcessItems → ProcessLightList → ProcessVisionList) | (town: ProcessTowners → ProcessItems → ProcessMissiles) → sound_update → CheckTriggers → CheckQuests → RedrawViewport → pfile_update (`Source/diablo.cpp:1523-1568`). This order is deterministic and must be preserved exactly.
- Interpolation for rendering: `ProgressToNextGameTick` is a 0..128 fraction (`nthread_UpdateProgressToNextGameTick`, `Source/nthread.cpp:246`); `AnimationInfo::baseValueFraction = 128` (`Source/engine/animationinfo.h:114`).

### GI-3 — Save-compat decision: **byte-compatible read/write of vanilla `.sv` saves**
- Save files are MPQ containers written per-hero as `single_%d.sv` / `spawn_%d.sv` / `multi_%d.sv` (+hellfire variants) under PrefPath (`Source/pfile.cpp:80-97` `GetSavePath`), with fixed passwords `"xrgyrkj1"` (single), `"szqnlsk1"` (multi), `"adslhfb1"` (spawn single), `"lshbkfg1"` (spawn multi) — `Source/pfile.cpp:60-63`.
- The inner "game" stream starts with a 4-char magic `RETL`/`SHAR`/`HELF`/`SHLF` and mixes big-endian (classic fields) and little-endian (newer fields) writes (`Source/loadsave.cpp:2763-2800` `SaveGameData`).
- **Decision**: Go must read (and for round-trip parity, write) these files byte-identically; SQLite is the *additional* native store (see §5). Round-trip tests: `test/pack_test.cpp:441-868`, `test/writehero_test.cpp`, `test/codec_test.cpp`.

### GI-4 — Asset-compat decision: **read retail MPQs + reuse devilutionX's derived assets**
- MPQ load order and priorities: devilutionx.mpq (9000) → fonts (9200) → language (9100, loaded after core, replaces) → DIABDAT/diabdat/spawn (1000) → hfbard (8110) → hfbarb (8120) → hellfire (8000) → hfmonk (8100) → hfmusic (8200) → hfvoice (8500) → mods (≥10000) — `Source/engine/assets.cpp:505-594,704-760`, priorities at `Source/engine/assets.hpp:309-312`. Higher priority wins on duplicate names (an ordered map keyed by priority).
- Compression actually required: PKWARE DCL (0x08), zlib (0x02), bzip2 (0x10) — implemented by the vendored `mpqfs` library (`build-v2/_deps/mpqfs-src/src/mpq_stream.c:180-225` after FetchContent from `https://github.com/diasurgical/mpqfs` per `3rdParty/mpqfs/CMakeLists.txt:9-11`); Huffman/ADPCM are explicitly *not* implemented there, and retail `diabdat.mpq` files for graphics/data decode fine with the three supported methods. Go's `github.com/suprsokr/go-mpq` (MIT) supports exactly PKWARE DCL + zlib + bzip2 read and omits Huffman "audio files only, not implemented" — https://pkg.go.dev/github.com/suprsokr/go-mpq (Feature Support table). **Gap risk is low but must be proven per-file with an extraction smoke test over diabdat.mpq + spawn.mpq + hellfire.mpq** (see §6 Gaps).
- devilutionX ships derived assets (`assets/`: txtdata TSV tables, CLX re-encodes, fonts, ui_art, levels L*.dun for setmaps, plrgfx, gendata, lua) packed into `devilutionx.mpq` — the Go build should consume the same source tree (vendored or via `ASSETS_VERSION`-pinned submodule/snapshot) rather than re-deriving.

### GI-5 — Platform targets & non-goals
- Targets: Linux/Windows/macOS desktop first (Ebitengine supports all three plus mobile/wasm — https://ebitengine.org/); headless mode for tests/CI is a first-class build tag (`HeadlessMode` global, `Source/headless_mode.cpp:5`; used by `test/timedemo_test.cpp:60`).
- Non-goals (cut from scope): ZeroTier/lwip virtual LAN (replaced by plain TCP), SDL movie (SMK) playback on day one (defer behind an interface; see §7), hardware cursor specifics (Ebiten draws its own cursor), console ports (Switch/3DS/vita), UWP, `discord/` rich presence (2 files, trivially re-addable later), NES-style `--spawn` demo installer flows.
- KEPT concepts: demo/timedemo mode as regression harness (`Source/engine/demomode.cpp`), headless mode, per-tick delta command system, spawn (shareware) gating via `gbIsSpawn` (`Source/quests.cpp:96`, `Source/monster.cpp:3159`, `Source/player.cpp:3479`), Hellfire as a data-driven extension (`gbIsHellfire` appears in 135 root-cpp lines).

### GI-6 — Language/stack pins
- Go 1.27.1, Ebitengine v2 (`github.com/hajimehoshi/ebiten/v2`), SQLite via **modernc.org/sqlite** (pure Go, no CGO → trivial cross-compile to all targets; mattn/go-sqlite3 rejected as CGO-bound) — rationale: the game is single-player logic + small blobs; driver performance is irrelevant at this write frequency.
- Networking over TCP with goroutines; MPQ reading via `github.com/suprsokr/go-mpq` (MIT) with a thin archive-priority layer; hand-written decoders for CEL/CL2/CLX/PCX/TRN/DUN/MIN/TIL/SOL + txtdata TSV + Diablo-codec/SHA1.
- Performance rules honored from day one: logic only in `Update()`, no allocations in `Draw()`, sprite sheets + `SubImage`, batched `DrawImage` with identical options merge into one draw call — https://ebitengine.org/en/documents/performancetips.html

---

## 1. Core loop + threading

### Purpose
Reproduce the fixed-tick simulation driver: 20 TPS logic, render at monitor rate, deterministic catch-up behavior, headless and demo variants, and map it onto Ebiten's `Update`/`Draw` split.

### Key sources (file:line)
- `Source/nthread.cpp:48` `NthreadHandler` (network thread loop), `:80` `nthread_send_and_recv_turn`, `:105` `nthread_recv_turns`, `:145` `nthread_start`, `:214` `nthread_has_500ms_passed`, `:246` `nthread_UpdateProgressToNextGameTick`
- `Source/diablo.cpp:146` (`gnTickDelay`), `:867-969` (`RunGameLoop`), `:925` (tick-due decision incl. demo), `:1523-1568` (`GameLogic`), `:3490-3507` (`game_loop`), `:3509` (`diablo_color_cyc_logic`)
- `Source/gamemenu.cpp:277` and `Source/multi.cpp:879` (tick rate 20–50 → `gnTickDelay`)
- `Source/engine/demomode.cpp:148` (`FILE *DemoRecording`), `:595-599` (record format: byte with `ProgressToNextGameTick | 0b10000000`, event type, payload), `:626` `InitRecording`, `:674` `IsRecording`
- `Source/headless_mode.cpp:5` (`bool HeadlessMode`)
- `Source/game_mode.hpp` (`GameLogicStep` enum used to tag the current phase for crash reporting)

### Data structures
```cpp
// Source/diablo.cpp:146
uint16_t gnTickDelay = 50;              // ms per logic tick; 1000/nTickRate
uint8_t  ProgressToNextGameTick;        // 0..128 render interpolation fraction
// Source/nthread.cpp:14-20 (file scope)
uint8_t  sgbNetUpdateRate;              // turns per logic tick batch
uint32_t gdwTurnsInTransit;             // from provider caps
int      last_tick;                     // SDL_GetTicks anchor
// Source/engine/demomode.cpp:43-58 (per-event records)
struct DemoMsg { enum EventType : uint8_t { GameTick = 0, ... }; ... };
// Source/engine/animationinfo.h:114
constexpr static uint8_t baseValueFraction = 128;
```

### Algorithms
1. Tick due check (`nthread_has_500ms_passed`, `Source/nthread.cpp:214-243`): `ticksElapsed = now - last_tick`; if `> gnTickDelay*10` (missed >10 ticks) reset `last_tick` when alone/single-player to avoid fast-forward storms; `drawGame = ticksElapsed <= gnTickDelay` (skip render to catch up when behind).
2. Multiplayer send cadence (`nthread_recv_turns`, `:105-143`): countdown `sgbPacketCountdown` per `sgbNetUpdateRate`, sync countdown every 4 batches; turns batched in transit `gdwTurnsInTransit` from provider caps (`nthread_start:145-176`, incl. `gdwNormalMsgSize = bytessec*rate/20 *3/4 /maxplayers`, doubling until ≥0x80). The turn consumer side is `multi_handle_delta` + `multi_msg_countdown` (`Source/multi.cpp:665`).
3. Catch-up loop (`game_loop`, `Source/diablo.cpp:3490`): on startup runs `nTickRate*3` deltas; normally 1 delta then `GameLogic()`, repeating while behind (`nthread_has_500ms_passed()` with null drawGame → return `ticksElapsed >= 0`, i.e. always true when not paused — used purely as a break-glass check).
4. Color cycling after each successful tick: caves/hell/nest/crypt palette animations (`diablo_color_cyc_logic`, `:3509-3529`).
5. Demo mode replaces the tick-due source: `demo::GetRunGameLoop(drawGame, processInput)` (`Source/diablo.cpp:925`) replays recorded events tagged with the interpolation byte, and `--timedemo` disables frame limiting (`Source/diablo.cpp:1075-1079` help text).
6. Game bootstrap order: `StartGame` (`Source/diablo.cpp:179`, called from `RunGameLoop:872`) → table loads in fixed order at `Source/diablo.cpp:2792-2799` (`LoadPlayerDataFiles`, `LoadSpellData`, `LoadMissileData`, `LoadMonsterData`, `LoadItemData`, `LoadObjectData` — loaders live at `Source/tables/itemdat.cpp:710`, `Source/tables/misdat.cpp:442`, `Source/tables/monstdat.cpp:472`, `Source/tables/objdat.cpp:257`) → player creation `CreatePlayer` (`Source/player.cpp:2300`) → `InitPlayer` (`Source/player.cpp:2479`) → per-level `CreateDungeon`.

### Determinism constraints
- Exactly one `GameLogic()` per delta; delta count per real-time window is a *timing* concern, never a simulation input. The sim never reads wall-clock time (the only clock reads are `nthread` pacing and `DoLighting`-independent UI effects).
- Pause semantics: `PauseMode != 0` or active gmenu (single-player) freezes ticks but rendering continues (`Source/nthread.cpp:247` guard).

### Go/Ebiten mapping decision
- `ebiten.SetTPS(20)` is **not** sufficient — Update() must not drift from `gnTickDelay` when the user changes in-game tick rate (20–50). Decision: run Ebiten at default 60 TPS with **an internal accumulator**: `Update()` accumulates `dt` and runs `GameLogic()` while `acc >= gnTickDelay`, capped at 10 ticks per Update (mirroring the >10-tick reset). `Draw()` uses the same `ProgressToNextGameTick` fraction (recomputed as `acc/gnTickDelay*128`) for interpolation — this reproduces devilutionX's 60-fps smoothness at 20 TPS exactly and matches Ebiten's "logic in Update only" guidance (https://ebitengine.org/en/documents/performancetips.html, local skill `/home/blaze/.agents/skills/ebiten/game-loop/SKILL.md`).
- The `nthread` network thread maps to a goroutine owning `send_and_recv_turn` pacing over a channel; the game goroutine is the only mutator of world state (GI-1).
- Headless mode = build tag `headless` with a null renderer; demo files parsed with the same reader in tests.

### Gaps/risks
- Ebiten's fixed TPS jitter differs from SDL_Delay pacing; accumulator drift must be clamped identically to the >10-tick rule or multiplayer sync suffers.
- `game_loop` startup triple-tick burst (`wait = nTickRate*3`) is easy to miss.

### Effort: M

### Open questions
- Do we ever want `--timedemo` uncapped rendering in Go, or is headless+state-hash enough? (Decision criterion: whether frame-exact parity vs the C++ binary is a milestone; default: state-hash only.)

---

## 2. World/level system

### Purpose
Generate and represent dungeons: DRLG for Cathedral/Catas/Caves/Hell, static Town, Crypt (Hellfire), quest set-maps, theme rooms, triggers, tile metadata, and per-level seeds.

### Key sources (file:line)
- `Source/levels/gendung_defs.hpp:5-9` (`DMAXX 40`, `DMAXY 40`, `MAXDUNX = 16+40*2+16 = 112`, `MAXDUNY`), `:11` (`NUMLEVELS 25`)
- `Source/levels/dun_tile_data.hpp:26-42` (`DungeonFlag` bits), `:49-56` (`THEME_LOC`, `MegaTile` 4×uint16 micros), `:60-120` (all world grids: `dungeon[DMAXX][DMAXY]`, `pdungeon`, `Protected`, `pMegaTiles`, `pDungeonCels`, `SOLData[MAXTILES=1379]`, `dPiece/dTransVal/dLight/dPreLight/dFlags/dPlayer/dMonster/dCorpse/dObject/dSpecial[MAXDUNX][MAXDUNY]`, `DungeonSeeds[NUMLEVELS]`, `LevelSeeds`)
- `Source/levels/dun_tile.hpp:12-71` (`TileType` enum: Square, TransparentSquare RLE, LeftTriangle 32×31 with even-row 2-byte pad, RightTriangle, LeftTrapezoid, RightTrapezoid), `:83-104` (`LevelCelBlock`: `type=(data&0x7000)>>12`, `frame=data&0xFFF`), `:105-116` (`TileProperties` bits: Solid/BlockLight/BlockMissile/Transparent/TransparentLeft/TransparentRight/Trap), `:124-130` (`MICROS{LevelCelBlock mt[16]}`, `DunFrameWidth=32`, `DunFrameHeight=32`, triangle height 31)
- `Source/levels/gendung.h:35-70` (`Miniset` with `search[6][6]`/`replace[6][6]`, `matches` respecting `Protected`, `place`), `:86-95` (`PlaceMiniSet` w/ `drlg1Quirk`), `DRLG_PlaceThemeRooms`, `DRLG_LPass3`, `LoadDungeonBase`, `PlaceDunTiles`
- `Source/levels/drlg_l1.cpp:1238-1246` (`Pass3` → `DRLG_LPass3(Dirt-1)`), `:1249-1266` (`PlaceMiniSetRandom` scanning full map with `GenerateRnd(100) >= rndper` gate), `:1268-1294` (`SelectChamber` picking quest chamber by availability + `VerticalLayout`), `:1298-1312` (`CreateL5Dungeon(rseed, entry)`: `SetRndSeed(rseed); GenerateLevel(entry); Pass3();`), `:1314-1330` (`LoadPreL1Dungeon`/`LoadL1Dungeon` from `.dun` bytes)
- `Source/levels/setmaps.cpp:112-165` (`LoadSetMap`: per-quest `.dun` loads, e.g. `SL_SKELKING → levels\l1data\sklkng1.dun + sklkng2.dun + sklkngt.dun` transparency + palette `l1_2.pal` + `AddSKingObjs` + `InitSKingTriggers`)
- `Source/levels/themes.cpp:377-449+` (`Theme_Barrel`, `Theme_Shrine`, `Theme_MonstPit`, `PlaceThemeMonsts`), theme id enum at `Source/tables/objdat.h:18-37` (THEME_BARREL..THEME_WEAPONRACK, THEME_NONE=-1), `ThemeStruct` at `Source/levels/themes.h:14-17`
- `Source/levels/town.cpp` (static town from `levels\towndata\town.dun`; town warp gating by quest progress `IsWarpOpen(DTYPE_CATACOMBS/CAVES/HELL/NEST/CRYPT)` at `Source/levels/town.cpp:210-234`), `Source/levels/crypt.cpp` (Hellfire crypt = L1 generator variant with `CryptSubstitution`, `PlaceCryptLights` — `Source/levels/drlg_l1.cpp:1218-1225`)
- L2/L3/L4 entry points: `CreateL2Dungeon(uint32_t rseed, lvl_entry entry)` at `Source/levels/drlg_l2.cpp:2830` (+`LoadL2Dungeon:2849`), `CreateL3Dungeon` at `Source/levels/drlg_l3.cpp:2185`, `CreateL4Dungeon` at `Source/levels/drlg_l4.cpp:1228` — all follow the same `SetRndSeed → generate → Pass3` skeleton as L1
- `Source/levels/tile_properties.cpp:11-41` (`IsTileNotSolid`/`IsTileSolid` — direct SOL lookups), `CanStep` corner-cut rule at `Source/levels/tile_properties.hpp:28-37`
- `Source/levels/gendung.cpp:326` `GetLevelType(int level)` (level→dungeon_type mapping incl. Hellfire 17-24)
- `Source/levels/trigs.cpp` (stairs/portal triggers; `InitSKingTriggers` et al.)
- `Source/lighting.cpp:117-160` (`DoLighting` 15-tile radius loop with per-quadrant bounds), `Source/lighting.h:30-37` (`Light{position,radius,oldRadius,isInvalid,hasChanged}`), `MAXLIGHTS 32` (`Source/engine/lighting_defs.hpp:7`), `LightFalloffs[16][128]` (`Source/engine/light_tables.hpp:16-23`)
- `.dun` container: header = two uint16 LE dims then tile/passthrough arrays — visible in fixture bytes `test/fixtures/diablo/1-2588.dun` (`0x28 0x00` ×2 = 40×40); consumed by `LoadFileInMem<uint16_t>` + `PlaceDunTiles` (`Source/levels/drlg_l1.cpp:1317-1322`)

### Data structures (Go targets)
```go
// world/dungeon.go
type DungeonFlag uint8 // Missile|Visible|DeadPlayer|Populated|MissileFireWall|MissileLightningWall|Lit|Explored
const DMAXX, DMAXY, MAXDUNX, MAXDUNY = 40, 40, 112, 112
type World struct {
    Dungeon   [DMAXX][DMAXY]uint8        // tile ids at generation resolution
    PDungeon  [DMAXX][DMAXY]uint8        // backup for regeneration parity
    Protected utils.Bitset2D             // 40×40
    DPiece    [MAXDUNX][MAXDUNY]uint16   // micro-tile ids
    DFlags    [MAXDUNX][MAXDUNY]DungeonFlag
    DLight    [MAXDUNX][MAXDUNY]uint8    // 0=bright..15
    DTransVal [MAXDUNX][MAXDUNY]int8
    DPlayer   [MAXDUNX][MAXDUNY]int8     // +id / -id-1 while moving
    DMonster  [MAXDUNX][MAXDUNY]int16
    DCorpse   [MAXDUNX][MAXDUNY]int8     // idx | dir<<5
    DObject   [MAXDUNX][MAXDUNY]int8
    DSpecial  [MAXDUNX][MAXDUNY]int8
    SOLData   [1379]TileProperties
    Micros    []MICROS                   // per pMegaTile: 16 LevelCelBlock
    Seeds     [25]uint32                 // DungeonSeeds
}
```

### Algorithms
1. **L1 Cathedral**: `SetRndSeed(seed)` → room-grid generation (`GenerateLevel`) with 3 quest chambers (`SelectChamber`), miniset substitutions (stairs/doors/lamps: `PlaceMiniSetRandom(LAMPS, 1600, true)` ×(5+`GenerateRnd(5)`)), transparency fixes (`FixTransparency/FixDirtTiles/FixCornerTiles`), `Substitution()`, shadows, then `Pass3()` which expands each 1×1 dungeon cell into a 2×2 mega-tile of 16 micros (`DRLG_LPass3`, `Source/levels/gendung.h:88`).
2. **Theme rooms**: after room detection (`DRLG_PlaceThemeRooms(min,max,floor,freq,rndSize)`), `CreateThemeRooms` paints a `theme_id` per room; theme counts capped `MAXTHEMES 50` (`Source/levels/dun_tile_data.hpp:21`); concrete theme painters at `Source/levels/themes.cpp:377` (`Theme_Barrel`), `:400` (`Theme_Shrine`), `:422` (`Theme_MonstPit`), `:349` (`PlaceThemeMonsts`) with object placement via `AddObject` (`Source/objects.cpp:276-345`).
2b. **L2 Catacombs** uses random rectangles: `RoomMax 10`/`RoomMin 4`, `roomSize.width = GenerateRnd(min(area.width,RoomMax)-RoomMin)+RoomMin` (`Source/levels/drlg_l2.cpp:1780-1786`), scatter pass gated by `GenerateRnd(100) >= rndper` (`:1626`); pre-set piece via `LoadPreL2Dungeon:2839`; crypt variant branches at `Source/levels/drlg_l1.cpp:1214` (`FixCryptDirtTiles`) and `:1221` (`CryptSubstitution`).
3. **Lighting**: static pre-light pass then dynamic `DoLighting` writes min(current, falloff) per tile; 16 radius tables of 128 entries; crypt/nest allow dim bases (`Source/lighting.cpp:150-156`). Runtime: `AddLight` (`Source/lighting.cpp:254`), `ChangeLightRadius:290`, `ProcessLightList:366`, `ProcessVisionList:437` — both called from `GameLogic` (`Source/diablo.cpp:1545-1546`).
4. **Triggers**: per-level trigger list built by `Init*Triggers` (e.g. `InitSKingTriggers`) and checked each tick by `CheckTriggers()` (called from `GameLogic`, `Source/diablo.cpp:1563`).
5. **Monster population per level**: `InitLevelMonsters`/`AddMonsterType` with placement flags incl. uniques (`PLACE_SCATTER`, `PLACE_UNIQUE` — `Source/monster.cpp:566-603` for WarlordOfBlood/Lachdan/Lazarus/RedVex), `InitMonster:195`, `CanPlaceMonster:282`, `PlaceMonster:292`.

### Determinism constraints
- Same seed ⇒ byte-identical `dungeon`, `dPiece`, themes, monster spawn. The test corpus proves this: `test/fixtures/diablo/*.dun` are named `level-seed.dun` and asserted by `test/drlg_l1_test.cpp` … `drlg_l4_test.cpp` against `CreateL5Dungeon`-family calls.
- `pdungeon` backup exists because re-entering a level re-runs parts of generation; ordering of miniset placement is part of the algorithm.

### Go/Ebiten mapping decision
- Package `world` owns the grids as flat arrays (`[MAXDUNX*MAXDUNY]` with index math) for cache locality; generation code is a near line-for-line port of drlg_l1..l4/town/crypt/setmaps/themes/trigs, sharing a `Miniset` engine.
- Micros decode once into `*ebiten.Image` sub-tiles at asset-load time (see §7), keeping the *logic* side pure indices (`dPiece`) — identical to C++.

### Gaps/risks
- L2/L3/L4 generators are separate algorithms (`drlg_l2.cpp` linear-room corridors, `drlg_l3.cpp` caves flood, `drlg_l4.cpp` rivers/crannies) — each is its own XL port with its own quirks; the fixtures cover them (`test/fixtures/levels/l2data`…).
- Crypt reuses L1 code paths with crypt-specific substitution — fork carefully, not copy-paste.

### Effort: XL

### Open questions
- Keep the 40×40 → 112×112 two-resolution model, or generate directly at micro resolution? (Decision: keep — save format stores both and parity tests depend on it.)

---

## 3. Entities

### Purpose
Player, Monster, Missile, Object, Item, Towner, Corpse — their state, animation progression, 8-way facings, movement, and death handling.

### Key sources (file:line)
- `Source/player.h:190-389` (full `Player` struct), `:211` start; inventory: `InvBody[NUM_INVLOC]`, `InvList[InventoryGridCells=40]` (`player.h:38`), `InvGrid[40]`, `SpdList[MaxBeltItems=8]` (`player.h:39`), `walkpath[MaxPathLengthPlayer=100]` (`Source/engine/path.h:20`), `_pSplLvl[64]`, `_pSplHotKey[NumHotkeys=12]` (`player.h:44`), spell bitmasks `_pISpells/_pMemSpells/_pAblSpells/_pScrlSpells` as uint64; `Players` is `std::vector<Player>` (`player.h:918`), `MAX_PLRS 4` (`Source/multi.h:26`)
- `Source/monster.h:181-216` (`MonsterSpritesData`, `CMonster` with `anims[6]`, `sounds[4][2]`, `LevelMonsterTypes[MaxLvlMTypes]`), `:213-310` (`Monster`: `rndItemSeed`, `aiSeed`, `goalVar1..3`, `var1..3`, `goal`, `enemyPosition`, `MonsterMode mode`, `pathCount`, `direction`, `enemy`, `MonsterAIID ai`, `intelligence`), `MaxMonsters 200` (`monster.h:38`), `Monsters[MaxMonsters]` + `ActiveMonsters[]` (`monster.h:494-495`), AI enum `Source/tables/monstdat.h:22-58` (Zombie…Lich, 40+ ids)
- `Source/missiles.h:25-59` (`MissilePosition` with pixel-precision `offset/velocity/traveled` as 16.16 fixed point "Pixels traveled as a numerator of 65,536"), `:89-200` (`Missile`: `_mitype`, `_mispllvl`, `_miDelFlag`, anim fields, `_midam`, `var1..var7`, `lastCollisionTargetHash`, `IsTrap() == _misource==-1`, `sourcePlayer/sourceMonster`); `Missiles` is `std::list<Missile>` (`missiles.h:238`)
- `Source/objects.h:32-80` (`Object`: `_otype`, trap/door flags, anim counters, `_oRndSeed`, `_olid` light id); `MAXOBJECTS 127` (`objects.h:30`), `Objects[MAXOBJECTS]` (`objects.h:284`)
- `Source/items.h:189-230+` (`Item`: `_iSeed`, `_iCreateInfo`, `_itype`, `_iIdentified`, `_iMagical`, name buffers, `_iCurs`, `_ivalue/_iIvalue`, damage/AC/durability, `ItemSpecialEffect _iFlags`, misc id, spell, charges, affix stat fields…); `MAXITEMS 127` (`items.h:25`), `Items[MAXITEMS+1]` (`items.h:487`)
- `Source/towners.h:48-63` (`Towner`: `talk` fn pointer, `animOrder` frame sequence, gossip topic, anim delay) — towners are data-driven from `txtdata\towners\towners.tsv` (fixture row: `TOWN_SMITH Griswold the Blacksmith 62 63 SouthWest 96 towners\smith\smithn 16 3 …` with a 100+-entry animOrder)
- `Source/dead.cpp` + `dCorpse` encoding `idx | dir<<5` (`Source/levels/dun_tile_data.hpp:110-116`), corpse render at `Source/engine/render/scrollrt.cpp:798-806`
- `Source/engine/animationinfo.h:112-131` (`baseValueFraction=128`, progress API), `Source/engine/animationinfo.cpp:21-95` (frame distribution across ticks, frame-skipping before attack hits `:115`)
- `Source/engine/direction.hpp` (8-way `Direction` + `Direction::NoDirection`), `Source/engine/actor_position.hpp` (`tile`, `future`, pixel `offset`)
- Monster data tables: `Source/tables/monstdat.cpp:52-55` (`MonstersData`, `UniqueMonstersData` vectors), loaded from `txtdata\monsters\monstdat.tsv` (`Source/tables/monstdat.cpp:409`) — fixture header shows 40+ columns incl. `frames[6]`, `rate[6]`, `ai`, `abilityFlags`, `immunities/resistances` (`test/../build-v2/mods/hf/txtdata/monsters/monstdat.tsv:1-2`)

### Data structures (example Go mirror — Player core)
```go
type Player struct {
    Name       [32]byte
    InvBody    [7]Item   // NUM_INVLOC
    InvList    [40]Item
    InvGrid    [40]int8  // -1 empty, else InvList index+1 (curse: 0..39)
    SpdList    [8]Item   // belt
    NumInv     int
    Strength, Magic, Dexterity, Vitality int  // + base variants
    HPBase, MaxHPBase, HP, MaxHP int
    ManaBase, MaxManaBase, Mana, MaxMana int
    Experience uint32
    Mode       PLRMode
    WalkPath   [100]int8
    Dir        Direction
    Class      HeroClass
    SplLvl     [64]uint8
    MemSpells  uint64
    // ...
}
```

### Algorithms
- **Animation progression** (`AnimationInfo`): frames are distributed over ticks with a tickModifier; `getAnimationProgress()` = `(progressToNextGameTick + ticksSinceSequenceStarted) * tickModifier / frames / 128` (`Source/engine/animationinfo.cpp:37-55`); render picks `getFrameToUseForRendering()` using `ProgressToNextGameTick` — this is what makes 20 TPS look smooth at 60 fps.
- **Monster AI**: `MonsterAIID` dispatch (40+ behaviors, `Source/tables/monstdat.h:22-58`); each monster has `goal`/`goalVar`s implementing seek/attack/retreat FSMs; AI runs only for active monsters (`ActiveMonsters` swap-list).
- **Missiles**: fixed-point 16.16 sub-tile motion (`MissilePosition.traveled`), collision via `CheckMissileCol(missile, damageType, minDamage, maxDamage, isDamageShifted, position, dontDeleteOnCollision, onlyHitWalking)` (`Source/missiles.cpp:479`) wrapped by `MoveMissileAndCheckMissileCol` (`:636`); chained spawns via `AddMissile` (`Source/missiles.cpp:728,748,828,1242` — e.g. Guardian firebolts, teleport explosions); behavior is table-driven by `MissileData.addFn/processFn` function pointers (`Source/tables/misdat.h:166-196`).
- **Items on ground**: `SpawnItem(monster, position, sendmsg, spawn)` (`Source/items.cpp:3422`) routes unique-monster treasure through `SpawnUnique` (`:3195`, `T_MASK` decode of the `treasure` field) and rolls durability via `ItemRndDur:1475`.
- **Towners**: populated from the TSV-backed `TownersData` array in `InitTowners` (`Source/towners.cpp:760`), ticked by `ProcessTowners:811` (called from `GameLogic` town branch, `Source/diablo.cpp:1556`).
- **Player stat recalc**: `CalcPlrInv(player, true)` after every equipment change (`Source/player.cpp:450-494` call sites) — the single funnel for derived stats (AC/damage/mana/HP).
- **Dead/corpse**: on monster death a `Corpse` record + `dCorpse` occupancy; players drop items per `ApplicationsPlrHit` death rules (multiplayer ear: `TEar` struct `Source/msg.h:549-555`).

### Determinism constraints
- Iteration order of `ActiveMonsters`/`ActiveItems`/`ActiveObjects` and the `Missiles` list order are gameplay-visible; Go must keep the same swap-remove/append-list structures and processing order from `GameLogic` (`ProcessMonsters` before `ProcessObjects` before `ProcessMissiles`…, `Source/diablo.cpp:1530-1553`). Tile queries use the constant helpers `InDungeonBounds` / `TileContainsMissile` / `IsTileVisible` (`Source/levels/dun_tile_data.hpp:143-175`) — port them as `inline` funcs, not methods, to keep hot loops clean.
- `AnimationInfo` math must use the same integer fixed-point (128ths), no floats.

### Go/Ebiten mapping decision
- Entity structs stay value-heavy arrays inside a `GameState` (like C++ globals) — no ECS; the fork semantics and save format both assume index-stable arrays. Skill `/home/blaze/.agents/skills/ebiten/ecs` is *not* applied (documented deviation: ECS would break parity tooling).
- `Missiles` as a container-list port: a slice pool with explicit free-list to mimic `std::list` splice semantics for save ordering.

### Gaps/risks
- `std::list` iteration order on missile deletion differs subtly from a slice pool; write a missile-order parity test early.
- Player struct is huge (~100 fields); port it *completely* on day one even if unused — saves pain when save code lands.

### Effort: XL

### Open questions
- Do we mirror `_pIFlags` bitmask powers via generated code from `itemdat` flags or hand-enum? (Default: hand-enum mirroring `ItemSpecialEffect` + `ItemSpecialEffectHf`.)

---

## 4. Gameplay systems

### Purpose
Combat resolution, spells, missile behaviors, item generation (base/affix/unique), durability, inventory grid + belt, stores, quests with per-level triggers, XP curve, spawn vs retail gating, Hellfire deltas.

### Key sources (file:line)
- **To-hit**: `Source/player.h:575` `GetMeleeToHit = level + dex/2 + _pIBonusToHit + baseMeleeToHit`; `:582-590` `GetMeleePiercingToHit` (+`_pIEnAc` unless Hellfire); `:594` ranged = `level + dex + …`; `:654-666` `CalculateArmorPierce` (Hellfire: `tmac >>= pIEnAc-1` or `-= 25%`; Barbarian melee bonus)
- **Melee hit & damage** (`PlrHitMonst`, `Source/player.cpp:524-660`): `hit = GenerateRnd(100)` (petrified ⇒ 0); `hper += GetMeleePiercingToHit() - CalculateArmorPierce(monster.armorClass, true)`; `clamp(hper, 5, 95)`; on hit: `dam = RandomIntBetween(_pIMinDam, _pIMaxDam); dam += dam*_pIBonusDam/100; dam += _pIBonusDamMod; int dam2 = dam<<6; dam += _pDamageMod;` warrior crit `GenerateRnd(100) < level` ⇒ ×2; sword/mace vs Undead/Animal ±50%; demon ×3 with `TripleDemonDamage`; Hellfire Devastation 5% ⇒ ×3, Doppelganger 10% spawns clone, Jesters `r=GenerateRnd(201); if r>=100 r=100+(r-100)*5; dam = dam*r/100`; final `dam <<= 6` (damage in 1/64ths!), `adjacentDamage >>= 2`. The helpers: `RandomIntLessThan` (`Source/engine/random.hpp:391`) and `RandomIntBetween(min, max, halfOpen)` (`:403-405`, `RandomIntLessThan(max-min+1)+min`); class to-hit bases loaded from TSV (`Source/tables/playerdat.cpp:203` reads `baseMeleeToHit`; accessor `GetClassAttributes:379`)
- **Monster→player**: symmetric roll at `Source/player.cpp:712` (`hit = GenerateRnd(100)` + `hper` clamp 5..95)
- **Spells**: `Source/tables/spelldat.h:225-252` (`SpellData`: name, sfx, bookCost10/staffCost10 (×10 for real price), `sManaCost`, flags (type Fire/Lightning/Magic), `sBookLvl/sStaffLvl`, minInt, `sMissiles[2]`, `sManaAdj/sMinMana`, staff charges) — loaded from `txtdata\spells\spelldat.tsv`
- **Missile table**: `Source/tables/misdat.h:166-196` — `AddFn addFn; ProcessFn processFn; SfxID castSound, hitSound; MissileGraphicID graphic; MissileDataFlags flags; MissileMovementDistribution movementDistribution;` with `damageType = flags & 0b111`; data from `txtdata\missiles\misdat.tsv`
- **Items**: base table `AllItemsList` + affixes `ItemPrefixes/ItemSuffixes` as `std::vector<PLStruct>` (`Source/tables/itemdat.h:628-655`); `PLStruct{PLName, power(ItemPower), PLMinLvl, PLIType(bitmask), PLGOE, PLChance, PLOk, minVal, maxVal, multVal}`; `UniqueItem{UIName, UICurs, UIItemId, UIMinLvl, UINumPL, UIValue, powers[6], mappingId}`; affix application `SaveItemAffix` (`Source/items.cpp:1048`) and prefix/suffix selection at `:1138/:1218-1224`
- **Durability**: decrement per blocked hit; `DUR_INDESTRUCTIBLE` sentinel, break→`RemoveEquipment`+`CalcPlrInv` (`Source/player.cpp:505-522`)
- **Inventory grid**: `InvGrid[40]` maps 10×4 cells to `InvList` indices; belt `SpdList[8]`; placement rules `CanUseScroll/CanPlaceItem` etc. in `Source/inv.cpp` (tests: `test/inv_test.cpp`, `test/inventory_ui_test.cpp`)
- **Stores**: `TalkID` enum (`Source/stores.h:44-70`: None, Smith, SmithBuy, SmithSell, SmithRepair, Witch, WitchBuy, WitchSell, WitchRecharge, NoMoney, NoRoom, Confirm, Boy, BoyBuy, Healer, Storyteller, HealerBuy, StorytellerIdentify, SmithPremiumBuy, Gossip, StorytellerIdentifyShow, Tavern, Drunk, Barmaid); `ActiveStore` state machine + `StartStore` re-entry (`Source/stores.cpp:504,911-929,1192,1293`); 2911 lines total
- **Quests**: `Quest` struct (`Source/tables/questdat.hpp:102-115`: `_qidx,_qactive,_qlevel,position,_qlvltype,_qslvl,_qlog,_qmsg,_qvar1,_qvar2`), `MAXQUESTS 24`; `QuestData` spawn-randomization fields `_qdrnd`, `isSinglePlayerOnly` (`:117-129`); pools initialized from `DungeonSeeds[15]` when `randomizeQuests` on (`Source/quests.cpp:90-93`); **spawn disables all quests** (`Source/quests.cpp:96-100` `if (gbIsSpawn) for(q) q._qactive = QUEST_NOTAVAIL`); per-quest var inits at `:102-110`; kill hook `CheckQuestKill:227`, state broadcast `ResyncQuests:397` (re-synced on quest network messages at `:610`)
- **XP curve**: `txtdata\Experience.tsv` (Level→XP: 1→2000, 2→4620, …) loaded by `Source/tables/playerdat.cpp:91` into `ExperienceData`
- **Spawn gating**: monster availability `MonsterAvailability::Retail` filtered at `Source/monster.cpp:3159`; Skeleton King set-level gate `Source/player.cpp:3479`; Diablo end-game gate `:3688-3701`
- **Hellfire**: `gbIsHellfire` branches across 135 lines in root cpp (e.g. armor-pierce semantics above, Spectral Arrow off-by-1 fix `Source/player.cpp:559-562`, 25 levels vs 17, 24 quests vs 16, premium items 15 vs 6 — `Source/loadsave.cpp:2772-2781`)

### Data structures (example: damage pipeline)
```
melee: dam = rand[min..max]                     // whole points
       dam += dam * _pIBonusDam / 100
       dam += _pIBonusDamMod
       dam2 = dam << 6                           // snapshot for Peril
       dam += _pDamageMod                        // str-derived bonus (whole)
       [class crit ×2 | undead/mace ×1.5 | ...]
       dam <<= 6                                // → 1/64 units
       [Jesters r∈[0,200] warp | Peril self-dmg dam2+GetHit<<6]
       ApplyPlrDamage/DamageMonster(dam)
```

### Algorithms (item generation sketch — from items.cpp flow)
1. Roll base index by dungeon level bands (`GenerateRnd` over `AllItemsList` filtered by `iMinMLvl`/spawn availability).
2. Roll quality: unique → check `UniqueItem` availability + `_iSeed`; else prefix/suffix chance by `PLMinLvl ≤ ilvl` and `PLChance` weights; curses possible (`PLGOE == Evil`, `PLOk == false`).
3. Apply powers (`SaveItemAffix`) clamping value to `[minVal, maxVal]`, multiply price by `multVal`.
4. Store `_iSeed` + `_iCreateInfo` so the item can be regenerated on load (this is why items store seeds, not stats, in saves — §5).

### Determinism constraints
- Every roll order matters (hit before damage before affixes); the `GenerateRnd` call sequence per action is parity-relevant.
- Damage is 1/64 fixed point after `<<6` — never floats.

### Go/Ebiten mapping decision
- Pure-Go package `gameplay` with subpackages `combat/spells/items/quests/stores`; tables loaded from the same TSVs into generated Go structs (a small `txtdata` codegen tool mirrors `PLStruct` etc.).
- Function-pointer tables (missile add/process) become registry maps `map[MissileID]struct{Add func(*Missile,*Params); Process func(*Missile)}` initialized in `init()` — same shape as `Source/tables/misdat.h:167-168`.

### Gaps/risks
- `stores.cpp` mixes UI + logic (2911 lines); the Go split must extract `store_transaction` logic (already test-separated in C++: `test/store_transaction_test.cpp`, `test/visual_store_test.cpp`, `test/vendor_test.cpp`) from widget code.
- Hellfire vs Diablo branch density is the #1 parity hazard; mandate a `dgame.Hellfire()` helper wrapping every divergence.

### Effort: XL

### Open questions
- Port all 24 quests up-front or Diablo-16 first then Hellfire-8? (Contract: milestones M2 = Diablo full, M4 = Hellfire; so 16 first, but keep `MAXQUESTS 24` arrays from day one.)

---

## 5. Persistence

### Purpose
Hero/game/stash persistence: legacy `.sv` byte-compat, pack structs, codec/SHA1 encryption, plus the new SQLite-native store.

### Key sources (file:line)
- `Source/pack.h:20-30` (`ItemPack`: iSeed u32, iCreateInfo u16, idx u16, bId/bDur/bMDur/bCh/bMCh u8, wValue u16, dwBuff u32), `:33-88` (`PlayerPack`: timestamps, dest params, pos, name[32], class, base stats, level, statPts, experience u32, gold, HP/Mana bases, `pSplLvl[37]` "set to 37 to make save games compatible", `pMemSpells` u64, ItemPacks for body/inv, `InvGrid[40]`, …), `:90+` `PlayerNetPack`; API `PackItem/UnPackItem/PackNetPlayer/UnPackNetPlayer` (`Source/pack.h:151-182`)
- `Source/loadsave.cpp:2763-2900` `SaveGameData` — order: magic `RETL|SHAR|HELF|SHLF` (LE32 of chars), level counts (`giNumberOfLevels = 25|17`, quests `24|16`, premium `15|6`), `setlevel/setlvlnum/currlevel/leveltype` (BE), view pos (BE), panel flags (LE), active counts (BE), per-level `DungeonSeeds[i]` + level types (BE), `SavePlayer`, quests, portals, `MonsterKillCounts` (+padding to `MaxMonsters`), then per-level `ActiveMonsters`/monster states/missiles (capped `MaxMissilesForSaveGame = 125`, `Source/loadsave.cpp:55`)/objects/lights/visions; `SaveLevel` at `:1929` with `SaveMonster` per active monster at `:1960` (monster serializer `:1503`), seeds at `:1902-1927`; `giNumberOfLevels` global `:51` with load loops `:543-546`
- `Source/loadsave.cpp:2008` `LoadLevel` (+ conversion path `pfile_convert_levels` `:2148-2169`)
- **Codec** (`Source/codec.cpp:92-136` `codec_decode`, `:137+` `codec_encode`): block cipher chaining Diablo-"SHA1" keystream: per 32-byte block `SHA1Result(ctx,dst); XOR dst⊕buf; SHA1Calculate(ctx,buf)`; trailer `{checksum u32, error u8, lastChunkSize u8}`; `codec_get_encoded_len` pads to 32-byte blocks + signature; key init from password via 3-round SHA-ingest (`CodecInitKey`)
- **SHA** (`Source/sha.cpp:8-25`): "Diablo's SHA1 differs from actual SHA1" (custom circular shift & schedule — port verbatim)
- **pfile** (`Source/pfile.cpp:60-97`): passwords & `GetSavePath` (`spawn_/share_/single_/multi_` prefixes by spawn/multi flags, `.sv` ext, `UNPACKED_SAVES` dir variant), `GetStashSavePath`; hero container entry `"hero"` written at `:178`; API list at `Source/pfile.h:100-128` (`pfile_write_hero`, `pfile_compare_hero_demo`, `pfile_ui_set_hero_infos`, `pfile_read_player_from_save`, …)
- Stash save format: page-list u32 + per-page 10×10 u16 grid + item list (`Source/loadsave.cpp:2740-2760`)
- `MAX_CHARACTERS 99` (`Source/pfile.h:24`)
- Tests: `test/pack_test.cpp:441-868` (UnPackItem for diablo/unique-bug/spawn/multi/hellfire/strip/empty/gold sizes), `test/codec_test.cpp`, `test/writehero_test.cpp`, `test/stash_test.cpp`

### Data structures (SQLite schema proposal)
```sql
PRAGMA user_version = 1;              -- schema version, bump + migrate on change
CREATE TABLE meta      (key TEXT PRIMARY KEY, value BLOB);      -- engine version, game flags
CREATE TABLE heroes (
  id INTEGER PRIMARY KEY, slot INTEGER UNIQUE,        -- 0..98
  name TEXT, class_id INTEGER, created_at INTEGER, updated_at INTEGER,
  pack BLOB NOT NULL,                                  -- PlayerPack bytes (legacy layout)
  spawn INTEGER, hellfire INTEGER, multiplayer INTEGER
);
CREATE TABLE games (                                   -- one row per hero per game
  hero_id INTEGER REFERENCES heroes(id), kind TEXT,    -- 'single'|'multi'|'spawn'...
  data BLOB NOT NULL,                                  -- SaveGameData stream (un-codeced)
  updated_at INTEGER, PRIMARY KEY (hero_id, kind)
);
CREATE TABLE levels (hero_id INTEGER, kind TEXT, level INTEGER,
  seed INTEGER, type INTEGER, blob BLOB,               -- SaveLevel payload
  PRIMARY KEY (hero_id, kind, level));
CREATE TABLE stash (hero_id INTEGER PRIMARY KEY, page_count INTEGER, blob BLOB);
CREATE TABLE settings (key TEXT PRIMARY KEY, value TEXT);
```
Rationale: blobs over normalization — the in-memory structs ARE the format; normalization would create a second source of truth. Versioning: `user_version` + a Go `migrate()` chain; hero `pack` column keeps legacy byte layout so `.sv` export is a memcpy.

### Algorithms
- Save cycle: pack player → write "hero" + "game"/level streams into an in-memory MPQ-ish container → codec-encrypt with mode password → write `single_N.sv`. The Go writer reuses the MPQ writer from go-mpq or writes the legacy container manually (devilutionX uses `SaveWriter`/`mpq_writer` — `Source/mpq/mpq_writer.cpp`). Hero entry is codec-encoded with `pfile_get_password()` at write (`Source/pfile.cpp:177`) and decoded symmetrically at read (`:628`).
- Demo reference compares hero bytes: `pfile_compare_hero_demo` (used by timedemo test as oracle, `test/timedemo_test.cpp:96-100`).

### Determinism constraints
- Save/load must round-trip bit-exactly (the C++ tests assert bytes); missile cap 125 on save, ordering = active list order.
- Codec is password- and mode-dependent — wrong password prefix yields garbage, so mode flags must be persisted *outside* the encrypted blob (they are: filename prefix).

### Go/Ebiten mapping decision
- `modernc.org/sqlite` behind a `Store` interface with two impls: `SQLiteStore` (native) and `LegacySVStore` (byte-compat `.sv` import/export). Both share the same pack/loadsave serialization code so parity is testable: `TestRoundTrip(save→sqlite→load→save) == identical bytes`.
- Codec/SHA1 hand-ported with C++ test vectors (fixtures in `test/codec_test.cpp`).

### Gaps/risks
- The `SaveWriter` MPQ write path (hash tables, block tables, best-sector packing) must be replicated for `.sv` export; go-mpq's *write* side compresses with zlib only and cannot encrypt — acceptable: legacy containers need plain STORE sectors + encryption of the codec layer only. Verify with a hex-diff against a C++-written save.
- `UNPACKED_SAVES` (directory mode) is a debug variant — support read-only.

### Effort: L

### Open questions
- Should multiplayer games also live in SQLite per-hero or a separate `multi.db`? (Default: same file, `kind` column distinguishes.)

---

## 6. Asset pipeline

### Purpose
Everything between MPQ bytes and renderable/playable data: archive stack, CEL/CL2/CLX/PCX decode, palettes + TRN, fonts, txtdata TSV, sound effects and music.

### Key sources (file:line)
- **Load order**: `Source/engine/assets.cpp:505-517` `LoadCoreArchives` (devilutionx 9000, fonts 9200 — fonts loaded early for error dialogs), `:521-527` `LoadLanguageArchive` (code-named MPQ at 9100), `:530-568` `LoadGameArchives` (`DIABDAT` uppercase `.MPQ` first for CD/GOG, then lowercase, then `spawn`; hfbard 8110 / hfbarb 8120), `:570-586` `LoadHellfireArchives` (hellfire 8000; hfmonk 8100 `mhaas.clx` probe, hfmusic 8200 `dlvlf.wav/mp3` probe, hfvoice 8500 `cowsut1` probe), `:704-760` `LoadModArchives`; search paths `GetMPQSearchPaths` `:441`; priorities `Source/engine/assets.hpp:309-312`
- **MPQ layer**: `Source/mpq/mpq_reader.cpp` wraps mpqfs (`mpqfs_open`/read/close via `mpqfs.h`); decompression methods implemented: PKWARE DCL 0x08 → `pk_explode_sector`, zlib 0x02, bzip2 0x10, anything else → `MPQFS_ERR_UNSUPPORTED_COMPRESSION` (`build-v2/_deps/mpqfs-src/src/mpq_stream.c:160-231`); encryption/decryption tables in `mpq_crypto.c`; **no Huffman** anywhere in the stack; `HaveMainData()` helper (`Source/engine/assets.hpp:344`) gates tests/UI on diabdat-or-spawn presence
- **Go option**: `github.com/suprsokr/go-mpq` v0.3.1 — MIT, pure Go, MPQ v1+v2 read+write, "PKWare DCL (0x08) ✅ read; Zlib ✅; BZip2 ✅; Huffman ❌ audio files only, not implemented" — https://pkg.go.dev/github.com/suprsokr/go-mpq (Compression Support + Limitations). Alternative `github.com/OpenDiablo2/mpq` is GPL-3.0 — license-incompatible with a permissive Go port preference, rejected.
- **CLX format** (devilutionX runtime sprite format): frame header 6 bytes `{u16 headerSize, u16 width, u16 height}`, pixel data = CL2 RLE — `Source/engine/clx_sprite.hpp:15-27`; sheet = u32 LE frame-count + offsets table (iterators at `:99-160`, `ClxSpriteListOrSheet` `:462`); loader `Source/engine/load_clx.cpp:27-45` (with optional on-load `CelToClx` conversion `:36-39`)
- **CL2 pixel RLE**: control byte `c`: `1..63` → copy c literal pixels; `64..127` → repeat next pixel `c-64`; `129..255` → RLE-run of `257-c` of next byte; `128` = end-of-line — per header comment referencing https://github.com/savagesteel/d1-file-formats/blob/master/PC-Mac/CL2.md (`Source/engine/clx_sprite.hpp:25`)
- **CEL**: loaded via width-hint `LoadCelListOrSheet(name, widthOrWidths)` (`Source/engine/load_cel.cpp:27-46`), converted to CLX in-memory by `CelToClx` (`Source/utils/cel_to_clx.*`)
- **PCX**: `Source/engine/load_pcx.cpp` (8-bit paletted, used for ui_art backgrounds, e.g. `gendata/*.pcx`)
- **TRN**: 256-byte palette translation tables applied at draw (`ClxDrawTRN`), loaded `Source/engine/trn.cpp`; unique monsters carry `uniqueMonsterTRN` (`Source/monster.h:213`)
- **Palettes**: per-level `.pal` files (`LoadPaletteAndInitBlending`, `Source/levels/setmaps.cpp:126`), 256 entries; `LightTables` 16 levels of 256 (`Source/engine/light_tables.cpp`); special tables `InfravisionTable/StoneTable/PauseTable` (`Source/lighting.h:42-45`); color cycling `palette_update_caves/hive/crypt` (`Source/diablo.cpp:3513-3521`)
- **Fonts**: `DiabloUI` + `engine/render/text_render.cpp`; devilutionx.mpq `fonts/` (+optional fonts.mpq 9200)
- **txtdata TSV**: D2-style tab-separated format (`assets/txtdata/Readme.md` — "values cannot contain tab characters"; MAY start UTF-8 BOM; header record SHOULD be first); parser `Source/data/file.hpp:37-110` (`DataFile::load`, `parseHeader` with column mapping) + `record_reader`/`value_reader`; loaders per domain, e.g. monsters `Source/tables/monstdat.cpp:409` (`txtdata\monsters\monstdat.tsv`, 113 lines), unique monsters `:459`, items/affixes, spells, missiles (`txtdata\missiles\misdat.tsv` + `missile_sprites.tsv`), sounds (`txtdata\sound\effects.tsv` ~1000 rows: `Walk Misc sfx\misc\walk1.wav`), towners, quests dialog matrix (`txtdata\towners\quest_dialog.tsv`), experience
- **Sound**: init `Aulib::init(sampleRate option, AUDIO_S16, channels, buffer)` (`Source/engine/sound.cpp:281`); default sample rate **22050** (`#define DEFAULT_AUDIO_SAMPLE_RATE 22050`, `Source/options.cpp:55`; choices 22050/44100/48000 at `:520`); music = WAV with automatic `.mp3` fallback swap (`Source/engine/sound.cpp:61-64` `mp3Path`), track tables `SpawnMusicTracks` `:143-146` (`music\stowne.wav`…); sfx `PlaySFX/PlaySfxLoc` (`Source/effects.cpp:208,217`, table-driven `sgSFX[]` TSnd pointers `:190,226`, position-attenuated `PlaySfxPriv` `:214`); TSnd struct `Source/engine/sound.h:39`; sfx catalog from `txtdata\sound\effects.tsv` (~1000 rows)

### Data structures
```go
// assets/clx.go
type ClxFrame struct { W, H uint16; Pixels []byte }  // palette indices, CL2-decoded
type ClxSheet []ClxFrame                             // or grouped lists for sheets
type Palette [256]color.RGBA
type TRN [256]uint8                                  // src index → dst index
```

### Algorithms
- **Archive stack**: ordered map priority→archive; lookup walks priorities descending (highest first) — replicate as `[]struct{prio int; arch *mpq.Archive}` sorted desc + `OpenAsset(name)` returning bytes (`Source/engine/assets.cpp:385-437` LoadMPQ/FindMPQ semantics).
- **CL2 decode loop** (per frame, top-down rows):
  ```
  for len(src) > 0 && y < h:
      c := src[0]; src = src[1:]
      switch {
      case c == 0:            continue          // some builds: end marker variants
      case c <= 63:           emit literals src[:c]; src = src[c:]
      case c >= 64 && c <=127:repeat src[0] c-64 times; src = src[1:]
      case c == 128:          y++ (row end)
      default:                repeat src[0] (257-c) times; src = src[1:]
      }
  ```
  (Exact branch table cross-checked against `blit_impl.hpp` during implementation; unit-test with fixture frames extracted by a debug dump from the C++ build.)
- **Palette→Ebiten**: decode each CLX frame to indices once; build one `*ebiten.Image` per (frame, palette-instance) at load: base image from main palette; per-TRN variants created by re-palettizing a copy (ColorScale cannot express arbitrary index remaps). Light levels handled at draw time via pre-multiplied shade images or `ColorScale` gray-scale multiply (16 light tables are monotone → `ColorScale{R: s, G: s, B: s}` with per-light `s` constants is a faithful, cheap approximation for most content; TRN tables like infravision remain index-based). Decision: **re-palettize for TRN, ColorScale for light** — verifiable against `LightTables` monotonicity.
- **txtdata**: streaming TSV reader (BOM-tolerant, header-driven, quoted fields per Readme EBNF) → typed structs via reflection or codegen; mirrors `DataFile` error semantics (`reportFatalFieldError`).

### Determinism constraints
- Asset bytes are inputs, not outputs — but derived decode must be stable: never sort map iterations when building atlases.

### Go/Ebiten mapping decision
- `assets` package: `FS`-like `Open(name string) ([]byte, error)` over the MPQ stack; `devilutionx.mpq` equivalent loaded from the vendored `assets/` tree (go:embed of the pre-built MPQ is 100+ MB — instead embed the raw tree and serve unpacked paths, matching devilutionX's `UNPACKED_MPQS` mode which it fully supports, `Source/engine/assets.cpp:551-558`).
- Font: hand-rolled bitmap font renderer over the CLX font sheets (devilutionX does the same — `text_render.cpp`), not Go fonts.

### Gaps/risks
- **go-mpq extraction smoke test is mandatory**: enumerate every file in diabdat/spawn/hellfire and decode all; any Huffman-marked sector would surface here. mpqfs's absence of Huffman and its ability to run the full game strongly implies zero usage, but prove it.
- CL2 branch semantics differ subtly between sources; freeze with fixtures.
- Music `.wav→.mp3` fallback: Ebiten audio decodes both (`/home/blaze/.agents/skills/ebiten/audio/SKILL.md` decoder table); resample 22050→device rate with `audio.Context` at 22050 to avoid double resampling, or upmix once at load. MP3s in retail MPQs are devilutionX additions (not in diabdat) — only needed if we ship them.

### Effort: XL

### Open questions
- Embed `devilutionx.mpq` binary or the unpacked `assets/` tree? (Default: unpacked tree via go:embed; smaller tooling, no MPQ writer needed at boot.)
- Do we ever need MPQ *writing* in Go beyond `.sv` export? (Only if mods-by-mpq is a goal — defer.)

---

## 7. Rendering

### Purpose
World rendering order, isometric math, lighting, transparency, automap, panels, text, cursor, scaling, cutscenes.

### Key sources (file:line)
- `Source/engine/render/scrollrt.cpp:1229` `DrawGame(fullOut, position, offset)`; `:949-988` `DrawFloor` (row walk: after each row `y += TILE_HEIGHT/2`, odd rows `x++/cols--/bufX += TILE_WIDTH/2`, even rows `y++/cols++/bufX -= TILE_WIDTH/2` — the classic isometric diamond walk), `:992+` `DrawTileContent` (`rows += MicroTileLen` because tall walls span micro rows), `:1000-1036` wall-behind pre-pass (`IsWall(tile) && IsWall(E|W neighbor) && walkable behind ⇒ draw tile+East first, skipNext`)
- **Per-tile draw order** (`DrawDungeon`, `Source/engine/render/scrollrt.cpp:790-980`): missiles(pre) → corpses (`dCorpse`, unique-TRN aware) → objects with `_oPreFlag` → items with `!_iPostDraw` → dead players → players (moving south/east players drawn via *negative* `dPlayer` ids so tile order paints them from the tile they leave, `:830-866`) → monsters → items `_iPostDraw` → objects (non-pre) → missiles → arch/special `dSpecial` (blended if `TransList[bMap]`, per-pixel-lightmap variant with `Lightmap::bleedUp`, `:907-943`; town tree leaves deferred one row, `:944-952`)
- **Missile render interpolation**: `UpdateMissilePositionForRendering` advances `tileForRendering/offsetForRendering` by `progress` fraction of fixed-point velocity with collision-aware clamping (`Source/engine/render/scrollrt.cpp:125-193`)
- **Lighting at draw**: `ClxDrawLight(out, pos, clx, lightTableIndex)` — `LightTables[idx]` as TRN, `ClxDrawLightBlended` for transparency (`Source/engine/render/scrollrt.cpp:249-258`); per-pixel option `Lightmap` path `:912-930`
- **Automap**: `Source/automap.cpp:1752` `DrawAutomap` (tile double-loop `map.x+j, map.y-j`, `:1847-1855`), `DrawAutomapTile:980`, `DrawAutomapPlr:1341`, `DrawAutomapText:1427`; `AutomapActive` `:1548`, `InitAutomapOnce:1556`, `StartAutomap:1702`; mode enum `AutomapType` + `CurrentAutomapType` (`Source/automap.h:83,91`); mode handling (minimap rect / transparent) at `Source/engine/render/automap_render.cpp:168-171`
- **Panels**: `Source/panels/` — `mainpanel.cpp` (control panel + orbs), `charpanel.cpp`, `spell_book.cpp`, `spell_list.cpp`, `spell_icons.cpp`, `quest_log.cpp`, `partypanel.cpp`, `info_box.cpp`, `console.cpp`; panel drawing primitives in `Source/control/control_panel.cpp:351` (`DrawPanelBox`) and button rects at `:435-445`; flasks in `Source/control/control_flasks.cpp`; belt/life/mana redraw in the `Source/control/` family; test `test/panel_state_test.cpp`
- **Text**: `Source/engine/render/text_render.cpp` (CLX glyph blitting, word-wrap, `DrawStringFormatArg` machinery at `:291-359`, `draw_color` variants) — integration-tested in `test/text_render_integration_test.cpp`
- **Cursor**: `ShouldShowCursor` (`Source/engine/render/scrollrt.cpp:236-247` — hidden on gamepad unless menus/teleport), software cursor blit `BlitCursor:214`, hardware cursor `Source/hwcursor.cpp:112` `SetHardwareCursorFromSurface`
- **Screen/scaling**: logical 640×480 (or 1920×1080) surface scaled to window; `InitBackbufferState`/`backbuffer_state.cpp`; `AdjustToScreenGeometry(forceResolution)` used by tests (`test/timedemo_test.cpp:89`)
- **Movies**: `Source/movie.cpp:37` `play_movie`, in-game `PlayInGameMovie:130`; SMK decode in `Source/storm/storm_svid.cpp` (frame-rate note `:406`)
- Color cycling palettes (see §6) animate after ticks, not frames (`Source/diablo.cpp:947-948` call site)
- Low-level tile blitters: `Source/engine/render/dun_render.cpp:908` `RenderTileType` (templated per `TileType` × light mode), `:991` `RenderTileDispatch` (per-pixel-light dispatch at `:994`), CLX sprite blits `Source/engine/render/clx_render.cpp:592` (`ClxDrawTRN` with TRN tables), backbuffer bookkeeping `Source/engine/backbuffer_state.cpp:21-28` (`BackbufferState`)

### Algorithms (isometric mapping)
```
screenX = (tile.x - tile.y) * TILE_WIDTH/2 + offset.x      // TILE 64×32
screenY = (tile.x + tile.y) * TILE_HEIGHT/2 + offset.y
```
Row iteration as in `DrawFloor` (§ above) — implement once, share between floor and content passes. The canonical transforms live on `Displacement::worldToScreen` / `screenToWorld` ("inverse matrix of the worldToScreen transformation", `Source/engine/displacement.hpp:151-164`) — port these two as the single source of truth for all screen math.

### Determinism constraints
- Rendering is *derived* (pure function of world + ProgressToNextGameTick) — no logic in Draw. Golden-frame tests hash the final `*ebiten.Image` pixels at fixed interpolation values.

### Go/Ebiten mapping decision
- Offscreen `*ebiten.Image` per frame region at native logical resolution, then one `DrawImage` with scaling options to screen (single draw command per https://ebitengine.org/en/documents/performancetips.html). Tiles/sprites pre-decoded to a texture atlas; per-frame `SubImage` + `DrawImage` with shared `*DrawImageOptions` to batch.
- Draw order mirrors `DrawDungeon` exactly; players-in-motion handled by the same negative-id tile trick using a Go `dPlayer` grid.
- Lighting: pre-render 16 shaded variants per sprite is memory-heavy (sprite count × 16); choose per-draw `ColorScale` for monotone light + per-pixel composite only for `Lightmap` mode parity (optional flag).
- Automap: separate `Draw` layer, vector-ish lines via `ebiten/vector` or small tile stamps from `automap` CLX.
- Movies: **defer** — define a `CutscenePlayer` interface with a "skip" stub; optionally use an existing Go SMK decoder if license-clean, else encode cutscenes once to `.webm`/frame dumps. Decision criterion: does M2 need intro/endings? Default: still images + text for M2, full SMK for M4 or never.

### Gaps/risks
- Transparency blending (`ClxDrawLightBlended` = 50% blend of sprite with TRN-light) must be replicated with `DrawImageOptions.ColorScale` alpha ≈ composite; verify against screenshots.
- Per-pixel lightmap mode (new in this fork line) is optional; skipping it changes visuals only, not logic.

### Effort: XL

### Open questions
- Logical resolution: fixed 640×480 (retail feel, crisp integer scale) vs dynamic `Layout` returning window size (modern). Default: fixed 640×480 with integer scaling option (matches assets natively).

---

## 8. Networking

### Purpose
Lockstep multiplayer: packet framing + encryption, the dvlnet protocol zoo, the CMD_* application layer, delta compression, timing sync, join/leave; and the CMD system's second life as the demo/replay format.

### Key sources (file:line)
- **Packet types** (`Source/dvlnet/packet.h:27-40`): `PT_MESSAGE 0x01, PT_TURN 0x02, PT_JOIN_REQUEST 0x11, PT_JOIN_ACCEPT 0x12, PT_CONNECT 0x13, PT_DISCONNECT 0x14, PT_INFO_REQUEST 0x21, PT_INFO_REPLY 0x22, PT_ECHO_REQUEST 0x31, PT_ECHO_REPLY 0x32`; `turn_t {seq_t SequenceNumber; int32_t Value}` (`:56-59`); encryption = NaCl `crypto_secretbox` (nonce‖MAC‖ciphertext; `Source/dvlnet/packet.cpp:225-240` with `crypto_secretbox_open_easy`, key from password via `key_t`/`GenerateCookie` `:213`); plaintext mode when `NOZIP`/stub `key_t = uint8_t` (`:48-55`)
- **base** (`Source/dvlnet/base.h:19-100+`): SNet* surface (`SNetSendMessage/ReceiveMessage/SendTurn/ReceiveTurns/GetTurnsInTransit/DropPlayer/GetProviderCaps`), per-player `PlayerState{isConnected, turnQueue deque, lastTurnValue, roundTripLatency}`, `current_turn/next_turn` seq_t, `SendTurnIfReady`, `message_t{sender,payload}` queue; `game_init_info` buffer for join-in-progress
- **Protocol zoo**: `base_protocol.h` (templated UDP-ish transport), `tcp_client.cpp`/`tcp_server.cpp` (asio; server: `acceptor(ioc, ep)` `Source/dvlnet/tcp_server.cpp:14-22`, `connections[MAX_PLRS]`, `NextFree()` `:45`, broadcast/route logic `:154-180`), `loopback.cpp` (single-process), `cdwrap.cpp` (provider cdict wrapper), `zerotier_lwip.cpp`/`zerotier_native.cpp` + `protocol_zt.cpp` (libzt virtual NIC — **cut**), `frame_queue.cpp` (length-prefixed framing)
- **Application commands** (`Source/msg.h:18-428`): `_cmd_id : uint8_t` — ~86 CMD_* values, e.g. `CMD_STAND(TCmd)`, `CMD_WALKXY(TCmdLoc)`, `CMD_ACK_PLRINFO(TCmdPlrInfoHdr)`, `CMD_ADDSTR/ADDAG/ADDDEX/ADDVIT(TCmdParam1 int16 delta)`, `CMD_GETITEM/AGETITEM/PUTITEM/SPAWNITEM(TCmdGItem/TCmdPItem)`, `CMD_RATTACKXY(TCmdLoc)`, `CMD_SPELLXY(TCmdLocParam4 x,y,spellID,spellType,spellLevel,spellFrom)`, `CMD_OPOBJXY/DISARMXY(TCmdLoc)`, `CMD_ATTACKID/ATTACKPID/RATTACKID/RATTACKPID(TCmdParam1)`, … `CMD_STRING(80-char max, MAX_SEND_STR_LEN 80 msg.h:22)`, … `CMD_OPENGRAVE`, `CMD_SPAWNMONSTER`, sentinels `FAKE_CMD_SETID/DROPID`, `CMD_INVALID 0xFF` (`:428`)
- **Command payload structs** (`Source/msg.h:432-470+`): `TCmd{bCmd}`, `TCmdLoc{bCmd,x,y}`, `TCmdLocParam1..4` (u16 params), `TCmdGItem` (`:560-579`: loc + item union + `bMaster/bPnum/bCursitem/bLevel/dwTime`), `TCmdPItem` (`:581-600`, with `FloorItem=CMD_STAND`/`PickedUpItem=CMD_WALKXY` sentinel aliases for delta compression), `TItem` (`:534-547` wIndx,wCI,dwSeed,bId,bDur,bMDur,bCh,bMCh,wValue,dwBuff,wToHit,wMaxDam), `TEar` (`:549-555` heroname[17])
- **Message packets**: `TPktHdr` (`Source/msg.h:694-709`: px,py,targx,targy,php,pmhp,mana,maxmana,bstr,bmag,bdex,pdir,wCheck=HeaderCheckVal,wLen) + `TPkt.body[493]`; megapacket buffer `TMegaPkt{spaceLeft; byte data[32000]}` (`Source/msg.cpp:207-212`); dispatch `ParseCmd` (`Source/msg.cpp:3341`, per-player dispatch loop at `:464`) → per-command `On_*` handlers (the ~2000 lines of msg.cpp); senders `NetSendCmd/NetSendCmdLoc/NetSendCmdLocParam1..4` (`Source/msg.cpp:2982-3066`), `NetSendCmdGItem2/Req2/Extra` (`:1152-1197`)
- **Delta compression** (`Source/msg.cpp:218-275`): `DLevel{TCmdPItem item[MAXITEMS]; map<pos,DObjectStr> object; map<size_t,DSpawnedMonster> spawnedMonsters; DMonsterStr monster[MaxMonsters]}`, `DMonsterStr{position,menemy,mactive,hp,mWhoHit}`, `LocalLevel` automap save `automapsv[DMAXX][DMAXY]`, `DPortal`, `MultiQuests{qstate,qlog,qvar1,qvar2,qmsg}`, `DJunk{portal[MAXPORTAL],quests[MAXQUESTS]}`; export/import `DeltaExportData` (`Source/msg.cpp:698` call site), `delta_kill_monster`, `DeltaSyncJunk`, `DeltaClearLevel` (`Source/msg.h:725-731`)
- **Sync/timing**: `multi_handle_delta()` consumed by `game_loop` (`Source/diablo.cpp:3496`); turn cadence from §1 (`nthread_recv_turns`); `ProcessGameMessagePackets` (`Source/multi.cpp:729-760`: header validation `wCheck`/`wLen`, trust gating `IsNetPlayerValid`, buffering during `gbBufferMsgs`); `run_delta_info` (`Source/msg.cpp:2763`)
- **Join/leave**: join handshake = `PT_JOIN_REQUEST(cookie)` → server replies `PT_JOIN_ACCEPT(+game_init_info)`; player info exchange `CMD_ACK_PLRINFO`/`TPCmdPlrInfoHdr`; drop via `SNetDropPlayer` + `msg_send_drop_pkt` (`Source/msg.h:718`)

### Algorithms (lockstep turn loop)
1. Each logic tick, local client appends its cmds to the turn's `turn_t.Value` batch (commands serialized since last turn).
2. `nthread_send_and_recv_turn` keeps `gdwTurnsInTransit` turns in flight; `nthread_recv_turns` fires every `sgbNetUpdateRate` ticks; when all 4 players' turns for `current_turn` are present, `multi_handle_delta` applies all `TPkt` bodies in deterministic (player-id) order and advances `current_turn`.
3. Deltas sync world drift: monster hp/position, object states, ground items (`TCmdPItem` with sentinel cmd-aliasing so a 3-byte `TCmd` can mean "clear item slot"), quests, portals.

### Determinism constraints
- Commands are applied in strict per-tick, per-player order; the CMD layer *is* the input tape — hence demo files record the same stream (`Source/engine/demomode.cpp`).
- `TPktHdr` mirrors player vitals for UI smoothing only — must not feed logic.

### Go/Ebiten mapping decision
- **Keep**: TCP client/server (goroutines + `net` package, no asio), loopback, cdwrap-equivalent provider selection, packet format, NaCl secretbox via `golang.org/x/crypto/nacl/secretbox` (byte-compatible), frame queue, the full CMD enum + payload structs (generated `encoding/binary`-free manual marshal to preserve exact widths/order), delta structs, turn logic.
- **Cut**: ZeroTier/lwip (`zerotier_*`, `protocol_zt`), UDP base_protocol (TCP only; devilutionX itself defaults TCP for direct IP).
- Server = `tcp_server` port: listener goroutine, `connections[4]`, drop/rejoin, `game_init_info` replay. Client = reconnect with cookie.
- **Demo format reuse**: `internal/demo` writes `{event type, mod, payload, ProgressToNextGameTick|0x80}` records (mirroring `Source/engine/demomode.cpp:595-599`) — same parser for `--demo` playback and tests.

### Gaps/risks
- Byte-exact packet layouts are security+parity sensitive; write a fuzz test comparing Go marshal against C++ hexdumps.
- `std::deque` turn buffering semantics under packet reordering must match (frame_queue handles fragmentation).

### Effort: L (single-player first) → XL (full MP)

### Open questions
- UDP base_protocol for LAN latency: needed? (Default: no — TCP only, revisit after M3.)
- Cross-play with C++ devilutionX peers: goal or not? (If yes, packet compat is non-negotiable and echo/timing fields must match exactly; default: Go-to-Go only, same wire format anyway.)

---

## 9. Input

### Purpose
Keyboard/mouse/gamepad/touch, virtual cursor gamepad mode, keymapper/padmapper, control-mode switching, and this fork's stick turn/walk behavior.

### Key sources (file:line)
- `Source/controls/` modules: `controller.cpp/h` (device abstraction), `controller_motion.cpp:35` `ScaleJoystickAxes` (radial deadzone per "thumbstick dead zones right"), `:129-146` `ScaleJoysticks` (left/right + per-device deadzones), `:192-205` motion event pump; `axis_direction.h` (`AxisDirection`, `AxisDirectionX/Y`), `game_controls.cpp` (gamepad→action mapping), `plrctrls.cpp` (virtual cursor + action state machine), `menu_controls.cpp`, `modifier_hints.cpp`, `touch/` (virtual gamepad), `devices/` (keyboard/joystick), `remap_keyboard.h`
- **Keymapper**: `Source/controls/keymapper.hpp` (rebindable actions with defaults, `KeymapperPress/Release` exported `:17-18`), `padmapper.hpp` (gamepad binds)
- **Fork-specific (HEAD `ca91d778f`)**: `Movement()` — `GetAnalogStickDirection(leftStickX, leftStickY)` → `FaceDir[dir.x][dir.y]`; `switch (GetAnalogStickPush(...)) { case Firm: WalkInDir; case Weak: TurnToDir; case None: }` (`Source/controls/plrctrls.cpp:1861-1882` post-commit); `TurnToDir(Player&, Direction)` (`:1776-1782`: skip if walking/can't change action; `_pdir = dir; StartStand(player, dir)`); `IsStickMovementSignificant()` now `GetAnalogStickPush(...) == StickPush::Firm || rightStick != 0` (`:1916-1920`); `Interact()` fake-attack no longer requires motion — uses `_pdir` when standing (`:557-563`); helper unification in `controller_motion.h` (`GetLeftStickDirection` merge, +15 lines) — from `git show ca91d778f` (diff hunks at plrctrls.cpp:557, :1781, :1861, :1903)
- Control mode: `Source/controls/control_mode.hpp` (`ControlTypes::KeyboardAndMouse` used by `ShouldShowCursor`, `Source/engine/render/scrollrt.cpp:228`)
- Per-frame hooks: `plrctrls_every_frame` (`Source/controls/plrctrls.cpp:2305`), `plrctrls_after_check_curs_move:2264`, `plrctrls_after_game_logic:2312` (called at `Source/diablo.cpp:1568`)

### Algorithms
- Virtual cursor: right-stick moves a screen-space cursor; actions (`Interact`, `Attack`, `Spell`) resolve cursor→world tile via the same pick logic as mouse (`CheckCursMove`);
- Deadzone: scaled radial freedom — magnitude below deadzone → 0, above → remapped `((m-dz)/(1-dz))^2` style curve (per `ScaleJoystickAxes` and the archived third-helix reference in its comment `:38`).

### Determinism constraints
- Input is only an *input* (cursor pos, button edges) — all resolution happens inside `ProcessInput`/player commands, so replays record device-independent cmds (that's why demo mode records SDL events *plus* the cmd stream; Go records cmds only + optional raw events for UI parity).

### Go/Ebiten mapping decision
- Ebiten input APIs (`/home/blaze/.agents/skills/ebiten/input/SKILL.md`): `ebiten.IsKeyPressed`, `inpututil.IsKeyJustPressed/Released`, `ebiten.CursorPosition`, `ebiten.IsMouseButtonPressed` + `inpututil.IsMouseButtonJustPressed`, `ebiten.Wheel()`, `ebiten.StandardGamepadAxisValue` / `IsStandardGamepadButtonPressed`, `inpututil.AppendJustPressedTouchIDs` + `ebiten.TouchPosition`.
- Port `ScaleJoystickAxes` verbatim (float math — safe, input-only). Implement `StickPush{None,Weak,Firm}` thresholds + `TurnToDir` semantics as first-class (it's the fork's headline feature).
- Keymapper: JSON/TOML config in save dir mirroring devilutionX diablo.ini keys; padmapper analogous.

### Gaps/risks
- Touch virtual gamepad (`controls/touch/`) is large; defer to post-M2 unless mobile is a target.
- Gamepad "standard mapping" differs from SDL controller DB on exotic pads — accept standard-only initially.

### Effort: M

### Open questions
- Do we support simultaneous kb+mouse AND gamepad profiles switching automatically (devilutionX `ControlMode` auto-switch on last-used device)? Default: yes, cheap to do in Ebiten.

---

## 10. UI flow

### Purpose
Title → main menu → hero select → game creation → in-game UI (menus, panels, dialogs, toasts, store screens, progress screens).

### Key sources (file:line)
- `Source/DiabloUI/`: `title.cpp`, `mainmenu.cpp` (items vector + `_mainmenu_selections`, `UiMainMenuSelect:34`, dialog build `:69-70` with `UiAddBackground/UiAddLogo`), `hero/selhero.cpp` (class select → name → stats), `multi/selconn.cpp` + `multi/selgame.cpp` (provider/address/game select), `credits.cpp` + `credits_lines.cpp`, `support_lines.cpp`, `selok/selyesno/selstart.cpp` (info dialogs), `settingsmenu.cpp`, `scrollbar.cpp`, `dialogs.cpp` (error/OK boxes), `text_input.hpp`, `button.cpp/h`, `ui_flags.hpp`, `ui_item.h`, `diabloui.cpp` (`UiInitialize:689`)
- Hero info contract with the engine: `_uiheroinfo{saveNumber u32; name[16]; level u8; heroclass; herorank u8; strength/magic/dexterity/vitality u16; hassaved bool; spawned bool}` (`Source/DiabloUI/diabloui.h:60-72`) with callbacks `gfnHeroInfo`/`fninfo`/`fncreate`/`fnremove`/`fnstats` and dialogs `UiSelHeroMultDialog` etc. (`Source/DiabloUI/diabloui.h:81,93`)
- In-game menu: `Source/gmenu.cpp:251` `gmenu_draw` / `:185` `gmenu_draw_pause`, slider API `gmenu_slider_get/steps` (`Source/gmenu.h:101,106`; used at `Source/gamemenu.cpp:274-277` for tick rate 20–50), `Source/gamemenu.cpp` (save/quit/options), `Source/help.cpp:167` `InitHelp` / `:191` `DrawHelp` / `:143` slider, `Source/interfac.cpp` (progress bars `StartProgress`), `Source/diablo_msg.cpp:45,125-130` (`InitDiabloMsg` toast queue with 3500 ms default duration), minitext (`Source/engine/render/text_render.cpp` small-font paths)
- Store screens: `Source/stores.cpp` UI halves (stext* buffers, 20-line store list, buy/sell/repair cursor flows, gossip paging)
- Progress screens on level load: `interface` art + `progress` bar (`Source/DiabloUI/progress.cpp`)
- Tests: `test/game_menu_test.cpp`, `test/inventory_ui_test.cpp`, `test/char_panel_test.cpp`, `test/spell_ui_test.cpp`, `test/ui_test.hpp`

### Data structures
```go
// ui/widget.go — hand-rolled minimal set mirroring DiabloUI primitives
type Widget interface{ Draw(dst *ebiten.Image); HandleInput() }
type Button struct{ Frames []*ebiten.Image; OnClick func(); ... }   // trishdbtn etc.
type MenuList struct{ Items []MenuItem; Selected int }
type Slider struct{ Min, Max, Val int }                              // gmenu_slider port
```

### Algorithms
- UI is a stack of screens: title → mainmenu → selhero → selconn/selgame (MP) → game; each screen owns focus + event routing; `UiInitialize` loads shared art; transitions fade via palette (`PaletteFadeIn(8)` at game start, `Source/diablo.cpp:888`).

### Determinism constraints
- None beyond rendering purity; store dialog *logic* (prices, stock) is in §4.

### Go/Ebiten mapping decision
- **Hand-rolled widgets**, not ebitenui: the Diablo UI is entirely art-driven (CLX buttons/backgrounds, bitmap fonts, fixed layouts); a generic widget lib fights the art. Implement ~6 primitives (Button, MenuList, Slider, Scrollbar, TextInput, Dialog) + a screen stack. ebitenui remains a fallback if text-input IME becomes painful (Open question).

### Gaps/risks
- Mouse/kb navigation parity (arrow keys + Enter + Esc) must work in all screens (devilutionX guarantees it for gamepad too).
- Localization strings change layouts (25 languages) — keep string lookup + wrap-aware text renderer early.

### Effort: L

### Open questions
- ebitenui for settings menu only (scrollable lists, dropdowns) while keeping art screens hand-rolled? (Default: yes for settings only.)

---

## 11. Platform & misc

### Purpose
Options system, translations, Lua scripting, discord RPC, capture, QoL modules, demo mode, spawn gating, headless.

### Key sources (file:line)
- **Options**: `Source/options.h:877` `struct Options` (Graphics/Audio/Gameplay/Controller/Language trees), `:530` `Fullscreen` example, audio `sampleRate` `Source/options.h:512` with values {22050, 44100, 48000} & default `DEFAULT_AUDIO_SAMPLE_RATE 22050` (`Source/options.cpp:55,520`); persistence to `diablo.ini` (`Source/utils/ini.cpp`, test `test/ini_test.cpp`); option-change callbacks (e.g. audio restart `Source/engine/sound.cpp:183-185`)
- **Translations**: 25 `.po` files under `Translations/`; runtime loads `.mo`/`.gmo` (`Source/utils/language.cpp:39-42` extensions list); language MPQ override at priority 9100 (`Source/engine/assets.cpp:521-527`)
- **Lua**: `Source/lua/` 59 files; sol2 bindings; modules: `audio` (`Source/lua/modules/audio.cpp:37-47` exposes `playSfx`, `playSfxLoc`, `SfxID` enum table), `dev/*` (debug: monsters/player/level/quests/towers/items/display/search — `Source/lua/modules/dev/*.cpp`), `i18n`, `render`, `hellfire`, `floatingnumbers`, `items`, `monsters`, `log`, `repl` + `autocomplete`, `lua_event` (event hooks), `lua_global` (VM lifecycle); game hooks `lua::GameStart` (`Source/diablo.cpp:895`); assets: `assets/lua/inspect.lua`, `repl_prelude.lua`, `mods/`; integrity gate: multiplayer refuses overridden `.lua/.tsv/.sol` (`Source/diablo.cpp:900-903` IsAssetIntegrityViolated)
- **Discord RPC**: `Source/discord/` (2 files; `StartGame/UpdateGame` call sites `Source/diablo.cpp:895-941`) — cut (see GI-5)
- **Capture**: `Source/capture.cpp:85` `CaptureScreen` → PCX via `utils/surface_to_pcx` (filename `.pcx` `:49`)
- **QoL**: `Source/qol/` — `stash.cpp` (shared stash, 100 pages), `xpbar.cpp`, `itemlabels.cpp`, `chatlog.cpp`, `floatingnumbers.cpp`, `monhealthbar.cpp`, `autopickup.cpp`, `visual_store.cpp`
- **Demo mode**: `Source/engine/demomode.cpp:39` `Version = 3`; record/playback/`--timedemo` (frame-limit disabled); demo paths in help `Source/diablo.cpp:1075-1079`; reference save comparison `pfile_compare_hero_demo`
- **Spawn gating**: quest wipe (`Source/quests.cpp:96-100`), monster availability filter (`Source/monster.cpp:3159`), set-level gates (`Source/player.cpp:3479`), endgame gates (`:3688-3701`); shareware assets come from `spawn.mpq` at priority 1000
- **Headless**: `Source/headless_mode.cpp:5` + `HeadlessMode` guards at `Source/diablo.cpp:187,727,873,968,3113,3276,3330-3332`; used by timedemo test (`test/timedemo_test.cpp:60`)
- **`restrict.cpp`**: NOT spawn gating — it's a writable-PrefPath probe (`ReadOnlyTest` writes+deletes `Diablo1ReadOnlyTest.foo`, `Source/restrict.cpp:19-28`). Spawn restrictions are `gbIsSpawn` branches (above).

### Algorithms
- Options are declarative `OptionEntry*` trees with categories, flags (CantChangeInGame), callbacks, and ini load/save — a Go `Options` struct + reflection-free explicit registry keeps parity simple.
- Lua: **decision** — embed [`github.com/yuin/gopher-lua`](https://github.com/yuin/gopher-lua) (MIT, Lua 5.1, pure Go) exposing a *reduced* API surface (dev/debug + audio + events) for the fork's own `assets/lua` scripts; do NOT attempt sol2-level parity of all modules. Rationale: scripts shipped in `assets/lua` are tooling (inspect, repl prelude), not gameplay; gameplay-critical data is TSV. Criterion to revisit: if mod ecosystem parity becomes a milestone, gopher-lua stays the engine, we grow bindings incrementally.

### Determinism constraints
- Options must not alter simulation (tickRate is negotiated/multiplayer-fixed; graphics options are render-only). Enforce: sim reads a frozen `SimOptions` snapshot.

### Go/Ebiten mapping decision
- `internal/options`: struct + JSON in config dir (not ini — but ship an ini importer for existing users); `internal/i18n`: `gotext`-style catalog or x/text with the existing `.po` files compiled to `.mo` via `msgfmt` at build time (pure-Go `github.com/leonelquinteros/gotext` reads `.mo` directly); capture → PNG (`image/png`), keep F-key binding; QoL modules ported as standalone packages with feature flags.

### Gaps/risks
- gopher-lua API drift vs sol2 semantics for the few shipped scripts — port `inspect.lua` expectations, not bindings, verbatim.
- gotext plural-forms for all 25 languages — verify with the 3 non-Latin scripts (ru, pl, tr?) early.

### Effort: M

### Open questions
- Keep chatlog/party QoL for M3 only? (Yes — single-player first.)

---

## 12. Testing strategy

### Purpose
What the C++ test suite already proves, what ports directly, and the parity methodology (seeded dungeons, timedemo replay, state hashing, benchmarks).

### Key sources (file:line)
- Suite inventory (`test/`): pure-logic families that port 1:1 to Go unit tests — `random_test.cpp`, `path_test.cpp`, `crawl_test.cpp`, `math_test.cpp`, `rectangle_test.cpp`, `pack_test.cpp` (441-868 item pack cases), `codec_test.cpp`, `writehero_test.cpp`, `quests_test.cpp`, `stores_test.cpp` + `store_transaction_test.cpp` + `vendor_test.cpp`, `inv_test.cpp`, `items_test.cpp`, `missiles_test.cpp`, `player_test.cpp`, `dead_test.cpp`, `townerdat_test.cpp`, `automap_test.cpp`, `drlg_common/l1/l2/l3/l4_test.cpp`, `tile_properties_test.cpp`, `palette_blending_test.cpp`, `appfat_test.cpp`, `data_file_test.cpp`, `format_int/parse_int/str_cat/utf8/sheen_bidi_test.cpp`, `static_vector_test.cpp`, `mod_identity_test.cpp`, `multi_logging_test.cpp`, `vision_test.cpp`, `char_panel/spell_ui/inventory_ui/game_menu/panel_state_test.cpp`, `text_render_integration_test.cpp`, `cursor_test.cpp`, `effects_test.cpp`, `diablo_test.cpp`, `stash_test.cpp`
- Benchmarks: `clx_render_benchmark.cpp`, `dun_render_benchmark.cpp`, `light_render_benchmark.cpp`, `crawl_benchmark.cpp`, `path_benchmark.cpp`, `palette_blending_benchmark.cpp` → Go `testing.B` equivalents
- **Dungeon fixtures**: `test/fixtures/diablo/*.dun` named `level-seed.dun` (e.g. `1-2588.dun`, `10-1630062353.dun`), `test/fixtures/hellfire/*.dun`, `test/fixtures/levels/l{1..4}data/` — asserted by the drlg tests as "same seed → same dungeon" (the acceptance methodology to replicate in Go)
- **Timedemo**: `test/timedemo_test.cpp:28-110` `RunTimedemo("WarriorLevel1to2")` — boots headless, loads MPQs, sets `gbIsSpawn`, replays `fixtures/timedemo/WarriorLevel1to2/demo_0.dmo`, then `pfile_compare_hero_demo(0)` asserts the resulting hero save equals `demo_0_reference_spawn_0.sv` (`:96-100`); skip guard when no MPQs (`:47-49`)
- **txtdata fixtures**: `test/fixtures/txtdata/*.tsv` (crlf, cr, lf, lf_no_trail, empty, empty_with_utf8_bom, sample, utf8_bom) — parser edge cases
- **Memory-map + light fixtures**: `test/fixtures/memory_map/`, `light_render_benchmark/`
- Text rendering: `test/fixtures/text_render_integration_test/`

### Algorithms (Go parity harness)
1. **Seed-dungeon parity**: for each `level-seed.dun` fixture, run the Go DRLG with that seed and compare the generated `dungeon[40][40]` + `dPiece[112][112]` against the fixture bytes. Ports of `drlg_l*_test` — pure unit tests, no C++ needed.
2. **State-hash timedemo**: port demo replay; after every N ticks hash a canonical serialization of (players, monsters, items, quests, DungeonSeeds) — either compare against C++-dumped hashes (generate a fixture corpus from the C++ build with a small patch, kept out of the repo) or compare final hero bytes exactly like the C++ test does (`demo_0_reference_spawn_0.sv` is directly reusable!).
3. **Binary cross-check** (optional, strongest): run headless C++ devilutionX (`--demo N`) and Go side by side, diff state hashes each tick; catches drift the first time it appears.
4. **Golden frames**: fixed interpolation values (0, 64, 128/128) → hash offscreen image; guards the render port (optional, render-only).
5. **Benchmarks**: port the five render/crawl/path benchmarks as `go test -bench`; budget targets from the C++ numbers on identical hardware.

### Determinism constraints
- Tests must never depend on wall clock; the timedemo port runs unbounded (`--timedemo` semantics).

### Go/Ebiten mapping decision
- Standard `testing` + table tests; `testify` only if desired (skill `/home/blaze/.agents/skills/go-fluent-testing/SKILL.md` Given/When/Then parts pattern is available but optional); fixtures copied verbatim from devilutionX `test/fixtures/` into `testdata/`; MPQ-dependent tests skip when no retail MPQs, exactly like C++ (`test/timedemo_test.cpp:47-49`).
- CI gate: `make vet test lint` (per repo AGENTS) + a parity job that regenerates `.dmo`/reference-save comparisons when fixtures update.

### Gaps/risks
- The timedemo reference save couples us to spawn.mpq availability in CI — mirror the C++ skip logic.
- State-hash corpus generation requires a C++ build step in CI; keep it optional/gated.

### Effort: L

### Open questions
- Accept `testify` dependency? (Default: stdlib only for the core, testify allowed in UI tests.)
- Golden-frame hashing across GPUs: Ebiten output should be deterministic per backend, but CI uses software GL — restrict golden frames to CI runner images.

---

## Appendix A — Skills for coders (verified paths)
All verified with `ls` on this machine:
- `/home/blaze/.agents/skills/ebiten/SKILL.md` (router), `game-loop/`, `input/`, `rendering/`, `audio/`, `resources/`, `ecs/`, `physics/`, `deployment/`
- `/home/blaze/.agents/skills/go-dev/SKILL.md`, `/home/blaze/.agents/skills/go-setup/SKILL.md`, `/home/blaze/.agents/skills/go-fluent-testing/SKILL.md`, `/home/blaze/.agents/skills/coding-rules/SKILL.md`, `/home/blaze/.agents/skills/beans/SKILL.md`

## Appendix B — Online references
- Ebitengine performance tips: https://ebitengine.org/en/documents/performancetips.html (Update-only logic, no hot-path allocations, batching)
- go-mpq: https://pkg.go.dev/github.com/suprsokr/go-mpq (v0.3.1, MIT; PKWARE/zlib/bzip2 read; Huffman not implemented — audio-only usage)
- OpenDiablo2/mpq: GPL-3.0 (rejected on license)
- CL2 file format reference: https://github.com/savagesteel/d1-file-formats/blob/master/PC-Mac/CL2.md
- mpqfs (devilutionX MPQ backend): https://github.com/diasurgical/mpqfs @ de99526 (via `3rdParty/mpqfs/CMakeLists.txt:9-11`)

## Appendix C — Effort roll-up
| # | Subsystem | Effort |
|---|-----------|--------|
| 1 | Core loop/threading | M |
| 2 | World/DRLG | XL |
| 3 | Entities | XL |
| 4 | Gameplay | XL |
| 5 | Persistence | L |
| 6 | Asset pipeline | XL |
| 7 | Rendering | XL |
| 8 | Networking | L→XL |
| 9 | Input | M |
| 10 | UI | L |
| 11 | Platform/misc | M |
| 12 | Testing/parity | L |

Total: a multi-month full rewrite; the milestone plan (see `/tmp/dg-bean-tree.md`) sequences a playable spawn demo early (M1) to de-risk engine core + assets before the XL gameplay/renderer push.
# devilutionX → Go/Ebiten/SQLite Conversion — Requirements Study (APPENDICES D–F)

> Continuation of `/tmp/dg-requirements.md` (§§GI, 1–12). This volume adds:
> Appendix D — enum & constant registry (ports checklist), Appendix E — Go package layout,
> Appendix F — per-milestone functional acceptance matrices, Appendix G — risk register,
> Appendix H — open-questions ledger with decision criteria.

---

## Appendix D — Enum & constant registry (port checklist)

Every constant here is cited from source and must exist in Go with identical numeric values
(they are serialized into saves/packets or index shared tables).

### D.1 World & levels
| Constant | Value | Source |
|---|---|---|
| `DMAXX`, `DMAXY` | 40, 40 | `Source/levels/gendung_defs.hpp:5-6` |
| `MAXDUNX`, `MAXDUNY` | 16+40*2+16 = 112 | `Source/levels/gendung_defs.hpp:8-9` |
| `NUMLEVELS` | 25 | `Source/levels/gendung_defs.hpp:11` |
| `MAXTILES` | 1379 | `Source/levels/dun_tile_data.hpp:24` |
| `MAXTHEMES` | 50 | `Source/levels/dun_tile_data.hpp:21` |
| `TILE_WIDTH/HEIGHT` | 64/32 | `Source/levels/dun_tile.hpp:9-10` |
| `DunFrameWidth/Height` | 32/32; triangle h=31 | `Source/levels/dun_tile.hpp:127-131` |
| `MAXLIGHTS` | 32 | `Source/engine/lighting_defs.hpp:7` |
| `NumLightRadiuses` | 16 | `Source/engine/light_tables.hpp:16` |
| `DungeonFlag` bits | Missile 1<<0 … Explored 1<<7; Saved=Populated\|Lit\|Explored; Loaded=Missile\|Visible\|DeadPlayer\|… | `Source/levels/dun_tile_data.hpp:28-45` |
| `TileProperties` bits | Solid, BlockLight, BlockMissile, Transparent, TransparentLeft/Right, Trap(1<<7) | `Source/levels/dun_tile.hpp:105-116` |
| theme_id | THEME_BARREL(0)…THEME_WEAPONRACK(15), THEME_NONE=-1 | `Source/tables/objdat.h:18-37` |

### D.2 Entities
| Constant | Value | Source |
|---|---|---|
| `MAX_PLRS` | 4 | `Source/multi.h:26` |
| `MaxMonsters` | 200 | `Source/monster.h:38` |
| `MaxLvlMTypes` | level monster types cap | `Source/monster.h` (LevelMonsterTypes decl) |
| `MAXOBJECTS` | 127 | `Source/objects.h:30` |
| `MAXITEMS` | 127 | `Source/items.h:25` |
| `InventoryGridCells` | 40 | `Source/player.h:38` |
| `MaxBeltItems` | 8 | `Source/player.h:39` |
| `NumHotkeys` | 12 | `Source/player.h:44` |
| `PlayerNameLength` | 32 | `Source/player.h:42` |
| `MaxPathLengthPlayer` | 100 | `Source/engine/path.h:20` |
| `MAX_CHARACTERS` | 99 | `Source/pfile.h:24` |
| `MAXQUESTS` | 24 | `Source/tables/questdat.hpp:14` |
| `MAXPORTAL` | portals per player | `Source/msg.h` (DPortal usage) |
| `Direction` | 8-way + NoDirection; left/right rotation tables | `Source/engine/direction.hpp:11,26,33` |
| `PLR_MODE` | PM_STAND, PM_WALK_NORTHWARDS, PM_WALK_SOUTHWARDS, PM_WALK_SIDEWAYS, … | `Source/player.h:110-113` |
| `MonsterMode` | Stand, MoveNorthwards, MoveSouthwards, MoveSideways, MeleeAttack, HitRecovery, Death, SpecialMeleeAttack, FadeIn, FadeOut, RangedAttack, … | `Source/monster.h:76-91` |
| `MonsterAIID` | Zombie(0), Fat, SkeletonMelee, SkeletonRanged, Scavenger, Rhino, GoatMelee, GoatRanged, Fallen, Magma, SkeletonKing, Bat, Gargoyle, Butcher, Succubus, Sneak, Storm, FireMan, Gharbad, Acid, AcidUnique, Golem, Zhar, Snotspill, Snake, Counselor, Mega, Diablo, Lazarus, LazarusSuccubus, Lachdanan, Warlord, FireBat, Torchant, HorkDemon, Lich… | `Source/tables/monstdat.h:22-58` |

### D.3 Networking & commands
| Constant | Value | Source |
|---|---|---|
| `packet_type` | PT_MESSAGE 0x01, PT_TURN 0x02, PT_JOIN_REQUEST 0x11, PT_JOIN_ACCEPT 0x12, PT_CONNECT 0x13, PT_DISCONNECT 0x14, PT_INFO_REQUEST 0x21, PT_INFO_REPLY 0x22, PT_ECHO_REQUEST 0x31, PT_ECHO_REPLY 0x32 | `Source/dvlnet/packet.h:27-40` |
| `plr_t/seq_t` | uint8/uint8 | `Source/dvlnet/packet.h:45-46` |
| `cookie_t/timestamp_t/key_t` | uint32/uint32/array<uchar, crypto_secretbox_KEYBYTES> | `Source/dvlnet/packet.h:47-50` |
| `PLR_MASTER/PLR_BROADCAST` | 0xFE/0xFF | `Source/dvlnet/packet.h:61-62` |
| `turn_t` | {seq u8…, Value int32} | `Source/dvlnet/packet.h:56-59` |
| `_cmd_id` | 86 values CMD_* (see §8) + FAKE_CMD_SETID/DROPID + CMD_INVALID 0xFF | `Source/msg.h:18-428` |
| `MAX_SEND_STR_LEN` | 80 | `Source/msg.h:22` |
| `HeaderCheckVal` | TPkt.wCheck sentinel | `Source/msg.h` (TPktHdr use) |
| `TPkt.body` | 493 bytes | `Source/msg.h:711-714` |
| `TMegaPkt.data` | 32000 bytes | `Source/msg.cpp:207-212` |

### D.4 Persistence
| Constant | Value | Source |
|---|---|---|
| Passwords | single `xrgyrkj1`, multi `szqnlsk1`, spawn-single `adslhfb1`, spawn-multi `lshbkfg1` | `Source/pfile.cpp:60-63` |
| Save magics | RETL / SHAR / HELF / SHLF | `Source/loadsave.cpp:2765-2772` |
| Counts | levels 25(HF)/17(D), quests 24/16, premium 15/6 | `Source/loadsave.cpp:2773-2781` |
| `pSplLvl` | [37] (not 64) for save compat | `Source/pack.h:65` |
| Codec block | 32 bytes + signature; `codec_get_encoded_len` pads + sig | `Source/codec.cpp:137-160` |
| ItemPack size | u32+u16+u16+5×u8+u16+u32 = 24B | `Source/pack.h:20-30` |

### D.5 Input & UI
| Constant | Value | Source |
|---|---|---|
| `StickPush` | None/Weak/Firm thresholds (fork HEAD) | `Source/controls/controller_motion.h` (+15 lines in ca91d778f) |
| `TalkID` | 23 states (None…Barmaid) | `Source/stores.h:44-70` |
| Automap modes | `AutomapType` (Normal/Transparent/Minimap) + `CurrentAutomapType` | `Source/automap.h:83,91` |
| Tick slider | 20..50 | `Source/gamemenu.cpp:275-277` |
| Audio rates | 22050 default; {22050,44100,48000} | `Source/options.cpp:55,520` |

### D.6 Demo format
| Field | Value | Source |
|---|---|---|
| Version | 3 | `Source/engine/demomode.cpp:39` |
| Record | `type u8`, `mod u16`, payload; tick records carry `ProgressToNextGameTick \| 0x80` | `Source/engine/demodode.cpp:595-599` (demomode.cpp) |
| Event types | GameTick=0, MouseMotion{x,y u16}, MouseButton{x,y,mod,button}, MouseWheel{x i16,y i16,mod u16}, Key{sym u32, mod u16} | `Source/engine/demomode.cpp:43-58` |

---

## Appendix E — Proposed Go package layout

```
devilutionGo/
├── cmd/
│   ├── dgo/main.go                 # ebiten entry, flag parsing (--demo, --timedemo, --spawn…)
│   └── dgoheadless/main.go         # headless runner (tests, CI parity)
├── internal/
│   ├── engine/                     # §1
│   │   ├── tick.go                 # accumulator, gnTickDelay, ProgressToNextGameTick
│   │   ├── gamelogic.go            # GameLogic() fixed order port
│   │   ├── random.go               # DiabloGenerator LCG + xoshiro seedGen
│   │   ├── path.go                 # A* (MaxPathNodes 1024, costs 100/101)
│   │   ├── animation.go            # AnimationInfo (128ths fixed point)
│   │   ├── direction.go            # 8-way + rotation tables
│   │   └── lighting.go             # DoLighting, LightFalloffs, vision
│   ├── world/                      # §2
│   │   ├── grid.go                 # dungeon/dPiece/dFlags… arrays
│   │   ├── drlg_l1.go … drlg_l4.go # CreateL{1..4}Dungeon ports
│   │   ├── town.go, crypt.go, setmaps.go
│   │   ├── themes.go, trigs.go, tileprops.go
│   │   └── miniset.go              # Miniset match/place engine
│   ├── entity/                     # §3
│   │   ├── player.go, monster.go, missile.go, object.go, item.go, towner.go, corpse.go
│   ├── gameplay/                   # §4
│   │   ├── combat.go               # to-hit/damage pipelines (<<6 fixed point)
│   │   ├── spells.go, missiles_table.go
│   │   ├── itemgen.go              # affix/unique rolls
│   │   ├── inventory.go, belt.go
│   │   ├── stores/                 # TalkID state machine + transaction logic
│   │   ├── quests.go, xp.go
│   ├── net/                        # §8
│   │   ├── packet.go               # PT_*, turn_t, secretbox seal/open
│   │   ├── tcp/server.go, tcp/client.go
│   │   ├── loopback.go, base.go    # SNet surface, PlayerState turn queues
│   │   ├── cmd.go                  # 86 CMD_* + payload structs (manual marshal)
│   │   ├── delta.go                # DLevel/DMonsterStr/DJunk sync
│   │   └── turn.go                 # nthread cadence port
│   ├── assets/                     # §6
│   │   ├── mpq.go                  # go-mpq wrapper + priority stack
│   │   ├── clx.go, cl2.go, cel.go, pcx.go, trn.go, palette.go, font.go
│   │   ├── txtdata/                # TSV parser + generated tables
│   │   └── embedded/               # go:embed of unpacked devilutionx assets
│   ├── render/                     # §7
│   │   ├── world.go                # DrawFloor/DrawTileContent/DrawDungeon order
│   │   ├── sprite.go               # atlas, SubImage, light ColorScale, TRN re-palettize
│   │   ├── automap.go, panels/, text.go, cursor.go, movie_stub.go
│   ├── ui/                         # §10
│   │   ├── screen.go (stack), widget.go, title.go, mainmenu.go, selhero.go, selgame.go
│   │   └── gmenu.go, dialog.go, progress.go
│   ├── persist/                    # §5
│   │   ├── pack.go, loadsave.go    # byte-exact serializers
│   │   ├── codec.go, diablosha.go  # save encryption
│   │   ├── sqlite.go               # modernc.org/sqlite store
│   │   ├── legacysv.go             # .sv import/export
│   │   └── pfile.go                # hero lifecycle, MAX_CHARACTERS 99
│   ├── input/                      # §9
│   │   ├── devices.go, gamepad.go (ScaleJoystickAxes, StickPush, TurnToDir)
│   │   ├── plrctrls.go, keymapper.go, padmapper.go, touch/ (later)
│   ├── platform/                   # §11
│   │   ├── options.go, ini_import.go, i18n/, lua/ (gopher-lua), qol/
│   ├── demo/                       # demo/timedemo record+replay (cmd stream)
│   └── game/                       # wires everything; ebiten.Game impl (Update/Draw/Layout)
└── testdata/                       # fixtures copied from devilutionX test/fixtures
```

Rules:
- `internal/game` is the ONLY package importing ebiten besides `cmd` — everything below it is engine-pure and unit-testable headless.
- No package may import upward. `gameplay` sees `entity`/`world`/`engine`; `render` reads state but never mutates.
- The `SimOptions` snapshot (frozen at game start) is the only options view gameplay may read.

---

## Appendix F — Milestone acceptance matrices

### F.1 M0 — Foundations (repo, CI, skeleton)
- [ ] `go build ./...` green on linux/amd64 + windows/amd64 (no CGO).
- [ ] `go vet`, `golangci-lint` (repo `.golangci.yml`), `go test ./...` wired into `justfile` targets mirroring `make vet test lint`.
- [ ] Ebiten window opens 640×480, fixed Layout, `Update` accumulator running at 20 TPS with a debug tick counter; `Draw` shows interpolated position of a moving test rect (progress fraction 0..128).
- [ ] `internal/engine/random.go` passes ported vectors from `test/random_test.cpp`; `internal/engine/path.go` passes `test/path_test.cpp` cases (seeded maps).
- [ ] Headless build tag compiles the same logic with a null renderer.

### F.2 M1 — Assets + Town + Playable Spawn slice
- [ ] MPQ stack loads spawn.mpq (+devilutionx assets unpacked-embed) with correct priority; `OpenAsset("levels\\towndata\\town.dun")` returns bytes (verify against hexdump of `test/fixtures/diablo/1-2588.dun`-style header: two u16 LE dims).
- [ ] go-mpq extraction smoke test enumerates 100% of files in spawn.mpq and decodes each (PKWARE/zlib only expected); CI artifact lists any failure.
- [ ] Town loads: tiles render from `towndata` CLX via decoded MICROS; palette + TRN correct (screenshot vs C++ reference at same view).
- [ ] Player (Warrior) walks in town at 20 TPS with 8-way facing, interpolated rendering, belt/inventory UI panels drawn from `ui_art`.
- [ ] One towner dialog (Griswold) + one store screen functional via TalkID state machine subset.
- [ ] Seed-dungeon parity: L1 DRLG produces fixture-identical `dungeon` grids for all `test/fixtures/diablo/*.dun` seeds.

### F.3 M2 — Full Diablo single-player
- [ ] All 16 quests + all 4 DRLG types + setmaps (sklkng, bonecha, foulwatr, vile) + themes + triggers.
- [ ] Combat parity: scripted scenario tests reproduce C++ damage/hit outcomes for N=1000 seeded rolls (port `PlrHitMonst` roll-sequence expectations).
- [ ] Save/load: `.sv` byte-round-trip against fixtures; SQLite store save/load→same world state hash.
- [ ] Demo record/replay of a full level-1→2 run matches reference hero bytes (reuse `demo_0_reference_spawn_0.sv`).
- [ ] All 20 TPS systems (missiles incl. Guardian chains, monster AIs of Cathedral+ Catacombs bestiary, lighting/vision) tick-parity via state-hash dump diff vs C++ headless run on 3 seeds.

### F.4 M3 — Multiplayer (Go↔Go)
- [ ] TCP host/join with secretbox-encrypted PT_* handshake; 2–4 players; lockstep turns at negotiated rate.
- [ ] Delta sync converges after forced divergence (kill monster on A, verify B hp/position within 1 tick).
- [ ] Chat (CMD_STRING), trade-safe item ownership (TCmdGItem bMaster/bPnum protocol), drop/rejoin.

### F.5 M4 — Hellfire
- [ ] hellfire.mpq + hf*.mpq load; Monk class; crypt/nest levels 17–24; 8 extra quests; Oliver/Scholar NPC set; Runes/Oils items; `gbIsHellfire` branch audit complete (135 root-cpp call sites mapped to Go equivalents).

---

## Appendix G — Risk register (top 10)

1. **DRLG parity drift** (L1–L4 quirks: `drlg1Quirk`, crypt `respectProtected=false` bug-compat) — mitigated by fixture-driven TDD; every generator PR must green all fixture seeds.
2. **RNG call-order divergence** in gameplay (extra/missing `GenerateRnd` shifts all downstream rolls) — mitigate with C++-side state-dump harness diffing `sglGameSeed` per tick.
3. **CL2/CEL decoder edge cases** (RLE 128 semantics, triangle padding removal `ReencodeDungeonCels`) — freeze with byte-level golden frames from C++ `DEBUG_CEL_TO_CL2_SIZE` builds.
4. **go-mpq gaps** (Huffman absent; write-side zlib-only) — smoke-test every retail file; if any Huffman sector appears, port `pk_explode`-style decoder for it (unlikely per mpqfs parity).
5. **Save byte-compat** (mixed endianness per field, padding, missile cap 125) — port `SaveHelper` exactly; round-trip against `.sv` corpus in CI.
6. **Ebiten performance at 112×112 tile scenes** (batching discipline) — benchmark early with `dun_render_benchmark` port; enforce no-alloc Draw via `go test -benchmem` + lint rule.
7. **Network lockstep divergence** (turn application order) — deterministic per-tick command ordering tests + fuzz marshal against C++ hexdumps.
8. **Store/dialog UX port size** (2911 lines mixing UI+logic) — extract pure transaction core first (mirroring `store_transaction_test`), UI after.
9. **Localization breadth** (25 languages, bidi via SheenBidi in C++) — start with 2 (en, pl), design string registry so `.po` ingestion is mechanical; defer bidi.
10. **Scope creep into QoL/Lua before core parity** — milestone-gated: no qol/, lua/ work before M2 complete (contract).

## Appendix H — Open-questions ledger (with decision criteria)

| # | Question | Options | Decide by |
|---|----------|---------|-----------|
| H1 | Timedemo frame-uncapped mode needed? | state-hash only vs golden frames | M2 exit review: if C++ cross-check harness proves sufficient, drop golden frames |
| H2 | Two-resolution world model (40×40 + 112×112) | keep vs micro-native | keep (save format + fixtures depend); revisit only if M2 perf suffers |
| H3 | Item power flags representation | hand-enums vs generated from TSV | codegen spike in M1; choose whichever passes affix fixture tests with less drift |
| H4 | Quest port order | 16-then-8 vs all-24-arrays-day-one | decided: arrays for 24 from day one, logic Diablo-first |
| H5 | MPQ writing beyond .sv export | none vs mod support | defer until mods milestone exists |
| H6 | Embedded assets form | unpacked tree (go:embed) vs devilutionx.mpq blob | unpacked (matches UNPACKED_MPQS support); blob only if embed size hurts builds |
| H7 | Logical resolution | fixed 640×480 vs dynamic | fixed + integer scale default; dynamic option post-M2 |
| H8 | UDP LAN transport | TCP-only vs base_protocol port | post-M3 latency data |
| H9 | Cross-play with C++ peers | Go-only vs wire-compat | wire format is identical anyway; decide on marketing/CI cost at M3 |
| H10 | ebitenui for settings screen | hand-rolled vs lib | settings-only lib trial in M1 |
| H11 | Movies | defer vs SMK decoder vs pre-encoded | M2: still+text; M4 review |
| H12 | Touch controls | defer vs port controls/touch | mobile milestone (post-M4) |
| H13 | testify in tests | stdlib vs testify | stdlib core / testify UI (decided, revisit never) |
| H14 | Golden-frame hashing scope | CI-runner-only vs all GPUs | CI-runner-only (decided) |
| H15 | Lua engine | gopher-lua subset vs none | gopher-lua subset for dev tooling (decided); full sol2 parity is a non-goal |

---

## Appendix I — Source-tree quick map (for coder orientation)

- Root `Source/*.cpp` (55 files): diablo (loop), gamemenu, gmenu, init, interfac, inv, items, levels glue via levels/, lighting, loadsave, msg, multi, nthread, objects* (root objects.cpp), pack, pfile, player, quests (root quests.cpp + tables/questdat), stores, towners, dead, doom, automap, capture, movie, codec, encrypt, sha, crawl, cursor, debug, effects, appfat, restrict.
- `Source/engine/` (70f): assets, random, path, animationinfo, direction, lighting_defs, light_tables, trn, palette, dx (surface), backbuffer_state, ticks, events, demomode, sound*, load_cel/cl2/clx/pcx, clx_sprite, render/ (scrollrt, dun_render, clx_render, light_render, text_render, automap_render, primitive_render).
- `Source/levels/` (32f): gendung*, drlg_l1..l4, town, crypt, setmaps, themes, trigs, dun_tile*, tile_properties, reencode_dun_cels, drlg_quests, parse_dungeon_type.
- `Source/tables/` (18f): itemdat, misdat, monstdat, objdat, playerdat, questdat, spelldat, textdat, townerdat + loaders (LoadItemData `Source/tables/itemdat.cpp:710`, LoadMissileData `misdat.cpp:442`, LoadMonsterData `monstdat.cpp:472`, LoadObjectData `objdat.cpp:257`).
- `Source/controls/` (36f incl. touch/, devices/), `Source/DiabloUI/` (33f incl. hero/, multi/), `Source/dvlnet/` (24f), `Source/lua/` (59f incl. modules/, modules/dev/), `Source/qol/` (16f), `Source/panels/` (19f), `Source/control/` (11f: panel, flasks, chat, gold, infobox), `Source/platform/` (50f), `Source/storm/` (4f: storm_net, storm_svid), `Source/mpq/` (8f), `Source/items/` (2f: validation), `Source/discord/` (2f), `Source/data/` (TSV parser: file/iterators/record_reader/value_reader).
- `test/` (70 files): see §12; fixtures: diablo/ hellfire/ levels/l{1..4}data txtdata/ timedemo/WarriorLevel1to2 memory_map light_render_benchmark text_render_integration_test.

---

## Appendix J — Worked examples (concrete number walkthroughs)

### J.1 Melee combat: Warrior lvl 5 vs Zombie (SkeletonMelee-class data), seed-for-seed

Setup (values from tables cited in §3/§4):
- Warrior 5: STR 30, DEX 20, level 5 → `GetMeleeToHit = 5 + 20/2 + 0 + baseMeleeToHit(warrior≈50 per playerdat.tsv)` (`Source/player.h:575`).
- Zombie armorClass from `txtdata\monsters\monstdat.tsv` row 1 (`MT_NZOMBIE … toHit 4 …`).
- Weapon: bare-hand range mind..maxd from class data; `_pDamageMod` derived from STR.

Roll sequence (must match C++ call order exactly — every `GenerateRnd` advances the global LCG):
```
1. hit = GenerateRnd(100)                              // player.cpp:542
2. hper = GetMeleePiercingToHit() - CalculateArmorPierce(ac, true)
   hper = clamp(hper, 5, 95)
3. if hit >= hper → miss (return)
4. dam = RandomIntBetween(_pIMinDam, _pIMaxDam)        // random.hpp:403
5. dam += dam * _pIBonusDam / 100
6. dam += _pIBonusDamMod
7. dam2 = dam << 6                                      // snapshot
8. dam += _pDamageMod                                   // STR-derived
9. warrior crit: if GenerateRnd(100) < 5 → dam *= 2     // player.cpp:575
10. monster.data().monsterClass == Undead:
    mace → dam += dam/2 ; sword → dam -= dam/2          // player.cpp:602-610
11. (hellfire branches: Devastation/Doppelganger/Jesters — gbIsHellfire==false here)
12. dam <<= 6                                           // 1/64 units
13. adjacentDamage ? dam >>= 2
14. monster.hp -= dam; hit-recovery mode if threshold
```
Go test: seed the LCG to a known state, run `PlrHitMonst` port 1000 times, compare the full roll-tape (each GenerateRnd result) against a C++ dump. This is the template for all combat parity tests.

### J.2 Level generation: `CreateL5Dungeon(seed, entry)` step tape

From `Source/levels/drlg_l1.cpp:1298-1312` + internals read in §2:
```
SetRndSeed(seed)                    // engine/random.cpp:64
UberRow = UberCol = 0
GenerateLevel(entry):
    InitDungeonFlags()              // clear dungeon/pdungeon/Protected/dungeon flags
    FirstRoom()…                    // quest chamber layout, VerticalLayout roll
    …RoomGen / corridor carving …   // (all GenerateRnd-ordered)
    PlaceMiniSet(stairs-down, …)    // drlg1Quirk variant
    PlaceMiniSetRandom(LAMPS, 1600) // ×(5 + GenerateRnd(5)) — drlg_l1.cpp:1231-1234
    FixTransparency(); FixDirtTiles(); FixCornerTiles()
    Substitution(); ApplyShadowsPatterns()
    FillFloor()
    memcpy(pdungeon, dungeon)
    DRLG_CheckQuests(SetPiece.position)
Pass3():
    DRLG_LPass3(Dirt - 1)           // expand 40×40 → 112×112 micros via pMegaTiles
    InitDungeonPieces()
```
Fixture contract: for every `test/fixtures/diablo/L-SEED.dun`, Go's `CreateL5Dungeon(SEED, ENTRY_MAIN)` must yield identical `dungeon[40][40]` (fixture bytes 0x08.. after the 2×u16 header) — the file layout was verified by hexdump (`1-2588.dun` starts `28 00 28 00` = 40,40).

### J.3 Network turn lifecycle (2 players, TCP)

```
T+0ms  A(host): game tick due → collects cmds since last turn
         packet_out PT_TURN{s=A, d=BROADCAST, turn={seq=N, value=cmdsBlob}}
         secretbox_seal(key from password, nonce || mac || pt_plaintext)
         frame_queue: [u32 len][ciphertext] → TCP
T+0ms  B: io goroutine reads frame → decrypt → base::HandleTurn
         turnQueue[A].push({N, cmds}) ; if turn N complete for all → runnable
T+50ms both: multi_handle_delta() applies A's and B's TPkt bodies
         in player-id order (ProcessGameMessagePackets, multi.cpp:729)
         → ParseCmd dispatch (msg.cpp:3341) → On_CMD_WALKXY etc.
         current_turn = N+1
```
Join: `PT_JOIN_REQUEST{cookie}` → host assigns plr slot → `PT_JOIN_ACCEPT{cookie, game_init_info}` → `CMD_ACK_PLRINFO` exchange → delta snapshot via `DeltaExportData(pnum)` (msg.cpp:698 caller).
Drop: `SNetDropPlayer` → `PT_DISCONNECT` broadcast + `msg_send_drop_pkt` (msg.h:718).

### J.4 Demo file record layout (Version 3)

From `Source/engine/demomode.cpp:39-58, 595-599, 816`:
```
header: version byte (3), then demo settings + player count …
records (LE):
  [u8 type][payload...]
  type & 0x80 → GameTick record: (type & 0x7F) == ProgressToNextGameTick
  Key:      [u32 sym][u16 mod]
  MouseBtn: [u16 x][u16 y][u16 mod][u8 button]
  MouseMove:[u16 x][u16 y]
  Wheel:    [i16 x][i16 y][u16 mod]
```
Go `internal/demo` reads/writes this exactly; the paired reference save (`demo_0_reference_spawn_0.sv`, fixture dir) is the oracle for `pfile_compare_hero_demo` port.

---

## Appendix K — C++ test → Go test port matrix

| C++ test (`test/`) | Go target | Kind | Notes |
|---|---|---|---|
| random_test.cpp | internal/engine/random_test.go | unit | LCG vectors, xoshiro jump |
| path_test.cpp + path_benchmark.cpp | internal/engine/path_test.go / _bench.go | unit+bench | A* 100/101 costs |
| crawl_test.cpp (+bench) | internal/engine/crawl_test.go | unit+bench | spiral search |
| math_test.cpp, rectangle_test.cpp | internal/engine/*_test.go | unit | pure |
| drlg_common/l1/l2/l3/l4_test.cpp | internal/world/drlg_*_test.go | unit+fixtures | seed→dungeon parity |
| tile_properties_test.cpp | internal/world/tileprops_test.go | unit | SOL flags |
| pack_test.cpp | internal/persist/pack_test.go | unit | ItemPack cases incl. unique-bug |
| codec_test.cpp | internal/persist/codec_test.go | unit | Diablo-SHA1 vectors |
| writehero_test.cpp | internal/persist/pfile_test.go | unit | hero round-trip |
| stash_test.cpp | internal/persist/stash_test.go | unit | 100-page grid |
| quests_test.cpp | internal/gameplay/quests_test.go | unit | pools, spawn wipe |
| stores/store_transaction/vendor/visual_store_test.cpp | internal/gameplay/stores/*_test.go | unit+UI | transaction core first |
| inv_test.cpp, inventory_ui_test.cpp | internal/gameplay/inventory_test.go, internal/ui/inv_test.go | unit/UI | 40-grid packing |
| items_test.cpp | internal/gameplay/itemgen_test.go | unit | affix/unique rolls |
| missiles_test.cpp | internal/entity/missile_test.go | unit | behavior table |
| player_test.cpp | internal/entity/player_test.go | unit | modes/stats |
| dead_test.cpp | internal/entity/dead_test.go | unit | corpse encoding |
| townerdat_test.cpp | internal/entity/towner_test.go | unit | TSV-driven |
| automap_test.cpp | internal/render/automap_test.go | unit | tile types |
| panel_state/char_panel/spell_ui/game_menu_test.cpp | internal/ui/*_test.go | UI | state machines |
| text_render_integration_test.cpp | internal/render/text_test.go | UI-int | fonts fixture |
| clx_render/dun_render/light_render/palette_blending benchmarks | internal/render/*_bench.go | bench | perf budgets |
| data_file_test.cpp (+ txtdata fixtures) | internal/assets/txtdata/*_test.go | unit | BOM/CRLF/CR edge cases |
| ini_test.cpp | internal/platform/ini_test.go | unit | options import |
| utf8_test.cpp, str_cat_test.cpp, format_int_test.cpp, parse_int_test.cpp | internal/str/*_test.go | unit | string utils |
| static_vector_test.cpp | internal/engine/staticvec_test.go | unit | pool container |
| multi_logging_test.cpp | internal/net/logging_test.go | unit | msg logging |
| mod_identity_test.cpp | internal/assets/mod_test.go | unit | mod manifests |
| vision_test.cpp | internal/engine/vision_test.go | unit | vision list |
| cursor_test.cpp, effects_test.cpp | internal/input, internal/platform | unit | |
| diablo_test.cpp | internal/game/game_test.go | int | boot smoke (headless) |
| timedemo_test.cpp | internal/demo/timedemo_test.go | parity | reference `.sv` oracle |
| appfat_test.cpp | n/a (error paths) | — | Go panics/fataler |
| animationinfo_test.cpp | internal/engine/animation_test.go | unit | 128ths math |
| codec/encrypt/sha | internal/persist/* | unit | covered above |
| assets_test.cpp | internal/assets/assets_test.go | unit | archive stack |
| sheen_bidi_test.cpp | deferred (i18n milestone) | — | post-M2 |

Porting rule: a C++ test ports **before** its feature ("test-first from fixtures"); every fixture file referenced gets copied verbatim into `testdata/`.

---

## Appendix L — Engine-internal API surface to mirror (Go signatures)

Signatures the port should keep 1:1 (renamed to Go style) so the C++ source remains readable as documentation.

```go
// engine/tick.go
func (g *Game) Has500msPassed() (runLogic bool, drawGame bool)        // nthread_has_500ms_passed
func (g *Game) UpdateProgressToNextGameTick()                          // 0..128
func GameLoop(startup bool) bool                                       // diablo.cpp:3490
func GameLogic()                                                       // diablo.cpp:1523 (fixed order!)

// engine/random.go
func SetRndSeed(seed uint32)
func GetLCGEngineState() uint32
func GenerateRandomNumber() uint32
func AdvanceRndSeed() int32      // abs(int32(lcg())) with INT_MIN case
func GenerateRnd(v int32) int32  // high-bits trick for v<=0x7FFF (random.hpp:76-84)
func FlipCoin(frequency uint) bool
func ReserveSeedSequence() *Xoshiro // seedGenerator.jump()

// engine/path.go
func FindPath(canStep func(from, to Point) bool, posOk func(Point) bool,
    start, dest Point, path []int8) int                                 // path.cpp:184
func FindClosestValidPosition(posOk func(Point) bool, start Point,
    minRadius, maxRadius uint) (Point, bool)                            // path.cpp:285

// world/gendung.go
func CreateDungeon(rseed uint32, entry LevelEntry)
func PlaceMiniSet(ms Miniset, tries int, drlg1Quirk bool) (Point, bool)
func (w *World) SetDungeonMicros()                                      // til/min expansion
func LoadLevelSOLData() error                                           // SOLData[1379]

// entity
func AddMissile(src, dst Point, dir Direction, id MissileID,
    caster MienemyType, source int, dam, splLvl int, parent *Missile) *Missile
func CheckMissileCol(m *Missile, dtype DamageType, minD, maxD int,
    shifted bool, pos Point, dontDelete bool, onlyWalking bool)         // missiles.cpp:479

// gameplay
func PlrHitMonst(p *Player, m *Monster, adjacentDamage bool) bool      // player.cpp:524
func CalcPlrInv(p *Player, loadgfx bool)                                // player.cpp:450 funnel
func SpawnItem(monster *Monster, pos Point, sendmsg, spawn bool)        // items.cpp:3422
func SpawnUnique(uid UniqueItemID, pos Point, lvl int, ...) *Item       // items.cpp:3195

// net
func NetSendCmd(bHiPri bool, cmd CmdID)                                 // msg.cpp:2982
func NetSendCmdLoc(playerID uint8, bHiPri bool, cmd CmdID, pos Point)   // :3009
func NetSendCmdLocParam1..4(...)                                        // :3027-3084
func ParseCmd(pnum uint8, data []byte) (n int)                          // msg.cpp:3341
func DeltaExportData(pnum uint8)                                        // msg.cpp:698

// persist
func PackPlayer(dst *PlayerPack, p *Player, manashield bool)
func UnPackPlayer(src *PlayerPack, p *Player, netSync bool) bool
func SaveGameData(w SaveWriter)                                         // loadsave.cpp:2763
func LoadGameData(r SaveReader) error
func CodecDecode(buf []byte, password string) int                       // codec.cpp:92
func CodecEncode(buf []byte, size64 int, password string)
```

Naming conventions: keep the C++ verb names (Go-cased) for everything that is a ported algorithm; new Go-only glue may use idiomatic names. This keeps `git grep`-style cross-referencing trivial during the port years.

## Appendix M — Threading & goroutine model (normative)

```
main goroutine (ebiten.RunGame)
 ├── Update(): input polling → command build → tick accumulator → GameLogic()×k (k≤10)
 ├── Draw(): pure render from (state, progress)
 └── Layout(): fixed 640×480

net goroutine (per connection + server accept loop)
 ├── owns: TCP conn read/write, frame queue, secretbox seal/open
 ├── sends decrypted packets to game via channel chan PacketOut
 └── receives packets-to-send via channel chan PacketOut
     (never touches world state; mirrors MemCrit discipline in nthread.cpp:48-77)

audio goroutine (ebiten audio.Context owns its thread)
assets loader (init-time only; no runtime loading in Draw — perf rule)

Synchronization rule: exactly ONE goroutine mutates GameState (the Update
caller). Network/deltas are buffered and applied at tick boundary inside
multi_handle_delta port — identical to C++ where NthreadHandler only
paces sends and the main loop applies messages (diablo.cpp:942).
```

## Appendix N — License & provenance notes

- devilutionX code: maintained as clean-room reimplementation of Diablo; its own sources carry no Blizzard code. Our Go port is a *derivative work of devilutionX* (its repo license — check `LICENSE.md` at repo root before publishing; historically GPL-2.0-ish with special exceptions; **verify at M0 and record in NOTICE**).
- `github.com/suprsokr/go-mpq`: MIT (verified https://pkg.go.dev/github.com/suprsokr/go-mpq).
- `github.com/OpenDiablo2/mpq`: GPL-3.0 — **rejected**.
- Ebitengine: Apache-2.0. modernc.org/sqlite: MIT (transpiled from SQLite public domain). gopher-lua: MIT.
- Retail MPQs remain user-supplied (same stance as devilutionX): the Go build ships only devilutionX-derived `assets/` content, never Blizzard data.

---

## Appendix O — Glossary & acronym map

| Term | Meaning | Where |
|---|---|---|
| TPS | ticks per second (logic rate) | §1 |
| DRLG | dungeon (random level) generation | §2 |
| DUN | level tile data file (`.dun`) | §2 |
| MIN/TIL | micro-tile & mega-tile definitions per dungeon type | §2 |
| SOL | "solid" tile property table (SOLData) | §2 |
| CEL/CL2/CLX | sprite formats; CLX = devilutionX runtime re-encode of CEL/CL2 | §6 |
| TRN | palette translation table (256 entries) | §6 |
| PCX | still-image format (UI art) | §6 |
| MPQ | Blizzard archive format | §6 |
| MPFS | devilutionX's mpqfs C library backend | §6 |
| CMD_* | network command enum (~86 values) | §8 |
| PT_* | dvlnet packet types (10 values) | §8 |
| TCmd* | command payload structs (TCmdLoc, TCmdGItem…) | §8 |
| TPkt | player→players message packet (header + 493B body) | §8 |
| delta | per-level world-diff sync (DLevel/DJunk) | §8 |
| lockstep | deterministic simultaneous-turn MP model | §8 |
| timedemo | recorded-input deterministic replay | §12 |
| hero/`.sv` | per-character save file (MPQ container) | §5 |
| ItemPack/PlayerPack | on-disk/save packed structs | §5 |
| spawn | shareware mode (spawn.mpq, gbIsSpawn) | §4/§11 |
| Hellfire | official expansion (gbIsHellfire, levels 17–24, Monk) | §4 |
| setmap | quest fixed maps (SL_*) | §2 |
| theme | special room generator (THEME_*) | §2 |
| towner | town NPC | §3 |
| gmenu | in-game menu engine | §10 |
| DiabloUI | out-of-game UI framework (title/menus/hero) | §10 |
| qol | quality-of-life modules (stash, xpbar, labels…) | §11 |
| plrctrls | virtual-cursor gamepad controller layer | §9 |
| StickPush | fork's stick deflection classification (None/Weak/Firm) | §9 |
| headless | no-window build mode for CI/tests | §1/§11 |

## Appendix P — Verification checklist for this document

- [x] All 12 numbered section headings present (§1..§12)
- [x] ≥100 distinct `Source/...:line` citations, each opened during research
- [x] Concrete struct/enum/formula excerpts in every section
- [x] Effort ratings + open questions per section
- [x] Online claims carry URLs (Ebitengine, go-mpq, CL2 spec, mpqfs)
- [x] Skill paths verified with `ls` (Appendix A)
- [x] Decisions justified; unknowns recorded as open questions with decision criteria (Appendix H)

Companion deliverable: `/tmp/dg-bean-tree.md` (epics/tasks/subtasks JSON).
