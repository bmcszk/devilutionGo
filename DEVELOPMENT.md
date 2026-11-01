# Development

Dev-facing documentation: build commands, repo layout, CI gates.
Player-facing README lives at [README.md](README.md).

## Repo layout (three layers, one-way imports)

```
cmd/
├─ devilutionGo/           # GUI entry point
└─ devilutionGo-headless/  # headless entry point for CI / parity tests
internal/
├─ engine/                 # project-agnostic engine: loop, ECS, render, audio,
│                          # input, asset pipeline, net, save, config
├─ game/                   # shared Diablo-style systems: grid, LOS, path-finding,
│                          # combat, inventory, camera, UI
└─ diablo/                 # Diablo-specific: levels, monsters, items, spells,
                           # quests, assets, save schema, net packets
shaders/                   # .kage shader sources (palette, light, TRN)
test/                      # golden fixtures + parity tests (no Go packages)
assets/                    # build-time only (extracted MPQ data, manifests)
```

**Import rule (binding):**

| Layer | May import |
|-------|------------|
| `internal/engine` | nothing from this repo's game layers |
| `internal/game` | `internal/engine` |
| `internal/diablo` | `internal/engine`, `internal/game` |

One-way: `engine ← game ← diablo`. Never the reverse — `internal/engine`
must stay reusable for any isometric ARPG. AGENTS.md is the single source of
truth for architecture and conventions.

## Commands

`just` is required ([github.com/casey/just](https://github.com/casey/just)).

| Command | What it does |
|---------|--------------|
| `just build-gui` | Build `bin/devilutionGo` (GUI binary) |
| `just build-headless` | Build `bin/devilutionGo-headless` (`headless` build tag) |
| `just vet` | `go vet ./...` |
| `just lint` | `golangci-lint run ./...` |
| `just test` | `go test -count=1 ./...` (golden-fixture tests included) |
| `just fmt` | `gofmt -s -w .` |
| `just check` | Fast local gate: beans ledger + vet + lint |
| `just pre-push` | Full push gate: check + test (run by the git hook) |
| `just ci` | Full CI pipeline: fmt + vet + lint + test + beans-lint |

CI runs the same gates in [`.github/workflows/ci.yml`](.github/workflows/ci.yml).

## Debug cheats

Upstream keeps its cheats in `#ifdef _DEBUG` builds and replays them through
the Lua console (`Source/diablo.cpp:896-903`); here the same surface is an
explicit opt-in so normal builds carry no cheat code path at all.

| Piece | Where |
|-------|-------|
| Enable | pass `--debug` (or `-debug`) to the binary; `game.NewDebugRouterFromFlag` then returns a router, otherwise `nil` |
| Router | `internal/game/debug.go` — `DebugRouter.Execute("<command>")` returns upstream's status message |
| Cheats | `internal/game/debug_cheats.go` — monster spawn, item spawn, level jump |
| Game hooks | `DebugEnv` (injected; a nil hook fails closed) |
| Map of the stars | `internal/game/doom.go` — `DoomPortal` tracks `DoomFlag` + the loaded cel |

Commands (upstream source in brackets):

| Command | Effect |
|---------|--------|
| `spawn-monster <name> [count]` | spawns up to `count` (default 1) monsters of the named type on the nearest walkable, unoccupied tiles crawled from the player (`Source/lua/modules/dev/monsters.cpp:98-168`) |
| `spawn-item <name>` | generates an item whose name matches, reports the iteration count (`Source/lua/modules/dev/items.cpp:85-119`) |
| `jump-level <n>` | `StartNewLvl` to dungeon level `n`, bounded by `MaxLevel` (16, or 24 with Hellfire) (`Source/lua/modules/dev/level/warp.cpp:18-30`) |

Fail-closed rules: a nil `*DebugRouter` (the default) or a nil hook makes every
command return `game.ErrDebugDisabled` and touch nothing; malformed commands
return an error before any hook runs.

Parity policy for these cheats:

* `spawn-monster` placement is deterministic: the crawl order is `crawl.cpp`'s
  and draws no random values (`TestDebugRouter_SpawnMonster_KeepsLCGState`).
  The monster roll itself belongs to `SpawnMonster` (`Source/monster.cpp`).
* `jump-level` is deterministic and consumes no RNG.
* `spawn-item` is **excluded from parity**: upstream seeds the LCG from
  `BetterRng` (`Source/lua/modules/dev/items.cpp:87-89`, deliberately
  non-deterministic so repeated lookups do not loop), so the go-forward state
  cannot be reproduced. The Go hook owns that seeding.
* Quest-level warps and the level-21 Tristram return need quest state and are
  out of scope; `JumpLevel` carries upstream's `StartNewLvl` mode.

## Pre-push hook

`scripts/install-hooks.sh` installs a pre-push hook running `just pre-push`.
Bypass a single push with `git push --no-verify`.
