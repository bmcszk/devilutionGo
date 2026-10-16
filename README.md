# devilutionGo

A 1:1, byte-exact port of [devilutionX](https://github.com/diasurgical/devilutionX)
(the modern C++ Diablo 1 / Hellfire engine) to **Go 1.27 + Ebitengine v2 + SQLite**.

Every change must keep the parity contract: identical algorithms, constants,
RNG, data values, save-file format, network packets, pixel-perfect rendering.
See [AGENTS.md](AGENTS.md) — the single source of truth for architecture and
conventions.

## Layout (three layers, one-way imports)

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

The dependency direction is one-way: `engine ← game ← diablo`. Never the
reverse — `internal/engine` must stay reusable for any isometric ARPG.

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

CI runs the same gates in
[`.github/workflows/ci.yml`](.github/workflows/ci.yml).

## Status

Scaffold stage (milestone M0). The package tree, build gates, and CI wiring
are in place; game logic lands per-milestone with parity fixtures. Roadmap and
fidelity gates: [AGENTS.md](AGENTS.md).
